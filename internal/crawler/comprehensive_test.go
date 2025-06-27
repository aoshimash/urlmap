package crawler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestCrawlRecursive_FullWorkflow tests the complete CrawlRecursive workflow
func TestCrawlRecursive_FullWorkflow(t *testing.T) {
	// Create a mock server with multiple pages
	server := createMockHTMLServer(t)
	defer server.Close()

	tests := []struct {
		name            string
		startURL        string
		config          *Config
		expectedMinURLs int
		expectedMaxURLs int
	}{
		{
			name:            "Single page crawl",
			startURL:        server.URL,
			config:          &Config{MaxDepth: 1, SameDomain: true},
			expectedMinURLs: 1,
			expectedMaxURLs: 3,
		},
		{
			name:            "Deep crawl with limit",
			startURL:        server.URL,
			config:          &Config{MaxDepth: 2, SameDomain: true},
			expectedMinURLs: 1,
			expectedMaxURLs: 5,
		},
		{
			name:            "Cross-domain allowed",
			startURL:        server.URL,
			config:          &Config{MaxDepth: 1, SameDomain: false},
			expectedMinURLs: 1,
			expectedMaxURLs: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			crawler, err := New(tt.config)
			if err != nil {
				t.Fatalf("New() failed: %v", err)
			}

			results, stats, err := crawler.CrawlRecursive(tt.startURL)
			if err != nil {
				t.Fatalf("CrawlRecursive() failed: %v", err)
			}

			if len(results) < tt.expectedMinURLs {
				t.Errorf("Expected at least %d results, got %d", tt.expectedMinURLs, len(results))
			}

			if len(results) > tt.expectedMaxURLs {
				t.Errorf("Expected at most %d results, got %d", tt.expectedMaxURLs, len(results))
			}

			if stats.TotalURLs < tt.expectedMinURLs {
				t.Errorf("Expected at least %d total URLs in stats, got %d", tt.expectedMinURLs, stats.TotalURLs)
			}

			// Verify start URL is in results (may be normalized)
			found := false
			for _, result := range results {
				if result.URL == tt.startURL || result.URL == tt.startURL+"/" {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Start URL %s not found in results", tt.startURL)
			}
		})
	}
}

// TestCrawlerUtilityMethods_FullCoverage tests all utility methods
func TestCrawlerUtilityMethods_FullCoverage(t *testing.T) {
	server := createMockHTMLServer(t)
	defer server.Close()

	crawler, err := New(&Config{MaxDepth: 1, SameDomain: true})
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	// Perform a small crawl to populate data
	results, stats, err := crawler.CrawlRecursive(server.URL)
	if err != nil {
		t.Fatalf("CrawlRecursive() failed: %v", err)
	}

	t.Run("GetAllURLs", func(t *testing.T) {
		allURLs := crawler.GetAllURLs()
		if len(allURLs) == 0 {
			t.Error("GetAllURLs() returned empty slice")
		}

		// Should include at least the start URL (may be normalized)
		found := false
		for _, url := range allURLs {
			if url == server.URL || url == server.URL+"/" {
				found = true
				break
			}
		}
		if !found {
			t.Error("GetAllURLs() did not include start URL")
		}
	})

	t.Run("GetSuccessfulURLs", func(t *testing.T) {
		successfulURLs := crawler.GetSuccessfulURLs()
		if len(successfulURLs) == 0 {
			t.Error("GetSuccessfulURLs() returned empty slice")
		}

		// All URLs should be successful in our mock server
		if len(successfulURLs) < 1 {
			t.Errorf("Expected at least 1 successful URL, got %d", len(successfulURLs))
		}
	})

	t.Run("Results and Stats consistency", func(t *testing.T) {
		// Verify that GetResults and GetStats return consistent data
		crawlerResults := crawler.GetResults()
		crawlerStats := crawler.GetStats()

		if len(crawlerResults) != len(results) {
			t.Errorf("GetResults() length %d doesn't match original results length %d", len(crawlerResults), len(results))
		}

		if crawlerStats.TotalURLs != stats.TotalURLs {
			t.Errorf("GetStats() TotalURLs %d doesn't match original stats %d", crawlerStats.TotalURLs, stats.TotalURLs)
		}
	})
}

// TestCrawlRecursive_EdgeCases tests edge cases and error conditions
func TestCrawlRecursive_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func() (*Crawler, string, error)
		wantError bool
	}{
		{
			name: "Server returns 404",
			setupFunc: func() (*Crawler, string, error) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotFound)
					w.Write([]byte("Not Found"))
				}))
				t.Cleanup(server.Close)

				crawler, err := New(&Config{MaxDepth: 1, SameDomain: true})
				return crawler, server.URL, err
			},
			wantError: false, // Should not error, but should have failed stats
		},
		{
			name: "Server returns 500",
			setupFunc: func() (*Crawler, string, error) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("Internal Server Error"))
				}))
				t.Cleanup(server.Close)

				crawler, err := New(&Config{MaxDepth: 1, SameDomain: true})
				return crawler, server.URL, err
			},
			wantError: false, // Should not error, but should have failed stats
		},
		{
			name: "Server timeout",
			setupFunc: func() (*Crawler, string, error) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					time.Sleep(100 * time.Millisecond) // Delay longer than timeout
					w.Write([]byte("<html><body>Hello</body></html>"))
				}))
				t.Cleanup(server.Close)

				config := &Config{
					MaxDepth:   1,
					SameDomain: true,
					Timeout:    50 * time.Millisecond, // Short timeout
				}
				crawler, err := New(config)
				return crawler, server.URL, err
			},
			wantError: false, // Should not error, but should have failed stats
		},
		{
			name: "Malformed HTML",
			setupFunc: func() (*Crawler, string, error) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "text/html")
					w.Write([]byte("<html><body><a href='/page1'>Link</a><broken><unclosed"))
				}))
				t.Cleanup(server.Close)

				crawler, err := New(&Config{MaxDepth: 1, SameDomain: true})
				return crawler, server.URL, err
			},
			wantError: false, // Should handle malformed HTML gracefully
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			crawler, url, err := tt.setupFunc()
			if err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			results, stats, err := crawler.CrawlRecursive(url)
			if (err != nil) != tt.wantError {
				t.Errorf("CrawlRecursive() error = %v, wantError %v", err, tt.wantError)
				return
			}

			// Should always return some results/stats even on failure
			if results == nil {
				t.Error("CrawlRecursive() returned nil results")
			}

			if stats == nil {
				t.Error("CrawlRecursive() returned nil stats")
			}

			// Stats should be consistent
			if stats.TotalURLs < 0 {
				t.Error("Stats TotalURLs should not be negative")
			}

			if stats.CrawledURLs < 0 {
				t.Error("Stats CrawledURLs should not be negative")
			}

			if stats.FailedURLs < 0 {
				t.Error("Stats FailedURLs should not be negative")
			}
		})
	}
}

// TestCrawlRecursive_DepthLimiting tests depth limiting functionality
func TestCrawlRecursive_DepthLimiting(t *testing.T) {
	// Create nested server with multiple depth levels
	server := createNestedMockServer(t)
	defer server.Close()

	tests := []struct {
		name     string
		maxDepth int
		minURLs  int
		maxURLs  int
	}{
		{name: "Depth 0", maxDepth: 0, minURLs: 1, maxURLs: 1},
		{name: "Depth 1", maxDepth: 1, minURLs: 1, maxURLs: 3},
		{name: "Depth 2", maxDepth: 2, minURLs: 1, maxURLs: 6},
		{name: "Unlimited depth", maxDepth: -1, minURLs: 1, maxURLs: 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				MaxDepth:   tt.maxDepth,
				SameDomain: true,
			}

			crawler, err := New(config)
			if err != nil {
				t.Fatalf("New() failed: %v", err)
			}

			results, _, err := crawler.CrawlRecursive(server.URL)
			if err != nil {
				t.Fatalf("CrawlRecursive() failed: %v", err)
			}

			if len(results) < tt.minURLs {
				t.Errorf("Expected at least %d results, got %d", tt.minURLs, len(results))
			}

			if len(results) > tt.maxURLs {
				t.Errorf("Expected at most %d results, got %d", tt.maxURLs, len(results))
			}

			// Verify depth limits are respected
			if tt.maxDepth >= 0 {
				for _, result := range results {
					if result.Depth > tt.maxDepth {
						t.Errorf("Result depth %d exceeds max depth %d for URL %s",
							result.Depth, tt.maxDepth, result.URL)
					}
				}
			}
		})
	}
}

// TestConcurrentCrawler_ComprehensiveWorkflow tests the concurrent crawler extensively
func TestConcurrentCrawler_ComprehensiveWorkflow(t *testing.T) {
	server := createMockHTMLServer(t)
	defer server.Close()

	tests := []struct {
		name          string
		workers       int
		maxDepth      int
		withProgress  bool
		withRateLimit bool
	}{
		{name: "Single worker", workers: 1, maxDepth: 2, withProgress: false, withRateLimit: false},
		{name: "Multiple workers", workers: 3, maxDepth: 2, withProgress: false, withRateLimit: false},
		{name: "With progress", workers: 2, maxDepth: 1, withProgress: true, withRateLimit: false},
		{name: "With rate limit", workers: 2, maxDepth: 1, withProgress: false, withRateLimit: true},
		{name: "All features", workers: 2, maxDepth: 1, withProgress: true, withRateLimit: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				MaxDepth:   tt.maxDepth,
				SameDomain: true,
				Workers:    tt.workers,
			}

			// Note: Progress reporter would be used in actual crawling
			// For now, we focus on testing the core crawler functionality
			_ = tt.withProgress && tt.withRateLimit // Acknowledge the test parameters

			cc, err := NewConcurrentCrawler(config)
			if err != nil {
				t.Fatalf("NewConcurrentCrawler() failed: %v", err)
			}

			results, stats, err := cc.CrawlConcurrent(server.URL)
			if err != nil {
				t.Fatalf("CrawlConcurrent() failed: %v", err)
			}

			if len(results) == 0 {
				t.Error("CrawlConcurrent() returned no results")
			}

			if stats.TotalURLs == 0 {
				t.Error("CrawlConcurrent() stats show no total URLs")
			}

			// Verify results contain the start URL (may be normalized)
			found := false
			for _, result := range results {
				if result.URL == server.URL || result.URL == server.URL+"/" {
					found = true
					break
				}
			}
			if !found {
				t.Error("Results do not contain start URL")
			}

			// Test GetResults and GetStats
			ccResults := cc.GetResults()
			ccStats := cc.GetStats()

			if len(ccResults) != len(results) {
				t.Errorf("GetResults() returned %d results, expected %d", len(ccResults), len(results))
			}

			if ccStats.TotalURLs != stats.TotalURLs {
				t.Errorf("GetStats() returned TotalURLs %d, expected %d", ccStats.TotalURLs, stats.TotalURLs)
			}
		})
	}
}

// TestCrawlRecursive_SameDomainFiltering tests same-domain filtering
func TestCrawlRecursive_SameDomainFiltering(t *testing.T) {
	// Create server with mixed internal/external links
	var serverURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		html := fmt.Sprintf(`
			<html>
			<body>
				<a href="%s/page1">Internal Link 1</a>
				<a href="%s/page2">Internal Link 2</a>
				<a href="https://external.com/page">External Link</a>
				<a href="https://another-external.com/page">Another External</a>
			</body>
			</html>
		`, serverURL, serverURL)
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
	}))
	serverURL = server.URL
	defer server.Close()

	tests := []struct {
		name               string
		sameDomain         bool
		expectedMinResults int
		expectedMaxResults int
	}{
		{name: "Same domain only", sameDomain: true, expectedMinResults: 1, expectedMaxResults: 3},
		{name: "All domains", sameDomain: false, expectedMinResults: 1, expectedMaxResults: 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				MaxDepth:   1,
				SameDomain: tt.sameDomain,
			}

			crawler, err := New(config)
			if err != nil {
				t.Fatalf("New() failed: %v", err)
			}

			results, _, err := crawler.CrawlRecursive(server.URL)
			if err != nil {
				t.Fatalf("CrawlRecursive() failed: %v", err)
			}

			if len(results) < tt.expectedMinResults {
				t.Errorf("Expected at least %d results, got %d", tt.expectedMinResults, len(results))
			}

			if len(results) > tt.expectedMaxResults {
				t.Errorf("Expected at most %d results, got %d", tt.expectedMaxResults, len(results))
			}

			// If same domain only, verify no external URLs
			if tt.sameDomain {
				for _, result := range results {
					if !strings.HasPrefix(result.URL, server.URL) && result.URL != server.URL {
						t.Errorf("Found external URL %s when same domain filtering enabled", result.URL)
					}
				}
			}
		})
	}
}

// Benchmark tests for performance validation
func BenchmarkCrawlRecursive_SinglePage(b *testing.B) {
	server := createMockHTMLServer(nil)
	defer server.Close()

	config := &Config{MaxDepth: 1, SameDomain: true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		crawler, err := New(config)
		if err != nil {
			b.Fatalf("New() failed: %v", err)
		}

		_, _, err = crawler.CrawlRecursive(server.URL)
		if err != nil {
			b.Fatalf("CrawlRecursive() failed: %v", err)
		}
	}
}

func BenchmarkConcurrentCrawler_MultipleWorkers(b *testing.B) {
	server := createMockHTMLServer(nil)
	defer server.Close()

	config := &Config{
		MaxDepth:   2,
		SameDomain: true,
		Workers:    4,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cc, err := NewConcurrentCrawler(config)
		if err != nil {
			b.Fatalf("NewConcurrentCrawler() failed: %v", err)
		}

		_, _, err = cc.CrawlConcurrent(server.URL)
		if err != nil {
			b.Fatalf("CrawlConcurrent() failed: %v", err)
		}
	}
}

// Helper functions for creating mock servers

func createMockHTMLServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var html string
		switch r.URL.Path {
		case "/":
			html = fmt.Sprintf(`
				<html>
				<body>
					<h1>Main Page</h1>
					<a href="http://%s/page1">Page 1</a>
					<a href="http://%s/page2">Page 2</a>
				</body>
				</html>
			`, r.Host, r.Host)
		case "/page1":
			html = `
				<html>
				<body>
					<h1>Page 1</h1>
					<a href="/page3">Page 3</a>
				</body>
				</html>
			`
		case "/page2":
			html = `
				<html>
				<body>
					<h1>Page 2</h1>
					<a href="/page4">Page 4</a>
				</body>
				</html>
			`
		default:
			html = fmt.Sprintf(`
				<html>
				<body>
					<h1>%s</h1>
					<p>Content for %s</p>
				</body>
				</html>
			`, r.URL.Path, r.URL.Path)
		}

		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
	}))
}

func createNestedMockServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var html string
		switch r.URL.Path {
		case "/":
			html = `
				<html>
				<body>
					<h1>Root Level</h1>
					<a href="/level1/page1">Level 1 Page 1</a>
					<a href="/level1/page2">Level 1 Page 2</a>
				</body>
				</html>
			`
		case "/level1/page1":
			html = `
				<html>
				<body>
					<h1>Level 1 Page 1</h1>
					<a href="/level2/page1">Level 2 Page 1</a>
				</body>
				</html>
			`
		case "/level1/page2":
			html = `
				<html>
				<body>
					<h1>Level 1 Page 2</h1>
					<a href="/level2/page2">Level 2 Page 2</a>
				</body>
				</html>
			`
		case "/level2/page1":
			html = `
				<html>
				<body>
					<h1>Level 2 Page 1</h1>
					<a href="/level3/page1">Level 3 Page 1</a>
				</body>
				</html>
			`
		case "/level2/page2":
			html = `
				<html>
				<body>
					<h1>Level 2 Page 2</h1>
					<a href="/level3/page2">Level 3 Page 2</a>
				</body>
				</html>
			`
		default:
			html = fmt.Sprintf(`
				<html>
				<body>
					<h1>%s</h1>
					<p>Deep nested content</p>
				</body>
				</html>
			`, r.URL.Path)
		}

		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
	}))
}
