package output

import (
	"fmt"
	"os"
	"sort"
)

// OutputURLs outputs URLs in plain text format to stdout
// URLs are deduplicated, sorted alphabetically, and output one per line
func OutputURLs(urls []string) error {
	// Remove duplicates
	uniqueURLs := removeDuplicates(urls)

	// Sort alphabetically
	sort.Strings(uniqueURLs)

	// Output to stdout (one per line)
	for _, url := range uniqueURLs {
		fmt.Println(url)
	}

	return nil
}

// removeDuplicates removes duplicate URLs from the slice
func removeDuplicates(urls []string) []string {
	// Handle empty input
	if len(urls) == 0 {
		return []string{}
	}

	// Use map to track unique URLs
	urlMap := make(map[string]bool)
	uniqueURLs := make([]string, 0)

	for _, url := range urls {
		if !urlMap[url] {
			urlMap[url] = true
			uniqueURLs = append(uniqueURLs, url)
		}
	}

	return uniqueURLs
}

// OutputURLsToFile outputs URLs to a specified file
// This is useful for testing and when output redirection is needed
func OutputURLsToFile(urls []string, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	// Remove duplicates and sort
	uniqueURLs := removeDuplicates(urls)
	sort.Strings(uniqueURLs)

	// Write to file
	for _, url := range uniqueURLs {
		if _, err := fmt.Fprintln(file, url); err != nil {
			return fmt.Errorf("failed to write URL to file: %w", err)
		}
	}

	return nil
}

// GetUniqueURLs returns deduplicated and sorted URLs without outputting them
// This is useful for testing and when you need the processed URLs for other operations
func GetUniqueURLs(urls []string) []string {
	uniqueURLs := removeDuplicates(urls)
	sort.Strings(uniqueURLs)
	return uniqueURLs
}
