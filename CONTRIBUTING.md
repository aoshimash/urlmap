# Contributing to urlmap

Thank you for your interest in contributing to urlmap! We welcome contributions from everyone, whether you're fixing bugs, adding features, improving documentation, or just asking questions.

## 📋 Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Testing](#testing)
- [Documentation](#documentation)
- [Submitting Changes](#submitting-changes)
- [Review Process](#review-process)
- [Community](#community)

## 🤝 Code of Conduct

This project adheres to a code of conduct that we expect all contributors to follow. Please be respectful and constructive in all interactions.

### Our Pledge

- Use welcoming and inclusive language
- Be respectful of differing viewpoints and experiences
- Gracefully accept constructive criticism
- Focus on what is best for the community
- Show empathy towards other community members

## 🚀 Getting Started

### Prerequisites

- **Go**: Version 1.21 or higher
- **Git**: For version control
- **golangci-lint**: For code linting (optional but recommended)

### First-time Contributors

If you're new to contributing to open source projects:

1. Look for issues labeled `good first issue` or `help wanted`
2. Start with documentation improvements or small bug fixes
3. Ask questions! We're here to help

## 🛠 Development Setup

### 1. Fork and Clone

```bash
# Fork the repository on GitHub, then clone your fork
git clone https://github.com/YOUR_USERNAME/urlmap.git
cd urlmap

# Add the upstream repository
git remote add upstream https://github.com/aoshimash/urlmap.git
```

### 2. Install Dependencies

```bash
# Download Go modules
go mod download

# Install development tools (optional)
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### 3. Verify Setup

```bash
# Build the project
go build -o urlmap ./cmd/urlmap

# Run tests
go test ./...

# Run linting
go vet ./...
golangci-lint run  # if installed
```

## 📝 Making Changes

### Branch Naming

Use descriptive branch names with prefixes:

- `feat/` - New features
- `fix/` - Bug fixes
- `docs/` - Documentation changes
- `refactor/` - Code refactoring
- `test/` - Adding tests
- `ci/` - CI/CD changes

Examples:
- `feat/add-json-output`
- `fix/handle-empty-urls`
- `docs/update-readme`

### Coding Standards

Follow these guidelines to maintain code quality:

#### Go Code Style

- Follow `gofmt` formatting
- Use meaningful variable and function names
- Write clear comments for public functions
- Keep functions small and focused
- Handle errors appropriately

#### Example:

```go
// ExtractLinks extracts all valid HTTP/HTTPS links from the given HTML content.
// It returns a slice of normalized URLs and any parsing errors encountered.
func ExtractLinks(html string, baseURL *url.URL) ([]string, error) {
    if html == "" {
        return nil, fmt.Errorf("empty HTML content")
    }

    // Implementation...
}
```

#### Project Structure

Follow the established project structure:

```
urlmap/
├── cmd/urlmap/          # CLI application entry point
├── internal/            # Private application code
│   ├── crawler/         # Core crawling logic
│   ├── parser/          # HTML parsing
│   └── ...
├── pkg/                 # Public libraries
├── test/                # Test files and fixtures
└── docs/                # Documentation
```

### Commit Messages

Write clear, descriptive commit messages:

```
<type>: <description>

<body>

<footer>
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `refactor`: Code refactoring
- `test`: Adding/updating tests
- `ci`: CI/CD changes

Examples:
```
feat: add JSON output format support

Add --output-format flag to support JSON output in addition to
the default plain text format. This enables better integration
with other tools and scripts.

Closes #123
```

```
fix: handle malformed URLs gracefully

Previously, the crawler would panic when encountering malformed
URLs. Now it logs a warning and continues processing.

Fixes #456
```

## 🧪 Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests verbosely
go test -v ./...

# Run specific package tests
go test ./internal/crawler/
```

### Writing Tests

- Write tests for all new functionality
- Aim for good test coverage (we target 80%+)
- Use table-driven tests where appropriate
- Include both positive and negative test cases

#### Example Test:

```go
func TestNormalizeURL(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {
            name:     "valid HTTP URL",
            input:    "http://example.com/path",
            expected: "http://example.com/path",
            wantErr:  false,
        },
        {
            name:     "invalid URL",
            input:    "not-a-url",
            expected: "",
            wantErr:  true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := NormalizeURL(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("NormalizeURL() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if result != tt.expected {
                t.Errorf("NormalizeURL() = %v, want %v", result, tt.expected)
            }
        })
    }
}
```

### Integration Tests

Run end-to-end tests to ensure the full application works:

```bash
# Run integration tests
go test ./test/integration/

# Run E2E tests
go test ./test/e2e/
```

## 📚 Documentation

### Types of Documentation

1. **Code Comments**: Document public functions and complex logic
2. **README Updates**: Keep installation and usage instructions current
3. **Architecture Docs**: Document significant design decisions
4. **API Documentation**: Use godoc-style comments

### Documentation Standards

- Write clear, concise documentation
- Include examples where helpful
- Keep documentation in sync with code changes
- Use proper markdown formatting

### Building Documentation

```bash
# Generate Go documentation
go doc ./...

# Serve documentation locally
godoc -http=:6060
```

## 📨 Submitting Changes

### Before Submitting

1. **Test your changes**:
   ```bash
   go test ./...
   go vet ./...
   ```

2. **Format your code**:
   ```bash
   go fmt ./...
   ```

3. **Run linting** (if available):
   ```bash
   golangci-lint run
   ```

### Pull Request Process

1. **Update your branch**:
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. **Push your changes**:
   ```bash
   git push origin your-branch-name
   ```

3. **Create a Pull Request**:
   - Use the GitHub web interface
   - Fill out the PR template completely
   - Link to any related issues

### Pull Request Template

When creating a PR, include:

```markdown
## Description
Brief description of the changes.

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Documentation update
- [ ] Refactoring

## Testing
- [ ] Tests pass locally
- [ ] New tests added (if applicable)
- [ ] Manual testing performed

## Checklist
- [ ] Code follows project style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] No breaking changes (or clearly documented)
```

## 🔍 Review Process

### What to Expect

1. **Automated Checks**: CI/CD will run tests and linting
2. **Maintainer Review**: Code review by project maintainers
3. **Feedback**: Constructive feedback and suggestions
4. **Iteration**: You may need to make changes based on feedback

### Review Criteria

Reviews focus on:
- **Correctness**: Does the code work as intended?
- **Style**: Does it follow project conventions?
- **Performance**: Are there any performance concerns?
- **Security**: Are there any security implications?
- **Maintainability**: Is the code easy to understand and modify?

## 💬 Community

### Getting Help

- **GitHub Issues**: For bug reports and feature requests
- **GitHub Discussions**: For questions and general discussion
- **Code Reviews**: For feedback on your contributions

### Communication Guidelines

- Be respectful and professional
- Provide context and examples when asking questions
- Search existing issues before creating new ones
- Use clear, descriptive titles for issues and PRs

## 🐛 Reporting Bugs

### Before Reporting

1. Search existing issues to avoid duplicates
2. Test with the latest version
3. Gather relevant system information

### Bug Report Template

```markdown
## Bug Description
Clear description of the issue.

## Steps to Reproduce
1. Run command: `urlmap ...`
2. Observe error: ...

## Expected Behavior
What should happen.

## Actual Behavior
What actually happens.

## Environment
- OS: [e.g., macOS 12.0]
- Go version: [e.g., 1.21.0]
- urlmap version: [e.g., v1.0.0]

## Additional Context
Any other relevant information.
```

## 💡 Feature Requests

### Before Requesting

- Check if the feature already exists
- Search existing feature requests
- Consider if it fits the project's scope

### Feature Request Template

```markdown
## Feature Description
Clear description of the proposed feature.

## Use Case
Why is this feature needed? What problem does it solve?

## Proposed Implementation
How do you envision this working?

## Alternatives Considered
What other approaches have you considered?
```

## 🏆 Recognition

Contributors who make significant contributions will be:
- Added to the project's contributors list
- Recognized in release notes
- Eligible for collaborator status (for regular contributors)

## 📄 License

By contributing to urlmap, you agree that your contributions will be licensed under the MIT License.

## 🙏 Acknowledgments

Thank you for contributing to urlmap! 🎉
