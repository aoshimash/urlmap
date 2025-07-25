package output

import (
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestRemoveDuplicates(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "no duplicates",
			input:    []string{"http://example.com", "http://test.com"},
			expected: []string{"http://example.com", "http://test.com"},
		},
		{
			name:     "with duplicates",
			input:    []string{"http://example.com", "http://test.com", "http://example.com"},
			expected: []string{"http://example.com", "http://test.com"},
		},
		{
			name:     "empty slice",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "all duplicates",
			input:    []string{"http://example.com", "http://example.com", "http://example.com"},
			expected: []string{"http://example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeDuplicates(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("removeDuplicates() returned %d URLs, expected %d", len(result), len(tt.expected))
				return
			}

			// Convert to sets for comparison (order doesn't matter for this function)
			resultSet := make(map[string]bool)
			expectedSet := make(map[string]bool)

			for _, url := range result {
				resultSet[url] = true
			}
			for _, url := range tt.expected {
				expectedSet[url] = true
			}

			if !reflect.DeepEqual(resultSet, expectedSet) {
				t.Errorf("removeDuplicates() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestGetUniqueURLs(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "unsorted URLs with duplicates",
			input:    []string{"http://z.com", "http://a.com", "http://z.com", "http://b.com"},
			expected: []string{"http://a.com", "http://b.com", "http://z.com"},
		},
		{
			name:     "already sorted URLs",
			input:    []string{"http://a.com", "http://b.com", "http://c.com"},
			expected: []string{"http://a.com", "http://b.com", "http://c.com"},
		},
		{
			name:     "empty slice",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "single URL",
			input:    []string{"http://example.com"},
			expected: []string{"http://example.com"},
		},
		{
			name:     "case sensitivity",
			input:    []string{"http://Example.com", "http://example.com"},
			expected: []string{"http://Example.com", "http://example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetUniqueURLs(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("GetUniqueURLs() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestOutputURLsToFile(t *testing.T) {
	// Create a temporary file for testing
	tmpFile := "test_output.txt"
	defer os.Remove(tmpFile)

	testURLs := []string{
		"http://z.com",
		"http://a.com",
		"http://z.com", // duplicate
		"http://b.com",
	}

	// Test file output
	err := OutputURLsToFile(testURLs, tmpFile)
	if err != nil {
		t.Fatalf("OutputURLsToFile() failed: %v", err)
	}

	// Read back the file content
	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	// Verify content
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	expected := []string{"http://a.com", "http://b.com", "http://z.com"}

	if !reflect.DeepEqual(lines, expected) {
		t.Errorf("File content = %v, expected %v", lines, expected)
	}
}

func TestOutputURLsToFileError(t *testing.T) {
	// Test with invalid file path
	err := OutputURLsToFile([]string{"http://example.com"}, "/invalid/path/file.txt")
	if err == nil {
		t.Error("OutputURLsToFile() should have failed with invalid path")
	}
}

// BenchmarkRemoveDuplicates benchmarks the performance of duplicate removal
func BenchmarkRemoveDuplicates(b *testing.B) {
	// Create test data with many duplicates
	urls := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		// Create URLs with many duplicates
		urls[i] = "http://example" + string(rune(i%100)) + ".com"
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		removeDuplicates(urls)
	}
}

// BenchmarkGetUniqueURLs benchmarks the complete process of deduplication and sorting
func BenchmarkGetUniqueURLs(b *testing.B) {
	// Create test data
	urls := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		urls[i] = "http://example" + string(rune(i%100)) + ".com"
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetUniqueURLs(urls)
	}
}

// TestLargeURLSet tests with a large number of URLs to ensure efficiency
func TestLargeURLSet(t *testing.T) {
	// Create a large set of URLs with duplicates
	urls := make([]string, 10000)
	for i := 0; i < 10000; i++ {
		// Create patterns that will result in duplicates
		urls[i] = "http://example" + string(rune(i%1000)) + ".com"
	}

	result := GetUniqueURLs(urls)

	// Should have exactly 1000 unique URLs
	if len(result) != 1000 {
		t.Errorf("Expected 1000 unique URLs, got %d", len(result))
	}

	// Verify sorting
	for i := 1; i < len(result); i++ {
		if result[i-1] >= result[i] {
			t.Errorf("URLs not properly sorted: %s >= %s", result[i-1], result[i])
		}
	}
}

func TestOutputURLsWithFormat(t *testing.T) {
	testURLs := []string{
		"https://example.com/page1",
		"https://example.com/page2",
		"https://example.com/page1", // duplicate
	}

	tests := []struct {
		name           string
		config         *OutputConfig
		expectedFormat string
		shouldError    bool
	}{
		{
			name:           "text format",
			config:         &OutputConfig{Format: FormatText},
			expectedFormat: "text",
			shouldError:    false,
		},
		{
			name:           "json format",
			config:         &OutputConfig{Format: FormatJSON},
			expectedFormat: "json",
			shouldError:    false,
		},
		{
			name:           "csv format",
			config:         &OutputConfig{Format: FormatCSV},
			expectedFormat: "csv",
			shouldError:    false,
		},
		{
			name:           "xml format",
			config:         &OutputConfig{Format: FormatXML},
			expectedFormat: "xml",
			shouldError:    false,
		},
		{
			name:           "nil config defaults to text",
			config:         nil,
			expectedFormat: "text",
			shouldError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We can't easily test stdout output without complex setup,
			// so we'll just verify the function doesn't error
			err := OutputURLsWithFormat(testURLs, tt.config)
			if tt.shouldError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestOutputJSON(t *testing.T) {
	testURLs := []string{
		"https://example.com/page2",
		"https://example.com/page1",
		"https://example.com/page1", // duplicate
	}

	// Test that the function doesn't error
	err := outputJSON(testURLs)
	if err != nil {
		t.Errorf("outputJSON() returned error: %v", err)
	}
}

func TestOutputCSV(t *testing.T) {
	testURLs := []string{
		"https://example.com/page2",
		"https://example.com/page1",
		"https://example.com/page1", // duplicate
	}

	// Test that the function doesn't error
	err := outputCSV(testURLs)
	if err != nil {
		t.Errorf("outputCSV() returned error: %v", err)
	}
}

func TestOutputXML(t *testing.T) {
	testURLs := []string{
		"https://example.com/page2",
		"https://example.com/page1",
		"https://example.com/page1", // duplicate
	}

	// Test that the function doesn't error
	err := outputXML(testURLs)
	if err != nil {
		t.Errorf("outputXML() returned error: %v", err)
	}
}
