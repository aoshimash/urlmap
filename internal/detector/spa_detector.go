package detector

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/aoshimash/urlmap/internal/client"
)

// SPADetector SPA検出機能を提供する構造体
type SPADetector struct {
	logger *slog.Logger
	cache  *DetectionCache
}

// DetectionResult SPA検出の結果を表す構造体
type DetectionResult struct {
	IsSPA      bool      `json:"is_spa"`
	Confidence float64   `json:"confidence"`
	Indicators []string  `json:"indicators"`
	Method     string    `json:"method"`
	Timestamp  time.Time `json:"timestamp"`
}

// NewSPADetector 新しいSPA検出器を作成
func NewSPADetector(logger *slog.Logger) *SPADetector {
	return &SPADetector{
		logger: logger,
		cache:  NewDetectionCache(1 * time.Hour), // 1時間のTTL
	}
}

// DetectSPA URLとHTMLコンテンツからSPAかどうかを検出
func (d *SPADetector) DetectSPA(url, htmlContent string) (*DetectionResult, error) {
	// キャッシュから結果を取得
	if cached, exists := d.cache.Get(d.extractDomain(url)); exists {
		d.logger.Debug("SPA detection result found in cache", "url", url)
		return cached, nil
	}

	result := &DetectionResult{
		IsSPA:      false,
		Confidence: 0.0,
		Indicators: []string{},
		Method:     "static_analysis",
		Timestamp:  time.Now(),
	}

	// 1. Framework検出
	if d.detectFramework(htmlContent) {
		result.IsSPA = true
		result.Confidence += 0.4
		result.Indicators = append(result.Indicators, "framework_detected")
	}

	// 2. DOM構造分析
	if d.detectSPAStructure(htmlContent) {
		result.IsSPA = true
		result.Confidence += 0.3
		result.Indicators = append(result.Indicators, "spa_structure")
	}

	// 3. リンク数分析（フレームワークが検出された場合のみ）
	if d.detectFramework(htmlContent) && d.detectLowLinkCount(htmlContent) {
		result.Confidence += 0.2
		result.Indicators = append(result.Indicators, "low_link_count")
	}

	// 4. 動的コンテンツ検出（フレームワークが検出された場合のみ）
	if d.detectFramework(htmlContent) && d.detectDynamicContent(htmlContent) {
		result.Confidence += 0.1
		result.Indicators = append(result.Indicators, "dynamic_content")
	}

	// 閾値判定（フレームワーク検出またはSPA構造 + 高い信頼度）
	result.IsSPA = result.Confidence >= 0.5 || (d.detectSPAStructure(htmlContent) && result.Confidence >= 0.3)

	// 結果をキャッシュに保存
	d.cache.Set(d.extractDomain(url), result)

	return result, nil
}

// detectFramework 主要なフレームワークの検出
func (d *SPADetector) detectFramework(htmlContent string) bool {
	// React検出
	reactIndicators := []string{
		"__REACT_DEVTOOLS_GLOBAL_HOOK__",
		"data-reactroot",
		"_reactInternalInstance",
		`<div id="root"></div>`,
		`<div id="app"></div>`,
		"react",
		"ReactDOM",
		"createElement",
	}

	// Vue検出
	vueIndicators := []string{
		"Vue.js",
		"__VUE__",
		"v-if", "v-for", "v-model",
		`<div id="app"></div>`,
		"vue",
		"Vue.component",
	}

	// Angular検出
	angularIndicators := []string{
		"ng-app", "ng-controller",
		"[ng-", "(ng-",
		"__ng_",
		"angular.module",
		"angular",
		"ng-",
	}

	// Next.js検出
	nextjsIndicators := []string{
		"__NEXT_DATA__",
		"_next/static",
		`<div id="__next"></div>`,
		"next",
		"Next.js",
	}

	// Svelte検出
	svelteIndicators := []string{
		"svelte",
		"__svelte__",
		"data-svelte",
	}

	allIndicators := append(reactIndicators, vueIndicators...)
	allIndicators = append(allIndicators, angularIndicators...)
	allIndicators = append(allIndicators, nextjsIndicators...)
	allIndicators = append(allIndicators, svelteIndicators...)

	for _, indicator := range allIndicators {
		if strings.Contains(strings.ToLower(htmlContent), strings.ToLower(indicator)) {
			d.logger.Debug("Framework detected", "indicator", indicator)
			return true
		}
	}

	return false
}

// detectSPAStructure SPA特有のDOM構造を検出
func (d *SPADetector) detectSPAStructure(htmlContent string) bool {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		d.logger.Warn("Failed to parse HTML for structure analysis", "error", err)
		return false
	}

	// SPA特有の構造パターン
	spaPatterns := []string{
		"#root", "#app", "#__next", "#main",
		"[data-reactroot]", "[data-vue]", "[data-svelte]",
	}

	for _, pattern := range spaPatterns {
		if doc.Find(pattern).Length() > 0 {
			d.logger.Debug("SPA structure detected", "pattern", pattern)
			return true
		}
	}

	// 空のbodyや最小限のHTML構造（ただし、十分なコンテンツがある場合は除外）
	body := doc.Find("body")
	if body.Length() > 0 {
		bodyHTML := body.Text()
		bodyText := strings.TrimSpace(bodyHTML)

		// 空に近いbodyで、かつSPA特有の構造がない場合はSPAの可能性が高い
		if len(bodyText) < 50 && doc.Find("h1, h2, h3, p").Length() == 0 {
			return true
		}
	}

	return false
}

// detectLowLinkCount リンク数が少ない場合の検出
func (d *SPADetector) detectLowLinkCount(htmlContent string) bool {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return false
	}

	linkCount := doc.Find("a[href]").Length()

	// リンク数が10個未満の場合はSPAの可能性
	if linkCount < 10 {
		d.logger.Debug("Low link count detected", "count", linkCount)
		return true
	}

	return false
}

// detectDynamicContent 動的コンテンツの検出
func (d *SPADetector) detectDynamicContent(htmlContent string) bool {
	// JavaScriptの存在確認
	jsPatterns := []string{
		"<script",
		"window.", "document.",
		"addEventListener",
		"fetch(", "XMLHttpRequest",
	}

	for _, pattern := range jsPatterns {
		if strings.Contains(htmlContent, pattern) {
			d.logger.Debug("Dynamic content detected", "pattern", pattern)
			return true
		}
	}

	return false
}

// VerifyWithJS 実際にJSレンダリングして検証
func (d *SPADetector) VerifyWithJS(ctx context.Context, url string, staticHTML string, jsClient *client.JSClient) (*DetectionResult, error) {
	d.logger.Info("Starting dynamic verification with JavaScript", "url", url)

	// JavaScript実行後のHTML取得
	jsHTML, err := jsClient.RenderPage(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to render with JS: %w", err)
	}

	// 静的HTMLとJS実行後HTMLの差分分析
	staticLinks := d.extractLinks(staticHTML)
	jsLinks := d.extractLinks(jsHTML)

	linkDifference := len(jsLinks) - len(staticLinks)

	if len(staticLinks) == 0 {
		// 静的リンクが0個の場合は、JSリンクが1個以上あればSPA
		result := &DetectionResult{
			IsSPA:      len(jsLinks) > 0,
			Confidence: math.Min(float64(len(jsLinks)), 1.0),
			Indicators: []string{fmt.Sprintf("js_links_%d", len(jsLinks))},
			Method:     "dynamic_verification",
			Timestamp:  time.Now(),
		}
		return result, nil
	}

	improvementRatio := float64(linkDifference) / float64(len(staticLinks))

	result := &DetectionResult{
		IsSPA:      improvementRatio > 0.5, // 50%以上リンクが増加
		Confidence: math.Min(improvementRatio, 1.0),
		Indicators: []string{fmt.Sprintf("link_improvement_%.1f%%", improvementRatio*100)},
		Method:     "dynamic_verification",
		Timestamp:  time.Now(),
	}

	d.logger.Info("Dynamic verification completed",
		"url", url,
		"static_links", len(staticLinks),
		"js_links", len(jsLinks),
		"improvement_ratio", improvementRatio,
		"is_spa", result.IsSPA,
	)

	return result, nil
}

// extractLinks HTMLからリンクを抽出
func (d *SPADetector) extractLinks(htmlContent string) []string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		d.logger.Warn("Failed to parse HTML for link extraction", "error", err)
		return []string{}
	}

	var links []string
	doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		if href, exists := s.Attr("href"); exists {
			links = append(links, href)
		}
	})

	return links
}

// extractDomain URLからドメインを抽出
func (d *SPADetector) extractDomain(url string) string {
	// 簡単なドメイン抽出（実際の実装ではより堅牢にする）
	if strings.HasPrefix(url, "http") {
		parts := strings.Split(url, "/")
		if len(parts) >= 3 {
			return parts[2]
		}
	}
	return url
}
