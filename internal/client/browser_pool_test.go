package client

import (
	"context"
	"log/slog"
	"testing"
	"time"
)

func TestNewBrowserPool(t *testing.T) {
	logger := slog.Default()

	// Test with default config
	pool, err := NewBrowserPool(nil, logger)
	if err != nil {
		t.Fatalf("Failed to create browser pool: %v", err)
	}
	defer pool.Close()

	// Test with custom config
	config := &JSConfig{
		Enabled:     true,
		BrowserType: "chromium",
		Headless:    true,
		Timeout:     30 * time.Second,
	}

	pool2, err := NewBrowserPool(config, logger)
	if err != nil {
		t.Fatalf("Failed to create browser pool with custom config: %v", err)
	}
	defer pool2.Close()
}

func TestBrowserPool_AcquireContext(t *testing.T) {
	logger := slog.Default()
	config := &JSConfig{
		Enabled:     true,
		BrowserType: "chromium",
		Headless:    true,
		Timeout:     30 * time.Second,
	}

	pool, err := NewBrowserPool(config, logger)
	if err != nil {
		t.Fatalf("Failed to create browser pool: %v", err)
	}
	defer pool.Close()

	// Acquire first context
	ctx1, err := pool.AcquireContext()
	if err != nil {
		t.Fatalf("Failed to acquire context: %v", err)
	}

	if ctx1.Context == nil {
		t.Error("Browser context is nil")
	}

	if ctx1.Pool != pool {
		t.Error("Context pool reference is incorrect")
	}

	// Release context
	ctx1.ReleaseContext()

	// Acquire second context (should reuse from pool)
	ctx2, err := pool.AcquireContext()
	if err != nil {
		t.Fatalf("Failed to acquire second context: %v", err)
	}

	if ctx2.Context == nil {
		t.Error("Second browser context is nil")
	}

	ctx2.ReleaseContext()
}

func TestBrowserPool_RenderPage(t *testing.T) {
	logger := slog.Default()
	config := &JSConfig{
		Enabled:     true,
		BrowserType: "chromium",
		Headless:    true,
		Timeout:     30 * time.Second,
	}

	pool, err := NewBrowserPool(config, logger)
	if err != nil {
		t.Fatalf("Failed to create browser pool: %v", err)
	}
	defer pool.Close()

	// Test rendering a simple page
	ctx := context.Background()
	content, err := pool.RenderPage(ctx, "https://example.com")
	if err != nil {
		t.Fatalf("Failed to render page: %v", err)
	}

	if content == "" {
		t.Error("Rendered content is empty")
	}

	// Check that content contains expected elements
	if len(content) < 100 {
		t.Error("Rendered content seems too short")
	}
}

func TestBrowserPool_GetPoolStats(t *testing.T) {
	logger := slog.Default()
	config := &JSConfig{
		Enabled:     true,
		BrowserType: "chromium",
		Headless:    true,
		Timeout:     30 * time.Second,
		WaitFor:     "networkidle",
	}

	pool, err := NewBrowserPool(config, logger)
	if err != nil {
		t.Fatalf("Failed to create browser pool: %v", err)
	}
	defer pool.Close()

	stats := pool.GetPoolStats()

	// Check required fields
	requiredFields := []string{"initialized", "closed", "pool_size", "max_contexts", "available"}
	for _, field := range requiredFields {
		if _, exists := stats[field]; !exists {
			t.Errorf("Stats missing required field: %s", field)
		}
	}

	// Check initial state
	if !stats["initialized"].(bool) {
		t.Error("Pool should be initialized")
	}

	if stats["closed"].(bool) {
		t.Error("Pool should not be closed initially")
	}

	if stats["max_contexts"].(int) != 10 {
		t.Error("Default max contexts should be 10")
	}
}

func TestBrowserPool_ConcurrentAccess(t *testing.T) {
	logger := slog.Default()
	config := &JSConfig{
		Enabled:     true,
		BrowserType: "chromium",
		Headless:    true,
		Timeout:     30 * time.Second,
		WaitFor:     "networkidle",
	}

	pool, err := NewBrowserPool(config, logger)
	if err != nil {
		t.Fatalf("Failed to create browser pool: %v", err)
	}
	defer pool.Close()

	// Test concurrent context acquisition
	done := make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func(id int) {
			defer func() { done <- true }()

			ctx, err := pool.AcquireContext()
			if err != nil {
				t.Errorf("Worker %d failed to acquire context: %v", id, err)
				return
			}

			// Simulate some work
			time.Sleep(10 * time.Millisecond)

			ctx.ReleaseContext()
		}(i)
	}

	// Wait for all workers to complete
	for i := 0; i < 5; i++ {
		<-done
	}

	// Check pool stats after concurrent access
	stats := pool.GetPoolStats()
	if stats["available"].(int) > 10 {
		t.Error("Pool should not exceed max contexts")
	}
}

func TestBrowserPool_Close(t *testing.T) {
	logger := slog.Default()
	config := &JSConfig{
		Enabled:     true,
		BrowserType: "chromium",
		Headless:    true,
		Timeout:     30 * time.Second,
		WaitFor:     "networkidle",
	}

	pool, err := NewBrowserPool(config, logger)
	if err != nil {
		t.Fatalf("Failed to create browser pool: %v", err)
	}

	// Close the pool
	err = pool.Close()
	if err != nil {
		t.Fatalf("Failed to close pool: %v", err)
	}

	// Try to acquire context after closing
	_, err = pool.AcquireContext()
	if err == nil {
		t.Error("Should not be able to acquire context after pool is closed")
	}

	// Check stats after closing
	stats := pool.GetPoolStats()
	if !stats["closed"].(bool) {
		t.Error("Pool should be marked as closed")
	}
}

func TestBrowserContext_ReleaseContext(t *testing.T) {
	logger := slog.Default()
	config := &JSConfig{
		Enabled:     true,
		BrowserType: "chromium",
		Headless:    true,
		Timeout:     30 * time.Second,
		WaitFor:     "networkidle",
	}

	pool, err := NewBrowserPool(config, logger)
	if err != nil {
		t.Fatalf("Failed to create browser pool: %v", err)
	}
	defer pool.Close()

	// Acquire context
	ctx, err := pool.AcquireContext()
	if err != nil {
		t.Fatalf("Failed to acquire context: %v", err)
	}

	// Release context
	ctx.ReleaseContext()

	// Try to release again (should be safe)
	ctx.ReleaseContext()

	// Check that context is properly released
	stats := pool.GetPoolStats()
	available := stats["available"].(int)
	if available < 0 || available > 10 {
		t.Errorf("Available contexts should be between 0 and 10, got %d", available)
	}
}
