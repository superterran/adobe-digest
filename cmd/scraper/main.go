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

	"github.com/PuerkitoBio/goquery"
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
	fmt.Println("🕷️  Adobe Security Bulletins Comprehensive Scraper")
	fmt.Println("📡 Scraping https://helpx.adobe.com/security/security-bulletin.html")

	// Load existing database to avoid duplicates
	dbPath := "data/security-bulletins.json"
	db := loadDatabase(dbPath)

	existingAPSBs := make(map[string]bool)
	for _, bulletin := range db.Bulletins {
		existingAPSBs[bulletin.APSB] = true
	}

	fmt.Printf("📊 Current database has %d bulletins\n", len(db.Bulletins))

	// Scrape all bulletins from the comprehensive page
	allBulletins, err := scrapeComprehensiveBulletins()
	if err != nil {
		log.Fatalf("Failed to scrape bulletins: %v", err)
	}

	fmt.Printf("🔍 Found %d total bulletins on Adobe's page\n", len(allBulletins))

	// Filter out existing bulletins
	var newBulletins []SecurityBulletin
	for _, bulletin := range allBulletins {
		if !existingAPSBs[bulletin.APSB] {
			newBulletins = append(newBulletins, bulletin)
		}
	}

	if len(newBulletins) == 0 {
		fmt.Println("✅ No new bulletins found - database is up to date")
		return
	}

	fmt.Printf("📥 Found %d new bulletins to add\n", len(newBulletins))

	// Add new bulletins to the front of the list (newest first)
	db.Bulletins = append(newBulletins, db.Bulletins...)
	db.LastUpdated = time.Now()

	// Save updated database
	if err := saveDatabase(dbPath, db); err != nil {
		log.Fatalf("Failed to save database: %v", err)
	}

	fmt.Printf("✅ Successfully added %d new bulletins\n", len(newBulletins))
	fmt.Printf("📊 Database now contains %d bulletins total\n", len(db.Bulletins))

	// Show some examples of what was added
	fmt.Println("\n📋 Sample of new bulletins added:")
	for i, bulletin := range newBulletins {
		if i >= 5 { // Show only first 5
			fmt.Printf("   ... and %d more\n", len(newBulletins)-5)
			break
		}
		fmt.Printf("  • %s: %s\n", bulletin.APSB, strings.TrimPrefix(bulletin.Title, bulletin.APSB+": "))
	}

	fmt.Println("\n🔄 Run 'go run cmd/content-generator/main.go generate' to update the site content")
}

func scrapeComprehensiveBulletins() ([]SecurityBulletin, error) {
	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	req, err := http.NewRequest("GET", "https://helpx.adobe.com/security/security-bulletin.html", nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	// Set comprehensive headers to appear like a real browser
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br, zstd")
	req.Header.Set("Sec-Ch-Ua", `"Not_A Brand";v="8", "Chromium";v="120", "Google Chrome";v="120"`)
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", `"Windows"`)
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")

	fmt.Println("🌐 Making request to Adobe...")

	// Add delay to be respectful
	time.Sleep(3 * time.Second)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	fmt.Printf("📡 Response status: %d\n", resp.StatusCode)

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parsing HTML: %w", err)
	}

	fmt.Println("🔍 Parsing HTML content...")

	var bulletins []SecurityBulletin
	bulletinMap := make(map[string]SecurityBulletin) // Prevent duplicates

	// Look for various patterns that might contain bulletin information
	selectors := []string{
		"a[href*='apsb']",     // Links containing 'apsb'
		"a[href*='security']", // Security-related links
		"*:contains('APSB')",  // Any element containing 'APSB'
	}

	for _, selector := range selectors {
		doc.Find(selector).Each(func(i int, s *goquery.Selection) {
			processBulletinElement(s, &bulletins, bulletinMap)
		})
	}

	// Also look for text patterns that might be bulletin references
	doc.Find("*").Each(func(i int, s *goquery.Selection) {
		text := s.Text()
		if containsAPSBPattern(text) {
			processTextForBulletins(s, text, &bulletins, bulletinMap)
		}
	})

	fmt.Printf("✅ Extracted %d unique bulletins\n", len(bulletins))

	return bulletins, nil
}

func processBulletinElement(s *goquery.Selection, bulletins *[]SecurityBulletin, bulletinMap map[string]SecurityBulletin) {
	href, hasHref := s.Attr("href")
	text := strings.TrimSpace(s.Text())

	// Extract APSB ID
	apsbID := extractAPSBFromText(text)
	if apsbID == "" && hasHref {
		apsbID = extractAPSBFromURL(href)
	}

	if apsbID == "" {
		return
	}

	// Skip if we already processed this APSB
	if _, exists := bulletinMap[apsbID]; exists {
		return
	}

	// Build full URL
	var fullURL string
	if hasHref {
		if strings.HasPrefix(href, "http") {
			fullURL = href
		} else if strings.HasPrefix(href, "/") {
			fullURL = "https://helpx.adobe.com" + href
		} else {
			fullURL = "https://helpx.adobe.com/security/products/" + href
		}
	} else {
		// Generate likely URL based on APSB ID and context
		fullURL = generateBulletinURL(apsbID, text)
	}

	// Create bulletin
	bulletin := SecurityBulletin{
		APSB:        apsbID,
		Title:       generateTitle(apsbID, text),
		Description: generateDescription(text, fullURL),
		URL:         fullURL,
		Date:        parseOrEstimateDate(text, apsbID),
		Products:    inferProducts(text, fullURL),
		Severity:    inferSeverity(text),
	}

	bulletinMap[apsbID] = bulletin
	*bulletins = append(*bulletins, bulletin)

	fmt.Printf("  📄 %s: %s\n", apsbID, strings.TrimPrefix(bulletin.Title, apsbID+": "))
}

func processTextForBulletins(s *goquery.Selection, text string, bulletins *[]SecurityBulletin, bulletinMap map[string]SecurityBulletin) {
	// Find all APSB IDs in the text
	re := regexp.MustCompile(`APSB\d{2}-\d{2,3}`)
	matches := re.FindAllString(text, -1)

	for _, apsbID := range matches {
		if _, exists := bulletinMap[apsbID]; exists {
			continue
		}

		// Try to extract more context around the APSB ID
		context := extractContextAroundAPSB(text, apsbID)

		bulletin := SecurityBulletin{
			APSB:        apsbID,
			Title:       generateTitle(apsbID, context),
			Description: generateDescription(context, ""),
			URL:         generateBulletinURL(apsbID, context),
			Date:        parseOrEstimateDate(context, apsbID),
			Products:    inferProducts(context, ""),
			Severity:    inferSeverity(context),
		}

		bulletinMap[apsbID] = bulletin
		*bulletins = append(*bulletins, bulletin)

		fmt.Printf("  📄 %s: %s\n", apsbID, strings.TrimPrefix(bulletin.Title, apsbID+": "))
	}
}

func containsAPSBPattern(text string) bool {
	re := regexp.MustCompile(`APSB\d{2}-\d{2,3}`)
	return re.MatchString(text)
}

func extractAPSBFromText(text string) string {
	re := regexp.MustCompile(`APSB\d{2}-\d{2,3}`)
	match := re.FindString(text)
	return strings.ToUpper(match)
}

func extractAPSBFromURL(url string) string {
	re := regexp.MustCompile(`apsb\d{2}-\d{2,3}`)
	match := re.FindString(strings.ToLower(url))
	return strings.ToUpper(match)
}

func extractContextAroundAPSB(text, apsbID string) string {
	// Find the APSB and get surrounding context
	index := strings.Index(text, apsbID)
	if index == -1 {
		return text
	}

	start := index - 50
	if start < 0 {
		start = 0
	}

	end := index + len(apsbID) + 100
	if end > len(text) {
		end = len(text)
	}

	return strings.TrimSpace(text[start:end])
}

func generateTitle(apsbID, text string) string {
	// Clean up the text to make a reasonable title
	title := strings.TrimSpace(text)
	title = strings.ReplaceAll(title, "\n", " ")
	title = strings.ReplaceAll(title, "\t", " ")

	// Remove extra whitespace
	re := regexp.MustCompile(`\s+`)
	title = re.ReplaceAllString(title, " ")

	// If the title doesn't start with the APSB ID, prepend it
	if !strings.HasPrefix(title, apsbID) {
		if title == "" {
			title = fmt.Sprintf("%s: Security update available", apsbID)
		} else {
			title = fmt.Sprintf("%s: %s", apsbID, title)
		}
	}

	// Limit title length
	if len(title) > 150 {
		title = title[:147] + "..."
	}

	return title
}

func generateDescription(text, url string) string {
	if text == "" {
		return "Adobe has released security updates. More details in the security bulletin."
	}

	desc := strings.TrimSpace(text)
	if len(desc) > 200 {
		desc = desc[:197] + "..."
	}

	if url != "" {
		desc += fmt.Sprintf(" More details at %s", url)
	}

	return desc
}

func generateBulletinURL(apsbID, context string) string {
	// Try to infer the product from context
	productPath := inferProductPath(context)
	return fmt.Sprintf("https://helpx.adobe.com/security/products/%s/%s.html", productPath, strings.ToLower(apsbID))
}

func inferProductPath(text string) string {
	text = strings.ToLower(text)

	productMap := map[string]string{
		"acrobat":            "acrobat",
		"reader":             "acrobat",
		"photoshop":          "photoshop",
		"illustrator":        "illustrator",
		"indesign":           "indesign",
		"after effects":      "after-effects",
		"premiere":           "premiere",
		"lightroom":          "lightroom",
		"bridge":             "bridge",
		"dreamweaver":        "dreamweaver",
		"animate":            "animate",
		"audition":           "audition",
		"dimension":          "dimension",
		"experience manager": "experience-manager",
		"aem":                "experience-manager",
		"commerce":           "commerce",
		"magento":            "magento",
		"coldfusion":         "coldfusion",
		"campaign":           "campaign",
		"substance":          "substance",
	}

	for keyword, path := range productMap {
		if strings.Contains(text, keyword) {
			return path
		}
	}

	return "other"
}

func parseOrEstimateDate(text, apsbID string) time.Time {
	// Try to extract year from APSB ID (APSB24-XX means 2024)
	re := regexp.MustCompile(`APSB(\d{2})-\d{2,3}`)
	matches := re.FindStringSubmatch(apsbID)

	if len(matches) > 1 {
		yearSuffix := matches[1]
		year := 2000 + parseInt(yearSuffix)

		// Default to middle of the year if no specific date found
		return time.Date(year, 6, 15, 0, 0, 0, 0, time.UTC)
	}

	// Fallback to current year
	return time.Date(time.Now().Year(), 6, 15, 0, 0, 0, 0, time.UTC)
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

func inferProducts(text, url string) []string {
	combined := strings.ToLower(text + " " + url)
	var products []string

	productMap := map[string][]string{
		"acrobat":            {"Adobe Acrobat", "Adobe Acrobat Reader"},
		"reader":             {"Adobe Acrobat Reader"},
		"photoshop":          {"Adobe Photoshop"},
		"illustrator":        {"Adobe Illustrator"},
		"indesign":           {"Adobe InDesign"},
		"after effects":      {"Adobe After Effects"},
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
	}

	for keyword, productList := range productMap {
		if strings.Contains(combined, keyword) {
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
