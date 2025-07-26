package metrics

import (
	"sync"
	"sync/atomic"
	"time"
)

// JSMetrics JavaScriptレンダリングのパフォーマンスメトリクス
type JSMetrics struct {
	// 基本統計
	RenderCount     int64
	TotalDuration   int64 // ナノ秒単位
	AverageDuration int64 // ナノ秒単位
	ErrorCount      int64

	// キャッシュ統計
	CacheHits   int64
	CacheMisses int64

	// ブラウザプール統計
	BrowserPoolSize  int64
	ActiveContexts   int64
	ContextCreations int64
	ContextReuses    int64

	// 並列処理統計
	ConcurrentRenders int64
	MaxConcurrent     int64

	// リソース使用量
	MemoryUsage int64 // バイト単位
	PeakMemory  int64 // バイト単位

	// 時間統計
	StartTime  time.Time
	LastUpdate time.Time

	mutex sync.RWMutex
}

// NewJSMetrics 新しいメトリクスインスタンスを作成
func NewJSMetrics() *JSMetrics {
	return &JSMetrics{
		StartTime:  time.Now(),
		LastUpdate: time.Now(),
	}
}

// RecordRender レンダリング結果を記録
func (m *JSMetrics) RecordRender(duration time.Duration, err error) {
	atomic.AddInt64(&m.RenderCount, 1)
	atomic.AddInt64(&m.TotalDuration, int64(duration))

	// 平均時間を更新
	m.mutex.Lock()
	count := atomic.LoadInt64(&m.RenderCount)
	if count > 0 {
		m.AverageDuration = atomic.LoadInt64(&m.TotalDuration) / count
	}
	m.LastUpdate = time.Now()
	m.mutex.Unlock()

	if err != nil {
		atomic.AddInt64(&m.ErrorCount, 1)
	}
}

// RecordCacheHit キャッシュヒットを記録
func (m *JSMetrics) RecordCacheHit() {
	atomic.AddInt64(&m.CacheHits, 1)
}

// RecordCacheMiss キャッシュミスを記録
func (m *JSMetrics) RecordCacheMiss() {
	atomic.AddInt64(&m.CacheMisses, 1)
}

// SetBrowserPoolSize ブラウザプールサイズを設定
func (m *JSMetrics) SetBrowserPoolSize(size int) {
	atomic.StoreInt64(&m.BrowserPoolSize, int64(size))
}

// SetActiveContexts アクティブコンテキスト数を設定
func (m *JSMetrics) SetActiveContexts(count int) {
	atomic.StoreInt64(&m.ActiveContexts, int64(count))
}

// RecordContextCreation コンテキスト作成を記録
func (m *JSMetrics) RecordContextCreation() {
	atomic.AddInt64(&m.ContextCreations, 1)
}

// RecordContextReuse コンテキスト再利用を記録
func (m *JSMetrics) RecordContextReuse() {
	atomic.AddInt64(&m.ContextReuses, 1)
}

// SetConcurrentRenders 並列レンダリング数を設定
func (m *JSMetrics) SetConcurrentRenders(count int) {
	atomic.StoreInt64(&m.ConcurrentRenders, int64(count))

	// 最大値を更新
	for {
		max := atomic.LoadInt64(&m.MaxConcurrent)
		if int64(count) <= max {
			break
		}
		if atomic.CompareAndSwapInt64(&m.MaxConcurrent, max, int64(count)) {
			break
		}
	}
}

// SetMemoryUsage メモリ使用量を設定
func (m *JSMetrics) SetMemoryUsage(bytes int) {
	atomic.StoreInt64(&m.MemoryUsage, int64(bytes))

	// ピークメモリを更新
	for {
		peak := atomic.LoadInt64(&m.PeakMemory)
		if int64(bytes) <= peak {
			break
		}
		if atomic.CompareAndSwapInt64(&m.PeakMemory, peak, int64(bytes)) {
			break
		}
	}
}

// GetSummary メトリクスのサマリーを取得
func (m *JSMetrics) GetSummary() map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	renderCount := atomic.LoadInt64(&m.RenderCount)
	totalDuration := atomic.LoadInt64(&m.TotalDuration)
	errorCount := atomic.LoadInt64(&m.ErrorCount)
	cacheHits := atomic.LoadInt64(&m.CacheHits)
	cacheMisses := atomic.LoadInt64(&m.CacheMisses)

	var errorRate float64
	if renderCount > 0 {
		errorRate = float64(errorCount) / float64(renderCount) * 100
	}

	var cacheHitRate float64
	if cacheHits+cacheMisses > 0 {
		cacheHitRate = float64(cacheHits) / float64(cacheHits+cacheMisses) * 100
	}

	var avgDuration time.Duration
	if renderCount > 0 {
		avgDuration = time.Duration(totalDuration / renderCount)
	}

	uptime := time.Since(m.StartTime)

	return map[string]interface{}{
		"renders_total":      renderCount,
		"renders_successful": renderCount - errorCount,
		"renders_failed":     errorCount,
		"error_rate_percent": errorRate,
		"total_duration":     time.Duration(totalDuration),
		"average_duration":   avgDuration,
		"uptime":             uptime,
		"cache_hits":         cacheHits,
		"cache_misses":       cacheMisses,
		"cache_hit_rate":     cacheHitRate,
		"browser_pool_size":  atomic.LoadInt64(&m.BrowserPoolSize),
		"active_contexts":    atomic.LoadInt64(&m.ActiveContexts),
		"context_creations":  atomic.LoadInt64(&m.ContextCreations),
		"context_reuses":     atomic.LoadInt64(&m.ContextReuses),
		"concurrent_renders": atomic.LoadInt64(&m.ConcurrentRenders),
		"max_concurrent":     atomic.LoadInt64(&m.MaxConcurrent),
		"memory_usage_bytes": atomic.LoadInt64(&m.MemoryUsage),
		"peak_memory_bytes":  atomic.LoadInt64(&m.PeakMemory),
		"last_update":        m.LastUpdate,
	}
}

// GetPerformanceReport パフォーマンスレポートを取得
func (m *JSMetrics) GetPerformanceReport() map[string]interface{} {
	summary := m.GetSummary()

	renderCount := summary["renders_total"].(int64)
	uptime := summary["uptime"].(time.Duration)

	var throughput float64
	if uptime.Seconds() > 0 {
		throughput = float64(renderCount) / uptime.Seconds()
	}

	avgDuration := summary["average_duration"].(time.Duration)
	errorRate := summary["error_rate_percent"].(float64)
	cacheHitRate := summary["cache_hit_rate"].(float64)

	// パフォーマンス評価
	var performanceGrade string
	switch {
	case throughput >= 10 && avgDuration < 2*time.Second && errorRate < 5:
		performanceGrade = "A"
	case throughput >= 5 && avgDuration < 5*time.Second && errorRate < 10:
		performanceGrade = "B"
	case throughput >= 2 && avgDuration < 10*time.Second && errorRate < 20:
		performanceGrade = "C"
	default:
		performanceGrade = "D"
	}

	return map[string]interface{}{
		"throughput_renders_per_second": throughput,
		"average_duration_ms":           avgDuration.Milliseconds(),
		"error_rate_percent":            errorRate,
		"cache_hit_rate_percent":        cacheHitRate,
		"performance_grade":             performanceGrade,
		"uptime_seconds":                uptime.Seconds(),
		"total_renders":                 renderCount,
		"successful_renders":            summary["renders_successful"],
		"failed_renders":                summary["renders_failed"],
	}
}

// Reset メトリクスをリセット
func (m *JSMetrics) Reset() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	atomic.StoreInt64(&m.RenderCount, 0)
	atomic.StoreInt64(&m.TotalDuration, 0)
	atomic.StoreInt64(&m.AverageDuration, 0)
	atomic.StoreInt64(&m.ErrorCount, 0)
	atomic.StoreInt64(&m.CacheHits, 0)
	atomic.StoreInt64(&m.CacheMisses, 0)
	atomic.StoreInt64(&m.ContextCreations, 0)
	atomic.StoreInt64(&m.ContextReuses, 0)
	atomic.StoreInt64(&m.ConcurrentRenders, 0)
	atomic.StoreInt64(&m.MaxConcurrent, 0)
	atomic.StoreInt64(&m.MemoryUsage, 0)
	atomic.StoreInt64(&m.PeakMemory, 0)

	m.StartTime = time.Now()
	m.LastUpdate = time.Now()
}

// IsHealthy メトリクスが健全かどうかを判定
func (m *JSMetrics) IsHealthy() bool {
	summary := m.GetSummary()

	errorRate := summary["error_rate_percent"].(float64)
	avgDuration := summary["average_duration"].(time.Duration)

	// エラー率が20%以下、平均時間が10秒以下
	return errorRate <= 20 && avgDuration <= 10*time.Second
}
