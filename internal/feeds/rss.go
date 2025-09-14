package feeds

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/feeds"
	"github.com/superterran/adobe-digest/internal/adobe"
)

// Generator handles RSS feed generation for security bulletins
type Generator struct {
	baseURL  string
	siteURL  string
	feedPath string
	maxItems int
}

// NewGenerator creates a new RSS feed generator
func NewGenerator(baseURL, siteURL, feedPath string, maxItems int) *Generator {
	return &Generator{
		baseURL:  baseURL,
		siteURL:  siteURL,
		feedPath: feedPath,
		maxItems: maxItems,
	}
}

// GenerateRSSFeed creates an RSS feed for the given security bulletins
func (g *Generator) GenerateRSSFeed(bulletins []adobe.SecurityBulletin, title, description string) error {
	// Create the feed
	feed := &feeds.Feed{
		Title:       title,
		Link:        &feeds.Link{Href: g.siteURL},
		Description: description,
		Author:      &feeds.Author{Name: "Adobe Digest", Email: "noreply@adobedigest.com"},
		Created:     time.Now(),
		Copyright:   fmt.Sprintf("© %d Adobe Digest - Data sourced from Adobe PSIRT", time.Now().Year()),
	}

	// Sort bulletins by publication date (newest first)
	sortedBulletins := make([]adobe.SecurityBulletin, len(bulletins))
	copy(sortedBulletins, bulletins)

	for i := 0; i < len(sortedBulletins)-1; i++ {
		for j := i + 1; j < len(sortedBulletins); j++ {
			if sortedBulletins[i].PublishedAt.Before(sortedBulletins[j].PublishedAt) {
				sortedBulletins[i], sortedBulletins[j] = sortedBulletins[j], sortedBulletins[i]
			}
		}
	}

	// Limit to maxItems
	itemCount := g.maxItems
	if len(sortedBulletins) < itemCount {
		itemCount = len(sortedBulletins)
	}

	// Add items to feed
	for i := 0; i < itemCount; i++ {
		bulletin := sortedBulletins[i]
		item := g.createFeedItem(bulletin)
		feed.Items = append(feed.Items, item)
	}

	// Generate RSS XML
	rssXML, err := feed.ToRss()
	if err != nil {
		return fmt.Errorf("generating RSS XML: %w", err)
	}

	// Create feed directory if it doesn't exist
	feedDir := filepath.Dir(g.feedPath)
	if err := os.MkdirAll(feedDir, 0755); err != nil {
		return fmt.Errorf("creating feed directory: %w", err)
	}

	// Write RSS file
	if err := os.WriteFile(g.feedPath, []byte(rssXML), 0644); err != nil {
		return fmt.Errorf("writing RSS file: %w", err)
	}

	return nil
}

// createFeedItem creates an RSS item from a security bulletin
func (g *Generator) createFeedItem(bulletin adobe.SecurityBulletin) *feeds.Item {
	// Generate local URL
	productSlug := getProductSlug(bulletin.Product)
	bulletinSlug := strings.ToLower(bulletin.ID)
	localURL := fmt.Sprintf("%s/bulletins/%s/%s/", g.baseURL, productSlug, bulletinSlug)

	// Create description with key information
	description := g.createItemDescription(bulletin)

	// Create categories
	var categories []string
	categories = append(categories, bulletin.Product)
	categories = append(categories, bulletin.PriorityLabel)

	// Add CVE categories
	for _, vuln := range bulletin.Vulnerabilities {
		if vuln.CVE != "" {
			categories = append(categories, vuln.CVE)
		}
	}

	item := &feeds.Item{
		Title:       fmt.Sprintf("%s: %s", bulletin.ID, bulletin.Title),
		Link:        &feeds.Link{Href: localURL},
		Description: description,
		Id:          bulletin.ID,
		Created:     bulletin.PublishedAt,
		Updated:     bulletin.UpdatedAt,
	}

	// Set enclosure for structured data (optional)
	if bulletin.URL != "" {
		item.Enclosure = &feeds.Enclosure{
			Url:    bulletin.URL,
			Type:   "text/html",
			Length: "0",
		}
	}

	return item
}

// createItemDescription creates a rich HTML description for RSS items
func (g *Generator) createItemDescription(bulletin adobe.SecurityBulletin) string {
	var parts []string

	// Summary
	if bulletin.Summary != "" {
		parts = append(parts, fmt.Sprintf("<p><strong>Summary:</strong> %s</p>", bulletin.Summary))
	}

	// Priority and CVSS
	if bulletin.Priority > 0 {
		parts = append(parts, fmt.Sprintf("<p><strong>Priority:</strong> %s (%d)</p>",
			bulletin.PriorityLabel, bulletin.Priority))
	}

	maxCVSS := bulletin.GetMaxCVSS()
	if maxCVSS > 0 {
		parts = append(parts, fmt.Sprintf("<p><strong>Max CVSS Score:</strong> %.1f</p>", maxCVSS))
	}

	// CVEs
	cves := bulletin.GetCVEs()
	if len(cves) > 0 {
		cveList := strings.Join(cves, ", ")
		parts = append(parts, fmt.Sprintf("<p><strong>CVEs:</strong> %s</p>", cveList))
	}

	// Affected versions (show first few)
	versions := bulletin.GetAffectedVersionStrings()
	if len(versions) > 0 {
		versionList := versions[0]
		if len(versions) > 1 {
			versionList += " and others"
		}
		parts = append(parts, fmt.Sprintf("<p><strong>Affected Versions:</strong> %s</p>", versionList))
	}

	// Critical vulnerabilities warning
	if bulletin.HasCriticalVulnerabilities() {
		parts = append(parts, "<p><strong style='color: red;'>⚠️ Contains Critical Vulnerabilities</strong></p>")
	}

	// Link to original bulletin
	parts = append(parts, fmt.Sprintf("<p><a href='%s'>View Original Adobe Bulletin</a></p>", bulletin.URL))

	return strings.Join(parts, "\n")
}

// GenerateProductRSSFeed creates a product-specific RSS feed
func (g *Generator) GenerateProductRSSFeed(bulletins []adobe.SecurityBulletin, productName string) error {
	// Filter bulletins for this product
	var productBulletins []adobe.SecurityBulletin
	for _, bulletin := range bulletins {
		if strings.Contains(strings.ToLower(bulletin.Product), strings.ToLower(productName)) ||
			strings.Contains(strings.ToLower(productName), strings.ToLower(bulletin.Product)) {
			productBulletins = append(productBulletins, bulletin)
		}
	}

	if len(productBulletins) == 0 {
		return nil // No bulletins for this product
	}

	// Create product-specific feed path
	productSlug := getProductSlug(productName)
	feedDir := filepath.Dir(g.feedPath)
	productFeedPath := filepath.Join(feedDir, fmt.Sprintf("%s.xml", productSlug))

	// Temporarily update feed path
	originalPath := g.feedPath
	g.feedPath = productFeedPath
	defer func() { g.feedPath = originalPath }()

	title := fmt.Sprintf("Adobe %s Security Bulletins", productName)
	description := fmt.Sprintf("Security updates and bulletins for Adobe %s", productName)

	return g.GenerateRSSFeed(productBulletins, title, description)
}

// GenerateAllFeeds generates both the main RSS feed and product-specific feeds
func (g *Generator) GenerateAllFeeds(bulletins []adobe.SecurityBulletin, products []adobe.ProductConfig) error {
	// Generate main feed
	if err := g.GenerateRSSFeed(bulletins,
		"Adobe Security Bulletins",
		"Latest security updates for Adobe Commerce, AEM, and Experience Platform"); err != nil {
		return fmt.Errorf("generating main RSS feed: %w", err)
	}

	// Generate product-specific feeds
	for _, product := range products {
		if product.Enabled {
			if err := g.GenerateProductRSSFeed(bulletins, product.DisplayName); err != nil {
				return fmt.Errorf("generating RSS feed for %s: %w", product.DisplayName, err)
			}
		}
	}

	return nil
}

// Helper functions

func getProductSlug(productName string) string {
	slug := strings.ToLower(productName)
	slug = strings.ReplaceAll(slug, "adobe ", "")
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "/", "-")
	return slug
}
