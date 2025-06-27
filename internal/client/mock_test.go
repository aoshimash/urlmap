package client

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// MockServer provides a test HTTP server for client testing
type MockServer struct {
	*httptest.Server
	requestCount int
	lastRequest  *http.Request
}

// NewMockServer creates a new mock HTTP server
func NewMockServer(handler http.HandlerFunc) *MockServer {
	server := httptest.NewServer(handler)
	return &MockServer{
		Server: server,
	}
}

// SuccessHandler returns a handler that always returns 200 OK
func SuccessHandler(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("Success"))
		if err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}
}

// DelayHandler returns a handler that delays response
func DelayHandler(delay time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(delay)
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("Delayed response"))
		if err != nil {
			// Silent fail in test
		}
	}
}

// ErrorHandler returns a handler that returns server errors
func ErrorHandler(statusCode int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		_, err := w.Write([]byte(fmt.Sprintf("Error %d", statusCode)))
		if err != nil {
			// Silent fail in test
		}
	}
}

// TestClientWithMockServer tests the client with a mock HTTP server
func TestClientWithMockServer(t *testing.T) {
	tests := []struct {
		name           string
		handler        http.HandlerFunc
		expectedStatus int
		expectedBody   string
		wantErr        bool
	}{
		{
			name:           "Successful request",
			handler:        SuccessHandler(t),
			expectedStatus: 200,
			expectedBody:   "Success",
			wantErr:        false,
		},
		{
			name:           "Server error",
			handler:        ErrorHandler(http.StatusInternalServerError),
			expectedStatus: 500,
			expectedBody:   "Error 500",
			wantErr:        false, // HTTP errors are not Go errors
		},
		{
			name:           "Not found error",
			handler:        ErrorHandler(http.StatusNotFound),
			expectedStatus: 404,
			expectedBody:   "Error 404",
			wantErr:        false,
		},
		{
			name:           "Bad gateway error",
			handler:        ErrorHandler(http.StatusBadGateway),
			expectedStatus: 502,
			expectedBody:   "Error 502",
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockServer := NewMockServer(tt.handler)
			defer mockServer.Close()

			client := NewDefaultClient()
			ctx := context.Background()

			resp, err := client.Get(ctx, mockServer.URL)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if resp.StatusCode() != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode())
			}

			body := string(resp.Body())
			if body != tt.expectedBody {
				t.Errorf("Expected body %q, got %q", tt.expectedBody, body)
			}
		})
	}
}

// TestClientRetryBehavior tests retry behavior with server errors
func TestClientRetryBehavior(t *testing.T) {
	tests := []struct {
		name        string
		statusCode  int
		expectRetry bool
		retryCount  int
		expectFinal int
	}{
		{
			name:        "Retry on 500 error",
			statusCode:  500,
			expectRetry: true,
			retryCount:  2,
			expectFinal: 500, // Still fails after retries
		},
		{
			name:        "No retry on 404 error",
			statusCode:  404,
			expectRetry: false,
			retryCount:  0,
			expectFinal: 404,
		},
		{
			name:        "Retry on 503 error",
			statusCode:  503,
			expectRetry: true,
			retryCount:  2,
			expectFinal: 503,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestCount := 0
			handler := func(w http.ResponseWriter, r *http.Request) {
				requestCount++
				w.WriteHeader(tt.statusCode)
				_, err := w.Write([]byte(fmt.Sprintf("Error %d", tt.statusCode)))
				if err != nil {
					// Silent fail in test
				}
			}

			mockServer := NewMockServer(handler)
			defer mockServer.Close()

			config := &Config{
				UserAgent:        "test-client",
				Timeout:          5 * time.Second,
				RetryCount:       2,
				RetryWaitTime:    10 * time.Millisecond, // Short wait for tests
				RetryMaxWaitTime: 50 * time.Millisecond,
			}

			client := NewClient(config)
			ctx := context.Background()

			resp, err := client.Get(ctx, mockServer.URL)

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if resp.StatusCode() != tt.expectFinal {
				t.Errorf("Expected final status %d, got %d", tt.expectFinal, resp.StatusCode())
			}

			expectedRequests := 1
			if tt.expectRetry {
				expectedRequests = tt.retryCount + 1
			}

			if requestCount != expectedRequests {
				t.Errorf("Expected %d requests, got %d", expectedRequests, requestCount)
			}
		})
	}
}

// TestClientTimeoutWithMock tests timeout behavior with mock server
func TestClientTimeoutWithMock(t *testing.T) {
	delayHandler := DelayHandler(2 * time.Second)
	mockServer := NewMockServer(delayHandler)
	defer mockServer.Close()

	config := &Config{
		UserAgent:        "test-client",
		Timeout:          100 * time.Millisecond, // Short timeout
		RetryCount:       0,                      // No retries for timeout test
		RetryWaitTime:    10 * time.Millisecond,
		RetryMaxWaitTime: 50 * time.Millisecond,
	}

	client := NewClient(config)
	ctx := context.Background()

	_, err := client.Get(ctx, mockServer.URL)

	if err == nil {
		t.Error("Expected timeout error but got none")
	}

	if !strings.Contains(err.Error(), "timeout") && !strings.Contains(err.Error(), "deadline") {
		t.Errorf("Expected timeout error, got: %v", err)
	}
}

// TestClientHeaders tests custom headers
func TestClientHeaders(t *testing.T) {
	var receivedHeaders http.Header
	handler := func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header.Clone()
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("OK"))
		if err != nil {
			// Silent fail in test
		}
	}

	mockServer := NewMockServer(handler)
	defer mockServer.Close()

	client := NewDefaultClient()
	ctx := context.Background()

	customHeaders := map[string]string{
		"X-Test-Header": "test-value",
		"Authorization": "Bearer token123",
	}

	_, err := client.GetWithHeaders(ctx, mockServer.URL, customHeaders)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	// Check if custom headers were sent
	for key, expectedValue := range customHeaders {
		actualValue := receivedHeaders.Get(key)
		if actualValue != expectedValue {
			t.Errorf("Expected header %s=%s, got %s", key, expectedValue, actualValue)
		}
	}

	// Check if User-Agent is set
	userAgent := receivedHeaders.Get("User-Agent")
	if userAgent == "" {
		t.Error("User-Agent header not set")
	}
}

// TestClientUserAgent tests User-Agent header
func TestClientUserAgent(t *testing.T) {
	var receivedUserAgent string
	handler := func(w http.ResponseWriter, r *http.Request) {
		receivedUserAgent = r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("OK"))
		if err != nil {
			// Silent fail in test
		}
	}

	mockServer := NewMockServer(handler)
	defer mockServer.Close()

	tests := []struct {
		name      string
		userAgent string
		expected  string
	}{
		{
			name:      "Default User-Agent",
			userAgent: DefaultUserAgent,
			expected:  DefaultUserAgent,
		},
		{
			name:      "Custom User-Agent",
			userAgent: "custom-crawler/2.0",
			expected:  "custom-crawler/2.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				UserAgent:        tt.userAgent,
				Timeout:          5 * time.Second,
				RetryCount:       0,
				RetryWaitTime:    10 * time.Millisecond,
				RetryMaxWaitTime: 50 * time.Millisecond,
			}

			client := NewClient(config)
			ctx := context.Background()

			_, err := client.Get(ctx, mockServer.URL)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if receivedUserAgent != tt.expected {
				t.Errorf("Expected User-Agent %s, got %s", tt.expected, receivedUserAgent)
			}
		})
	}
}

// TestClientPostWithMock tests POST requests with mock server
func TestClientPostWithMock(t *testing.T) {
	var receivedMethod string

	handler := func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		w.WriteHeader(http.StatusCreated)
		_, err := w.Write([]byte("Created"))
		if err != nil {
			// Silent fail in test
		}
	}

	mockServer := NewMockServer(handler)
	defer mockServer.Close()

	client := NewDefaultClient()
	ctx := context.Background()

	testData := map[string]interface{}{
		"key": "value",
		"num": 42,
	}

	resp, err := client.Post(ctx, mockServer.URL, testData)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	if receivedMethod != "POST" {
		t.Errorf("Expected POST method, got %s", receivedMethod)
	}

	if resp.StatusCode() != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode())
	}

	responseBody := string(resp.Body())
	if responseBody != "Created" {
		t.Errorf("Expected response body 'Created', got %s", responseBody)
	}
}

// BenchmarkClientGet benchmarks GET requests
func BenchmarkClientGet(b *testing.B) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("Benchmark response"))
		if err != nil {
			// Silent fail in benchmark
		}
	}

	mockServer := NewMockServer(handler)
	defer mockServer.Close()

	client := NewDefaultClient()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.Get(ctx, mockServer.URL)
		if err != nil {
			b.Errorf("Unexpected error: %v", err)
		}
	}
}
