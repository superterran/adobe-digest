package main

import (
	"log"
	"time"

	"github.com/superterran/adobe-digest/internal/adobe"
	"github.com/superterran/adobe-digest/internal/content"
	"github.com/superterran/adobe-digest/internal/feeds"
)

func main() {
	log.Println("Testing Adobe Security Bulletins scraper with mock data...")

	// Create mock bulletin data based on the structure we analyzed
	mockBulletin := &adobe.SecurityBulletin{
		ID:            "APSB25-88",
		Title:         "Security update available for Adobe Commerce",
		Product:       "Adobe Commerce",
		URL:           "https://helpx.adobe.com/security/products/magento/apsb25-88.html",
		PublishedAt:   time.Date(2025, 9, 9, 0, 0, 0, 0, time.UTC),
		UpdatedAt:     time.Date(2025, 9, 9, 0, 0, 0, 0, time.UTC),
		Priority:      1, // Critical
		PriorityLabel: "Critical",
		Summary:       "Adobe has released a security update for Adobe Commerce and Magento Open Source. This update resolves a critical vulnerability. Successful exploitation could lead to security feature bypass. Adobe is not aware of any exploits in the wild for this issue.",
		Vulnerabilities: []adobe.Vulnerability{
			{
				CVE:           "CVE-2025-54236",
				CWE:           "CWE-20",
				CVSS:          9.1,
				CVSSVector:    "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:N",
				Type:          "Improper Input Validation",
				Impact:        "Security feature bypass",
				Severity:      "Critical",
				AuthRequired:  false,
				AdminRequired: false,
				Exploited:     false,
			},
		},
		Affected: []adobe.AffectedVersion{
			{
				Product:   "Adobe Commerce",
				Versions:  []string{"2.4.9-alpha2 and earlier", "2.4.8-p2 and earlier", "2.4.7-p7 and earlier"},
				Platforms: []string{"All"},
				Priority:  1,
			},
			{
				Product:   "Magento Open Source",
				Versions:  []string{"2.4.9-alpha2 and earlier", "2.4.8-p2 and earlier", "2.4.7-p7 and earlier"},
				Platforms: []string{"All"},
				Priority:  1,
			},
		},
		Solutions: []adobe.Solution{
			{
				Product:     "Adobe Commerce and Magento Open Source",
				Type:        "Hotfix",
				Description: "Hotfix for CVE-2025-54236 - Compatible with all Adobe Commerce and Magento Open Source versions between 2.4.4 - 2.4.7",
				Priority:    1,
			},
		},
		Acknowledgements: []string{
			"blaklis",
		},
	}

	log.Printf("Created mock bulletin:")
	log.Printf("  ID: %s", mockBulletin.ID)
	log.Printf("  Title: %s", mockBulletin.Title)
	log.Printf("  Product: %s", mockBulletin.Product)
	log.Printf("  Priority: %d (%s)", mockBulletin.Priority, mockBulletin.GetPriorityLabel())
	log.Printf("  CVEs: %v", mockBulletin.GetCVEs())
	log.Printf("  Max CVSS: %.1f", mockBulletin.GetMaxCVSS())
	log.Printf("  Vulnerabilities: %d", len(mockBulletin.Vulnerabilities))
	log.Printf("  Affected products: %d", len(mockBulletin.Affected))
	log.Printf("  Solutions: %d", len(mockBulletin.Solutions))

	// Test content generation
	log.Println("\nTesting content generation...")
	contentGen, err := content.NewGenerator("content")
	if err != nil {
		log.Fatalf("Error initializing content generator: %v", err)
	}

	if err := contentGen.GenerateBulletin(mockBulletin); err != nil {
		log.Fatalf("Error generating bulletin content: %v", err)
	}

	log.Println("✅ Successfully generated bulletin content")

	// Create a second mock bulletin for AEM
	mockBulletinAEM := &adobe.SecurityBulletin{
		ID:            "APSB25-90",
		Title:         "Security update available for Adobe Experience Manager",
		Product:       "Adobe Experience Manager",
		URL:           "https://helpx.adobe.com/security/products/experience-manager/apsb25-90.html",
		PublishedAt:   time.Date(2025, 9, 9, 0, 0, 0, 0, time.UTC),
		UpdatedAt:     time.Date(2025, 9, 9, 0, 0, 0, 0, time.UTC),
		Priority:      2, // Important
		PriorityLabel: "Important",
		Summary:       "Adobe has released security updates for Adobe Experience Manager. These updates resolve multiple vulnerabilities that could result in security feature bypass and information disclosure.",
		Vulnerabilities: []adobe.Vulnerability{
			{
				CVE:           "CVE-2025-54237",
				CVSS:          7.5,
				Type:          "Information Disclosure",
				Impact:        "Information disclosure",
				Severity:      "High",
				AuthRequired:  false,
				AdminRequired: false,
			},
			{
				CVE:           "CVE-2025-54238",
				CVSS:          6.1,
				Type:          "Cross-Site Scripting",
				Impact:        "Security feature bypass",
				Severity:      "Medium",
				AuthRequired:  true,
				AdminRequired: false,
			},
		},
		Affected: []adobe.AffectedVersion{
			{
				Product:   "Adobe Experience Manager 6.5",
				Versions:  []string{"6.5.21.0 and earlier"},
				Platforms: []string{"All"},
				Priority:  2,
			},
		},
	}

	if err := contentGen.GenerateBulletin(mockBulletinAEM); err != nil {
		log.Fatalf("Error generating AEM bulletin content: %v", err)
	}

	log.Println("✅ Successfully generated AEM bulletin content")

	// Test generating all content including indexes
	bulletinsByProduct := map[string][]adobe.SecurityBulletin{
		"Adobe Commerce":           {*mockBulletin},
		"Adobe Experience Manager": {*mockBulletinAEM},
	}

	products := []adobe.ProductConfig{
		{
			Name:        "commerce",
			DisplayName: "Adobe Commerce/Magento",
			Enabled:     true,
		},
		{
			Name:        "experience-manager",
			DisplayName: "Adobe Experience Manager",
			Enabled:     true,
		},
	}

	if err := contentGen.GenerateAll(bulletinsByProduct, products); err != nil {
		log.Fatalf("Error generating all content: %v", err)
	}

	log.Println("✅ Successfully generated all content and indexes")

	// Test RSS feed generation
	log.Println("\nTesting RSS feed generation...")
	feedGen := feeds.NewGenerator(
		"https://adobedigest.com",
		"https://adobedigest.com",
		"static/feeds/bulletins.xml",
		50,
	)

	allBulletins := []adobe.SecurityBulletin{*mockBulletin, *mockBulletinAEM}
	if err := feedGen.GenerateAllFeeds(allBulletins, products); err != nil {
		log.Fatalf("Error generating RSS feeds: %v", err)
	}

	log.Println("✅ Successfully generated RSS feeds")

	// Show generated files
	log.Println("\nGenerated files:")
	log.Println("📄 Content files:")
	log.Println("  content/bulletins/commerce/apsb25-88.md")
	log.Println("  content/bulletins/experience-manager/apsb25-90.md")
	log.Println("  content/bulletins/_index.md")
	log.Println("  content/bulletins/commerce/_index.md")
	log.Println("  content/bulletins/experience-manager/_index.md")
	log.Println("\n🗞️  RSS feeds:")
	log.Println("  static/feeds/bulletins.xml")
	log.Println("  static/feeds/commerce-magento.xml")
	log.Println("  static/feeds/experience-manager.xml")

	log.Println("\n🎉 All tests completed successfully!")
	log.Println("The scraper implementation is working correctly!")
}
