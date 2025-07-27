package client

import (
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUnifiedClient_DefaultConfig(t *testing.T) {
	client, err := NewUnifiedClient(nil, slog.Default())

	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.NotNil(t, client.httpClient)
	assert.Nil(t, client.jsClient) // JS should be disabled by default
	assert.False(t, client.IsJSEnabled())
}

func TestNewUnifiedClient_HTTPOnly(t *testing.T) {
	config := &UnifiedConfig{
		UserAgent: "test-agent",
		JSConfig:  &JSConfig{Enabled: false},
	}

	client, err := NewUnifiedClient(config, slog.Default())

	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.NotNil(t, client.httpClient)
	assert.Nil(t, client.jsClient)
	assert.False(t, client.IsJSEnabled())
	assert.Equal(t, "test-agent", client.config.UserAgent)
}

func TestNewUnifiedClient_JSEnabled(t *testing.T) {

	config := &UnifiedConfig{
		UserAgent: "test-agent",
		JSConfig: &JSConfig{
			Enabled:     true,
			BrowserType: "chromium",
			Headless:    true,
			Timeout:     10 * time.Second,
			WaitFor:     "networkidle",
			UserAgent:   "", // Should inherit from UnifiedConfig
		},
	}

	// This will fail without Playwright installation, but we test the configuration
	client, err := NewUnifiedClient(config, slog.Default())

	if err != nil {
		// Expected in environments without Playwright setup
		assert.Contains(t, err.Error(), "failed to create JS client")
		return
	}

	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.NotNil(t, client.httpClient)
	assert.NotNil(t, client.jsClient)
	assert.True(t, client.IsJSEnabled())
	assert.Equal(t, "test-agent", client.jsClient.config.UserAgent)

	// Clean up
	client.Close()
}

func TestNewUnifiedClient_InvalidJSConfig(t *testing.T) {
	config := &UnifiedConfig{
		UserAgent: "test-agent",
		JSConfig: &JSConfig{
			Enabled:     true,
			BrowserType: "invalid-browser",
			Timeout:     10 * time.Second,
			WaitFor:     "networkidle",
		},
	}

	client, err := NewUnifiedClient(config, slog.Default())

	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "invalid browser type")
}

func TestUnifiedClient_Get_HTTPOnly(t *testing.T) {
	t.Skip("Requires external HTTP request")

	// Example of what an integration test would look like:
	/*
		config := &UnifiedConfig{
			UserAgent: "test-agent",
			JSConfig:  &JSConfig{Enabled: false},
		}

		client, err := NewUnifiedClient(config, slog.Default())
		require.NoError(t, err)
		defer client.Close()

		ctx := context.Background()
		response, err := client.Get(ctx, "https://httpbin.org/get")

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, 200, response.StatusCode())
		assert.Contains(t, response.String(), "test-agent")
	*/
}

func TestUnifiedClient_Close(t *testing.T) {
	config := &UnifiedConfig{
		UserAgent: "test-agent",
		JSConfig:  &JSConfig{Enabled: false},
	}

	client, err := NewUnifiedClient(config, slog.Default())
	require.NoError(t, err)

	// Should not panic or error
	err = client.Close()
	assert.NoError(t, err)
}

func TestHTTPResponseWrapper(t *testing.T) {
	// Use a simple test endpoint (this would be mocked in a real test)
	t.Skip("Requires external HTTP request")

	// Example of what this test would look like:
	/*
		ctx := context.Background()
		response, err := httpClient.Get(ctx, "https://httpbin.org/get")
		require.NoError(t, err)

		wrapper := &HTTPResponseWrapper{response: response}

		assert.Equal(t, 200, wrapper.StatusCode())
		assert.NotEmpty(t, wrapper.String())
		assert.Contains(t, wrapper.String(), "httpbin.org")
	*/
}

func TestUnifiedClient_GetWithFallback_NoJS(t *testing.T) {
	t.Skip("Requires external HTTP request or mocking")

	// Example of what this test would look like with proper mocking:
	/*
		config := &UnifiedConfig{
			UserAgent: "test-agent",
			JSConfig:  &JSConfig{Enabled: false},
		}

		client, err := NewUnifiedClient(config, slog.Default())
		require.NoError(t, err)
		defer client.Close()

		// With JS disabled, GetWithFallback should behave like regular Get
		ctx := context.Background()
		response, err := client.GetWithFallback(ctx, "https://httpbin.org/get")

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, 200, response.StatusCode())
	*/
}

func TestUnifiedResponse_Interface(t *testing.T) {
	// Test that JSResponse implements UnifiedResponse interface
	jsResponse := &JSResponse{
		URL:     "https://example.com",
		Content: "<html><body>Test</body></html>",
		Status:  200,
		Headers: make(map[string]string),
		Host:    "example.com",
	}

	var response UnifiedResponse = jsResponse

	assert.Equal(t, "<html><body>Test</body></html>", response.String())
	assert.Equal(t, 200, response.StatusCode())

	// Test that HTTPResponseWrapper would implement UnifiedResponse interface
	// (We can't create a real one without HTTP calls, but we can test the interface)
	assert.Implements(t, (*UnifiedResponse)(nil), &HTTPResponseWrapper{})
}
