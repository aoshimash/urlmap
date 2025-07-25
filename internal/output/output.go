package output

import (
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"sort"
	"time"
)

// OutputFormat represents the supported output formats
type OutputFormat string

const (
	FormatText OutputFormat = "text"
	FormatJSON OutputFormat = "json"
	FormatCSV  OutputFormat = "csv"
	FormatXML  OutputFormat = "xml"
)

// OutputConfig holds configuration for output formatting
type OutputConfig struct {
	Format OutputFormat
}

// URLResult represents a single URL result with metadata
type URLResult struct {
	URL       string    `json:"url" xml:"url"`
	Timestamp time.Time `json:"timestamp" xml:"timestamp"`
	Depth     int       `json:"depth,omitempty" xml:"depth,omitempty"`
}

// CrawlOutput represents the complete crawl output
type CrawlOutput struct {
	URLs      []URLResult `json:"urls" xml:"urls>url"`
	Timestamp time.Time   `json:"timestamp" xml:"timestamp"`
	Total     int         `json:"total" xml:"total"`
}

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

// OutputURLsWithFormat outputs URLs in the specified format
func OutputURLsWithFormat(urls []string, config *OutputConfig) error {
	if config == nil {
		config = &OutputConfig{Format: FormatText}
	}

	switch config.Format {
	case FormatJSON:
		return outputJSON(urls)
	case FormatCSV:
		return outputCSV(urls)
	case FormatXML:
		return outputXML(urls)
	case FormatText:
		fallthrough
	default:
		return OutputURLs(urls)
	}
}

// outputJSON outputs URLs in JSON format
func outputJSON(urls []string) error {
	uniqueURLs := removeDuplicates(urls)
	sort.Strings(uniqueURLs)

	urlResults := make([]URLResult, len(uniqueURLs))
	timestamp := time.Now()

	for i, url := range uniqueURLs {
		urlResults[i] = URLResult{
			URL:       url,
			Timestamp: timestamp,
		}
	}

	output := CrawlOutput{
		URLs:      urlResults,
		Timestamp: timestamp,
		Total:     len(urlResults),
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

// outputCSV outputs URLs in CSV format
func outputCSV(urls []string) error {
	uniqueURLs := removeDuplicates(urls)
	sort.Strings(uniqueURLs)

	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	// Write header
	if err := writer.Write([]string{"url", "timestamp"}); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	timestamp := time.Now().Format(time.RFC3339)

	// Write data
	for _, url := range uniqueURLs {
		if err := writer.Write([]string{url, timestamp}); err != nil {
			return fmt.Errorf("failed to write CSV record: %w", err)
		}
	}

	return nil
}

// outputXML outputs URLs in XML format
func outputXML(urls []string) error {
	uniqueURLs := removeDuplicates(urls)
	sort.Strings(uniqueURLs)

	urlResults := make([]URLResult, len(uniqueURLs))
	timestamp := time.Now()

	for i, url := range uniqueURLs {
		urlResults[i] = URLResult{
			URL:       url,
			Timestamp: timestamp,
		}
	}

	output := CrawlOutput{
		URLs:      urlResults,
		Timestamp: timestamp,
		Total:     len(urlResults),
	}

	xmlData, err := xml.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal XML: %w", err)
	}

	fmt.Print(xml.Header)
	fmt.Println(string(xmlData))
	return nil
}
