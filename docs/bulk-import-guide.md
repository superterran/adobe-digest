# Adobe Security Bulletin Management

This repository contains tools for managing a comprehensive database of Adobe security bulletins.

## Quick Start for Bulk Import

Since Adobe's CDN blocks automated scraping, we use a manual approach that makes it easy to import large numbers of bulletins:

### 1. Extract Bulletin Data

1. Go to https://helpx.adobe.com/security/security-bulletin.html
2. Copy bulletin entries (they look like this):
   ```
   APSB25-85 - Security update available for Adobe Acrobat Reader - September 9, 2025
   APSB25-84 - Security update available for Adobe Photoshop - September 2, 2025
   APSB25-83 - Security update available for Adobe After Effects - August 13, 2025
   ```
3. Paste them into a text file (e.g., `bulletins-to-import.txt`)

### 2. Parse and Import

```bash
# Parse the text file into JSON format
python3 tools/parse-bulletins.py bulletins-to-import.txt import-data.json

# Import into the database (handles duplicates automatically)
go run cmd/bulk-importer/main.go data/security-bulletins.json import-data.json

# Generate the Hugo content
go run cmd/content-generator/main.go generate
```

### 3. Build and Deploy

```bash
# Build the site
hugo

# Or run locally
hugo server
```

## Tools Overview

### Bulk Importer (`cmd/bulk-importer/main.go`)
- Imports multiple bulletins from JSON files
- Automatically detects and skips duplicates
- Validates data before importing
- Maintains chronological order (newest first)

### Bulletin Parser (`tools/parse-bulletins.py`)
- Converts copy-pasted bulletin lines into JSON format
- Automatically infers product categories
- Generates proper URLs
- Handles various date formats

### Content Generator (`cmd/content-generator/main.go`)
- Generates Hugo content from the database
- Creates individual bulletin pages
- Builds category pages for each product
- Generates RSS feeds

## Database Structure

The `data/security-bulletins.json` file contains:

```json
{
  "last_updated": "2025-01-27T10:30:00Z",
  "bulletins": [
    {
      "apsb": "APSB25-85",
      "title": "APSB25-85: Security update available for Adobe Acrobat Reader",
      "description": "Adobe has released security updates...",
      "url": "https://helpx.adobe.com/security/products/acrobat/apsb25-85.html",
      "date": "2025-09-09T00:00:00Z",
      "products": ["Adobe Acrobat Reader DC"],
      "severity": "Critical"
    }
  ]
}
```

## Comprehensive Coverage

To achieve comprehensive coverage of Adobe's security bulletins:

1. **Recent Bulletins**: Check the main page regularly for new APSB releases
2. **Historical Data**: Use the product-specific pages to find older bulletins
3. **Product Categories**: Ensure coverage across all Adobe product lines:
   - Creative Cloud (Photoshop, Illustrator, After Effects, etc.)
   - Document Cloud (Acrobat, Reader)
   - Experience Cloud (Experience Manager, Campaign)
   - Commerce (Magento Commerce)
   - Developer Tools (ColdFusion, Dreamweaver)

## Automation with GitHub Actions

The repository includes GitHub Actions workflows for:
- Manual bulletin addition (single bulletins)
- Bulk import processing
- Automated content generation
- Site deployment

To add bulletins via GitHub Actions:
1. Go to the Actions tab
2. Select "Add Security Bulletin" or "Bulk Import Bulletins"
3. Provide the required information
4. The workflow will update the database and regenerate the site

## Site Features

- **Comprehensive Listing**: Homepage shows all bulletins with filtering
- **Product Categories**: Navigation organized by Adobe product lines
- **Search and Filter**: Interactive filtering by product and severity
- **RSS Feeds**: Automated RSS generation for the entire feed and per-product
- **Responsive Design**: Works on desktop and mobile devices

## Future Enhancements

- Automatic severity detection from bulletin content
- Enhanced product categorization
- Vulnerability database integration
- Email notifications for critical bulletins
- API endpoints for programmatic access
