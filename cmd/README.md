# Adobe Digest Command Line Tools

This directory contains various command-line tools for managing Adobe security bulletins and generating the Hugo site.

## Available Tools

### 🤖 adobe-scraper
**Unified scraper with multiple modes**
```bash
go run cmd/adobe-scraper/main.go <command>
```

**Commands:**
- `auto` - Try automated scraping (may not work due to JavaScript)
- `manual` - Parse bulletin data from stdin (recommended)
- `import <file>` - Import bulletins from JSON file
- `test` - Test connection to Adobe's page

**Examples:**
```bash
# Test connection
go run cmd/adobe-scraper/main.go test

# Manual input (most reliable)
echo "| APSB25-85 : Title | 09/14/2025 | 09/14/2025 |" | go run cmd/adobe-scraper/main.go manual

# Try automated (likely to fail)
go run cmd/adobe-scraper/main.go auto

# Import from file
go run cmd/adobe-scraper/main.go import bulletins.json
```

### 🏗️ content-generator
**Hugo site content generator**
```bash
go run cmd/content-generator/main.go generate
```
- Generates all Hugo content from bulletins database
- Creates product pages, bulletin pages, RSS feeds
- Builds interactive homepage with filtering
- Essential for updating the site after adding bulletins

### 📦 bulk-importer
**Import multiple bulletins from JSON**
```bash
go run cmd/bulk-importer/main.go
```
- Imports bulletins from external JSON files
- Useful for large data migrations
- Maintains database integrity

## Recommended Workflow

For **manual updates** (most reliable):
1. Visit https://helpx.adobe.com/security/security-bulletin.html
2. Copy recent bulletin table data
3. Run: `echo "table data" | go run cmd/adobe-scraper/main.go manual`
4. Run: `go run cmd/content-generator/main.go generate`
5. Commit and push changes

For **automated updates** (GitHub Actions):
- The `adobe-scraper auto` runs every 6 hours via GitHub Actions
- Falls back to manual approach if automated scraping fails
- Automatically generates content and deploys

**Note**: Adobe's page uses JavaScript to load bulletin data dynamically, so automated scraping typically doesn't work. The manual approach is much more reliable.

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

