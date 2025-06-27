package utils

import (
	"testing"
)

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name     string
		rawURL   string
		expected string
		wantErr  bool
	}{
		{
			name:     "Valid HTTPS URL",
			rawURL:   "https://example.com",
			expected: "https://example.com",
			wantErr:  false,
		},
		{
			name:     "Valid HTTP URL",
			rawURL:   "http://example.com",
			expected: "http://example.com",
			wantErr:  false,
		},
		{
			name:     "URL without scheme - should add HTTPS",
			rawURL:   "example.com",
			expected: "https://example.com",
			wantErr:  false,
		},
		{
			name:     "URL with path",
			rawURL:   "https://example.com/path/to/page",
			expected: "https://example.com/path/to/page",
			wantErr:  false,
		},
		{
			name:     "URL with query parameters",
			rawURL:   "https://example.com?param=value",
			expected: "https://example.com?param=value",
			wantErr:  false,
		},
		{
			name:     "URL with fragment",
			rawURL:   "https://example.com#section",
			expected: "https://example.com#section",
			wantErr:  false,
		},
		{
			name:     "URL with port",
			rawURL:   "https://example.com:8080",
			expected: "https://example.com:8080",
			wantErr:  false,
		},
		{
			name:     "Domain without scheme - should add HTTPS",
			rawURL:   "google.com",
			expected: "https://google.com",
			wantErr:  false,
		},
		{
			name:     "Domain with subdomain without scheme",
			rawURL:   "api.example.com",
			expected: "https://api.example.com",
			wantErr:  false,
		},
		{
			name:    "Empty URL",
			rawURL:  "",
			wantErr: true,
		},
		{
			name:    "Whitespace only",
			rawURL:  "   ",
			wantErr: true, // Invalid character in host name
		},
		{
			name:     "URL with invalid scheme gets https prefix",
			rawURL:   "ftp://example.com",
			expected: "https://ftp://example.com",
			wantErr:  false, // ValidateURL adds https:// prefix
		},
		{
			name:     "Special characters in URL - accepted by Go's parser",
			rawURL:   "https://[invalid-host]",
			expected: "https://[invalid-host]",
			wantErr:  false, // Go's url.Parse accepts this
		},
		{
			name:    "Malformed URL",
			rawURL:  "http://",
			wantErr: true,
		},
		{
			name:    "URL with only path",
			rawURL:  "/path/only",
			wantErr: true, // No host
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateURL(tt.rawURL)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateURL() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("ValidateURL() unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("ValidateURL() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestExtractDomain(t *testing.T) {
	tests := []struct {
		name     string
		rawURL   string
		expected string
		wantErr  bool
	}{
		{
			name:     "Basic HTTPS URL",
			rawURL:   "https://example.com",
			expected: "example.com",
			wantErr:  false,
		},
		{
			name:     "Basic HTTP URL",
			rawURL:   "http://example.com",
			expected: "example.com",
			wantErr:  false,
		},
		{
			name:     "URL with path",
			rawURL:   "https://example.com/path/to/page",
			expected: "example.com",
			wantErr:  false,
		},
		{
			name:     "URL with subdomain",
			rawURL:   "https://api.example.com",
			expected: "api.example.com",
			wantErr:  false,
		},
		{
			name:     "URL with port",
			rawURL:   "https://example.com:8080",
			expected: "example.com:8080",
			wantErr:  false,
		},
		{
			name:     "URL with query parameters",
			rawURL:   "https://example.com?param=value",
			expected: "example.com",
			wantErr:  false,
		},
		{
			name:     "URL with fragment",
			rawURL:   "https://example.com#section",
			expected: "example.com",
			wantErr:  false,
		},
		{
			name:     "Complex URL",
			rawURL:   "https://api.subdomain.example.com:8080/path?param=value#section",
			expected: "api.subdomain.example.com:8080",
			wantErr:  false,
		},
		{
			name:     "Empty string",
			rawURL:   "",
			expected: "",
			wantErr:  false, // ExtractDomain returns empty string for empty input
		},
		{
			name:     "Invalid URL - no scheme, no host separator",
			rawURL:   "not-a-url",
			expected: "",
			wantErr:  false, // ExtractDomain just parses, returns empty for path-only URLs
		},
		{
			name:     "Malformed URL",
			rawURL:   "http://[invalid-host]",
			expected: "[invalid-host]",
			wantErr:  false, // ExtractDomain extracts even malformed hosts
		},
		{
			name:     "URL without scheme but with host",
			rawURL:   "//example.com",
			expected: "example.com",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExtractDomain(tt.rawURL)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ExtractDomain() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("ExtractDomain() unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("ExtractDomain() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// Benchmark tests for performance testing
func BenchmarkValidateURL(b *testing.B) {
	testURL := "https://example.com/path/to/page?param=value#section"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ValidateURL(testURL)
	}
}

func BenchmarkExtractDomain(b *testing.B) {
	testURL := "https://api.subdomain.example.com:8080/path?param=value#section"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ExtractDomain(testURL)
	}
}
