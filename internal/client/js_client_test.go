package client

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"
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
		Timeout:     30 * time.Second,
		WaitFor:     "networkidle",
	}

	client2, err := NewJSClient(config, logger)
	if err != nil {
		t.Fatalf("Failed to create JS client with custom config: %v", err)
	}
	defer client2.Close()
}

func TestJSClient_RenderPage(t *testing.T) {
	logger := slog.Default()
	config := &JSConfig{
		Enabled:     true,
		BrowserType: "chromium",
		Headless:    true,
		Timeout:     30 * time.Second,
		WaitFor:     "networkidle",
	}

	client, err := NewJSClient(config, logger)
	if err != nil {
		t.Fatalf("Failed to create JS client: %v", err)
	}
	defer client.Close()

	// Test rendering a simple page
	ctx := context.Background()
	content, err := client.RenderPage(ctx, "https://example.com")
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

func TestJSClient_Get(t *testing.T) {
	logger := slog.Default()
	config := &JSConfig{
		Enabled:     true,
		BrowserType: "chromium",
		Headless:    true,
		Timeout:     30 * time.Second,
		WaitFor:     "networkidle",
	}

	client, err := NewJSClient(config, logger)
	if err != nil {
		t.Fatalf("Failed to create JS client: %v", err)
	}
	defer client.Close()

	// Test getting a page
	ctx := context.Background()
	response, err := client.Get(ctx, "https://example.com")
	if err != nil {
		t.Fatalf("Failed to get page: %v", err)
	}

	if response == nil {
		t.Fatal("Response is nil")
	}

	if response.URL != "https://example.com" {
		t.Errorf("Expected URL 'https://example.com', got '%s'", response.URL)
	}

	if response.Content == "" {
		t.Error("Response content is empty")
	}

	if response.Status != 200 {
		t.Errorf("Expected status 200, got %d", response.Status)
	}

	if response.Host != "example.com" {
		t.Errorf("Expected host 'example.com', got '%s'", response.Host)
	}
}

func TestJSClient_GetPoolStats(t *testing.T) {
	logger := slog.Default()
	config := &JSConfig{
		Enabled:     true,
		BrowserType: "chromium",
		Headless:    true,
		Timeout:     30 * time.Second,
		WaitFor:     "networkidle",
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
	requiredFields := []string{"initialized", "closed", "pool_size", "max_contexts", "available"}
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
		Timeout:     30 * time.Second,
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
