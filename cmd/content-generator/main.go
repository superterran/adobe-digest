package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/feeds"
)

// SecurityBulletin represents a security bulletin entry
type SecurityBulletin struct {
	APSB        string    `json:"apsb"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	URL         string    `json:"url"`
	Date        time.Time `json:"date"`
	Products    []string  `json:"products"`
	Severity    string    `json:"severity"`
}

// BulletinDatabase holds all security bulletins
type BulletinDatabase struct {
	LastUpdated time.Time          `json:"last_updated"`
	Bulletins   []SecurityBulletin `json:"bulletins"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Adobe Security Content Generator")
		fmt.Println("Usage:")
		fmt.Println("  go run cmd/content-generator/main.go generate  - Generate all content from database")
		fmt.Println("  go run cmd/content-generator/main.go add [json] - Add bulletin and regenerate")
		fmt.Println("  go run cmd/content-generator/main.go clean     - Clean old generated content")
		os.Exit(1)
	}

	command := os.Args[1]
	dataFile := "data/security-bulletins.json"

	switch command {
	case "generate":
		generateAllContent(dataFile)
	case "add":
		if len(os.Args) < 3 {
			log.Fatal("Please provide bulletin JSON data")
		}
		addBulletinAndRegenerate(dataFile, os.Args[2])
	case "clean":
		cleanGeneratedContent()
	default:
		log.Fatalf("Unknown command: %s", command)
	}
}

func generateAllContent(dataFile string) {
	fmt.Println("🏗️  Generating all content from security bulletins database...")

	// Load database
	db, err := loadDatabase(dataFile)
	if err != nil {
		log.Fatalf("Failed to load database: %v", err)
	}

	fmt.Printf("📊 Loaded %d bulletins from database\n", len(db.Bulletins))

	// Generate Hugo content files
	if err := generateHugoContent(db); err != nil {
		log.Fatalf("Failed to generate Hugo content: %v", err)
	}

	// Generate RSS feed
	if err := generateRSSFeed(db); err != nil {
		log.Fatalf("Failed to generate RSS feed: %v", err)
	}

	// Generate homepage data
	if err := generateHomepageData(db); err != nil {
		log.Fatalf("Failed to generate homepage data: %v", err)
	}

	fmt.Println("✅ All content generated successfully!")
}

func addBulletinAndRegenerate(dataFile, jsonData string) {
	// Load existing database
	db, err := loadDatabase(dataFile)
	if err != nil {
		log.Fatalf("Failed to load database: %v", err)
	}

	// Parse new bulletin
	var bulletin SecurityBulletin
	if err := json.Unmarshal([]byte(jsonData), &bulletin); err != nil {
		log.Fatalf("Failed to parse bulletin JSON: %v", err)
	}

	// Add to database (at the front for newest first)
	db.Bulletins = append([]SecurityBulletin{bulletin}, db.Bulletins...)
	db.LastUpdated = time.Now()

	// Save updated database
	if err := saveDatabase(dataFile, db); err != nil {
		log.Fatalf("Failed to save database: %v", err)
	}

	fmt.Printf("✅ Added bulletin %s: %s\n", bulletin.APSB, bulletin.Title)

	// Regenerate all content
	generateAllContent(dataFile)
}

func cleanGeneratedContent() {
	fmt.Println("🧹 Cleaning generated content...")

	// Remove generated bulletin files
	bulletinDirs := []string{
		"content/bulletins",
		"content/products",
	}

	for _, dir := range bulletinDirs {
		if err := os.RemoveAll(dir); err != nil {
			log.Printf("Warning: failed to remove %s: %v", dir, err)
		} else {
			fmt.Printf("🗑️  Removed %s\n", dir)
		}
	}

	// Remove generated feeds
	feedFiles := []string{
		"public/adobe-security.xml",
		"static/feeds",
	}

	for _, file := range feedFiles {
		if err := os.RemoveAll(file); err != nil {
			log.Printf("Warning: failed to remove %s: %v", file, err)
		} else {
			fmt.Printf("🗑️  Removed %s\n", file)
		}
	}

	fmt.Println("✅ Cleanup completed!")
}

func loadDatabase(dataFile string) (*BulletinDatabase, error) {
	data, err := os.ReadFile(dataFile)
	if err != nil {
		return nil, fmt.Errorf("reading database file: %w", err)
	}

	var db BulletinDatabase
	if err := json.Unmarshal(data, &db); err != nil {
		return nil, fmt.Errorf("unmarshaling database: %w", err)
	}

	return &db, nil
}

func saveDatabase(dataFile string, db *BulletinDatabase) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(dataFile), 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	data, err := json.MarshalIndent(db, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling database: %w", err)
	}

	return os.WriteFile(dataFile, data, 0644)
}

func generateHugoContent(db *BulletinDatabase) error {
	fmt.Println("📝 Generating Hugo content files...")

	// Create bulletins directory
	bulletinDir := "content/bulletins"
	if err := os.MkdirAll(bulletinDir, 0755); err != nil {
		return fmt.Errorf("creating bulletins directory: %w", err)
	}

	// Group bulletins by product for organization
	productGroups := make(map[string][]SecurityBulletin)
	for _, bulletin := range db.Bulletins {
		for _, product := range bulletin.Products {
			productGroups[product] = append(productGroups[product], bulletin)
		}
	}

	// Generate individual bulletin pages
	for _, bulletin := range db.Bulletins {
		if err := generateBulletinPage(bulletin, bulletinDir); err != nil {
			return fmt.Errorf("generating bulletin page for %s: %w", bulletin.APSB, err)
		}
	}

	// Generate product index pages
	if err := generateProductPages(productGroups); err != nil {
		return fmt.Errorf("generating product pages: %w", err)
	}

	// Generate main bulletins index
	if err := generateBulletinsIndex(db.Bulletins); err != nil {
		return fmt.Errorf("generating bulletins index: %w", err)
	}

	fmt.Printf("✅ Generated Hugo content for %d bulletins\n", len(db.Bulletins))
	return nil
}

func generateBulletinPage(bulletin SecurityBulletin, bulletinDir string) error {
	// Create filename from APSB ID
	filename := fmt.Sprintf("%s.md", strings.ToLower(bulletin.APSB))
	filepath := filepath.Join(bulletinDir, filename)

	// Clean title to avoid duplicating APSB ID
	cleanTitle := bulletin.Title
	if strings.HasPrefix(cleanTitle, bulletin.APSB+": ") {
		cleanTitle = strings.TrimPrefix(cleanTitle, bulletin.APSB+": ")
	}

	// Generate frontmatter and content
	content := fmt.Sprintf(`---
title: "%s: %s"
description: "%s"
date: %s
apsb_id: "%s"
severity: "%s"
products: %s
external_url: "%s"
draft: false
---

## %s: %s

**Severity**: %s  
**Date**: %s  
**Products**: %s  
**APSB ID**: %s

### Description

%s

### Affected Products

%s

### Official Advisory

For complete details, patches, and remediation steps, please refer to the official Adobe security advisory:

[🔗 View Official Advisory](%s)

---

*This information is sourced from Adobe's official Product Security Incident Response Team (PSIRT) advisories. Always refer to the official Adobe advisory for authoritative information and remediation guidance.*
`,
		bulletin.APSB, cleanTitle,
		bulletin.Description,
		bulletin.Date.Format("2006-01-02T15:04:05Z07:00"),
		bulletin.APSB,
		bulletin.Severity,
		formatProductsForYAML(bulletin.Products),
		bulletin.URL,
		bulletin.APSB, cleanTitle,
		bulletin.Severity,
		bulletin.Date.Format("January 2, 2006"),
		strings.Join(bulletin.Products, ", "),
		bulletin.APSB,
		bulletin.Description,
		formatProductsList(bulletin.Products),
		bulletin.URL,
	)

	return os.WriteFile(filepath, []byte(content), 0644)
}

func generateProductPages(productGroups map[string][]SecurityBulletin) error {
	fmt.Println("📦 Generating product pages...")

	productDir := "content/products"
	if err := os.MkdirAll(productDir, 0755); err != nil {
		return fmt.Errorf("creating products directory: %w", err)
	}

	for product, bulletins := range productGroups {
		// Create safe filename
		filename := strings.ToLower(strings.ReplaceAll(product, " ", "-"))
		filename = strings.ReplaceAll(filename, "/", "-")
		filepath := filepath.Join(productDir, fmt.Sprintf("%s.md", filename))

		// Generate product page content
		content := fmt.Sprintf(`---
title: "%s Security Bulletins"
description: "Security bulletins and advisories for %s"
product: "%s"
bulletin_count: %d
draft: false
---

# %s Security Bulletins

This page contains all security bulletins for **%s**.

## Statistics

- **Total Bulletins**: %d
- **Critical**: %d
- **Important**: %d
- **Moderate**: %d

## Recent Bulletins

`,
			product, product, product, len(bulletins),
			product, product, len(bulletins),
			countBySeverity(bulletins, "Critical"),
			countBySeverity(bulletins, "Important"),
			countBySeverity(bulletins, "Moderate"),
		)

		// Add bulletin list
		for _, bulletin := range bulletins {
			content += fmt.Sprintf(`### [%s](%s) - %s

**Date**: %s | **Severity**: %s

%s

---

`, bulletin.APSB, bulletin.URL, bulletin.Title,
				bulletin.Date.Format("January 2, 2006"),
				bulletin.Severity, bulletin.Description)
		}

		if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
			return fmt.Errorf("writing product page for %s: %w", product, err)
		}
	}

	return nil
}

func generateBulletinsIndex(bulletins []SecurityBulletin) error {
	fmt.Println("📑 Generating bulletins index...")

	filepath := "content/bulletins/_index.md"

	content := fmt.Sprintf(`---
title: "Security Bulletins"
description: "Adobe Security Bulletins and Advisories"
weight: 1
---

# Adobe Security Bulletins

This section contains all tracked Adobe security bulletins, organized by product and severity.

## Quick Stats

- **Total Bulletins**: %d
- **Critical**: %d
- **Important**: %d
- **Moderate**: %d

## Recent Bulletins

`,
		len(bulletins),
		countBySeverity(bulletins, "Critical"),
		countBySeverity(bulletins, "Important"),
		countBySeverity(bulletins, "Moderate"),
	)

	// Add recent bulletins (last 10)
	recentBulletins := bulletins
	if len(recentBulletins) > 10 {
		recentBulletins = recentBulletins[:10]
	}

	for _, bulletin := range recentBulletins {
		content += fmt.Sprintf(`- [**%s**](%s) - %s (%s) - %s
`,
			bulletin.APSB,
			bulletin.URL,
			bulletin.Title,
			bulletin.Severity,
			bulletin.Date.Format("Jan 2, 2006"),
		)
	}

	return os.WriteFile(filepath, []byte(content), 0644)
}

func generateRSSFeed(db *BulletinDatabase) error {
	fmt.Println("📡 Generating RSS feed...")

	// Create RSS feed
	feed := &feeds.Feed{
		Title:       "Adobe Security Bulletins",
		Link:        &feeds.Link{Href: "https://adobedigest.com/"},
		Description: "Latest security bulletins for Adobe products - manually curated and verified",
		Author:      &feeds.Author{Name: "Adobe Security Digest"},
		Created:     db.LastUpdated,
		Copyright:   "Adobe Inc. - Republished for community awareness",
	}

	// Add bulletins as RSS items (limit to 25 most recent)
	recentBulletins := db.Bulletins
	if len(recentBulletins) > 25 {
		recentBulletins = recentBulletins[:25]
	}

	for _, bulletin := range recentBulletins {
		item := &feeds.Item{
			Title: fmt.Sprintf("%s: %s", bulletin.APSB, bulletin.Title),
			Link:  &feeds.Link{Href: bulletin.URL},
			Description: fmt.Sprintf("%s\n\nProducts: %s\nSeverity: %s\n\nView full advisory: %s",
				bulletin.Description,
				strings.Join(bulletin.Products, ", "),
				bulletin.Severity,
				bulletin.URL),
			Author:  &feeds.Author{Name: "Adobe Security Team"},
			Created: bulletin.Date,
			Id:      bulletin.APSB,
		}
		feed.Items = append(feed.Items, item)
	}

	// Generate RSS XML
	rss, err := feed.ToRss()
	if err != nil {
		return fmt.Errorf("generating RSS: %w", err)
	}

	// Save RSS file
	rssFile := "public/adobe-security.xml"
	if err := os.MkdirAll(filepath.Dir(rssFile), 0755); err != nil {
		return fmt.Errorf("creating RSS directory: %w", err)
	}

	if err := os.WriteFile(rssFile, []byte(rss), 0644); err != nil {
		return fmt.Errorf("writing RSS file: %w", err)
	}

	fmt.Printf("✅ Generated RSS feed with %d items\n", len(feed.Items))
	return nil
}

func generateHomepageData(db *BulletinDatabase) error {
	fmt.Println("🏠 Generating homepage data...")

	// Calculate statistics
	severityStats := make(map[string]int)
	for _, bulletin := range db.Bulletins {
		severityStats[bulletin.Severity]++
	}

	// Get recent bulletins (last 10)
	recentBulletins := db.Bulletins
	if len(recentBulletins) > 10 {
		recentBulletins = recentBulletins[:10]
	}

	// Create homepage data
	homePageData := map[string]interface{}{
		"last_updated":     db.LastUpdated.Format("January 2, 2006 at 3:04 PM"),
		"recent_bulletins": recentBulletins,
		"total_bulletins":  len(db.Bulletins),
		"critical_count":   severityStats["Critical"],
		"important_count":  severityStats["Important"],
		"severity_stats":   severityStats,
	}

	// Save homepage data
	outputFile := "data/homepage.json"
	if err := os.MkdirAll(filepath.Dir(outputFile), 0755); err != nil {
		return fmt.Errorf("creating data directory: %w", err)
	}

	data, err := json.MarshalIndent(homePageData, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling homepage data: %w", err)
	}

	if err := os.WriteFile(outputFile, data, 0644); err != nil {
		return fmt.Errorf("writing homepage data: %w", err)
	}

	// Also copy to assets for Hugo
	assetsDir := "assets/data"
	if err := os.MkdirAll(assetsDir, 0755); err != nil {
		return fmt.Errorf("creating assets directory: %w", err)
	}

	assetsFile := filepath.Join(assetsDir, "homepage.json")
	if err := os.WriteFile(assetsFile, data, 0644); err != nil {
		return fmt.Errorf("writing assets homepage data: %w", err)
	}

	fmt.Println("✅ Generated homepage data")
	return nil
}

// Helper functions
func formatProductsForYAML(products []string) string {
	result := "["
	for i, product := range products {
		if i > 0 {
			result += ", "
		}
		result += fmt.Sprintf(`"%s"`, product)
	}
	result += "]"
	return result
}

func formatProductsList(products []string) string {
	result := ""
	for _, product := range products {
		result += fmt.Sprintf("- %s\n", product)
	}
	return result
}

func countBySeverity(bulletins []SecurityBulletin, severity string) int {
	count := 0
	for _, bulletin := range bulletins {
		if bulletin.Severity == severity {
			count++
		}
	}
	return count
}
