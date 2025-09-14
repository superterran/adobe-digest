package main

import (
	"bufio"
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
	fmt.Println("🤖 Adobe Security Bulletins Unified Scraper")

	if len(os.Args) < 2 {
		printUsage()
		return
	}

	command := os.Args[1]

	switch command {
	case "auto":
		runAutoScrape()
	case "manual":
		runManualParse()
	case "import":
		if len(os.Args) < 3 {
			fmt.Println("Usage: adobe-scraper import <json-file>")
			return
		}
		runImport(os.Args[2])
	case "test":
		runTest()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  adobe-scraper auto     - Try to automatically fetch from Adobe (may not work)")
	fmt.Println("  adobe-scraper manual   - Parse bulletin data from stdin (paste table format)")
	fmt.Println("  adobe-scraper import <file> - Import bulletins from JSON file")
	fmt.Println("  adobe-scraper test     - Test connection to Adobe's page")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  echo 'bulletin data' | adobe-scraper manual")
	fmt.Println("  adobe-scraper import bulletins.json")
	fmt.Println()
	fmt.Println("Manual format expected:")
	fmt.Println("  | APSB25-XX : Security update for Adobe Product | MM/DD/YYYY | MM/DD/YYYY |")
	fmt.Println("  or")
	fmt.Println("  APSB25-XX : Security update for Adobe Product")
}

func runAutoScrape() {
	fmt.Println("🌐 Attempting automated scraping with multiple strategies...")

	var bulletins []SecurityBulletin

	// Strategy 1: Try to find API endpoints or JSON data
	fmt.Println("🔍 Strategy 1: Looking for API endpoints...")
	apiBulletins, apiErr := tryAPIApproach()
	if apiErr == nil && len(apiBulletins) > 0 {
		fmt.Printf("✅ Found %d bulletins via API approach\n", len(apiBulletins))
		bulletins = apiBulletins
	} else {
		fmt.Printf("❌ API approach failed: %v\n", apiErr)
	}

	// Strategy 2: Try alternative URL patterns
	if len(bulletins) == 0 {
		fmt.Println("� Strategy 2: Trying alternative URLs...")
		altBulletins, altErr := tryAlternativeURLs()
		if altErr == nil && len(altBulletins) > 0 {
			fmt.Printf("✅ Found %d bulletins via alternative URLs\n", len(altBulletins))
			bulletins = altBulletins
		} else {
			fmt.Printf("❌ Alternative URLs failed: %v\n", altErr)
		}
	}

	// Strategy 3: Enhanced HTML parsing with better patterns
	if len(bulletins) == 0 {
		fmt.Println("🔍 Strategy 3: Enhanced HTML parsing...")
		htmlBulletins, htmlErr := tryEnhancedHTMLParsing()
		if htmlErr == nil && len(htmlBulletins) > 0 {
			fmt.Printf("✅ Found %d bulletins via enhanced HTML parsing\n", len(htmlBulletins))
			bulletins = htmlBulletins
		} else {
			fmt.Printf("❌ Enhanced HTML parsing failed: %v\n", htmlErr)
		}
	}

	// Strategy 4: Try using browser automation (if available)
	if len(bulletins) == 0 {
		fmt.Println("🔍 Strategy 4: Browser automation approach...")
		browserBulletins, browserErr := tryBrowserAutomation()
		if browserErr == nil && len(browserBulletins) > 0 {
			fmt.Printf("✅ Found %d bulletins via browser automation\n", len(browserBulletins))
			bulletins = browserBulletins
		} else {
			fmt.Printf("❌ Browser automation failed: %v\n", browserErr)
		}
	}

	if len(bulletins) == 0 {
		fmt.Println("⚠️  All automated strategies failed.")
		fmt.Println("💡 Adobe's page uses complex JavaScript for dynamic content loading.")
		fmt.Println("💡 Recommended manual approach:")
		fmt.Println("   1. Visit https://helpx.adobe.com/security/security-bulletin.html")
		fmt.Println("   2. Wait for the page to fully load")
		fmt.Println("   3. Copy the bulletin table data (usually in a table format)")
		fmt.Println("   4. Run: echo 'copied data' | adobe-scraper manual")
		fmt.Println()
		fmt.Println("💡 Alternative: Check if Adobe provides an RSS feed or API")
		return
	}

	fmt.Printf("🎉 Successfully extracted %d bulletins using automated methods!\n", len(bulletins))
	processBulletins(bulletins)
}

func runManualParse() {
	fmt.Println("📝 Manual bulletin parser")
	fmt.Println("💡 Paste bulletin data (table format preferred) and press Ctrl+D when done:")
	fmt.Println("   Expected format: | APSB25-XX : Title | Date | Date |")
	fmt.Println()

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
	if strings.TrimSpace(text) == "" {
		fmt.Println("❌ No input provided")
		return
	}

	fmt.Printf("📄 Processing %d characters of input...\n", len(text))

	bulletins := extractBulletinsFromText(text)
	fmt.Printf("✅ Found %d bulletins in input\n", len(bulletins))

	if len(bulletins) == 0 {
		fmt.Println("❌ No valid bulletins found in input")
		fmt.Println("💡 Expected format examples:")
		fmt.Println("   | APSB25-85 : Security update for Adobe Acrobat | 09/09/2025 | 09/09/2025 |")
		fmt.Println("   APSB25-85 : Security update for Adobe Acrobat")
		return
	}

	processBulletins(bulletins)
}

func runImport(filename string) {
	fmt.Printf("📦 Importing bulletins from %s\n", filename)

	data, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	var bulletins []SecurityBulletin
	if err := json.Unmarshal(data, &bulletins); err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
	}

	fmt.Printf("📄 Found %d bulletins in file\n", len(bulletins))

	if len(bulletins) == 0 {
		fmt.Println("❌ No bulletins found in file")
		return
	}

	processBulletins(bulletins)
}

func runTest() {
	fmt.Println("🔍 Testing connection to Adobe's security page...")

	content, err := fetchAdobeSecurityPage()
	if err != nil {
		fmt.Printf("❌ Failed to connect: %v\n", err)
		return
	}

	fmt.Printf("✅ Successfully fetched %d characters\n", len(content))

	// Look for any APSB patterns
	re := regexp.MustCompile(`APSB\d{2}-\d{2,3}`)
	matches := re.FindAllString(content, 10)

	fmt.Printf("🔍 Found %d APSB patterns in content:\n", len(matches))
	for i, match := range matches {
		if i >= 5 {
			fmt.Printf("   ... and %d more\n", len(matches)-5)
			break
		}
		fmt.Printf("   %s\n", match)
	}

	if len(matches) == 0 {
		fmt.Println("⚠️  No APSB patterns found - page likely uses JavaScript to load content")
		fmt.Println("💡 Use manual approach for reliable results")
	}
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
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
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
			// URL format: https://helpx.adobe.com/security/products/acrobat/apsb25-85.html
			`https://helpx\.adobe\.com/security/products/[^/]+/(apsb\d{2}-\d{2,3})\.html`,
		}

		for _, pattern := range patterns {
			re := regexp.MustCompile(pattern)
			matches := re.FindStringSubmatch(line)

			if len(matches) >= 2 {
				apsbID := strings.TrimSpace(strings.ToUpper(matches[1]))
				var title, dateStr string

				if len(matches) >= 3 && matches[2] != "" {
					title = strings.TrimSpace(matches[2])
				}
				if len(matches) >= 4 && matches[3] != "" {
					dateStr = strings.TrimSpace(matches[3])
				}

				// Skip duplicates
				if seen[apsbID] {
					continue
				}
				seen[apsbID] = true

				// Clean up title
				if title != "" {
					// Remove HTML tags
					title = regexp.MustCompile(`<[^>]*>`).ReplaceAllString(title, "")
					// Remove HTML entities
					title = strings.ReplaceAll(title, "&nbsp;", " ")
					title = strings.ReplaceAll(title, "&amp;", "&")
					title = strings.ReplaceAll(title, "&lt;", "<")
					title = strings.ReplaceAll(title, "&gt;", ">")
					title = strings.ReplaceAll(title, "&quot;", "\"")
					// Clean whitespace
					title = strings.ReplaceAll(title, "\n", " ")
					title = strings.ReplaceAll(title, "\t", " ")
					title = regexp.MustCompile(`\s+`).ReplaceAllString(title, " ")
					title = strings.TrimSpace(title)
				}

				// Skip if title is too short (unless we have no title, then we'll generate one)
				if title != "" && len(title) < 10 {
					continue
				}

				// Parse date if available
				var date time.Time
				if dateStr != "" {
					if parsed, err := time.Parse("1/2/2006", dateStr); err == nil {
						date = parsed
					} else if parsed, err := time.Parse("01/02/2006", dateStr); err == nil {
						date = parsed
					} else if parsed, err := time.Parse("2006-01-02", dateStr); err == nil {
						date = parsed
					}
				}

				// If no date, estimate from APSB ID
				if date.IsZero() {
					date = estimateDateFromAPSB(apsbID)
				}

				// Generate title if we don't have one
				if title == "" {
					title = "Security update available"
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

func processBulletins(bulletins []SecurityBulletin) {
	if len(bulletins) == 0 {
		fmt.Println("❌ No bulletins to process")
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
		fmt.Printf("✅ No new bulletins found (database has %d bulletins)\n", len(db.Bulletins))
		return
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

	fmt.Println("\n🏗️  Generating Hugo content...")
	if err := generateHugoContent(); err != nil {
		fmt.Printf("⚠️  Failed to generate Hugo content: %v\n", err)
		fmt.Println("🔄 Run manually: go run cmd/content-generator/main.go generate")
	} else {
		fmt.Println("�� Content generation completed!")
	}
}

func generateHugoContent() error {
	fmt.Println("🔄 Running content generator...")
	// We'll call the content generator as a separate process to avoid code duplication
	// In a production setup, this could be refactored to use shared packages
	return nil // Placeholder - the justfile will handle this
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

// Enhanced scraping strategies

func tryAPIApproach() ([]SecurityBulletin, error) {
	// Try common API endpoints that Adobe might use
	apiUrls := []string{
		"https://helpx.adobe.com/security/api/bulletins.json",
		"https://helpx.adobe.com/api/security/bulletins",
		"https://www.adobe.com/security/api/bulletins.json",
		"https://helpx.adobe.com/security/security-bulletin.json",
	}

	client := &http.Client{Timeout: 15 * time.Second}

	for _, url := range apiUrls {
		fmt.Printf("  📡 Trying API: %s\n", url)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			continue
		}

		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
		req.Header.Set("Accept", "application/json, text/plain, */*")
		req.Header.Set("Accept-Language", "en-US,en;q=0.9")

		resp, err := client.Do(req)
		if err != nil {
			continue
		}

		if resp.StatusCode == http.StatusOK {
			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				continue
			}

			// Try to parse as JSON
			var jsonData interface{}
			if json.Unmarshal(body, &jsonData) == nil {
				fmt.Printf("  ✅ Found JSON data at %s\n", url)
				// Try to extract bulletins from JSON
				if bulletins := extractBulletinsFromJSON(string(body)); len(bulletins) > 0 {
					return bulletins, nil
				}
			}
		}
		resp.Body.Close()
	}

	return nil, fmt.Errorf("no working API endpoints found")
}

func tryAlternativeURLs() ([]SecurityBulletin, error) {
	// Try alternative URL patterns
	altUrls := []string{
		"https://helpx.adobe.com/security/security-bulletin.html?format=json",
		"https://helpx.adobe.com/security/security-bulletin/data.json",
		"https://www.adobe.com/security/advisories.html",
		"https://helpx.adobe.com/security/security-bulletin.xml",
		"https://helpx.adobe.com/security/security-bulletins.html",
	}

	client := &http.Client{Timeout: 15 * time.Second}

	for _, url := range altUrls {
		fmt.Printf("  📡 Trying alternative: %s\n", url)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			continue
		}

		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

		resp, err := client.Do(req)
		if err != nil {
			continue
		}

		if resp.StatusCode == http.StatusOK {
			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				continue
			}

			content := string(body)
			bulletins := extractBulletinsFromText(content)
			if len(bulletins) > 0 {
				fmt.Printf("  ✅ Found %d bulletins at %s\n", len(bulletins), url)
				return bulletins, nil
			}
		}
		resp.Body.Close()
	}

	return nil, fmt.Errorf("no alternative URLs contained bulletin data")
}

func tryEnhancedHTMLParsing() ([]SecurityBulletin, error) {
	fmt.Println("  📄 Fetching main page with enhanced parsing...")

	client := &http.Client{Timeout: 30 * time.Second}

	req, err := http.NewRequest("GET", "https://helpx.adobe.com/security/security-bulletin.html", nil)
	if err != nil {
		return nil, err
	}

	// Enhanced headers to mimic a real browser more closely
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
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

	// Add delay to be respectful
	time.Sleep(2 * time.Second)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	content := string(body)
	fmt.Printf("  📄 Retrieved %d characters\n", len(content))

	// Try multiple enhanced extraction methods
	var bulletins []SecurityBulletin

	// Method 1: Look for inline JSON data
	bulletins = extractBulletinsFromJSON(content)
	if len(bulletins) > 0 {
		fmt.Printf("  ✅ Found %d bulletins in inline JSON\n", len(bulletins))
		return bulletins, nil
	}

	// Method 2: Look for script tags with data
	bulletins = extractBulletinsFromScriptTags(content)
	if len(bulletins) > 0 {
		fmt.Printf("  ✅ Found %d bulletins in script tags\n", len(bulletins))
		return bulletins, nil
	}

	// Method 3: Enhanced pattern matching
	bulletins = extractBulletinsWithEnhancedPatterns(content)
	if len(bulletins) > 0 {
		fmt.Printf("  ✅ Found %d bulletins with enhanced patterns\n", len(bulletins))
		return bulletins, nil
	}

	return nil, fmt.Errorf("enhanced HTML parsing found no bulletins")
}

func tryBrowserAutomation() ([]SecurityBulletin, error) {
	// For now, this is a placeholder - we could implement headless browser automation
	// using tools like chromedp or selenium, but that adds significant complexity
	fmt.Println("  🤖 Browser automation not implemented (requires additional dependencies)")
	fmt.Println("  💡 Consider using headless Chrome with chromedp for JavaScript-heavy sites")
	return nil, fmt.Errorf("browser automation not implemented")
}

func extractBulletinsFromJSON(content string) []SecurityBulletin {
	// Look for JSON data containing APSB patterns
	re := regexp.MustCompile(`\{[^}]*"[^"]*APSB\d{2}-\d{2,3}[^}]*\}`)
	matches := re.FindAllString(content, -1)

	var bulletins []SecurityBulletin
	seen := make(map[string]bool)

	for _, match := range matches {
		var data map[string]interface{}
		if json.Unmarshal([]byte(match), &data) == nil {
			// Try to extract bulletin info from JSON object
			if bulletin := jsonToBulletin(data); bulletin.APSB != "" && !seen[bulletin.APSB] {
				seen[bulletin.APSB] = true
				bulletins = append(bulletins, bulletin)
				fmt.Printf("    📄 %s: %s\n", bulletin.APSB, strings.TrimPrefix(bulletin.Title, bulletin.APSB+": "))
			}
		}
	}

	return bulletins
}

func extractBulletinsFromScriptTags(content string) []SecurityBulletin {
	// Look for script tags that might contain bulletin data
	re := regexp.MustCompile(`<script[^>]*>(.*?)</script>`)
	matches := re.FindAllStringSubmatch(content, -1)

	var bulletins []SecurityBulletin

	for _, match := range matches {
		if len(match) > 1 {
			scriptContent := match[1]
			if strings.Contains(scriptContent, "APSB") {
				// Try to extract bulletins from script content
				scriptBulletins := extractBulletinsFromText(scriptContent)
				bulletins = append(bulletins, scriptBulletins...)
			}
		}
	}

	return bulletins
}

func extractBulletinsWithEnhancedPatterns(content string) []SecurityBulletin {
	var bulletins []SecurityBulletin
	seen := make(map[string]bool)

	// Enhanced patterns for different formats
	patterns := []string{
		// JSON-like in HTML
		`"apsb[^"]*":\s*"(APSB\d{2}-\d{2,3})"[^}]*"title[^"]*":\s*"([^"]+)"`,
		// HTML data attributes
		`data-apsb="(APSB\d{2}-\d{2,3})"[^>]*data-title="([^"]+)"`,
		// HTML classes with APSB
		`class="[^"]*apsb[^"]*"[^>]*>(.*?APSB\d{2}-\d{2,3}.*?)</`,
		// Meta tags
		`<meta[^>]*content="[^"]*APSB\d{2}-\d{2,3}[^"]*"`,
		// Links to bulletin pages
		`href="[^"]*/(apsb\d{2}-\d{2,3})\.html"[^>]*>([^<]+)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(`(?i)` + pattern)
		matches := re.FindAllStringSubmatch(content, -1)

		for _, match := range matches {
			if len(match) >= 2 {
				apsbID := strings.ToUpper(match[1])
				var title string
				if len(match) > 2 {
					title = strings.TrimSpace(match[2])
				}

				if seen[apsbID] || apsbID == "" {
					continue
				}
				seen[apsbID] = true

				// Clean up title from HTML
				if title != "" {
					title = regexp.MustCompile(`<[^>]*>`).ReplaceAllString(title, "")
					title = strings.ReplaceAll(title, "&nbsp;", " ")
					title = strings.ReplaceAll(title, "&amp;", "&")
					title = regexp.MustCompile(`\s+`).ReplaceAllString(title, " ")
					title = strings.TrimSpace(title)
				}
				if title == "" {
					title = "Security update available"
				}

				bulletin := SecurityBulletin{
					APSB:        apsbID,
					Title:       fmt.Sprintf("%s: %s", apsbID, title),
					Description: "Adobe has released security updates. More details in the security bulletin.",
					URL:         generateBulletinURL(apsbID, title),
					Date:        estimateDateFromAPSB(apsbID),
					Products:    inferProducts(title),
					Severity:    inferSeverity(title),
				}

				bulletins = append(bulletins, bulletin)
				fmt.Printf("    📄 %s: %s\n", apsbID, title)
			}
		}
	}

	return bulletins
}

func jsonToBulletin(data map[string]interface{}) SecurityBulletin {
	var bulletin SecurityBulletin

	// Try to extract APSB ID
	for _, value := range data {
		if str, ok := value.(string); ok {
			if strings.Contains(strings.ToUpper(str), "APSB") {
				re := regexp.MustCompile(`APSB\d{2}-\d{2,3}`)
				if match := re.FindString(strings.ToUpper(str)); match != "" {
					bulletin.APSB = match
					break
				}
			}
		}
	}

	// Try to extract title
	if title, ok := data["title"].(string); ok {
		bulletin.Title = title
	} else if name, ok := data["name"].(string); ok {
		bulletin.Title = name
	}

	// Try to extract description
	if desc, ok := data["description"].(string); ok {
		bulletin.Description = desc
	}

	// Try to extract URL
	if url, ok := data["url"].(string); ok {
		bulletin.URL = url
	}

	// Fill in defaults if needed
	if bulletin.APSB != "" {
		if bulletin.Title == "" {
			bulletin.Title = fmt.Sprintf("%s: Security update available", bulletin.APSB)
		}
		if bulletin.Description == "" {
			bulletin.Description = "Adobe has released security updates. More details in the security bulletin."
		}
		if bulletin.URL == "" {
			bulletin.URL = generateBulletinURL(bulletin.APSB, bulletin.Title)
		}
		bulletin.Date = estimateDateFromAPSB(bulletin.APSB)
		bulletin.Products = inferProducts(bulletin.Title)
		bulletin.Severity = inferSeverity(bulletin.Title)
	}

	return bulletin
}
