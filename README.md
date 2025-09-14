# Adobe Security Digest

A **clean, reliable approach** to Adobe Security Bulletins - manually curated database that generates Hugo content, RSS feeds, and a comprehensive security tracking website.

## 🎯 Why This Approach?

Adobe's Akamai CDN aggressively blocks automated scrapers, making traditional scraping unreliable. Our manual curation approach provides:

- ✅ **100% Reliable** - No CDN blocking or timeout issues
- ✅ **Rich Content** - Full Hugo pages for each bulletin  
- ✅ **Multiple Formats** - Website, RSS feeds, JSON data
- ✅ **Easy Updates** - Simple JSON input, automated generation
- ✅ **GitHub Actions** - Automated workflows for content management

## 🚀 Quick Start

### Local Development

```bash
# Clone and setup
git clone https://github.com/superterran/adobe-digest.git
cd adobe-digest
go mod download

# Generate all content from database
go run cmd/content-generator/main.go generate

# Build and serve Hugo site  
hugo server --port 1314
```

### Adding New Bulletins

Via GitHub Actions (Recommended):
1. Go to **Actions** → **Update Adobe Security Content**
2. Click **Run workflow**
3. Paste bulletin JSON:

```json
{
  "apsb": "APSB24-XX",
  "title": "Security update available for Adobe Commerce",
  "description": "Adobe has released security updates resolving vulnerabilities.",
  "url": "https://helpx.adobe.com/security/products/magento/apsb24-xx.html", 
  "date": "2024-12-01T00:00:00Z",
  "products": ["Adobe Commerce", "Magento Open Source"],
  "severity": "Critical"
}
```

## 📊 Generated Content

The content generator creates:

### Hugo Pages
- **Individual Bulletins** (`/bulletins/apsb24-xx/`) - Detailed pages for each bulletin
- **Product Pages** (`/products/adobe-commerce/`) - Bulletins grouped by product  
- **Index Pages** (`/bulletins/`) - Overview and navigation
- **Homepage Data** - Dynamic statistics and recent bulletins

### RSS Feeds  
- **Main Feed** (`/adobe-security.xml`) - All bulletins in RSS 2.0 format

### Data Files
- **Homepage JSON** (`data/homepage.json`) - Statistics for Hugo templates
- **Database** (`data/security-bulletins.json`) - Master bulletin database

## 🏗️ Architecture

```
├── cmd/content-generator/     # Main content generation tool
├── data/
│   ├── security-bulletins.json   # Master database (manual)  
│   └── homepage.json             # Generated homepage data
├── content/
│   ├── bulletins/                # Generated bulletin pages
│   └── products/                 # Generated product pages  
├── layouts/                      # Hugo templates
└── public/
    └── adobe-security.xml        # Generated RSS feed
```

## 🛠️ Commands

```bash
# Generate all content from database
go run cmd/content-generator/main.go generate

# Add new bulletin and regenerate 
go run cmd/content-generator/main.go add '{"apsb":"APSB24-XX",...}'

# Clean all generated content
go run cmd/content-generator/main.go clean

# Build Hugo site
hugo --minify

# Serve locally
hugo server --port 1314
```

## 🔄 Automation

### GitHub Actions Workflow
- **Manual Trigger**: Add bulletins via web interface
- **Scheduled**: Weekly content regeneration  
- **Automated**: Content generation → Hugo build → GitHub Pages deploy

### Content Generation Flow
1. **Database Update** - Add bulletin to JSON database
2. **Hugo Content** - Generate markdown files for bulletins/products
3. **RSS Generation** - Create RSS feed from database
4. **Homepage Data** - Generate statistics and recent bulletins
5. **Site Build** - Hugo builds static site with all content

## 📈 Current Status

- **Total Bulletins**: 5 (Adobe Commerce focused)
- **Products Tracked**: Adobe Commerce, Magento Open Source
- **Content Types**: Individual bulletins, product summaries, RSS feeds
- **Last Updated**: Automatically tracked in database

## 🔍 Bulletin Sources

Monitor these sources for new Adobe security bulletins:
- [Adobe Security Advisories](https://helpx.adobe.com/security.html)
- [Adobe Commerce Security](https://helpx.adobe.com/security/products/magento.html)
- [Adobe PSIRT Blog](https://blog.adobe.com/en/publish/2024/03/12/psirt-adobe-product-security-incident-response-team)

## 🚢 Deployment

The site auto-deploys to GitHub Pages at [adobedigest.com](https://adobedigest.com) when:
- New bulletins are added via GitHub Actions
- Content is manually regenerated  
- Changes are pushed to main branch

## 📝 Content Structure

Each bulletin generates:
- **Detailed page** with full advisory information
- **Structured frontmatter** for Hugo processing  
- **External links** to official Adobe advisories
- **Product categorization** for easy navigation
- **RSS feed entries** with full descriptions

---

**Note**: This is an **unofficial** security bulletin aggregation service. Always refer to [Adobe's official PSIRT advisories](https://helpx.adobe.com/security.html) for authoritative security information and remediation guidance.
