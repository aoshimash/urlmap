package detector

import (
	"sync"
	"time"
)

// DetectionCache SPA検出結果のキャッシュ
type DetectionCache struct {
	cache map[string]*DetectionResult
	mutex sync.RWMutex
	ttl   time.Duration
}

// NewDetectionCache 新しい検出キャッシュを作成
func NewDetectionCache(ttl time.Duration) *DetectionCache {
	return &DetectionCache{
		cache: make(map[string]*DetectionResult),
		ttl:   ttl,
	}
}

// Get キャッシュから検出結果を取得
func (dc *DetectionCache) Get(domain string) (*DetectionResult, bool) {
	dc.mutex.RLock()
	defer dc.mutex.RUnlock()

	result, exists := dc.cache[domain]
	if !exists {
		return nil, false
	}

	// TTL確認
	if time.Since(result.Timestamp) > dc.ttl {
		delete(dc.cache, domain)
		return nil, false
	}

	return result, true
}

// Set キャッシュに検出結果を保存
func (dc *DetectionCache) Set(domain string, result *DetectionResult) {
	dc.mutex.Lock()
	defer dc.mutex.Unlock()

	dc.cache[domain] = result
}

// Clear キャッシュをクリア
func (dc *DetectionCache) Clear() {
	dc.mutex.Lock()
	defer dc.mutex.Unlock()

	dc.cache = make(map[string]*DetectionResult)
}

// Cleanup 期限切れのエントリを削除
func (dc *DetectionCache) Cleanup() {
	dc.mutex.Lock()
	defer dc.mutex.Unlock()

	now := time.Now()
	for domain, result := range dc.cache {
		if now.Sub(result.Timestamp) > dc.ttl {
			delete(dc.cache, domain)
		}
	}
}

// Size キャッシュサイズを取得
func (dc *DetectionCache) Size() int {
	dc.mutex.RLock()
	defer dc.mutex.RUnlock()

	return len(dc.cache)
}

// GetStats キャッシュ統計を取得
func (dc *DetectionCache) GetStats() map[string]interface{} {
	dc.mutex.RLock()
	defer dc.mutex.RUnlock()

	// 期限切れエントリのカウント
	expiredCount := 0
	now := time.Now()
	for _, result := range dc.cache {
		if now.Sub(result.Timestamp) > dc.ttl {
			expiredCount++
		}
	}

	return map[string]interface{}{
		"total_entries":   len(dc.cache),
		"expired_entries": expiredCount,
		"ttl_seconds":     dc.ttl.Seconds(),
	}
}
