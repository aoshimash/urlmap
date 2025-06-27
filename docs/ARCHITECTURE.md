# crawld Architecture

This document provides a comprehensive overview of the crawld architecture, design decisions, and implementation details.

## 📋 Table of Contents

- [Overview](#overview)
- [System Architecture](#system-architecture)
- [Core Components](#core-components)
- [Data Flow](#data-flow)
- [Concurrency Model](#concurrency-model)
- [Error Handling](#error-handling)
- [Configuration Management](#configuration-management)
- [Testing Strategy](#testing-strategy)
- [Performance Considerations](#performance-considerations)
- [Security Considerations](#security-considerations)

## 🏗 Overview

crawld is a concurrent web crawler designed for discovering and extracting links from websites. It follows a modular architecture with clear separation of concerns, enabling maintainability, testability, and extensibility.

### Key Design Principles

1. **Modularity**: Each component has a single responsibility
2. **Concurrency**: Leverages Go's goroutines for parallel processing
3. **Configurability**: Flexible configuration options for different use cases
4. **Reliability**: Robust error handling and graceful degradation
5. **Performance**: Optimized for high-throughput crawling
6. **Safety**: Thread-safe operations and resource management

## 🎯 System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        CLI Layer                            │
│  ┌─────────────────┐  ┌─────────────────┐  ┌──────────────┐ │
│  │   cmd/crawld    │  │     Cobra       │  │    Config    │ │
│  │   (main.go)     │  │   CLI Engine    │  │   Parsing    │ │
│  └─────────────────┘  └─────────────────┘  └──────────────┘ │
└─────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────┐
│                     Core Engine                             │
│  ┌─────────────────┐  ┌─────────────────┐  ┌──────────────┐ │
│  │   Concurrent    │  │     Worker      │  │  Progress    │ │
│  │    Crawler      │  │     Pool        │  │  Reporter    │ │
│  └─────────────────┘  └─────────────────┘  └──────────────┘ │
└─────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────┐
│                  Processing Layer                           │
│  ┌─────────────────┐  ┌─────────────────┐  ┌──────────────┐ │
│  │  HTTP Client    │  │  HTML Parser    │  │ URL Manager  │ │
│  │  (Resty)        │  │  (goquery)      │  │  Validator   │ │
│  └─────────────────┘  └─────────────────┘  └──────────────┘ │
└─────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────┐
│                     Output Layer                            │
│  ┌─────────────────┐  ┌─────────────────┐  ┌──────────────┐ │
│  │   URL Output    │  │  Log Output     │  │  Statistics  │ │
│  │   (stdout)      │  │   (stderr)      │  │  Reporting   │ │
│  └─────────────────┘  └─────────────────┘  └──────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## 🧩 Core Components

### 1. CLI Layer (`cmd/crawld/`)

**Purpose**: Entry point and user interface

**Components**:
- `main.go`: Application entry point
- Command parsing using Cobra framework
- Configuration validation and setup
- Signal handling for graceful shutdown

**Key Responsibilities**:
- Parse command-line arguments
- Initialize logging configuration
- Set up crawler with user parameters
- Handle interrupt signals
- Output results to stdout/stderr

### 2. Crawler Engine (`internal/crawler/`)

**Purpose**: Core crawling logic and coordination

**Components**:
- `Crawler`: Sequential crawler implementation
- `ConcurrentCrawler`: Parallel crawler with worker pool
- `Config`: Crawler configuration structure
- Job and result structures

**Key Responsibilities**:
- Coordinate crawling operations
- Manage worker pool for concurrent processing
- Track visited URLs to prevent duplicates
- Collect and aggregate results
- Provide crawling statistics

#### Worker Pool Architecture

```
                    ┌─────────────┐
                    │   Master    │
                    │  Goroutine  │
                    └──────┬──────┘
                           │
              ┌────────────┼────────────┐
              │            │            │
         ┌────▼────┐  ┌────▼────┐  ┌────▼────┐
         │Worker 1 │  │Worker 2 │  │Worker N │
         └─────────┘  └─────────┘  └─────────┘
              │            │            │
              └────────────┼────────────┘
                           │
                    ┌──────▼──────┐
                    │   Results   │
                    │ Collector   │
                    └─────────────┘
```

### 3. HTTP Client (`internal/client/`)

**Purpose**: HTTP request handling and management

**Components**:
- `Client`: HTTP client wrapper
- `Config`: Client configuration
- Request/response handling
- Retry logic and timeout management

**Key Responsibilities**:
- Execute HTTP requests
- Handle redirects and status codes
- Implement request timeouts
- Provide user-agent configuration
- Support for custom headers

### 4. HTML Parser (`internal/parser/`)

**Purpose**: HTML content parsing and link extraction

**Components**:
- `LinkExtractor`: Main parsing interface
- URL extraction logic
- HTML document processing

**Key Responsibilities**:
- Parse HTML content using goquery
- Extract href attributes from anchor tags
- Resolve relative URLs to absolute URLs
- Filter out non-HTTP/HTTPS links
- Handle malformed HTML gracefully

### 5. URL Management (`internal/url/`)

**Purpose**: URL validation, normalization, and filtering

**Components**:
- URL validation functions
- Normalization utilities
- Domain extraction logic

**Key Responsibilities**:
- Validate URL format and schemes
- Normalize URLs for consistency
- Extract domain names for filtering
- Handle edge cases in URL processing
- Support same-domain filtering

### 6. Progress Reporting (`internal/progress/`)

**Purpose**: Real-time progress tracking and statistics

**Components**:
- `ProgressReporter`: Progress tracking interface
- Statistics collection and aggregation
- Performance metrics

**Key Responsibilities**:
- Track crawling progress in real-time
- Collect performance statistics
- Report discovered/crawled/failed URLs
- Calculate crawling rates and timing
- Provide user feedback during operation

### 7. Configuration (`internal/config/`)

**Purpose**: Application configuration and logging setup

**Components**:
- Logging configuration
- Structured logging setup
- Configuration validation

**Key Responsibilities**:
- Set up structured logging with slog
- Configure log levels and output
- Validate configuration parameters
- Provide logging utilities

### 8. Output Management (`internal/output/`)

**Purpose**: Result formatting and output

**Components**:
- URL output formatting
- Result aggregation

**Key Responsibilities**:
- Format URLs for output
- Handle different output formats
- Manage stdout/stderr separation
- Support future output format extensions

## 🔄 Data Flow

### High-Level Flow

```
1. CLI Parsing → 2. Configuration → 3. Crawler Creation → 4. URL Processing → 5. Output
```

### Detailed Processing Flow

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│  Start URL  │───▶│  URL Queue  │───▶│   Worker    │
└─────────────┘    └─────────────┘    │    Pool     │
                                      └──────┬──────┘
                                             │
┌─────────────┐    ┌─────────────┐    ┌─────▼─────┐
│ Final URLs  │◀───│  Results    │◀───│  Process  │
└─────────────┘    │ Collector   │    │   Page    │
                   └─────────────┘    └───────────┘
```

### Concurrency Flow

1. **Job Distribution**: Master goroutine distributes URLs to worker pool
2. **Parallel Processing**: Workers fetch and parse pages concurrently
3. **Result Collection**: Results are collected through channels
4. **Link Discovery**: New links are added back to the job queue
5. **Duplicate Prevention**: Visited URLs are tracked to avoid cycles
6. **Completion Detection**: Crawling completes when no more jobs remain

## ⚡ Concurrency Model

### Goroutine Architecture

```go
type ConcurrentCrawler struct {
    jobs     chan CrawlJob      // Job distribution channel
    results  chan CrawlResult   // Result collection channel
    visited  sync.Map           // Thread-safe visited URLs
    wg       sync.WaitGroup     // Worker synchronization
    ctx      context.Context    // Cancellation context
    cancel   context.CancelFunc // Cancellation function
}
```

### Worker Pool Pattern

1. **Master Goroutine**: Manages overall crawling process
2. **Worker Goroutines**: Process individual crawl jobs
3. **Result Collector**: Aggregates results from workers
4. **Progress Updater**: Updates statistics and progress

### Synchronization Mechanisms

- **Channels**: For job distribution and result collection
- **sync.Map**: Thread-safe visited URL tracking
- **sync.WaitGroup**: Worker lifecycle management
- **context.Context**: Graceful cancellation support
- **Mutexes**: Protecting shared state where needed

## 🚨 Error Handling

### Error Categories

1. **Network Errors**: Connection failures, timeouts, DNS issues
2. **HTTP Errors**: 4xx/5xx status codes, malformed responses
3. **Parsing Errors**: Invalid HTML, encoding issues
4. **URL Errors**: Malformed URLs, unsupported schemes
5. **System Errors**: File I/O, memory allocation failures

### Error Handling Strategy

```go
// Example error handling pattern
func (c *ConcurrentCrawler) processJob(job CrawlJob) {
    result := CrawlResult{URL: job.URL, Depth: job.Depth}

    // Network request with error handling
    resp, err := c.client.Get(job.URL)
    if err != nil {
        result.Error = fmt.Errorf("failed to fetch %s: %w", job.URL, err)
        c.results <- result
        return
    }

    // Parse content with error handling
    links, err := c.parser.ExtractLinks(resp.Body, job.URL)
    if err != nil {
        c.logger.Warn("Failed to parse HTML", "url", job.URL, "error", err)
        // Continue with partial results
    }

    result.Links = links
    c.results <- result
}
```

### Recovery Mechanisms

- **Graceful Degradation**: Continue processing despite individual failures
- **Retry Logic**: Built into HTTP client for transient failures
- **Timeout Handling**: Prevent hanging on slow or unresponsive servers
- **Resource Cleanup**: Proper cleanup on errors and cancellation

## ⚙️ Configuration Management

### Configuration Sources

1. **Command-line Flags**: Primary configuration method
2. **Default Values**: Sensible defaults for all parameters
3. **Environment Variables**: Future extension possibility

### Configuration Structure

```go
type Config struct {
    MaxDepth       int              // Crawling depth limit
    SameDomain     bool             // Domain filtering
    UserAgent      string           // HTTP user agent
    Timeout        time.Duration    // Request timeout
    Workers        int              // Concurrent workers
    ShowProgress   bool             // Progress display
    Logger         *slog.Logger     // Logging instance
    ProgressConfig *progress.Config // Progress configuration
}
```

### Validation

- Parameter range checking
- URL format validation
- Dependency verification
- Resource limit validation

## 🧪 Testing Strategy

### Test Categories

1. **Unit Tests**: Individual component testing
2. **Integration Tests**: Component interaction testing
3. **End-to-End Tests**: Full application testing
4. **Performance Tests**: Load and stress testing

### Test Structure

```
test/
├── unit/           # Unit tests (co-located with code)
├── integration/    # Integration tests
├── e2e/           # End-to-end tests
├── fixtures/      # Test data and mock servers
└── shared/        # Common test utilities
```

### Mock Strategy

- HTTP server mocking for network tests
- Interface-based mocking for dependencies
- Test fixtures for HTML content
- Deterministic test scenarios

## 🚀 Performance Considerations

### Optimization Strategies

1. **Connection Pooling**: Reuse HTTP connections
2. **Concurrent Processing**: Parallel worker pools
3. **Memory Management**: Efficient data structures
4. **Resource Limits**: Configurable worker counts
5. **Early Termination**: Depth and domain limits

### Performance Metrics

- **Throughput**: URLs processed per second
- **Latency**: Response time per URL
- **Memory Usage**: Peak and average memory consumption
- **Error Rate**: Failed requests percentage
- **Concurrency**: Active worker utilization

### Bottleneck Identification

1. **Network I/O**: Usually the primary bottleneck
2. **CPU Usage**: HTML parsing and URL processing
3. **Memory**: URL storage and result collection
4. **Disk I/O**: Logging and output generation

## 🔒 Security Considerations

### Security Measures

1. **URL Validation**: Prevent malicious URL injection
2. **Request Limits**: Rate limiting and timeout protection
3. **Memory Limits**: Prevent resource exhaustion
4. **Input Sanitization**: Safe HTML parsing
5. **TLS Verification**: Proper certificate validation

### Attack Prevention

- **DoS Protection**: Rate limiting and worker limits
- **SSRF Prevention**: URL scheme and domain validation
- **XSS Prevention**: Safe HTML content handling
- **Resource Exhaustion**: Memory and time limits

### Privacy Considerations

- **User Agent**: Configurable identification
- **Respectful Crawling**: Rate limiting support
- **Data Handling**: No persistent storage of content
- **Logging**: Configurable log levels

## 📈 Future Extensions

### Planned Enhancements

1. **Output Formats**: JSON, CSV, XML support
2. **Plugin System**: Custom processing hooks
3. **Distributed Crawling**: Multi-node coordination
4. **Robots.txt Support**: Configurable respect for robots.txt
5. **Web UI**: Browser-based monitoring interface

### Extension Points

- **Parser Interface**: Support for different content types
- **Client Interface**: Support for different HTTP libraries
- **Output Interface**: Support for different output formats
- **Storage Interface**: Support for persistent storage

---

This architecture provides a solid foundation for a scalable, maintainable, and extensible web crawler while maintaining simplicity and performance.
