package client

import (
	"sync"
	"sync/atomic"
	"time"
)

// CacheEntry represents a cached render result
type CacheEntry struct {
	URL        string
	Content    string
	Headers    map[string]string
	StatusCode int
	Timestamp  time.Time
}

// RenderCache is a thread-safe in-memory cache for rendered pages
type RenderCache struct {
	// Thread-safe storage
	entries sync.Map // key: string (URL), value: *CacheEntry

	// Configuration
	maxSize int64         // Maximum number of entries
	ttl     time.Duration // Time to live for cache entries

	// Size tracking
	size int64 // Current number of entries (atomic)

	// For eviction
	accessOrder sync.Map // key: string (URL), value: time.Time (last access time)
	mu          sync.Mutex
}

// NewRenderCache creates a new render cache with the given configuration
func NewRenderCache(maxSize int, ttl time.Duration) *RenderCache {
	return &RenderCache{
		maxSize: int64(maxSize),
		ttl:     ttl,
		size:    0,
	}
}

// Get retrieves a cached entry if it exists and is not expired
func (c *RenderCache) Get(url string) (*CacheEntry, bool) {
	// Try to get the entry
	value, exists := c.entries.Load(url)
	if !exists {
		return nil, false
	}

	entry := value.(*CacheEntry)

	// Check if the entry has expired
	if time.Since(entry.Timestamp) > c.ttl {
		// Remove expired entry
		c.Delete(url)
		return nil, false
	}

	// Update access time
	c.accessOrder.Store(url, time.Now())

	return entry, true
}

// Set stores a new cache entry, evicting oldest entries if necessary
func (c *RenderCache) Set(url string, content string, headers map[string]string, statusCode int) {
	// Create new entry
	entry := &CacheEntry{
		URL:        url,
		Content:    content,
		Headers:    headers,
		StatusCode: statusCode,
		Timestamp:  time.Now(),
	}

	// Check if we're updating an existing entry
	_, exists := c.entries.Load(url)

	// Store the entry
	c.entries.Store(url, entry)
	c.accessOrder.Store(url, time.Now())

	// Update size if this is a new entry
	if !exists {
		newSize := atomic.AddInt64(&c.size, 1)

		// Check if we need to evict
		if newSize > c.maxSize {
			c.evictOldest()
		}
	}
}

// Delete removes an entry from the cache
func (c *RenderCache) Delete(url string) {
	if _, exists := c.entries.LoadAndDelete(url); exists {
		c.accessOrder.Delete(url)
		atomic.AddInt64(&c.size, -1)
	}
}

// evictOldest removes the least recently accessed entries until we're under maxSize
func (c *RenderCache) evictOldest() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Collect all entries with their access times
	type accessEntry struct {
		url        string
		accessTime time.Time
	}

	var entries []accessEntry
	c.accessOrder.Range(func(key, value interface{}) bool {
		entries = append(entries, accessEntry{
			url:        key.(string),
			accessTime: value.(time.Time),
		})
		return true
	})

	// Sort by access time (oldest first)
	// Using simple bubble sort for small datasets
	for i := 0; i < len(entries); i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[i].accessTime.After(entries[j].accessTime) {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}

	// Evict oldest entries until we're under maxSize
	currentSize := atomic.LoadInt64(&c.size)
	for i := 0; i < len(entries) && currentSize > c.maxSize; i++ {
		c.Delete(entries[i].url)
		currentSize = atomic.LoadInt64(&c.size)
	}
}

// Size returns the current number of entries in the cache
func (c *RenderCache) Size() int {
	return int(atomic.LoadInt64(&c.size))
}

// Clear removes all entries from the cache
func (c *RenderCache) Clear() {
	c.entries.Range(func(key, value interface{}) bool {
		c.entries.Delete(key)
		c.accessOrder.Delete(key)
		return true
	})
	atomic.StoreInt64(&c.size, 0)
}

// Stats returns cache statistics
func (c *RenderCache) Stats() map[string]interface{} {
	var expiredCount int
	var validCount int

	c.entries.Range(func(key, value interface{}) bool {
		entry := value.(*CacheEntry)
		if time.Since(entry.Timestamp) > c.ttl {
			expiredCount++
		} else {
			validCount++
		}
		return true
	})

	return map[string]interface{}{
		"size":            c.Size(),
		"max_size":        c.maxSize,
		"ttl_seconds":     c.ttl.Seconds(),
		"valid_entries":   validCount,
		"expired_entries": expiredCount,
	}
}

