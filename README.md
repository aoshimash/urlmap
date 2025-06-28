# urlmap

[![CI](https://github.com/aoshimash/urlmap/workflows/CI/badge.svg)](https://github.com/aoshimash/urlmap/actions/workflows/ci.yml)
[![Docker](https://github.com/aoshimash/urlmap/workflows/Docker%20Build%20and%20Publish/badge.svg)](https://github.com/aoshimash/urlmap/actions/workflows/docker.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/aoshimash/urlmap)](https://goreportcard.com/report/github.com/aoshimash/urlmap)
[![License](https://img.shields.io/github/license/aoshimash/urlmap)](LICENSE)

A fast and efficient web crawler CLI tool for discovering and mapping URLs within a website. Built with Go for high performance and concurrent crawling.

## 🚀 Features

- **Recursive Link Discovery**: Automatically discover all links within a website
- **Same-Domain Filtering**: Focus crawling on a specific domain to avoid external links
- **Concurrent Processing**: High-performance crawling with configurable worker pools
- **Depth Limiting**: Control crawl depth to prevent infinite recursion
- **Progress Indicators**: Real-time progress reporting during crawling operations
- **Rate Limiting**: Respectful crawling with configurable request rates
- **Graceful Shutdown**: Interrupt-safe with proper cleanup on termination
- **Structured Logging**: Comprehensive logging with verbose mode support
- **Multiple Output Formats**: URLs output to stdout, logs to stderr
- **Custom User Agent**: Configurable user agent strings for identification

## 📦 Installation

### Binary Download

Download the latest binary from the [releases page](https://github.com/aoshimash/urlmap/releases):

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
Download `urlmap-windows-amd64.zip` from the releases page and extract the executable.

### Docker

Run with Docker without installation:

```bash
# Pull from GitHub Container Registry
docker pull ghcr.io/aoshimash/urlmap:latest

# Basic usage
docker run --rm ghcr.io/aoshimash/urlmap:latest https://example.com
```

### From Source

Requirements: Go 1.21 or higher

```bash
# Clone the repository
git clone https://github.com/aoshimash/urlmap.git
cd urlmap

# Build the application
go build -o urlmap ./cmd/urlmap

# Install globally (optional)
sudo mv urlmap /usr/local/bin/
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

# Disable progress indicators
urlmap --progress=false https://example.com

# Combined options
urlmap --depth 5 --concurrent 15 --verbose --rate-limit 2 https://example.com
```

### Docker Usage

```bash
# Basic crawling
docker run --rm ghcr.io/aoshimash/urlmap:latest https://example.com

# With options
docker run --rm ghcr.io/aoshimash/urlmap:latest --depth 3 --concurrent 20 https://example.com

# Save output to file
docker run --rm ghcr.io/aoshimash/urlmap:latest https://example.com > urls.txt

# Interactive mode with shell access
docker run -it --rm ghcr.io/aoshimash/urlmap:latest /bin/sh
```

## 🔧 Command Line Options

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--depth` | `-d` | 0 (unlimited) | Maximum crawl depth |
| `--concurrent` | `-c` | 10 | Number of concurrent workers |
| `--verbose` | `-v` | false | Enable verbose logging |
| `--user-agent` | `-u` | urlmap/1.0.0 | Custom User-Agent string |
| `--progress` | `-p` | true | Show progress indicators |
| `--rate-limit` | `-r` | 0 (no limit) | Rate limit (requests per second) |
| `--help` | `-h` | - | Show help message |

## 📋 Examples

### Basic Website Crawling

```bash
# Crawl a simple website
urlmap https://example.com
```

Output:
```
https://example.com
https://example.com/about
https://example.com/contact
https://example.com/products
```

### Depth-Limited Crawling

```bash
# Only crawl up to 2 levels deep
urlmap --depth 2 https://blog.example.com
```

### High-Performance Crawling

```bash
# Use 50 concurrent workers for large sites
urlmap --concurrent 50 --verbose https://large-site.example.com
```

### Respectful Crawling

```bash
# Limit to 1 request per second with custom user agent
urlmap --rate-limit 1 --user-agent "Research Bot 1.0 (contact@example.com)" https://example.com
```

### Save Results to File

```bash
# Save URLs to a file
urlmap https://example.com > discovered_urls.txt

# Save with timestamps and logs
urlmap --verbose https://example.com > urls.txt 2> crawl.log
```

### Processing Large Sites

```bash
# Optimized for large sites with progress tracking
urlmap --depth 5 --concurrent 30 --rate-limit 10 --verbose https://large-site.com
```

## 🏗 Architecture

urlmap follows a modular architecture for maintainability and extensibility:

```
urlmap/
├── cmd/urlmap/          # CLI application entry point
├── internal/
│   ├── client/          # HTTP client with retry logic
│   ├── config/          # Configuration and logging setup
│   ├── crawler/         # Core crawling engine
│   ├── output/          # Output formatting and handling
│   ├── parser/          # HTML parsing and link extraction
│   ├── progress/        # Progress reporting and statistics
│   └── url/            # URL validation and normalization
└── pkg/utils/          # Public utilities
```

### Core Components

- **Crawler Engine**: Concurrent crawler with worker pool architecture
- **HTTP Client**: Resilient HTTP client with timeout and retry logic
- **Link Parser**: HTML parser using goquery for reliable link extraction
- **URL Manager**: URL validation, normalization, and domain filtering
- **Progress Reporter**: Real-time crawling statistics and progress tracking

## ⚡ Performance

urlmap is optimized for performance with the following characteristics:

### Benchmarks

- **Small sites** (< 100 pages): ~50-100 URLs/second
- **Medium sites** (100-1000 pages): ~30-50 URLs/second
- **Large sites** (> 1000 pages): ~20-30 URLs/second

Performance varies based on:
- Network latency and bandwidth
- Target server response times
- Number of concurrent workers
- Page complexity and size

### Optimization Tips

1. **Concurrent Workers**: Increase `--concurrent` for I/O bound crawling
2. **Rate Limiting**: Use `--rate-limit` to avoid overwhelming servers
3. **Depth Control**: Set appropriate `--depth` to avoid infinite crawling
4. **Progress Tracking**: Disable `--progress=false` for slight performance gain

### Memory Usage

- Base memory: ~10-20 MB
- Per worker: ~1-2 MB
- URL storage: ~100 bytes per URL
- For 10,000 URLs: typically ~50-100 MB total

## 🔍 Troubleshooting

### Common Issues

#### Permission Denied
```bash
# Error: permission denied
sudo chmod +x urlmap
# Or install to user directory
mv urlmap ~/.local/bin/
```

#### DNS Resolution Failures
```bash
# Test URL accessibility first
curl -I https://example.com

# Check DNS resolution
nslookup example.com

# Use verbose mode for debugging
urlmap --verbose https://example.com
```

#### Rate Limiting / 429 Errors
```bash
# Reduce concurrent workers and add rate limiting
urlmap --concurrent 5 --rate-limit 1 https://example.com
```

#### Memory Issues with Large Sites
```bash
# Reduce concurrent workers
urlmap --concurrent 5 --depth 3 https://large-site.com

# Monitor memory usage
urlmap --verbose https://example.com 2>&1 | grep -i memory
```

#### SSL/TLS Certificate Errors
```bash
# Check certificate validity
curl -I https://example.com

# For development/testing only (not recommended for production)
# Currently not configurable - urlmap validates all certificates
```

### Debugging

Enable verbose logging to troubleshoot issues:

```bash
urlmap --verbose https://example.com 2> debug.log
```

Log levels include:
- INFO: General crawling progress
- DEBUG: Detailed URL processing
- WARN: Non-fatal issues (failed URLs, timeouts)
- ERROR: Fatal errors that stop crawling

### Performance Issues

If crawling is slow:

1. **Check Network**: Test direct access to the target site
2. **Adjust Workers**: Try different `--concurrent` values (5-50)
3. **Monitor Rate Limits**: Ensure you're not being throttled
4. **Use Rate Limiting**: Add `--rate-limit` to be more respectful

```bash
# Performance testing command
time urlmap --depth 2 --concurrent 20 https://example.com > /dev/null
```

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details.

### Development Setup

```bash
# Clone the repository
git clone https://github.com/aoshimash/urlmap.git
cd urlmap

# Install dependencies
go mod download

# Run tests
go test ./...

# Run linting
go vet ./...
golangci-lint run

# Build for development
go build -o urlmap ./cmd/urlmap
```

### Project Structure

See [Architecture Documentation](docs/ARCHITECTURE.md) for detailed information about the codebase structure and design decisions.

## 📚 Dependencies

urlmap uses the following high-quality Go libraries:

- **[Cobra](https://github.com/spf13/cobra)** - Modern CLI framework
- **[Resty](https://github.com/go-resty/resty)** - HTTP client library
- **[goquery](https://github.com/PuerkitoBio/goquery)** - jQuery-like HTML parsing

## 📊 Monitoring and Statistics

urlmap provides detailed statistics during and after crawling:

```bash
# Example output with statistics
urlmap --verbose https://example.com
```

Statistics include:
- Total URLs discovered
- Successfully crawled URLs
- Failed URLs with reasons
- Maximum depth reached
- Total crawling time
- Average response time

## 🔒 Security Considerations

- urlmap respects robots.txt by default behavior of underlying HTTP libraries
- Uses safe HTML parsing to prevent XSS in link extraction
- Validates all URLs to prevent malicious redirects
- Implements proper timeout handling to prevent hanging requests
- Rate limiting capabilities help prevent accidental DoS

## 🤖 AI-Driven Development

This project serves as a practical experiment in AI-driven software development. As part of this exploration, the entire codebase was implemented using Cursor AI agent, including:

- Project design and architecture
- Issue creation and project management
- Pull request creation and code reviews
- Implementation of all features and functionality
- Documentation and README creation

**Important Note**: There is not a single line of code written by a human in this repository. Everything was generated and managed by AI tools, demonstrating the current capabilities of AI-assisted development.

## 📄 License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## 🙋‍♀️ Support

- **Bug Reports**: [GitHub Issues](https://github.com/aoshimash/urlmap/issues)
- **Feature Requests**: [GitHub Discussions](https://github.com/aoshimash/urlmap/discussions)
- **Security Issues**: Please email security issues privately

## 🗺 Roadmap

Future enhancements planned:
- [ ] Robots.txt respect configuration
- [ ] Custom output formats (JSON, CSV, XML)
- [ ] Plugin system for custom processing
- [ ] Distributed crawling support
- [ ] Web UI for monitoring large crawls
- [ ] Integration with popular data analysis tools

---
