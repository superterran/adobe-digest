package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/superterran/adobe-digest/internal/adobe"
	"github.com/superterran/adobe-digest/internal/content"
	"github.com/superterran/adobe-digest/internal/feeds"
	"github.com/superterran/adobe-digest/internal/storage"
)

func main() {
	log.Println("Starting Adobe Security Bulletins scraper...")

	// Load configuration
	config, err := loadConfig("configs/scraper.yaml")
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	// Initialize cache
	cache := storage.NewCache(config.Output.CacheFile)
	if err := cache.Load(); err != nil {
		log.Fatalf("Error loading cache: %v", err)
	}

	// Parse durations from config
	rateLimit, err := time.ParseDuration(config.Scraper.RateLimit)
	if err != nil {
		log.Fatalf("Error parsing rate limit: %v", err)
	}

	timeout, err := time.ParseDuration(config.Scraper.Timeout)
	if err != nil {
		log.Fatalf("Error parsing timeout: %v", err)
	}

	// Initialize HTTP client
	client := adobe.NewClient(config.Scraper.UserAgent, timeout, rateLimit)
	defer client.Close()

	// Initialize parser
	parser := adobe.NewParser(client)

	// Initialize content generator
	contentGen, err := content.NewGenerator(config.Output.ContentDir)
	if err != nil {
		log.Fatalf("Error initializing content generator: %v", err)
	}

	// Initialize RSS feed generator
	feedGen := feeds.NewGenerator(
		config.RSS.Link,
		config.RSS.Link,
		config.Output.RSSFile,
		config.RSS.MaxItems,
	)

	// Track new and updated bulletins
	var allBulletins []adobe.SecurityBulletin
	bulletinsByProduct := make(map[string][]adobe.SecurityBulletin)
	newBulletinCount := 0
	updatedBulletinCount := 0

	// Process each enabled product
	for _, product := range config.Products {
		if !product.Enabled {
			log.Printf("Skipping disabled product: %s", product.DisplayName)
			continue
		}

		log.Printf("Processing product: %s", product.DisplayName)
		log.Printf("Fetching bulletin list from: %s", product.URL)

		// Get bulletin summaries from product page
		summaries, err := parser.ParseProductPage(product.URL, product.DisplayName)
		if err != nil {
			log.Printf("Error parsing product page for %s: %v", product.DisplayName, err)
			continue
		}

		log.Printf("Found %d bulletins for %s", len(summaries), product.DisplayName)

		var productBulletins []adobe.SecurityBulletin

		// Process each bulletin
		for _, summary := range summaries {
			// Check if bulletin is new or updated
			if !cache.IsNewOrUpdated(summary) {
				log.Printf("Skipping unchanged bulletin: %s", summary.ID)
				continue
			}

			isNew := false
			if _, exists := cache.GetBulletin(summary.ID); !exists {
				isNew = true
				newBulletinCount++
			} else {
				updatedBulletinCount++
			}

			log.Printf("Processing %s bulletin: %s",
				map[bool]string{true: "new", false: "updated"}[isNew], summary.ID)

			// Parse full bulletin details
			bulletin, err := parser.ParseBulletin(summary.URL)
			if err != nil {
				log.Printf("Error parsing bulletin %s: %v", summary.ID, err)
				continue
			}

			// Ensure bulletin has all necessary data
			if bulletin.PublishedAt.IsZero() {
				bulletin.PublishedAt = summary.PublishedAt
			}
			if bulletin.UpdatedAt.IsZero() {
				bulletin.UpdatedAt = summary.UpdatedAt
			}
			if bulletin.Product == "" || bulletin.Product == "Adobe Product" {
				bulletin.Product = product.DisplayName
			}

			// Add to collections
			allBulletins = append(allBulletins, *bulletin)
			productBulletins = append(productBulletins, *bulletin)

			// Update cache
			cache.SetBulletin(summary)

			log.Printf("Successfully processed bulletin: %s (Priority: %d, CVEs: %d)",
				bulletin.ID, bulletin.Priority, len(bulletin.Vulnerabilities))
		}

		if len(productBulletins) > 0 {
			bulletinsByProduct[product.DisplayName] = productBulletins
		}
	}

	log.Printf("Processing complete. New: %d, Updated: %d, Total: %d",
		newBulletinCount, updatedBulletinCount, len(allBulletins))

	// Generate content if we have new or updated bulletins
	if len(allBulletins) > 0 {
		log.Println("Generating Hugo content...")

		// Generate all content (bulletins, indexes)
		if err := contentGen.GenerateAll(bulletinsByProduct, config.Products); err != nil {
			log.Fatalf("Error generating content: %v", err)
		}

		log.Println("Generating RSS feeds...")

		// Generate RSS feeds
		if err := feedGen.GenerateAllFeeds(allBulletins, config.Products); err != nil {
			log.Fatalf("Error generating RSS feeds: %v", err)
		}

		log.Printf("Successfully generated content for %d bulletins", len(allBulletins))
	} else {
		log.Println("No new or updated bulletins found")
	}

	// Update cache with last run time
	cache.SetLastRun(time.Now())

	// Clean up old bulletins (keep 2 years)
	removed := cache.RemoveOldBulletins(2 * 365 * 24 * time.Hour)
	if removed > 0 {
		log.Printf("Removed %d old bulletins from cache", removed)
	}

	// Save cache
	if err := cache.Save(); err != nil {
		log.Fatalf("Error saving cache: %v", err)
	}

	// Print statistics
	stats := cache.GetStats()
	log.Printf("Cache statistics:")
	log.Printf("  Total bulletins: %d", stats.TotalBulletins)
	log.Printf("  Recent bulletins (30 days): %d", stats.RecentBulletins)
	log.Printf("  By product:")
	for product, count := range stats.ByProduct {
		log.Printf("    %s: %d", product, count)
	}

	log.Println("Scraper completed successfully!")
}

// loadConfig loads the scraper configuration from file
func loadConfig(configPath string) (*adobe.ScraperConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var config adobe.ScraperConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}

	// Set defaults if not specified
	if config.Scraper.UserAgent == "" {
		config.Scraper.UserAgent = "AdobeDigest-SecurityScraper/1.0"
	}
	if config.Scraper.RateLimit == "" {
		config.Scraper.RateLimit = "2s"
	}
	if config.Scraper.Timeout == "" {
		config.Scraper.Timeout = "30s"
	}
	if config.Scraper.RetryAttempts == 0 {
		config.Scraper.RetryAttempts = 3
	}
	if config.RSS.MaxItems == 0 {
		config.RSS.MaxItems = 50
	}

	// Ensure output directories exist
	if config.Output.ContentDir == "" {
		config.Output.ContentDir = "content"
	}
	if config.Output.RSSFile == "" {
		config.Output.RSSFile = "static/feeds/bulletins.xml"
	}
	if config.Output.CacheFile == "" {
		config.Output.CacheFile = ".scraper-cache.json"
	}

	// Create output directories
	if err := os.MkdirAll(config.Output.ContentDir, 0755); err != nil {
		return nil, fmt.Errorf("creating content directory: %w", err)
	}

	feedDir := filepath.Dir(config.Output.RSSFile)
	if err := os.MkdirAll(feedDir, 0755); err != nil {
		return nil, fmt.Errorf("creating feed directory: %w", err)
	}

	return &config, nil
}
