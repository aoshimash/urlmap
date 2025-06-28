package crawler

import (
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/aoshimash/urlmap/internal/progress"
)

func TestMain(m *testing.M) {
	// Disable logs during testing
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelError,
	})))
	os.Exit(m.Run())
}

func TestNew(t *testing.T) {
	crawler, err := New(nil)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	if crawler == nil {
		t.Fatal("New() returned nil crawler")
	}

	// Test that all fields are properly initialized
	if crawler.client == nil {
		t.Error("Crawler client is nil")
	}

	if crawler.parser == nil {
		t.Error("Crawler parser is nil")
	}

	if crawler.logger == nil {
		t.Error("Crawler logger is nil")
	}

	if crawler.visited == nil {
		t.Error("Crawler visited map is nil")
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	if config == nil {
		t.Fatal("DefaultConfig returned nil")
	}

	if config.MaxDepth != 3 {
		t.Errorf("Expected MaxDepth 3, got %d", config.MaxDepth)
	}

	if !config.SameDomain {
		t.Error("Expected SameDomain to be true")
	}

	if config.UserAgent != "urlmap/1.0" {
		t.Errorf("Expected UserAgent 'urlmap/1.0', got %s", config.UserAgent)
	}

	if config.Timeout != 30*time.Second {
		t.Errorf("Expected Timeout 30s, got %v", config.Timeout)
	}
}

func TestCrawlRecursive_InvalidURL(t *testing.T) {
	crawler, err := New(nil)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	invalidURLs := []string{
		"",
		"not-a-url",
		"ftp://example.com",
		"javascript:void(0)",
	}

	for _, url := range invalidURLs {
		t.Run("invalid_url_"+url, func(t *testing.T) {
			results, stats, err := crawler.CrawlRecursive(url)
			if err == nil {
				t.Error("Expected error for invalid URL")
			}

			if len(results) != 0 {
				t.Errorf("Expected 0 results, got %d", len(results))
			}

			if stats.TotalURLs != 0 {
				t.Errorf("Expected 0 total URLs, got %d", stats.TotalURLs)
			}

			// Reset for next test
			crawler.Reset()
		})
	}
}

func TestCrawlerUtilityMethods(t *testing.T) {
	crawler, err := New(nil)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	// Test Reset
	crawler.visited["test"] = true
	crawler.results = append(crawler.results, CrawlResult{URL: "test"})
	crawler.stats.TotalURLs = 5

	crawler.Reset()

	if len(crawler.visited) != 0 {
		t.Error("Reset() did not clear visited map")
	}

	if len(crawler.results) != 0 {
		t.Error("Reset() did not clear results")
	}

	if crawler.stats.TotalURLs != 0 {
		t.Error("Reset() did not reset stats")
	}

	// Test GetResults and GetStats
	testResult := CrawlResult{URL: "test", Depth: 1}
	crawler.results = append(crawler.results, testResult)
	crawler.stats.CrawledURLs = 1

	results := crawler.GetResults()
	if len(results) != 1 || results[0].URL != "test" {
		t.Error("GetResults() did not return correct results")
	}

	stats := crawler.GetStats()
	if stats.CrawledURLs != 1 {
		t.Error("GetStats() did not return correct stats")
	}
}

func TestNewConcurrentCrawler(t *testing.T) {
	config := &Config{
		MaxDepth:   2,
		SameDomain: true,
		UserAgent:  "test/1.0",
		Timeout:    10 * time.Second,
		Workers:    5,
	}

	cc, err := NewConcurrentCrawler(config)
	if err != nil {
		t.Fatalf("NewConcurrentCrawler() failed: %v", err)
	}

	if cc == nil {
		t.Fatal("NewConcurrentCrawler() returned nil")
	}

	// Test that all fields are properly initialized
	if cc.Crawler == nil {
		t.Error("ConcurrentCrawler.Crawler is nil")
	}

	if cc.jobs == nil {
		t.Error("ConcurrentCrawler.jobs channel is nil")
	}

	if cc.results == nil {
		t.Error("ConcurrentCrawler.results channel is nil")
	}

	if cc.ctx == nil {
		t.Error("ConcurrentCrawler.ctx is nil")
	}

	if cc.cancel == nil {
		t.Error("ConcurrentCrawler.cancel is nil")
	}

	if cc.workers != 5 {
		t.Errorf("Expected workers 5, got %d", cc.workers)
	}
}

func TestConcurrentCrawler_InvalidURL(t *testing.T) {
	cc, err := NewConcurrentCrawler(nil)
	if err != nil {
		t.Fatalf("NewConcurrentCrawler() failed: %v", err)
	}

	invalidURLs := []string{
		"",
		"not-a-url",
		"ftp://example.com",
		"javascript:void(0)",
	}

	for _, url := range invalidURLs {
		t.Run("invalid_url_"+url, func(t *testing.T) {
			results, stats, err := cc.CrawlConcurrent(url)
			if err == nil {
				t.Error("Expected error for invalid URL")
			}

			if len(results) != 0 {
				t.Errorf("Expected 0 results, got %d", len(results))
			}

			if stats.TotalURLs != 0 {
				t.Errorf("Expected 0 total URLs, got %d", stats.TotalURLs)
			}
		})
	}
}

func TestConcurrentCrawler_Cancel(t *testing.T) {
	cc, err := NewConcurrentCrawler(&Config{
		Workers:  2,
		MaxDepth: 10, // Deep crawl to ensure it doesn't finish quickly
		Timeout:  1 * time.Second,
	})
	if err != nil {
		t.Fatalf("NewConcurrentCrawler() failed: %v", err)
	}

	// Start crawling in a goroutine
	done := make(chan struct{})
	go func() {
		defer close(done)
		// Use a URL that would take time to crawl if it were real
		cc.CrawlConcurrent("https://httpbin.org/delay/5")
	}()

	// Cancel immediately
	cc.Cancel()

	// Wait for completion with timeout
	select {
	case <-done:
		// Test passed - crawl was cancelled
	case <-time.After(2 * time.Second):
		t.Error("Crawl did not complete within expected time after cancellation")
	}
}

func TestConcurrentCrawler_ThreadSafety(t *testing.T) {
	// This test ensures no race conditions occur during concurrent operations
	cc, err := NewConcurrentCrawler(&Config{
		Workers:  2,
		MaxDepth: 1, // Limit depth to prevent long running test
		Timeout:  2 * time.Second,
	})
	if err != nil {
		t.Fatalf("NewConcurrentCrawler() failed: %v", err)
	}

	// Test concurrent access to GetResults and GetStats
	done := make(chan struct{})
	go func() {
		defer close(done)
		// Use a simpler URL that's more likely to work
		cc.CrawlConcurrent("https://example.com")
	}()

	// Concurrently call GetResults and GetStats
	for i := 0; i < 5; i++ {
		go func() {
			cc.GetResults()
			cc.GetStats()
		}()
	}

	// Wait for completion with a shorter timeout
	select {
	case <-done:
		// Test completed successfully
	case <-time.After(5 * time.Second):
		t.Error("Test timed out")
		cc.Cancel()
		<-done // Wait for goroutine to finish
	}
}

func TestConcurrentCrawler_WorkerConfiguration(t *testing.T) {
	testCases := []struct {
		name     string
		workers  int
		expected int
	}{
		{"default_workers", 0, 10},
		{"custom_workers", 5, 5},
		{"negative_workers", -1, 10},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &Config{
				Workers: tc.workers,
			}

			cc, err := NewConcurrentCrawler(config)
			if err != nil {
				t.Fatalf("NewConcurrentCrawler() failed: %v", err)
			}

			if cc.workers != tc.expected {
				t.Errorf("Expected %d workers, got %d", tc.expected, cc.workers)
			}
		})
	}
}

// Benchmark to compare concurrent vs sequential performance
func BenchmarkCrawler_Sequential(b *testing.B) {
	config := DefaultConfig()
	config.MaxDepth = 1 // Limit depth for benchmark

	c, err := New(config)
	if err != nil {
		b.Fatalf("New() failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = c.CrawlRecursive("https://httpbin.org/html")
		c.Reset()
	}
}

func BenchmarkCrawler_Concurrent(b *testing.B) {
	config := DefaultConfig()
	config.MaxDepth = 1 // Limit depth for benchmark

	cc, err := NewConcurrentCrawler(config)
	if err != nil {
		b.Fatalf("NewConcurrentCrawler() failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = cc.CrawlConcurrent("https://httpbin.org/html")
	}
}

// TestConcurrentCrawler_WithProgress tests concurrent crawler with progress reporting
func TestConcurrentCrawler_WithProgress(t *testing.T) {
	// Create a config with progress reporting enabled
	config := &Config{
		MaxDepth:     1,
		SameDomain:   true,
		UserAgent:    "test-agent",
		Workers:      2,
		ShowProgress: true,
		ProgressConfig: &progress.Config{
			ShowProgress: true,
			RateLimit:    0, // No rate limiting for test
		},
	}

	crawler, err := NewConcurrentCrawler(config)
	if err != nil {
		t.Fatalf("Failed to create concurrent crawler: %v", err)
	}

	// Check that progress reporter is initialized
	if crawler.progress == nil {
		t.Error("Expected progress reporter to be initialized")
	}

	if crawler.progress.IsRateLimited() {
		t.Error("Expected rate limiting to be disabled")
	}
}

// TestConcurrentCrawler_WithRateLimit tests concurrent crawler with rate limiting
func TestConcurrentCrawler_WithRateLimit(t *testing.T) {
	// Create a config with rate limiting enabled
	config := &Config{
		MaxDepth:     1,
		SameDomain:   true,
		UserAgent:    "test-agent",
		Workers:      2,
		ShowProgress: true,
		ProgressConfig: &progress.Config{
			ShowProgress: true,
			RateLimit:    10.0, // 10 requests per second
		},
	}

	crawler, err := NewConcurrentCrawler(config)
	if err != nil {
		t.Fatalf("Failed to create concurrent crawler: %v", err)
	}

	// Check that progress reporter is initialized with rate limiting
	if crawler.progress == nil {
		t.Error("Expected progress reporter to be initialized")
	}

	if !crawler.progress.IsRateLimited() {
		t.Error("Expected rate limiting to be enabled")
	}
}

// TestConcurrentCrawler_WithoutProgress tests concurrent crawler without progress reporting
func TestConcurrentCrawler_WithoutProgress(t *testing.T) {
	// Create a config with progress reporting disabled
	config := &Config{
		MaxDepth:     1,
		SameDomain:   true,
		UserAgent:    "test-agent",
		Workers:      2,
		ShowProgress: false,
	}

	crawler, err := NewConcurrentCrawler(config)
	if err != nil {
		t.Fatalf("Failed to create concurrent crawler: %v", err)
	}

	// Check that progress reporter is not initialized
	if crawler.progress != nil {
		t.Error("Expected progress reporter to not be initialized")
	}
}
