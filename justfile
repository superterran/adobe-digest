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