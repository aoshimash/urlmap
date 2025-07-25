package client

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/go-resty/resty/v2"
)

// UnifiedClient provides both HTTP and JavaScript rendering capabilities
type UnifiedClient struct {
	httpClient *Client
	jsClient   *JSClient
	config     *UnifiedConfig
	logger     *slog.Logger
}

// UnifiedConfig holds configuration for the unified client
type UnifiedConfig struct {
	// HTTP client configuration
	UserAgent string

	// JavaScript client configuration
	JSConfig *JSConfig
}

// UnifiedResponse represents a response from either HTTP or JS client
type UnifiedResponse interface {
	String() string
	StatusCode() int
}

// NewUnifiedClient creates a new unified client that can use both HTTP and JS rendering
func NewUnifiedClient(config *UnifiedConfig, logger *slog.Logger) (*UnifiedClient, error) {
	if config == nil {
		config = &UnifiedConfig{
			UserAgent: "urlmap/1.0",
			JSConfig:  DefaultJSConfig(),
		}
	}

	if logger == nil {
		logger = slog.Default()
	}

	// Create HTTP client
	httpConfig := &Config{
		UserAgent: config.UserAgent,
	}
	httpClient := NewClient(httpConfig)

	// Create JS client (only if enabled)
	var jsClient *JSClient
	var err error
	if config.JSConfig != nil && config.JSConfig.Enabled {
		// Ensure UserAgent consistency
		if config.JSConfig.UserAgent == "" {
			config.JSConfig.UserAgent = config.UserAgent
		}

		jsClient, err = NewJSClient(config.JSConfig, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to create JS client: %w", err)
		}
	}

	return &UnifiedClient{
		httpClient: httpClient,
		jsClient:   jsClient,
		config:     config,
		logger:     logger,
	}, nil
}

// Get fetches content using the appropriate client (HTTP or JS)
func (c *UnifiedClient) Get(ctx context.Context, url string) (UnifiedResponse, error) {
	// If JS rendering is enabled, use JS client
	if c.jsClient != nil && c.config.JSConfig.Enabled {
		c.logger.Debug("Using JavaScript client", "url", url)
		return c.jsClient.Get(ctx, url)
	}

	// Otherwise, use HTTP client
	c.logger.Debug("Using HTTP client", "url", url)
	response, err := c.httpClient.Get(ctx, url)
	if err != nil {
		return nil, err
	}

	return &HTTPResponseWrapper{response: response}, nil
}

// GetWithFallback attempts JS rendering first, falls back to HTTP on error
func (c *UnifiedClient) GetWithFallback(ctx context.Context, url string) (UnifiedResponse, error) {
	// Try JS client first if available and fallback is enabled
	if c.jsClient != nil && c.config.JSConfig.Enabled && c.config.JSConfig.Fallback {
		c.logger.Debug("Attempting JavaScript rendering with fallback", "url", url)

		jsResponse, err := c.jsClient.Get(ctx, url)
		if err != nil {
			c.logger.Warn("JavaScript rendering failed, falling back to HTTP",
				"url", url, "error", err)

			// Fallback to HTTP client
			response, httpErr := c.httpClient.Get(ctx, url)
			if httpErr != nil {
				return nil, fmt.Errorf("both JS and HTTP clients failed - JS error: %w, HTTP error: %v", err, httpErr)
			}

			return &HTTPResponseWrapper{response: response}, nil
		}

		return jsResponse, nil
	}

	// No JS client or fallback disabled, use regular Get
	return c.Get(ctx, url)
}

// Close cleans up resources for both clients
func (c *UnifiedClient) Close() error {
	var errors []error

	if c.jsClient != nil {
		if err := c.jsClient.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close JS client: %w", err))
		}
	}

	// HTTP client doesn't need explicit closing

	if len(errors) > 0 {
		return fmt.Errorf("errors closing unified client: %v", errors)
	}

	c.logger.Debug("Unified client closed")
	return nil
}

// IsJSEnabled returns whether JavaScript rendering is enabled
func (c *UnifiedClient) IsJSEnabled() bool {
	return c.jsClient != nil && c.config.JSConfig.Enabled
}

// GetJSConfig returns the JavaScript configuration
func (c *UnifiedClient) GetJSConfig() *JSConfig {
	if c.config == nil {
		return nil
	}
	return c.config.JSConfig
}

// GetJSClient returns the JavaScript client
func (c *UnifiedClient) GetJSClient() *JSClient {
	return c.jsClient
}

// GetHTTPClient returns the HTTP client
func (c *UnifiedClient) GetHTTPClient() *Client {
	return c.httpClient
}

// HTTPResponseWrapper wraps the HTTP response to implement UnifiedResponse
type HTTPResponseWrapper struct {
	response *resty.Response
}

// String returns the response body as string
func (w *HTTPResponseWrapper) String() string {
	return w.response.String()
}

// StatusCode returns the HTTP status code
func (w *HTTPResponseWrapper) StatusCode() int {
	return w.response.StatusCode()
}
