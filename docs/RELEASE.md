# Release Process

This document describes the automated release process for urlmap and how to create new releases.

## 🚀 Automated Release Process

urlmap uses an automated release process that:
- Builds multi-platform binaries on git tag pushes
- Creates compressed archives with checksums
- Generates automatic release notes
- Publishes releases to GitHub Releases

## Supported Platforms

The release process builds binaries for the following platforms:

| Platform | Architecture | Binary Name | Archive Format |
|----------|-------------|-------------|----------------|
| Linux | amd64 | `urlmap-linux-amd64` | tar.gz |
| Linux | arm64 | `urlmap-linux-arm64` | tar.gz |
| macOS | amd64 (Intel) | `urlmap-darwin-amd64` | tar.gz |
| macOS | arm64 (Apple Silicon) | `urlmap-darwin-arm64` | tar.gz |
| Windows | amd64 | `urlmap-windows-amd64.exe` | zip |

## Creating a Release

### 1. Prepare the Release

1. Ensure all changes are merged to the `main` branch
2. Update any necessary documentation
3. Run tests to ensure everything works:
   ```bash
   go test ./...
   ```

### 2. Create and Push a Git Tag

Releases are triggered by pushing git tags that start with `v`. Use semantic versioning:

```bash
# Create an annotated tag with release notes
git tag -a v1.0.0 -m "Release v1.0.0

- Initial stable release
- Added web crawling functionality
- Multi-platform binary support"

# Push the tag to trigger the release
git push origin v1.0.0
```

### 3. Monitor the Release Process

1. Go to the [Actions tab](https://github.com/aoshimash/urlmap/actions) to monitor the build
2. The release workflow will:
   - Build binaries for all platforms
   - Create compressed archives
   - Generate checksums
   - Create a GitHub Release with assets
3. Check the [Releases page](https://github.com/aoshimash/urlmap/releases) when complete

## Version Information

Each binary includes build-time version information accessible via the `version` command:

```bash
$ urlmap version
urlmap version v1.0.0
commit: abc1234
built: 2024-01-01T12:00:00Z
```

The version information includes:
- **Version**: The git tag (e.g., `v1.0.0`)
- **Commit**: Short git commit hash
- **Built**: Build timestamp in UTC

## Local Development Builds

For local development and testing, use the build script:

```bash
# Build with default "dev" version
./scripts/build-release.sh

# Build with specific version
./scripts/build-release.sh v1.0.0-dev
```

This creates the same multi-platform binaries locally in the `bin/` directory.

## Release Notes

Release notes are automatically generated from:
1. Git tag annotations (if the tag is annotated)
2. Commit messages since the last tag (if no tag annotation)

For better release notes, use annotated tags with detailed messages:

```bash
git tag -a v1.1.0 -m "Release v1.1.0

### New Features
- Added concurrent crawling support
- Improved error handling

### Bug Fixes
- Fixed URL parsing edge cases
- Resolved memory leak in parser

### Breaking Changes
- Changed default crawl depth from unlimited to 3"
```

## Binary Verification

Each release includes a `checksums.txt` file for verifying binary integrity:

```bash
# Download and verify a binary
curl -L -O https://github.com/aoshimash/urlmap/releases/download/v1.0.0/urlmap-linux-amd64.tar.gz
curl -L -O https://github.com/aoshimash/urlmap/releases/download/v1.0.0/checksums.txt

# Verify checksum
sha256sum -c checksums.txt --ignore-missing
```

## Troubleshooting

### Release Build Fails

1. Check the [Actions tab](https://github.com/aoshimash/urlmap/actions) for error details
2. Common issues:
   - Build failures due to code issues
   - Missing dependencies
   - Test failures
3. Fix issues and create a new tag if needed

### Binary Issues

1. Test binaries locally using `./scripts/build-release.sh`
2. Verify cross-compilation works for all platforms
3. Check that version information is embedded correctly

### Permission Issues

The release workflow requires:
- `contents: write` permission to create releases
- `GITHUB_TOKEN` with appropriate permissions

## Semantic Versioning

urlmap follows [Semantic Versioning](https://semver.org/):

- **MAJOR** version (X.0.0): Incompatible API changes
- **MINOR** version (0.X.0): New features (backward compatible)
- **PATCH** version (0.0.X): Bug fixes (backward compatible)
- **Pre-release** (1.0.0-alpha.1): Development versions

Examples:
- `v1.0.0` - First stable release
- `v1.1.0`
