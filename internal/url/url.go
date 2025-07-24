package url

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

var (
	ErrInvalidURL        = errors.New("invalid URL")
	ErrUnsupportedScheme = errors.New("unsupported URL scheme")
	ErrEmptyURL          = errors.New("URL cannot be empty")
)

// IsValidURL validates if the given string is a valid HTTP/HTTPS URL
func IsValidURL(rawURL string) bool {
	if rawURL = strings.TrimSpace(rawURL); rawURL == "" {
		return false
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

	// Only accept HTTP and HTTPS protocols
	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return false
	}

	// Must have a host
	if parsed.Host == "" {
		return false
	}

	return true
}

// ExtractDomain extracts the domain/hostname from a URL
func ExtractDomain(rawURL string) (string, error) {
	if rawURL = strings.TrimSpace(rawURL); rawURL == "" {
		return "", ErrEmptyURL
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %w", err)
	}

	if parsed.Host == "" {
		return "", ErrInvalidURL
	}

	// Extract hostname without port
	hostname := parsed.Hostname()
	if hostname == "" {
		return "", ErrInvalidURL
	}

	return hostname, nil
}

// ResolveURL resolves a relative URL against a base URL to create an absolute URL
func ResolveURL(baseURL, relativeURL string) (string, error) {
	if baseURL = strings.TrimSpace(baseURL); baseURL == "" {
		return "", fmt.Errorf("base URL cannot be empty")
	}

	if relativeURL = strings.TrimSpace(relativeURL); relativeURL == "" {
		return "", fmt.Errorf("relative URL cannot be empty")
	}

	// Parse base URL
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse base URL: %w", err)
	}

	// Parse relative URL
	relative, err := url.Parse(relativeURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse relative URL: %w", err)
	}

	// Resolve relative URL against base URL
	resolved := base.ResolveReference(relative)

	return resolved.String(), nil
}

// NormalizeURL normalizes a URL by removing fragments and handling trailing slashes
func NormalizeURL(rawURL string) (string, error) {
	if rawURL = strings.TrimSpace(rawURL); rawURL == "" {
		return "", ErrEmptyURL
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %w", err)
	}

	// Remove fragment
	parsed.Fragment = ""

	// Normalize path - remove trailing slash for non-root paths
	if parsed.Path != "/" && strings.HasSuffix(parsed.Path, "/") {
		parsed.Path = strings.TrimSuffix(parsed.Path, "/")
	}

	// Ensure root path is "/"
	if parsed.Path == "" {
		parsed.Path = "/"
	}

	return parsed.String(), nil
}

// IsSameDomain checks if two URLs belong to the same domain
func IsSameDomain(url1, url2 string) (bool, error) {
	domain1, err := ExtractDomain(url1)
	if err != nil {
		return false, fmt.Errorf("failed to extract domain from first URL: %w", err)
	}

	domain2, err := ExtractDomain(url2)
	if err != nil {
		return false, fmt.Errorf("failed to extract domain from second URL: %w", err)
	}

	return strings.EqualFold(domain1, domain2), nil
}

// IsSamePathPrefix checks if targetURL is under the same path prefix as baseURL
// This function first checks if the URLs belong to the same domain, then checks
// if the target URL's path starts with the base URL's path prefix
func IsSamePathPrefix(baseURL, targetURL string) (bool, error) {
	// First check if they belong to the same domain
	sameDomain, err := IsSameDomain(baseURL, targetURL)
	if err != nil {
		return false, fmt.Errorf("failed to check domain similarity: %w", err)
	}
	if !sameDomain {
		return false, nil
	}

	// Parse both URLs to extract paths
	baseParsed, err := url.Parse(baseURL)
	if err != nil {
		return false, fmt.Errorf("failed to parse base URL: %w", err)
	}

	targetParsed, err := url.Parse(targetURL)
	if err != nil {
		return false, fmt.Errorf("failed to parse target URL: %w", err)
	}

	// Normalize paths - ensure they end with / for proper prefix matching
	basePath := baseParsed.Path
	targetPath := targetParsed.Path

	// Ensure paths end with / for directory-style prefix matching
	if basePath != "/" && !strings.HasSuffix(basePath, "/") {
		basePath += "/"
	}
	if targetPath != "/" && !strings.HasSuffix(targetPath, "/") {
		targetPath += "/"
	}

	// Check if target path starts with base path
	return strings.HasPrefix(targetPath, basePath), nil
}

// ShouldSkipURL checks if a URL should be skipped based on common patterns
func ShouldSkipURL(rawURL string) bool {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return true
	}

	// Skip common non-HTTP schemes
	lowerURL := strings.ToLower(rawURL)
	skipPrefixes := []string{
		"javascript:",
		"mailto:",
		"tel:",
		"ftp:",
		"file:",
		"data:",
		"#",
	}

	for _, prefix := range skipPrefixes {
		if strings.HasPrefix(lowerURL, prefix) {
			return true
		}
	}

	return false
}
