package progress

import (
	"bytes"
	"log/slog"
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	// Disable logs during testing
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelError,
	})))
	os.Exit(m.Run())
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	if config == nil {
		t.Fatal("DefaultConfig returned nil")
	}

	if config.UpdateInterval != 1*time.Second {
		t.Errorf("Expected UpdateInterval 1s, got %v", config.UpdateInterval)
	}

	if !config.ShowProgress {
		t.Error("Expected ShowProgress to be true")
	}

	if config.Output != os.Stderr {
		t.Error("Expected Output to be stderr")
	}

	if config.RateLimit != 0 {
		t.Errorf("Expected RateLimit 0, got %f", config.RateLimit)
	}
}

func TestNewProgressReporter(t *testing.T) {
	config := &Config{
		UpdateInterval: 500 * time.Millisecond,
		ShowProgress:   true,
		Output:         &bytes.Buffer{},
		RateLimit:      5.0,
	}

	pr := NewProgressReporter(config)
	if pr == nil {
		t.Fatal("NewProgressReporter returned nil")
	}

	if pr.updateInterval != config.UpdateInterval {
		t.Errorf("Expected UpdateInterval %v, got %v", config.UpdateInterval, pr.updateInterval)
	}

	if !pr.showProgress {
		t.Error("Expected ShowProgress to be true")
	}

	if pr.rateLimiter == nil {
		t.Error("Expected rate limiter to be created")
	}

	if !pr.rateLimiter.enabled {
		t.Error("Expected rate limiter to be enabled")
	}

	pr.Stop()
}

func TestNewProgressReporter_WithNilConfig(t *testing.T) {
	pr := NewProgressReporter(nil)
	if pr == nil {
		t.Fatal("NewProgressReporter returned nil")
	}

	if pr.updateInterval != 1*time.Second {
		t.Errorf("Expected default UpdateInterval 1s, got %v", pr.updateInterval)
	}

	if pr.rateLimiter != nil {
		t.Error("Expected no rate limiter with default config")
	}

	pr.Stop()
}

func TestStatsUpdates(t *testing.T) {
	buf := &bytes.Buffer{}
	config := &Config{
		ShowProgress: false, // Don't start progress display for this test
		Output:       buf,
	}

	pr := NewProgressReporter(config)
	defer pr.Stop()

	// Test initial stats
	stats := pr.GetStats()
	if stats.URLsProcessed != 0 {
		t.Errorf("Expected URLsProcessed 0, got %d", stats.URLsProcessed)
	}

	// Test increment methods
	pr.IncrementProcessed()
	pr.IncrementDiscovered()
	pr.IncrementFailed()
	pr.IncrementSkipped()

	stats = pr.GetStats()
	if stats.URLsProcessed != 1 {
		t.Errorf("Expected URLsProcessed 1, got %d", stats.URLsProcessed)
	}
	if stats.URLsDiscovered != 1 {
		t.Errorf("Expected URLsDiscovered 1, got %d", stats.URLsDiscovered)
	}
	if stats.URLsFailed != 1 {
		t.Errorf("Expected URLsFailed 1, got %d", stats.URLsFailed)
	}
	if stats.URLsSkipped != 1 {
		t.Errorf("Expected URLsSkipped 1, got %d", stats.URLsSkipped)
	}

	// Test bulk update
	pr.UpdateStats(10, 20, 2, 3, 5, 8)
	stats = pr.GetStats()
	if stats.URLsProcessed != 10 {
		t.Errorf("Expected URLsProcessed 10, got %d", stats.URLsProcessed)
	}
	if stats.URLsDiscovered != 20 {
		t.Errorf("Expected URLsDiscovered 20, got %d", stats.URLsDiscovered)
	}
	if stats.ActiveWorkers != 5 {
		t.Errorf("Expected ActiveWorkers 5, got %d", stats.ActiveWorkers)
	}
	if stats.QueueSize != 8 {
		t.Errorf("Expected QueueSize 8, got %d", stats.QueueSize)
	}
}

func TestProgressDisplay(t *testing.T) {
	buf := &bytes.Buffer{}
	config := &Config{
		UpdateInterval: 100 * time.Millisecond,
		ShowProgress:   true,
		Output:         buf,
	}

	pr := NewProgressReporter(config)
	pr.Start()

	// Update stats and wait for progress display
	pr.UpdateStats(5, 10, 0, 0, 2, 3)
	time.Sleep(150 * time.Millisecond)

	pr.Stop()

	output := buf.String()
	if output == "" {
		t.Error("Expected progress output, got empty string")
	}

	// Check if the output contains expected elements
	expectedElements := []string{"URLs processed", "URLs/sec", "workers", "queued"}
	for _, element := range expectedElements {
		if !bytes.Contains([]byte(output), []byte(element)) {
			t.Errorf("Expected output to contain '%s', got: %s", element, output)
		}
	}
}

func TestRateLimiter(t *testing.T) {
	// Test rate limiter creation
	rl := NewRateLimiter(10.0) // 10 requests per second
	if rl == nil {
		t.Fatal("NewRateLimiter returned nil")
	}

	if !rl.enabled {
		t.Error("Expected rate limiter to be enabled")
	}

	if rl.requestsPerSecond != 10.0 {
		t.Errorf("Expected requestsPerSecond 10.0, got %f", rl.requestsPerSecond)
	}

	// Test that Wait() doesn't block immediately (tokens should be available)
	start := time.Now()
	rl.Wait()
	elapsed := time.Since(start)

	if elapsed > 50*time.Millisecond {
		t.Errorf("Rate limiter Wait() took too long: %v", elapsed)
	}

	rl.Stop()
}

func TestRateLimiter_Disabled(t *testing.T) {
	// Test with rate limit of 0 (disabled)
	rl := NewRateLimiter(0)
	if rl == nil {
		t.Fatal("NewRateLimiter returned nil")
	}

	if rl.enabled {
		t.Error("Expected rate limiter to be disabled")
	}

	// Wait should not block when disabled
	start := time.Now()
	rl.Wait()
	elapsed := time.Since(start)

	if elapsed > 10*time.Millisecond {
		t.Errorf("Disabled rate limiter Wait() took too long: %v", elapsed)
	}
}

func TestProgressReporter_RateLimit(t *testing.T) {
	config := &Config{
		ShowProgress: false,
		RateLimit:    100.0, // High rate to avoid blocking in test
	}

	pr := NewProgressReporter(config)
	defer pr.Stop()

	if !pr.IsRateLimited() {
		t.Error("Expected progress reporter to be rate limited")
	}

	// Test WaitForRateLimit doesn't block with high rate
	start := time.Now()
	pr.WaitForRateLimit()
	elapsed := time.Since(start)

	if elapsed > 50*time.Millisecond {
		t.Errorf("WaitForRateLimit took too long: %v", elapsed)
	}
}

func TestProgressReporter_NoRateLimit(t *testing.T) {
	config := &Config{
		ShowProgress: false,
		RateLimit:    0, // No rate limiting
	}

	pr := NewProgressReporter(config)
	defer pr.Stop()

	if pr.IsRateLimited() {
		t.Error("Expected progress reporter to not be rate limited")
	}

	// WaitForRateLimit should not block when no rate limiting
	start := time.Now()
	pr.WaitForRateLimit()
	elapsed := time.Since(start)

	if elapsed > 10*time.Millisecond {
		t.Errorf("WaitForRateLimit took too long without rate limiting: %v", elapsed)
	}
}

func TestProcessingRate(t *testing.T) {
	config := &Config{
		ShowProgress: false,
	}

	pr := NewProgressReporter(config)
	defer pr.Stop()

	// Wait a bit to ensure elapsed time > 0
	time.Sleep(10 * time.Millisecond)

	// Update with some processed URLs
	pr.UpdateStats(10, 15, 0, 0, 2, 3)

	stats := pr.GetStats()
	if stats.ProcessingRate <= 0 {
		t.Errorf("Expected positive processing rate, got %f", stats.ProcessingRate)
	}

	if stats.ElapsedTime <= 0 {
		t.Errorf("Expected positive elapsed time, got %v", stats.ElapsedTime)
	}
}

func TestFinalStats(t *testing.T) {
	buf := &bytes.Buffer{}
	config := &Config{
		ShowProgress: true,
		Output:       buf,
	}

	pr := NewProgressReporter(config)

	// Update stats before stopping
	pr.UpdateStats(25, 30, 2, 1, 0, 0)

	pr.Stop()

	output := buf.String()
	expectedElements := []string{
		"Crawling completed",
		"URLs discovered: 30",
		"URLs processed:  25",
		"URLs failed:     2",
		"URLs skipped:    1",
		"Average rate:",
	}

	for _, element := range expectedElements {
		if !bytes.Contains([]byte(output), []byte(element)) {
			t.Errorf("Expected final stats to contain '%s', got: %s", element, output)
		}
	}
}
