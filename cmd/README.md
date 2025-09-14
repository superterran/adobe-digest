# Adobe Security Digest - Command Line Tools

Professional command-line tools for automated Adobe security bulletin management and content generation.

---

## 🛠️ Available Tools

### 🤖 `adobe-scraper` - Multi-Strategy Bulletin Scraper

**Purpose**: Reliably extract Adobe security bulletins using multiple automated strategies.

```bash
go run cmd/adobe-scraper/main.go <command>
```

#### Commands

| Command | Description | Reliability | Use Case |
|---------|-------------|-------------|----------|
| `auto` | Automated multi-strategy scraping | ⭐⭐⭐⭐ | GitHub Actions, scheduled runs |
| `manual` | Interactive manual bulletin entry | ⭐⭐⭐⭐⭐ | Manual updates, testing |
| `test` | Test connection to Adobe's endpoints | ⭐⭐⭐⭐⭐ | Debugging, monitoring |
| `import <file>` | Bulk import from JSON file | ⭐⭐⭐⭐⭐ | Data migration, backups |

#### Scraping Strategies

The `auto` command employs multiple strategies in sequence:

1. **API Discovery** - Searches for Adobe's bulletin API endpoints
2. **Alternative URLs** - Uses Adobe's JSON-format security bulletin URLs
3. **Enhanced HTML Parsing** - Intelligent content extraction with headers
4. **Browser Automation** - Handles JavaScript-heavy dynamic content (planned)

#### Examples

```bash
# Test all scraping strategies
go run cmd/adobe-scraper/main.go test

# Run automated scraping (production)
go run cmd/adobe-scraper/main.go auto

# Manual bulletin entry
go run cmd/adobe-scraper/main.go manual
# Then paste: | APSB25-85 : Security update | 09/14/2025 | 09/14/2025 |

# Import from backup file
go run cmd/adobe-scraper/main.go import backup-bulletins.json
```

---

### 🏗️ `content-generator` - Hugo Site Builder

**Purpose**: Transform bulletin database into complete Hugo website with RSS feeds.

```bash
go run cmd/content-generator/main.go generate
```

#### Generated Content

- **📄 Individual Bulletin Pages** (`content/bulletins/apsb25-xx.md`)
- **🏢 Product Pages** (`content/products/adobe-photoshop.md`) 
- **📡 RSS Feeds** (39 feeds: global + products + individual products)
- **🏠 Homepage Data** (statistics and recent bulletins)

#### Features

- **Duplicate Prevention** - Intelligent APSB ID cleaning
- **Product Organization** - Automatic product categorization
- **RSS Generation** - Full RSS 2.0 compliance with rich descriptions
- **Template Integration** - Clean Hugo frontmatter generation

#### Output Structure
```
content/
├── bulletins/          # Individual security advisory pages
├── products/           # Product-specific bulletin collections
public/
├── adobe-security.xml  # Global RSS feed (25 recent)
├── feeds/
│   ├── products.xml    # All products RSS (50 recent)
│   └── adobe-*.xml     # Product-specific RSS feeds (25 each)
```

---

### 📦 `bulk-importer` - Data Import Utility

**Purpose**: Import large sets of bulletin data from external sources.

```bash
go run cmd/bulk-importer/main.go <source-file>
```

#### Features
- **Data Validation** - Ensures bulletin integrity before import
- **Duplicate Detection** - Prevents duplicate APSB entries
- **Batch Processing** - Efficient handling of large datasets
- **Database Integrity** - Maintains consistent data structure

---

## 🔄 Recommended Workflows

### Production Automated Updates
```bash
# This runs automatically every 6 hours via GitHub Actions
go run cmd/adobe-scraper/main.go auto
go run cmd/content-generator/main.go generate
# → Commit and push changes
# → Trigger deployment
```

### Development/Testing
```bash
# Test scraper connectivity
go run cmd/adobe-scraper/main.go test

# Run manual scraping for testing
go run cmd/adobe-scraper/main.go auto

# Generate content for local development
go run cmd/content-generator/main.go generate

# Serve Hugo site locally
hugo server --bind 0.0.0.0 --port 1313
```

### Manual Content Updates
```bash
# For adding specific bulletins manually
go run cmd/adobe-scraper/main.go manual
# → Enter bulletin data when prompted

# Regenerate all content
go run cmd/content-generator/main.go generate

# Test locally before committing
just dev
```

---

## 📊 Database Schema

All tools interact with the centralized bulletin database at `data/security-bulletins.json`:

```json
{
  "last_updated": "2025-09-14T19:29:40.123Z",
  "bulletins": [
    {
      "apsb": "APSB25-85",
      "title": "APSB25-85: Security update available for Adobe Acrobat Reader",
      "description": "Adobe has released security updates. More details in the security bulletin.",
      "url": "https://helpx.adobe.com/security/products/acrobat/apsb25-85.html",
      "date": "2025-09-09T00:00:00Z",
      "products": ["Adobe Acrobat", "Adobe Acrobat Reader"],
      "severity": "Important"
    }
  ]
}
```

### Data Quality Features
- **Automatic Timestamps** - `last_updated` tracks generation time
- **Consistent Formatting** - Standardized APSB IDs and product names
- **Validation** - Required fields and format checking
- **Deduplication** - Prevents duplicate entries by APSB ID

---

## 🔧 Integration with Hugo

The tools integrate seamlessly with Hugo's content structure:

### Content Generation Flow
```
Database → Content Generator → Hugo Content → Static Site
    ↓              ↓               ↓            ↓
JSON data → Markdown files → HTML pages → RSS feeds
```

### Template Integration
- **Frontmatter** - Rich YAML metadata for Hugo processing
- **Content Body** - Clean markdown with structured information
- **Cross-references** - Automatic linking between bulletins and products
- **SEO Optimization** - Proper meta tags and structured data

---

## 🚀 Production Deployment

These tools are designed for automated production use:

- **GitHub Actions** - Runs `adobe-scraper auto` every 6 hours
- **Error Handling** - Graceful failures with detailed logging
- **Incremental Updates** - Only processes new/changed bulletins
- **Zero Downtime** - Hugo builds complete site from scratch each time

### Monitoring & Debugging

```bash
# Check scraper status
go run cmd/adobe-scraper/main.go test

# Validate database integrity
go run cmd/content-generator/main.go generate --dry-run

# Monitor RSS feed quality
curl -s https://adobedigest.com/adobe-security.xml | head -20
```

---

**🏢 Enterprise Ready**: Built for reliability, scalability, and professional deployment environments.

