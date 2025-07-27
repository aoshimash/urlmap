package client

import (
	"time"

	"github.com/playwright-community/playwright-go"
)

// TimeoutStrategy defines different timeout settings for various operations
type TimeoutStrategy struct {
	Navigation time.Duration
	Script     time.Duration
	Resource   time.Duration
}

// DefaultTimeoutStrategy returns the default timeout strategy
func DefaultTimeoutStrategy() TimeoutStrategy {
	return TimeoutStrategy{
		Navigation: 30 * time.Second,
		Script:     10 * time.Second,
		Resource:   5 * time.Second,
	}
}

// OptimizePage applies performance optimizations to a page
func OptimizePage(page playwright.Page) error {
	// Block unnecessary resources to improve performance
	err := page.Route("**/*", func(route playwright.Route) {
		resourceType := route.Request().ResourceType()
		switch resourceType {
		case "image", "font", "media", "manifest", "other":
			// Block these resource types
			route.Abort()
			return
		case "stylesheet":
			// Optionally block CSS if not needed for link extraction
			// For now, we'll allow it as some sites may use CSS for layout
			route.Continue()
			return
		default:
			// Allow script, document, xhr, fetch, etc.
			route.Continue()
		}
	})
	if err != nil {
		return err
	}

	// Set viewport to a reasonable size to reduce rendering overhead
	err = page.SetViewportSize(1280, 720)
	if err != nil {
		return err
	}

	// Add initialization script to optimize JavaScript execution
	scriptContent := `
		// Disable webdriver detection
		Object.defineProperty(navigator, 'webdriver', {
			get: () => false,
		});
		
		// Override permissions API if it exists
		if (window.navigator.permissions && window.navigator.permissions.query) {
			const originalQuery = window.navigator.permissions.query;
			window.navigator.permissions.query = (parameters) => (
				parameters.name === 'notifications' ?
					Promise.resolve({ state: Notification.permission }) :
					originalQuery(parameters)
			);
		}
		
		// Reduce animation overhead
		if (window.CSS && CSS.supports && CSS.supports('animation', 'none')) {
			const style = document.createElement('style');
			style.textContent = '*, *::before, *::after { animation-duration: 0s !important; animation-delay: 0s !important; transition-duration: 0s !important; transition-delay: 0s !important; }';
			document.head.appendChild(style);
		}
	`
	err = page.AddInitScript(playwright.Script{Content: &scriptContent})
	if err != nil {
		return err
	}

	return nil
}

// ApplyTimeoutStrategy applies timeout settings to a page
func ApplyTimeoutStrategy(page playwright.Page, strategy TimeoutStrategy) {
	// Set navigation timeout
	page.SetDefaultNavigationTimeout(float64(strategy.Navigation.Milliseconds()))

	// Set general timeout for other operations
	page.SetDefaultTimeout(float64(strategy.Script.Milliseconds()))
}

// GetOptimizedBrowserOptions returns optimized browser launch options
func GetOptimizedBrowserOptions(headless bool) playwright.BrowserTypeLaunchOptions {
	return playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(headless),
		Args: []string{
			"--disable-blink-features=AutomationControlled",
			"--disable-dev-shm-usage",
			"--no-sandbox",
			"--disable-setuid-sandbox",
			"--disable-gpu",
			"--disable-accelerated-2d-canvas",
			"--disable-background-timer-throttling",
			"--disable-backgrounding-occluded-windows",
			"--disable-renderer-backgrounding",
			"--disable-features=TranslateUI",
			"--disable-ipc-flooding-protection",
		},
	}
}

// GetOptimizedContextOptions returns optimized browser context options
func GetOptimizedContextOptions(userAgent string) playwright.BrowserNewContextOptions {
	options := playwright.BrowserNewContextOptions{
		UserAgent:         playwright.String(userAgent),
		JavaScriptEnabled: playwright.Bool(true),
		IgnoreHttpsErrors: playwright.Bool(true),
		// Set a generic locale to avoid locale-specific loading
		Locale: playwright.String("en-US"),
		// Set timezone to avoid timezone detection
		TimezoneId: playwright.String("UTC"),
	}

	return options
}
