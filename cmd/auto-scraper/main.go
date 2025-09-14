package main

import (
	"encoding/json"
	"fmt"
	"io"
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
	fmt.Println("🤖 Adobe Security Bulletins Automated Scraper")
	fmt.Println("📡 Fetching https://helpx.adobe.com/security/security-bulletin.html")

	// Fetch the Adobe security page
	content, err := fetchAdobeSecurityPage()
	if err != nil {
		log.Fatalf("Failed to fetch Adobe security page: %v", err)
	}

	fmt.Printf("📄 Retrieved %d characters of content\n", len(content))

	// Parse bulletins from the content
	bulletins := extractBulletinsFromText(content)
	fmt.Printf("🔍 Extracted %d bulletins from content\n", len(bulletins))

	if len(bulletins) == 0 {
		fmt.Println("⚠️  No bulletins found. This might indicate a parsing issue.")
		fmt.Println("💡 Try running the manual parser for debugging.")
		os.Exit(0)
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
		fmt.Printf("✅ No new bulletins found (database has %d bulletins)\n", len(db.Bulletins))
		os.Exit(0)
	}

	fmt.Printf("📥 Adding %d new bulletins to database\n", len(newBulletins))

	// Add new bulletins to the front (most recent first)
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
		if i >= 10 { // Show first 10
			fmt.Printf("   ... and %d more\n", len(newBulletins)-10)
			break
		}
		fmt.Printf("  • %s: %s\n", bulletin.APSB, strings.TrimPrefix(bulletin.Title, bulletin.APSB+": "))
	}

	// Generate Hugo content
	fmt.Println("\n🏗️  Generating Hugo content...")
	if err := generateHugoContent(); err != nil {
		log.Fatalf("Failed to generate Hugo content: %v", err)
	}

	fmt.Println("✅ Automated scraping completed successfully!")
}

func fetchAdobeSecurityPage() (string, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", "https://helpx.adobe.com/security/security-bulletin.html", nil)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	// Set headers to mimic a real browser
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("DNT", "1")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP error: %d %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response: %w", err)
	}

	return string(body), nil
}

func extractBulletinsFromText(text string) []SecurityBulletin {
	var bulletins []SecurityBulletin
	seen := make(map[string]bool)

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
				break // Found a match, no need to try other patterns
			}
		}
	}

	return bulletins
}

func generateHugoContent() error {
	// For GitHub Actions, we'll rely on the workflow to run the content generator
	// This is a placeholder - in the actual workflow, we'll run both commands separately
	fmt.Println("📝 Content generation will be handled by the workflow")
	return nil
}

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
