# urlmap v0.1.0 - Initial Release (Development Version)

🎉 **We're excited to announce the initial release of urlmap v0.1.0!**

urlmap is a high-performance Go-based CLI tool for efficiently discovering and mapping URLs within websites.

⚠️ **Note**: This is an initial development version (0.x.x). APIs and features may change in future versions.

## ✨ Key Features

### 🚀 Core Functionality
- **Recursive Link Discovery**: Automatically discover all links within a website
- **Same-Domain Filtering**: Focus crawling on specific domains to exclude external links
- **Concurrent Processing**: High-performance crawling with configurable worker pools
- **Depth Limiting**: Control crawl depth to prevent infinite recursion
- **Progress Indicators**: Real-time progress reporting during crawling operations

### ⚡ Performance Features
- **Rate Limiting**: Respectful crawling with configurable request rates
- **Graceful Shutdown**: Proper cleanup on interruption
- **Structured Logging**: Comprehensive logging with verbose mode support
- **Multiple Output Formats**: URLs output to stdout, logs to stderr
- **Custom User Agent**: Configurable user agent strings for identification

## 📦 Installation

### Binary Download

#### Linux (x86_64)
```bash
curl -L -o urlmap.tar.gz https://github.com/aoshimash/urlmap/releases/download/v0.1.0/urlmap-linux-amd64.tar.gz
tar -xzf urlmap.tar.gz
chmod +x urlmap
sudo mv urlmap /usr/local/bin/
```

#### macOS (Apple Silicon)
```bash
curl -L -o urlmap.tar.gz https://github.com/aoshimash/urlmap/releases/download/v0.1.0/urlmap-darwin-arm64.tar.gz
tar -xzf urlmap.tar.gz
chmod +x urlmap
sudo mv urlmap /usr/local/bin/
```

#### Windows
Download `urlmap-windows-amd64.zip` from the [releases page](https://github.com/aoshimash/urlmap/releases/tag/v0.1.0) and extract the executable.

### Docker
```bash
docker pull ghcr.io/aoshimash/urlmap:v0.1.0
docker run --rm ghcr.io/aoshimash/urlmap:v0.1.0 https://example.com
```

### Build from Source
```bash
git clone https://github.com/aoshimash/urlmap.git
cd urlmap
go build -o urlmap ./cmd/urlmap
```

## 🎯 Usage Examples

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

## 🔧 Command Line Options

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--depth` | `-d` | 0 (unlimited) | Maximum crawl depth |
| `--concurrent` | `-c` | 10 | Number of concurrent requests |
| `--verbose` | `-v` | false | Enable verbose logging |
| `--user-agent` | `-u` | urlmap/0.1.0 | Custom User-Agent string |
| `--progress` | `-p` | true | Show progress indicators |
| `--rate-limit` | `-r` | 0 (no limit) | Rate limit (requests per second) |

## 🚧 Development Status

This initial release (v0.1.0) includes the following implemented features:

✅ **Implemented**:
- Basic web crawling functionality
- High-performance concurrent crawling
- Same-domain filtering
- Depth limiting functionality
- Progress reporting and logging
- CLI interface
- Docker support

🔄 **Planned for future versions** (v0.2.0+):
- WebDriver support for JavaScript-rendered content
- Output format options (JSON, CSV, XML)
- Enhanced filtering capabilities
- Plugin system
- Performance optimizations
- Improved error handling

## ⚠️ Known Limitations

- Does not support JavaScript-dynamically generated links
- Current output format is plain text only
- May have unexpected behavior with some complex website structures

## 🧪 Test Coverage

- **Unit Tests**: Comprehensive test suite for all components
- **Integration Tests**: CLI functionality integration tests
- **E2E Tests**: Complete workflow testing
- **Performance Tests**: Concurrency and stress testing

## 📚 Documentation

- [README](README.md) - Basic usage and installation instructions
- [Architecture Documentation](docs/ARCHITECTURE.md) - Detailed design documentation
- [Performance Guide](docs/PERFORMANCE.md) - Performance optimization
- [Contributing Guide](CONTRIBUTING.md) - How to contribute to development

## 🙏 Acknowledgments

This project uses the following open source libraries:

- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Resty](https://github.com/go-resty/resty) - HTTP client
- [goquery](https://github.com/PuerkitoBio/goquery) - HTML parser

## 📄 License

This project is released under the MIT License. See the [LICENSE](LICENSE) file for details.

## 📞 Feedback & Support

As this is an initial development version, we actively welcome feedback!

- 🐛 Bug Reports: [Issues](https://github.com/aoshimash/urlmap/issues)
- 💡 Feature Requests: [Issues](https://github.com/aoshimash/urlmap/issues)
- 🤔 Questions & Suggestions: [Discussions](https://github.com/aoshimash/urlmap/discussions)
- 📖 Documentation: [Wiki](https://github.com/aoshimash/urlmap/wiki)

## ⚡ Quick Start

Simple examples for first-time users:

```bash
# 1. Download from releases page or use Docker
docker run --rm ghcr.io/aoshimash/urlmap:v0.1.0 https://example.com

# 2. Try with a small site
urlmap --depth 2 --verbose https://example.com

# 3. Save results to file
urlmap https://example.com > urls.txt
```

---

**Try v0.1.0 today! Your feedback will help make future versions even better.** 🚀

**Note**: This version is a development release. Please test thoroughly before using in production environments.
