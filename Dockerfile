# Build stage
FROM golang:1.24-alpine AS builder

LABEL org.opencontainers.image.source="https://github.com/aoshimash/urlmap"
LABEL org.opencontainers.image.description="A fast and efficient web crawler CLI tool for discovering URLs within a domain"
LABEL org.opencontainers.image.licenses="MIT"

# Install git for go modules
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build \
    -a -installsuffix cgo \
    -ldflags '-extldflags "-static"' \
    -o urlmap ./cmd/urlmap

# Final stage
FROM alpine:latest

LABEL org.opencontainers.image.source="https://github.com/aoshimash/urlmap"
LABEL org.opencontainers.image.description="A fast and efficient web crawler CLI tool for discovering URLs within a domain"
LABEL org.opencontainers.image.licenses="MIT"

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/urlmap /usr/local/bin/urlmap

# Make sure binary is executable
RUN chmod +x /usr/local/bin/urlmap

# Use the binary as entrypoint
ENTRYPOINT ["/usr/local/bin/urlmap"]
