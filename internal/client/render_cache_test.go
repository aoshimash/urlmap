package client

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestNewRenderCache(t *testing.T) {
	cache := NewRenderCache(100, 5*time.Minute)

	if cache == nil {
		t.Fatal("NewRenderCache returned nil")
	}

	if cache.maxSize != 100 {
		t.Errorf("Expected maxSize 100, got %d", cache.maxSize)
	}

	if cache.ttl != 5*time.Minute {
		t.Errorf("Expected ttl 5 minutes, got %v", cache.ttl)
	}

	if cache.Size() != 0 {
		t.Errorf("Expected initial size 0, got %d", cache.Size())
	}
}

func TestRenderCache_SetAndGet(t *testing.T) {
	cache := NewRenderCache(10, 1*time.Hour)

	// Test basic set and get
	url := "https://example.com"
	content := "<html><body>Test</body></html>"
	headers := map[string]string{"Content-Type": "text/html"}
	statusCode := 200

	cache.Set(url, content, headers, statusCode)

	// Test successful get
	entry, found := cache.Get(url)
	if !found {
		t.Error("Expected to find cached entry")
	}

	if entry.URL != url {
		t.Errorf("Expected URL %s, got %s", url, entry.URL)
	}

	if entry.Content != content {
		t.Errorf("Expected content %s, got %s", content, entry.Content)
	}

	if entry.StatusCode != statusCode {
		t.Errorf("Expected status code %d, got %d", statusCode, entry.StatusCode)
	}

	// Test cache size
	if cache.Size() != 1 {
		t.Errorf("Expected cache size 1, got %d", cache.Size())
	}
}

func TestRenderCache_TTLExpiration(t *testing.T) {
	cache := NewRenderCache(10, 100*time.Millisecond)

	url := "https://example.com"
	content := "<html><body>Test</body></html>"
	cache.Set(url, content, nil, 200)

	// Entry should be found immediately
	_, found := cache.Get(url)
	if !found {
		t.Error("Expected to find cached entry immediately after setting")
	}

	// Wait for TTL to expire
	time.Sleep(150 * time.Millisecond)

	// Entry should be expired and removed
	_, found = cache.Get(url)
	if found {
		t.Error("Expected entry to be expired and not found")
	}

	// Size should be 0 after expiration
	if cache.Size() != 0 {
		t.Errorf("Expected cache size 0 after expiration, got %d", cache.Size())
	}
}

func TestRenderCache_Eviction(t *testing.T) {
	cache := NewRenderCache(3, 1*time.Hour)

	// Fill cache to capacity
	cache.Set("url1", "content1", nil, 200)
	time.Sleep(10 * time.Millisecond)
	cache.Set("url2", "content2", nil, 200)
	time.Sleep(10 * time.Millisecond)
	cache.Set("url3", "content3", nil, 200)

	if cache.Size() != 3 {
		t.Errorf("Expected cache size 3, got %d", cache.Size())
	}

	// Access url1 to make it more recently used
	cache.Get("url1")

	// Add a fourth entry, should evict the oldest (url2)
	cache.Set("url4", "content4", nil, 200)

	// Cache should still have size 3
	if cache.Size() != 3 {
		t.Errorf("Expected cache size 3 after eviction, got %d", cache.Size())
	}

	// url2 should have been evicted (oldest non-accessed)
	_, found := cache.Get("url2")
	if found {
		t.Error("Expected url2 to be evicted")
	}

	// Other URLs should still be present
	_, found = cache.Get("url1")
	if !found {
		t.Error("Expected url1 to still be in cache")
	}

	_, found = cache.Get("url3")
	if !found {
		t.Error("Expected url3 to still be in cache")
	}

	_, found = cache.Get("url4")
	if !found {
		t.Error("Expected url4 to still be in cache")
	}
}

func TestRenderCache_Delete(t *testing.T) {
	cache := NewRenderCache(10, 1*time.Hour)

	url := "https://example.com"
	cache.Set(url, "content", nil, 200)

	// Verify entry exists
	_, found := cache.Get(url)
	if !found {
		t.Error("Expected to find cached entry")
	}

	// Delete the entry
	cache.Delete(url)

	// Verify entry is gone
	_, found = cache.Get(url)
	if found {
		t.Error("Expected entry to be deleted")
	}

	// Verify size is 0
	if cache.Size() != 0 {
		t.Errorf("Expected cache size 0 after deletion, got %d", cache.Size())
	}
}

func TestRenderCache_Clear(t *testing.T) {
	cache := NewRenderCache(10, 1*time.Hour)

	// Add multiple entries
	for i := 0; i < 5; i++ {
		url := fmt.Sprintf("https://example.com/page%d", i)
		cache.Set(url, fmt.Sprintf("content%d", i), nil, 200)
	}

	if cache.Size() != 5 {
		t.Errorf("Expected cache size 5, got %d", cache.Size())
	}

	// Clear the cache
	cache.Clear()

	// Verify all entries are gone
	if cache.Size() != 0 {
		t.Errorf("Expected cache size 0 after clear, got %d", cache.Size())
	}

	// Verify individual entries are gone
	for i := 0; i < 5; i++ {
		url := fmt.Sprintf("https://example.com/page%d", i)
		_, found := cache.Get(url)
		if found {
			t.Errorf("Expected %s to be cleared", url)
		}
	}
}

func TestRenderCache_UpdateExisting(t *testing.T) {
	cache := NewRenderCache(10, 1*time.Hour)

	url := "https://example.com"
	cache.Set(url, "original content", nil, 200)

	if cache.Size() != 1 {
		t.Errorf("Expected cache size 1, got %d", cache.Size())
	}

	// Update the same URL
	cache.Set(url, "updated content", nil, 200)

	// Size should still be 1
	if cache.Size() != 1 {
		t.Errorf("Expected cache size 1 after update, got %d", cache.Size())
	}

	// Content should be updated
	entry, found := cache.Get(url)
	if !found {
		t.Error("Expected to find cached entry")
	}

	if entry.Content != "updated content" {
		t.Errorf("Expected updated content, got %s", entry.Content)
	}
}

func TestRenderCache_ConcurrentAccess(t *testing.T) {
	cache := NewRenderCache(100, 1*time.Hour)

	// Number of goroutines and operations
	numGoroutines := 10
	numOperations := 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Launch concurrent goroutines
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			for j := 0; j < numOperations; j++ {
				url := fmt.Sprintf("https://example.com/page%d-%d", id, j)

				// Alternate between set and get operations
				if j%2 == 0 {
					cache.Set(url, fmt.Sprintf("content%d-%d", id, j), nil, 200)
				} else {
					// Try to get a previously set URL
					prevURL := fmt.Sprintf("https://example.com/page%d-%d", id, j-1)
					cache.Get(prevURL)
				}

				// Occasionally delete an entry
				if j%10 == 0 && j > 0 {
					deleteURL := fmt.Sprintf("https://example.com/page%d-%d", id, j-10)
					cache.Delete(deleteURL)
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify cache is in a valid state
	size := cache.Size()
	if size < 0 || size > 100 {
		t.Errorf("Cache size out of bounds: %d", size)
	}

	// Verify we can still perform operations
	cache.Set("final-test", "final-content", nil, 200)
	entry, found := cache.Get("final-test")
	if !found {
		t.Error("Failed to set and get after concurrent access")
	}
	if entry.Content != "final-content" {
		t.Errorf("Expected final-content, got %s", entry.Content)
	}
}

func TestRenderCache_Stats(t *testing.T) {
	cache := NewRenderCache(10, 100*time.Millisecond)

	// Add some entries
	cache.Set("url1", "content1", nil, 200)
	cache.Set("url2", "content2", nil, 200)

	// Wait for entries to expire
	time.Sleep(150 * time.Millisecond)

	// Add a fresh entry
	cache.Set("url3", "content3", nil, 200)

	stats := cache.Stats()

	// Check stats structure
	if stats["max_size"].(int64) != 10 {
		t.Errorf("Expected max_size 10, got %v", stats["max_size"])
	}

	if stats["ttl_seconds"].(float64) != 0.1 {
		t.Errorf("Expected ttl_seconds 0.1, got %v", stats["ttl_seconds"])
	}

	// Size should be 3 (2 expired + 1 valid)
	if stats["size"].(int) != 3 {
		t.Errorf("Expected size 3, got %v", stats["size"])
	}

	if stats["valid_entries"].(int) != 1 {
		t.Errorf("Expected 1 valid entry, got %v", stats["valid_entries"])
	}

	if stats["expired_entries"].(int) != 2 {
		t.Errorf("Expected 2 expired entries, got %v", stats["expired_entries"])
	}
}

func TestRenderCache_ComplexEviction(t *testing.T) {
	cache := NewRenderCache(5, 1*time.Hour)

	// Fill cache with entries accessed at different times
	urls := []string{"url1", "url2", "url3", "url4", "url5"}
	for i, url := range urls {
		cache.Set(url, fmt.Sprintf("content%d", i), nil, 200)
		time.Sleep(10 * time.Millisecond)

		// Access some entries to change their access time
		if i == 1 {
			cache.Get("url1") // Make url1 more recently accessed than url2
		}
	}

	// Now access pattern: url5 (newest), url4, url3, url2, url1 (recently accessed)
	// Add new entries to trigger eviction
	cache.Set("url6", "content6", nil, 200)

	// url2 should be evicted (oldest access time among the original entries)
	_, found := cache.Get("url2")
	if found {
		t.Error("Expected url2 to be evicted")
	}

	// Check other URLs are still present
	expectedPresent := []string{"url1", "url3", "url4", "url5", "url6"}
	for _, url := range expectedPresent {
		_, found := cache.Get(url)
		if !found {
			t.Errorf("Expected %s to be in cache", url)
		}
	}
}

