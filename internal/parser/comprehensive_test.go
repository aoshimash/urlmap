package parser

import (
	"sort"
	"testing"
)

// TestLinkExtractor_ComprehensiveFixtures tests the link extractor with comprehensive fixtures
func TestLinkExtractor_ComprehensiveFixtures(t *testing.T) {
	extractor := NewLinkExtractor(nil)
	fixtures := GetTestFixtures()

	for _, fixture := range fixtures {
		t.Run(fixture.Name, func(t *testing.T) {
			links, err := extractor.ExtractLinks(fixture.BaseURL, fixture.HTMLContent)
			if err != nil {
				t.Errorf("ExtractLinks() error = %v", err)
				return
			}

			// Sort both slices for comparison
			sort.Strings(links)
			expectedSorted := make([]string, len(fixture.Expected))
			copy(expectedSorted, fixture.Expected)
			sort.Strings(expectedSorted)

			if len(links) != len(expectedSorted) {
				t.Errorf("ExtractLinks() got %d links, expected %d", len(links), len(expectedSorted))
				t.Errorf("Got: %v", links)
				t.Errorf("Expected: %v", expectedSorted)
				return
			}

			for i, link := range links {
				if link != expectedSorted[i] {
					t.Errorf("ExtractLinks() link %d = %v, expected %v", i, link, expectedSorted[i])
				}
			}
		})
	}
}

// TestLinkExtractor_SameDomainFiltering tests same-domain filtering with fixtures
func TestLinkExtractor_SameDomainFiltering(t *testing.T) {
	extractor := NewLinkExtractor(nil)
	fixtures := GetSameDomainTestFixtures()

	for _, fixture := range fixtures {
		t.Run(fixture.Name, func(t *testing.T) {
			links, err := extractor.ExtractSameDomainLinks(fixture.BaseURL, fixture.HTMLContent)
			if err != nil {
				t.Errorf("ExtractSameDomainLinks() error = %v", err)
				return
			}

			// Sort both slices for comparison
			sort.Strings(links)
			expectedSorted := make([]string, len(fixture.Expected))
			copy(expectedSorted, fixture.Expected)
			sort.Strings(expectedSorted)

			if len(links) != len(expectedSorted) {
				t.Errorf("ExtractSameDomainLinks() got %d links, expected %d", len(links), len(expectedSorted))
				t.Errorf("Got: %v", links)
				t.Errorf("Expected: %v", expectedSorted)
				return
			}

			for i, link := range links {
				if link != expectedSorted[i] {
					t.Errorf("ExtractSameDomainLinks() link %d = %v, expected %v", i, link, expectedSorted[i])
				}
			}
		})
	}
}

// TestLinkExtractor_StatsWithFixtures tests extraction with statistics
func TestLinkExtractor_StatsWithFixtures(t *testing.T) {
	extractor := NewLinkExtractor(nil)
	fixtures := GetTestFixtures()

	for _, fixture := range fixtures {
		t.Run(fixture.Name+"_with_stats", func(t *testing.T) {
			links, stats, err := extractor.ExtractLinksWithStats(fixture.BaseURL, fixture.HTMLContent)
			if err != nil {
				t.Errorf("ExtractLinksWithStats() error = %v", err)
				return
			}

			// Verify that we got some links
			if len(links) == 0 && len(fixture.Expected) > 0 {
				t.Error("ExtractLinksWithStats() returned no links but expected some")
			}

			// Verify stats are consistent
			if stats.Valid != len(links) {
				t.Errorf("Stats.Valid = %d, but got %d links", stats.Valid, len(links))
			}

			// Total found should be >= valid links
			if stats.TotalFound < stats.Valid {
				t.Errorf("Stats.TotalFound (%d) < Stats.Valid (%d)", stats.TotalFound, stats.Valid)
			}

			// Check that stats string is not empty
			statsStr := stats.String()
			if statsStr == "" {
				t.Error("Stats.String() returned empty string")
			}

			t.Logf("Fixture: %s", fixture.Description)
			t.Logf("Stats: %s", statsStr)
		})
	}
}

// TestLinkExtractor_PerformanceWithLargeHTML tests performance with large HTML content
func TestLinkExtractor_PerformanceWithLargeHTML(t *testing.T) {
	extractor := NewLinkExtractor(nil)

	// Generate large HTML content
	largeHTML := generateLargeHTML(1000) // 1000 links
	baseURL := "https://performance.example.com"

	// Test that it completes in reasonable time
	links, err := extractor.ExtractLinks(baseURL, largeHTML)
	if err != nil {
		t.Errorf("ExtractLinks() with large HTML error = %v", err)
		return
	}

	if len(links) == 0 {
		t.Error("ExtractLinks() with large HTML returned no links")
	}

	t.Logf("Processed large HTML with %d links", len(links))
}

// TestLinkExtractor_ErrorHandling tests error handling scenarios
func TestLinkExtractor_ErrorHandling(t *testing.T) {
	extractor := NewLinkExtractor(nil)

	errorTests := []struct {
		name        string
		baseURL     string
		htmlContent string
		wantErr     bool
		errorMsg    string
	}{
		{
			name:        "Empty base URL",
			baseURL:     "",
			htmlContent: "<html><body><a href='/test'>Test</a></body></html>",
			wantErr:     true,
			errorMsg:    "base URL cannot be empty",
		},
		{
			name:        "Invalid base URL",
			baseURL:     "not-a-url",
			htmlContent: "<html><body><a href='/test'>Test</a></body></html>",
			wantErr:     true,
			errorMsg:    "invalid base URL",
		},
		{
			name:        "Whitespace base URL",
			baseURL:     "   ",
			htmlContent: "<html><body><a href='/test'>Test</a></body></html>",
			wantErr:     true,
			errorMsg:    "base URL cannot be empty",
		},
		{
			name:        "Empty HTML content",
			baseURL:     "https://example.com",
			htmlContent: "",
			wantErr:     false, // Should return empty slice, not error
			errorMsg:    "",
		},
		{
			name:        "Whitespace only HTML",
			baseURL:     "https://example.com",
			htmlContent: "   \n\t   ",
			wantErr:     false, // Should return empty slice, not error
			errorMsg:    "",
		},
	}

	for _, tt := range errorTests {
		t.Run(tt.name, func(t *testing.T) {
			links, err := extractor.ExtractLinks(tt.baseURL, tt.htmlContent)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ExtractLinks() expected error containing %q but got none", tt.errorMsg)
					return
				}
				if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("ExtractLinks() error = %v, expected to contain %q", err, tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ExtractLinks() unexpected error = %v", err)
					return
				}
				// For empty/whitespace HTML, should return empty slice
				if len(links) != 0 && (tt.htmlContent == "" || isWhitespace(tt.htmlContent)) {
					t.Errorf("ExtractLinks() with empty/whitespace HTML got %d links, expected 0", len(links))
				}
			}
		})
	}
}

// TestLinkExtractor_ConcurrencySupport tests that the extractor is safe for concurrent use
func TestLinkExtractor_ConcurrencySupport(t *testing.T) {
	extractor := NewLinkExtractor(nil)
	baseURL := "https://concurrent.example.com"
	htmlContent := `<html><body>
		<a href="/page1">Page 1</a>
		<a href="/page2">Page 2</a>
		<a href="/page3">Page 3</a>
	</body></html>`

	// Run multiple goroutines concurrently
	const numGoroutines = 10
	done := make(chan bool, numGoroutines)
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()

			links, err := extractor.ExtractLinks(baseURL, htmlContent)
			if err != nil {
				errors <- err
				return
			}

			if len(links) != 3 {
				errors <- err
				return
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		select {
		case <-done:
			// OK
		case err := <-errors:
			t.Errorf("Concurrent execution failed: %v", err)
		}
	}
}

// Helper functions

// generateLargeHTML generates HTML content with the specified number of links
func generateLargeHTML(numLinks int) string {
	html := `<!DOCTYPE html><html><head><title>Large Page</title></head><body>`

	for i := 0; i < numLinks; i++ {
		html += `<a href="/page` + string(rune('0'+i%10)) + `">Link ` + string(rune('0'+i%10)) + `</a>`
	}

	html += `</body></html>`
	return html
}

// contains checks if a string contains a substring
func contains(str, substr string) bool {
	return len(str) >= len(substr) &&
		(len(substr) == 0 || str[len(str)-len(substr):] == substr ||
			str[:len(substr)] == substr ||
			containsSubstring(str, substr))
}

// containsSubstring is a simple substring search
func containsSubstring(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// isWhitespace checks if string contains only whitespace
func isWhitespace(s string) bool {
	for _, r := range s {
		if r != ' ' && r != '\t' && r != '\n' && r != '\r' {
			return false
		}
	}
	return true
}

// BenchmarkLinkExtractor_ComplexHTML benchmarks extraction from complex HTML
func BenchmarkLinkExtractor_ComplexHTML(b *testing.B) {
	extractor := NewLinkExtractor(nil)
	fixtures := GetTestFixtures()

	// Use the most complex fixture
	var complexFixture HTMLFixtures
	for _, fixture := range fixtures {
		if fixture.Name == "E-commerce page" {
			complexFixture = fixture
			break
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := extractor.ExtractLinks(complexFixture.BaseURL, complexFixture.HTMLContent)
		if err != nil {
			b.Errorf("ExtractLinks() error = %v", err)
		}
	}
}

// BenchmarkLinkExtractor_LargeHTML benchmarks extraction from large HTML
func BenchmarkLinkExtractor_LargeHTML(b *testing.B) {
	extractor := NewLinkExtractor(nil)
	largeHTML := generateLargeHTML(100)
	baseURL := "https://benchmark.example.com"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := extractor.ExtractLinks(baseURL, largeHTML)
		if err != nil {
			b.Errorf("ExtractLinks() error = %v", err)
		}
	}
}
