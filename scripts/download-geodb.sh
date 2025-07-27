#!/bin/bash

set -e

# Configuration
DATA_DIR="./data"
DB_NAME="GeoLite2-Country"
BACKUP_DIR="${DATA_DIR}/backup"

# Create directories
mkdir -p "$DATA_DIR" "$BACKUP_DIR"

# Function to download sample database for testing
download_sample() {
    echo "Downloading sample GeoIP database for testing..."
    
    # Create a sample database file (placeholder)
    touch "${DATA_DIR}/${DB_NAME}.mmdb"
    
    echo "Sample database created at ${DATA_DIR}/${DB_NAME}.mmdb"
    echo "Note: This is a placeholder. For production use, download the real MaxMind database."
}

# Function to download real MaxMind database
download_maxmind() {
    local license_key=$1
    
    if [ -z "$license_key" ]; then
        echo "Error: MaxMind license key is required for production database"
        echo "Usage: $0 [LICENSE_KEY]"
        echo "Get your license key from: https://www.maxmind.com/en/my_license_key"
        return 1
    fi
    
    echo "Downloading MaxMind GeoLite2 Country database..."
    
    # Backup existing database
    if [ -f "${DATA_DIR}/${DB_NAME}.mmdb" ]; then
        echo "Backing up existing database..."
        cp "${DATA_DIR}/${DB_NAME}.mmdb" "${BACKUP_DIR}/${DB_NAME}-$(date +%Y%m%d-%H%M%S).mmdb"
    fi
    
    # Download the database
    local url="https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-Country&license_key=${license_key}&suffix=tar.gz"
    local temp_file="${DATA_DIR}/${DB_NAME}.tar.gz"
    
    curl -L -o "$temp_file" "$url"
    
    # Extract the database
    tar -xzf "$temp_file" -C "$DATA_DIR" --strip-components=1 --wildcards "*/${DB_NAME}.mmdb"
    
    # Clean up
    rm -f "$temp_file"
    
    echo "Database downloaded successfully to ${DATA_DIR}/${DB_NAME}.mmdb"
}

# Main logic
if [ $# -eq 0 ]; then
    download_sample
else
    download_maxmind "$1"
fi