package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
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

func main() {
	fmt.Println("🕷️  Adobe Security Bulletins Manual Parser")
	fmt.Println("📝 Paste the bulletin data from fetch_webpage and press Ctrl+D when done:")

	// Read from stdin
	scanner := bufio.NewScanner(os.Stdin)
	var content strings.Builder

	for scanner.Scan() {
		content.WriteString(scanner.Text())
		content.WriteString("\n")
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading input: %v", err)
	}

	text := content.String()
	fmt.Printf("📄 Processing %d characters of input...\n", len(text))

	// Parse bulletins from the text
	bulletins := extractBulletinsFromText(text)

	fmt.Printf("✅ Found %d bulletins\n", len(bulletins))

	if len(bulletins) == 0 {
		fmt.Println("❌ No bulletins found in input")
		return
	}

	// Load existing database
	dbPath := "data/security-bulletins.json"
	db := loadDatabase(dbPath)

	existingAPSBs := make(map[string]bool)
	for _, bulletin := range db.Bulletins {
		existingAPSBs[bulletin.APSB] = true
	}

	// Filter new bulletins
	var newBulletins []SecurityBulletin
	for _, bulletin := range bulletins {
		if !existingAPSBs[bulletin.APSB] {
			newBulletins = append(newBulletins, bulletin)
		}
	}

	if len(newBulletins) == 0 {
		fmt.Printf("✅ No new bulletins found (database has %d)\n", len(db.Bulletins))
		return
	}

	fmt.Printf("📥 Adding %d new bulletins to database\n", len(newBulletins))

	// Add new bulletins to the front
	db.Bulletins = append(newBulletins, db.Bulletins...)
	db.LastUpdated = time.Now()

	// Save database
	if err := saveDatabase(dbPath, db); err != nil {
		log.Fatalf("Failed to save database: %v", err)
	}

	fmt.Printf("✅ Database now contains %d bulletins\n", len(db.Bulletins))

	// Show sample of new bulletins
	fmt.Println("\n📋 New bulletins added:")
	for i, bulletin := range newBulletins {
		if i >= 20 { // Show first 20
			fmt.Printf("   ... and %d more\n", len(newBulletins)-20)
			break
		}
		fmt.Printf("  • %s: %s\n", bulletin.APSB, strings.TrimPrefix(bulletin.Title, bulletin.APSB+": "))
	}

	fmt.Println("\n🔄 Run 'go run cmd/content-generator/main.go generate' to update site")
}

func extractBulletinsFromText(text string) []SecurityBulletin {
	var bulletins []SecurityBulletin
	seen := make(map[string]bool)

	// Look for patterns like:
	// "APSB25-85 : Security update available for Adobe Acrobat Reader  | 09/09/2025 | 09/09/2025 |"
	// "| APSB25-85 : Security update available for Adobe Acrobat Reader | 09/09/2025 | 09/09/2025 |"

	lines := strings.Split(text, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines or lines that don't contain APSB
		if line == "" || !strings.Contains(line, "APSB") {
			continue
		}

		// Match various patterns
		patterns := []string{
			// Table format: | APSB25-85 : Security update available for Adobe Acrobat Reader  | 09/09/2025 | 09/09/2025 |
			`\|\s*(APSB\d{2}-\d{2,3})\s*[:\-]\s*([^|]+?)\s*\|\s*(\d{1,2}/\d{1,2}/\d{4})\s*\|`,
			// Simple format: APSB25-85 : Security update available for Adobe Acrobat Reader
			`(APSB\d{2}-\d{2,3})\s*[:\-]\s*(.+?)(?:\s+(\d{1,2}/\d{1,2}/\d{4}))?$`,
		}

		for _, pattern := range patterns {
			re := regexp.MustCompile(pattern)
			matches := re.FindStringSubmatch(line)

			if len(matches) >= 3 {
				apsbID := strings.TrimSpace(matches[1])
				title := strings.TrimSpace(matches[2])

				// Skip duplicates
				if seen[apsbID] {
					continue
				}
				seen[apsbID] = true

				// Clean up title
				title = strings.ReplaceAll(title, "\n", " ")
				title = strings.ReplaceAll(title, "\t", " ")
				title = regexp.MustCompile(`\s+`).ReplaceAllString(title, " ")
				title = strings.TrimSpace(title)

				// Skip if title is too short or doesn't look right
				if len(title) < 10 {
					continue
				}

				// Parse date if available
				var date time.Time
				if len(matches) > 3 && matches[3] != "" {
					if parsed, err := time.Parse("1/2/2006", matches[3]); err == nil {
						date = parsed
					} else if parsed, err := time.Parse("01/02/2006", matches[3]); err == nil {
						date = parsed
					}
				}

				// If no date, estimate from APSB ID
				if date.IsZero() {
					date = estimateDateFromAPSB(apsbID)
				}

				// Generate URL
				url := generateBulletinURL(apsbID, title)

				// Infer products
				products := inferProducts(title)

				// Infer severity
				severity := inferSeverity(title)

				bulletin := SecurityBulletin{
					APSB:        apsbID,
					Title:       fmt.Sprintf("%s: %s", apsbID, title),
					Description: "Adobe has released security updates. More details in the security bulletin.",
					URL:         url,
					Date:        date,
					Products:    products,
					Severity:    severity,
				}

				bulletins = append(bulletins, bulletin)
				fmt.Printf("  📄 %s: %s\n", apsbID, title)
				break // Found a match, no need to try other patterns
			}
		}
	}

	return bulletins
}

// ... (same helper functions as before)

func estimateDateFromAPSB(apsbID string) time.Time {
	re := regexp.MustCompile(`APSB(\d{2})-(\d{2,3})`)
	matches := re.FindStringSubmatch(apsbID)

	if len(matches) >= 3 {
		year := 2000 + parseInt(matches[1])
		bulletinNum := parseInt(matches[2])
		month := (bulletinNum / 8) + 1
		if month > 12 {
			month = 12
		}
		if month < 1 {
			month = 1
		}

		return time.Date(year, time.Month(month), 15, 0, 0, 0, 0, time.UTC)
	}

	return time.Now()
}

func parseInt(s string) int {
	var result int
	for _, r := range s {
		if r >= '0' && r <= '9' {
			result = result*10 + int(r-'0')
		}
	}
	return result
}

func generateBulletinURL(apsbID, title string) string {
	productPath := inferProductPath(title)
	return fmt.Sprintf("https://helpx.adobe.com/security/products/%s/%s.html", productPath, strings.ToLower(apsbID))
}

func inferProductPath(text string) string {
	text = strings.ToLower(text)

	if strings.Contains(text, "acrobat") || strings.Contains(text, "reader") {
		return "acrobat"
	} else if strings.Contains(text, "photoshop") {
		return "photoshop"
	} else if strings.Contains(text, "after effects") {
		return "after-effects"
	} else if strings.Contains(text, "illustrator") {
		return "illustrator"
	} else if strings.Contains(text, "premiere") {
		return "premiere"
	} else if strings.Contains(text, "lightroom") {
		return "lightroom"
	} else if strings.Contains(text, "indesign") {
		return "indesign"
	} else if strings.Contains(text, "dreamweaver") {
		return "dreamweaver"
	} else if strings.Contains(text, "animate") {
		return "animate"
	} else if strings.Contains(text, "experience manager") || strings.Contains(text, "aem") {
		return "experience-manager"
	} else if strings.Contains(text, "commerce") || strings.Contains(text, "magento") {
		return "commerce"
	} else if strings.Contains(text, "coldfusion") {
		return "coldfusion"
	} else if strings.Contains(text, "substance") {
		return "substance"
	} else if strings.Contains(text, "bridge") {
		return "bridge"
	} else if strings.Contains(text, "audition") {
		return "audition"
	} else if strings.Contains(text, "dimension") {
		return "dimension"
	} else if strings.Contains(text, "framemaker") {
		return "framemaker"
	} else if strings.Contains(text, "connect") {
		return "connect"
	}

	return "other"
}

func inferProducts(text string) []string {
	text = strings.ToLower(text)
	var products []string

	productMap := map[string][]string{
		"acrobat":            {"Adobe Acrobat", "Adobe Acrobat Reader"},
		"reader":             {"Adobe Acrobat Reader"},
		"photoshop":          {"Adobe Photoshop"},
		"illustrator":        {"Adobe Illustrator"},
		"indesign":           {"Adobe InDesign"},
		"after effects":      {"Adobe After Effects"},
		"premiere pro":       {"Adobe Premiere Pro"},
		"premiere":           {"Adobe Premiere Pro"},
		"lightroom":          {"Adobe Lightroom"},
		"bridge":             {"Adobe Bridge"},
		"dreamweaver":        {"Adobe Dreamweaver"},
		"animate":            {"Adobe Animate"},
		"audition":           {"Adobe Audition"},
		"dimension":          {"Adobe Dimension"},
		"experience manager": {"Adobe Experience Manager"},
		"aem":                {"Adobe Experience Manager"},
		"commerce":           {"Adobe Commerce"},
		"magento":            {"Adobe Commerce", "Magento Open Source"},
		"coldfusion":         {"Adobe ColdFusion"},
		"campaign":           {"Adobe Campaign"},
		"substance":          {"Adobe Substance 3D"},
		"framemaker":         {"Adobe FrameMaker"},
		"connect":            {"Adobe Connect"},
		"media encoder":      {"Adobe Media Encoder"},
	}

	for keyword, productList := range productMap {
		if strings.Contains(text, keyword) {
			products = append(products, productList...)
		}
	}

	if len(products) == 0 {
		products = []string{"Adobe Product"}
	}

	// Remove duplicates
	seen := make(map[string]bool)
	var unique []string
	for _, product := range products {
		if !seen[product] {
			seen[product] = true
			unique = append(unique, product)
		}
	}

	return unique
}

func inferSeverity(text string) string {
	text = strings.ToLower(text)

	if strings.Contains(text, "critical") {
		return "Critical"
	} else if strings.Contains(text, "important") {
		return "Important"
	} else if strings.Contains(text, "moderate") {
		return "Moderate"
	}

	return "Important" // Default
}

func loadDatabase(dbPath string) *BulletinDatabase {
	data, err := os.ReadFile(dbPath)
	if err != nil {
		return &BulletinDatabase{
			LastUpdated: time.Now(),
			Bulletins:   []SecurityBulletin{},
		}
	}

	var db BulletinDatabase
	if err := json.Unmarshal(data, &db); err != nil {
		log.Printf("Warning: Could not parse existing database: %v", err)
		return &BulletinDatabase{
			LastUpdated: time.Now(),
			Bulletins:   []SecurityBulletin{},
		}
	}

	return &db
}

func saveDatabase(dbPath string, db *BulletinDatabase) error {
	data, err := json.MarshalIndent(db, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling database: %w", err)
	}

	return os.WriteFile(dbPath, data, 0644)
}
