package detector

import (
	"log/slog"
	"testing"
	"time"
)

func TestSPADetector_DetectSPA(t *testing.T) {
	logger := slog.Default()
	detector := NewSPADetector(logger)

	tests := []struct {
		name        string
		url         string
		htmlContent string
		expectedSPA bool
		description string
	}{
		{
			name:        "React SPA",
			url:         "https://react-app.com",
			htmlContent: `<div id="root"><div data-reactroot>Hello React</div></div>`,
			expectedSPA: true,
			description: "Reactアプリケーションの検出",
		},
		{
			name:        "Vue SPA",
			url:         "https://vue-app.com",
			htmlContent: `<div id="app"><div v-if="show">Hello Vue</div></div>`,
			expectedSPA: true,
			description: "Vueアプリケーションの検出",
		},
		{
			name:        "Angular SPA",
			url:         "https://angular-app.com",
			htmlContent: `<div ng-app="myApp"><div ng-controller="myCtrl">Hello Angular</div></div>`,
			expectedSPA: true,
			description: "Angularアプリケーションの検出",
		},
		{
			name:        "Next.js SPA",
			url:         "https://nextjs-app.com",
			htmlContent: `<div id="__next"><script>window.__NEXT_DATA__={}</script></div>`,
			expectedSPA: true,
			description: "Next.jsアプリケーションの検出",
		},
		{
			name:        "Static HTML",
			url:         "https://static-site.com",
			htmlContent: `<html><body><h1>Hello World</h1><a href="/about">About</a><a href="/contact">Contact</a></body></html>`,
			expectedSPA: false,
			description: "静的HTMLサイトの検出",
		},
		{
			name:        "Empty body SPA",
			url:         "https://spa-empty.com",
			htmlContent: `<html><body><div id="app"></div></body></html>`,
			expectedSPA: true,
			description: "空のbodyを持つSPAの検出",
		},
		{
			name:        "Low link count",
			url:         "https://low-links.com",
			htmlContent: `<html><body><h1>Welcome</h1><p>This is a static site</p><a href="/1">Link 1</a><a href="/2">Link 2</a></body></html>`,
			expectedSPA: false, // リンク数が少なくても、フレームワークの指標がない場合はSPAではない
			description: "リンク数が少ないサイトの検出",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := detector.DetectSPA(tt.url, tt.htmlContent)
			if err != nil {
				t.Fatalf("DetectSPA failed: %v", err)
			}

			if result.IsSPA != tt.expectedSPA {
				t.Errorf("Expected SPA: %v, got: %v (confidence: %.2f, indicators: %v)",
					tt.expectedSPA, result.IsSPA, result.Confidence, result.Indicators)
			}

			// 信頼度の範囲チェック
			if result.Confidence < 0.0 || result.Confidence > 1.0 {
				t.Errorf("Confidence out of range [0,1]: %.2f", result.Confidence)
			}

			// メソッドが設定されているかチェック
			if result.Method == "" {
				t.Error("Method should not be empty")
			}
		})
	}
}

func TestSPADetector_detectFramework(t *testing.T) {
	logger := slog.Default()
	detector := NewSPADetector(logger)

	tests := []struct {
		name        string
		htmlContent string
		expected    bool
		framework   string
	}{
		{
			name:        "React detection",
			htmlContent: `<div data-reactroot>React App</div>`,
			expected:    true,
			framework:   "React",
		},
		{
			name:        "Vue detection",
			htmlContent: `<div v-if="show">Vue App</div>`,
			expected:    true,
			framework:   "Vue",
		},
		{
			name:        "Angular detection",
			htmlContent: `<div ng-app="app">Angular App</div>`,
			expected:    true,
			framework:   "Angular",
		},
		{
			name:        "Next.js detection",
			htmlContent: `<script>window.__NEXT_DATA__={}</script>`,
			expected:    true,
			framework:   "Next.js",
		},
		{
			name:        "No framework",
			htmlContent: `<div>Plain HTML</div>`,
			expected:    false,
			framework:   "None",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.detectFramework(tt.htmlContent)
			if result != tt.expected {
				t.Errorf("Expected %s detection: %v, got: %v", tt.framework, tt.expected, result)
			}
		})
	}
}

func TestSPADetector_detectSPAStructure(t *testing.T) {
	logger := slog.Default()
	detector := NewSPADetector(logger)

	tests := []struct {
		name        string
		htmlContent string
		expected    bool
		description string
	}{
		{
			name:        "SPA structure with root div",
			htmlContent: `<div id="root"></div>`,
			expected:    true,
			description: "root divを持つSPA構造",
		},
		{
			name:        "SPA structure with app div",
			htmlContent: `<div id="app"></div>`,
			expected:    true,
			description: "app divを持つSPA構造",
		},
		{
			name:        "SPA structure with __next div",
			htmlContent: `<div id="__next"></div>`,
			expected:    true,
			description: "__next divを持つSPA構造",
		},
		{
			name:        "Empty body",
			htmlContent: `<html><body><div></div></body></html>`,
			expected:    true,
			description: "空のコンテナを持つ構造",
		},
		{
			name:        "Regular HTML",
			htmlContent: `<html><body><h1>Title</h1><p>Content</p></body></html>`,
			expected:    false,
			description: "通常のHTML構造",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.detectSPAStructure(tt.htmlContent)
			if result != tt.expected {
				t.Errorf("Expected SPA structure: %v, got: %v", tt.expected, result)
			}
		})
	}
}

func TestSPADetector_detectLowLinkCount(t *testing.T) {
	logger := slog.Default()
	detector := NewSPADetector(logger)

	tests := []struct {
		name        string
		htmlContent string
		expected    bool
		linkCount   int
	}{
		{
			name:        "Low link count",
			htmlContent: `<a href="/1">Link 1</a><a href="/2">Link 2</a>`,
			expected:    true,
			linkCount:   2,
		},
		{
			name:        "High link count",
			htmlContent: `<nav><a href="/1">Link 1</a><a href="/2">Link 2</a><a href="/3">Link 3</a><a href="/4">Link 4</a></nav><a href="/5">Link 5</a><a href="/6">Link 6</a><a href="/7">Link 7</a><a href="/8">Link 8</a><a href="/9">Link 9</a><a href="/10">Link 10</a><a href="/11">Link 11</a>`,
			expected:    false,
			linkCount:   11,
		},
		{
			name:        "No links",
			htmlContent: `<div>No links here</div>`,
			expected:    true,
			linkCount:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.detectLowLinkCount(tt.htmlContent)
			if result != tt.expected {
				t.Errorf("Expected low link count: %v, got: %v (link count: %d)", tt.expected, result, tt.linkCount)
			}
		})
	}
}

func TestSPADetector_detectDynamicContent(t *testing.T) {
	logger := slog.Default()
	detector := NewSPADetector(logger)

	tests := []struct {
		name        string
		htmlContent string
		expected    bool
		description string
	}{
		{
			name:        "JavaScript content",
			htmlContent: `<script>window.addEventListener('load', function() {})</script>`,
			expected:    true,
			description: "JavaScriptコードを含む",
		},
		{
			name:        "Template variable",
			htmlContent: `<div>{{message}}</div>`,
			expected:    true,
			description: "テンプレート変数を含む",
		},
		{
			name:        "No dynamic content",
			htmlContent: `<div>Static content</div>`,
			expected:    false,
			description: "動的コンテンツなし",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.detectDynamicContent(tt.htmlContent)
			if result != tt.expected {
				t.Errorf("Expected dynamic content: %v, got: %v", tt.expected, result)
			}
		})
	}
}

func TestSPADetector_extractLinks(t *testing.T) {
	logger := slog.Default()
	detector := NewSPADetector(logger)

	tests := []struct {
		name        string
		htmlContent string
		expected    []string
	}{
		{
			name:        "Multiple links",
			htmlContent: `<a href="/about">About</a><a href="/contact">Contact</a><a href="/help">Help</a>`,
			expected:    []string{"/about", "/contact", "/help"},
		},
		{
			name:        "No links",
			htmlContent: `<div>No links here</div>`,
			expected:    []string{},
		},
		{
			name:        "Links with text",
			htmlContent: `<a href="/page1">Page 1</a><a href="/page2">Page 2</a>`,
			expected:    []string{"/page1", "/page2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.extractLinks(tt.htmlContent)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d links, got %d", len(tt.expected), len(result))
			}

			for i, expected := range tt.expected {
				if i >= len(result) || result[i] != expected {
					t.Errorf("Expected link %d: %s, got: %s", i, expected, result[i])
				}
			}
		})
	}
}

func TestDetectionCache(t *testing.T) {
	cache := NewDetectionCache(1 * time.Hour)

	// テスト用の結果を作成
	result := &DetectionResult{
		IsSPA:      true,
		Confidence: 0.8,
		Indicators: []string{"framework_detected"},
		Method:     "static_analysis",
		Timestamp:  time.Now(),
	}

	// キャッシュに保存
	cache.Set("example.com", result)

	// キャッシュから取得
	cached, exists := cache.Get("example.com")
	if !exists {
		t.Error("Expected cached result to exist")
	}

	if cached.IsSPA != result.IsSPA {
		t.Errorf("Expected IsSPA: %v, got: %v", result.IsSPA, cached.IsSPA)
	}

	if cached.Confidence != result.Confidence {
		t.Errorf("Expected Confidence: %.2f, got: %.2f", result.Confidence, cached.Confidence)
	}

	// 存在しないドメインのテスト
	_, exists = cache.Get("nonexistent.com")
	if exists {
		t.Error("Expected non-existent domain to not be cached")
	}

	// キャッシュサイズのテスト
	size := cache.Size()
	if size != 1 {
		t.Errorf("Expected cache size: 1, got: %d", size)
	}
}
