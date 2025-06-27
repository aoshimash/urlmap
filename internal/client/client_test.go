package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.UserAgent != DefaultUserAgent {
		t.Errorf("Expected UserAgent %q, got %q", DefaultUserAgent, config.UserAgent)
	}

	if config.Timeout != DefaultTimeout {
		t.Errorf("Expected Timeout %v, got %v", DefaultTimeout, config.Timeout)
	}

	if config.RetryCount != DefaultRetryCount {
		t.Errorf("Expected RetryCount %d, got %d", DefaultRetryCount, config.RetryCount)
	}

	if config.RetryWaitTime != DefaultRetryWaitTime {
		t.Errorf("Expected RetryWaitTime %v, got %v", DefaultRetryWaitTime, config.RetryWaitTime)
	}

	if config.RetryMaxWaitTime != DefaultRetryMaxWaitTime {
		t.Errorf("Expected RetryMaxWaitTime %v, got %v", DefaultRetryMaxWaitTime, config.RetryMaxWaitTime)
	}
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
	}{
		{
			name:   "With default config",
			config: DefaultConfig(),
		},
		{
			name: "With custom config",
			config: &Config{
				UserAgent:        "test-client/1.0",
				Timeout:          10 * time.Second,
				RetryCount:       2,
				RetryWaitTime:    500 * time.Millisecond,
				RetryMaxWaitTime: 2 * time.Second,
			},
		},
		{
			name:   "With nil config (should use defaults)",
			config: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.config)

			if client == nil {
				t.Fatal("Expected non-nil client")
			}

			if client.client == nil {
				t.Fatal("Expected non-nil Resty client")
			}

			if client.config == nil {
				t.Fatal("Expected non-nil config")
			}

			// When config is nil, should use defaults
			if tt.config == nil {
				if client.config.UserAgent != DefaultUserAgent {
					t.Errorf("Expected default UserAgent %q, got %q", DefaultUserAgent, client.config.UserAgent)
				}
			} else {
				if client.config.UserAgent != tt.config.UserAgent {
					t.Errorf("Expected UserAgent %q, got %q", tt.config.UserAgent, client.config.UserAgent)
				}
				if client.config.Timeout != tt.config.Timeout {
					t.Errorf("Expected Timeout %v, got %v", tt.config.Timeout, client.config.Timeout)
				}
			}
		})
	}
}

func TestNewDefaultClient(t *testing.T) {
	client := NewDefaultClient()

	if client == nil {
		t.Fatal("Expected non-nil client")
	}

	if client.config.UserAgent != DefaultUserAgent {
		t.Errorf("Expected default UserAgent %q, got %q", DefaultUserAgent, client.config.UserAgent)
	}

	if client.config.Timeout != DefaultTimeout {
		t.Errorf("Expected default Timeout %v, got %v", DefaultTimeout, client.config.Timeout)
	}
}

func TestClientGet(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check User-Agent header
		userAgent := r.Header.Get("User-Agent")
		if userAgent != DefaultUserAgent {
			t.Errorf("Expected User-Agent %q, got %q", DefaultUserAgent, userAgent)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, World!"))
	}))
	defer server.Close()

	client := NewDefaultClient()
	ctx := context.Background()

	resp, err := client.Get(ctx, server.URL)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if resp.StatusCode() != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode())
	}

	body := string(resp.Body())
	expected := "Hello, World!"
	if body != expected {
		t.Errorf("Expected body %q, got %q", expected, body)
	}
}

func TestClientGetWithHeaders(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check custom header
		customHeader := r.Header.Get("X-Custom-Header")
		if customHeader != "test-value" {
			t.Errorf("Expected X-Custom-Header %q, got %q", "test-value", customHeader)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Custom header received"))
	}))
	defer server.Close()

	client := NewDefaultClient()
	ctx := context.Background()
	headers := map[string]string{
		"X-Custom-Header": "test-value",
	}

	resp, err := client.GetWithHeaders(ctx, server.URL, headers)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if resp.StatusCode() != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode())
	}
}

func TestClientPost(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("Created"))
	}))
	defer server.Close()

	client := NewDefaultClient()
	ctx := context.Background()
	body := map[string]string{"key": "value"}

	resp, err := client.Post(ctx, server.URL, body)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if resp.StatusCode() != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, resp.StatusCode())
	}
}

func TestSetUserAgent(t *testing.T) {
	client := NewDefaultClient()
	newUserAgent := "test-client/2.0"

	client.SetUserAgent(newUserAgent)

	if client.config.UserAgent != newUserAgent {
		t.Errorf("Expected UserAgent %q, got %q", newUserAgent, client.config.UserAgent)
	}
}

func TestSetTimeout(t *testing.T) {
	client := NewDefaultClient()
	newTimeout := 15 * time.Second

	client.SetTimeout(newTimeout)

	if client.config.Timeout != newTimeout {
		t.Errorf("Expected Timeout %v, got %v", newTimeout, client.config.Timeout)
	}
}

func TestStatusCheckFunctions(t *testing.T) {
	tests := []struct {
		name        string
		statusCode  int
		isSuccess   bool
		isClientErr bool
		isServerErr bool
		statusMsg   string
	}{
		{"Success 200", 200, true, false, false, "Success"},
		{"Success 201", 201, true, false, false, "Success"},
		{"Success 299", 299, true, false, false, "Success"},
		{"Redirection 301", 301, false, false, false, "Redirection"},
		{"Redirection 302", 302, false, false, false, "Redirection"},
		{"Client Error 400", 400, false, true, false, "Client Error"},
		{"Client Error 404", 404, false, true, false, "Client Error"},
		{"Client Error 499", 499, false, true, false, "Client Error"},
		{"Server Error 500", 500, false, false, true, "Server Error"},
		{"Server Error 503", 503, false, false, true, "Server Error"},
		{"Redirection 399", 399, false, false, false, "Redirection"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server with specific status code
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			client := NewDefaultClient()
			ctx := context.Background()
			actualResp, err := client.Get(ctx, server.URL)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if IsSuccess(actualResp) != tt.isSuccess {
				t.Errorf("IsSuccess() = %v, want %v", IsSuccess(actualResp), tt.isSuccess)
			}

			if IsClientError(actualResp) != tt.isClientErr {
				t.Errorf("IsClientError() = %v, want %v", IsClientError(actualResp), tt.isClientErr)
			}

			if IsServerError(actualResp) != tt.isServerErr {
				t.Errorf("IsServerError() = %v, want %v", IsServerError(actualResp), tt.isServerErr)
			}

			if GetStatusMessage(actualResp) != tt.statusMsg {
				t.Errorf("GetStatusMessage() = %q, want %q", GetStatusMessage(actualResp), tt.statusMsg)
			}
		})
	}
}

func TestClientTimeout(t *testing.T) {
	// Create test server with delay
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second) // Delay longer than client timeout
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create client with short timeout
	config := &Config{
		UserAgent:        DefaultUserAgent,
		Timeout:          100 * time.Millisecond, // Very short timeout
		RetryCount:       0,                      // No retries for this test
		RetryWaitTime:    DefaultRetryWaitTime,
		RetryMaxWaitTime: DefaultRetryMaxWaitTime,
	}
	client := NewClient(config)
	ctx := context.Background()

	_, err := client.Get(ctx, server.URL)
	if err == nil {
		t.Error("Expected timeout error, but got none")
	}
}

func TestGetClient(t *testing.T) {
	client := NewDefaultClient()
	restyClient := client.GetClient()

	if restyClient == nil {
		t.Error("Expected non-nil Resty client")
	}

	if restyClient != client.client {
		t.Error("GetClient() should return the same Resty client instance")
	}
}

func TestGetConfig(t *testing.T) {
	config := &Config{
		UserAgent:        "test-agent",
		Timeout:          5 * time.Second,
		RetryCount:       2,
		RetryWaitTime:    500 * time.Millisecond,
		RetryMaxWaitTime: 3 * time.Second,
	}

	client := NewClient(config)
	returnedConfig := client.GetConfig()

	if returnedConfig == nil {
		t.Error("Expected non-nil config")
	}

	if returnedConfig.UserAgent != config.UserAgent {
		t.Errorf("Expected UserAgent %q, got %q", config.UserAgent, returnedConfig.UserAgent)
	}

	if returnedConfig.Timeout != config.Timeout {
		t.Errorf("Expected Timeout %v, got %v", config.Timeout, returnedConfig.Timeout)
	}
}
