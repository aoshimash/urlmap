package parser

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/aoshimash/urlmap/internal/client"
	"github.com/aoshimash/urlmap/internal/url"
)

// LinkExtractor provides functionality to extract and filter links from HTML content
type LinkExtractor struct {
	logger *slog.Logger
	client *client.UnifiedClient
}

// NewLinkExtractor creates a new LinkExtractor instance
func NewLinkExtractor(logger *slog.Logger) *LinkExtractor {
	if logger == nil {
		logger = slog.Default()
	}
	return &LinkExtractor{
		logger: logger,
		client: nil, // Will be set when needed
	}
}

// NewLinkExtractorWithClient creates a new LinkExtractor instance with a unified client
func NewLinkExtractorWithClient(logger *slog.Logger, unifiedClient *client.UnifiedClient) *LinkExtractor {
	if logger == nil {
		logger = slog.Default()
	}
	return &LinkExtractor{
		logger: logger,
		client: unifiedClient,
	}
}

// ExtractLinksFromURL fetches content from URL and extracts links using the unified client
// This method supports both HTTP and JavaScript rendering based on client configuration
func (le *LinkExtractor) ExtractLinksFromURL(ctx context.Context, targetURL string) ([]string, error) {
	if le.client == nil {
		return nil, fmt.Errorf("unified client not configured - use NewLinkExtractorWithClient or call ExtractLinks directly")
	}

	if targetURL = strings.TrimSpace(targetURL); targetURL == "" {
		return nil, fmt.Errorf("target URL cannot be empty")
	}

	// Validate URL
	if !url.IsValidURL(targetURL) {
		return nil, fmt.Errorf("invalid target URL: %s", targetURL)
	}

	le.logger.Debug("Fetching content from URL",
		"url", targetURL,
		"js_enabled", le.client.IsJSEnabled())

	// Fetch content using unified client (HTTP or JS rendering)
	response, err := le.client.Get(ctx, targetURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch content from %s: %w", targetURL, err)
	}

	htmlContent := response.String()

	le.logger.Debug("Content fetched successfully",
		"url", targetURL,
		"content_length", len(htmlContent),
		"status_code", response.StatusCode())

	// Extract links from the fetched content
	return le.ExtractLinks(targetURL, htmlContent)
}

// ExtractLinksFromURLWithFallback fetches content with fallback support
// This method attempts JavaScript rendering first, then falls back to HTTP if enabled
func (le *LinkExtractor) ExtractLinksFromURLWithFallback(ctx context.Context, targetURL string) ([]string, error) {
	if le.client == nil {
		return nil, fmt.Errorf("unified client not configured - use NewLinkExtractorWithClient or call ExtractLinks directly")
	}

	if targetURL = strings.TrimSpace(targetURL); targetURL == "" {
		return nil, fmt.Errorf("target URL cannot be empty")
	}

	// Validate URL
	if !url.IsValidURL(targetURL) {
		return nil, fmt.Errorf("invalid target URL: %s", targetURL)
	}

	le.logger.Debug("Fetching content from URL with fallback",
		"url", targetURL,
		"js_enabled", le.client.IsJSEnabled())

	// Fetch content using unified client with fallback
	response, err := le.client.GetWithFallback(ctx, targetURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch content from %s: %w", targetURL, err)
	}

	htmlContent := response.String()

	le.logger.Debug("Content fetched successfully with fallback",
		"url", targetURL,
		"content_length", len(htmlContent),
		"status_code", response.StatusCode())

	// Extract links from the fetched content
	return le.ExtractLinks(targetURL, htmlContent)
}

// ExtractLinks extracts and filters links from HTML content
// baseURL is used to resolve relative URLs to absolute URLs
// htmlContent is the HTML content to parse
// Returns a slice of valid, filtered absolute URLs
func (le *LinkExtractor) ExtractLinks(baseURL, htmlContent string) ([]string, error) {
	if baseURL = strings.TrimSpace(baseURL); baseURL == "" {
		return nil, fmt.Errorf("base URL cannot be empty")
	}

	if htmlContent = strings.TrimSpace(htmlContent); htmlContent == "" {
		le.logger.Debug("Empty HTML content provided")
		return []string{}, nil
	}

	// Validate base URL
	if !url.IsValidURL(baseURL) {
		return nil, fmt.Errorf("invalid base URL: %s", baseURL)
	}

	// Parse HTML document
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML content: %w", err)
	}

	le.logger.Debug("Starting link extraction", "base_url", baseURL)

	var validLinks []string
	var totalFound int
	var validCount int

	// Extract all links from anchor tags with href attribute
	doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists {
			return
		}

		totalFound++
		href = strings.TrimSpace(href)

		// Skip empty hrefs
		if href == "" {
			le.logger.Debug("Skipping empty href")
			return
		}

		// Skip URLs that should be filtered out
		if url.ShouldSkipURL(href) {
			le.logger.Debug("Skipping filtered URL", "url", href)
			return
		}

		// Handle relative URLs - resolve them to absolute
		var absoluteURL string
		if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
			// Already absolute URL
			absoluteURL = href
		} else {
			// Relative URL - resolve against base URL
			resolved, err := url.ResolveURL(baseURL, href)
			if err != nil {
				le.logger.Debug("Failed to resolve relative URL", "href", href, "error", err)
				return
			}
			absoluteURL = resolved
		}

		// Validate the final URL
		if !url.IsValidURL(absoluteURL) {
			le.logger.Debug("Invalid URL after resolution", "url", absoluteURL)
			return
		}

		// Normalize the URL
		normalizedURL, err := url.NormalizeURL(absoluteURL)
		if err != nil {
			le.logger.Debug("Failed to normalize URL", "url", absoluteURL, "error", err)
			return
		}

		validLinks = append(validLinks, normalizedURL)
		validCount++
		le.logger.Debug("Added valid link", "url", normalizedURL)
	})

	le.logger.Info("Link extraction completed",
		"total_found", totalFound,
		"valid_count", validCount,
		"base_url", baseURL)

	return validLinks, nil
}

// ExtractSameDomainLinks extracts links that belong to the same domain as the base URL
func (le *LinkExtractor) ExtractSameDomainLinks(baseURL, htmlContent string) ([]string, error) {
	allLinks, err := le.ExtractLinks(baseURL, htmlContent)
	if err != nil {
		return nil, err
	}

	if len(allLinks) == 0 {
		return []string{}, nil
	}

	le.logger.Debug("Filtering links for same domain", "base_url", baseURL, "total_links", len(allLinks))

	var sameDomainLinks []string
	for _, link := range allLinks {
		isSame, err := url.IsSameDomain(baseURL, link)
		if err != nil {
			le.logger.Debug("Failed to check domain", "base_url", baseURL, "link", link, "error", err)
			continue
		}

		if isSame {
			sameDomainLinks = append(sameDomainLinks, link)
			le.logger.Debug("Added same-domain link", "url", link)
		} else {
			le.logger.Debug("Skipped external domain link", "url", link)
		}
	}

	le.logger.Info("Same-domain filtering completed",
		"base_url", baseURL,
		"total_links", len(allLinks),
		"same_domain_links", len(sameDomainLinks))

	return sameDomainLinks, nil
}

// ExtractLinksWithStats extracts links and returns statistics
func (le *LinkExtractor) ExtractLinksWithStats(baseURL, htmlContent string) ([]string, *ExtractionStats, error) {
	stats := &ExtractionStats{}

	if baseURL = strings.TrimSpace(baseURL); baseURL == "" {
		return nil, stats, fmt.Errorf("base URL cannot be empty")
	}

	if htmlContent = strings.TrimSpace(htmlContent); htmlContent == "" {
		return []string{}, stats, nil
	}

	// Parse HTML document
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, stats, fmt.Errorf("failed to parse HTML content: %w", err)
	}

	var validLinks []string

	// Extract all links from anchor tags with href attribute
	doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists {
			return
		}

		stats.TotalFound++
		href = strings.TrimSpace(href)

		// Skip empty hrefs
		if href == "" {
			stats.EmptyHrefs++
			return
		}

		// Skip URLs that should be filtered out
		if url.ShouldSkipURL(href) {
			stats.FilteredOut++
			return
		}

		// Handle relative URLs
		var absoluteURL string
		if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
			absoluteURL = href
		} else {
			stats.RelativeURLs++
			resolved, err := url.ResolveURL(baseURL, href)
			if err != nil {
				stats.ResolutionErrors++
				return
			}
			absoluteURL = resolved
		}

		// Validate the final URL
		if !url.IsValidURL(absoluteURL) {
			stats.InvalidURLs++
			return
		}

		// Normalize the URL
		normalizedURL, err := url.NormalizeURL(absoluteURL)
		if err != nil {
			stats.NormalizationErrors++
			return
		}

		validLinks = append(validLinks, normalizedURL)
		stats.Valid++
	})

	return validLinks, stats, nil
}

// ExtractionStats holds statistics about link extraction
type ExtractionStats struct {
	TotalFound          int // Total anchor tags with href found
	Valid               int // Valid links extracted
	EmptyHrefs          int // Empty href attributes
	FilteredOut         int // Links filtered out (javascript:, mailto:, etc.)
	RelativeURLs        int // Relative URLs that were resolved
	ResolutionErrors    int // Errors during relative URL resolution
	InvalidURLs         int // Invalid URLs after resolution
	NormalizationErrors int // Errors during URL normalization
}

// String returns a human-readable representation of the stats
func (s *ExtractionStats) String() string {
	return fmt.Sprintf("ExtractionStats{Total: %d, Valid: %d, Empty: %d, Filtered: %d, Relative: %d, ResolutionErr: %d, Invalid: %d, NormalizationErr: %d}",
		s.TotalFound, s.Valid, s.EmptyHrefs, s.FilteredOut, s.RelativeURLs, s.ResolutionErrors, s.InvalidURLs, s.NormalizationErrors)
}
