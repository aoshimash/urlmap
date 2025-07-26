package cache

import (
	"fmt"
	"testing"
	"time"
)

func TestNewRenderCache(t *testing.T) {
	cache := NewRenderCache(100, 1*time.Hour)

	if cache.maxSize != 100 {
		t.Errorf("Expected maxSize: 100, got: %d", cache.maxSize)
	}

	if cache.ttl != 1*time.Hour {
		t.Errorf("Expected ttl: 1h, got: %v", cache.ttl)
	}

	if cache.Size() != 0 {
		t.Errorf("Expected initial size: 0, got: %d", cache.Size())
	}
}

func TestRenderCache_SetAndGet(t *testing.T) {
	cache := NewRenderCache(10, 1*time.Hour)

	// 正常な保存と取得
	cache.Set("https://example.com", "<html>test</html>")

	html, exists := cache.Get("https://example.com")
	if !exists {
		t.Error("Expected cached result to exist")
	}

	if html != "<html>test</html>" {
		t.Errorf("Expected HTML: <html>test</html>, got: %s", html)
	}

	// 存在しないURLのテスト
	_, exists = cache.Get("https://nonexistent.com")
	if exists {
		t.Error("Expected non-existent URL to not be cached")
	}
}

func TestRenderCache_LRU(t *testing.T) {
	cache := NewRenderCache(3, 1*time.Hour)

	// 3つのエントリを追加
	cache.Set("url1", "html1")
	cache.Set("url2", "html2")
	cache.Set("url3", "html3")

	// 4つ目のエントリを追加（LRUエビクションが発生）
	cache.Set("url4", "html4")

	// 最初のエントリが削除されているはず
	_, exists := cache.Get("url1")
	if exists {
		t.Error("Expected url1 to be evicted by LRU")
	}

	// 他のエントリは存在するはず
	_, exists = cache.Get("url2")
	if !exists {
		t.Error("Expected url2 to still exist")
	}

	_, exists = cache.Get("url3")
	if !exists {
		t.Error("Expected url3 to still exist")
	}

	_, exists = cache.Get("url4")
	if !exists {
		t.Error("Expected url4 to still exist")
	}
}

func TestRenderCache_TTL(t *testing.T) {
	cache := NewRenderCache(10, 100*time.Millisecond)

	cache.Set("https://example.com", "<html>test</html>")

	// すぐに取得（成功するはず）
	html, exists := cache.Get("https://example.com")
	if !exists {
		t.Error("Expected cached result to exist immediately")
	}

	if html != "<html>test</html>" {
		t.Errorf("Expected HTML: <html>test</html>, got: %s", html)
	}

	// TTLを超えて待機
	time.Sleep(150 * time.Millisecond)

	// 期限切れで取得できないはず
	_, exists = cache.Get("https://example.com")
	if exists {
		t.Error("Expected cached result to be expired")
	}
}

func TestRenderCache_Stats(t *testing.T) {
	cache := NewRenderCache(10, 1*time.Hour)

	// ヒットとミスの統計をテスト
	cache.Set("url1", "html1")

	// ヒット
	cache.Get("url1")

	// ミス
	cache.Get("nonexistent")

	stats := cache.GetStats()

	if stats["hits"] != int64(1) {
		t.Errorf("Expected hits: 1, got: %v", stats["hits"])
	}

	if stats["misses"] != int64(1) {
		t.Errorf("Expected misses: 1, got: %v", stats["misses"])
	}

	hitRate := stats["hit_rate"].(float64)
	if hitRate != 50.0 {
		t.Errorf("Expected hit rate: 50.0, got: %f", hitRate)
	}
}

func TestRenderCache_Clear(t *testing.T) {
	cache := NewRenderCache(10, 1*time.Hour)

	cache.Set("url1", "html1")
	cache.Set("url2", "html2")

	if cache.Size() != 2 {
		t.Errorf("Expected size: 2, got: %d", cache.Size())
	}

	cache.Clear()

	if cache.Size() != 0 {
		t.Errorf("Expected size after clear: 0, got: %d", cache.Size())
	}

	_, exists := cache.Get("url1")
	if exists {
		t.Error("Expected url1 to not exist after clear")
	}
}

func TestRenderCache_Cleanup(t *testing.T) {
	cache := NewRenderCache(10, 100*time.Millisecond)

	cache.Set("url1", "html1")
	cache.Set("url2", "html2")

	// すぐにクリーンアップ（何も削除されないはず）
	removed := cache.Cleanup()
	if removed != 0 {
		t.Errorf("Expected 0 removed, got: %d", removed)
	}

	// TTLを超えて待機
	time.Sleep(150 * time.Millisecond)

	// クリーンアップで期限切れエントリが削除されるはず
	removed = cache.Cleanup()
	if removed != 2 {
		t.Errorf("Expected 2 removed, got: %d", removed)
	}

	if cache.Size() != 0 {
		t.Errorf("Expected size after cleanup: 0, got: %d", cache.Size())
	}
}

func TestRenderCache_ConcurrentAccess(t *testing.T) {
	cache := NewRenderCache(100, 1*time.Hour)

	// 並行アクセステスト
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			url := fmt.Sprintf("url%d", id)
			html := fmt.Sprintf("html%d", id)

			cache.Set(url, html)

			// 複数回アクセス
			for j := 0; j < 5; j++ {
				cache.Get(url)
			}

			done <- true
		}(i)
	}

	// 全てのゴルーチンの完了を待機
	for i := 0; i < 10; i++ {
		<-done
	}

	// キャッシュサイズの確認
	if cache.Size() != 10 {
		t.Errorf("Expected final size: 10, got: %d", cache.Size())
	}

	// 統計情報の確認
	stats := cache.GetStats()
	if stats["hits"] != int64(50) { // 10 * 5 = 50 hits
		t.Errorf("Expected hits: 50, got: %v", stats["hits"])
	}
}
