package metrics

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewJSMetrics(t *testing.T) {
	metrics := NewJSMetrics()

	if metrics.RenderCount != 0 {
		t.Errorf("Expected initial RenderCount: 0, got: %d", metrics.RenderCount)
	}

	if metrics.ErrorCount != 0 {
		t.Errorf("Expected initial ErrorCount: 0, got: %d", metrics.ErrorCount)
	}

	if metrics.StartTime.IsZero() {
		t.Error("Expected StartTime to be set")
	}
}

func TestRecordRender(t *testing.T) {
	metrics := NewJSMetrics()

	// 成功したレンダリング
	metrics.RecordRender(2*time.Second, nil)

	if atomic.LoadInt64(&metrics.RenderCount) != 1 {
		t.Errorf("Expected RenderCount: 1, got: %d", metrics.RenderCount)
	}

	if atomic.LoadInt64(&metrics.ErrorCount) != 0 {
		t.Errorf("Expected ErrorCount: 0, got: %d", metrics.ErrorCount)
	}

	// 失敗したレンダリング
	metrics.RecordRender(1*time.Second, errors.New("test error"))

	if atomic.LoadInt64(&metrics.RenderCount) != 2 {
		t.Errorf("Expected RenderCount: 2, got: %d", metrics.RenderCount)
	}

	if atomic.LoadInt64(&metrics.ErrorCount) != 1 {
		t.Errorf("Expected ErrorCount: 1, got: %d", metrics.ErrorCount)
	}
}

func TestCacheMetrics(t *testing.T) {
	metrics := NewJSMetrics()

	// キャッシュヒット
	metrics.RecordCacheHit()
	metrics.RecordCacheHit()

	// キャッシュミス
	metrics.RecordCacheMiss()

	if atomic.LoadInt64(&metrics.CacheHits) != 2 {
		t.Errorf("Expected CacheHits: 2, got: %d", metrics.CacheHits)
	}

	if atomic.LoadInt64(&metrics.CacheMisses) != 1 {
		t.Errorf("Expected CacheMisses: 1, got: %d", metrics.CacheMisses)
	}
}

func TestBrowserPoolMetrics(t *testing.T) {
	metrics := NewJSMetrics()

	metrics.SetBrowserPoolSize(5)
	metrics.SetActiveContexts(3)
	metrics.RecordContextCreation()
	metrics.RecordContextReuse()

	if atomic.LoadInt64(&metrics.BrowserPoolSize) != 5 {
		t.Errorf("Expected BrowserPoolSize: 5, got: %d", metrics.BrowserPoolSize)
	}

	if atomic.LoadInt64(&metrics.ActiveContexts) != 3 {
		t.Errorf("Expected ActiveContexts: 3, got: %d", metrics.ActiveContexts)
	}

	if atomic.LoadInt64(&metrics.ContextCreations) != 1 {
		t.Errorf("Expected ContextCreations: 1, got: %d", metrics.ContextCreations)
	}

	if atomic.LoadInt64(&metrics.ContextReuses) != 1 {
		t.Errorf("Expected ContextReuses: 1, got: %d", metrics.ContextReuses)
	}
}

func TestConcurrentRenders(t *testing.T) {
	metrics := NewJSMetrics()

	metrics.SetConcurrentRenders(3)
	metrics.SetConcurrentRenders(5)
	metrics.SetConcurrentRenders(2)

	if atomic.LoadInt64(&metrics.ConcurrentRenders) != 2 {
		t.Errorf("Expected ConcurrentRenders: 2, got: %d", metrics.ConcurrentRenders)
	}

	if atomic.LoadInt64(&metrics.MaxConcurrent) != 5 {
		t.Errorf("Expected MaxConcurrent: 5, got: %d", metrics.MaxConcurrent)
	}
}

func TestMemoryMetrics(t *testing.T) {
	metrics := NewJSMetrics()

	metrics.SetMemoryUsage(1000)
	metrics.SetMemoryUsage(2000)
	metrics.SetMemoryUsage(1500)

	if atomic.LoadInt64(&metrics.MemoryUsage) != 1500 {
		t.Errorf("Expected MemoryUsage: 1500, got: %d", metrics.MemoryUsage)
	}

	if atomic.LoadInt64(&metrics.PeakMemory) != 2000 {
		t.Errorf("Expected PeakMemory: 2000, got: %d", metrics.PeakMemory)
	}
}

func TestGetSummary(t *testing.T) {
	metrics := NewJSMetrics()

	// テストデータを設定
	metrics.RecordRender(2*time.Second, nil)
	metrics.RecordRender(1*time.Second, errors.New("test error"))
	metrics.RecordCacheHit()
	metrics.RecordCacheMiss()
	metrics.SetBrowserPoolSize(3)

	summary := metrics.GetSummary()

	if summary["renders_total"] != int64(2) {
		t.Errorf("Expected renders_total: 2, got: %v", summary["renders_total"])
	}

	if summary["renders_failed"] != int64(1) {
		t.Errorf("Expected renders_failed: 1, got: %v", summary["renders_failed"])
	}

	errorRate := summary["error_rate_percent"].(float64)
	if errorRate != 50.0 {
		t.Errorf("Expected error_rate_percent: 50.0, got: %f", errorRate)
	}

	cacheHitRate := summary["cache_hit_rate"].(float64)
	if cacheHitRate != 50.0 {
		t.Errorf("Expected cache_hit_rate: 50.0, got: %f", cacheHitRate)
	}

	if summary["browser_pool_size"] != int64(3) {
		t.Errorf("Expected browser_pool_size: 3, got: %v", summary["browser_pool_size"])
	}
}

func TestGetPerformanceReport(t *testing.T) {
	metrics := NewJSMetrics()

	// パフォーマンステストデータを設定
	for i := 0; i < 10; i++ {
		metrics.RecordRender(1*time.Second, nil)
	}

	time.Sleep(100 * time.Millisecond) // 少し時間を進める

	report := metrics.GetPerformanceReport()

	throughput := report["throughput_renders_per_second"].(float64)
	if throughput <= 0 {
		t.Errorf("Expected positive throughput, got: %f", throughput)
	}

	avgDuration := report["average_duration_ms"].(int64)
	if avgDuration != 1000 {
		t.Errorf("Expected average_duration_ms: 1000, got: %d", avgDuration)
	}

	errorRate := report["error_rate_percent"].(float64)
	if errorRate != 0.0 {
		t.Errorf("Expected error_rate_percent: 0.0, got: %f", errorRate)
	}

	grade := report["performance_grade"].(string)
	if grade == "" {
		t.Error("Expected performance_grade to be set")
	}
}

func TestIsHealthy(t *testing.T) {
	metrics := NewJSMetrics()

	// 健全な状態（エラーなし、短い時間）
	metrics.RecordRender(1*time.Second, nil)

	if !metrics.IsHealthy() {
		t.Error("Expected metrics to be healthy")
	}

	// 不健全な状態（エラー率高）
	for i := 0; i < 10; i++ {
		metrics.RecordRender(1*time.Second, errors.New("test error"))
	}

	if metrics.IsHealthy() {
		t.Error("Expected metrics to be unhealthy due to high error rate")
	}
}

func TestReset(t *testing.T) {
	metrics := NewJSMetrics()

	// テストデータを設定
	metrics.RecordRender(1*time.Second, nil)
	metrics.RecordCacheHit()

	// リセット
	metrics.Reset()

	if atomic.LoadInt64(&metrics.RenderCount) != 0 {
		t.Errorf("Expected RenderCount after reset: 0, got: %d", metrics.RenderCount)
	}

	if atomic.LoadInt64(&metrics.CacheHits) != 0 {
		t.Errorf("Expected CacheHits after reset: 0, got: %d", metrics.CacheHits)
	}

	if atomic.LoadInt64(&metrics.BrowserPoolSize) != 0 {
		t.Errorf("Expected BrowserPoolSize after reset: 0, got: %d", metrics.BrowserPoolSize)
	}

	if metrics.StartTime.IsZero() {
		t.Error("Expected StartTime to be reset")
	}
}
