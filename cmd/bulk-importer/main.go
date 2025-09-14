package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
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

// BulkImportData represents the structure for importing multiple bulletins
type BulkImportData struct {
	Bulletins []SecurityBulletin `json:"bulletins"`
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Adobe Security Bulk Importer")
		fmt.Println("Usage:")
		fmt.Println("  go run cmd/bulk-importer/main.go [database-file] [import-file]")
		fmt.Println()
		fmt.Println("Import file should be JSON with structure:")
		fmt.Println(`{
  "bulletins": [
    {
      "apsb": "APSB25-85",
      "title": "Security update available for Adobe Acrobat Reader",
      "description": "Adobe has released security updates...",
      "url": "https://helpx.adobe.com/security/products/acrobat/apsb25-85.html",
      "date": "2025-09-09T00:00:00Z",
      "products": ["Adobe Acrobat Reader DC"],
      "severity": "Critical"
    }
  ]
}`)
		os.Exit(1)
	}

	databaseFile := os.Args[1]
	importFile := os.Args[2]

	// Load existing database
	db, err := loadDatabase(databaseFile)
	if err != nil {
		log.Fatalf("Failed to load database: %v", err)
	}

	// Load import data
	importData, err := loadImportFile(importFile)
	if err != nil {
		log.Fatalf("Failed to load import file: %v", err)
	}

	fmt.Printf("📊 Current database has %d bulletins\n", len(db.Bulletins))
	fmt.Printf("📥 Import file contains %d bulletins\n", len(importData.Bulletins))

	// Create a map of existing bulletins for quick lookup
	existingAPSBs := make(map[string]bool)
	for _, bulletin := range db.Bulletins {
		existingAPSBs[bulletin.APSB] = true
	}

	// Add new bulletins
	var newBulletins []SecurityBulletin
	var duplicates []string

	for _, bulletin := range importData.Bulletins {
		if existingAPSBs[bulletin.APSB] {
			duplicates = append(duplicates, bulletin.APSB)
			continue
		}

		// Validate required fields
		if bulletin.APSB == "" || bulletin.Title == "" || bulletin.URL == "" {
			log.Printf("⚠️  Skipping invalid bulletin: %+v", bulletin)
			continue
		}

		// Set default values if missing
		if bulletin.Severity == "" {
			bulletin.Severity = "Important"
		}
		if bulletin.Date.IsZero() {
			bulletin.Date = time.Now()
		}

		newBulletins = append(newBulletins, bulletin)
	}

	if len(duplicates) > 0 {
		fmt.Printf("⚠️  Found %d duplicate bulletins (skipped): %s\n", len(duplicates), strings.Join(duplicates, ", "))
	}

	if len(newBulletins) == 0 {
		fmt.Println("✅ No new bulletins to import")
		return
	}

	// Add new bulletins to the front of the list (newest first)
	db.Bulletins = append(newBulletins, db.Bulletins...)
	db.LastUpdated = time.Now()

	// Save updated database
	if err := saveDatabase(databaseFile, db); err != nil {
		log.Fatalf("Failed to save database: %v", err)
	}

	fmt.Printf("✅ Successfully imported %d new bulletins\n", len(newBulletins))
	fmt.Printf("📊 Database now contains %d bulletins total\n", len(db.Bulletins))

	// Print summary of new bulletins
	fmt.Println("\n📋 New bulletins added:")
	for _, bulletin := range newBulletins {
		fmt.Printf("  • %s: %s (%s)\n", bulletin.APSB, bulletin.Title, bulletin.Severity)
	}

	fmt.Println("\n🔄 Run 'go run cmd/content-generator/main.go generate' to update the site content")
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

func loadImportFile(importFile string) (*BulkImportData, error) {
	data, err := os.ReadFile(importFile)
	if err != nil {
		return nil, fmt.Errorf("reading import file: %w", err)
	}

	var importData BulkImportData
	if err := json.Unmarshal(data, &importData); err != nil {
		return nil, fmt.Errorf("unmarshaling import file: %w", err)
	}

	return &importData, nil
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
