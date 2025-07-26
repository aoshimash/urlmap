package detector

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// SPADetector SPA検出器
type SPADetector struct {
	logger *slog.Logger
	cache  *DetectionCache
}

// DetectionResult SPA検出結果
type DetectionResult struct {
	IsSPA      bool      `json:"is_spa"`
	Confidence float64   `json:"confidence"`
	Indicators []string  `json:"indicators"`
	Method     string    `json:"method"`
	Timestamp  time.Time `json:"timestamp"`
}

// NewSPADetector 新しいSPA検出器を作成
func NewSPADetector(logger *slog.Logger) *SPADetector {
	if logger == nil {
		logger = slog.Default()
	}
	return &SPADetector{
		logger: logger,
		cache:  NewDetectionCache(1 * time.Hour), // 1時間のTTL
	}
}

// DetectSPA 静的解析によるSPA検出
func (d *SPADetector) DetectSPA(url, htmlContent string) (*DetectionResult, error) {
	// キャッシュから結果を取得
	domain := d.extractDomain(url)
	if cached, exists := d.cache.Get(domain); exists {
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
		result.Confidence += 0.4
		result.Indicators = append(result.Indicators, "framework_detected")
	}

	// 2. DOM構造分析
	if d.detectSPAStructure(htmlContent) {
		result.Confidence += 0.3
		result.Indicators = append(result.Indicators, "spa_structure")
	}

	// 3. リンク数分析
	if d.detectLowLinkCount(htmlContent) {
		result.Confidence += 0.2
		result.Indicators = append(result.Indicators, "low_link_count")
	}

	// 4. 動的コンテンツ検出
	if d.detectDynamicContent(htmlContent) {
		result.Confidence += 0.1
		result.Indicators = append(result.Indicators, "dynamic_content")
	}

	// 閾値判定
	result.IsSPA = result.Confidence >= 0.5

	d.logger.Debug("SPA detection completed",
		"url", url,
		"is_spa", result.IsSPA,
		"confidence", result.Confidence,
		"indicators", result.Indicators,
	)

	return result, nil
}

// detectFramework Framework検出
func (d *SPADetector) detectFramework(htmlContent string) bool {
	// React検出
	reactIndicators := []string{
		"__REACT_DEVTOOLS_GLOBAL_HOOK__",
		"data-reactroot",
		"_reactInternalInstance",
		"<div id=\"root\"></div>",
		"<div id=\"app\"></div>",
		"react",
		"ReactDOM",
		"createElement",
	}

	// Vue検出
	vueIndicators := []string{
		"Vue.js",
		"__VUE__",
		"v-if", "v-for", "v-model",
		"<div id=\"app\"></div>",
		"vue",
		"Vue.component",
	}

	// Angular検出
	angularIndicators := []string{
		"ng-app", "ng-controller",
		"[ng-", "(ng-",
		"__ng__",
		"angular.module",
		"angular",
		"ng-",
	}

	// Next.js検出
	nextjsIndicators := []string{
		"__NEXT_DATA__",
		"_next/static",
		"<div id=\"__next\"></div>",
		"next",
		"Next.js",
	}

	// Svelte検出
	svelteIndicators := []string{
		"svelte",
		"__svelte__",
		"<div id=\"svelte\"></div>",
	}

	// すべてのインジケーターを結合
	allIndicators := append(reactIndicators, vueIndicators...)
	allIndicators = append(allIndicators, angularIndicators...)
	allIndicators = append(allIndicators, nextjsIndicators...)
	allIndicators = append(allIndicators, svelteIndicators...)

	// 大文字小文字を区別しない検索
	lowerHTML := strings.ToLower(htmlContent)
	for _, indicator := range allIndicators {
		if strings.Contains(lowerHTML, strings.ToLower(indicator)) {
			return true
		}
	}

	return false
}

// detectSPAStructure SPA構造の検出
func (d *SPADetector) detectSPAStructure(htmlContent string) bool {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return false
	}

	// 1. 単一のルート要素の存在
	rootSelectors := []string{"#root", "#app", "#__next", "#svelte", "[data-reactroot]"}
	for _, selector := range rootSelectors {
		if doc.Find(selector).Length() > 0 {
			return true
		}
	}

	// 2. 空のコンテナ要素の存在
	emptyContainers := doc.Find("div, main, section").FilterFunction(func(i int, s *goquery.Selection) bool {
		return s.Children().Length() == 0 && strings.TrimSpace(s.Text()) == ""
	})
	if emptyContainers.Length() > 0 {
		return true
	}

	// 3. スクリプトタグの比率
	scriptCount := doc.Find("script").Length()
	totalElements := doc.Find("*").Length()
	if totalElements > 0 && float64(scriptCount)/float64(totalElements) > 0.1 {
		return true
	}

	return false
}

// detectLowLinkCount リンク数の分析
func (d *SPADetector) detectLowLinkCount(htmlContent string) bool {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return false
	}

	// リンク数をカウント
	linkCount := doc.Find("a[href]").Length()
	navCount := doc.Find("nav a[href]").Length()

	// ナビゲーションリンクが少ない場合（SPAの特徴）
	if navCount <= 3 {
		return true
	}

	// 全体的にリンクが少ない場合
	if linkCount <= 10 {
		return true
	}

	return false
}

// detectDynamicContent 動的コンテンツの検出
func (d *SPADetector) detectDynamicContent(htmlContent string) bool {
	// 1. テンプレート変数の検出
	templatePatterns := []string{
		`\{\{.*?\}\}`, // Vue/Handlebars
		`\{.*?\}`,     // React/JSX
		`\[\[.*?\]\]`, // Angular
		`\$\{.*?\}`,   // Template literals
	}

	for _, pattern := range templatePatterns {
		re := regexp.MustCompile(pattern)
		if re.MatchString(htmlContent) {
			return true
		}
	}

	// 2. 動的属性の検出
	dynamicAttrs := []string{
		"v-if", "v-for", "v-model", "v-show",
		"ng-if", "ng-repeat", "ng-model",
		"data-bind", "data-react",
	}

	for _, attr := range dynamicAttrs {
		if strings.Contains(htmlContent, attr) {
			return true
		}
	}

	return false
}

// VerifyWithJS 動的検証（JavaScript実行前後の比較）
func (d *SPADetector) VerifyWithJS(url, staticHTML string, jsClient interface{}) (*DetectionResult, error) {
	// JSClientの型チェック
	client, ok := jsClient.(interface {
		Get(ctx context.Context, url string) (interface{}, error)
	})
	if !ok {
		return nil, fmt.Errorf("invalid JS client type")
	}

	// JavaScript実行後のHTML取得
	ctx := context.Background()
	jsResponse, err := client.Get(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to render with JS: %w", err)
	}

	// レスポンスからHTMLを抽出
	var jsHTML string
	if response, ok := jsResponse.(interface{ GetBody() string }); ok {
		jsHTML = response.GetBody()
	} else {
		return nil, fmt.Errorf("failed to extract HTML from JS response")
	}

	// 静的HTMLとJS実行後HTMLの差分分析
	staticLinks := d.extractLinks(staticHTML)
	jsLinks := d.extractLinks(jsHTML)

	linkDifference := len(jsLinks) - len(staticLinks)
	improvementRatio := 0.0
	if len(staticLinks) > 0 {
		improvementRatio = float64(linkDifference) / float64(len(staticLinks))
	}

	result := &DetectionResult{
		IsSPA:      improvementRatio > 0.5, // 50%以上リンクが増加
		Confidence: math.Min(improvementRatio, 1.0),
		Indicators: []string{fmt.Sprintf("link_improvement_%.1f%%", improvementRatio*100)},
		Method:     "dynamic_verification",
		Timestamp:  time.Now(),
	}

	d.logger.Debug("Dynamic verification completed",
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

// extractDomainはURLからドメイン部分を抽出します
func (d *SPADetector) extractDomain(rawurl string) string {
	u, err := url.Parse(rawurl)
	if err != nil {
		d.logger.Debug("Failed to parse URL for domain extraction", "url", rawurl, "error", err)
		return rawurl
	}
	return u.Host
}

// urlParseはnet/url.Parseのエイリアス
var urlParse = func(rawurl string) (*url.URL, error) {
	return url.Parse(rawurl)
}

// GetConfidenceLevel 信頼度レベルの取得
func (d *SPADetector) GetConfidenceLevel(confidence float64) string {
	switch {
	case confidence >= 0.8:
		return "high"
	case confidence >= 0.6:
		return "medium"
	case confidence >= 0.4:
		return "low"
	default:
		return "very_low"
	}
}
