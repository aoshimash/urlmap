# Crawld

[![CI](https://github.com/aoshimash/crawld/workflows/CI/badge.svg)](https://github.com/aoshimash/crawld/actions/workflows/ci.yml)
[![Docker](https://github.com/aoshimash/crawld/workflows/Docker%20Build%20and%20Publish/badge.svg)](https://github.com/aoshimash/crawld/actions/workflows/docker.yml)

A web crawler daemon for collecting and processing web content.

## 🚀 Features

- CLI-based web crawler
- HTML parsing and content extraction
- HTTP client with retry capabilities
- Modular architecture for extensibility

## 📋 Project Structure

```
crawld/
├── cmd/                # Command-line applications
│   └── crawld/         # Main CLI application
├── internal/           # Private application code
│   └── crawler/        # Core crawler logic
├── pkg/               # Public library code
│   └── utils/         # Utility functions
├── go.mod             # Go module file
├── go.sum             # Go dependency checksums
└── README.md          # This file
```

## 🛠 Installation

```bash
# Clone the repository
git clone https://github.com/aoshimash/crawld.git
cd crawld

# Build the application
go build -o bin/crawld ./cmd/crawld

# Run the application
./bin/crawld
```

## 🎯 Usage

### Binary Usage

```bash
# Basic usage
crawld --help

# Crawl a website
crawld https://example.com

# Advanced usage with options
crawld -d 3 -c 5 --verbose https://example.com
```

### Docker Usage

```bash
# Pull from GitHub Container Registry
docker pull ghcr.io/aoshimash/crawld:latest

# Run with Docker
docker run --rm ghcr.io/aoshimash/crawld:latest --help

# Crawl a website using Docker
docker run --rm ghcr.io/aoshimash/crawld:latest https://example.com

# Advanced usage with Docker
docker run --rm ghcr.io/aoshimash/crawld:latest -d 3 -c 5 --verbose https://example.com
```

## 🧪 Development

### Prerequisites

- Go 1.21 or higher

### Continuous Integration

This project uses GitHub Actions for continuous integration. The CI workflow:

- Tests on Go 1.21 and 1.22
- Runs `go fmt`, `go vet`, and `go test`
- Includes dependency caching for faster builds
- Generates code coverage reports

### Build

```bash
go build ./...
```

### Test

```bash
go test ./...
```

### Lint

```bash
go vet ./...
```

## 📚 Dependencies

- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Resty](https://github.com/go-resty/resty) - HTTP client library
- [goquery](https://github.com/PuerkitoBio/goquery) - HTML parsing

## 📝 License

See [LICENSE](LICENSE) file for details.
A fast recursive web crawler for extracting all URLs from websites.
