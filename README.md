# Adobe Digest

> **Comprehensive Security Intelligence Platform for Adobe Products**

A professional security monitoring platform that aggregates Adobe security bulletins, vulnerability data, and security analysis. Currently focused on automated Adobe security bulletin tracking with plans to expand into comprehensive security research, blog articles, and multi-source threat intelligence.

**� Live Site**: [adobedigest.com](https://adobedigest.com)

---

## Overview

Adobe Digest provides comprehensive security intelligence for Adobe products through automated monitoring and professional analysis. Currently specializing in Adobe security bulletin aggregation, the platform is designed to expand into broader security research, vulnerability analysis, and threat intelligence from multiple sources.

### ✨ **Current Features**
- **📡 Automated Adobe Bulletin Monitoring** - Multi-strategy scraping handles Adobe's dynamic content
- **🔍 Product-Specific Organization** - Security advisories organized by Adobe product  
- **📰 Comprehensive RSS Feeds** - Global and product-specific feeds (39 total)
- **🏗️ Professional Static Site** - Fast, reliable Hugo-based platform
- **⚡ Automated Deployment** - GitHub Actions with 6-hour update cycles
- **📊 Extensive Coverage** - 150+ security bulletins across 35+ Adobe products

### 🚀 **Planned Expansions**
- **📝 Security Blog Articles** - In-depth vulnerability analysis and security research
- **🔐 Multi-Source Intelligence** - Additional security feeds beyond Adobe
- **📈 Threat Analytics** - Security trend analysis and reporting
- **🔔 Advanced Alerting** - Real-time notifications for critical vulnerabilities

### 🎯 **Why Adobe Security Digest?**
- **Reliability** - Automated scraping with fallback strategies
- **Accessibility** - Clean, searchable interface with RSS feeds
- **Organization** - Bulletins grouped by product for targeted monitoring
- **Automation** - Updates every 6 hours via GitHub Actions
- **Open Source** - Transparent, community-driven development

## 🚀 Quick Start

### For Users

**RSS Feeds**: Subscribe to security updates for specific Adobe products or all products:
- **All Products**: `https://adobedigest.com/adobe-security.xml`
- **Products Overview**: `https://adobedigest.com/feeds/products.xml`
- **Specific Products**: `https://adobedigest.com/feeds/{product-name}.xml`

**Website**: Browse and search security bulletins at [adobedigest.com](https://adobedigest.com)

### For Developers

```bash
# Clone and setup
git clone https://github.com/superterran/adobe-digest.git
cd adobe-digest
go mod download

# Run automated scraper
just scrape

# Generate content and serve locally
just dev
```

**Using Just Commands:**
```bash
just scrape          # Run automated scraper
just run             # Start Hugo development server  
just dev             # Clean and start dev server
just clean-all       # Clean all generated content
```

## 🏗️ Architecture

### Multi-Strategy Automated Scraper
The `adobe-scraper` employs multiple strategies to reliably extract bulletin data:

1. **API Discovery** - Attempts to find Adobe's bulletin API endpoints
2. **Alternative URLs** - Uses Adobe's JSON-format security bulletin URLs  
3. **Enhanced HTML Parsing** - Intelligent content extraction from web pages
4. **Browser Automation** - Handles JavaScript-heavy dynamic content

### Content Generation Pipeline
```
Adobe Security Page → Scraper → Database → Content Generator → Hugo Site
                                    ↓
                              RSS Feeds (39 feeds)
```

### Generated Content

**🌐 Website Content:**
- **Individual Bulletins** (`/bulletins/apsb25-xx/`) - Detailed security advisory pages
- **Product Pages** (`/products/adobe-photoshop/`) - Product-specific bulletin collections
- **Homepage** - Statistics, recent bulletins, and navigation
- **Search & Browse** - Organized access to all security advisories

**📡 RSS Feeds:**
- **Global Feed** (`/adobe-security.xml`) - 25 most recent bulletins
- **Products Feed** (`/feeds/products.xml`) - 50 recent bulletins organized by product
- **38 Product-Specific Feeds** (`/feeds/{product}.xml`) - Dedicated feeds per Adobe product

**📊 Data:**
- **Master Database** (`data/security-bulletins.json`) - Structured bulletin data
- **Automated Caching** - Intelligent scraping with duplicate prevention

## 🔧 Development

### Project Structure
```
├── cmd/
│   ├── adobe-scraper/          # Multi-strategy bulletin scraper
│   ├── content-generator/      # Hugo content and RSS generation
│   └── bulk-importer/          # Bulk data import utilities
├── data/
│   └── security-bulletins.json # Master bulletin database
├── content/                    # Generated Hugo content
├── layouts/                    # Hugo templates and overrides
├── public/                     # Generated static site and RSS feeds
└── .github/workflows/          # Automated CI/CD pipelines
```

### Commands

**Just Commands** (Recommended):
```bash
just scrape          # Run automated scraper
just run             # Start development server
just dev             # Clean and start development server  
just clean-all       # Remove all generated content
```

**Direct Go Commands**:
```bash
# Run scraper (multiple modes)
go run cmd/adobe-scraper/main.go auto
go run cmd/adobe-scraper/main.go manual
go run cmd/adobe-scraper/main.go test

# Generate all content
go run cmd/content-generator/main.go generate

# Build Hugo site
hugo --minify
```

## ⚙️ Automation & Deployment

### GitHub Actions Workflows

**🤖 Automated Scraping** (`scraper.yml`):
- Runs every 6 hours automatically
- Uses multi-strategy scraper to find new bulletins
- Commits and pushes changes when new bulletins are found
- Triggers deployment automatically

**🚀 Site Deployment** (`deploy.yml`):
- Triggered by content changes or manual dispatch
- Builds Hugo site with latest content and RSS feeds
- Deploys to GitHub Pages at [adobedigest.com](https://adobedigest.com)

### Content Pipeline
```
Scheduled Run → Scraper → New Bulletins → Content Generator → Site Build → Deploy
     ↓              ↓           ↓              ↓               ↓         ↓
Every 6hrs    Multi-strategy  Database    Hugo + RSS     Static Site  GitHub Pages
```

## � Current Coverage

- **📈 Total Bulletins**: 150+ security advisories
- **🏢 Products Tracked**: 35+ Adobe products (Creative Cloud, Document Cloud, Experience Cloud)
- **📡 RSS Feeds**: 39 feeds (1 global + 1 products + 37 product-specific)
- **🔄 Update Frequency**: Every 6 hours via automated scraping
- **⚡ Last Updated**: Tracked automatically in database

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details.

### Development Setup
1. Fork the repository
2. Clone your fork: `git clone https://github.com/your-username/adobe-digest.git`
3. Install dependencies: `go mod download`
4. Run tests: `go test ./...`
5. Make your changes and submit a pull request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 👨‍💻 Author & Sponsorship

**Created by**: [Doug Hatcher](https://doughatcher.com)  
**Sponsored by**: [Blue Acorn iCi](https://blueacornici.com)

## ⚠️ Disclaimer

This is an **unofficial** security bulletin aggregation service. Always refer to [Adobe's official PSIRT advisories](https://helpx.adobe.com/security.html) for authoritative security information and remediation guidance.

Adobe and all Adobe product names are trademarks of Adobe Inc. This project is not affiliated with or endorsed by Adobe Inc.
