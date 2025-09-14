#!/bin/bash

# Adobe Security Bulletins Scraper Runner
# This script runs the scraper with proper error handling and logging

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" &> /dev/null && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_ROOT"

echo "Adobe Security Bulletins Scraper"
echo "=================================="
echo "Starting at: $(date)"
echo "Working directory: $(pwd)"
echo ""

# Check if Go is available
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed or not in PATH"
    exit 1
fi

# Check if config file exists
if [ ! -f "configs/scraper.yaml" ]; then
    echo "Error: Configuration file not found at configs/scraper.yaml"
    exit 1
fi

# Create necessary directories
mkdir -p content/bulletins
mkdir -p static/feeds

# Run the scraper
echo "Running scraper..."
if go run cmd/scraper/main.go; then
    echo ""
    echo "✅ Scraper completed successfully!"
    
    # Show what was generated
    echo ""
    echo "Generated files:"
    find content/bulletins -name "*.md" -newer .scraper-cache.json 2>/dev/null | head -10 || echo "  No new content files"
    find static/feeds -name "*.xml" 2>/dev/null | head -5 || echo "  No RSS files found"
    
    if [ -f ".scraper-cache.json" ]; then
        echo ""
        echo "Cache statistics:"
        if command -v jq &> /dev/null; then
            jq -r '"  Total bulletins: " + (.bulletins | keys | length | tostring)' .scraper-cache.json 2>/dev/null || echo "  Cache file exists but cannot parse"
        else
            echo "  Cache file updated: $(stat -c %y .scraper-cache.json 2>/dev/null || stat -f %Sm .scraper-cache.json 2>/dev/null || echo "unknown")"
        fi
    fi
else
    echo ""
    echo "❌ Scraper failed!"
    exit 1
fi

echo ""
echo "Completed at: $(date)"
