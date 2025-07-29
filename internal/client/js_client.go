package client

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
)

// JSClient provides JavaScript rendering capabilities using a shared browser pool
type JSClient struct {
	pool   *BrowserPool
	config *JSConfig
	logger *slog.Logger
	cache  *RenderCache
}

// NewJSClient creates a new JavaScript client with the given configuration
func NewJSClient(config *JSConfig, logger *slog.Logger) (*JSClient, error) {
	if config == nil {
		config = DefaultJSConfig()
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid JS config: %w", err)
	}

	if logger == nil {
		logger = slog.Default()
	}

	// Create browser pool
	pool, err := NewBrowserPool(config, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create browser pool: %w", err)
	}

	client := &JSClient{
		pool:   pool,
		config: config,
		logger: logger,
	}

	// Create cache if enabled
	if config.CacheEnabled {
		client.cache = NewRenderCache(config.CacheSize, config.CacheTTL)
		logger.Info("JavaScript render cache enabled",
			"max_size", config.CacheSize,
			"ttl", config.CacheTTL)
	}

	return client, nil
}

// RenderPage renders a page with JavaScript and returns the final HTML content
func (c *JSClient) RenderPage(ctx context.Context, targetURL string) (string, error) {
	if !c.config.Enabled {
		return "", fmt.Errorf("JavaScript rendering is not enabled")
	}

	return c.pool.RenderPage(ctx, targetURL)
}

// Get implements a similar interface to the HTTP client for compatibility
func (c *JSClient) Get(ctx context.Context, targetURL string) (*JSResponse, error) {
	// Check cache first if enabled
	if c.cache != nil {
		if entry, hit := c.cache.Get(targetURL); hit {
			c.logger.Debug("Cache hit for URL", "url", targetURL)

			// Parse URL for response metadata
			parsedURL, err := url.Parse(targetURL)
			if err != nil {
				return nil, fmt.Errorf("failed to parse URL: %w", err)
			}

			return &JSResponse{
				URL:       targetURL,
				Content:   entry.Content,
				Status:    entry.StatusCode,
				Headers:   entry.Headers,
				Host:      parsedURL.Host,
				FromCache: true,
			}, nil
		}
		c.logger.Debug("Cache miss for URL", "url", targetURL)
	}

	// Not in cache, render the page
	content, err := c.RenderPage(ctx, targetURL)
	if err != nil {
		return nil, err
	}

	// Parse URL for response metadata
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	response := &JSResponse{
		URL:       targetURL,
		Content:   content,
		Status:    200, // Assume success if we got content
		Headers:   make(map[string]string),
		Host:      parsedURL.Host,
		FromCache: false,
	}

	// Store in cache if enabled
	if c.cache != nil {
		c.cache.Set(targetURL, content, response.Headers, response.Status)
		c.logger.Debug("Stored render result in cache", "url", targetURL)
	}

	return response, nil
}

// Close cleans up the JavaScript client resources
func (c *JSClient) Close() error {
	if c.pool != nil {
		if err := c.pool.Close(); err != nil {
			c.logger.Warn("Failed to close browser pool", "error", err)
		}
	}

	c.logger.Debug("JavaScript client closed")
	return nil
}

// GetPoolStats returns statistics about the browser pool
func (c *JSClient) GetPoolStats() map[string]interface{} {
	if c.pool == nil {
		return nil
	}
	return c.pool.GetPoolStats()
}

// GetCacheStats returns statistics about the render cache
func (c *JSClient) GetCacheStats() map[string]interface{} {
	if c.cache == nil {
		return nil
	}
	return c.cache.Stats()
}

// JSResponse represents a response from JavaScript rendering
type JSResponse struct {
	URL       string
	Content   string
	Status    int
	Headers   map[string]string
	Host      string
	FromCache bool // Indicates if this response was served from cache
}

// String returns the rendered HTML content
func (r *JSResponse) String() string {
	return r.Content
}

// StatusCode returns the HTTP status code (always 200 for successful JS rendering)
func (r *JSResponse) StatusCode() int {
	return r.Status
}
