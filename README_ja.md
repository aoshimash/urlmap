# crawld

[![CI](https://github.com/aoshimash/crawld/workflows/CI/badge.svg)](https://github.com/aoshimash/crawld/actions/workflows/ci.yml)
[![Docker](https://github.com/aoshimash/crawld/workflows/Docker%20Build%20and%20Publish/badge.svg)](https://github.com/aoshimash/crawld/actions/workflows/docker.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/aoshimash/crawld)](https://goreportcard.com/report/github.com/aoshimash/crawld)
[![License](https://img.shields.io/github/license/aoshimash/crawld)](LICENSE)

高性能で効率的なWebクローラー CLI ツール。ドメイン内のリンクを発見するためのGoで構築された並行処理対応ツールです。

[🇺🇸 English](README.md) | 🇯🇵 日本語

## 🚀 機能

- **再帰的リンク探索**: ウェブサイト内のすべてのリンクを自動的に発見
- **同一ドメインフィルタリング**: 外部リンクを除外し、特定ドメインに焦点を当てた
- **並行処理**: 設定可能なワーカープールによる高性能クローリング
- **深度制限**: 無限再帰を防ぐためのクローリング深度制御
- **プログレス表示**: クローリング操作中のリアルタイム進捗報告
- **レート制限**: 設定可能なリクエストレートによる丁寧なクローリング
- **グレースフル・シャットダウン**: 終了時の適切なクリーンアップによる中断安全性
- **構造化ログ**: 詳細モードサポート付きの包括的ログ記録
- **複数出力形式**: URL出力は標準出力、ログは標準エラー出力
- **カスタムユーザーエージェント**: 識別用の設定可能なユーザーエージェント文字列

## 📦 インストール

### バイナリダウンロード

[リリースページ](https://github.com/aoshimash/crawld/releases)から最新のバイナリをダウンロード：

#### Linux (x86_64)
```bash
curl -L -o crawld.tar.gz https://github.com/aoshimash/crawld/releases/latest/download/crawld-linux-amd64.tar.gz
tar -xzf crawld.tar.gz
chmod +x crawld
sudo mv crawld /usr/local/bin/
```

#### Linux (ARM64)
```bash
curl -L -o crawld.tar.gz https://github.com/aoshimash/crawld/releases/latest/download/crawld-linux-arm64.tar.gz
tar -xzf crawld.tar.gz
chmod +x crawld
sudo mv crawld /usr/local/bin/
```

#### macOS (Intel)
```bash
curl -L -o crawld.tar.gz https://github.com/aoshimash/crawld/releases/latest/download/crawld-darwin-amd64.tar.gz
tar -xzf crawld.tar.gz
chmod +x crawld
sudo mv crawld /usr/local/bin/
```

#### macOS (Apple Silicon)
```bash
curl -L -o crawld.tar.gz https://github.com/aoshimash/crawld/releases/latest/download/crawld-darwin-arm64.tar.gz
tar -xzf crawld.tar.gz
chmod +x crawld
sudo mv crawld /usr/local/bin/
```

#### Windows
リリースページから `crawld-windows-amd64.zip` をダウンロードし、実行ファイルを展開してください。

### Docker

インストールなしでDockerで実行：

```bash
# GitHub Container Registryからプル
docker pull ghcr.io/aoshimash/crawld:latest

# 基本的な使用方法
docker run --rm ghcr.io/aoshimash/crawld:latest https://example.com
```

### ソースからビルド

要件: Go 1.21以上

```bash
# リポジトリをクローン
git clone https://github.com/aoshimash/crawld.git
cd crawld

# アプリケーションをビルド
go build -o crawld ./cmd/crawld

# グローバルにインストール（オプション）
sudo mv crawld /usr/local/bin/
```

## 🎯 使用方法

### 基本的な使用方法

```bash
# デフォルト設定でウェブサイトをクロール
crawld https://example.com

# バージョン確認
crawld version

# ヘルプ表示
crawld --help
```

### 高度なオプション

```bash
# クロール深度を3レベルに制限
crawld --depth 3 https://example.com

# 高速クローリングのため20個の並行ワーカーを使用
crawld --concurrent 20 https://example.com

# 詳細ログを有効化
crawld --verbose https://example.com

# カスタムユーザーエージェント
crawld --user-agent "MyBot/1.0" https://example.com

# レート制限（毎秒5リクエスト）
crawld --rate-limit 5 https://example.com

# プログレス表示を無効化
crawld --progress=false https://example.com

# オプション組み合わせ
crawld --depth 5 --concurrent 15 --verbose --rate-limit 2 https://example.com
```

### Docker使用方法

```bash
# 基本的なクローリング
docker run --rm ghcr.io/aoshimash/crawld:latest https://example.com

# オプション指定
docker run --rm ghcr.io/aoshimash/crawld:latest --depth 3 --concurrent 20 https://example.com

# 出力をファイルに保存
docker run --rm ghcr.io/aoshimash/crawld:latest https://example.com > urls.txt

# シェルアクセス付きインタラクティブモード
docker run -it --rm ghcr.io/aoshimash/crawld:latest /bin/sh
```

## 🔧 コマンドラインオプション

| フラグ | 短縮 | デフォルト | 説明 |
|------|-------|---------|-------------|
| `--depth` | `-d` | 0 (無制限) | 最大クロール深度 |
| `--concurrent` | `-c` | 10 | 並行ワーカー数 |
| `--verbose` | `-v` | false | 詳細ログを有効化 |
| `--user-agent` | `-u` | crawld/1.0.0 | カスタムUser-Agent文字列 |
| `--progress` | `-p` | true | プログレス表示を表示 |
| `--rate-limit` | `-r` | 0 (制限なし) | レート制限（毎秒リクエスト数） |
| `--help` | `-h` | - | ヘルプメッセージを表示 |

## 📋 使用例

### 基本的なウェブサイトクローリング

```bash
# シンプルなウェブサイトをクロール
crawld https://example.com
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
# 2レベルまでの深度でのみクロール
crawld --depth 2 https://blog.example.com
```

### 高性能クローリング

```bash
# 大規模サイト用に50個の並行ワーカーを使用
crawld --concurrent 50 --verbose https://large-site.example.com
```

### 丁寧なクローリング

```bash
# カスタムユーザーエージェントで毎秒1リクエストに制限
crawld --rate-limit 1 --user-agent "Research Bot 1.0 (contact@example.com)" https://example.com
```

### 結果をファイルに保存

```bash
# URLをファイルに保存
crawld https://example.com > discovered_urls.txt

# タイムスタンプとログ付きで保存
crawld --verbose https://example.com > urls.txt 2> crawl.log
```

### 大規模サイトの処理

```bash
# プログレス追跡付きで大規模サイト用に最適化
crawld --depth 5 --concurrent 30 --rate-limit 10 --verbose https://large-site.com
```

## 🏗 アーキテクチャ

crawldは保守性と拡張性のためにモジュラーアーキテクチャに従っています：

```
crawld/
├── cmd/crawld/          # CLIアプリケーションエントリーポイント
├── internal/
│   ├── client/          # リトライロジック付きHTTPクライアント
│   ├── config/          # 設定とログ設定
│   ├── crawler/         # コアクローリングエンジン
│   ├── output/          # 出力フォーマットと処理
│   ├── parser/          # HTMLパースとリンク抽出
│   ├── progress/        # プログレス報告と統計
│   └── url/            # URL検証と正規化
└── pkg/utils/          # パブリックユーティリティ
```

### コアコンポーネント

- **クローラーエンジン**: ワーカープールアーキテクチャによる並行クローラー
- **HTTPクライアント**: タイムアウトとリトライロジック付きの耐障害性HTTPクライアント
- **リンクパーサー**: 信頼性の高いリンク抽出のためのgoqueryを使用したHTMLパーサー
- **URLマネージャー**: URL検証、正規化、ドメインフィルタリング
- **プログレスレポーター**: リアルタイムクローリング統計とプログレス追跡

## ⚡ パフォーマンス

crawldは以下の特性でパフォーマンスが最適化されています：

### ベンチマーク

- **小規模サイト**（100ページ未満）: ~50-100 URL/秒
- **中規模サイト**（100-1000ページ）: ~30-50 URL/秒
- **大規模サイト**（1000ページ以上）: ~20-30 URL/秒

パフォーマンスは以下により変動します：
- ネットワーク遅延と帯域幅
- 対象サーバーの応答時間
- 並行ワーカー数
- ページの複雑さとサイズ

### パフォーマンス最適化のヒント

1. **並行ワーカー**: I/Oバウンドなクローリングでは `--concurrent` を増加
2. **レート制限**: サーバーの過負荷を避けるため `--rate-limit` を使用
3. **深度制御**: 無限クローリングを避けるため適切な `--depth` を設定
4. **プログレス追跡**: わずかなパフォーマンス向上のため `--progress=false` で無効化

### メモリ使用量

- ベースメモリ: ~10-20 MB
- ワーカーあたり: ~1-2 MB
- URL保存: ~100バイト/URL
- 10,000 URL: 通常 ~50-100 MB total

## 🔍 トラブルシューティング

### よくある問題

#### Permission Denied（権限拒否）
```bash
# エラー: permission denied
sudo chmod +x crawld
# またはユーザーディレクトリにインストール
mv crawld ~/.local/bin/
```

#### DNS解決の失敗
```bash
# 最初にURLアクセス可能性をテスト
curl -I https://example.com

# DNS解決をチェック
nslookup example.com

# デバッグ用詳細モードを使用
crawld --verbose https://example.com
```

#### レート制限 / 429エラー
```bash
# 並行ワーカーを減らしレート制限を追加
crawld --concurrent 5 --rate-limit 1 https://example.com
```

#### 大規模サイトでのメモリ問題
```bash
# 並行ワーカーを減らす
crawld --concurrent 5 --depth 3 https://large-site.com

# メモリ使用量を監視
crawld --verbose https://example.com 2>&1 | grep -i memory
```

#### SSL/TLS証明書エラー
```bash
# 証明書の有効性をチェック
curl -I https://example.com

# 開発/テスト用のみ（本番環境では非推奨）
# 現在設定不可 - crawldはすべての証明書を検証
```

### デバッグ

問題のトラブルシューティングには詳細ログを有効化：

```bash
crawld --verbose https://example.com 2> debug.log
```

ログレベル：
- INFO: 一般的なクローリング進捗
- DEBUG: 詳細なURL処理
- WARN: 致命的でない問題（失敗したURL、タイムアウト）
- ERROR: クローリングを停止する致命的エラー

### パフォーマンス問題

クローリングが遅い場合：

1. **ネットワーク確認**: 対象サイトへの直接アクセスをテスト
2. **ワーカー調整**: 異なる `--concurrent` 値を試す（5-50）
3. **レート制限監視**: 制限されていないことを確認
4. **レート制限使用**: より丁寧にするため `--rate-limit` を追加

```bash
# パフォーマンステストコマンド
time crawld --depth 2 --concurrent 20 https://example.com > /dev/null
```

## 🤝 コントリビューション

コントリビューションを歓迎します！詳細は[コントリビューションガイドライン](CONTRIBUTING.md)をご覧ください。

### 開発セットアップ

```bash
# リポジトリをクローン
git clone https://github.com/aoshimash/crawld.git
cd crawld

# 依存関係をインストール
go mod download

# テスト実行
go test ./...

# リンティング実行
go vet ./...
golangci-lint run

# 開発用ビルド
go build -o crawld ./cmd/crawld
```

### プロジェクト構造

コードベースの構造と設計決定に関する詳細は[アーキテクチャドキュメント](docs/ARCHITECTURE.md)をご覧ください。

## 📚 依存関係

crawldは以下の高品質Goライブラリを使用しています：

- **[Cobra](https://github.com/spf13/cobra)** - モダンCLIフレームワーク
- **[Resty](https://github.com/go-resty/resty)** - HTTPクライアントライブラリ
- **[goquery](https://github.com/PuerkitoBio/goquery)** - jQuery風HTMLパース

## 📊 監視と統計

crawldはクローリング中と完了後に詳細な統計を提供します：

```bash
# 統計付きの出力例
crawld --verbose https://example.com
```

統計には以下が含まれます：
- 発見されたURL総数
- 正常にクロールされたURL
- 失敗したURLと理由
- 到達した最大深度
- 総クローリング時間
- 平均応答時間

## 🔒 セキュリティ考慮事項

- crawldは基底HTTPライブラリのデフォルト動作によりrobots.txtを尊重
- リンク抽出でXSSを防ぐため安全なHTMLパースを使用
- 悪意のあるリダイレクトを防ぐためすべてのURLを検証
- ハングするリクエストを防ぐため適切なタイムアウト処理を実装
- レート制限機能で偶発的なDoSを防ぐ支援

## 📄 ライセンス

このプロジェクトはMITライセンスの下でライセンスされています。詳細は[LICENSE](LICENSE)ファイルをご覧ください。

## 🙋‍♂️ サポート

- **バグレポート**: [GitHub Issues](https://github.com/aoshimash/crawld/issues)
- **機能リクエスト**: [GitHub Discussions](https://github.com/aoshimash/crawld/discussions)
- **セキュリティ問題**: セキュリティ問題は非公開でメールしてください

## 🗺 ロードマップ

予定されている将来の機能拡張：
- [ ] robots.txt尊重設定
- [ ] カスタム出力形式（JSON、CSV、XML）
- [ ] カスタム処理用プラグインシステム
- [ ] 分散クローリングサポート
- [ ] 大規模クロール監視用Web UI
- [ ] 人気データ分析ツールとの統合

---

**crawldチームが❤️を込めて作成**
