package performance

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aoshimash/urlmap/internal/cache"
	"github.com/aoshimash/urlmap/internal/client"
	"github.com/aoshimash/urlmap/internal/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// BenchmarkJavaScriptRendering JavaScriptレンダリングのベンチマークテスト
func BenchmarkJavaScriptRendering(b *testing.B) {
	// テスト用の設定
	config := &client.OptimizedJSConfig{
		JSConfig: &client.JSConfig{
			Enabled:     true,
			BrowserType: "chromium",
			Headless:    true,
			Timeout:     30 * time.Second,
			WaitFor:     "networkidle",
		},
		CacheSize:       100,
		CacheTTL:        1 * time.Hour,
		ConcurrentLimit: 3,
		EnableMetrics:   true,
	}

	// クライアントを作成
	client, err := client.NewOptimizedJSClient(config, nil)
	require.NoError(b, err)
	defer client.Close()

	// テストURL
	testURLs := []string{
		"https://httpbin.org/html",
		"https://httpbin.org/json",
		"https://httpbin.org/xml",
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		url := testURLs[i%len(testURLs)]

		start := time.Now()
		_, err := client.Get(context.Background(), url)
		duration := time.Since(start)

		require.NoError(b, err)

		// パフォーマンス要件の確認
		if duration > 10*time.Second {
			b.Errorf("Rendering took too long: %v for %s", duration, url)
		}
	}
}

// BenchmarkRenderCache レンダリングキャッシュのベンチマークテスト
func BenchmarkRenderCache(b *testing.B) {
	// キャッシュを作成
	renderCache := cache.NewRenderCache(1000, 1*time.Hour)

	// テストデータ
	testData := "test HTML content"
	testURL := "https://example.com"

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// キャッシュに設定
		renderCache.Set(testURL, testData)

		// キャッシュから取得
		_, found := renderCache.Get(testURL)
		assert.True(b, found)
	}
}

// BenchmarkJSMetrics メトリクスのベンチマークテスト
func BenchmarkJSMetrics(b *testing.B) {
	metrics := metrics.NewJSMetrics()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// レンダリング記録
		metrics.RecordRender(100*time.Millisecond, nil)

		// キャッシュヒット記録
		metrics.RecordCacheHit()

		// ブラウザプール統計更新
		metrics.SetBrowserPoolSize(3)
		metrics.SetActiveContexts(2)

		// 並行処理統計更新
		metrics.SetConcurrentRenders(5)

		// メモリ使用量更新
		metrics.SetMemoryUsage(1024 * 1024) // 1MB
	}
}

// BenchmarkBrowserPool ブラウザプールのベンチマークテスト
func BenchmarkBrowserPool(b *testing.B) {
	// ブラウザプールを作成
	pool, err := client.NewBrowserPool(&client.JSConfig{
		BrowserType: "chromium",
		Headless:    true,
	}, nil)
	require.NoError(b, err)
	defer pool.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// コンテキストを取得
		browserCtx, err := pool.AcquireContext()
		require.NoError(b, err)

		// 簡単なページをレンダリング
		_, err = pool.RenderPage(context.Background(), "https://httpbin.org/html")
		require.NoError(b, err)

		// コンテキストを返却
		browserCtx.ReleaseContext()
	}
}

// TestPerformanceRequirements パフォーマンス要件のテスト
func TestPerformanceRequirements(t *testing.T) {
	// テスト用の設定
	config := &client.OptimizedJSConfig{
		JSConfig: &client.JSConfig{
			Enabled:     true,
			BrowserType: "chromium",
			Headless:    true,
			Timeout:     30 * time.Second,
			WaitFor:     "networkidle",
		},
		CacheSize:       100,
		CacheTTL:        1 * time.Hour,
		ConcurrentLimit: 3,
		EnableMetrics:   true,
	}

	// クライアントを作成
	client, err := client.NewOptimizedJSClient(config, nil)
	require.NoError(t, err)
	defer client.Close()

	// メトリクスを取得
	metrics := client.GetStats()

	// パフォーマンス要件のテスト
	t.Run("InitializationTime", func(t *testing.T) {
		// 初期化時間は5秒以内
		if startTime, ok := metrics["start_time"].(time.Time); ok {
			assert.Less(t, time.Since(startTime), 5*time.Second)
		}
	})

	t.Run("MemoryUsage", func(t *testing.T) {
		// メモリ使用量は1GB以内
		if memUsage, ok := metrics["memory_usage"].(int64); ok {
			assert.Less(t, memUsage, int64(1024*1024*1024))
		}
	})

	t.Run("BrowserPoolSize", func(t *testing.T) {
		// ブラウザプールサイズは設定値以下
		if poolSize, ok := metrics["browser_pool_size"].(int); ok {
			assert.LessOrEqual(t, poolSize, 3)
		}
	})
}

// TestConcurrentPerformance 並行処理のパフォーマンステスト
func TestConcurrentPerformance(t *testing.T) {
	// テスト用の設定
	config := &client.OptimizedJSConfig{
		JSConfig: &client.JSConfig{
			Enabled:     true,
			BrowserType: "chromium",
			Headless:    true,
			Timeout:     30 * time.Second,
			WaitFor:     "networkidle",
		},
		CacheSize:       100,
		CacheTTL:        1 * time.Hour,
		ConcurrentLimit: 5,
		EnableMetrics:   true,
	}

	// クライアントを作成
	client, err := client.NewOptimizedJSClient(config, nil)
	require.NoError(t, err)
	defer client.Close()

	// 並行処理テスト
	t.Run("ConcurrentRendering", func(t *testing.T) {
		urls := []string{
			"https://httpbin.org/html",
			"https://httpbin.org/json",
			"https://httpbin.org/xml",
		}

		start := time.Now()
		results := make(chan error, len(urls))

		// 並行してレンダリング
		for _, url := range urls {
			go func(u string) {
				_, err := client.Get(context.Background(), u)
				results <- err
			}(url)
		}

		// 結果を収集
		for i := 0; i < len(urls); i++ {
			err := <-results
			require.NoError(t, err)
		}

		duration := time.Since(start)

		// 並行処理の効率性を確認
		// 3つのURLを並行処理で30秒以内に完了
		assert.Less(t, duration, 30*time.Second)
	})
}

// TestCachePerformance キャッシュのパフォーマンステスト
func TestCachePerformance(t *testing.T) {
	// キャッシュを作成
	renderCache := cache.NewRenderCache(1000, 1*time.Hour)

	// テストデータ
	testData := "test HTML content"
	testURL := "https://example.com"

	t.Run("CacheHitPerformance", func(t *testing.T) {
		// キャッシュに設定
		renderCache.Set(testURL, testData)

		// キャッシュヒットの性能をテスト
		start := time.Now()
		for i := 0; i < 1000; i++ {
			_, found := renderCache.Get(testURL)
			assert.True(t, found)
		}
		duration := time.Since(start)

		// 1000回のキャッシュヒットが1秒以内
		assert.Less(t, duration, 1*time.Second)
	})

	t.Run("CacheEvictionPerformance", func(t *testing.T) {
		// キャッシュの上限を超えるデータを設定
		for i := 0; i < 1100; i++ {
			renderCache.Set(fmt.Sprintf("https://example%d.com", i), "test data")
		}

		// キャッシュサイズが制限内であることを確認
		assert.LessOrEqual(t, renderCache.Size(), 1000)
	})
}

// TestMemoryPerformance メモリ使用量のテスト
func TestMemoryPerformance(t *testing.T) {
	// テスト用の設定
	config := &client.OptimizedJSConfig{
		JSConfig: &client.JSConfig{
			Enabled:     true,
			BrowserType: "chromium",
			Headless:    true,
			Timeout:     30 * time.Second,
			WaitFor:     "networkidle",
		},
		CacheSize:       100,
		CacheTTL:        1 * time.Hour,
		ConcurrentLimit: 3,
		EnableMetrics:   true,
	}

	// クライアントを作成
	client, err := client.NewOptimizedJSClient(config, nil)
	require.NoError(t, err)
	defer client.Close()

	// メモリ使用量を監視
	initialStats := client.GetStats()

	// 複数のレンダリングを実行
	for i := 0; i < 10; i++ {
		_, err := client.Get(context.Background(), "https://httpbin.org/html")
		require.NoError(t, err)
	}

	finalStats := client.GetStats()

	// メモリ使用量の増加が制限内であることを確認
	if initialMem, ok1 := initialStats["memory_usage"].(int64); ok1 {
		if finalMem, ok2 := finalStats["memory_usage"].(int64); ok2 {
			memoryIncrease := finalMem - initialMem
			assert.Less(t, memoryIncrease, int64(100*1024*1024)) // 100MB以内
		}
	}
}
