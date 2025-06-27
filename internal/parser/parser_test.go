package parser

import (
	"log/slog"
	"os"
	"testing"
)

const (
	testBaseURL = "https://example.com"
)

func TestMain(m *testing.M) {
	// Disable logs during testing
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelError, // Only show errors
	})))
	os.Exit(m.Run())
}

func TestLinkExtractor_ExtractLinks(t *testing.T) {
	extractor := NewLinkExtractor(nil)

	tests := []struct {
		name        string
		baseURL     string
		htmlContent string
		expected    []string
		expectError bool
	}{
		{
			name:        "Empty base URL",
			baseURL:     "",
			htmlContent: `<html><body><a href="/test">Test</a></body></html>`,
			expectError: true,
		},
		{
			name:        "Invalid base URL",
			baseURL:     "not-a-url",
			htmlContent: `<html><body><a href="/test">Test</a></body></html>`,
			expectError: true,
		},
		{
			name:        "Empty HTML content",
			baseURL:     testBaseURL,
			htmlContent: "",
			expected:    []string{},
			expectError: false,
		},
		{
			name:        "HTML with no links",
			baseURL:     testBaseURL,
			htmlContent: `<html><body><p>No links here</p></body></html>`,
			expected:    []string{},
			expectError: false,
		},
		{
			name:    "Single absolute link",
			baseURL: testBaseURL,
			htmlContent: `<html><body>
				<a href="https://example.com/page1">Page 1</a>
			</body></html>`,
			expected:    []string{"https://example.com/page1"},
			expectError: false,
		},
		{
			name:    "Single relative link",
			baseURL: testBaseURL,
			htmlContent: `<html><body>
				<a href="/page1">Page 1</a>
			</body></html>`,
			expected:    []string{"https://example.com/page1"},
			expectError: false,
		},
		{
			name:    "Multiple mixed links",
			baseURL: testBaseURL,
			htmlContent: `<html><body>
				<a href="https://example.com/absolute">Absolute</a>
				<a href="/relative">Relative</a>
				<a href="./relative2">Relative2</a>
				<a href="../parent">Parent</a>
			</body></html>`,
			expected: []string{
				"https://example.com/absolute",
				"https://example.com/relative",
				"https://example.com/relative2",
				"https://example.com/parent",
			},
			expectError: false,
		},
		{
			name:    "Links with fragments - fragments should be removed",
			baseURL: testBaseURL,
			htmlContent: `<html><body>
				<a href="https://example.com/page#section">With Fragment</a>
				<a href="/relative#fragment">Relative with Fragment</a>
			</body></html>`,
			expected: []string{
				"https://example.com/page",
				"https://example.com/relative",
			},
			expectError: false,
		},
		{
			name:    "Filter out invalid schemes",
			baseURL: testBaseURL,
			htmlContent: `<html><body>
				<a href="javascript:void(0)">JavaScript</a>
				<a href="mailto:test@example.com">Email</a>
				<a href="tel:+1234567890">Phone</a>
				<a href="ftp://example.com/file">FTP</a>
				<a href="https://example.com/valid">Valid</a>
			</body></html>`,
			expected:    []string{"https://example.com/valid"},
			expectError: false,
		},
		{
			name:    "Filter out fragment-only links",
			baseURL: testBaseURL,
			htmlContent: `<html><body>
				<a href="#section1">Section 1</a>
				<a href="#section2">Section 2</a>
				<a href="https://example.com/valid">Valid</a>
			</body></html>`,
			expected:    []string{"https://example.com/valid"},
			expectError: false,
		},
		{
			name:    "Empty href attributes",
			baseURL: testBaseURL,
			htmlContent: `<html><body>
				<a href="">Empty</a>
				<a href="   ">Whitespace Only</a>
				<a href="https://example.com/valid">Valid</a>
			</body></html>`,
			expected:    []string{"https://example.com/valid"},
			expectError: false,
		},
		{
			name:    "Anchor tags without href",
			baseURL: testBaseURL,
			htmlContent: `<html><body>
				<a name="anchor">Named Anchor</a>
				<a id="target">ID Target</a>
				<a href="https://example.com/valid">Valid</a>
			</body></html>`,
			expected:    []string{"https://example.com/valid"},
			expectError: false,
		},
		{
			name:    "Complex HTML structure",
			baseURL: testBaseURL,
			htmlContent: `<html>
			<head><title>Test Page</title></head>
			<body>
				<nav>
					<a href="/home">Home</a>
					<a href="/about">About</a>
				</nav>
				<main>
					<article>
						<h1>Article Title</h1>
						<p>Some text with <a href="https://external.com">external link</a></p>
						<p>And <a href="/internal">internal link</a></p>
					</article>
				</main>
				<footer>
					<a href="/contact">Contact</a>
					<a href="mailto:contact@example.com">Email</a>
				</footer>
			</body>
			</html>`,
			expected: []string{
				"https://example.com/home",
				"https://example.com/about",
				"https://external.com/",
				"https://example.com/internal",
				"https://example.com/contact",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := extractor.ExtractLinks(tt.baseURL, tt.htmlContent)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d links, got %d", len(tt.expected), len(result))
				t.Errorf("Expected: %v", tt.expected)
				t.Errorf("Got: %v", result)
				return
			}

			// Convert to map for easier comparison (order doesn't matter)
			expectedMap := make(map[string]bool)
			for _, link := range tt.expected {
				expectedMap[link] = true
			}

			for _, link := range result {
				if !expectedMap[link] {
					t.Errorf("Unexpected link in result: %s", link)
				}
			}
		})
	}
}

func TestLinkExtractor_ExtractSameDomainLinks(t *testing.T) {
	extractor := NewLinkExtractor(nil)

	tests := []struct {
		name        string
		baseURL     string
		htmlContent string
		expected    []string
		expectError bool
	}{
		{
			name:    "Mixed domain links",
			baseURL: testBaseURL,
			htmlContent: `<html><body>
				<a href="https://example.com/page1">Same Domain 1</a>
				<a href="https://other.com/page2">Other Domain</a>
				<a href="/relative">Relative (Same Domain)</a>
				<a href="https://example.com/page2">Same Domain 2</a>
			</body></html>`,
			expected: []string{
				"https://example.com/page1",
				"https://example.com/relative",
				"https://example.com/page2",
			},
			expectError: false,
		},
		{
			name:    "All external links",
			baseURL: testBaseURL,
			htmlContent: `<html><body>
				<a href="https://other1.com/page1">External 1</a>
				<a href="https://other2.com/page2">External 2</a>
			</body></html>`,
			expected:    []string{},
			expectError: false,
		},
		{
			name:    "All same domain links",
			baseURL: testBaseURL,
			htmlContent: `<html><body>
				<a href="https://example.com/page1">Same 1</a>
				<a href="/page2">Same 2</a>
				<a href="./page3">Same 3</a>
			</body></html>`,
			expected: []string{
				"https://example.com/page1",
				"https://example.com/page2",
				"https://example.com/page3",
			},
			expectError: false,
		},
		{
			name:    "Subdomain handling",
			baseURL: "https://www.example.com",
			htmlContent: `<html><body>
				<a href="https://www.example.com/page1">WWW Same</a>
				<a href="https://example.com/page2">Non-WWW Different</a>
				<a href="https://sub.example.com/page3">Subdomain Different</a>
			</body></html>`,
			expected: []string{
				"https://www.example.com/page1",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := extractor.ExtractSameDomainLinks(tt.baseURL, tt.htmlContent)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d links, got %d", len(tt.expected), len(result))
				t.Errorf("Expected: %v", tt.expected)
				t.Errorf("Got: %v", result)
				return
			}

			// Convert to map for easier comparison
			expectedMap := make(map[string]bool)
			for _, link := range tt.expected {
				expectedMap[link] = true
			}

			for _, link := range result {
				if !expectedMap[link] {
					t.Errorf("Unexpected link in result: %s", link)
				}
			}
		})
	}
}

func TestLinkExtractor_ExtractLinksWithStats(t *testing.T) {
	extractor := NewLinkExtractor(nil)

	htmlContent := `<html><body>
		<a href="https://example.com/valid1">Valid 1</a>
		<a href="/valid2">Valid 2 (relative)</a>
		<a href="">Empty href</a>
		<a href="javascript:void(0)">JavaScript</a>
		<a href="mailto:test@example.com">Email</a>
		<a href="#fragment">Fragment only</a>
		<a name="anchor">No href</a>
		<a href="   ">Whitespace only</a>
		<a href="invalid-url">Invalid URL</a>
		<a href="https://example.com/valid3">Valid 3</a>
	</body></html>`

	links, stats, err := extractor.ExtractLinksWithStats(testBaseURL, htmlContent)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	// Verify stats
	expectedStats := &ExtractionStats{
		TotalFound:          9, // 9 anchor tags with href attribute
		Valid:               4, // 4 valid links (including resolved "invalid-url")
		EmptyHrefs:          2, // Empty and whitespace-only hrefs
		FilteredOut:         3, // javascript:, mailto:, #fragment
		RelativeURLs:        2, // /valid2 and invalid-url (treated as relative)
		ResolutionErrors:    0,
		InvalidURLs:         0, // "invalid-url" resolves to valid URL
		NormalizationErrors: 0,
	}

	if stats.TotalFound != expectedStats.TotalFound {
		t.Errorf("Expected TotalFound %d, got %d", expectedStats.TotalFound, stats.TotalFound)
	}
	if stats.Valid != expectedStats.Valid {
		t.Errorf("Expected Valid %d, got %d", expectedStats.Valid, stats.Valid)
	}
	if stats.EmptyHrefs != expectedStats.EmptyHrefs {
		t.Errorf("Expected EmptyHrefs %d, got %d", expectedStats.EmptyHrefs, stats.EmptyHrefs)
	}
	if stats.FilteredOut != expectedStats.FilteredOut {
		t.Errorf("Expected FilteredOut %d, got %d", expectedStats.FilteredOut, stats.FilteredOut)
	}
	if stats.RelativeURLs != expectedStats.RelativeURLs {
		t.Errorf("Expected RelativeURLs %d, got %d", expectedStats.RelativeURLs, stats.RelativeURLs)
	}

	// Verify links
	expectedLinks := []string{
		"https://example.com/valid1",
		"https://example.com/valid2",
		"https://example.com/invalid-url", // This becomes a valid URL when resolved
		"https://example.com/valid3",
	}

	if len(links) != len(expectedLinks) {
		t.Errorf("Expected %d links, got %d", len(expectedLinks), len(links))
		t.Errorf("Expected: %v", expectedLinks)
		t.Errorf("Got: %v", links)
	}
}

func TestNewLinkExtractor(t *testing.T) {
	// Test with nil logger
	extractor1 := NewLinkExtractor(nil)
	if extractor1 == nil {
		t.Error("Expected non-nil extractor")
	}
	if extractor1.logger == nil {
		t.Error("Expected non-nil logger")
	}

	// Test with custom logger
	customLogger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	extractor2 := NewLinkExtractor(customLogger)
	if extractor2 == nil {
		t.Error("Expected non-nil extractor")
	}
	if extractor2.logger != customLogger {
		t.Error("Expected custom logger to be used")
	}
}

func TestExtractionStats_String(t *testing.T) {
	stats := &ExtractionStats{
		TotalFound:          10,
		Valid:               5,
		EmptyHrefs:          2,
		FilteredOut:         2,
		RelativeURLs:        3,
		ResolutionErrors:    1,
		InvalidURLs:         1,
		NormalizationErrors: 0,
	}

	result := stats.String()
	expected := "ExtractionStats{Total: 10, Valid: 5, Empty: 2, Filtered: 2, Relative: 3, ResolutionErr: 1, Invalid: 1, NormalizationErr: 0}"

	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestLinkExtractor_MalformedHTML(t *testing.T) {
	extractor := NewLinkExtractor(nil)

	tests := []struct {
		name        string
		htmlContent string
		expectError bool
	}{
		{
			name:        "Unclosed tags",
			htmlContent: `<html><body><a href="/test">Test`,
			expectError: false, // goquery handles malformed HTML gracefully
		},
		{
			name:        "Invalid nested tags",
			htmlContent: `<a href="/outer"><a href="/inner">Nested</a></a>`,
			expectError: false,
		},
		{
			name:        "Mixed case attributes",
			htmlContent: `<A HREF="/test">Test</A>`,
			expectError: false,
		},
		{
			name:        "Attributes without quotes",
			htmlContent: `<a href=/test>Test</a>`,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := extractor.ExtractLinks(testBaseURL, tt.htmlContent)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// Benchmark tests
func BenchmarkLinkExtractor_ExtractLinks(b *testing.B) {
	extractor := NewLinkExtractor(nil)
	htmlContent := `<html><body>`
	for i := 0; i < 100; i++ {
		htmlContent += `<a href="/page` + string(rune(i)) + `">Page ` + string(rune(i)) + `</a>`
	}
	htmlContent += `</body></html>`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := extractor.ExtractLinks(testBaseURL, htmlContent)
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}
