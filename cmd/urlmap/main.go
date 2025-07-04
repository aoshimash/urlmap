package main

import (
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/aoshimash/urlmap/internal/config"
	"github.com/aoshimash/urlmap/internal/crawler"
	"github.com/aoshimash/urlmap/internal/output"
	"github.com/aoshimash/urlmap/internal/progress"
	"github.com/spf13/cobra"
)

var (
	// Build-time variables for version information
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

// Command line flags
var (
	depth        int
	verbose      bool
	userAgent    string
	concurrent   int
	showProgress bool
	rateLimit    float64
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "urlmap <URL>",
	Short: "A web crawler for mapping site URLs",
	Long: `Urlmap is a web crawler for discovering and mapping all URLs within a website.

This tool crawls web pages starting from a given URL and discovers all links
within the same domain, creating a comprehensive URL map of the site.

Examples:
  urlmap https://example.com
  urlmap -d 3 -c 5 https://example.com
  urlmap --verbose --user-agent "MyBot/1.0" https://example.com`,
	Args: cobra.ExactArgs(1), // Require exactly one URL argument
	RunE: runCrawl,
}

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("urlmap version %s\n", version)
		fmt.Printf("commit: %s\n", commit)
		fmt.Printf("built: %s\n", date)
	},
}

func init() {
	// Add flags to the root command
	rootCmd.Flags().IntVarP(&depth, "depth", "d", -1, "Maximum crawl depth (-1 = unlimited)")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")
	rootCmd.Flags().StringVarP(&userAgent, "user-agent", "u", "urlmap/0.2.0 (+https://github.com/aoshimash/urlmap)", "Custom User-Agent string")
	rootCmd.Flags().IntVarP(&concurrent, "concurrent", "c", 10, "Number of concurrent requests")
	rootCmd.Flags().BoolVarP(&showProgress, "progress", "p", true, "Show progress indicators (default: true)")
	rootCmd.Flags().Float64VarP(&rateLimit, "rate-limit", "r", 0, "Rate limit requests per second (0 = no limit)")

	// Add subcommands
	rootCmd.AddCommand(versionCmd)
}

func runCrawl(cmd *cobra.Command, args []string) error {
	// Validate URL argument
	targetURL := args[0]
	parsedURL, err := url.Parse(targetURL)
	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		return fmt.Errorf("invalid URL: %s (must be http or https)", targetURL)
	}

	// Set up logging based on verbose flag
	loggingConfig := config.NewLoggingConfig(verbose)
	loggingConfig.SetupLogger()
	logger := slog.Default()

	// Log the start of crawl operation with structured logging
	config.LogCrawlStart(targetURL, depth, concurrent, userAgent)

	// Create progress configuration
	progressConfig := &progress.Config{
		ShowProgress: showProgress,
		RateLimit:    rateLimit,
		Logger:       logger,
	}

	// Create crawler configuration
	crawlerConfig := &crawler.Config{
		MaxDepth:       depth,
		SameDomain:     true, // For now, limit to same domain
		UserAgent:      userAgent,
		Logger:         logger,
		Workers:        concurrent,
		ShowProgress:   showProgress,
		ProgressConfig: progressConfig,
	}

	// Create and configure the concurrent crawler
	c, err := crawler.NewConcurrentCrawler(crawlerConfig)
	if err != nil {
		return fmt.Errorf("failed to create crawler: %w", err)
	}

	// Set up signal handling for graceful shutdown

	// Create a channel to listen for interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start crawling in a goroutine
	type crawlResult struct {
		results []crawler.CrawlResult
		stats   *crawler.CrawlStats
		err     error
	}

	resultChan := make(chan crawlResult, 1)
	go func() {
		results, stats, err := c.CrawlConcurrent(targetURL)
		resultChan <- crawlResult{results: results, stats: stats, err: err}
	}()

	// Wait for either completion or interruption
	var results []crawler.CrawlResult
	var stats *crawler.CrawlStats
	var crawlErr error

	select {
	case result := <-resultChan:
		results = result.results
		stats = result.stats
		crawlErr = result.err
	case <-sigChan:
		logger.Info("Received interrupt signal, stopping crawl...")
		c.Cancel()
		// Wait for crawl to stop gracefully
		result := <-resultChan
		results = result.results
		stats = result.stats
		crawlErr = result.err
		logger.Info("Crawl stopped gracefully")
	}

	if crawlErr != nil {
		return fmt.Errorf("crawl failed: %w", crawlErr)
	}

	// Extract all URLs from the results
	var allURLs []string
	for _, result := range results {
		allURLs = append(allURLs, result.URL)
	}

	// Output URLs to stdout (logs are already going to stderr)
	if err := output.OutputURLs(allURLs); err != nil {
		return fmt.Errorf("failed to output URLs: %w", err)
	}

	// Log completion stats to stderr
	config.LogCrawlComplete(targetURL, stats.CrawledURLs, stats.FailedURLs)

	return nil
}

func Execute() error {
	return rootCmd.Execute()
}

func main() {
	if err := Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
