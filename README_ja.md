# urlmap

[![CI](https://github.com/aoshimash/urlmap/workflows/CI/badge.svg)](https://github.com/aoshimash/urlmap/actions/workflows/ci.yml)
[![Docker](https://github.com/aoshimash/urlmap/workflows/Docker%20Build%20and%20Publish/badge.svg)](https://github.com/aoshimash/urlmap/actions/workflows/docker.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/aoshimash/urlmap)](https://goreportcard.com/report/github.com/aoshimash/urlmap)
[![License](https://img.shields.io/github/license/aoshimash/urlmap)](LICENSE)

高速で効率的なウェブクローラーCLIツールです。ドメイン内のURLを発見・マッピングします。Goで構築されており、高性能な並行クローリングが可能です。

## 🚀 機能

- **再帰的リンク発見**: ウェブサイト内のすべてのリンクを自動発見
- **同一ドメインフィルタリング**: 外部リンクを避け、特定ドメインに集中
- **並行処理**: 設定可能なワーカープールによる高性能クローリング
- **深度制限**: 無限再帰を防ぐためのクロール深度制御
- **プログレス表示**: クローリング操作中のリアルタイム進捗レポート
- **レート制限**: 設定可能なリクエストレートによる配慮されたクローリング
- **グレースフル終了**: 中断時の適切なクリーンアップを伴う中断セーフ
- **構造化ログ**: 詳細モードサポートによる包括的ログ
- **複数出力形式**: URLを標準出力、ログを標準エラー出力
- **カスタムユーザーエージェント**: 識別のための設定可能なユーザーエージェント文字列

## 📦 インストール

### バイナリダウンロード

[リリースページ](https://github.com/aoshimash/urlmap/releases)から最新のバイナリをダウンロードしてください：

#### Linux (x86_64)
```bash
curl -L -o urlmap.tar.gz https://github.com/aoshimash/urlmap/releases/latest/download/urlmap-linux-amd64.tar.gz
tar -xzf urlmap.tar.gz
chmod +x urlmap
sudo mv urlmap /usr/local/bin/
```

#### Linux (ARM64)
```bash
curl -L -o urlmap.tar.gz https://github.com/aoshimash/urlmap/releases/latest/download/urlmap-linux-arm64.tar.gz
tar -xzf urlmap.tar.gz
chmod +x urlmap
sudo mv urlmap /usr/local/bin/
```

#### macOS (Intel)
```bash
curl -L -o urlmap.tar.gz https://github.com/aoshimash/urlmap/releases/latest/download/urlmap-darwin-amd64.tar.gz
tar -xzf urlmap.tar.gz
chmod +x urlmap
sudo mv urlmap /usr/local/bin/
```

#### macOS (Apple Silicon)
```bash
curl -L -o urlmap.tar.gz https://github.com/aoshimash/urlmap/releases/latest/download/urlmap-darwin-arm64.tar.gz
tar -xzf urlmap.tar.gz
chmod +x urlmap
sudo mv urlmap /usr/local/bin/
```

#### Windows
リリースページから`urlmap-windows-amd64.zip`をダウンロードし、実行ファイルを展開してください。

### Docker

インストールせずにDockerで実行：

```bash
# GitHub Container Registryからプル
docker pull ghcr.io/aoshimash/urlmap:latest

# 基本的な使用方法
docker run --rm ghcr.io/aoshimash/urlmap:latest https://example.com
```

### ソースからビルド

要件: Go 1.21以上

```bash
# リポジトリをクローン
git clone https://github.com/aoshimash/urlmap.git
cd urlmap

# アプリケーションをビルド
go build -o urlmap ./cmd/urlmap

# グローバルインストール（オプション）
sudo mv urlmap /usr/local/bin/
```

## 🎯 使用方法

### 基本的な使用方法

```bash
# デフォルト設定でウェブサイトをクロール
urlmap https://example.com

# バージョン確認
urlmap version

# ヘルプを表示
urlmap --help
```

### 高度なオプション

```bash
# クロール深度を3レベルに制限
urlmap --depth 3 https://example.com

# より高速なクローリングのために20の並行ワーカーを使用
urlmap --concurrent 20 https://example.com

# 詳細ログを有効化
urlmap --verbose https://example.com

# カスタムユーザーエージェント
urlmap --user-agent "MyBot/1.0" https://example.com

# レート制限（秒あたり5リクエスト）
urlmap --rate-limit 5 https://example.com

# プログレス表示を無効化
urlmap --progress=false https://example.com

# オプションの組み合わせ
urlmap --depth 5 --concurrent 15 --verbose --rate-limit 2 https://example.com
```

### Docker使用法

```bash
# 基本的なクローリング
docker run --rm ghcr.io/aoshimash/urlmap:latest https://example.com

# オプション付き
docker run --rm ghcr.io/aoshimash/urlmap:latest --depth 3 --concurrent 20 https://example.com

# 出力をファイルに保存
docker run --rm ghcr.io/aoshimash/urlmap:latest https://example.com > urls.txt

# シェルアクセス付きのインタラクティブモード
docker run -it --rm ghcr.io/aoshimash/urlmap:latest /bin/sh
```

## 🔧 コマンドラインオプション

| フラグ | 短縮形 | デフォルト | 説明 |
|------|-------|---------|-------------|
| `--depth` | `-d` | -1 (無制限) | 最大クロール深度 |
| `--concurrent` | `-c` | 10 | 並行ワーカー数 |
| `--verbose` | `-v` | false | 詳細ログを有効化 |
| `--user-agent` | `-u` | urlmap/1.0.0 | カスタムUser-Agent文字列 |
| `--progress` | `-p` | true | プログレス表示 |
| `--rate-limit` | `-r` | 0 (制限なし) | レート制限（秒あたりリクエスト数） |
| `--help` | `-h` | - | ヘルプメッセージを表示 |

## 📋 例

### 基本的なウェブサイトクローリング

```bash
# シンプルなウェブサイトをクロール
urlmap https://example.com
```

出力:
```
https://example.com
https://example.com/about
https://example.com/contact
https://example.com/products
```

### 深度制限クローリング

```bash
# 2レベルまでのみクロール
urlmap --depth 2 https://blog.example.com
```

### 高性能クローリング

```bash
# 大規模サイト用に50の並行ワーカーを使用
urlmap --concurrent 50 --verbose https://large-site.example.com
```

### 配慮されたクローリング

```bash
# カスタムユーザーエージェントで秒あたり1リクエストに制限
urlmap --rate-limit 1 --user-agent "Research Bot 1.0 (contact@example.com)" https://example.com
```

### 結果をファイルに保存

```bash
# URLをファイルに保存
urlmap https://example.com > discovered_urls.txt

# タイムスタンプとログ付きで保存
urlmap --verbose https://example.com > urls.txt 2> crawl.log
```

## 🏗 アーキテクチャ

urlmapは保守性と拡張性のためのモジュラーアーキテクチャに従っています：

```
urlmap/
├── cmd/urlmap/          # CLIアプリケーションエントリーポイント
├── internal/
│   ├── client/          # リトライロジック付きHTTPクライアント
│   ├── config/          # 設定管理とログ設定
│   ├── crawler/         # コアクローリングロジック
│   ├── output/          # 出力フォーマッティング
│   ├── parser/          # HTMLパースとリンク抽出
│   ├── progress/        # プログレス追跡とレポート
│   └── url/            # URL正規化とバリデーション
├── pkg/
│   └── utils/          # ユーティリティ関数
└── test/               # テストとフィクスチャ
```

## 📚 依存関係

urlmapは以下の高品質なGoライブラリを使用しています：

- **[Cobra](https://github.com/spf13/cobra)** - モダンなCLIフレームワーク
- **[Resty](https://github.com/go-resty/resty)** - リトライとタイムアウト機能付きHTTPクライアント
- **[goquery](https://github.com/PuerkitoBio/goquery)** - jQuery風HTMLパーサー
- **標準ライブラリ** - 並行性、ログ、HTTP処理

## 🔒 セキュリティ上の考慮事項

- urlmapは基盤となるHTTPライブラリのデフォルト動作でrobots.txtを尊重します
- リンク抽出でXSSを防ぐための安全なHTMLパースを使用
- 悪意のあるリダイレクトを防ぐためのすべてのURLバリデーション
- デフォルトでHTTPS証明書検証を強制

## 🙋‍♀️ サポート

- **バグレポート**: [GitHub Issues](https://github.com/aoshimash/urlmap/issues)
- **機能リクエスト**: [GitHub Discussions](https://github.com/aoshimash/urlmap/discussions)
- **セキュリティ問題**: セキュリティ問題はプライベートでメールしてください

## 🤖 AI駆動開発

このプロジェクトはAI駆動ソフトウェア開発の実用的な実験として機能しています。この探究の一環として、コードベース全体がCursor AI agentを使用して実装されました：

- プロジェクト設計とアーキテクチャ
- Issue作成とプロジェクト管理
- プルリクエスト作成とコードレビュー
- すべての機能と機能性の実装
- ドキュメントとREADMEの作成

**重要な注意**: このリポジトリには人間が書いたコードは一行もありません。すべてがAIツールによって生成・管理されており、AI支援開発の現在の能力を実証しています。

## 📄 ライセンス

このプロジェクトはMITライセンスの下でライセンスされています - 詳細は[LICENSE](LICENSE)ファイルを参照してください。

---
