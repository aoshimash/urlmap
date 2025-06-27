# Crawld

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

```bash
# Basic usage
crawld --help

# Example usage (to be implemented)
crawld crawl https://example.com
```

## 🧪 Development

### Prerequisites

- Go 1.21 or higher

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
