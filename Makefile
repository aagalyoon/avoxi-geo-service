.PHONY: help build run test clean docker-build docker-run download-db proto lint

# Variables
BINARY_NAME=geo-service
DOCKER_IMAGE=geo-service
VERSION?=latest
GOOS?=linux
GOARCH?=amd64

# Default target
help:
	@echo "Available targets:"
	@echo "  build        - Build the binary"
	@echo "  run          - Run the application locally"
	@echo "  test         - Run tests"
	@echo "  lint         - Run linters"
	@echo "  clean        - Clean build artifacts"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run Docker container"
	@echo "  download-db  - Download GeoIP database"
	@echo "  proto        - Generate protobuf code"

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags="-w -s" -o $(BINARY_NAME) cmd/server/main.go

# Run the application
run:
	@echo "Running $(BINARY_NAME)..."
	go run cmd/server/main.go

# Run tests
test:
	@echo "Running tests..."
	go test -v -cover ./...

# Run tests with race detector
test-race:
	@echo "Running tests with race detector..."
	go test -v -race ./...

# Run linters
lint:
	@echo "Running linters..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install from https://golangci-lint.run/usage/install/"; \
	fi

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	rm -rf dist/

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(VERSION) .

# Run Docker container
docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 -p 9090:9090 $(DOCKER_IMAGE):$(VERSION)

# Download GeoIP database
download-db:
	@echo "Downloading GeoIP database..."
	@./scripts/download-geodb.sh

# Generate protobuf code
proto:
	@echo "Generating protobuf code..."
	@if command -v protoc > /dev/null; then \
		protoc --go_out=. --go-grpc_out=. pkg/proto/*.proto; \
	else \
		echo "protoc not installed. Please install protocol buffers compiler."; \
	fi