package client

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/playwright-community/playwright-go"
)

// TestError wraps an error with additional context for better debugging
type TestError struct {
	Err            error
	BrowserLogs    []string
	ConsoleLogs    []string
	NetworkLogs    []string
	ScreenshotPath string
}

func (e *TestError) Error() string {
	msg := fmt.Sprintf("Test failed: %v", e.Err)
	if len(e.BrowserLogs) > 0 {
		msg += fmt.Sprintf("\nBrowser logs: %v", e.BrowserLogs)
	}
	if len(e.ConsoleLogs) > 0 {
		msg += fmt.Sprintf("\nConsole logs: %v", e.ConsoleLogs)
	}
	if len(e.NetworkLogs) > 0 {
		msg += fmt.Sprintf("\nNetwork logs: %v", e.NetworkLogs)
	}
	if e.ScreenshotPath != "" {
		msg += fmt.Sprintf("\nScreenshot saved to: %s", e.ScreenshotPath)
	}
	return msg
}

// CapturePageDebugInfo captures debug information from a Playwright page
func CapturePageDebugInfo(t *testing.T, page playwright.Page, testName string) *TestError {
	debugInfo := &TestError{}

	// Capture console logs
	page.OnConsole(func(msg playwright.ConsoleMessage) {
		debugInfo.ConsoleLogs = append(debugInfo.ConsoleLogs,
			fmt.Sprintf("[%s] %s", msg.Type(), msg.Text()))
	})

	// Capture network requests
	page.OnRequest(func(req playwright.Request) {
		debugInfo.NetworkLogs = append(debugInfo.NetworkLogs,
			fmt.Sprintf("Request: %s %s", req.Method(), req.URL()))
	})

	page.OnResponse(func(resp playwright.Response) {
		debugInfo.NetworkLogs = append(debugInfo.NetworkLogs,
			fmt.Sprintf("Response: %d %s", resp.Status(), resp.URL()))
	})

	// Capture screenshot on failure
	if t.Failed() {
		screenshotDir := filepath.Join(os.TempDir(), "urlmap-test-screenshots")
		if err := os.MkdirAll(screenshotDir, 0755); err == nil {
			screenshotPath := filepath.Join(screenshotDir, fmt.Sprintf("%s.png", testName))
			if _, err := page.Screenshot(playwright.PageScreenshotOptions{
				Path:     playwright.String(screenshotPath),
				FullPage: playwright.Bool(true),
			}); err == nil {
				debugInfo.ScreenshotPath = screenshotPath
			}
		}
	}

	return debugInfo
}

// SetupPageDebugHandlers sets up debug handlers for a page
func SetupPageDebugHandlers(page playwright.Page) (consoleLogs []string, networkLogs []string) {
	// Capture console logs
	page.OnConsole(func(msg playwright.ConsoleMessage) {
		consoleLogs = append(consoleLogs,
			fmt.Sprintf("[%s] %s", msg.Type(), msg.Text()))
	})

	// Capture network activity
	page.OnRequest(func(req playwright.Request) {
		networkLogs = append(networkLogs,
			fmt.Sprintf("Request: %s %s", req.Method(), req.URL()))
	})

	page.OnResponse(func(resp playwright.Response) {
		networkLogs = append(networkLogs,
			fmt.Sprintf("Response: %d %s", resp.Status(), resp.URL()))
	})

	return consoleLogs, networkLogs
}

// LogTestDebugInfo logs debug information when a test fails
func LogTestDebugInfo(t *testing.T, testName string, consoleLogs, networkLogs []string, err error) {
	if err == nil {
		return
	}

	t.Errorf("Test %s failed: %v", testName, err)

	if len(consoleLogs) > 0 {
		t.Logf("Console logs:")
		for _, log := range consoleLogs {
			t.Logf("  %s", log)
		}
	}

	if len(networkLogs) > 0 {
		t.Logf("Network activity:")
		for _, log := range networkLogs {
			t.Logf("  %s", log)
		}
	}
}

// CaptureScreenshotOnFailure captures a screenshot if the test has failed
func CaptureScreenshotOnFailure(t *testing.T, page playwright.Page, testName string) string {
	if !t.Failed() {
		return ""
	}

	screenshotDir := filepath.Join(os.TempDir(), "urlmap-test-screenshots")
	if err := os.MkdirAll(screenshotDir, 0755); err != nil {
		t.Logf("Failed to create screenshot directory: %v", err)
		return ""
	}

	screenshotPath := filepath.Join(screenshotDir, fmt.Sprintf("%s.png", testName))
	if _, err := page.Screenshot(playwright.PageScreenshotOptions{
		Path:     playwright.String(screenshotPath),
		FullPage: playwright.Bool(true),
	}); err != nil {
		t.Logf("Failed to capture screenshot: %v", err)
		return ""
	}

	t.Logf("Screenshot saved to: %s", screenshotPath)
	return screenshotPath
}

// RetryTest runs a test function with retry logic for potentially flaky tests
func RetryTest(t *testing.T, maxAttempts int, testFunc func() error) {
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		lastErr = testFunc()
		if lastErr == nil {
			return // Test passed
		}

		if attempt < maxAttempts {
			t.Logf("Test attempt %d/%d failed: %v. Retrying...", attempt, maxAttempts, lastErr)
			// Small delay between retries
			time.Sleep(time.Second)
		}
	}

	// All attempts failed
	t.Errorf("Test failed after %d attempts. Last error: %v", maxAttempts, lastErr)
}

// time import for RetryTest
func init() {
	// Ensure time is imported
	_ = time.Second
}

