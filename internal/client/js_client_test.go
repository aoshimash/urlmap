package client

import (
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/aoshimash/urlmap/test/shared"
)

func TestNewJSClient(t *testing.T) {
	logger := slog.Default()

	// Test with default config
	client, err := NewJSClient(nil, logger)
	if err != nil {
		t.Fatalf("Failed to create JS client: %v", err)
	}
	defer client.Close()

	// Test with custom config
	config := &JSConfig{
		Enabled:     true,
		BrowserType: "chromium",
		Headless:    true,
		Timeout:     90 * time.Second,
		WaitFor:     "networkidle",
		PoolSize:    2,
	}

	client2, err := NewJSClient(config, logger)
	if err != nil {
		t.Fatalf("Failed to create JS client with custom config: %v", err)
	}
	defer client2.Close()
}

func TestJSClient_RenderPage(t *testing.T) {
	// Create test server for more reliable testing in CI
	testServer := shared.CreateBasicTestServer()
	defer testServer.Close()

	logger := slog.Default()
	config := &JSConfig{
		Enabled:     true,
		BrowserType: "chromium",
		Headless:    true,
		Timeout:     30 * time.Second,
		WaitFor:     "networkidle",
		PoolSize:    2,
	}

	client, err := NewJSClient(config, logger)
	if err != nil {
		t.Fatalf("Failed to create JS client: %v", err)
	}
	defer client.Close()

	// Test rendering a simple page
	ctx := context.Background()
	content, err := client.RenderPage(ctx, testServer.URL)
	if err != nil {
		t.Fatalf("Failed to render page: %v", err)
	}

	if content == "" {
		t.Error("Rendered content is empty")
	}

	// Check that content contains expected elements from test server
	if !strings.Contains(content, "Test Home Page") {
		t.Error("Rendered content does not contain expected title")
	}

	if !strings.Contains(content, "Page 1") || !strings.Contains(content, "Page 2") {
		t.Error("Rendered content does not contain expected links")
	}
}

func TestJSClient_Get(t *testing.T) {
	// Create test server for more reliable testing in CI
	testServer := shared.CreateBasicTestServer()
	defer testServer.Close()

	logger := slog.Default()
	config := &JSConfig{
		Enabled:     true,
		BrowserType: "chromium",
		Headless:    true,
		Timeout:     30 * time.Second,
		WaitFor:     "networkidle",
		PoolSize:    2,
	}

	client, err := NewJSClient(config, logger)
	if err != nil {
		t.Fatalf("Failed to create JS client: %v", err)
	}
	defer client.Close()

	// Test getting a page
	ctx := context.Background()
	response, err := client.Get(ctx, testServer.URL)
	if err != nil {
		t.Fatalf("Failed to get page: %v", err)
	}

	if response == nil {
		t.Fatal("Response is nil")
	}

	if response.URL != testServer.URL {
		t.Errorf("Expected URL '%s', got '%s'", testServer.URL, response.URL)
	}

	if response.Content == "" {
		t.Error("Response content is empty")
	}

	if response.Status != 200 {
		t.Errorf("Expected status 200, got %d", response.Status)
	}

	// Check that content contains expected elements from test server
	if !strings.Contains(response.Content, "Test Home Page") {
		t.Error("Response content does not contain expected title")
	}
}

func TestJSClient_GetPoolStats(t *testing.T) {
	logger := slog.Default()
	config := &JSConfig{
		Enabled:     true,
		BrowserType: "chromium",
		Headless:    true,
		Timeout:     90 * time.Second,
		WaitFor:     "networkidle",
		PoolSize:    2,
	}

	client, err := NewJSClient(config, logger)
	if err != nil {
		t.Fatalf("Failed to create JS client: %v", err)
	}
	defer client.Close()

	stats := client.GetPoolStats()
	if stats == nil {
		t.Error("Pool stats should not be nil")
	}

	// Check required fields
	requiredFields := []string{"initialized", "closed", "browsers_active", "browsers_available", "browser_pool_size", "context_pool_size", "contexts_available", "max_contexts"}
	for _, field := range requiredFields {
		if _, exists := stats[field]; !exists {
			t.Errorf("Stats missing required field: %s", field)
		}
	}
}

func TestJSClient_Disabled(t *testing.T) {
	logger := slog.Default()
	config := &JSConfig{
		Enabled:     false,
		BrowserType: "chromium",
		Headless:    true,
		Timeout:     90 * time.Second,
		PoolSize:    2,
	}

	client, err := NewJSClient(config, logger)
	if err != nil {
		t.Fatalf("Failed to create JS client: %v", err)
	}
	defer client.Close()

	// Test that rendering fails when disabled
	ctx := context.Background()
	_, err = client.RenderPage(ctx, "https://example.com")
	if err == nil {
		t.Error("Should fail to render when JS is disabled")
	}

	expectedErr := "JavaScript rendering is not enabled"
	if err.Error() != expectedErr {
		t.Errorf("Expected error '%s', got '%s'", expectedErr, err.Error())
	}
}

func TestJSResponse_String(t *testing.T) {
	response := &JSResponse{
		URL:     "https://example.com",
		Content: "<html><body>Test content</body></html>",
		Status:  200,
		Headers: map[string]string{"Content-Type": "text/html"},
		Host:    "example.com",
	}

	content := response.String()
	if content != "<html><body>Test content</body></html>" {
		t.Errorf("Expected content '<html><body>Test content</body></html>', got '%s'", content)
	}
}

func TestJSResponse_StatusCode(t *testing.T) {
	response := &JSResponse{
		URL:     "https://example.com",
		Content: "<html><body>Test content</body></html>",
		Status:  200,
		Headers: map[string]string{"Content-Type": "text/html"},
		Host:    "example.com",
	}

	status := response.StatusCode()
	if status != 200 {
		t.Errorf("Expected status 200, got %d", status)
	}
}

func TestJSClient_CacheHit(t *testing.T) {
	// Create test server
	testServer := shared.CreateBasicTestServer()
	defer testServer.Close()

	logger := slog.Default()
	config := &JSConfig{
		Enabled:      true,
		BrowserType:  "chromium",
		Headless:     true,
		Timeout:      30 * time.Second,
		WaitFor:      "networkidle",
		PoolSize:     1,
		CacheEnabled: true,
		CacheSize:    10,
		CacheTTL:     1 * time.Hour,
	}

	client, err := NewJSClient(config, logger)
	if err != nil {
		t.Fatalf("Failed to create JS client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// First request - should not be cached
	response1, err := client.Get(ctx, testServer.URL)
	if err != nil {
		t.Fatalf("Failed to get page: %v", err)
	}

	if response1.FromCache {
		t.Error("First request should not be from cache")
	}

	// Second request - should be cached
	response2, err := client.Get(ctx, testServer.URL)
	if err != nil {
		t.Fatalf("Failed to get cached page: %v", err)
	}

	if !response2.FromCache {
		t.Error("Second request should be from cache")
	}

	// Content should be the same
	if response1.Content != response2.Content {
		t.Error("Cached content should match original content")
	}

	// Check cache stats
	cacheStats := client.GetCacheStats()
	if cacheStats == nil {
		t.Fatal("Cache stats should not be nil")
	}

	if cacheStats["size"].(int) != 1 {
		t.Errorf("Expected cache size 1, got %v", cacheStats["size"])
	}
}

func TestJSClient_CacheExpiration(t *testing.T) {
	// Create test server
	testServer := shared.CreateBasicTestServer()
	defer testServer.Close()

	logger := slog.Default()
	config := &JSConfig{
		Enabled:      true,
		BrowserType:  "chromium",
		Headless:     true,
		Timeout:      30 * time.Second,
		WaitFor:      "networkidle",
		PoolSize:     1,
		CacheEnabled: true,
		CacheSize:    10,
		CacheTTL:     100 * time.Millisecond, // Short TTL for testing
	}

	client, err := NewJSClient(config, logger)
	if err != nil {
		t.Fatalf("Failed to create JS client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// First request
	response1, err := client.Get(ctx, testServer.URL)
	if err != nil {
		t.Fatalf("Failed to get page: %v", err)
	}

	if response1.FromCache {
		t.Error("First request should not be from cache")
	}

	// Wait for cache to expire
	time.Sleep(150 * time.Millisecond)

	// Second request - should not be cached (expired)
	response2, err := client.Get(ctx, testServer.URL)
	if err != nil {
		t.Fatalf("Failed to get page after expiration: %v", err)
	}

	if response2.FromCache {
		t.Error("Request after expiration should not be from cache")
	}
}

func TestJSClient_CacheDisabled(t *testing.T) {
	// Create test server
	testServer := shared.CreateBasicTestServer()
	defer testServer.Close()

	logger := slog.Default()
	config := &JSConfig{
		Enabled:      true,
		BrowserType:  "chromium",
		Headless:     true,
		Timeout:      30 * time.Second,
		WaitFor:      "networkidle",
		PoolSize:     1,
		CacheEnabled: false, // Cache disabled
	}

	client, err := NewJSClient(config, logger)
	if err != nil {
		t.Fatalf("Failed to create JS client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// First request
	response1, err := client.Get(ctx, testServer.URL)
	if err != nil {
		t.Fatalf("Failed to get page: %v", err)
	}

	// Second request - should not be cached
	response2, err := client.Get(ctx, testServer.URL)
	if err != nil {
		t.Fatalf("Failed to get page: %v", err)
	}

	if response1.FromCache || response2.FromCache {
		t.Error("No requests should be from cache when caching is disabled")
	}

	// Cache stats should be nil
	cacheStats := client.GetCacheStats()
	if cacheStats != nil {
		t.Error("Cache stats should be nil when caching is disabled")
	}
}
