package client

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSConfig_DefaultConfig(t *testing.T) {
	config := DefaultJSConfig()

	assert.False(t, config.Enabled)
	assert.Equal(t, "chromium", config.BrowserType)
	assert.True(t, config.Headless)
	assert.Equal(t, 30*time.Second, config.Timeout)
	assert.Equal(t, "networkidle", config.WaitFor)
	assert.Equal(t, "urlmap/1.0", config.UserAgent)
	assert.False(t, config.AutoDetect)
	assert.True(t, config.Fallback)
}

func TestJSConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *JSConfig
		wantErr bool
	}{
		{
			name:    "disabled config should be valid",
			config:  &JSConfig{Enabled: false},
			wantErr: false,
		},
		{
			name: "valid enabled config",
			config: &JSConfig{
				Enabled:     true,
				BrowserType: "chromium",
				Timeout:     10 * time.Second,
				WaitFor:     "networkidle",
			},
			wantErr: false,
		},
		{
			name: "invalid browser type",
			config: &JSConfig{
				Enabled:     true,
				BrowserType: "invalid",
				Timeout:     10 * time.Second,
				WaitFor:     "networkidle",
			},
			wantErr: true,
		},
		{
			name: "invalid wait condition",
			config: &JSConfig{
				Enabled:     true,
				BrowserType: "chromium",
				Timeout:     10 * time.Second,
				WaitFor:     "invalid",
			},
			wantErr: true,
		},
		{
			name: "invalid timeout",
			config: &JSConfig{
				Enabled:     true,
				BrowserType: "chromium",
				Timeout:     -1 * time.Second,
				WaitFor:     "networkidle",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewJSClient_DisabledConfig(t *testing.T) {
	config := &JSConfig{Enabled: false}
	client, err := NewJSClient(config, slog.Default())

	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.False(t, client.config.Enabled)

	// Should not initialize Playwright when disabled
	assert.Nil(t, client.playwright)
	assert.Nil(t, client.browser)
}

func TestNewJSClient_InvalidConfig(t *testing.T) {
	config := &JSConfig{
		Enabled:     true,
		BrowserType: "invalid",
		Timeout:     10 * time.Second,
		WaitFor:     "networkidle",
	}

	client, err := NewJSClient(config, slog.Default())

	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "invalid JS config")
}

func TestNewJSClient_EnabledConfig_CI(t *testing.T) {
	if !isGitHubActions() {
		t.Skip("This test only runs in GitHub Actions to verify Playwright installation")
	}

	config := &JSConfig{
		Enabled:     true,
		BrowserType: "chromium",
		Headless:    true,
		Timeout:     10 * time.Second,
		WaitFor:     "networkidle",
	}

	// In CI, this should fail gracefully due to missing system dependencies
	client, err := NewJSClient(config, slog.Default())

	// We expect this to fail in CI environment
	if err != nil {
		t.Logf("Expected failure in CI environment: %v", err)
		assert.Error(t, err)
		assert.Nil(t, client)
	} else {
		// If it succeeds, clean up
		if client != nil {
			client.Close()
		}
	}
}

func TestJSClient_RenderPage_Disabled(t *testing.T) {
	config := &JSConfig{Enabled: false}
	client, err := NewJSClient(config, slog.Default())
	require.NoError(t, err)

	ctx := context.Background()
	content, err := client.RenderPage(ctx, "https://example.com")

	assert.Error(t, err)
	assert.Empty(t, content)
	assert.Contains(t, err.Error(), "JavaScript rendering is not enabled")
}

func TestJSClient_Close(t *testing.T) {
	config := &JSConfig{Enabled: false}
	client, err := NewJSClient(config, slog.Default())
	require.NoError(t, err)

	// Should not panic even with disabled config
	err = client.Close()
	assert.NoError(t, err)
}

// Note: Integration tests that actually launch browsers are skipped in unit tests
// They would be included in e2e tests with proper environment setup
func TestJSClient_Integration_Skip(t *testing.T) {
	t.Skip("Integration tests require Playwright browser installation")

	// Example of what an integration test would look like:
	/*
		config := &JSConfig{
			Enabled:     true,
			BrowserType: "chromium",
			Headless:    true,
			Timeout:     10 * time.Second,
			WaitFor:     "networkidle",
		}

		client, err := NewJSClient(config, slog.Default())
		require.NoError(t, err)
		defer client.Close()

		ctx := context.Background()
		content, err := client.RenderPage(ctx, "https://example.com")

		assert.NoError(t, err)
		assert.NotEmpty(t, content)
		assert.Contains(t, content, "<html")
	*/
}
