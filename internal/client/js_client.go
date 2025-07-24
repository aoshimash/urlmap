package client

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"

	"github.com/playwright-community/playwright-go"
)

// JSClient provides JavaScript rendering capabilities using Playwright
type JSClient struct {
	playwright *playwright.Playwright
	browser    playwright.Browser
	config     *JSConfig
	logger     *slog.Logger
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

	client := &JSClient{
		config: config,
		logger: logger,
	}

	// Initialize Playwright if JS rendering is enabled
	if config.Enabled {
		if err := client.initPlaywright(); err != nil {
			return nil, fmt.Errorf("failed to initialize Playwright: %w", err)
		}
	}

	return client, nil
}

// initPlaywright initializes the Playwright browser instance
func (c *JSClient) initPlaywright() error {
	c.logger.Debug("Initializing Playwright", "browser_type", c.config.BrowserType)

	// Install Playwright (this will be a no-op if already installed)
	err := playwright.Install()
	if err != nil {
		return fmt.Errorf("failed to install Playwright: %w", err)
	}

	// Run Playwright
	pw, err := playwright.Run()
	if err != nil {
		return fmt.Errorf("failed to run Playwright: %w", err)
	}
	c.playwright = pw

	// Launch browser
	var browserType playwright.BrowserType
	switch c.config.BrowserType {
	case "chromium":
		browserType = c.playwright.Chromium
	case "firefox":
		browserType = c.playwright.Firefox
	case "webkit":
		browserType = c.playwright.WebKit
	default:
		return fmt.Errorf("unsupported browser type: %s", c.config.BrowserType)
	}

	browser, err := browserType.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(c.config.Headless),
	})
	if err != nil {
		return fmt.Errorf("failed to launch browser: %w", err)
	}
	c.browser = browser

	c.logger.Info("Playwright initialized successfully",
		"browser_type", c.config.BrowserType,
		"headless", c.config.Headless)

	return nil
}

// RenderPage renders a page with JavaScript and returns the final HTML content
func (c *JSClient) RenderPage(ctx context.Context, targetURL string) (string, error) {
	if !c.config.Enabled {
		return "", fmt.Errorf("JavaScript rendering is not enabled")
	}

	if c.browser == nil {
		return "", fmt.Errorf("browser not initialized")
	}

	c.logger.Debug("Starting JavaScript rendering", "url", targetURL)

	// Create a new browser context
	context, err := c.browser.NewContext(playwright.BrowserNewContextOptions{
		UserAgent: playwright.String(c.config.UserAgent),
	})
	if err != nil {
		return "", fmt.Errorf("failed to create browser context: %w", err)
	}
	defer context.Close()

	// Create a new page
	page, err := context.NewPage()
	if err != nil {
		return "", fmt.Errorf("failed to create page: %w", err)
	}
	defer page.Close()

	// Set timeout
	page.SetDefaultTimeout(float64(c.config.Timeout.Milliseconds()))

	// Navigate to the URL
	var waitUntil *playwright.WaitUntilState
	switch c.config.WaitFor {
	case "networkidle":
		waitUntil = playwright.WaitUntilStateNetworkidle
	case "domcontentloaded":
		waitUntil = playwright.WaitUntilStateDomcontentloaded
	case "load":
		waitUntil = playwright.WaitUntilStateLoad
	default:
		waitUntil = playwright.WaitUntilStateNetworkidle
	}

	_, err = page.Goto(targetURL, playwright.PageGotoOptions{
		WaitUntil: waitUntil,
		Timeout:   playwright.Float(float64(c.config.Timeout.Milliseconds())),
	})
	if err != nil {
		return "", fmt.Errorf("failed to navigate to URL %s: %w", targetURL, err)
	}

	// Get the final HTML content
	content, err := page.Content()
	if err != nil {
		return "", fmt.Errorf("failed to get page content: %w", err)
	}

	c.logger.Debug("JavaScript rendering completed",
		"url", targetURL,
		"content_length", len(content))

	return content, nil
}

// Get implements a similar interface to the HTTP client for compatibility
func (c *JSClient) Get(ctx context.Context, targetURL string) (*JSResponse, error) {
	content, err := c.RenderPage(ctx, targetURL)
	if err != nil {
		return nil, err
	}

	// Parse URL for response metadata
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	return &JSResponse{
		URL:     targetURL,
		Content: content,
		Status:  200, // Assume success if we got content
		Headers: make(map[string]string),
		Host:    parsedURL.Host,
	}, nil
}

// Close cleans up the JavaScript client resources
func (c *JSClient) Close() error {
	if c.browser != nil {
		if err := c.browser.Close(); err != nil {
			c.logger.Warn("Failed to close browser", "error", err)
		}
		c.browser = nil
	}

	if c.playwright != nil {
		if err := c.playwright.Stop(); err != nil {
			c.logger.Warn("Failed to stop Playwright", "error", err)
		}
	}

	c.logger.Debug("JavaScript client closed")
	return nil
}

// JSResponse represents a response from JavaScript rendering
type JSResponse struct {
	URL     string
	Content string
	Status  int
	Headers map[string]string
	Host    string
}

// String returns the rendered HTML content
func (r *JSResponse) String() string {
	return r.Content
}

// StatusCode returns the HTTP status code (always 200 for successful JS rendering)
func (r *JSResponse) StatusCode() int {
	return r.Status
}
