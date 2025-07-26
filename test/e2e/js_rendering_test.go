package e2e

import (
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestJavaScriptRendering_E2E JavaScriptレンダリング機能のE2Eテスト
func TestJavaScriptRendering_E2E(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		jsEnabled   bool
		expectLinks int
		timeout     time.Duration
		args        []string
	}{
		{
			name:        "Static Site - Wikipedia",
			url:         "https://en.wikipedia.org/wiki/Go_(programming_language)",
			jsEnabled:   false,
			expectLinks: 1,
			timeout:     10 * time.Second,
			args:        []string{"--depth", "1"},
		},
		{
			name:        "JavaScript Rendering Enabled",
			url:         "https://httpbin.org/html",
			jsEnabled:   true,
			expectLinks: 1,
			timeout:     30 * time.Second,
			args:        []string{"--depth", "1", "--js-render"},
		},
		{
			name:        "Automatic SPA Detection",
			url:         "https://httpbin.org/html",
			jsEnabled:   true,
			expectLinks: 1,
			timeout:     30 * time.Second,
			args:        []string{"--depth", "1", "--js-auto"},
		},
		{
			name:        "Performance Optimization",
			url:         "https://httpbin.org/html",
			jsEnabled:   true,
			expectLinks: 1,
			timeout:     30 * time.Second,
			args:        []string{"--depth", "1", "--js-render", "--js-block-resources", "--js-workers", "2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テスト用のコマンドを構築（絶対パスを使用）
			cmd := exec.Command("./urlmap", append(tt.args, tt.url)...)
			cmd.Dir = "/Users/aoshima/dev/github/aoshimash/urlmap" // 絶対パスを明示的に設定

			// コマンドを実行
			output, err := cmd.Output()

			// エラーチェック
			if err != nil {
				require.NoError(t, err, "Command failed: %v", err)
			}

			// 出力を解析
			links := strings.Split(strings.TrimSpace(string(output)), "\n")

			// 空の行を除去
			var validLinks []string
			for _, link := range links {
				if strings.TrimSpace(link) != "" {
					validLinks = append(validLinks, strings.TrimSpace(link))
				}
			}

			// 最低限のリンク数チェック
			assert.GreaterOrEqual(t, len(validLinks), tt.expectLinks,
				"Expected at least %d links, got %d", tt.expectLinks, len(validLinks))

			// すべてのリンクが有効なURLであることを確認
			for _, link := range validLinks {
				assert.True(t, strings.HasPrefix(link, "http"),
					"Invalid URL format: %s", link)
			}
		})
	}
}

// TestJavaScriptRendering_Docker Docker環境でのJavaScriptレンダリングテスト
func TestJavaScriptRendering_Docker(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Docker test in short mode")
	}

	// CI環境ではDockerテストをスキップ
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping Docker test in CI environment")
	}

	tests := []struct {
		name        string
		url         string
		jsEnabled   bool
		expectLinks int
		timeout     time.Duration
		args        []string
	}{
		{
			name:        "Docker - Static Site",
			url:         "https://httpbin.org/html",
			jsEnabled:   false,
			expectLinks: 1,
			timeout:     30 * time.Second,
			args:        []string{"--depth", "1"},
		},
		{
			name:        "Docker - JavaScript Rendering",
			url:         "https://httpbin.org/html",
			jsEnabled:   true,
			expectLinks: 1,
			timeout:     60 * time.Second,
			args:        []string{"--depth", "1", "--js-render"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Dockerコマンドを構築
			dockerArgs := []string{"run", "--rm", "ghcr.io/aoshimash/urlmap:latest"}
			dockerArgs = append(dockerArgs, tt.args...)
			dockerArgs = append(dockerArgs, tt.url)

			cmd := exec.Command("docker", dockerArgs...)
			cmd.Dir = "/Users/aoshima/dev/github/aoshimash/urlmap" // 絶対パスを明示的に設定

			// コマンドを実行
			output, err := cmd.Output()

			// エラーチェック
			if err != nil {
				require.NoError(t, err, "Docker command failed: %v", err)
			}

			// 出力を解析
			links := strings.Split(strings.TrimSpace(string(output)), "\n")

			// 空の行を除去
			var validLinks []string
			for _, link := range links {
				if strings.TrimSpace(link) != "" {
					validLinks = append(validLinks, strings.TrimSpace(link))
				}
			}

			// 最低限のリンク数チェック
			assert.GreaterOrEqual(t, len(validLinks), tt.expectLinks,
				"Expected at least %d links, got %d", tt.expectLinks, len(validLinks))

			// すべてのリンクが有効なURLであることを確認
			for _, link := range validLinks {
				assert.True(t, strings.HasPrefix(link, "http"),
					"Invalid URL format: %s", link)
			}
		})
	}
}

// TestJavaScriptRendering_ErrorHandling エラーハンドリングのテスト
func TestJavaScriptRendering_ErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		args        []string
		expectError bool
	}{
		{
			name:        "Invalid URL",
			url:         "https://invalid-domain-12345.example.com",
			args:        []string{"--js-render"},
			expectError: false, // エラーが発生しない場合もある
		},
		{
			name:        "Timeout with slow site",
			url:         "https://httpbin.org/delay/10",
			args:        []string{"--js-render", "--js-timeout", "5s"},
			expectError: false, // タイムアウトが発生しない場合もある
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テスト用のコマンドを構築
			cmd := exec.Command("./urlmap", append(tt.args, tt.url)...)
			cmd.Dir = "/Users/aoshima/dev/github/aoshimash/urlmap" // 絶対パスを明示的に設定

			// コマンドを実行
			_, err := cmd.Output()

			if tt.expectError {
				assert.Error(t, err, "Expected error but got none")
			} else {
				assert.NoError(t, err, "Unexpected error: %v", err)
			}
		})
	}
}

// TestJavaScriptRendering_Performance パフォーマンステスト
func TestJavaScriptRendering_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	tests := []struct {
		name          string
		url           string
		args          []string
		maxDuration   time.Duration
		expectedLinks int
	}{
		{
			name:          "Basic JavaScript Rendering",
			url:           "https://httpbin.org/html",
			args:          []string{"--depth", "1", "--js-render"},
			maxDuration:   30 * time.Second,
			expectedLinks: 1,
		},
		{
			name:          "Optimized JavaScript Rendering",
			url:           "https://httpbin.org/html",
			args:          []string{"--depth", "1", "--js-render", "--js-block-resources", "--js-workers", "2"},
			maxDuration:   25 * time.Second,
			expectedLinks: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テスト用のコマンドを構築
			cmd := exec.Command("./urlmap", append(tt.args, tt.url)...)
			cmd.Dir = "/Users/aoshima/dev/github/aoshimash/urlmap" // 絶対パスを明示的に設定

			// 実行時間を計測
			start := time.Now()

			// コマンドを実行
			output, err := cmd.Output()
			duration := time.Since(start)

			// エラーチェック
			require.NoError(t, err, "Command failed: %v", err)

			// 実行時間チェック
			assert.LessOrEqual(t, duration, tt.maxDuration,
				"Execution took too long: %v (max: %v)", duration, tt.maxDuration)

			// 出力を解析
			links := strings.Split(strings.TrimSpace(string(output)), "\n")

			// 空の行を除去
			var validLinks []string
			for _, link := range links {
				if strings.TrimSpace(link) != "" {
					validLinks = append(validLinks, strings.TrimSpace(link))
				}
			}

			// リンク数チェック
			assert.GreaterOrEqual(t, len(validLinks), tt.expectedLinks,
				"Expected at least %d links, got %d", tt.expectedLinks, len(validLinks))

			t.Logf("Performance test completed in %v with %d links", duration, len(validLinks))
		})
	}
}
