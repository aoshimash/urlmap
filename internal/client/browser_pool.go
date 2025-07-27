package client

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/playwright-community/playwright-go"
)

// Global initialization lock to prevent concurrent Playwright installations
var globalInitMu sync.Mutex

// BrowserPool manages shared browser instances for JavaScript rendering
type BrowserPool struct {
	playwright *playwright.Playwright
	browsers   []playwright.Browser // Multiple browser instances
	config     *JSConfig
	logger     *slog.Logger

	// Pool management
	contextPool chan *BrowserContext
	maxContexts int
	mu          sync.RWMutex

	// Browser instance management
	currentBrowserIdx int
	browserMu         sync.Mutex

	// Lifecycle management
	initialized bool
	closed      bool
	closeMu     sync.Mutex
}

// BrowserContext wraps a Playwright browser context with additional metadata
type BrowserContext struct {
	Context   playwright.BrowserContext
	CreatedAt time.Time
	LastUsed  time.Time
	UseCount  int
	Pool      *BrowserPool
}

// NewBrowserPool creates a new browser pool with the given configuration
func NewBrowserPool(config *JSConfig, logger *slog.Logger) (*BrowserPool, error) {
	if config == nil {
		config = DefaultJSConfig()
	}

	if logger == nil {
		logger = slog.Default()
	}

	poolSize := config.PoolSize
	if poolSize <= 0 {
		poolSize = 2 // Default pool size
	}

	pool := &BrowserPool{
		config:            config,
		logger:            logger,
		maxContexts:       10, // Default max contexts per browser
		contextPool:       make(chan *BrowserContext, 10*poolSize),
		browsers:          make([]playwright.Browser, 0, poolSize),
		currentBrowserIdx: 0,
	}

	// Initialize the pool if JS rendering is enabled
	if config.Enabled {
		if err := pool.initialize(); err != nil {
			return nil, fmt.Errorf("failed to initialize browser pool: %w", err)
		}
	}

	return pool, nil
}

// initialize sets up the Playwright browser instance
func (p *BrowserPool) initialize() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.initialized {
		return nil
	}

	// Use global lock for Playwright initialization
	globalInitMu.Lock()
	defer globalInitMu.Unlock()

	p.logger.Debug("Initializing browser pool", "browser_type", p.config.BrowserType)

	// Install Playwright (this will be a no-op if already installed)
	p.logger.Debug("Installing Playwright browsers...")
	err := playwright.Install()
	if err != nil {
		return fmt.Errorf("failed to install Playwright: %w", err)
	}
	p.logger.Debug("Playwright browsers installed")

	// Run Playwright
	pw, err := playwright.Run()
	if err != nil {
		return fmt.Errorf("failed to run Playwright: %w", err)
	}
	p.playwright = pw

	// Get browser type
	var browserType playwright.BrowserType
	switch p.config.BrowserType {
	case "chromium":
		browserType = p.playwright.Chromium
	case "firefox":
		browserType = p.playwright.Firefox
	case "webkit":
		browserType = p.playwright.WebKit
	default:
		return fmt.Errorf("unsupported browser type: %s", p.config.BrowserType)
	}

	// Launch the first browser instance
	browser, err := browserType.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(p.config.Headless),
	})
	if err != nil {
		return fmt.Errorf("failed to launch first browser: %w", err)
	}
	p.browsers = append(p.browsers, browser)

	p.initialized = true
	p.logger.Info("Browser pool initialized successfully",
		"browser_type", p.config.BrowserType,
		"headless", p.config.Headless,
		"pool_size", len(p.browsers),
		"max_pool_size", cap(p.browsers),
		"max_contexts_per_browser", p.maxContexts)

	return nil
}

// launchNewBrowserLocked creates and launches a new browser instance (must be called with browserMu held)
func (p *BrowserPool) launchNewBrowserLocked() (int, error) {

	// Check if we've reached the pool size limit
	if len(p.browsers) >= cap(p.browsers) {
		return -1, fmt.Errorf("browser pool is at capacity: %d", cap(p.browsers))
	}

	// Get browser type
	var browserType playwright.BrowserType
	switch p.config.BrowserType {
	case "chromium":
		browserType = p.playwright.Chromium
	case "firefox":
		browserType = p.playwright.Firefox
	case "webkit":
		browserType = p.playwright.WebKit
	default:
		return -1, fmt.Errorf("unsupported browser type: %s", p.config.BrowserType)
	}

	// Launch new browser
	browser, err := browserType.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(p.config.Headless),
	})
	if err != nil {
		return -1, fmt.Errorf("failed to launch browser: %w", err)
	}

	// Add to browsers slice
	idx := len(p.browsers)
	p.browsers = append(p.browsers, browser)

	p.logger.Info("Launched new browser instance",
		"browser_index", idx,
		"total_browsers", len(p.browsers))

	return idx, nil
}

// getBrowser gets a browser using round-robin
func (p *BrowserPool) getBrowser() (playwright.Browser, error) {
	p.browserMu.Lock()
	defer p.browserMu.Unlock()

	// If no browsers exist, create the first one
	if len(p.browsers) == 0 {
		return nil, fmt.Errorf("no browsers initialized")
	}

	// Create more browsers if needed and under limit
	if len(p.browsers) < cap(p.browsers) {
		_, err := p.launchNewBrowserLocked()
		if err != nil {
			// Log but don't fail - use existing browsers
			p.logger.Debug("Failed to launch additional browser", "error", err)
		}
	}

	// Use round-robin to distribute load
	browser := p.browsers[p.currentBrowserIdx]
	p.currentBrowserIdx = (p.currentBrowserIdx + 1) % len(p.browsers)

	return browser, nil
}

// AcquireContext gets a browser context from the pool
func (p *BrowserPool) AcquireContext() (*BrowserContext, error) {
	p.closeMu.Lock()
	if p.closed {
		p.closeMu.Unlock()
		return nil, fmt.Errorf("browser pool is closed")
	}
	p.closeMu.Unlock()

	// Try to get an existing context from the pool
	select {
	case ctx := <-p.contextPool:
		ctx.LastUsed = time.Now()
		ctx.UseCount++
		p.logger.Debug("Reused browser context", "use_count", ctx.UseCount)
		return ctx, nil
	default:
		// Create a new context if pool is empty
		return p.createNewContext()
	}
}

// ReleaseContext returns a browser context to the pool
func (p *BrowserContext) ReleaseContext() {
	if p.Pool == nil {
		return
	}

	// Check if pool is still open
	p.Pool.closeMu.Lock()
	if p.Pool.closed {
		p.Pool.closeMu.Unlock()
		// Pool is closed, close the context
		p.Context.Close()
		return
	}
	p.Pool.closeMu.Unlock()

	// Try to return to pool, but don't block if pool is full
	select {
	case p.Pool.contextPool <- p:
		p.Pool.logger.Debug("Returned browser context to pool")
	default:
		// Pool is full, close the context
		p.Pool.logger.Debug("Pool full, closing browser context")
		p.Context.Close()
	}
}

// createNewContext creates a new browser context
func (p *BrowserPool) createNewContext() (*BrowserContext, error) {
	// Get a browser from the pool
	browser, err := p.getBrowser()
	if err != nil {
		return nil, fmt.Errorf("failed to get browser from pool: %w", err)
	}

	context, err := browser.NewContext(playwright.BrowserNewContextOptions{
		UserAgent: playwright.String(p.config.UserAgent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create browser context: %w", err)
	}

	ctx := &BrowserContext{
		Context:   context,
		CreatedAt: time.Now(),
		LastUsed:  time.Now(),
		UseCount:  1,
		Pool:      p,
	}

	p.logger.Debug("Created new browser context")
	return ctx, nil
}

// RenderPage renders a page using a context from the pool
func (p *BrowserPool) RenderPage(ctx context.Context, targetURL string) (string, error) {
	if !p.config.Enabled {
		return "", fmt.Errorf("JavaScript rendering is not enabled")
	}

	browserCtx, err := p.AcquireContext()
	if err != nil {
		return "", fmt.Errorf("failed to acquire browser context: %w", err)
	}
	defer browserCtx.ReleaseContext()

	p.logger.Debug("Starting JavaScript rendering", "url", targetURL)

	// Create a new page
	page, err := browserCtx.Context.NewPage()
	if err != nil {
		return "", fmt.Errorf("failed to create page: %w", err)
	}
	defer page.Close()

	// Apply minimal performance optimizations
	// Block only the most resource-intensive content types
	if !testing.Testing() {
		page.Route("**/*", func(route playwright.Route) {
			resourceType := route.Request().ResourceType()
			switch resourceType {
			case "image", "media", "font":
				// Block only heavy resources
				route.Abort()
				return
			default:
				route.Continue()
			}
		})
	}

	// Setup debug handlers if running in test mode
	var consoleLogs, networkLogs []string
	if testing.Testing() {
		consoleLogs, networkLogs = SetupPageDebugHandlers(page)
	}

	// Set timeout
	page.SetDefaultTimeout(float64(p.config.Timeout.Milliseconds()))

	// Navigate to the URL
	var waitUntil *playwright.WaitUntilState
	switch p.config.WaitFor {
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
		Timeout:   playwright.Float(float64(p.config.Timeout.Milliseconds())),
	})
	if err != nil {
		// Log debug info when running in test mode
		if testing.Testing() && (len(consoleLogs) > 0 || len(networkLogs) > 0) {
			p.logger.Error("Navigation failed with debug info",
				"url", targetURL,
				"error", err,
				"console_logs", consoleLogs,
				"network_logs", networkLogs)
		}
		return "", fmt.Errorf("failed to navigate to URL %s: %w", targetURL, err)
	}

	// Get the final HTML content
	content, err := page.Content()
	if err != nil {
		// Log debug info when running in test mode
		if testing.Testing() && (len(consoleLogs) > 0 || len(networkLogs) > 0) {
			p.logger.Error("Failed to get content with debug info",
				"url", targetURL,
				"error", err,
				"console_logs", consoleLogs,
				"network_logs", networkLogs)
		}
		return "", fmt.Errorf("failed to get page content: %w", err)
	}

	p.logger.Debug("JavaScript rendering completed",
		"url", targetURL,
		"content_length", len(content))

	return content, nil
}

// GetPoolStats returns statistics about the browser pool
func (p *BrowserPool) GetPoolStats() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return map[string]interface{}{
		"initialized":        p.initialized,
		"closed":             p.closed,
		"browsers_active":    len(p.browsers),
		"browsers_available": len(p.browsers), // All browsers are available via round-robin
		"browser_pool_size":  cap(p.browsers),
		"context_pool_size":  cap(p.contextPool),
		"contexts_available": len(p.contextPool),
		"max_contexts":       p.maxContexts,
	}
}

// Close cleans up all browser pool resources
func (p *BrowserPool) Close() error {
	p.closeMu.Lock()
	defer p.closeMu.Unlock()

	if p.closed {
		return nil
	}

	p.logger.Debug("Closing browser pool")

	// Close all contexts in the pool
	close(p.contextPool)
	for ctx := range p.contextPool {
		if ctx.Context != nil {
			ctx.Context.Close()
		}
	}

	// Close all browsers
	var closeErrors []error
	for i, browser := range p.browsers {
		if browser != nil {
			p.logger.Debug("Closing browser", "browser_index", i)
			if err := browser.Close(); err != nil {
				p.logger.Warn("Failed to close browser", "browser_index", i, "error", err)
				closeErrors = append(closeErrors, fmt.Errorf("browser %d: %w", i, err))
			}
		}
	}
	p.browsers = nil

	// Stop Playwright
	if p.playwright != nil {
		if err := p.playwright.Stop(); err != nil {
			p.logger.Warn("Failed to stop Playwright", "error", err)
		}
		p.playwright = nil
	}

	p.closed = true
	p.logger.Info("Browser pool closed")
	return nil
}
