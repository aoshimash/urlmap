# urlmap v0.2.0 - Codebase Unification and Documentation Improvements

🚀 **We're excited to announce urlmap v0.2.0!**

This maintenance release includes complete codebase unification and comprehensive documentation improvements.

## ✨ Key Improvements in This Release

### 🔧 Codebase Improvements
- **Complete Naming Unification**: Unified all remaining `crawld` references to `urlmap` for consistency
- **Code Cleanup**: Improved naming consistency across internal components
- **Enhanced Maintainability**: Better code organization and structure

### 📚 Documentation Enhancements
- **AI-Driven Development Section**: Added comprehensive AI-driven development section to README
- **Developer Experience**: More detailed and accessible documentation
- **Improved Clarity**: Better explanations and examples throughout

## 🆙 Changes from v0.1.0

### Changed
- refactor: Replace all remaining crawld references with urlmap
- docs: Add AI-driven development section to README
- improve: Enhanced code consistency across internal components

### Fixed
- Resolved naming inconsistencies between internal components
- Improved overall codebase consistency and maintainability

## 📦 Installation

### Binary Download

Download the latest binary from the [releases page](https://github.com/aoshimash/urlmap/releases/tag/v0.2.0):

#### Linux (x86_64)
```bash
curl -L -o urlmap.tar.gz https://github.com/aoshimash/urlmap/releases/download/v0.2.0/urlmap-linux-amd64.tar.gz
tar -xzf urlmap.tar.gz
chmod +x urlmap
sudo mv urlmap /usr/local/bin/
```

#### Linux (ARM64)
```bash
curl -L -o urlmap.tar.gz https://github.com/aoshimash/urlmap/releases/download/v0.2.0/urlmap-linux-arm64.tar.gz
tar -xzf urlmap.tar.gz
chmod +x urlmap
sudo mv urlmap /usr/local/bin/
```

#### macOS (Intel)
```bash
curl -L -o urlmap.tar.gz https://github.com/aoshimash/urlmap/releases/download/v0.2.0/urlmap-darwin-amd64.tar.gz
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
Download `urlmap-windows-amd64.zip` from the [releases page](https://github.com/aoshimash/urlmap/releases/tag/v0.2.0) and extract the executable.

### Docker

Run with Docker without installation:

```bash
# Pull from GitHub Container Registry
docker pull ghcr.io/aoshimash/urlmap:v0.2.0

# Basic usage
docker run --rm ghcr.io/aoshimash/urlmap:v0.2.0 https://example.com
```

## 🎯 Usage

### Basic Usage

```bash
# Crawl a website with default settings
urlmap https://example.com

# Check version
urlmap version

# Get help
urlmap --help
```

### Advanced Options

```bash
# Limit crawl depth to 3 levels
urlmap --depth 3 https://example.com

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
# Basic crawling
docker run --rm ghcr.io/aoshimash/urlmap:v0.2.0 https://example.com

# With options
docker run --rm ghcr.io/aoshimash/urlmap:v0.2.0 --depth 3 --concurrent 20 https://example.com

# Save output to file
docker run --rm ghcr.io/aoshimash/urlmap:v0.2.0 https://example.com > urls.txt
```

## 🔄 Upgrading from v0.1.0

To upgrade from v0.1.0 to v0.2.0:

```bash
# Download new binary
curl -L -o urlmap.tar.gz https://github.com/aoshimash/urlmap/releases/download/v0.2.0/urlmap-linux-amd64.tar.gz
tar -xzf urlmap.tar.gz
chmod +x urlmap
sudo mv urlmap /usr/local/bin/

# Or update Docker image
docker pull ghcr.io/aoshimash/urlmap:v0.2.0
```

No configuration changes are required. All existing command-line options remain the same.

## ⚡ Performance

v0.2.0 improves maintainability through enhanced code consistency. Performance characteristics include:

- **High-Speed Crawling**: Efficient concurrent processing for fast website mapping
- **Memory Efficient**: Optimized memory usage for large-scale site crawling
- **Graceful Shutdown**: Interrupt-safe operation with proper cleanup
- **Rate Limiting**: Respectful crawling with configurable request rates

## 🔧 Command Line Options

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--depth` | `-d` | 0 (unlimited) | Maximum crawl depth |
| `--concurrent` | `-c` | 10 | Number of concurrent workers |
| `--verbose` | `-v` | false | Enable verbose logging |
| `--user-agent` | `-u` | urlmap/0.2.0 | Custom User-Agent string |
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

Planned features for v0.3.0:

- **JavaScript Support**: WebDriver integration for JavaScript-rendered content
- **Multiple Output Formats**: JSON, CSV, XML export options
- **Enhanced Filtering**: Advanced URL filtering capabilities
- **Plugin System**: Extensible architecture for custom functionality
- **Performance Optimizations**: Further improvements to crawling speed and efficiency

---

**Try v0.2.0 today! Your feedback helps make future versions even better.** 🚀

**Note**: This is a stable release. Please test thoroughly before using in production environments.
