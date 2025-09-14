run: ## run the dev server
  hugo serve

clean-scraped: ## wipe all scraped security bulletin content and caches
  @echo "Removing scraped security bulletin content..."
  rm -rf content/bulletins/
  rm -rf static/feeds/
  rm -f .scraper-cache.json
  @echo "Clearing Hugo caches..."
  rm -rf public/
  rm -rf resources/
  hugo mod clean
  @echo "Scraped content and caches cleaned successfully"

clean-all: ## wipe everything including build outputs and caches
  @echo "Performing complete cleanup..."
  rm -rf content/bulletins/
  rm -rf static/feeds/
  rm -rf public/
  rm -rf resources/
  rm -f .scraper-cache.json
  hugo mod clean
  go clean -cache -modcache -testcache
  @echo "Complete cleanup finished"

dev: clean-scraped run ## clean scraped content and start dev server

scrape: ## run the automated scraper to fetch new bulletins
  @echo "🤖 Running automated scraper..."
  go run cmd/adobe-scraper/main.go auto
  @echo "🏗️  Generating Hugo content..."
  go run cmd/content-generator/main.go generate
  @echo "✅ Scraping and content generation complete!"

scrape-manual: ## run manual parser (paste bulletin data and press Ctrl+D)
  @echo "📝 Manual bulletin parser - paste table format data and press Ctrl+D:"
  @echo "Example: | APSB25-XX : Security update for Adobe Product | MM/DD/YYYY | MM/DD/YYYY |"
  @echo ""
  go run cmd/adobe-scraper/main.go manual
  @echo "🏗️  Generating Hugo content..."
  go run cmd/content-generator/main.go generate
  @echo "✅ Manual parsing and content generation complete!"

scrape-test: ## test connection to Adobe's security page
  @echo "🔍 Testing Adobe security page connection..."
  go run cmd/adobe-scraper/main.go test

scrape-import: ## import bulletins from JSON file (usage: just scrape-import file.json)
  @echo "📦 Import bulletins from JSON file..."
  @echo "Usage: just scrape-import <file.json>"

update: scrape ## alias for scrape command