# Adobe Digest Command Line Tools

This directory contains various command-line tools for managing Adobe security bulletins and generating the Hugo site.

## Available Tools

### 🤖 auto-scraper
**Automated bulletin scraper for GitHub Actions**
```bash
go run cmd/auto-scraper/main.go
```
- Fetches Adobe's security bulletin page automatically
- Parses bulletin data from HTML content  
- Updates the database with new bulletins
- Designed for GitHub Actions automation
- Handles duplicate detection and date estimation

### 📝 manual-parser  
**Interactive parser for manual data input**
```bash
echo "bulletin data" | go run cmd/manual-parser/main.go
```
- Processes bulletin data from stdin (paste and Ctrl+D)
- Supports table format: `| APSB25-85 : Title | Date | Date |`
- Best for adding new bulletins when automated scraping fails
- Integrates with existing database seamlessly

### 🏗️ content-generator
**Hugo site content generator**
```bash
go run cmd/content-generator/main.go generate
```
- Generates all Hugo content from bulletins database
- Creates product pages, bulletin pages, RSS feeds
- Builds interactive homepage with filtering
- Essential for updating the site after adding bulletins

### 🔍 debug-scraper
**Debug tool for analyzing Adobe's page structure**
```bash
go run cmd/debug-scraper/main.go
```
- Downloads and analyzes Adobe's security page
- Shows content structure and parsing issues
- Helpful for troubleshooting scraping problems

### 📦 bulk-importer
**Import multiple bulletins from JSON**
```bash
go run cmd/bulk-importer/main.go
```
- Imports bulletins from external JSON files
- Useful for large data migrations
- Maintains database integrity

### 🕷️ scraper (legacy)
**Original comprehensive scraper**
```bash
go run cmd/scraper/main.go
```
- Complex HTML parsing approach
- May not work reliably due to Adobe's dynamic content
- Kept for reference

### 🎯 simple-scraper (legacy)  
**Simplified regex-based scraper**
```bash
go run cmd/simple-scraper/main.go
```
- Regex pattern matching approach
- Limited effectiveness with Adobe's current page structure
- Fallback option

## Recommended Workflow

For **manual updates** (most reliable):
1. Visit https://helpx.adobe.com/security/security-bulletin.html
2. Copy recent bulletin table data
3. Run: `echo "table data" | go run cmd/manual-parser/main.go`
4. Run: `go run cmd/content-generator/main.go generate`
5. Commit and push changes

For **automated updates** (GitHub Actions):
- The `auto-scraper` runs every 6 hours via GitHub Actions
- Falls back to manual parser if needed
- Automatically generates content and deploys

## Database Structure

All tools work with `data/security-bulletins.json`:
```json
{
  "last_updated": "2025-09-14T...",
  "bulletins": [
    {
      "apsb": "APSB25-85",
      "title": "APSB25-85: Security update available for Adobe Acrobat Reader",
      "description": "Adobe has released security updates...",
      "url": "https://helpx.adobe.com/security/products/acrobat/apsb25-85.html",
      "date": "2025-09-09T00:00:00Z",
      "products": ["Adobe Acrobat Reader"],
      "severity": "Important"
    }
  ]
}
```

