package crawler

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/aoshimash/urlmap/internal/client"
	"github.com/aoshimash/urlmap/internal/parser"
	"github.com/aoshimash/urlmap/internal/progress"
	"github.com/aoshimash/urlmap/internal/url"
)

// CrawlJob represents a job to be processed by a worker
type CrawlJob struct {
	URL   string // URL to crawl
	Depth int    // Depth of this URL in the crawl tree
}

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
	workers    int                   // Number of concurrent workers
}

// ConcurrentCrawler handles concurrent crawling with worker pool
type ConcurrentCrawler struct {
	*Crawler                                // Embed the original crawler
	jobs         chan CrawlJob              // Channel for distributing jobs
	results      chan CrawlResult           // Channel for collecting results
	visited      sync.Map                   // Thread-safe visited URLs tracker
	mu           sync.RWMutex               // Mutex for protecting shared state
	wg           sync.WaitGroup             // WaitGroup for worker synchronization
	ctx          context.Context            // Context for cancellation
	cancel       context.CancelFunc         // Cancel function
	resultsList  []CrawlResult              // Thread-safe results collection
	activeJobs   int                        // Number of active jobs
	activeJobsMu sync.Mutex                 // Mutex for active jobs counter
	progress     *progress.ProgressReporter // Progress reporter for tracking crawl progress
}

// Config holds configuration for the crawler
type Config struct {
	MaxDepth       int              // Maximum depth to crawl (0 = no limit)
	SameDomain     bool             // Whether to limit crawling to same domain
	UserAgent      string           // User agent to use for requests
	Timeout        time.Duration    // Request timeout
	Logger         *slog.Logger     // Logger instance
	Workers        int              // Number of concurrent workers
	ShowProgress   bool             // Whether to show progress indicators
	ProgressConfig *progress.Config // Progress reporting configuration
}

// DefaultConfig returns a default crawler configuration
func DefaultConfig() *Config {
	return &Config{
		MaxDepth:       3,
		SameDomain:     true,
		UserAgent:      "crawld/1.0",
		Timeout:        30 * time.Second,
		Logger:         slog.Default(),
		Workers:        10,
		ShowProgress:   true,
		ProgressConfig: progress.DefaultConfig(),
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

	workers := config.Workers
	if workers <= 0 {
		workers = 10 // Default number of workers
	}

	return &Crawler{
		client:     httpClient,
		parser:     linkExtractor,
		logger:     config.Logger,
		visited:    make(map[string]bool),
		maxDepth:   config.MaxDepth,
		sameDomain: config.SameDomain,
		results:    make([]CrawlResult, 0),
		stats:      CrawlStats{},
		workers:    workers,
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

// NewConcurrentCrawler creates a new concurrent crawler with worker pool
func NewConcurrentCrawler(config *Config) (*ConcurrentCrawler, error) {
	crawler, err := New(config)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	cc := &ConcurrentCrawler{
		Crawler:     crawler,
		jobs:        make(chan CrawlJob, crawler.workers*2), // Buffer for better performance
		results:     make(chan CrawlResult, crawler.workers*2),
		visited:     sync.Map{},
		ctx:         ctx,
		cancel:      cancel,
		resultsList: make([]CrawlResult, 0),
	}

	// Setup progress reporter if enabled
	if config != nil && config.ShowProgress {
		progressConfig := config.ProgressConfig
		if progressConfig == nil {
			progressConfig = progress.DefaultConfig()
		}
		// Override logger if specified in crawler config
		if config.Logger != nil {
			progressConfig.Logger = config.Logger
		}
		cc.progress = progress.NewProgressReporter(progressConfig)
	}

	return cc, nil
}

// CrawlConcurrent performs concurrent crawling starting from the given URL
func (cc *ConcurrentCrawler) CrawlConcurrent(startURL string) ([]CrawlResult, *CrawlStats, error) {
	cc.logger.Info("Starting concurrent crawl", "start_url", startURL, "max_depth", cc.maxDepth, "workers", cc.workers)
	cc.mu.Lock()
	cc.stats.StartTime = time.Now()
	cc.mu.Unlock()

	// Start progress reporter if enabled
	if cc.progress != nil {
		cc.progress.Start()
		defer cc.progress.Stop()
	}

	// Validate and normalize the start URL
	if !url.IsValidURL(startURL) {
		return nil, &cc.stats, fmt.Errorf("invalid start URL: %s", startURL)
	}

	normalizedURL, err := url.NormalizeURL(startURL)
	if err != nil {
		return nil, &cc.stats, fmt.Errorf("failed to normalize start URL: %w", err)
	}

	// Extract base domain for same-domain filtering
	if cc.sameDomain {
		cc.baseDomain, err = url.ExtractDomain(normalizedURL)
		if err != nil {
			return nil, &cc.stats, fmt.Errorf("failed to extract base domain: %w", err)
		}
		cc.logger.Debug("Same-domain filtering enabled", "base_domain", cc.baseDomain)
	}

	// Start workers
	for i := 0; i < cc.workers; i++ {
		cc.wg.Add(1)
		go cc.worker(i)
	}

	// Start result collector
	go cc.resultCollector()

	// Start progress updater if progress reporting is enabled
	if cc.progress != nil {
		go cc.progressUpdater()
	}

	// Add the start URL to the job queue
	cc.visited.Store(normalizedURL, true)
	cc.mu.Lock()
	cc.stats.TotalURLs = 1
	cc.mu.Unlock()
	cc.addJob(CrawlJob{URL: normalizedURL, Depth: 0})

	// Wait for all jobs to complete
	cc.wg.Wait()
	close(cc.results)

	// Wait a bit for result collector to finish
	time.Sleep(100 * time.Millisecond)

	cc.mu.Lock()
	startTime := cc.stats.StartTime
	cc.stats.TotalTime = time.Since(startTime)
	cc.mu.Unlock()

	cc.logger.Info("Concurrent crawling completed",
		"total_urls", cc.stats.TotalURLs,
		"crawled_urls", cc.stats.CrawledURLs,
		"failed_urls", cc.stats.FailedURLs,
		"skipped_urls", cc.stats.SkippedURLs,
		"max_depth_reached", cc.stats.MaxDepthReached,
		"total_time", cc.stats.TotalTime)

	return cc.resultsList, &cc.stats, nil
}

// worker processes jobs from the job queue
func (cc *ConcurrentCrawler) worker(id int) {
	defer cc.wg.Done()
	cc.logger.Debug("Worker started", "worker_id", id)

	for {
		select {
		case job, ok := <-cc.jobs:
			if !ok {
				cc.logger.Debug("Worker stopping - jobs channel closed", "worker_id", id)
				return
			}
			cc.processJob(job, id)
		case <-cc.ctx.Done():
			cc.logger.Debug("Worker stopping - context cancelled", "worker_id", id)
			return
		}
	}
}

// processJob processes a single crawl job
func (cc *ConcurrentCrawler) processJob(job CrawlJob, workerID int) {
	defer func() {
		cc.activeJobsMu.Lock()
		cc.activeJobs--
		isLastJob := cc.activeJobs == 0
		cc.activeJobsMu.Unlock()

		if isLastJob {
			// No more active jobs, close the jobs channel
			close(cc.jobs)
		}
	}()

	cc.logger.Debug("Processing job", "worker_id", workerID, "url", job.URL, "depth", job.Depth)

	// Apply rate limiting if progress reporter is configured with rate limiting
	if cc.progress != nil {
		cc.progress.WaitForRateLimit()
	}

	// Check depth limit
	if cc.maxDepth > 0 && job.Depth >= cc.maxDepth {
		cc.logger.Debug("Skipping job due to depth limit", "url", job.URL, "depth", job.Depth)
		cc.mu.Lock()
		cc.stats.SkippedURLs++
		cc.mu.Unlock()

		// Update progress statistics
		if cc.progress != nil {
			cc.progress.IncrementSkipped()
		}
		return
	}

	// Crawl the URL
	result := cc.crawlSingleConcurrent(job.URL, job.Depth)

	// Update progress statistics based on result
	if cc.progress != nil {
		cc.progress.IncrementProcessed()
		if result.Error != nil {
			cc.progress.IncrementFailed()
		}
	}

	// Send result to collector
	select {
	case cc.results <- result:
	case <-cc.ctx.Done():
		return
	}

	// If successful, add new links to job queue
	if result.Error == nil {
		cc.addLinksToQueue(result.Links, job.Depth)
	}

	// Update max depth reached
	cc.mu.Lock()
	if job.Depth > cc.stats.MaxDepthReached {
		cc.stats.MaxDepthReached = job.Depth
	}
	cc.mu.Unlock()
}

// crawlSingleConcurrent crawls a single URL (thread-safe version)
func (cc *ConcurrentCrawler) crawlSingleConcurrent(targetURL string, depth int) CrawlResult {
	result := CrawlResult{
		URL:       targetURL,
		Depth:     depth,
		FetchTime: time.Now(),
	}

	cc.logger.Debug("Fetching URL", "url", targetURL, "depth", depth)
	startTime := time.Now()

	// Fetch the page
	response, err := cc.client.Get(cc.ctx, targetURL)
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
	if cc.sameDomain {
		result.Links, err = cc.parser.ExtractSameDomainLinks(targetURL, htmlContent)
	} else {
		result.Links, err = cc.parser.ExtractLinks(targetURL, htmlContent)
	}

	if err != nil {
		result.Error = fmt.Errorf("failed to extract links: %w", err)
		return result
	}

	cc.logger.Debug("Extracted links", "url", targetURL, "link_count", len(result.Links))
	return result
}

// addLinksToQueue adds extracted links to the job queue
func (cc *ConcurrentCrawler) addLinksToQueue(links []string, currentDepth int) {
	for _, link := range links {
		// Skip if already visited
		if _, loaded := cc.visited.LoadOrStore(link, true); loaded {
			continue
		}

		// Apply same-domain filtering if enabled
		if cc.sameDomain {
			isSame, err := url.IsSameDomain(cc.baseDomain, link)
			if err != nil || !isSame {
				cc.logger.Debug("Skipping external domain link", "link", link)
				continue
			}
		}

		// Add to job queue
		cc.addJob(CrawlJob{URL: link, Depth: currentDepth + 1})

		cc.mu.Lock()
		cc.stats.TotalURLs++
		cc.mu.Unlock()

		// Update progress statistics
		if cc.progress != nil {
			cc.progress.IncrementDiscovered()
		}

		cc.logger.Debug("Added URL to queue", "url", link, "depth", currentDepth+1)
	}
}

// addJob adds a job to the job queue with proper synchronization
func (cc *ConcurrentCrawler) addJob(job CrawlJob) {
	cc.activeJobsMu.Lock()
	cc.activeJobs++
	cc.activeJobsMu.Unlock()

	select {
	case cc.jobs <- job:
		// Job added successfully
	case <-cc.ctx.Done():
		cc.activeJobsMu.Lock()
		cc.activeJobs--
		cc.activeJobsMu.Unlock()
		return
	}
}

// resultCollector collects results from workers
func (cc *ConcurrentCrawler) resultCollector() {
	for result := range cc.results {
		cc.mu.Lock()
		cc.resultsList = append(cc.resultsList, result)

		if result.Error != nil {
			cc.stats.FailedURLs++
			cc.logger.Warn("Failed to crawl URL", "url", result.URL, "error", result.Error)
		} else {
			cc.stats.CrawledURLs++
			cc.logger.Info("Successfully crawled URL", "url", result.URL, "links_found", len(result.Links))
		}
		cc.mu.Unlock()
	}
}

// Cancel cancels the crawling operation
func (cc *ConcurrentCrawler) Cancel() {
	cc.cancel()
}

// GetResults returns the crawling results (thread-safe)
func (cc *ConcurrentCrawler) GetResults() []CrawlResult {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	results := make([]CrawlResult, len(cc.resultsList))
	copy(results, cc.resultsList)
	return results
}

// GetStats returns the crawling statistics (thread-safe)
func (cc *ConcurrentCrawler) GetStats() *CrawlStats {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	statsCopy := cc.stats
	return &statsCopy
}

// progressUpdater periodically updates progress statistics
func (cc *ConcurrentCrawler) progressUpdater() {
	ticker := time.NewTicker(500 * time.Millisecond) // Update every 500ms
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cc.updateProgressStats()
		case <-cc.ctx.Done():
			return
		}
	}
}

// updateProgressStats updates the progress reporter with current statistics
func (cc *ConcurrentCrawler) updateProgressStats() {
	if cc.progress == nil {
		return
	}

	cc.mu.RLock()
	totalURLs := int64(cc.stats.TotalURLs)
	crawledURLs := int64(cc.stats.CrawledURLs)
	failedURLs := int64(cc.stats.FailedURLs)
	skippedURLs := int64(cc.stats.SkippedURLs)
	cc.mu.RUnlock()

	cc.activeJobsMu.Lock()
	activeJobs := cc.activeJobs
	cc.activeJobsMu.Unlock()

	// Calculate queue size (approximation)
	queueSize := len(cc.jobs)

	// Update progress with current statistics
	cc.progress.UpdateStats(
		crawledURLs, // URLs processed
		totalURLs,   // URLs discovered
		failedURLs,  // URLs failed
		skippedURLs, // URLs skipped
		activeJobs,  // Active workers (approximation)
		queueSize,   // Queue size
	)
}
