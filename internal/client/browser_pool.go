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

// BrowserPool manages shared browser instances for JavaScript rendering
type BrowserPool struct {
	playwright *playwright.Playwright
	browser    playwright.Browser
	config     *JSConfig
	logger     *slog.Logger

	// Pool management
	contextPool chan *BrowserContext
	maxContexts int
	mu          sync.RWMutex

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

	pool := &BrowserPool{
		config:      config,
		logger:      logger,
		maxContexts: 10, // Default max contexts
		contextPool: make(chan *BrowserContext, 10),
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

	p.logger.Debug("Initializing browser pool", "browser_type", p.config.BrowserType)

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
	p.playwright = pw

	// Launch browser
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

	browser, err := browserType.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(p.config.Headless),
	})
	if err != nil {
		return fmt.Errorf("failed to launch browser: %w", err)
	}
	p.browser = browser

	p.initialized = true
	p.logger.Info("Browser pool initialized successfully",
		"browser_type", p.config.BrowserType,
		"headless", p.config.Headless,
		"max_contexts", p.maxContexts)

	return nil
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
	if p.browser == nil {
		return nil, fmt.Errorf("browser not initialized")
	}

	context, err := p.browser.NewContext(playwright.BrowserNewContextOptions{
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
		"initialized":  p.initialized,
		"closed":       p.closed,
		"pool_size":    cap(p.contextPool),
		"max_contexts": p.maxContexts,
		"available":    len(p.contextPool),
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

	// Close browser
	if p.browser != nil {
		if err := p.browser.Close(); err != nil {
			p.logger.Warn("Failed to close browser", "error", err)
		}
		p.browser = nil
	}

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
