package config

import (
	"log/slog"
	"os"
)

// LoggingConfig holds the configuration for the logging system
type LoggingConfig struct {
	Level   slog.Level
	Verbose bool
}

// NewLoggingConfig creates a new logging configuration
func NewLoggingConfig(verbose bool) *LoggingConfig {
	level := slog.LevelWarn // Default level: only warnings and errors
	if verbose {
		level = slog.LevelInfo // Verbose mode: include progress information
	}

	return &LoggingConfig{
		Level:   level,
		Verbose: verbose,
	}
}

// SetupLogger configures the global logger with the given configuration
func (c *LoggingConfig) SetupLogger() {
	// Create a text handler that outputs to stderr
	opts := &slog.HandlerOptions{
		Level: c.Level,
	}

	handler := slog.NewTextHandler(os.Stderr, opts)
	logger := slog.New(handler)

	// Set as the default logger
	slog.SetDefault(logger)
}

// LogCrawlStart logs the start of a crawl operation with structured data
func LogCrawlStart(url string, maxDepth int, concurrent int, userAgent string) {
	slog.Info("Starting crawl",
		"url", url,
		"maxDepth", maxDepth,
		"concurrent", concurrent,
		"userAgent", userAgent,
	)
}

// LogCrawlProgress logs crawl progress information
func LogCrawlProgress(url string, depth int, statusCode int) {
	slog.Info("Crawling URL",
		"url", url,
		"depth", depth,
		"statusCode", statusCode,
	)
}

// LogCrawlError logs crawl errors with structured context
func LogCrawlError(url string, depth int, err error) {
	slog.Warn("Failed to fetch URL",
		"url", url,
		"depth", depth,
		"error", err.Error(),
	)
}

// LogCrawlComplete logs the completion of a crawl operation
func LogCrawlComplete(url string, totalPages int, errors int) {
	slog.Info("Crawl completed",
		"startUrl", url,
		"totalPages", totalPages,
		"errors", errors,
	)
}

// LogDebug logs debug information (for future development use)
func LogDebug(message string, args ...any) {
	slog.Debug(message, args...)
}

// LogInfo logs information messages
func LogInfo(message string, args ...any) {
	slog.Info(message, args...)
}

// LogWarn logs warning messages
func LogWarn(message string, args ...any) {
	slog.Warn(message, args...)
}

// LogError logs error messages
func LogError(message string, args ...any) {
	slog.Error(message, args...)
}
