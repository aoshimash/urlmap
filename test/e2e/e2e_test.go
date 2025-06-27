package e2e

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const (
	// Complex HTML structure for comprehensive testing
	indexHTML = `<!DOCTYPE html>
<html>
<head>
    <title>E2E Test Home</title>
    <meta name="description" content="E2E testing home page">
</head>
<body>
    <nav>
        <a href="/about">About</a>
        <a href="/products">Products</a>
        <a href="/contact">Contact</a>
    </nav>
    <main>
        <h1>Welcome to E2E Test Site</h1>
        <p>This is a comprehensive test site for crawling.</p>
        <div class="links">
            <a href="/blog">Blog</a>
            <a href="/blog/post1">Blog Post 1</a>
            <a href="/services">Services</a>
        </div>
    </main>
    <footer>
        <a href="/privacy">Privacy Policy</a>
        <a href="/terms">Terms of Service</a>
    </footer>
</body>
</html>`

	aboutHTML = `<!DOCTYPE html>
<html>
<head>
    <title>About - E2E Test</title>
</head>
<body>
    <h1>About Us</h1>
    <p>Learn more about our company.</p>
    <a href="/">Home</a>
    <a href="/team">Our Team</a>
</body>
</html>`

	productsHTML = `<!DOCTYPE html>
<html>
<head>
    <title>Products - E2E Test</title>
</head>
<body>
    <h1>Our Products</h1>
    <div class="product-list">
        <a href="/products/widget1">Widget 1</a>
        <a href="/products/widget2">Widget 2</a>
        <a href="/products/gadget1">Gadget 1</a>
    </div>
    <a href="/">Home</a>
</body>
</html>`

	blogHTML = `<!DOCTYPE html>
<html>
<head>
    <title>Blog - E2E Test</title>
</head>
<body>
    <h1>Blog</h1>
    <div class="posts">
        <a href="/blog/post1">First Post</a>
        <a href="/blog/post2">Second Post</a>
        <a href="/blog/post3">Third Post</a>
    </div>
    <a href="/">Home</a>
</body>
</html>`

	errorHTML = `<!DOCTYPE html>
<html>
<head>
    <title>Error 404</title>
</head>
<body>
    <h1>Page Not Found</h1>
    <p>The requested page could not be found.</p>
    <a href="/">Go Home</a>
</body>
</html>`
)

// Statistics tracking for server behavior
type ServerStats struct {
	RequestCount int64
	PathCounts   map[string]int64
	mutex        sync.RWMutex
}

func NewServerStats() *ServerStats {
	return &ServerStats{
		PathCounts: make(map[string]int64),
	}
}

func (s *ServerStats) IncrementRequest(path string) {
	atomic.AddInt64(&s.RequestCount, 1)
	s.mutex.Lock()
	s.PathCounts[path]++
	s.mutex.Unlock()
}

func (s *ServerStats) GetStats() (int64, map[string]int64) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	pathCounts := make(map[string]int64)
	for k, v := range s.PathCounts {
		pathCounts[k] = v
	}

	return atomic.LoadInt64(&s.RequestCount), pathCounts
}

func createComplexTestServer() (*httptest.Server, *ServerStats) {
	stats := NewServerStats()
	mux := http.NewServeMux()

	// Middleware to track requests
	trackingMiddleware := func(handler http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			stats.IncrementRequest(r.URL.Path)
			handler(w, r)
		}
	}

	// Main routes
	mux.HandleFunc("/", trackingMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, indexHTML)
	}))

	mux.HandleFunc("/about", trackingMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, aboutHTML)
	}))

	mux.HandleFunc("/products", trackingMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, productsHTML)
	}))

	mux.HandleFunc("/blog", trackingMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, blogHTML)
	}))

	// Dynamic content routes
	mux.HandleFunc("/products/", trackingMiddleware(func(w http.ResponseWriter, r *http.Request) {
		productName := strings.TrimPrefix(r.URL.Path, "/products/")
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<html><body><h1>Product: %s</h1><a href="/products">Back to Products</a></body></html>`, productName)
	}))

	mux.HandleFunc("/blog/", trackingMiddleware(func(w http.ResponseWriter, r *http.Request) {
		postName := strings.TrimPrefix(r.URL.Path, "/blog/")
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<html><body><h1>Blog Post: %s</h1><a href="/blog">Back to Blog</a></body></html>`, postName)
	}))

	// Slow response endpoint for testing performance
	mux.HandleFunc("/slow", trackingMiddleware(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<html><body><h1>Slow Page</h1><p>This page loads slowly.</p></body></html>`)
	}))

	// Error endpoints
	mux.HandleFunc("/error", trackingMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal Server Error")
	}))

	mux.HandleFunc("/not-found", trackingMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, errorHTML)
	}))

	return httptest.NewServer(mux), stats
}

func setupE2ETest(t *testing.T) (string, func()) {
	// Create temporary directory for test
	tempDir := t.TempDir()

	// Build the urlmap binary
	binaryPath := filepath.Join(tempDir, "urlmap")

	// Get project root directory
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	// Navigate to project root (assuming we're in test/e2e)
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

func TestE2E_CompleteWorkflow(t *testing.T) {
	// Setup complex test server
	server, stats := createComplexTestServer()
	defer server.Close()

	// Build binary
	binaryPath, cleanup := setupE2ETest(t)
	defer cleanup()

	// Run complete crawl with various flags
	cmd := exec.Command(binaryPath,
		"--depth=3",
		"--concurrent=5",
		"--verbose",
		"--progress=false", // Disable progress for cleaner test output
		server.URL)

	output, err := cmd.CombinedOutput()

	// Verify results
	if err != nil {
		t.Fatalf("E2E test failed: %v\nOutput: %s", err, string(output))
	}

	outputStr := string(output)

	// Verify that at least the base URL was crawled
	if !strings.Contains(outputStr, server.URL) {
		t.Errorf("Expected output to contain base URL %s, got: %s", server.URL, outputStr)
	}

	// Check that links were found (should be mentioned in verbose output)
	if !strings.Contains(outputStr, "links_found") {
		t.Error("Expected to see links_found in verbose output")
	}

	// Check crawling statistics
	if !strings.Contains(outputStr, "crawled_urls=1") {
		t.Error("Expected to see crawled URLs count in output")
	}

	// Verify server statistics
	totalRequests, pathCounts := stats.GetStats()
	if totalRequests == 0 {
		t.Error("Expected server to receive requests")
	}

	t.Logf("Total requests: %d", totalRequests)
	t.Logf("Path counts: %+v", pathCounts)
}

func TestE2E_ConcurrencyStressTest(t *testing.T) {
	// Setup test server
	server, stats := createComplexTestServer()
	defer server.Close()

	// Build binary
	binaryPath, cleanup := setupE2ETest(t)
	defer cleanup()

	// Run stress test with high concurrency
	cmd := exec.Command(binaryPath,
		"--depth=2",
		"--concurrent=20",
		"--progress=false",
		server.URL)

	start := time.Now()
	output, err := cmd.CombinedOutput()
	duration := time.Since(start)

	// Verify results
	if err != nil {
		t.Fatalf("Concurrency stress test failed: %v\nOutput: %s", err, string(output))
	}

	// Verify performance characteristics
	totalRequests, _ := stats.GetStats()
	if totalRequests == 0 {
		t.Error("Expected server to receive requests")
	}

	t.Logf("Stress test completed in %v with %d requests", duration, totalRequests)

	// Basic performance check - should complete within reasonable time
	if duration > 30*time.Second {
		t.Errorf("Stress test took too long: %v", duration)
	}
}

func TestE2E_ErrorHandling(t *testing.T) {
	// Setup test server
	server, _ := createComplexTestServer()
	defer server.Close()

	// Build binary
	binaryPath, cleanup := setupE2ETest(t)
	defer cleanup()

	// Test with mixed success/error scenarios
	// Add error URLs to the server by creating a custom HTML page
	testHTML := fmt.Sprintf(`<!DOCTYPE html>
<html>
<body>
    <h1>Error Test Page</h1>
    <a href="%s/about">Valid Link</a>
    <a href="%s/error">Error Link</a>
    <a href="%s/not-found">Not Found Link</a>
</body>
</html>`, server.URL, server.URL, server.URL)

	// Create a temporary server with error links
	errorMux := http.NewServeMux()
	errorMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, testHTML)
	})
	errorMux.HandleFunc("/about", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<html><body><h1>About</h1></body></html>`)
	})
	errorMux.HandleFunc("/error", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal Server Error")
	})
	errorMux.HandleFunc("/not-found", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "Not Found")
	})

	errorServer := httptest.NewServer(errorMux)
	defer errorServer.Close()

	// Run crawl on error-prone server
	cmd := exec.Command(binaryPath,
		"--depth=2",
		"--verbose",
		"--progress=false",
		errorServer.URL)

	output, err := cmd.CombinedOutput()

	// Should not fail completely due to some errors
	if err != nil {
		// Log the error but continue - partial failures are acceptable
		t.Logf("Command had partial failures (expected): %v", err)
	}

	outputStr := string(output)

	// Should still crawl valid pages
	if !strings.Contains(outputStr, errorServer.URL) {
		t.Errorf("Expected output to contain base URL: %s", outputStr)
	}

	t.Logf("Error handling test output: %s", outputStr)
}

func TestE2E_OutputFormatValidation(t *testing.T) {
	// Setup test server
	server, _ := createComplexTestServer()
	defer server.Close()

	// Build binary
	binaryPath, cleanup := setupE2ETest(t)
	defer cleanup()

	// Run crawl
	cmd := exec.Command(binaryPath,
		"--depth=2",
		"--progress=false",
		server.URL)

	output, err := cmd.Output()

	// Verify results
	if err != nil {
		t.Fatalf("Output format test failed: %v", err)
	}

	outputStr := string(output)
	lines := strings.Split(strings.TrimSpace(outputStr), "\n")

	// Verify output format
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Each line should be a valid URL
		if !strings.HasPrefix(line, "http") {
			t.Errorf("Line %d is not a valid URL: %s", i+1, line)
		}

		// Should contain the server URL
		if !strings.Contains(line, server.URL) {
			t.Errorf("Line %d does not contain server URL: %s", i+1, line)
		}
	}

	// Should have found at least one URL
	if len(lines) < 1 {
		t.Errorf("Expected at least one URL, got %d: %v", len(lines), lines)
	}
}

func TestE2E_SignalHandling(t *testing.T) {
	// This test is more complex and may not work reliably in all environments
	// Skip in short test mode
	if testing.Short() {
		t.Skip("Skipping signal handling test in short mode")
	}

	// Setup test server with slow responses
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Include many links to make crawling take time
		html := `<html><body><h1>Slow Server</h1>`
		for i := 0; i < 100; i++ {
			html += fmt.Sprintf(`<a href="/page%d">Page %d</a>`, i, i)
		}
		html += `</body></html>`

		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, html)
	})

	mux.HandleFunc("/page", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1 * time.Second) // Slow response
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<html><body><h1>%s</h1></body></html>`, r.URL.Path)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	// Build binary
	binaryPath, cleanup := setupE2ETest(t)
	defer cleanup()

	// Start crawl process
	cmd := exec.Command(binaryPath,
		"--depth=3",
		"--concurrent=5",
		"--progress=false",
		server.URL)

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start crawl command: %v", err)
	}

	// Let it run for a bit
	time.Sleep(3 * time.Second)

	// Send interrupt signal
	if err := cmd.Process.Signal(os.Interrupt); err != nil {
		t.Fatalf("Failed to send interrupt signal: %v", err)
	}

	// Wait for graceful shutdown
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		// Process should exit gracefully
		if err != nil {
			t.Logf("Process exited with error (may be expected): %v", err)
		}
	case <-time.After(10 * time.Second):
		// Force kill if it doesn't stop gracefully
		cmd.Process.Kill()
		t.Error("Process did not stop gracefully within timeout")
	}
}
