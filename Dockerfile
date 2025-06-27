# Build stage - Use Go official image matching go.mod version
FROM golang:1.24 AS builder

# Add metadata
LABEL org.opencontainers.image.source="https://github.com/aoshimash/crawld"
LABEL org.opencontainers.image.description="Web crawler daemon"
LABEL org.opencontainers.image.licenses="MIT"

# Install ca-certificates for HTTPS requests
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -ldflags="-s -w -X main.version=docker -X main.commit=$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown') -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    -o crawld ./cmd/crawld

# Runtime stage - Use distroless
FROM gcr.io/distroless/static-debian11:nonroot

# Add metadata
LABEL org.opencontainers.image.source="https://github.com/aoshimash/crawld"
LABEL org.opencontainers.image.description="Web crawler daemon"
LABEL org.opencontainers.image.licenses="MIT"

# Copy CA certificates from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the binary from builder stage
COPY --from=builder /app/crawld /usr/local/bin/crawld

# Use non-root user (already set by distroless:nonroot)
USER nonroot:nonroot

# Set entrypoint
ENTRYPOINT ["/usr/local/bin/crawld"]

# Default help command
CMD ["--help"]
