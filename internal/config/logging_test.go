package config

import (
	"bytes"
	"log/slog"
	"testing"
)

func TestNewLoggingConfig(t *testing.T) {
	tests := []struct {
		name     string
		verbose  bool
		expected slog.Level
	}{
		{
			name:     "Default (non-verbose) mode",
			verbose:  false,
			expected: slog.LevelWarn,
		},
		{
			name:     "Verbose mode",
			verbose:  true,
			expected: slog.LevelInfo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewLoggingConfig(tt.verbose)

			if config.Level != tt.expected {
				t.Errorf("NewLoggingConfig(%v) level = %v, want %v", tt.verbose, config.Level, tt.expected)
			}

			if config.Verbose != tt.verbose {
				t.Errorf("NewLoggingConfig(%v) verbose = %v, want %v", tt.verbose, config.Verbose, tt.verbose)
			}
		})
	}
}

func TestSetupLogger(t *testing.T) {
	// Test that SetupLogger creates a logger with correct configuration
	config := NewLoggingConfig(true)

	// Capture the original default logger
	originalLogger := slog.Default()
	defer slog.SetDefault(originalLogger)

	// Setup logger
	config.SetupLogger()

	// Verify that the default logger was set
	newLogger := slog.Default()
	if newLogger == originalLogger {
		t.Error("SetupLogger() did not set a new default logger")
	}
}

func TestLoggingLevels(t *testing.T) {
	tests := []struct {
		name         string
		verbose      bool
		logFunc      func()
		expectOutput bool
	}{
		{
			name:         "Warn level in non-verbose mode - should show warning",
			verbose:      false,
			logFunc:      func() { LogWarn("test warning") },
			expectOutput: true,
		},
		{
			name:         "Info level in non-verbose mode - should not show info",
			verbose:      false,
			logFunc:      func() { LogInfo("test info") },
			expectOutput: false,
		},
		{
			name:         "Info level in verbose mode - should show info",
			verbose:      true,
			logFunc:      func() { LogInfo("test info") },
			expectOutput: true,
		},
		{
			name:         "Error level should always show in non-verbose mode",
			verbose:      false,
			logFunc:      func() { LogError("test error") },
			expectOutput: true,
		},
		{
			name:         "Error level should always show in verbose mode",
			verbose:      true,
			logFunc:      func() { LogError("test error") },
			expectOutput: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a buffer to capture log output
			var buf bytes.Buffer

			// Create logging config based on verbose flag
			config := NewLoggingConfig(tt.verbose)

			// Create a logger that writes to our buffer using the correct level
			handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
				Level: config.Level,
			})
			logger := slog.New(handler)

			// Set as default logger
			originalLogger := slog.Default()
			slog.SetDefault(logger)
			defer slog.SetDefault(originalLogger)

			// Call the log function
			tt.logFunc()

			// Check if output was produced as expected
			gotOutput := buf.Len() > 0
			if gotOutput != tt.expectOutput {
				t.Errorf("Expected output: %v, got output: %v, buffer content: %q", tt.expectOutput, gotOutput, buf.String())
			}
		})
	}
}

func TestStructuredLoggingFunctions(t *testing.T) {
	// Test that structured logging functions work without panicking
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug, // Capture all log levels
	})
	logger := slog.New(handler)

	originalLogger := slog.Default()
	slog.SetDefault(logger)
	defer slog.SetDefault(originalLogger)

	// Test all structured logging functions
	LogCrawlStart("https://example.com", 3, 10, "test-agent")
	LogCrawlProgress("https://example.com/page", 1, 200)
	LogCrawlError("https://example.com/error", 2, &testError{"test error"})
	LogCrawlComplete("https://example.com", 5, 1)
	LogDebug("debug message", "key", "value")
	LogInfo("info message", "key", "value")
	LogWarn("warn message", "key", "value")
	LogError("error message", "key", "value")

	// Check that output was produced
	if buf.Len() == 0 {
		t.Error("No log output was produced by structured logging functions")
	}

	// Verify that the output contains expected structured data
	output := buf.String()
	expectedStrings := []string{
		"Starting crawl",
		"url=https://example.com",
		"maxDepth=3",
		"Crawling URL",
		"Failed to fetch URL",
		"Crawl completed",
	}

	for _, expected := range expectedStrings {
		if !bytes.Contains([]byte(output), []byte(expected)) {
			t.Errorf("Expected output to contain %q, but it didn't. Output: %s", expected, output)
		}
	}
}

// testError is a simple error type for testing
type testError struct {
	message string
}

func (e *testError) Error() string {
	return e.message
}
