package robots

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewRobotsChecker(t *testing.T) {
	checker := NewRobotsChecker("TestBot/1.0", nil)

	if checker.userAgent != "TestBot/1.0" {
		t.Errorf("Expected userAgent to be 'TestBot/1.0', got %s", checker.userAgent)
	}

	if checker.cache == nil {
		t.Error("Expected cache to be initialized")
	}

	if checker.logger == nil {
		t.Error("Expected logger to be initialized")
	}
}

func TestMatchesUserAgent(t *testing.T) {
	checker := NewRobotsChecker("MyBot/1.0 (http://example.com)", slog.Default())

	tests := []struct {
		pattern  string
		expected bool
	}{
		{"*", true},
		{"MyBot", true},
		{"mybot", true},
		{"OtherBot", false},
		{"Bot", true}, // partial match
		{"", false},
	}

	for _, test := range tests {
		result := checker.matchesUserAgent(test.pattern)
		if result != test.expected {
			t.Errorf("matchesUserAgent(%q) = %v, expected %v", test.pattern, result, test.expected)
		}
	}
}

func TestPathMatches(t *testing.T) {
	checker := NewRobotsChecker("TestBot/1.0", slog.Default())

	tests := []struct {
		pattern  string
		urlPath  string
		expected bool
	}{
		{"/admin", "/admin", true},
		{"/admin", "/admin/", true},     // robots.txt prefix matching
		{"/admin", "/admin/page", true}, // robots.txt prefix matching
		{"/admin/", "/admin/page", true},
		{"/admin/*", "/admin/page", true},
		{"/admin/*", "/admin/", true},
		{"/admin/*", "/other", false},
		{"", "/any", false},
		{"/", "/", true},
		{"/", "/any", true},
	}

	for _, test := range tests {
		result := checker.pathMatches(test.pattern, test.urlPath)
		if result != test.expected {
			t.Errorf("pathMatches(%q, %q) = %v, expected %v",
				test.pattern, test.urlPath, result, test.expected)
		}
	}
}

func TestCheckRules(t *testing.T) {
	checker := NewRobotsChecker("TestBot/1.0", slog.Default())

	rules := []Rule{
		{UserAgent: "TestBot", Directive: "Disallow", Path: "/admin"},
		{UserAgent: "TestBot", Directive: "Allow", Path: "/admin/public"},
		{UserAgent: "TestBot", Directive: "Disallow", Path: "/private/*"},
	}

	tests := []struct {
		urlPath  string
		expected bool
	}{
		{"/", true},              // default allowed
		{"/admin", false},        // disallowed
		{"/admin/public", true},  // explicitly allowed
		{"/admin/secret", false}, // disallowed by /admin rule
		{"/private/data", false}, // disallowed by wildcard
		{"/public", true},        // not covered by rules
	}

	for _, test := range tests {
		result := checker.checkRules(rules, test.urlPath)
		if result != test.expected {
			t.Errorf("checkRules for path %q = %v, expected %v",
				test.urlPath, result, test.expected)
		}
	}
}

func TestFetchRobots(t *testing.T) {
	// Create a test server with robots.txt
	robotsContent := `User-agent: *
Disallow: /admin/
Disallow: /private/
Allow: /admin/public/

User-agent: TestBot
Disallow: /special/
Crawl-delay: 2

Sitemap: https://example.com/sitemap.xml
`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/robots.txt" {
			w.Header().Set("Content-Type", "text/plain")
			fmt.Fprint(w, robotsContent)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	checker := NewRobotsChecker("TestBot/1.0", slog.Default())

	robotsData, err := checker.fetchRobots(server.URL)
	if err != nil {
		t.Fatalf("fetchRobots failed: %v", err)
	}

	// Check that rules were parsed
	if len(robotsData.rules) == 0 {
		t.Error("Expected rules to be parsed")
	}

	// Check crawl delay
	expectedDelay := 2 * time.Second
	if robotsData.crawlDelay != expectedDelay {
		t.Errorf("Expected crawl delay %v, got %v", expectedDelay, robotsData.crawlDelay)
	}

	// Check sitemaps
	if len(robotsData.sitemaps) != 1 || robotsData.sitemaps[0] != "https://example.com/sitemap.xml" {
		t.Errorf("Expected sitemap to be parsed correctly, got %v", robotsData.sitemaps)
	}
}

func TestIsAllowed(t *testing.T) {
	robotsContent := `User-agent: TestBot
Disallow: /admin/
Allow: /admin/public/
Disallow: /private/*
`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/robots.txt" {
			fmt.Fprint(w, robotsContent)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	checker := NewRobotsChecker("TestBot/1.0", slog.Default())

	tests := []struct {
		url      string
		expected bool
	}{
		{server.URL + "/", true},
		{server.URL + "/admin/", false},
		{server.URL + "/admin/public/", true},
		{server.URL + "/private/data", false},
		{server.URL + "/allowed", true},
	}

	for _, test := range tests {
		result, err := checker.IsAllowed(test.url)
		if err != nil {
			t.Errorf("IsAllowed(%q) failed: %v", test.url, err)
			continue
		}

		if result != test.expected {
			t.Errorf("IsAllowed(%q) = %v, expected %v", test.url, result, test.expected)
		}
	}
}

func TestIsAllowedWithMissingRobots(t *testing.T) {
	// Server that returns 404 for robots.txt
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	checker := NewRobotsChecker("TestBot/1.0", slog.Default())

	// Should allow by default when robots.txt is missing
	allowed, err := checker.IsAllowed(server.URL + "/any-path")
	if err != nil {
		t.Errorf("IsAllowed failed: %v", err)
	}

	if !allowed {
		t.Error("Expected URL to be allowed when robots.txt is missing")
	}
}

func TestGetCrawlDelay(t *testing.T) {
	robotsContent := `User-agent: TestBot
Crawl-delay: 5
`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/robots.txt" {
			fmt.Fprint(w, robotsContent)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	checker := NewRobotsChecker("TestBot/1.0", slog.Default())

	delay, err := checker.GetCrawlDelay(server.URL)
	if err != nil {
		t.Fatalf("GetCrawlDelay failed: %v", err)
	}

	expectedDelay := 5 * time.Second
	if delay != expectedDelay {
		t.Errorf("Expected crawl delay %v, got %v", expectedDelay, delay)
	}
}

func TestCacheFunction(t *testing.T) {
	checker := NewRobotsChecker("TestBot/1.0", slog.Default())

	// Initially cache should be empty
	if checker.GetCacheSize() != 0 {
		t.Error("Expected cache to be empty initially")
	}

	// Add some dummy data to cache
	checker.cache["https://example.com"] = &RobotsData{
		rules: []Rule{{UserAgent: "TestBot", Directive: "Disallow", Path: "/test"}},
	}

	if checker.GetCacheSize() != 1 {
		t.Error("Expected cache size to be 1")
	}

	// Clear cache
	checker.ClearCache()
	if checker.GetCacheSize() != 0 {
		t.Error("Expected cache to be empty after clearing")
	}
}

func TestInvalidURL(t *testing.T) {
	checker := NewRobotsChecker("TestBot/1.0", slog.Default())

	_, err := checker.IsAllowed("not-a-valid-url")
	if err == nil {
		t.Error("Expected error for invalid URL")
	}

	_, err = checker.GetCrawlDelay("not-a-valid-url")
	if err == nil {
		t.Error("Expected error for invalid URL")
	}
}

func TestRobotsParsingEdgeCases(t *testing.T) {
	robotsContent := `# This is a comment
User-agent: TestBot
Disallow:

User-agent: *
Disallow: /admin

# Another comment
Invalid-line-without-colon

User-agent: OtherBot
Allow: /special
`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/robots.txt" {
			fmt.Fprint(w, robotsContent)
		}
	}))
	defer server.Close()

	checker := NewRobotsChecker("TestBot/1.0", slog.Default())

	robotsData, err := checker.fetchRobots(server.URL)
	if err != nil {
		t.Fatalf("fetchRobots failed: %v", err)
	}

	// Should have parsed valid rules and ignored invalid ones
	if len(robotsData.rules) == 0 {
		t.Error("Expected some rules to be parsed")
	}
}

func BenchmarkIsAllowed(b *testing.B) {
	robotsContent := `User-agent: TestBot
Disallow: /admin/
Allow: /admin/public/
Disallow: /private/*
`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/robots.txt" {
			fmt.Fprint(w, robotsContent)
		}
	}))
	defer server.Close()

	checker := NewRobotsChecker("TestBot/1.0", slog.Default())
	testURL := server.URL + "/test-path"

	// Pre-populate cache
	checker.IsAllowed(testURL)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		checker.IsAllowed(testURL)
	}
}
