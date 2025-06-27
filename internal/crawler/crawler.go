package crawler

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/aoshimash/crawld/internal/client"
	"github.com/aoshimash/crawld/internal/parser"
	"github.com/aoshimash/crawld/internal/url"
)

// CrawlResult represents the result of crawling a single URL
type CrawlResult struct {
	URL          string        // The URL that was crawled
	Depth        int           // The depth at which this URL was found
	Links        []string      // Links found on this page
	Error        error         // Error if crawling failed
	FetchTime    time.Time     // When this URL was crawled
	ResponseTime time.Duration // Time taken to fetch this URL
	StatusCode   int           // HTTP status code
}

// CrawlStats holds statistics about the crawling process
type CrawlStats struct {
	TotalURLs       int           // Total URLs discovered
	CrawledURLs     int           // URLs successfully crawled
	FailedURLs      int           // URLs that failed to crawl
	SkippedURLs     int           // URLs skipped (duplicates, depth limit)
	MaxDepthReached int           // Maximum depth reached
	TotalTime       time.Duration // Total crawling time
	StartTime       time.Time     // When crawling started
}

// Crawler represents a web crawler instance with recursive capabilities
type Crawler struct {
	client     *client.Client        // HTTP client for fetching pages
	parser     *parser.LinkExtractor // HTML parser for extracting links
	logger     *slog.Logger          // Logger for structured logging
	visited    map[string]bool       // Track visited URLs to prevent duplicates
	maxDepth   int                   // Maximum crawling depth
	sameDomain bool                  // Whether to limit crawling to same domain
	baseDomain string                // Base domain for same-domain filtering
	results    []CrawlResult         // Results of crawling operations
	stats      CrawlStats            // Crawling statistics
}

// Config holds configuration for the crawler
type Config struct {
	MaxDepth   int           // Maximum depth to crawl (0 = no limit)
	SameDomain bool          // Whether to limit crawling to same domain
	UserAgent  string        // User agent to use for requests
	Timeout    time.Duration // Request timeout
	Logger     *slog.Logger  // Logger instance
}

// DefaultConfig returns a default crawler configuration
func DefaultConfig() *Config {
	return &Config{
		MaxDepth:   3,
		SameDomain: true,
		UserAgent:  "crawld/1.0",
		Timeout:    30 * time.Second,
		Logger:     slog.Default(),
	}
}

// New creates a new crawler instance with the given configuration
func New(config *Config) (*Crawler, error) {
	if config == nil {
		config = DefaultConfig()
	}

	if config.Logger == nil {
		config.Logger = slog.Default()
	}

	// Create HTTP client
	httpClient := client.NewClient(&client.Config{
		UserAgent: config.UserAgent,
		Timeout:   config.Timeout,
	})

	// Create link extractor
	linkExtractor := parser.NewLinkExtractor(config.Logger)

	return &Crawler{
		client:     httpClient,
		parser:     linkExtractor,
		logger:     config.Logger,
		visited:    make(map[string]bool),
		maxDepth:   config.MaxDepth,
		sameDomain: config.SameDomain,
		results:    make([]CrawlResult, 0),
		stats:      CrawlStats{},
	}, nil
}

// CrawlRecursive performs recursive crawling starting from the given URL
func (c *Crawler) CrawlRecursive(startURL string) ([]CrawlResult, *CrawlStats, error) {
	c.logger.Info("Starting recursive crawl", "start_url", startURL, "max_depth", c.maxDepth)
	c.stats.StartTime = time.Now()

	// Validate and normalize the start URL
	if !url.IsValidURL(startURL) {
		return nil, &c.stats, fmt.Errorf("invalid start URL: %s", startURL)
	}

	normalizedURL, err := url.NormalizeURL(startURL)
	if err != nil {
		return nil, &c.stats, fmt.Errorf("failed to normalize start URL: %w", err)
	}

	// Extract base domain for same-domain filtering
	if c.sameDomain {
		c.baseDomain, err = url.ExtractDomain(normalizedURL)
		if err != nil {
			return nil, &c.stats, fmt.Errorf("failed to extract base domain: %w", err)
		}
		c.logger.Debug("Same-domain filtering enabled", "base_domain", c.baseDomain)
	}

	// Initialize crawling queue with the start URL
	type queueItem struct {
		url   string
		depth int
	}

	queue := []queueItem{{url: normalizedURL, depth: 0}}
	c.visited[normalizedURL] = true
	c.stats.TotalURLs = 1

	// Process queue until empty
	for len(queue) > 0 {
		// Dequeue the next URL
		current := queue[0]
		queue = queue[1:]

		c.logger.Debug("Processing URL", "url", current.url, "depth", current.depth, "queue_size", len(queue))

		// Check depth limit
		if c.maxDepth > 0 && current.depth >= c.maxDepth {
			c.logger.Debug("Skipping URL due to depth limit", "url", current.url, "depth", current.depth)
			c.stats.SkippedURLs++
			continue
		}

		// Crawl the current URL
		result := c.crawlSingle(current.url, current.depth)
		c.results = append(c.results, result)

		// Update statistics
		if result.Error != nil {
			c.stats.FailedURLs++
			c.logger.Warn("Failed to crawl URL", "url", current.url, "error", result.Error)
		} else {
			c.stats.CrawledURLs++
			c.logger.Info("Successfully crawled URL", "url", current.url, "links_found", len(result.Links))

			// Add new links to queue
			for _, link := range result.Links {
				// Skip if already visited
				if c.visited[link] {
					continue
				}

				// Apply same-domain filtering if enabled
				if c.sameDomain {
					isSame, err := url.IsSameDomain(c.baseDomain, link)
					if err != nil || !isSame {
						c.logger.Debug("Skipping external domain link", "link", link)
						continue
					}
				}

				// Add to queue and mark as visited
				queue = append(queue, queueItem{url: link, depth: current.depth + 1})
				c.visited[link] = true
				c.stats.TotalURLs++

				c.logger.Debug("Added URL to queue", "url", link, "depth", current.depth+1)
			}
		}

		// Update max depth reached
		if current.depth > c.stats.MaxDepthReached {
			c.stats.MaxDepthReached = current.depth
		}
	}

	c.stats.TotalTime = time.Since(c.stats.StartTime)
	c.logger.Info("Crawling completed",
		"total_urls", c.stats.TotalURLs,
		"crawled_urls", c.stats.CrawledURLs,
		"failed_urls", c.stats.FailedURLs,
		"skipped_urls", c.stats.SkippedURLs,
		"max_depth_reached", c.stats.MaxDepthReached,
		"total_time", c.stats.TotalTime)

	return c.results, &c.stats, nil
}

// crawlSingle crawls a single URL and returns the result
func (c *Crawler) crawlSingle(targetURL string, depth int) CrawlResult {
	result := CrawlResult{
		URL:       targetURL,
		Depth:     depth,
		FetchTime: time.Now(),
	}

	c.logger.Debug("Fetching URL", "url", targetURL, "depth", depth)
	startTime := time.Now()

	// Fetch the page
	response, err := c.client.Get(context.Background(), targetURL)
	result.ResponseTime = time.Since(startTime)

	if err != nil {
		result.Error = fmt.Errorf("failed to fetch URL: %w", err)
		return result
	}

	result.StatusCode = response.StatusCode()

	// Check for successful response
	if response.StatusCode() < 200 || response.StatusCode() >= 400 {
		result.Error = fmt.Errorf("HTTP error: %d", response.StatusCode())
		return result
	}

	// Extract links from the page
	htmlContent := response.String()
	if c.sameDomain {
		result.Links, err = c.parser.ExtractSameDomainLinks(targetURL, htmlContent)
	} else {
		result.Links, err = c.parser.ExtractLinks(targetURL, htmlContent)
	}

	if err != nil {
		result.Error = fmt.Errorf("failed to extract links: %w", err)
		return result
	}

	c.logger.Debug("Extracted links", "url", targetURL, "link_count", len(result.Links))
	return result
}

// GetResults returns the crawling results
func (c *Crawler) GetResults() []CrawlResult {
	return c.results
}

// GetStats returns the crawling statistics
func (c *Crawler) GetStats() *CrawlStats {
	return &c.stats
}

// Reset clears the crawler state for a new crawling session
func (c *Crawler) Reset() {
	c.visited = make(map[string]bool)
	c.results = make([]CrawlResult, 0)
	c.stats = CrawlStats{}
	c.baseDomain = ""
}

// GetAllURLs returns all discovered URLs (both crawled and failed)
func (c *Crawler) GetAllURLs() []string {
	urls := make([]string, 0, len(c.results))
	for _, result := range c.results {
		urls = append(urls, result.URL)
	}
	return urls
}

// GetSuccessfulURLs returns only successfully crawled URLs
func (c *Crawler) GetSuccessfulURLs() []string {
	urls := make([]string, 0)
	for _, result := range c.results {
		if result.Error == nil {
			urls = append(urls, result.URL)
		}
	}
	return urls
}
