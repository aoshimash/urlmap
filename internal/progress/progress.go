package progress

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"
	"time"
)

// Stats holds performance and progress statistics
type Stats struct {
	StartTime      time.Time     // When crawling started
	URLsProcessed  int64         // Number of URLs processed
	URLsDiscovered int64         // Total URLs discovered
	URLsFailed     int64         // Number of failed URLs
	URLsSkipped    int64         // Number of skipped URLs
	ActiveWorkers  int           // Number of active workers
	QueueSize      int           // Current queue size
	ProcessingRate float64       // URLs per second
	ElapsedTime    time.Duration // Time elapsed since start
	LastUpdateTime time.Time     // Last time stats were updated
}

// ProgressReporter manages progress display and statistics
type ProgressReporter struct {
	stats          Stats
	mu             sync.RWMutex
	ticker         *time.Ticker
	output         io.Writer
	logger         *slog.Logger
	updateInterval time.Duration
	showProgress   bool
	rateLimiter    *RateLimiter
	done           chan struct{}
	wg             sync.WaitGroup
}

// RateLimiter controls the rate of requests
type RateLimiter struct {
	requestsPerSecond float64
	ticker            *time.Ticker
	tokens            chan struct{}
	mu                sync.Mutex
	enabled           bool
}

// Config holds configuration for progress reporting and rate limiting
type Config struct {
	UpdateInterval time.Duration // How often to update progress (default: 1s)
	ShowProgress   bool          // Whether to show progress indicators
	Output         io.Writer     // Where to write progress (default: stderr)
	Logger         *slog.Logger  // Logger instance
	RateLimit      float64       // Requests per second (0 = no limit)
}

// DefaultConfig returns a default progress configuration
func DefaultConfig() *Config {
	return &Config{
		UpdateInterval: 1 * time.Second,
		ShowProgress:   true,
		Output:         os.Stderr,
		Logger:         slog.Default(),
		RateLimit:      0, // No rate limiting by default
	}
}

// NewProgressReporter creates a new progress reporter
func NewProgressReporter(config *Config) *ProgressReporter {
	if config == nil {
		config = DefaultConfig()
	}

	if config.Output == nil {
		config.Output = os.Stderr
	}

	if config.Logger == nil {
		config.Logger = slog.Default()
	}

	if config.UpdateInterval <= 0 {
		config.UpdateInterval = 1 * time.Second
	}

	pr := &ProgressReporter{
		stats: Stats{
			StartTime:      time.Now(),
			LastUpdateTime: time.Now(),
		},
		output:         config.Output,
		logger:         config.Logger,
		updateInterval: config.UpdateInterval,
		showProgress:   config.ShowProgress,
		done:           make(chan struct{}),
	}

	// Setup rate limiter if specified
	if config.RateLimit > 0 {
		pr.rateLimiter = NewRateLimiter(config.RateLimit)
	}

	return pr
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(requestsPerSecond float64) *RateLimiter {
	if requestsPerSecond <= 0 {
		return &RateLimiter{enabled: false}
	}

	interval := time.Duration(float64(time.Second) / requestsPerSecond)

	rl := &RateLimiter{
		requestsPerSecond: requestsPerSecond,
		ticker:            time.NewTicker(interval),
		tokens:            make(chan struct{}, int(requestsPerSecond)+1), // Buffer for burst
		enabled:           true,
	}

	// Fill initial tokens
	for i := 0; i < cap(rl.tokens); i++ {
		rl.tokens <- struct{}{}
	}

	// Start token refill goroutine
	go rl.refillTokens()

	return rl
}

// refillTokens continuously refills the token bucket
func (rl *RateLimiter) refillTokens() {
	for range rl.ticker.C {
		select {
		case rl.tokens <- struct{}{}:
			// Token added
		default:
			// Buffer full, skip
		}
	}
}

// Wait blocks until a token is available (rate limiting)
func (rl *RateLimiter) Wait() {
	if !rl.enabled {
		return
	}

	<-rl.tokens
}

// Stop stops the rate limiter
func (rl *RateLimiter) Stop() {
	if rl.enabled && rl.ticker != nil {
		rl.ticker.Stop()
	}
}

// Start starts the progress reporter
func (pr *ProgressReporter) Start() {
	if !pr.showProgress {
		return
	}

	pr.ticker = time.NewTicker(pr.updateInterval)
	pr.wg.Add(1)

	go func() {
		defer pr.wg.Done()
		for {
			select {
			case <-pr.ticker.C:
				pr.updateProgress()
			case <-pr.done:
				return
			}
		}
	}()
}

// Stop stops the progress reporter and displays final statistics
func (pr *ProgressReporter) Stop() {
	if pr.ticker != nil {
		pr.ticker.Stop()
	}

	close(pr.done)
	pr.wg.Wait()

	if pr.rateLimiter != nil {
		pr.rateLimiter.Stop()
	}

	// Display final statistics
	pr.displayFinalStats()
}

// UpdateStats updates the current statistics
func (pr *ProgressReporter) UpdateStats(processed, discovered, failed, skipped int64, activeWorkers, queueSize int) {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	pr.stats.URLsProcessed = processed
	pr.stats.URLsDiscovered = discovered
	pr.stats.URLsFailed = failed
	pr.stats.URLsSkipped = skipped
	pr.stats.ActiveWorkers = activeWorkers
	pr.stats.QueueSize = queueSize
	pr.stats.ElapsedTime = time.Since(pr.stats.StartTime)

	// Calculate processing rate
	if pr.stats.ElapsedTime.Seconds() > 0 {
		pr.stats.ProcessingRate = float64(pr.stats.URLsProcessed) / pr.stats.ElapsedTime.Seconds()
	}

	pr.stats.LastUpdateTime = time.Now()
}

// IncrementProcessed increments the processed URLs counter
func (pr *ProgressReporter) IncrementProcessed() {
	pr.mu.Lock()
	defer pr.mu.Unlock()
	pr.stats.URLsProcessed++
}

// IncrementDiscovered increments the discovered URLs counter
func (pr *ProgressReporter) IncrementDiscovered() {
	pr.mu.Lock()
	defer pr.mu.Unlock()
	pr.stats.URLsDiscovered++
}

// IncrementFailed increments the failed URLs counter
func (pr *ProgressReporter) IncrementFailed() {
	pr.mu.Lock()
	defer pr.mu.Unlock()
	pr.stats.URLsFailed++
}

// IncrementSkipped increments the skipped URLs counter
func (pr *ProgressReporter) IncrementSkipped() {
	pr.mu.Lock()
	defer pr.mu.Unlock()
	pr.stats.URLsSkipped++
}

// WaitForRateLimit waits for rate limiting if enabled
func (pr *ProgressReporter) WaitForRateLimit() {
	if pr.rateLimiter != nil {
		pr.rateLimiter.Wait()
	}
}

// GetStats returns a copy of current statistics
func (pr *ProgressReporter) GetStats() Stats {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	// Update elapsed time and rate
	stats := pr.stats
	stats.ElapsedTime = time.Since(pr.stats.StartTime)
	if stats.ElapsedTime.Seconds() > 0 {
		stats.ProcessingRate = float64(stats.URLsProcessed) / stats.ElapsedTime.Seconds()
	}

	return stats
}

// updateProgress displays current progress information
func (pr *ProgressReporter) updateProgress() {
	stats := pr.GetStats()

	if stats.URLsProcessed == 0 && stats.URLsDiscovered == 0 {
		return // Nothing to show yet
	}

	// Format progress message
	var progressMsg string
	if stats.QueueSize > 0 {
		// Still crawling
		progressMsg = fmt.Sprintf("\rCrawling: %d/%d URLs processed (%.1f URLs/sec) [%d workers, %d queued]",
			stats.URLsProcessed,
			stats.URLsDiscovered,
			stats.ProcessingRate,
			stats.ActiveWorkers,
			stats.QueueSize)
	} else {
		// Likely finished or paused
		progressMsg = fmt.Sprintf("\rProcessed: %d URLs (%.1f URLs/sec, %.1fs elapsed)",
			stats.URLsProcessed,
			stats.ProcessingRate,
			stats.ElapsedTime.Seconds())
	}

	fmt.Fprint(pr.output, progressMsg)
}

// displayFinalStats displays final crawling statistics
func (pr *ProgressReporter) displayFinalStats() {
	if !pr.showProgress {
		return
	}

	stats := pr.GetStats()

	// Clear the progress line
	fmt.Fprint(pr.output, "\r")

	// Display final summary
	fmt.Fprintf(pr.output, "Crawling completed in %.2fs:\n", stats.ElapsedTime.Seconds())
	fmt.Fprintf(pr.output, "  URLs discovered: %d\n", stats.URLsDiscovered)
	fmt.Fprintf(pr.output, "  URLs processed:  %d\n", stats.URLsProcessed)

	if stats.URLsFailed > 0 {
		fmt.Fprintf(pr.output, "  URLs failed:     %d\n", stats.URLsFailed)
	}

	if stats.URLsSkipped > 0 {
		fmt.Fprintf(pr.output, "  URLs skipped:    %d\n", stats.URLsSkipped)
	}

	fmt.Fprintf(pr.output, "  Average rate:    %.1f URLs/sec\n", stats.ProcessingRate)

	if pr.rateLimiter != nil && pr.rateLimiter.enabled {
		fmt.Fprintf(pr.output, "  Rate limit:      %.1f requests/sec\n", pr.rateLimiter.requestsPerSecond)
	}

	fmt.Fprintln(pr.output)
}

// IsRateLimited returns true if rate limiting is enabled
func (pr *ProgressReporter) IsRateLimited() bool {
	return pr.rateLimiter != nil && pr.rateLimiter.enabled
}
