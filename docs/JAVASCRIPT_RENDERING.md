# JavaScript レンダリング機能ガイド

## 概要

urlmapは静的HTMLに加えて、JavaScript実行後の完全なHTMLからもリンクを発見できます。この機能により、SPA（Single Page Application）や動的コンテンツを含むWebサイトからも効果的にリンクを収集できます。

## 対応サイト例

- **React**: Create React App、Next.js、Gatsby
- **Vue**: Vue CLI、Nuxt.js
- **Angular**: Angular CLI
- **その他**: Svelte、Ember.js、動的CMS

## 基本的な使用方法

### 手動指定

```bash
# JavaScriptレンダリングを有効化
urlmap https://spa-site.com --js-render

# 詳細設定
urlmap https://site.com \
  --js-render \
  --js-timeout 10s \
  --js-wait networkidle \
  --js-browser chromium
```

### 自動検出

```bash
# SPA自動検出
urlmap https://site.com --js-auto

# 厳格モード（動的検証も実行）
urlmap https://site.com --js-auto-strict
```

### Docker使用

```bash
# 基本的な使用
docker run --rm ghcr.io/aoshimash/urlmap:latest \
  https://spa-site.com --js-render

# 出力ファイル保存
docker run --rm -v $(pwd)/output:/app/output \
  ghcr.io/aoshimash/urlmap:latest \
  https://site.com --js-render --output /app/output/sitemap.txt
```

## オプション一覧

| オプション | 説明 | デフォルト |
|------------|------|------------|
| `--js-render` | JavaScriptレンダリング有効化 | false |
| `--js-auto` | 自動SPA検出 | false |
| `--js-auto-strict` | 厳格な自動SPA検出 | false |
| `--js-browser` | ブラウザタイプ (chromium/firefox/webkit) | chromium |
| `--js-timeout` | ページ読み込みタイムアウト | 30s |
| `--js-wait` | 待機条件 (networkidle/domcontentloaded) | networkidle |
| `--js-fallback` | HTTPフォールバック有効化 | false |
| `--js-threshold` | 自動検出の閾値 | 0.3 |

### パフォーマンス最適化オプション

| オプション | 説明 | デフォルト |
|------------|------|------------|
| `--js-pool-size` | ブラウザプール最大サイズ | 3 |
| `--js-workers` | 並列ワーカー数 | 5 |
| `--js-cache-size` | キャッシュ最大エントリ数 | 1000 |
| `--js-cache-ttl` | キャッシュTTL | 1h |
| `--js-block-resources` | 画像・CSS等をブロック | false |
| `--js-metrics` | パフォーマンスメトリクス収集 | false |

## パフォーマンス指標

- **静的サイト**: ~100ms/page（変更なし）
- **JSサイト**: 3-5秒/page（初回ブラウザ起動込み）
- **キャッシュヒット時**: <100ms/page
- **並列処理**: 5-10 pages/秒

## トラブルシューティング

### よくある問題

#### 1. ブラウザ起動失敗
```
Error: Failed to launch browser
```
**解決策**:
- Dockerコンテナ使用を推奨
- システムの依存関係を確認

#### 2. タイムアウトエラー
```
Error: Page load timeout
```
**解決策**:
```bash
# タイムアウトを延長
urlmap https://slow-site.com --js-render --js-timeout 60s
```

#### 3. メモリ不足
```
Error: Out of memory
```
**解決策**:
```bash
# ブラウザプールサイズを削減
urlmap https://site.com --js-render --js-pool-size 1
```

## ベストプラクティス

### 1. サイトタイプ別の推奨設定

#### SPAサイト（React/Vue/Angular）
```bash
urlmap https://spa-site.com \
  --js-auto \
  --js-timeout 15s \
  --js-wait networkidle
```

#### 重いサイト
```bash
urlmap https://heavy-site.com \
  --js-render \
  --js-timeout 30s \
  --js-block-resources \
  --js-pool-size 2
```

#### 大規模クロール
```bash
urlmap https://large-site.com \
  --js-auto \
  --js-workers 10 \
  --js-cache-size 5000
```

### 2. パフォーマンス最適化

#### キャッシュ活用
```bash
# 長期間のキャッシュ
urlmap https://site.com --js-render --js-cache-ttl 1h
```

#### リソースブロック
```bash
# 不要なリソースをブロックして高速化
urlmap https://site.com --js-render --js-block-resources
```

#### 並列処理の調整
```bash
# 高並列処理（メモリ使用量に注意）
urlmap https://site.com --js-render --js-workers 10 --js-pool-size 5
```

## 使用例

### 基本的なSPAクロール
```bash
# Reactサイトのクロール
urlmap https://react-app.com --js-render --depth 3

# 出力例
https://react-app.com/
https://react-app.com/about
https://react-app.com/contact
https://react-app.com/products
https://react-app.com/products/item1
```

### 自動検出でのクロール
```bash
# 自動でSPAかどうかを判定
urlmap https://unknown-site.com --js-auto --depth 2
```

### パフォーマンス重視のクロール
```bash
# 高速化オプションを有効化
urlmap https://large-spa.com \
  --js-render \
  --js-block-resources \
  --js-workers 8 \
  --js-cache-size 2000 \
  --js-metrics
```

## 制限事項

1. **メモリ使用量**: JavaScriptレンダリングは大量のメモリを使用します
2. **処理速度**: 静的サイトと比べて大幅に遅くなります
3. **ブラウザ依存**: Playwrightのブラウザインストールが必要です
4. **ネットワーク**: 外部リソースの読み込みに依存します

## サポート

問題が発生した場合は、以下の情報を含めて報告してください：

- 使用したコマンド
- エラーメッセージ
- 対象URL
- システム情報（OS、メモリ等）
- Docker使用の有無
