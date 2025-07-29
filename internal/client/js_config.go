package client

import (
	"fmt"
	"time"
)

// JSConfig holds configuration for JavaScript rendering
type JSConfig struct {
	// Enabled indicates whether JavaScript rendering is enabled
	Enabled bool

	// BrowserType specifies which browser to use (chromium, firefox, webkit)
	BrowserType string

	// Headless indicates whether to run browser in headless mode
	Headless bool

	// Timeout specifies the maximum time to wait for page load
	Timeout time.Duration

	// WaitFor specifies what to wait for before considering page loaded
	// Options: "networkidle", "domcontentloaded", "load"
	WaitFor string

	// UserAgent to use for requests
	UserAgent string

	// AutoDetect indicates whether to automatically detect SPA sites
	AutoDetect bool

	// StrictMode indicates whether to use strict detection with dynamic verification
	StrictMode bool

	// Threshold specifies the confidence threshold for SPA detection (0.0-1.0)
	Threshold float64

	// Fallback indicates whether to fallback to HTTP client on JS errors
	Fallback bool

	// PoolSize specifies the number of browser instances in the pool
	PoolSize int

	// CacheEnabled indicates whether to cache rendered pages
	CacheEnabled bool

	// CacheSize specifies the maximum number of cache entries
	CacheSize int

	// CacheTTL specifies how long cache entries remain valid
	CacheTTL time.Duration
}

// DefaultJSConfig returns a default JavaScript configuration
func DefaultJSConfig() *JSConfig {
	return &JSConfig{
		Enabled:      false,
		BrowserType:  "chromium",
		Headless:     true,
		Timeout:      30 * time.Second,
		WaitFor:      "networkidle",
		UserAgent:    "urlmap/1.0",
		AutoDetect:   false,
		StrictMode:   false,
		Threshold:    0.5,
		Fallback:     true,
		PoolSize:     2,
		CacheEnabled: false,
		CacheSize:    100,
		CacheTTL:     5 * time.Minute,
	}
}

// Validate checks if the configuration is valid
func (c *JSConfig) Validate() error {
	if !c.Enabled {
		return nil
	}

	// Validate browser type
	validBrowsers := []string{"chromium", "firefox", "webkit"}
	validBrowser := false
	for _, browser := range validBrowsers {
		if c.BrowserType == browser {
			validBrowser = true
			break
		}
	}
	if !validBrowser {
		return fmt.Errorf("invalid browser type: %s, must be one of: %v", c.BrowserType, validBrowsers)
	}

	// Validate wait condition
	validWaitConditions := []string{"networkidle", "domcontentloaded", "load"}
	validWaitCondition := false
	for _, condition := range validWaitConditions {
		if c.WaitFor == condition {
			validWaitCondition = true
			break
		}
	}
	if !validWaitCondition {
		return fmt.Errorf("invalid wait condition: %s, must be one of: %v", c.WaitFor, validWaitConditions)
	}

	// Validate timeout
	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive, got: %v", c.Timeout)
	}

	// Validate pool size
	if c.PoolSize <= 0 {
		return fmt.Errorf("pool size must be positive, got: %v", c.PoolSize)
	}

	// Validate cache configuration
	if c.CacheEnabled {
		if c.CacheSize <= 0 {
			return fmt.Errorf("cache size must be positive when cache is enabled, got: %v", c.CacheSize)
		}
		if c.CacheTTL <= 0 {
			return fmt.Errorf("cache TTL must be positive when cache is enabled, got: %v", c.CacheTTL)
		}
	}

	return nil
}
