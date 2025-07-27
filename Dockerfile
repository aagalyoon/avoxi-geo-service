# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o geo-service cmd/server/main.go

# Runtime stage
FROM alpine:3.18

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata curl bash

# Create non-root user
RUN addgroup -g 1000 -S appgroup && \
    adduser -u 1000 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/geo-service /app/geo-service

# Copy scripts
COPY --from=builder /app/scripts /app/scripts
RUN chmod +x /app/scripts/*.sh

# Copy web assets
COPY --from=builder /app/web /app/web

# Create data directory
RUN mkdir -p /app/data /app/data/backup && \
    chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Download initial GeoIP database (sample for testing)
RUN /app/scripts/download-geodb.sh

# Expose ports
EXPOSE 8080 9090

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/api/v1/health || exit 1

# Default environment variables
ENV HTTP_PORT=8080 \
    GRPC_PORT=9090 \
    GEOIP_DB_PATH=/app/data/GeoLite2-Country.mmdb \
    LOG_LEVEL=info \
    DEMO_MODE=false

# Run the application
ENTRYPOINT ["/app/geo-service"]