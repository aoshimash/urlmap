# urlmap v0.3.0 - Enhanced Default Behavior and Bug Fixes

🚀 **We're excited to announce urlmap v0.3.0!**

This release focuses on improving the default crawling behavior and fixing important issues to enhance the user experience.

## ✨ Key Improvements in This Release

### 🔧 Enhanced Default Behavior
- **Unlimited Default Crawling**: Changed default crawl depth from 3 to unlimited (-1) for more comprehensive site mapping
- **Improved User Experience**: Users no longer need to specify depth manually for complete site crawling

### 🐛 Bug Fixes
- **Go Report Card Badge**: Fixed Go Report Card badge URL to resolve cache issues and display correct project status
- **Documentation Updates**: Improved documentation accuracy and clarity

## 🆙 Changes from v0.2.0

### Changed
- feat: Change default crawl depth to unlimited (-1)
- fix: Update Go Report Card badge URL to resolve cache issue
- docs: Translate v0.2.0 release notes to English

### Improved
- Enhanced default crawling behavior for better out-of-the-box experience
- Better project status visibility through fixed badge
- Improved documentation accessibility

## 📦 Installation

### Binary Download

Download the latest binary from the [releases page](https://github.com/aoshimash/urlmap/releases/tag/v0.3.0):

#### Linux (x86_64)
```bash
curl -L -o urlmap.tar.gz https://github.com/aoshimash/urlmap/releases/download/v0.3.0/urlmap-linux-amd64.tar.gz
tar -xzf urlmap.tar.gz
chmod +x urlmap
sudo mv urlmap /usr/local/bin/
```

#### Linux (ARM64)
```bash
curl -L -o urlmap.tar.gz https://github.com/aoshimash/urlmap/releases/download/v0.3.0/urlmap-linux-arm64.tar.gz
tar -xzf urlmap.tar.gz
chmod +x urlmap
sudo mv urlmap /usr/local/bin/
```

#### macOS (Intel)
```bash
curl -L -o urlmap.tar.gz https://github.com/aoshimash/urlmap/releases/download/v0.3.0/urlmap-darwin-amd64.tar.gz
tar -xzf urlmap.tar.gz
chmod +x urlmap
sudo mv urlmap /usr/local/bin/
```

#### macOS (Apple Silicon)
```bash
curl -L -o urlmap.tar.gz https://github.com/aoshimash/urlmap/releases/download/v0.3.0/urlmap-darwin-arm64.tar.gz
tar -xzf urlmap.tar.gz
chmod +x urlmap
sudo mv urlmap /usr/local/bin/
```

#### Windows
Download `urlmap-windows-amd64.zip` from the [releases page](https://github.com/aoshimash/urlmap/releases/tag/v0.3.0) and extract the executable.

### Docker

Run with Docker without installation:

```bash
# Pull from GitHub Container Registry
docker pull ghcr.io/aoshimash/urlmap:v0.3.0

# Basic usage (now crawls entire site by default)
docker run --rm ghcr.io/aoshimash/urlmap:v0.3.0 https://example.com
```

## 🎯 Usage

### Basic Usage

```bash
# Crawl entire website (unlimited depth by default)
urlmap https://example.com

# Limit crawl depth if needed
urlmap --depth 3 https://example.com

# Check version
urlmap version

# Get help
urlmap --help
```

### Advanced Options

```bash
# Use 20 concurrent workers for faster crawling
urlmap --concurrent 20 https://example.com

# Enable verbose logging
urlmap --verbose https://example.com

# Custom user agent
urlmap --user-agent "MyBot/1.0" https://example.com

# Rate limiting (5 requests per second)
urlmap --rate-limit 5 https://example.com

# Combined options
urlmap --depth 5 --concurrent 15 --verbose --rate-limit 2 https://example.com
```

### Docker Usage

```bash
# Basic crawling (unlimited depth by default)
docker run --rm ghcr.io/aoshimash/urlmap:v0.3.0 https://example.com

# With options
docker run --rm ghcr.io/aoshimash/urlmap:v0.3.0 --depth 3 --concurrent 20 https://example.com

# Save output to file
docker run --rm ghcr.io/aoshimash/urlmap:v0.3.0 https://example.com > urls.txt
```

## 🔄 Upgrading from v0.2.0

To upgrade from v0.2.0 to v0.3.0:

```bash
# Download new binary
curl -L -o urlmap.tar.gz https://github.com/aoshimash/urlmap/releases/download/v0.3.0/urlmap-linux-amd64.tar.gz
tar -xzf urlmap.tar.gz
chmod +x urlmap
sudo mv urlmap /usr/local/bin/

# Or update Docker image
docker pull ghcr.io/aoshimash/urlmap:v0.3.0
```

**⚠️ Breaking Change Notice**: The default crawl depth has changed from 3 to unlimited (-1). If you were relying on the previous default depth limit, you may need to explicitly specify `--depth 3` in your commands.

## ⚡ Performance

v0.3.0 maintains excellent performance characteristics:

- **Comprehensive Crawling**: Now crawls entire sites by default for complete mapping
- **High-Speed Processing**: Efficient concurrent processing for fast website mapping
- **Memory Efficient**: Optimized memory usage for large-scale site crawling
- **Graceful Shutdown**: Interrupt-safe operation with proper cleanup
- **Rate Limiting**: Respectful crawling with configurable request rates

## 🔧 Command Line Options

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--depth` | `-d` | -1 (unlimited) | Maximum crawl depth |
| `--concurrent` | `-c` | 10 | Number of concurrent workers |
| `--verbose` | `-v` | false | Enable verbose logging |
| `--user-agent` | `-u` | urlmap/0.3.0 | Custom User-Agent string |
| `--progress` | `-p` | true | Show progress indicators |
| `--rate-limit` | `-r` | 0 (no limit) | Rate limit (requests per second) |
| `--help` | `-h` | - | Show help message |

## 🧪 Testing

- **Unit Tests**: Comprehensive test suite for all components
- **Integration Tests**: CLI functionality integration tests
- **E2E Tests**: Complete workflow testing
- **Performance Tests**: Concurrency and stress testing

## 🙏 Acknowledgments

This project uses the following open source libraries:

- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Resty](https://github.com/go-resty/resty) - HTTP client
- [goquery](https://github.com/PuerkitoBio/goquery) - HTML parser

## 📄 License

This project is released under the MIT License. See the [LICENSE](LICENSE) file for details.

## 📞 Feedback & Support

- 🐛 Bug Reports: [Issues](https://github.com/aoshimash/urlmap/issues)
- 💡 Feature Requests: [Issues](https://github.com/aoshimash/urlmap/issues)
- 🤔 Questions & Suggestions: [Discussions](https://github.com/aoshimash/urlmap/discussions)
- 📖 Documentation: [Wiki](https://github.com/aoshimash/urlmap/wiki)

## 🔮 What's Next?

Looking ahead to future releases:

- Enhanced output formats (JSON, XML, CSV)
- Advanced filtering capabilities
- Performance optimizations for very large sites
- Plugin system for custom processing

---

Thank you for using urlmap! 🚀
