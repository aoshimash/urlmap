# Integration and E2E Tests

This directory contains integration tests and end-to-end (E2E) tests for crawld.

## Directory Structure

```
test/
├── integration/        # Integration tests for CLI functionality
├── e2e/               # End-to-end tests with complete workflows
├── shared/            # Shared test utilities and helpers
└── README.md          # This documentation
```

## Test Categories

### Integration Tests (`test/integration/`)

Integration tests verify the complete CLI functionality by:
- Building and executing the crawld binary
- Testing various command-line flag combinations
- Verifying output format and content
- Testing error scenarios and edge cases

Key test files:
- `cli_test.go` - Core CLI functionality tests

### E2E Tests (`test/e2e/`)

End-to-end tests provide comprehensive workflow testing by:
- Setting up complex test servers with realistic content
- Testing complete crawling workflows with various configurations
- Verifying performance characteristics under load
- Testing error handling and recovery
- Validating signal handling and graceful shutdown

Key test files:
- `e2e_test.go` - Comprehensive E2E test scenarios

### Shared Utilities (`test/shared/`)

Shared utilities provide common functionality for tests:
- Test server creation with various configurations
- Binary building and management
- Test environment setup and cleanup
- Common assertion helpers

Key files:
- `testutils.go` - Shared test utilities and helpers

## Running Tests

### Prerequisites

- Go 1.19 or later
- All project dependencies installed (`go mod download`)

### Running All Tests

```bash
# Run all tests from project root
go test ./test/...

# Run with verbose output
go test -v ./test/...

# Run with race detection
go test -race ./test/...
```

### Running Specific Test Categories

```bash
# Integration tests only
go test ./test/integration/

# E2E tests only
go test ./test/e2e/

# Skip long-running tests
go test -short ./test/...
```

### Running Individual Tests

```bash
# Run specific test function
go test -run TestCrawlCommand_BasicFunctionality ./test/integration/

# Run specific E2E test
go test -run TestE2E_CompleteWorkflow ./test/e2e/
```

## Test Environment

### Test Servers

Tests use `httptest.Server` to create controlled test environments:
- **Basic Server**: Simple HTML pages with standard navigation
- **Complex Server**: Multi-level site structure with various content types
- **Error Server**: Simulates various HTTP error conditions
- **Slow Server**: Tests timeout and performance behavior

### Binary Building

Tests automatically build the crawld binary:
- Built to temporary directory for each test run
- Cleaned up automatically after tests complete
- Uses same build process as production builds

### Test Data

Test HTML content includes:
- Standard navigation patterns
- Various link structures (relative, absolute)
- Nested page hierarchies
- Error conditions (404, 500, timeouts)

## Test Scenarios Covered

### CLI Integration Tests

- [x] Basic crawling functionality
- [x] Depth limiting (`--depth` flag)
- [x] Verbose output (`--verbose` flag)
- [x] Concurrency settings (`--concurrent` flag)
- [x] Custom User-Agent (`--user-agent` flag)
- [x] Invalid URL handling
- [x] Network error handling
- [x] Version command (`version` subcommand)

### E2E Test Scenarios

- [x] Complete workflow with complex site structure
- [x] High concurrency stress testing
- [x] Mixed success/error scenario handling
- [x] Output format validation
- [x] Signal handling and graceful shutdown
- [x] Performance characteristics validation

### Error Scenarios

- [x] Invalid URLs and malformed inputs
- [x] Network connectivity issues
- [x] HTTP error responses (4xx, 5xx)
- [x] Timeout conditions
- [x] Resource exhaustion scenarios

### Performance Testing

- [x] Concurrent request handling
- [x] Large site crawling
- [x] Memory usage patterns
- [x] Response time characteristics

## Test Configuration

### Environment Variables

Tests respect the following environment variables:
- `HTTP_TIMEOUT` - Override default HTTP timeout for tests
- `TEST_VERBOSE` - Enable verbose test output

### Test Timeouts

Default timeouts are configured for reliability:
- Individual test timeout: 30 seconds
- Binary build timeout: 60 seconds
- Server response timeout: 10 seconds

### Cleanup

Tests automatically clean up:
- Built binaries
- Temporary directories
- Test servers
- Background processes

## Troubleshooting

### Common Issues

**Binary build failures:**
- Ensure all dependencies are installed: `go mod download`
- Verify Go version compatibility
- Check build environment and permissions

**Test server connection issues:**
- Tests use ephemeral ports automatically assigned by the OS
- Firewall or security software may interfere
- Check system resource availability

**Test timeouts:**
- Adjust timeout values in test configuration
- Check system load and available resources
- Consider running with `-short` flag to skip long tests

**Race condition failures:**
- Run with `-race` flag to identify data races
- Check concurrent access to shared resources
- Verify proper synchronization in test code

### Getting Help

For test-related issues:
1. Run tests with `-v` flag for detailed output
2. Check test logs for specific error messages
3. Verify test environment setup and prerequisites
4. Review individual test function documentation

## Contributing to Tests

When adding new tests:

1. **Integration Tests**: Add to `test/integration/cli_test.go`
   - Test new CLI flags or functionality
   - Follow existing naming conventions
   - Include both success and error cases

2. **E2E Tests**: Add to `test/e2e/e2e_test.go`
   - Test complete user workflows
   - Include performance and load testing aspects
   - Test realistic usage scenarios

3. **Shared Utilities**: Extend `test/shared/testutils.go`
   - Add reusable test helpers
   - Create new test server configurations
   - Provide common assertion functions

### Test Guidelines

- Use descriptive test names that explain the scenario
- Include both positive and negative test cases
- Clean up resources in defer statements
- Use test helpers to reduce code duplication
- Add appropriate timeout handling
- Document complex test scenarios

### Code Coverage

Generate coverage reports:

```bash
# Generate coverage profile
go test -coverprofile=coverage.out ./test/...

# View coverage report
go tool cover -html=coverage.out

# View coverage summary
go tool cover -func=coverage.out
```

Target coverage goals:
- Integration tests: >90% CLI code coverage
- E2E tests: >95% workflow coverage
- Combined: >85% total project coverage
