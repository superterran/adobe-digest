package main

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

func main() {
	fmt.Println("🔍 Adobe Page Content Analyzer")

	content, err := fetchPageContent("https://helpx.adobe.com/security/security-bulletin.html")
	if err != nil {
		log.Fatalf("Failed to fetch page: %v", err)
	}

	fmt.Printf("📄 Page content length: %d characters\n", len(content))

	// Look for APSB patterns in the content
	fmt.Println("\n🔍 Looking for APSB patterns...")

	// Simple pattern to find any APSB references
	re := regexp.MustCompile(`APSB\d{2}-\d{2,3}[^a-zA-Z]*[^|]*`)
	matches := re.FindAllString(content, 20) // Get first 20 matches

	fmt.Printf("Found %d APSB matches:\n", len(matches))
	for i, match := range matches {
		if i >= 10 { // Show first 10
			fmt.Printf("  ... and %d more\n", len(matches)-10)
			break
		}

		// Clean up for display
		cleaned := strings.ReplaceAll(match, "\n", " ")
		cleaned = strings.ReplaceAll(cleaned, "\t", " ")
		cleaned = regexp.MustCompile(`\s+`).ReplaceAllString(cleaned, " ")
		cleaned = strings.TrimSpace(cleaned)

		if len(cleaned) > 100 {
			cleaned = cleaned[:100] + "..."
		}

		fmt.Printf("  %d: %s\n", i+1, cleaned)
	}

	// Also look for table-like structures
	fmt.Println("\n🔍 Looking for table patterns...")
	tableRe := regexp.MustCompile(`\|[^|]*APSB\d{2}-\d{2,3}[^|]*\|`)
	tableMatches := tableRe.FindAllString(content, 10)

	fmt.Printf("Found %d table matches:\n", len(tableMatches))
	for i, match := range tableMatches {
		cleaned := strings.TrimSpace(strings.ReplaceAll(match, "|", ""))
		fmt.Printf("  %d: %s\n", i+1, cleaned)
	}
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
