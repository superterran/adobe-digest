package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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
	fmt.Println("🕷️  Adobe Security Bulletins Scraper v2")
	fmt.Println("📡 Fetching https://helpx.adobe.com/security/security-bulletin.html")

	// Get the page content
	content, err := fetchPageContent("https://helpx.adobe.com/security/security-bulletin.html")
	if err != nil {
		log.Fatalf("Failed to fetch page: %v", err)
	}

	fmt.Println("🔍 Parsing bulletin entries...")

	// Extract all APSB entries from the content
	bulletins := extractBulletins(content)

	fmt.Printf("✅ Found %d bulletins\n", len(bulletins))

	if len(bulletins) == 0 {
		fmt.Println("❌ No bulletins found - check if page structure changed")
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
	fmt.Println("\n📋 Sample of new bulletins:")
	for i, bulletin := range newBulletins {
		if i >= 10 { // Show first 10
			fmt.Printf("   ... and %d more\n", len(newBulletins)-10)
			break
		}
		fmt.Printf("  • %s: %s\n", bulletin.APSB, strings.TrimPrefix(bulletin.Title, bulletin.APSB+": "))
	}

	fmt.Println("\n🔄 Run 'go run cmd/content-generator/main.go generate' to update site")
}

func fetchPageContent(url string) (string, error) {
	client := &http.Client{Timeout: 60 * time.Second}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	buf := make([]byte, 2*1024*1024) // 2MB buffer
	n, _ := resp.Body.Read(buf)

	return string(buf[:n]), nil
}

func extractBulletins(content string) []SecurityBulletin {
	var bulletins []SecurityBulletin
	seen := make(map[string]bool)

	// Regex to match APSB entries
	// Looking for patterns like: "APSB25-85 : Security update available for Adobe Acrobat Reader"
	re := regexp.MustCompile(`(APSB\d{2}-\d{2,3})\s*[:\-]\s*([^|]+?)(?:\s*\|\s*(\d{1,2}/\d{1,2}/\d{4})|$)`)

	matches := re.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) < 3 {
			continue
		}

		apsbID := strings.TrimSpace(match[1])
		title := strings.TrimSpace(match[2])

		// Skip duplicates
		if seen[apsbID] {
			continue
		}
		seen[apsbID] = true

		// Clean up the title
		title = strings.ReplaceAll(title, "\n", " ")
		title = strings.ReplaceAll(title, "\t", " ")
		title = regexp.MustCompile(`\s+`).ReplaceAllString(title, " ")
		title = strings.TrimSpace(title)

		// Skip if title is too short or doesn't look right
		if len(title) < 10 || !strings.Contains(strings.ToLower(title), "security") {
			continue
		}

		// Parse date if available
		var date time.Time
		if len(match) > 3 && match[3] != "" {
			if parsed, err := time.Parse("1/2/2006", match[3]); err == nil {
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
	}

	return bulletins
}

func estimateDateFromAPSB(apsbID string) time.Time {
	// Extract year from APSB ID (APSB24-XX means 2024)
	re := regexp.MustCompile(`APSB(\d{2})-(\d{2,3})`)
	matches := re.FindStringSubmatch(apsbID)

	if len(matches) >= 3 {
		year := 2000 + parseInt(matches[1])
		// Use bulletin number to estimate month (rough approximation)
		bulletinNum := parseInt(matches[2])
		month := (bulletinNum / 8) + 1 // Rough estimate
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
