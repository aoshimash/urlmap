#!/bin/bash

# Build script for release binaries
# Usage: ./scripts/build-release.sh [version]

set -e

# Default version if not provided
VERSION=${1:-"dev"}
COMMIT=$(git rev-parse --short HEAD)
BUILD_DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Building crawld release binaries...${NC}"
echo "Version: $VERSION"
echo "Commit: $COMMIT"
echo "Build Date: $BUILD_DATE"
echo

# Create bin directory
mkdir -p bin

# Set ldflags for version info
LDFLAGS="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${BUILD_DATE}"

# Function to build for a specific platform
build_platform() {
    local os=$1
    local arch=$2
    local ext=$3
    local filename="crawld-${os}-${arch}${ext}"

    echo -e "${YELLOW}Building for ${os}/${arch}...${NC}"
    GOOS=$os GOARCH=$arch go build -ldflags="${LDFLAGS}" -o "bin/${filename}" ./cmd/crawld

    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ ${filename} built successfully${NC}"
    else
        echo -e "${RED}✗ Failed to build ${filename}${NC}"
        exit 1
    fi
}

# Build for all platforms
build_platform "linux" "amd64" ""
build_platform "linux" "arm64" ""
build_platform "darwin" "amd64" ""
build_platform "darwin" "arm64" ""
build_platform "windows" "amd64" ".exe"

echo
echo -e "${GREEN}Creating compressed archives...${NC}"

cd bin

# Create tar.gz archives for Unix systems
tar -czf crawld-linux-amd64.tar.gz crawld-linux-amd64
tar -czf crawld-linux-arm64.tar.gz crawld-linux-arm64
tar -czf crawld-darwin-amd64.tar.gz crawld-darwin-amd64
tar -czf crawld-darwin-arm64.tar.gz crawld-darwin-arm64

# Create zip archive for Windows
zip -q crawld-windows-amd64.zip crawld-windows-amd64.exe

echo -e "${GREEN}Generating checksums...${NC}"
sha256sum *.tar.gz *.zip > checksums.txt

echo
echo -e "${GREEN}Build complete! Files created:${NC}"
ls -la crawld-*
echo
echo -e "${GREEN}Checksums:${NC}"
cat checksums.txt

echo
echo -e "${GREEN}Test version output:${NC}"
./crawld-linux-amd64 version || echo -e "${RED}Failed to run version command${NC}"

cd ..
