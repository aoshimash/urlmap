package cache

import (
	"container/list"
	"sync"
	"sync/atomic"
	"time"
)

// RenderCache レンダリング結果をキャッシュする高性能キャッシュ
type RenderCache struct {
	cache   map[string]*list.Element
	lru     *list.List
	mutex   sync.RWMutex
	maxSize int
	ttl     time.Duration
	stats   *CacheStats
}

// CacheEntry キャッシュエントリを表す構造体
type CacheEntry struct {
	Key         string
	HTML        string
	Timestamp   time.Time
	AccessCount int64
	Size        int
}

// CacheStats キャッシュ統計情報
type CacheStats struct {
	Hits      int64
	Misses    int64
	Evictions int64
	Size      int64
}

// NewRenderCache 新しいレンダリングキャッシュを作成
func NewRenderCache(maxSize int, ttl time.Duration) *RenderCache {
	return &RenderCache{
		cache:   make(map[string]*list.Element),
		lru:     list.New(),
		maxSize: maxSize,
		ttl:     ttl,
		stats:   &CacheStats{},
	}
}

// Get キャッシュからレンダリング結果を取得
func (rc *RenderCache) Get(url string) (string, bool) {
	rc.mutex.RLock()
	element, exists := rc.cache[url]
	if !exists {
		rc.mutex.RUnlock()
		atomic.AddInt64(&rc.stats.Misses, 1)
		return "", false
	}
	rc.mutex.RUnlock()

	entry := element.Value.(*CacheEntry)

	// TTL確認
	if time.Since(entry.Timestamp) > rc.ttl {
		rc.mutex.Lock()
		rc.removeElement(element)
		rc.mutex.Unlock()
		atomic.AddInt64(&rc.stats.Evictions, 1)
		return "", false
	}

	// アクセスカウントを増加
	atomic.AddInt64(&entry.AccessCount, 1)

	// LRUリストの先頭に移動
	rc.mutex.Lock()
	rc.lru.MoveToFront(element)
	rc.mutex.Unlock()

	atomic.AddInt64(&rc.stats.Hits, 1)
	return entry.HTML, true
}

// Set レンダリング結果をキャッシュに保存
func (rc *RenderCache) Set(url, html string) {
	rc.mutex.Lock()
	defer rc.mutex.Unlock()

	// サイズ制限チェック
	if len(rc.cache) >= rc.maxSize {
		rc.evictLRU()
	}

	entry := &CacheEntry{
		Key:         url,
		HTML:        html,
		Timestamp:   time.Now(),
		AccessCount: 1,
		Size:        len(html),
	}

	element := rc.lru.PushFront(entry)
	rc.cache[url] = element

	atomic.StoreInt64(&rc.stats.Size, int64(len(rc.cache)))
}

// evictLRU LRUアルゴリズムでエントリを削除
func (rc *RenderCache) evictLRU() {
	if rc.lru.Len() == 0 {
		return
	}

	// 最後の要素（最も古い）を削除
	element := rc.lru.Back()
	if element != nil {
		entry := element.Value.(*CacheEntry)
		delete(rc.cache, entry.Key)
		rc.lru.Remove(element)
		atomic.AddInt64(&rc.stats.Evictions, 1)
	}
}

// removeElement 特定のエントリを削除
func (rc *RenderCache) removeElement(element *list.Element) {
	entry := element.Value.(*CacheEntry)
	delete(rc.cache, entry.Key)
	rc.lru.Remove(element)
}

// Clear キャッシュをクリア
func (rc *RenderCache) Clear() {
	rc.mutex.Lock()
	defer rc.mutex.Unlock()

	rc.cache = make(map[string]*list.Element)
	rc.lru.Init()
	atomic.StoreInt64(&rc.stats.Size, 0)
}

// Size キャッシュサイズを取得
func (rc *RenderCache) Size() int {
	rc.mutex.RLock()
	defer rc.mutex.RUnlock()
	return len(rc.cache)
}

// GetStats キャッシュ統計を取得
func (rc *RenderCache) GetStats() map[string]interface{} {
	hits := atomic.LoadInt64(&rc.stats.Hits)
	misses := atomic.LoadInt64(&rc.stats.Misses)
	evictions := atomic.LoadInt64(&rc.stats.Evictions)
	size := atomic.LoadInt64(&rc.stats.Size)

	var hitRate float64
	if hits+misses > 0 {
		hitRate = float64(hits) / float64(hits+misses) * 100
	}

	return map[string]interface{}{
		"hits":      hits,
		"misses":    misses,
		"evictions": evictions,
		"size":      size,
		"hit_rate":  hitRate,
		"max_size":  rc.maxSize,
		"ttl":       rc.ttl.String(),
	}
}

// Cleanup 期限切れのエントリを削除
func (rc *RenderCache) Cleanup() int {
	rc.mutex.Lock()
	defer rc.mutex.Unlock()

	removed := 0
	for _, element := range rc.cache {
		entry := element.Value.(*CacheEntry)
		if time.Since(entry.Timestamp) > rc.ttl {
			rc.removeElement(element)
			removed++
		}
	}

	atomic.AddInt64(&rc.stats.Evictions, int64(removed))
	atomic.StoreInt64(&rc.stats.Size, int64(len(rc.cache)))

	return removed
}
