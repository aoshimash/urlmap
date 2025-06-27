package shared

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// TestEnvironment represents a test environment setup
type TestEnvironment struct {
	Server     *httptest.Server
	BinaryPath string
	TempDir    string
}

// SetupTestEnvironment creates a complete test environment
func SetupTestEnvironment(t *testing.T) *TestEnvironment {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "crawld-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Create test server
	server := CreateBasicTestServer()

	// Build binary
	binaryPath := BuildTestBinary(t)

	return &TestEnvironment{
		Server:     server,
		BinaryPath: binaryPath,
		TempDir:    tempDir,
	}
}

// Cleanup cleans up the test environment
func (te *TestEnvironment) Cleanup() {
	if te.Server != nil {
		te.Server.Close()
	}
	if te.BinaryPath != "" {
		os.Remove(te.BinaryPath)
	}
	if te.TempDir != "" {
		os.RemoveAll(te.TempDir)
	}
}

// CreateBasicTestServer creates a basic test server for testing
func CreateBasicTestServer() *httptest.Server {
	mux := http.NewServeMux()

	// Basic HTML pages
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<!DOCTYPE html>
<html>
<head><title>Test Home</title></head>
<body>
    <h1>Test Home Page</h1>
    <a href="/page1">Page 1</a>
    <a href="/page2">Page 2</a>
    <a href="/nested/deep">Deep Nested</a>
</body>
</html>`)
	})

	mux.HandleFunc("/page1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<!DOCTYPE html>
<html>
<head><title>Page 1</title></head>
<body>
    <h1>Page 1</h1>
    <a href="/">Home</a>
    <a href="/page2">Page 2</a>
</body>
</html>`)
	})

	mux.HandleFunc("/page2", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<!DOCTYPE html>
<html>
<head><title>Page 2</title></head>
<body>
    <h1>Page 2</h1>
    <a href="/">Home</a>
    <a href="/page1">Page 1</a>
</body>
</html>`)
	})

	mux.HandleFunc("/nested/deep", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<!DOCTYPE html>
<html>
<head><title>Deep Page</title></head>
<body>
    <h1>Deep Nested Page</h1>
    <a href="/">Home</a>
</body>
</html>`)
	})

	return httptest.NewServer(mux)
}

// findProjectRoot finds the project root directory
func findProjectRoot() (string, error) {
	// Get the current file's directory
	_, filename, _, _ := runtime.Caller(0)
	currentDir := filepath.Dir(filename)

	// Walk up the directory tree to find go.mod
	for {
		if _, err := os.Stat(filepath.Join(currentDir, "go.mod")); err == nil {
			return currentDir, nil
		}

		parent := filepath.Dir(currentDir)
		if parent == currentDir {
			return "", fmt.Errorf("could not find project root (go.mod not found)")
		}
		currentDir = parent
	}
}

// BuildTestBinary builds the urlmap binary for testing
func BuildTestBinary(t *testing.T) string {
	tempDir := t.TempDir()
	binaryPath := filepath.Join(tempDir, "urlmap")

	// Change to project root
	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Fatalf("Failed to find project root: %v", err)
	}

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(oldDir)

	if err := os.Chdir(projectRoot); err != nil {
		t.Fatalf("Failed to change to project root: %v", err)
	}

	// Build the binary
	cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/urlmap")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build urlmap binary: %v", err)
	}

	return binaryPath
}

// RunCrawldCommand runs the crawld command with given arguments
func RunCrawldCommand(binaryPath string, args ...string) (string, string, error) {
	cmd := exec.Command(binaryPath, args...)

	// Set a reasonable timeout
	timeout := 30 * time.Second
	timer := time.AfterFunc(timeout, func() {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	})
	defer timer.Stop()

	output, err := cmd.CombinedOutput()

	// Separate stdout and stderr if needed
	// For now, return combined output
	return string(output), "", err
}

// AssertURLsInOutput checks if expected URLs are present in the output
func AssertURLsInOutput(t *testing.T, output string, expectedURLs []string) {
	for _, expectedURL := range expectedURLs {
		if !strings.Contains(output, expectedURL) {
			t.Errorf("Expected output to contain URL %s, but it was not found in: %s", expectedURL, output)
		}
	}
}

// AssertNoError checks that no error occurred
func AssertNoError(t *testing.T, err error, context string) {
	if err != nil {
		t.Fatalf("%s: %v", context, err)
	}
}

// AssertError checks that an error occurred
func AssertError(t *testing.T, err error, context string) {
	if err == nil {
		t.Fatalf("%s: expected error but got none", context)
	}
}

// AssertStringContains checks that a string contains a substring
func AssertStringContains(t *testing.T, str, substr, context string) {
	if !strings.Contains(str, substr) {
		t.Errorf("%s: expected string to contain %q, got: %s", context, substr, str)
	}
}

// AssertStringNotContains checks that a string does not contain a substring
func AssertStringNotContains(t *testing.T, str, substr, context string) {
	if strings.Contains(str, substr) {
		t.Errorf("%s: expected string to not contain %q, got: %s", context, substr, str)
	}
}

// CreateServerWithError creates a test server that returns errors for certain paths
func CreateServerWithError(errorPaths []string) *httptest.Server {
	mux := http.NewServeMux()

	// Default handler
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Check if this path should return an error
		for _, errorPath := range errorPaths {
			if r.URL.Path == errorPath {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprint(w, "Internal Server Error")
				return
			}
		}

		// Normal response
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head><title>Test Page</title></head>
<body>
    <h1>Test Page: %s</h1>
    <a href="/test1">Test 1</a>
    <a href="/test2">Test 2</a>
    <a href="/error">Error Page</a>
</body>
</html>`, r.URL.Path)
	})

	return httptest.NewServer(mux)
}

// CreateSlowServer creates a test server with slow responses
func CreateSlowServer(delay time.Duration) *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(delay)
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<!DOCTYPE html>
<html>
<head><title>Slow Page</title></head>
<body>
    <h1>Slow Loading Page</h1>
    <p>This page took a while to load.</p>
</body>
</html>`)
	})

	return httptest.NewServer(mux)
}

// GetProjectRoot returns the project root directory
func GetProjectRoot() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "..", "..")
}

// CreateComplexSiteServer creates a test server with a complex site structure
func CreateComplexSiteServer() *httptest.Server {
	mux := http.NewServeMux()

	// Home page with multiple navigation links
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<!DOCTYPE html>
<html>
<head><title>Complex Site Home</title></head>
<body>
    <nav>
        <a href="/about">About</a>
        <a href="/products">Products</a>
        <a href="/blog">Blog</a>
        <a href="/contact">Contact</a>
    </nav>
    <main>
        <h1>Welcome to Complex Site</h1>
        <p>This is a complex site for testing.</p>
        <section>
            <a href="/services">Services</a>
            <a href="/support">Support</a>
        </section>
    </main>
</body>
</html>`)
	})

	// About page
	mux.HandleFunc("/about", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<!DOCTYPE html>
<html>
<head><title>About Us</title></head>
<body>
    <h1>About Us</h1>
    <p>Learn about our company.</p>
    <a href="/">Home</a>
    <a href="/team">Team</a>
    <a href="/history">History</a>
</body>
</html>`)
	})

	// Products page with subcategories
	mux.HandleFunc("/products", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<!DOCTYPE html>
<html>
<head><title>Products</title></head>
<body>
    <h1>Our Products</h1>
    <div class="categories">
        <a href="/products/software">Software</a>
        <a href="/products/hardware">Hardware</a>
        <a href="/products/services">Services</a>
    </div>
    <a href="/">Home</a>
</body>
</html>`)
	})

	// Blog page with posts
	mux.HandleFunc("/blog", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<!DOCTYPE html>
<html>
<head><title>Blog</title></head>
<body>
    <h1>Blog</h1>
    <div class="posts">
        <a href="/blog/post1">How to Use Our Product</a>
        <a href="/blog/post2">Latest Updates</a>
        <a href="/blog/post3">Best Practices</a>
    </div>
    <a href="/">Home</a>
</body>
</html>`)
	})

	// Dynamic handlers for subcategories
	mux.HandleFunc("/products/", func(w http.ResponseWriter, r *http.Request) {
		category := strings.TrimPrefix(r.URL.Path, "/products/")
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head><title>%s Products</title></head>
<body>
    <h1>%s Products</h1>
    <p>Details about our %s products.</p>
    <a href="/products">Back to Products</a>
    <a href="/">Home</a>
</body>
</html>`, category, category, category)
	})

	mux.HandleFunc("/blog/", func(w http.ResponseWriter, r *http.Request) {
		post := strings.TrimPrefix(r.URL.Path, "/blog/")
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head><title>Blog Post: %s</title></head>
<body>
    <h1>Blog Post: %s</h1>
    <p>Content of the blog post.</p>
    <a href="/blog">Back to Blog</a>
    <a href="/">Home</a>
</body>
</html>`, post, post)
	})

	return httptest.NewServer(mux)
}
