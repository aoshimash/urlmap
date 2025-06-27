---
title: "CLI Development Guidelines"
description: "Cobra and CLI-specific development standards for crawld"
type: "Auto Attached"
patterns: ["cmd/**/*.go", "**/*_cmd.go", "**/*command*.go"]
---

# CLI Development Guidelines

## Cobra Framework Standards

### Command Structure
- Use Cobra for all CLI functionality
- Organize commands in `cmd/` directory
- Each subcommand in separate file if complex

### Flag Definitions
- Use long flags with meaningful names: `--depth`, `--output-format`
- Provide short flags for common options: `-d`, `-o`
- Set reasonable default values
- Use persistent flags for options that apply to subcommands

### Example Command Structure:
```go
var rootCmd = &cobra.Command{
    Use:   "crawld [URL]",
    Short: "Crawl URLs and extract links recursively",
    Long: `crawld is a CLI tool that crawls web pages starting from a given URL
and recursively discovers all links within the same domain.`,
    Args: cobra.ExactArgs(1),
    RunE: runCrawl,
}

func init() {
    rootCmd.PersistentFlags().IntVarP(&maxDepth, "depth", "d", 0, "Maximum crawl depth (0 = unlimited)")
    rootCmd.PersistentFlags().StringVarP(&userAgent, "user-agent", "u", defaultUserAgent, "Custom User-Agent string")
    rootCmd.PersistentFlags().IntVarP(&concurrent, "concurrent", "c", 10, "Number of concurrent requests")
    rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")
}
```

## CLI Best Practices

### User Experience
- Provide helpful usage examples in command descriptions
- Use consistent flag naming across commands
- Implement `--help` with clear descriptions
- Support `--version` flag
- Provide progress indicators for long operations

### Output Standards
- Default output: plain text, one URL per line
- Support `--quiet` flag to suppress progress messages
- Use stderr for logs and progress, stdout for results
- Ensure output is pipeline-friendly

### Configuration
- Use environment variable prefix: `CRAWLD_`
- Support config file if needed (later enhancement)
- Flag precedence: CLI flags > environment variables > config file > defaults

### Error Handling
- Use appropriate exit codes:
  - 0: Success
  - 1: General error
  - 2: Invalid arguments
  - 3: Network error
- Provide actionable error messages
- Show usage information for argument errors

## Logging Integration

### slog Configuration
```go
// Configure slog based on verbosity
func setupLogging(verbose bool) {
    level := slog.LevelWarn
    if verbose {
        level = slog.LevelInfo
    }

    handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
        Level: level,
    })
    logger := slog.New(handler)
    slog.SetDefault(logger)
}
```

### Logging Levels
- **Debug**: Detailed execution flow (only in debug builds)
- **Info**: Progress information when verbose mode is enabled
- **Warn**: Recoverable issues (invalid URLs, timeouts)
- **Error**: Fatal errors that stop execution
