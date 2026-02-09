# ============================================
# Build stage
# ============================================
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# Copy dependency files
COPY go.mod go.sum ./
COPY vendor/ ./vendor/

# Copy source code
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY main.go ./

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -trimpath \
    -o /usr/local/bin/notion-as-mcp \
    .

# ============================================
# Runtime stage
# ============================================
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata bash python3

# Create cache directory
RUN mkdir -p /home/notion/.cache/notion-as-mcp && \
    chmod 755 /home/notion/.cache/notion-as-mcp

# Copy binary from builder
COPY --from=builder /usr/local/bin/notion-as-mcp /usr/local/bin/notion-as-mcp

# Copy .env.example as reference
COPY .env.example /etc/notion-as-mcp/.env.example

# Set working directory
WORKDIR /home/notion

# Create non-root user
RUN addgroup -g 1000 notion && \
    adduser -D -u 1000 -G notion notion && \
    chown -R notion:notion /home/notion

# Switch to non-root user
USER notion

# Expose SSE port
EXPOSE 3100

# Expose configuration via environment variables
ENV NOTION_API_KEY=""
ENV NOTION_DATABASE_ID=""
ENV NOTION_TYPE_FIELD="Type"
ENV CACHE_TTL="5m"
ENV CACHE_DIR="/home/notion/.cache/notion-as-mcp"
ENV LOG_LEVEL="info"
ENV EXEC_TIMEOUT="30s"
ENV EXEC_LANGUAGES="bash,python,js"
ENV POLL_INTERVAL="60s"
ENV REFRESH_ON_START="true"
ENV SERVER_HOST="0.0.0.0"
ENV SERVER_PORT="3100"
ENV TRANSPORT_TYPE="streamable"

# Set entrypoint
ENTRYPOINT ["notion-as-mcp"]
CMD ["serve"]
