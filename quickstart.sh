#!/bin/bash

set -e

echo "=== IP Geo Service Quick Start ==="
echo

# Check for Go installation
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed. Please install Go 1.21 or higher."
    exit 1
fi

echo "✓ Go is installed: $(go version)"

# Create data directory
mkdir -p data

# Download dependencies
echo
echo "Downloading Go dependencies..."
go mod download

# Build the service
echo
echo "Building the service..."
make build

# Check if GeoIP database exists
if [ ! -f "data/GeoLite2-Country.mmdb" ]; then
    echo
    echo "GeoIP database not found. Running in demo mode..."
    echo "To use real data, download the MaxMind GeoLite2 database."
fi

# Run the service
echo
echo "Starting the service..."
echo "Web UI will be available at: http://localhost:8080"
echo "Press Ctrl+C to stop"
echo

./geo-service -demo-mode=true