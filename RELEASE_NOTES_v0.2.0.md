# urlmap v0.2.0 - コードベース統一とドキュメント改善

🚀 **urlmap v0.2.0のリリースをお知らせします！**

この維持版では、コードベースの完全な統一とドキュメントの改善が行われました。

## ✨ このリリースの主な改善点

### 🔧 コードベース改善
- **完全な名前統一**: すべてのcrawldへの参照をurlmapに統一し、一貫性を向上
- **コードクリーンアップ**: 内部コンポーネント間の命名の一貫性を改善

### 📚 ドキュメント改善
- **AI駆動開発セクション追加**: READMEにAI駆動開発に関する新しいセクションを追加
- **開発者体験向上**: より詳細で理解しやすいドキュメント

## 🆙 v0.1.0からの変更点

### Changed
- refactor: すべての残りのcrawld参照をurlmapに置換
- docs: READMEにAI駆動開発セクションを追加

### Fixed
- 内部コンポーネント間の命名不整合を解決
- コードベース全体での一貫性を改善

## 📦 インストール

### バイナリダウンロード

#### Linux (x86_64)
```bash
curl -L -o urlmap.tar.gz https://github.com/aoshimash/urlmap/releases/download/v0.2.0/urlmap-linux-amd64.tar.gz
tar -xzf urlmap.tar.gz
chmod +x urlmap
sudo mv urlmap /usr/local/bin/
```

#### macOS (Apple Silicon)
```bash
curl -L -o urlmap.tar.gz https://github.com/aoshimash/urlmap/releases/download/v0.2.0/urlmap-darwin-arm64.tar.gz
tar -xzf urlmap.tar.gz
chmod +x urlmap
sudo mv urlmap /usr/local/bin/
```

#### Windows
[リリースページ](https://github.com/aoshimash/urlmap/releases/tag/v0.2.0)から`urlmap-windows-amd64.zip`をダウンロードして展開してください。

### Docker
```bash
docker pull ghcr.io/aoshimash/urlmap:v0.2.0
docker run --rm ghcr.io/aoshimash/urlmap:v0.2.0 https://example.com
```

## 🎯 基本的な使用方法

```bash
# 基本的なクロール
urlmap https://example.com

# 深度制限付きクロール
urlmap --depth 3 https://example.com

# 並行処理数を指定
urlmap --concurrent 20 https://example.com

# 詳細ログ出力
urlmap --verbose https://example.com

# バージョン確認
urlmap version
```

## 🔄 アップグレード

v0.1.0からのアップグレードは以下の手順で行えます：

```bash
# 新しいバイナリをダウンロード
curl -L -o urlmap.tar.gz https://github.com/aoshimash/urlmap/releases/download/v0.2.0/urlmap-linux-amd64.tar.gz
tar -xzf urlmap.tar.gz
chmod +x urlmap
sudo mv urlmap /usr/local/bin/

# またはDockerイメージを更新
docker pull ghcr.io/aoshimash/urlmap:v0.2.0
```

## ⚡ パフォーマンス

v0.2.0では、コードの一貫性向上により保守性が改善されました。パフォーマンス特性は以下の通りです：

- **高速クロール**: 並行処理による高性能クロール
- **メモリ効率**: 大規模サイトでの効率的なメモリ使用
- **割り込み安全**: 適切なクリーンアップによる安全な中断

## 🧪 テスト

- **単体テスト**: 全コンポーネントの包括的テストスイート
- **統合テスト**: CLI機能の統合テスト
- **E2Eテスト**: 完全なワークフローテスト

## 🙏 謝辞

このプロジェクトは以下のオープンソースライブラリを使用しています：

- [Cobra](https://github.com/spf13/cobra) - CLIフレームワーク
- [Resty](https://github.com/go-resty/resty) - HTTPクライアント
- [goquery](https://github.com/PuerkitoBio/goquery) - HTMLパーサー

## 📄 ライセンス

このプロジェクトはMITライセンスの下でリリースされています。詳細は[LICENSE](LICENSE)ファイルをご覧ください。

## 📞 フィードバック・サポート

- 🐛 バグ報告: [Issues](https://github.com/aoshimash/urlmap/issues)
- 💡 機能要望: [Issues](https://github.com/aoshimash/urlmap/issues)
- 🤔 質問・提案: [Discussions](https://github.com/aoshimash/urlmap/discussions)
- 📖 ドキュメント: [Wiki](https://github.com/aoshimash/urlmap/wiki)

## 🚀 次のバージョンに向けて

v0.3.0では以下の機能を計画しています：

- JavaScript動的生成リンクのサポート (WebDriver)
- 出力フォーマット選択肢 (JSON, CSV, XML)
- 強化されたフィルタリング機能
- プラグインシステム
- パフォーマンス最適化

---

**v0.2.0をお試しください！フィードバックお待ちしています。** 🚀

**注意**: このバージョンは安定版です。本番環境での使用前に十分にテストしてください。
