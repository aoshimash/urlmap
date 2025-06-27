package crawler

import (
	"log/slog"
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	// Disable logs during testing
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelError,
	})))
	os.Exit(m.Run())
}

func TestNew(t *testing.T) {
	crawler, err := New(nil)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	if crawler == nil {
		t.Fatal("New() returned nil crawler")
	}

	// Test that all fields are properly initialized
	if crawler.client == nil {
		t.Error("Crawler client is nil")
	}

	if crawler.parser == nil {
		t.Error("Crawler parser is nil")
	}

	if crawler.logger == nil {
		t.Error("Crawler logger is nil")
	}

	if crawler.visited == nil {
		t.Error("Crawler visited map is nil")
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	if config == nil {
		t.Fatal("DefaultConfig returned nil")
	}

	if config.MaxDepth != 3 {
		t.Errorf("Expected MaxDepth 3, got %d", config.MaxDepth)
	}

	if !config.SameDomain {
		t.Error("Expected SameDomain to be true")
	}

	if config.UserAgent != "crawld/1.0" {
		t.Errorf("Expected UserAgent 'crawld/1.0', got %s", config.UserAgent)
	}

	if config.Timeout != 30*time.Second {
		t.Errorf("Expected Timeout 30s, got %v", config.Timeout)
	}
}

func TestCrawlRecursive_InvalidURL(t *testing.T) {
	crawler, err := New(nil)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	invalidURLs := []string{
		"",
		"not-a-url",
		"ftp://example.com",
		"javascript:void(0)",
	}

	for _, url := range invalidURLs {
		t.Run("invalid_url_"+url, func(t *testing.T) {
			results, stats, err := crawler.CrawlRecursive(url)
			if err == nil {
				t.Error("Expected error for invalid URL")
			}

			if len(results) != 0 {
				t.Errorf("Expected 0 results, got %d", len(results))
			}

			if stats.TotalURLs != 0 {
				t.Errorf("Expected 0 total URLs, got %d", stats.TotalURLs)
			}

			// Reset for next test
			crawler.Reset()
		})
	}
}

func TestCrawlerUtilityMethods(t *testing.T) {
	crawler, err := New(nil)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	// Test Reset
	crawler.visited["test"] = true
	crawler.results = append(crawler.results, CrawlResult{URL: "test"})
	crawler.stats.TotalURLs = 5

	crawler.Reset()

	if len(crawler.visited) != 0 {
		t.Error("Reset() did not clear visited map")
	}

	if len(crawler.results) != 0 {
		t.Error("Reset() did not clear results")
	}

	if crawler.stats.TotalURLs != 0 {
		t.Error("Reset() did not reset stats")
	}

	// Test GetResults and GetStats
	testResult := CrawlResult{URL: "test", Depth: 1}
	crawler.results = append(crawler.results, testResult)
	crawler.stats.CrawledURLs = 1

	results := crawler.GetResults()
	if len(results) != 1 || results[0].URL != "test" {
		t.Error("GetResults() did not return correct results")
	}

	stats := crawler.GetStats()
	if stats.CrawledURLs != 1 {
		t.Error("GetStats() did not return correct stats")
	}
}
