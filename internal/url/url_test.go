package url

import (
	"testing"
)

func TestIsValidURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		// Valid URLs
		{"Valid HTTP URL", "http://example.com", true},
		{"Valid HTTPS URL", "https://example.com", true},
		{"Valid URL with path", "https://example.com/path", true},
		{"Valid URL with query", "https://example.com?query=value", true},
		{"Valid URL with fragment", "https://example.com#fragment", true},
		{"Valid URL with port", "https://example.com:8080", true},
		{"Valid URL with subdomain", "https://sub.example.com", true},

		// Invalid URLs
		{"Empty string", "", false},
		{"Whitespace only", "   ", false},
		{"No scheme", "example.com", false},
		{"FTP scheme", "ftp://example.com", false},
		{"File scheme", "file:///path", false},
		{"JavaScript scheme", "javascript:alert('test')", false},
		{"Mailto scheme", "mailto:test@example.com", false},
		{"Invalid URL", "http://", false},
		{"No host", "http://", false},
		{"Malformed URL", "http:///invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidURL(tt.input)
			if result != tt.expected {
				t.Errorf("IsValidURL(%q) = %v; want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestExtractDomain(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		shouldError bool
	}{
		// Valid cases
		{"Basic HTTP URL", "http://example.com", "example.com", false},
		{"Basic HTTPS URL", "https://example.com", "example.com", false},
		{"URL with path", "https://example.com/path", "example.com", false},
		{"URL with port", "https://example.com:8080", "example.com", false},
		{"URL with subdomain", "https://sub.example.com", "sub.example.com", false},
		{"URL with query", "https://example.com?query=value", "example.com", false},
		{"URL with fragment", "https://example.com#fragment", "example.com", false},

		// Error cases
		{"Empty string", "", "", true},
		{"Whitespace only", "   ", "", true},
		{"Invalid URL", "http://", "", true},
		{"No host", "http:///path", "", true},
		{"Malformed URL", "://invalid", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExtractDomain(tt.input)

			if tt.shouldError {
				if err == nil {
					t.Errorf("ExtractDomain(%q) expected error but got none", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("ExtractDomain(%q) unexpected error: %v", tt.input, err)
				}
				if result != tt.expected {
					t.Errorf("ExtractDomain(%q) = %q; want %q", tt.input, result, tt.expected)
				}
			}
		})
	}
}

func TestResolveURL(t *testing.T) {
	tests := []struct {
		name        string
		baseURL     string
		relativeURL string
		expected    string
		shouldError bool
	}{
		// Valid cases
		{"Absolute path", "https://example.com", "/path", "https://example.com/path", false},
		{"Relative path", "https://example.com/dir/", "file.html", "https://example.com/dir/file.html", false},
		{"Parent directory", "https://example.com/dir/subdir/", "../file.html", "https://example.com/dir/file.html", false},
		{"Current directory", "https://example.com/dir/", "./file.html", "https://example.com/dir/file.html", false},
		{"Query parameters", "https://example.com", "?query=value", "https://example.com?query=value", false},
		{"Fragment", "https://example.com", "#fragment", "https://example.com#fragment", false},
		{"Absolute URL", "https://example.com", "https://other.com/path", "https://other.com/path", false},

		// Error cases
		{"Empty base URL", "", "/path", "", true},
		{"Empty relative URL", "https://example.com", "", "", true},
		{"Whitespace base URL", "   ", "/path", "", true},
		{"Whitespace relative URL", "https://example.com", "   ", "", true},
		{"Invalid base URL", "://invalid", "/path", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ResolveURL(tt.baseURL, tt.relativeURL)

			if tt.shouldError {
				if err == nil {
					t.Errorf("ResolveURL(%q, %q) expected error but got none", tt.baseURL, tt.relativeURL)
				}
			} else {
				if err != nil {
					t.Errorf("ResolveURL(%q, %q) unexpected error: %v", tt.baseURL, tt.relativeURL, err)
				}
				if result != tt.expected {
					t.Errorf("ResolveURL(%q, %q) = %q; want %q", tt.baseURL, tt.relativeURL, result, tt.expected)
				}
			}
		})
	}
}

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		shouldError bool
	}{
		// Valid cases
		{"Basic URL", "https://example.com", "https://example.com/", false},
		{"URL with trailing slash", "https://example.com/", "https://example.com/", false},
		{"URL with path trailing slash", "https://example.com/path/", "https://example.com/path", false},
		{"URL with fragment", "https://example.com/path#fragment", "https://example.com/path", false},
		{"URL with query and fragment", "https://example.com/path?query=value#fragment", "https://example.com/path?query=value", false},
		{"Root path", "https://example.com/", "https://example.com/", false},
		{"Empty path", "https://example.com", "https://example.com/", false},

		// Error cases
		{"Empty string", "", "", true},
		{"Whitespace only", "   ", "", true},
		{"Invalid URL", "://invalid", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NormalizeURL(tt.input)

			if tt.shouldError {
				if err == nil {
					t.Errorf("NormalizeURL(%q) expected error but got none", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("NormalizeURL(%q) unexpected error: %v", tt.input, err)
				}
				if result != tt.expected {
					t.Errorf("NormalizeURL(%q) = %q; want %q", tt.input, result, tt.expected)
				}
			}
		})
	}
}

func TestIsSameDomain(t *testing.T) {
	tests := []struct {
		name        string
		url1        string
		url2        string
		expected    bool
		shouldError bool
	}{
		// Valid cases
		{"Same domain", "https://example.com", "https://example.com/path", true, false},
		{"Same domain different schemes", "http://example.com", "https://example.com", true, false},
		{"Same domain with ports", "https://example.com:8080", "https://example.com:9090", true, false},
		{"Different domains", "https://example.com", "https://other.com", false, false},
		{"Subdomain vs main domain", "https://sub.example.com", "https://example.com", false, false},
		{"Case insensitive", "https://Example.COM", "https://example.com", true, false},

		// Error cases
		{"Invalid first URL", "invalid", "https://example.com", false, true},
		{"Invalid second URL", "https://example.com", "invalid", false, true},
		{"Both invalid", "invalid1", "invalid2", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := IsSameDomain(tt.url1, tt.url2)

			if tt.shouldError {
				if err == nil {
					t.Errorf("IsSameDomain(%q, %q) expected error but got none", tt.url1, tt.url2)
				}
			} else {
				if err != nil {
					t.Errorf("IsSameDomain(%q, %q) unexpected error: %v", tt.url1, tt.url2, err)
				}
				if result != tt.expected {
					t.Errorf("IsSameDomain(%q, %q) = %v; want %v", tt.url1, tt.url2, result, tt.expected)
				}
			}
		})
	}
}

func TestShouldSkipURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		// Should skip
		{"Empty string", "", true},
		{"Whitespace only", "   ", true},
		{"JavaScript URL", "javascript:alert('test')", true},
		{"Mailto URL", "mailto:test@example.com", true},
		{"Tel URL", "tel:+1234567890", true},
		{"FTP URL", "ftp://example.com", true},
		{"File URL", "file:///path", true},
		{"Data URL", "data:text/plain;base64,SGVsbG8=", true},
		{"Fragment only", "#fragment", true},
		{"JavaScript case insensitive", "JAVASCRIPT:alert('test')", true},

		// Should not skip
		{"Valid HTTP URL", "http://example.com", false},
		{"Valid HTTPS URL", "https://example.com", false},
		{"Relative path", "/path", false},
		{"Relative URL", "./path", false},
		{"Query only", "?query=value", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShouldSkipURL(tt.input)
			if result != tt.expected {
				t.Errorf("ShouldSkipURL(%q) = %v; want %v", tt.input, result, tt.expected)
			}
		})
	}
}
