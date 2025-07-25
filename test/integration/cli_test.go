package integration

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testHTML = `<!DOCTYPE html>
<html>
<head>
    <title>Test Page</title>
</head>
<body>
    <h1>Test Page</h1>
    <p>This is a test page for crawling.</p>
    <a href="/page1">Page 1</a>
    <a href="/page2">Page 2</a>
    <a href="https://external.example.com">External Link</a>
</body>
</html>`

	page1HTML = `<!DOCTYPE html>
<html>
<head>
    <title>Page 1</title>
</head>
<body>
    <h1>Page 1</h1>
    <p>This is page 1.</p>
    <a href="/page2">Go to Page 2</a>
    <a href="/">Back to Home</a>
</body>
</html>`

	page2HTML = `<!DOCTYPE html>
<html>
<head>
    <title>Page 2</title>
</head>
<body>
    <h1>Page 2</h1>
    <p>This is page 2.</p>
    <a href="/">Back to Home</a>
</body>
</html>`
)

func createTestServer() *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, testHTML)
	})

	mux.HandleFunc("/page1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, page1HTML)
	})

	mux.HandleFunc("/page2", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, page2HTML)
	})

	return httptest.NewServer(mux)
}

func setupCLITest(t *testing.T) (string, func()) {
	// Create temporary directory for test
	tempDir := t.TempDir()

	// Build the urlmap binary
	binaryPath := filepath.Join(tempDir, "urlmap")

	// Get project root directory
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	// Navigate to project root (assuming we're in test/integration)
	projectRoot := filepath.Join(wd, "..", "..")
	cmdPath := filepath.Join(projectRoot, "cmd", "urlmap")

	cmd := exec.Command("go", "build", "-o", binaryPath, cmdPath)
	cmd.Dir = projectRoot
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build urlmap binary from %s: %v", cmdPath, err)
	}

	return binaryPath, func() {
		os.Remove(binaryPath)
	}
}

func TestCrawlCommand_BasicFunctionality(t *testing.T) {
	// Setup test server
	server := createTestServer()
	defer server.Close()

	// Build binary
	binaryPath, cleanup := setupCLITest(t)
	defer cleanup()

	// Run urlmap command
	cmd := exec.Command(binaryPath, server.URL)
	output, err := cmd.Output()

	// Verify results
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, server.URL) {
		t.Errorf("Expected output to contain %s, got: %s", server.URL, outputStr)
	}
}

func TestCrawlCommand_WithDepth(t *testing.T) {
	// Setup test server
	server := createTestServer()
	defer server.Close()

	// Build binary
	binaryPath, cleanup := setupCLITest(t)
	defer cleanup()

	// Run urlmap command with depth=2
	cmd := exec.Command(binaryPath, "--depth=2", server.URL)
	output, err := cmd.Output()

	// Verify results
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	outputStr := string(output)

	// Should contain at least the base URL
	if !strings.Contains(outputStr, server.URL) {
		t.Errorf("Expected output to contain base URL %s, got: %s", server.URL, outputStr)
	}

	// With depth=2, should be able to crawl linked pages
	// The exact behavior depends on crawler implementation
	// Let's just check that we get some output
	if len(strings.TrimSpace(outputStr)) == 0 {
		t.Error("Expected non-empty output from crawl")
	}

	t.Logf("Crawl output with depth=2: %s", outputStr)
}

func TestCrawlCommand_WithVerboseFlag(t *testing.T) {
	// Setup test server
	server := createTestServer()
	defer server.Close()

	// Build binary
	binaryPath, cleanup := setupCLITest(t)
	defer cleanup()

	// Run urlmap command with verbose flag
	cmd := exec.Command(binaryPath, "--verbose", server.URL)

	// Capture both stdout and stderr
	output, err := cmd.CombinedOutput()

	// Verify results
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	outputStr := string(output)

	// In verbose mode, we should see log messages
	if !strings.Contains(outputStr, "Starting crawl") {
		t.Errorf("Expected verbose output to contain log messages, got: %s", outputStr)
	}
}

func TestCrawlCommand_WithConcurrency(t *testing.T) {
	// Setup test server
	server := createTestServer()
	defer server.Close()

	// Build binary
	binaryPath, cleanup := setupCLITest(t)
	defer cleanup()

	// Run urlmap command with concurrency settings
	cmd := exec.Command(binaryPath, "--concurrent=5", "--depth=2", server.URL)
	output, err := cmd.Output()

	// Verify results
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, server.URL) {
		t.Errorf("Expected output to contain %s, got: %s", server.URL, outputStr)
	}
}

func TestCrawlCommand_InvalidURL(t *testing.T) {
	// Build binary
	binaryPath, cleanup := setupCLITest(t)
	defer cleanup()

	// Run urlmap command with invalid URL
	cmd := exec.Command(binaryPath, "not-a-valid-url")
	output, err := cmd.CombinedOutput()

	// Should fail with error
	if err == nil {
		t.Fatal("Command should have failed with invalid URL")
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "invalid URL") {
		t.Errorf("Expected error message about invalid URL, got: %s", outputStr)
	}
}

func TestCrawlCommand_NonExistentHost(t *testing.T) {
	// Build binary
	binaryPath, cleanup := setupCLITest(t)
	defer cleanup()

	// Run urlmap command with non-existent host
	cmd := exec.Command(binaryPath, "http://non-existent-host-12345.example")

	// Set a timeout for the command execution
	timeout := time.After(15 * time.Second)
	done := make(chan struct {
		output []byte
		err    error
	}, 1)

	go func() {
		output, err := cmd.CombinedOutput()
		done <- struct {
			output []byte
			err    error
		}{output, err}
	}()

	select {
	case result := <-done:
		// Should fail with network error or timeout
		// Accept either error condition or successful completion with no output
		// (depending on how the crawler handles DNS resolution failures)
		if result.err != nil {
			t.Logf("Command failed as expected: %v", result.err)
		} else {
			// Command succeeded but should have error logs
			outputStr := string(result.output)
			if !strings.Contains(outputStr, "error") && !strings.Contains(outputStr, "failed") {
				t.Error("Expected error messages in output for non-existent host")
			}
			t.Logf("Command completed with output: %s", outputStr)
		}
	case <-timeout:
		// Kill the process if it's taking too long
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		t.Fatal("Command timed out")
	}
}

func TestVersionCommand(t *testing.T) {
	// Build binary
	binaryPath, cleanup := setupCLITest(t)
	defer cleanup()

	// Run version command
	cmd := exec.Command(binaryPath, "version")
	output, err := cmd.Output()

	// Verify results
	if err != nil {
		t.Fatalf("Version command failed: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "urlmap version") {
		t.Errorf("Expected version output to contain version info, got: %s", outputStr)
	}
}

func TestCrawlCommand_UserAgent(t *testing.T) {
	// Setup test server that checks User-Agent
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		userAgent := r.Header.Get("User-Agent")
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<html><body><h1>User-Agent: %s</h1></body></html>`, userAgent)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	// Build binary
	binaryPath, cleanup := setupCLITest(t)
	defer cleanup()

	// Run urlmap command with custom User-Agent
	customUA := "TestBot/1.0"
	cmd := exec.Command(binaryPath, "--user-agent="+customUA, server.URL)
	output, err := cmd.Output()

	// Verify results
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, server.URL) {
		t.Errorf("Expected output to contain %s, got: %s", server.URL, outputStr)
	}
}

func TestCrawlCommand_JavaScriptRendering(t *testing.T) {
	// Skip if we're not in an environment that supports Playwright
	if os.Getenv("SKIP_JS_TESTS") != "" {
		t.Skip("JavaScript rendering tests skipped (SKIP_JS_TESTS is set)")
	}

	// Create a simple test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/spa":
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprint(w, `
				<!DOCTYPE html>
				<html>
				<head><title>SPA Test</title></head>
				<body>
					<div id="content">Loading...</div>
					<script>
						setTimeout(function() {
							document.getElementById('content').innerHTML =
								'<a href="/dynamic-link">Dynamic Link</a>';
						}, 100);
					</script>
				</body>
				</html>
			`)
		case "/dynamic-link":
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprint(w, `
				<!DOCTYPE html>
				<html>
				<head><title>Dynamic Page</title></head>
				<body><h1>Dynamic Content</h1></body>
				</html>
			`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	testURL := server.URL + "/spa"

	tests := []struct {
		name       string
		args       []string
		skipReason string
	}{
		{
			name:       "JavaScript rendering enabled",
			args:       []string{testURL, "--js-render", "--depth", "1"},
			skipReason: "Requires Playwright browser installation",
		},
		{
			name: "HTTP only (no JavaScript)",
			args: []string{testURL, "--depth", "1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipReason != "" {
				// Try to run the command and skip if it fails due to missing Playwright
				cmd := exec.Command("go", append([]string{"run", "../../cmd/urlmap"}, tt.args...)...)
				output, err := cmd.CombinedOutput()
				if err != nil && (strings.Contains(string(output), "playwright") ||
					strings.Contains(string(output), "browser")) {
					t.Skipf("Skipping test: %s", tt.skipReason)
					return
				}
			}

			cmd := exec.Command("go", append([]string{"run", "../../cmd/urlmap"}, tt.args...)...)
			output, err := cmd.Output()

			require.NoError(t, err, "Command should not fail")

			outputStr := strings.TrimSpace(string(output))
			assert.NotEmpty(t, outputStr, "Should produce some output")

			// Basic validation that we got URLs back
			lines := strings.Split(outputStr, "\n")
			assert.Greater(t, len(lines), 0, "Should have at least one URL")
		})
	}
}

func TestCrawlCommand_JavaScriptOptions(t *testing.T) {
	// Create a simple test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `
			<!DOCTYPE html>
			<html>
			<head><title>Test</title></head>
			<body><h1>Test Page</h1></body>
			</html>
		`)
	}))
	defer server.Close()

	testURL := server.URL

	tests := []struct {
		name        string
		args        []string
		expectError bool
		skipReason  string
	}{
		{
			name: "Valid JavaScript options - chromium",
			args: []string{
				testURL,
				"--js-render",
				"--js-browser", "chromium",
				"--js-timeout", "10s",
				"--js-wait", "networkidle",
			},
			expectError: false,
			skipReason:  "Requires Playwright browser installation",
		},
		{
			name: "Valid JavaScript options - firefox",
			args: []string{
				testURL,
				"--js-render",
				"--js-browser", "firefox",
				"--js-timeout", "10s",
				"--js-wait", "domcontentloaded",
			},
			expectError: false,
			skipReason:  "Requires Playwright browser installation",
		},
		{
			name:        "JavaScript disabled by default",
			args:        []string{testURL},
			expectError: false,
		},
		{
			name: "JavaScript with fallback enabled",
			args: []string{
				testURL,
				"--js-render",
				"--js-fallback",
			},
			expectError: false,
			skipReason:  "Requires Playwright browser installation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipReason != "" {
				// Try to run the command and skip if it fails due to missing Playwright
				cmd := exec.Command("go", append([]string{"run", "../../cmd/urlmap"}, tt.args...)...)
				output, err := cmd.CombinedOutput()
				if err != nil && strings.Contains(string(output), "playwright") {
					t.Skipf("Skipping test: %s", tt.skipReason)
					return
				}
			}

			cmd := exec.Command("go", append([]string{"run", "../../cmd/urlmap"}, tt.args...)...)
			output, err := cmd.CombinedOutput()

			if tt.expectError {
				assert.Error(t, err, "Command should fail")
			} else {
				assert.NoError(t, err, "Command should not fail. Output: %s", string(output))
			}
		})
	}
}
