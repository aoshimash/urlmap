package client

import (
	"context"
	"log/slog"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	// DefaultUserAgent is the default User-Agent string
	DefaultUserAgent = "crawld/1.0.0 (+https://github.com/aoshimash/crawld)"
	// DefaultTimeout is the default HTTP timeout
	DefaultTimeout = 30 * time.Second
	// DefaultRetryCount is the default number of retry attempts
	DefaultRetryCount = 3
	// DefaultRetryWaitTime is the default wait time between retries
	DefaultRetryWaitTime = 1 * time.Second
	// DefaultRetryMaxWaitTime is the maximum wait time between retries
	DefaultRetryMaxWaitTime = 5 * time.Second
)

// Config holds the HTTP client configuration
type Config struct {
	UserAgent        string
	Timeout          time.Duration
	RetryCount       int
	RetryWaitTime    time.Duration
	RetryMaxWaitTime time.Duration
}

// DefaultConfig returns the default client configuration
func DefaultConfig() *Config {
	return &Config{
		UserAgent:        DefaultUserAgent,
		Timeout:          DefaultTimeout,
		RetryCount:       DefaultRetryCount,
		RetryWaitTime:    DefaultRetryWaitTime,
		RetryMaxWaitTime: DefaultRetryMaxWaitTime,
	}
}

// Client wraps the Resty client with additional functionality
type Client struct {
	client *resty.Client
	config *Config
}

// NewClient creates a new HTTP client with the given configuration
func NewClient(config *Config) *Client {
	if config == nil {
		config = DefaultConfig()
	}

	client := resty.New()

	// Basic configuration
	client.SetTimeout(config.Timeout)
	client.SetHeader("User-Agent", config.UserAgent)

	// Retry configuration
	client.SetRetryCount(config.RetryCount)
	client.SetRetryWaitTime(config.RetryWaitTime)
	client.SetRetryMaxWaitTime(config.RetryMaxWaitTime)

	// Retry conditions - only retry on server errors (5xx)
	client.AddRetryCondition(func(r *resty.Response, err error) bool {
		// Retry on network errors
		if err != nil {
			slog.Debug("Retrying due to network error", "error", err)
			return true
		}

		// Retry on 5xx server errors, but not on 4xx client errors
		if r.StatusCode() >= 500 {
			slog.Debug("Retrying due to server error", "status_code", r.StatusCode())
			return true
		}

		// Don't retry on 4xx client errors
		return false
	})

	// Request and response hooks for logging
	client.OnBeforeRequest(func(c *resty.Client, req *resty.Request) error {
		slog.Debug("HTTP request starting",
			"method", req.Method,
			"url", req.URL,
			"user_agent", req.Header.Get("User-Agent"),
		)
		return nil
	})

	client.OnAfterResponse(func(c *resty.Client, resp *resty.Response) error {
		duration := resp.Time()
		slog.Debug("HTTP request completed",
			"method", resp.Request.Method,
			"url", resp.Request.URL,
			"status_code", resp.StatusCode(),
			"duration", duration,
			"response_size", len(resp.Body()),
		)
		return nil
	})

	client.OnError(func(req *resty.Request, err error) {
		slog.Error("HTTP request failed",
			"method", req.Method,
			"url", req.URL,
			"error", err,
		)
	})

	return &Client{
		client: client,
		config: config,
	}
}

// NewDefaultClient creates a new HTTP client with default configuration
func NewDefaultClient() *Client {
	return NewClient(DefaultConfig())
}

// Get performs a GET request to the specified URL
func (c *Client) Get(ctx context.Context, url string) (*resty.Response, error) {
	return c.client.R().
		SetContext(ctx).
		Get(url)
}

// GetWithHeaders performs a GET request with custom headers
func (c *Client) GetWithHeaders(ctx context.Context, url string, headers map[string]string) (*resty.Response, error) {
	return c.client.R().
		SetContext(ctx).
		SetHeaders(headers).
		Get(url)
}

// Post performs a POST request with the given body
func (c *Client) Post(ctx context.Context, url string, body interface{}) (*resty.Response, error) {
	return c.client.R().
		SetContext(ctx).
		SetBody(body).
		Post(url)
}

// GetClient returns the underlying Resty client for advanced usage
func (c *Client) GetClient() *resty.Client {
	return c.client
}

// GetConfig returns the client configuration
func (c *Client) GetConfig() *Config {
	return c.config
}

// SetUserAgent updates the User-Agent string
func (c *Client) SetUserAgent(userAgent string) {
	c.client.SetHeader("User-Agent", userAgent)
	c.config.UserAgent = userAgent
}

// SetTimeout updates the request timeout
func (c *Client) SetTimeout(timeout time.Duration) {
	c.client.SetTimeout(timeout)
	c.config.Timeout = timeout
}

// IsSuccess checks if the HTTP response indicates success (2xx status code)
func IsSuccess(resp *resty.Response) bool {
	return resp.StatusCode() >= 200 && resp.StatusCode() < 300
}

// IsClientError checks if the HTTP response indicates a client error (4xx status code)
func IsClientError(resp *resty.Response) bool {
	return resp.StatusCode() >= 400 && resp.StatusCode() < 500
}

// IsServerError checks if the HTTP response indicates a server error (5xx status code)
func IsServerError(resp *resty.Response) bool {
	return resp.StatusCode() >= 500
}

// GetStatusMessage returns a human-readable status message for the response
func GetStatusMessage(resp *resty.Response) string {
	statusCode := resp.StatusCode()

	switch {
	case statusCode >= 200 && statusCode < 300:
		return "Success"
	case statusCode >= 300 && statusCode < 400:
		return "Redirection"
	case statusCode >= 400 && statusCode < 500:
		return "Client Error"
	case statusCode >= 500:
		return "Server Error"
	default:
		return "Unknown"
	}
}
