package client

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"sync"
	"time"

	"github.com/aoshimash/urlmap/internal/cache"
	"github.com/aoshimash/urlmap/internal/metrics"
	"github.com/playwright-community/playwright-go"
)

// OptimizedJSClient 最適化されたJavaScriptクライアント
type OptimizedJSClient struct {
	browserPool *BrowserPool
	renderCache *cache.RenderCache
	metrics     *metrics.JSMetrics
	config      *JSConfig
	logger      *slog.Logger

	// 並列処理制御
	concurrentLimit int
	semaphore       chan struct{}

	// リソース管理
	contextPool  map[string]*BrowserContext
	contextMutex sync.RWMutex

	// ライフサイクル管理
	initialized bool
	closed      bool
	closeMu     sync.Mutex
}

// OptimizedJSConfig 最適化されたJSクライアントの設定
type OptimizedJSConfig struct {
	*JSConfig

	// キャッシュ設定
	CacheSize int
	CacheTTL  time.Duration

	// 並列処理設定
	ConcurrentLimit int

	// メトリクス設定
	EnableMetrics bool
}

// DefaultOptimizedJSConfig デフォルトの最適化設定
func DefaultOptimizedJSConfig() *OptimizedJSConfig {
	return &OptimizedJSConfig{
		JSConfig:        DefaultJSConfig(),
		CacheSize:       1000,
		CacheTTL:        1 * time.Hour,
		ConcurrentLimit: 5,
		EnableMetrics:   true,
	}
}

// NewOptimizedJSClient 新しい最適化されたJSクライアントを作成
func NewOptimizedJSClient(config *OptimizedJSConfig, logger *slog.Logger) (*OptimizedJSClient, error) {
	if config == nil {
		config = DefaultOptimizedJSConfig()
	}

	if logger == nil {
		logger = slog.Default()
	}

	// ブラウザプールを作成
	browserPool, err := NewBrowserPool(config.JSConfig, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create browser pool: %w", err)
	}

	// レンダリングキャッシュを作成
	renderCache := cache.NewRenderCache(config.CacheSize, config.CacheTTL)

	// メトリクスを作成
	var jsMetrics *metrics.JSMetrics
	if config.EnableMetrics {
		jsMetrics = metrics.NewJSMetrics()
	}

	client := &OptimizedJSClient{
		browserPool:     browserPool,
		renderCache:     renderCache,
		metrics:         jsMetrics,
		config:          config.JSConfig,
		logger:          logger,
		concurrentLimit: config.ConcurrentLimit,
		semaphore:       make(chan struct{}, config.ConcurrentLimit),
		contextPool:     make(map[string]*BrowserContext),
	}

	client.initialized = true
	return client, nil
}

// RenderPage 最適化されたページレンダリング
func (c *OptimizedJSClient) RenderPage(ctx context.Context, targetURL string) (string, error) {
	if !c.initialized {
		return "", fmt.Errorf("client not initialized")
	}

	startTime := time.Now()

	// キャッシュから取得を試行
	if cached, exists := c.renderCache.Get(targetURL); exists {
		c.logger.Debug("Cache hit", "url", targetURL)
		if c.metrics != nil {
			c.metrics.RecordCacheHit()
			c.metrics.RecordRender(time.Since(startTime), nil)
		}
		return cached, nil
	}

	if c.metrics != nil {
		c.metrics.RecordCacheMiss()
	}

	// セマフォで並列数を制限
	select {
	case c.semaphore <- struct{}{}:
		defer func() { <-c.semaphore }()
	case <-ctx.Done():
		return "", ctx.Err()
	}

	// 実際のレンダリング実行
	content, err := c.performRender(ctx, targetURL)

	// メトリクス記録
	if c.metrics != nil {
		c.metrics.RecordRender(time.Since(startTime), err)
	}

	// 成功した場合のみキャッシュに保存
	if err == nil {
		c.renderCache.Set(targetURL, content)
	}

	return content, err
}

// performRender 実際のレンダリング処理
func (c *OptimizedJSClient) performRender(ctx context.Context, targetURL string) (string, error) {
	// ブラウザコンテキストを取得
	browserCtx, err := c.getOrCreateContext(targetURL)
	if err != nil {
		return "", fmt.Errorf("failed to get browser context: %w", err)
	}

	// ページを作成
	page, err := browserCtx.Context.NewPage()
	if err != nil {
		return "", fmt.Errorf("failed to create page: %w", err)
	}
	defer page.Close()

	// ページ最適化
	if err := c.optimizePage(page); err != nil {
		return "", fmt.Errorf("failed to optimize page: %w", err)
	}

	// ページ読み込み
	var waitUntil *playwright.WaitUntilState
	switch c.config.WaitFor {
	case "networkidle":
		waitUntil = playwright.WaitUntilStateNetworkidle
	case "domcontentloaded":
		waitUntil = playwright.WaitUntilStateDomcontentloaded
	case "load":
		waitUntil = playwright.WaitUntilStateLoad
	default:
		waitUntil = playwright.WaitUntilStateNetworkidle
	}

	_, err = page.Goto(targetURL, playwright.PageGotoOptions{
		WaitUntil: waitUntil,
		Timeout:   playwright.Float(float64(c.config.Timeout.Milliseconds())),
	})
	if err != nil {
		return "", fmt.Errorf("failed to navigate to URL %s: %w", targetURL, err)
	}

	// HTML取得
	content, err := page.Content()
	if err != nil {
		return "", fmt.Errorf("failed to get page content: %w", err)
	}

	return content, nil
}

// getOrCreateContext ドメインベースのコンテキスト取得
func (c *OptimizedJSClient) getOrCreateContext(url string) (*BrowserContext, error) {
	domain := c.extractDomain(url)

	c.contextMutex.RLock()
	if ctx, exists := c.contextPool[domain]; exists {
		c.contextMutex.RUnlock()
		if c.metrics != nil {
			c.metrics.RecordContextReuse()
		}
		return ctx, nil
	}
	c.contextMutex.RUnlock()

	// 新しいコンテキストを作成
	browserCtx, err := c.browserPool.AcquireContext()
	if err != nil {
		return nil, fmt.Errorf("failed to acquire browser context: %w", err)
	}

	c.contextMutex.Lock()
	c.contextPool[domain] = browserCtx
	c.contextMutex.Unlock()

	if c.metrics != nil {
		c.metrics.RecordContextCreation()
	}

	return browserCtx, nil
}

// optimizePage ページ最適化
func (c *OptimizedJSClient) optimizePage(page playwright.Page) error {
	// 不要なリソースをブロック（オプション）
	if c.config.BlockResources {
		if err := page.Route("**/*", func(route playwright.Route) {
			resourceType := route.Request().ResourceType()

			// 画像、フォント、スタイルシートをブロック
			switch resourceType {
			case "image", "font", "stylesheet", "media":
				route.Abort()
				return
			}

			route.Continue()
		}); err != nil {
			return err
		}
	}

	// ユーザーエージェント設定
	if c.config.UserAgent != "" {
		if err := page.SetExtraHTTPHeaders(map[string]string{
			"User-Agent": c.config.UserAgent,
		}); err != nil {
			return err
		}
	}

	return nil
}

// extractDomain URLからドメインを抽出
func (c *OptimizedJSClient) extractDomain(url string) string {
	// 簡易的なドメイン抽出（実際の実装ではnet/urlを使用）
	if len(url) > 8 && url[:8] == "https://" {
		url = url[8:]
	} else if len(url) > 7 && url[:7] == "http://" {
		url = url[7:]
	}

	for i, char := range url {
		if char == '/' {
			return url[:i]
		}
	}
	return url
}

// Get 統一インターフェース用のGetメソッド
func (c *OptimizedJSClient) Get(ctx context.Context, targetURL string) (*JSResponse, error) {
	content, err := c.RenderPage(ctx, targetURL)
	if err != nil {
		return nil, err
	}

	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	return &JSResponse{
		URL:     targetURL,
		Content: content,
		Status:  200,
		Headers: make(map[string]string),
		Host:    parsedURL.Host,
	}, nil
}

// GetStats 統計情報を取得
func (c *OptimizedJSClient) GetStats() map[string]interface{} {
	stats := make(map[string]interface{})

	// ブラウザプール統計
	if poolStats := c.browserPool.GetPoolStats(); poolStats != nil {
		stats["browser_pool"] = poolStats
	}

	// キャッシュ統計
	if cacheStats := c.renderCache.GetStats(); cacheStats != nil {
		stats["render_cache"] = cacheStats
	}

	// メトリクス統計
	if c.metrics != nil {
		stats["metrics"] = c.metrics.GetSummary()
		stats["performance"] = c.metrics.GetPerformanceReport()
	}

	return stats
}

// Close リソースをクリーンアップ
func (c *OptimizedJSClient) Close() error {
	c.closeMu.Lock()
	defer c.closeMu.Unlock()

	if c.closed {
		return nil
	}

	var errors []error

	// ブラウザプールを閉じる
	if c.browserPool != nil {
		if err := c.browserPool.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close browser pool: %w", err))
		}
	}

	// キャッシュをクリア
	if c.renderCache != nil {
		c.renderCache.Clear()
	}

	c.closed = true
	c.logger.Debug("Optimized JS client closed")

	if len(errors) > 0 {
		return fmt.Errorf("errors closing optimized JS client: %v", errors)
	}

	return nil
}
