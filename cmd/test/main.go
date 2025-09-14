package main

import (
	"log"
	"time"

	"github.com/superterran/adobe-digest/internal/adobe"
	"github.com/superterran/adobe-digest/internal/content"
	"github.com/superterran/adobe-digest/internal/feeds"
)

func main() {
	log.Println("Testing Adobe Security Bulletins scraper with known bulletin...")

	// Initialize HTTP client with conservative settings
	client := adobe.NewClient(
		"Mozilla/5.0 (compatible; AdobeDigest-SecurityScraper/1.0; +https://adobedigest.com)",
		45*time.Second,
		5*time.Second,
	)
	defer client.Close()

	// Initialize parser
	parser := adobe.NewParser(client)

	// Test with the known bulletin from the README
	testURL := "https://helpx.adobe.com/security/products/magento/apsb25-88.html"
	log.Printf("Testing with bulletin: %s", testURL)

	// Parse the bulletin
	bulletin, err := parser.ParseBulletin(testURL)
	if err != nil {
		log.Fatalf("Error parsing bulletin: %v", err)
	}

	log.Printf("Successfully parsed bulletin:")
	log.Printf("  ID: %s", bulletin.ID)
	log.Printf("  Title: %s", bulletin.Title)
	log.Printf("  Product: %s", bulletin.Product)
	log.Printf("  Priority: %d (%s)", bulletin.Priority, bulletin.GetPriorityLabel())
	log.Printf("  CVEs: %v", bulletin.GetCVEs())
	log.Printf("  Max CVSS: %.1f", bulletin.GetMaxCVSS())
	log.Printf("  Vulnerabilities: %d", len(bulletin.Vulnerabilities))
	log.Printf("  Affected products: %d", len(bulletin.Affected))
	log.Printf("  Solutions: %d", len(bulletin.Solutions))

	// Set some default values if missing
	if bulletin.PublishedAt.IsZero() {
		bulletin.PublishedAt = time.Date(2025, 9, 9, 0, 0, 0, 0, time.UTC)
	}
	if bulletin.UpdatedAt.IsZero() {
		bulletin.UpdatedAt = bulletin.PublishedAt
	}
	if bulletin.Product == "" {
		bulletin.Product = "Adobe Commerce"
	}
	bulletin.PriorityLabel = bulletin.GetPriorityLabel()

	// Test content generation
	log.Println("Testing content generation...")
	contentGen, err := content.NewGenerator("content")
	if err != nil {
		log.Fatalf("Error initializing content generator: %v", err)
	}

	if err := contentGen.GenerateBulletin(bulletin); err != nil {
		log.Fatalf("Error generating bulletin content: %v", err)
	}

	log.Println("✅ Successfully generated bulletin content")

	// Test RSS feed generation
	log.Println("Testing RSS feed generation...")
	feedGen := feeds.NewGenerator(
		"https://adobedigest.com",
		"https://adobedigest.com",
		"static/feeds/test-bulletins.xml",
		10,
	)

	bulletins := []adobe.SecurityBulletin{*bulletin}
	if err := feedGen.GenerateRSSFeed(bulletins, "Test Adobe Security Bulletins", "Test feed"); err != nil {
		log.Fatalf("Error generating RSS feed: %v", err)
	}

	log.Println("✅ Successfully generated RSS feed")
	log.Println("Test completed successfully!")
}
