package main

import (
	"fmt"
	"net/url"
	"os"

	"github.com/aoshimash/crawld/internal/config"
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
	depth      int
	verbose    bool
	userAgent  string
	concurrent int
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "crawld <URL>",
	Short: "A web crawler daemon",
	Long: `Crawld is a web crawler daemon for collecting and processing web content.

This tool crawls web pages starting from a given URL and can be configured
with various options to control the crawling behavior.

Examples:
  crawld https://example.com
  crawld -d 3 -c 5 https://example.com
  crawld --verbose --user-agent "MyBot/1.0" https://example.com`,
	Args: cobra.ExactArgs(1), // Require exactly one URL argument
	RunE: runCrawl,
}

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("crawld version %s\n", version)
		fmt.Printf("commit: %s\n", commit)
		fmt.Printf("built: %s\n", date)
	},
}

func init() {
	// Add flags to the root command
	rootCmd.Flags().IntVarP(&depth, "depth", "d", 0, "Maximum crawl depth (0 = unlimited)")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")
	rootCmd.Flags().StringVarP(&userAgent, "user-agent", "u", "crawld/1.0.0 (+https://github.com/aoshimash/crawld)", "Custom User-Agent string")
	rootCmd.Flags().IntVarP(&concurrent, "concurrent", "c", 10, "Number of concurrent requests")

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

	// Log the start of crawl operation with structured logging
	config.LogCrawlStart(targetURL, depth, concurrent, userAgent)

	// TODO: Implement actual crawling logic
	// Output goes to stdout, logs go to stderr
	fmt.Printf("Crawling %s with depth %d and %d concurrent requests\n", targetURL, depth, concurrent)
	fmt.Printf("User-Agent: %s\n", userAgent)

	// Example of structured logging calls
	if verbose {
		config.LogInfo("Verbose logging enabled")
	}

	// Example of how to log progress and errors (for future implementation)
	// config.LogCrawlProgress(targetURL, 0, 200)
	// config.LogCrawlError(targetURL, 0, fmt.Errorf("example error"))
	// config.LogCrawlComplete(targetURL, 1, 0)

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
