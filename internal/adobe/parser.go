package adobe

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var (
	// Regular expressions for parsing
	bulletinIDRegex = regexp.MustCompile(`APSB\d{2}-\d{2,3}`)
	cveRegex        = regexp.MustCompile(`CVE-\d{4}-\d{4,7}`)
	cvssRegex       = regexp.MustCompile(`(\d+\.\d+)`)
	dateRegex       = regexp.MustCompile(`(\d{1,2})/(\d{1,2})/(\d{4})`)
	priorityRegex   = regexp.MustCompile(`Priority:\s*(\d+)`)
)

// Parser handles parsing of Adobe security bulletin pages
type Parser struct {
	client *Client
}

// NewParser creates a new parser with the given HTTP client
func NewParser(client *Client) *Parser {
	return &Parser{
		client: client,
	}
}

// ParseProductPage extracts bulletin summaries from a product's security page
func (p *Parser) ParseProductPage(url, productName string) ([]BulletinSummary, error) {
	html, err := p.client.GetBodyWithRetry(url, 3)
	if err != nil {
		return nil, fmt.Errorf("fetching product page: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("parsing HTML: %w", err)
	}

	var summaries []BulletinSummary

	// Find tables containing bulletin information
	doc.Find("table").Each(func(i int, table *goquery.Selection) {
		table.Find("tr").Each(func(j int, row *goquery.Selection) {
			// Skip header rows
			if row.Find("th").Length() > 0 {
				return
			}

			cells := row.Find("td")
			if cells.Length() < 3 {
				return
			}

			titleCell := cells.Eq(0)
			publishedCell := cells.Eq(1)
			updatedCell := cells.Eq(2)

			// Extract bulletin ID and title
			titleText := strings.TrimSpace(titleCell.Text())
			bulletinID := bulletinIDRegex.FindString(titleText)
			if bulletinID == "" {
				return
			}

			// Extract URL - look for links in the title cell
			var bulletinURL string
			titleCell.Find("a").Each(func(k int, link *goquery.Selection) {
				if href, exists := link.Attr("href"); exists {
					if strings.Contains(href, bulletinID) {
						if strings.HasPrefix(href, "http") {
							bulletinURL = href
						} else {
							bulletinURL = "https://helpx.adobe.com" + href
						}
					}
				}
			})

			// If no direct link found, construct URL
			if bulletinURL == "" {
				productSlug := getProductSlug(productName)
				bulletinURL = fmt.Sprintf("https://helpx.adobe.com/security/products/%s/%s.html",
					productSlug, strings.ToLower(bulletinID))
			}

			// Parse dates
			publishedAt := parseDate(strings.TrimSpace(publishedCell.Text()))
			updatedAt := parseDate(strings.TrimSpace(updatedCell.Text()))

			summary := BulletinSummary{
				ID:          bulletinID,
				Title:       titleText,
				URL:         bulletinURL,
				PublishedAt: publishedAt,
				UpdatedAt:   updatedAt,
				Product:     productName,
			}

			summaries = append(summaries, summary)
		})
	})

	return summaries, nil
}

// ParseBulletin extracts detailed information from a individual bulletin page
func (p *Parser) ParseBulletin(url string) (*SecurityBulletin, error) {
	html, err := p.client.GetBodyWithRetry(url, 3)
	if err != nil {
		return nil, fmt.Errorf("fetching bulletin page: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("parsing HTML: %w", err)
	}

	bulletin := &SecurityBulletin{
		URL:     url,
		Content: html,
	}

	// Extract bulletin ID from URL or title
	bulletin.ID = bulletinIDRegex.FindString(url)
	if bulletin.ID == "" {
		bulletin.ID = bulletinIDRegex.FindString(doc.Find("title").Text())
	}

	// Extract title
	bulletin.Title = strings.TrimSpace(doc.Find("h1").First().Text())
	if bulletin.Title == "" {
		bulletin.Title = strings.TrimSpace(doc.Find("title").Text())
	}

	// Extract product name from title
	bulletin.Product = extractProductFromTitle(bulletin.Title)

	// Extract dates and priority from the main table
	p.extractMetadata(doc, bulletin)

	// Extract summary
	bulletin.Summary = p.extractSummary(doc)

	// Extract vulnerabilities
	bulletin.Vulnerabilities = p.extractVulnerabilities(doc)

	// Extract affected versions
	bulletin.Affected = p.extractAffectedVersions(doc)

	// Extract solutions
	bulletin.Solutions = p.extractSolutions(doc)

	// Extract acknowledgements
	bulletin.Acknowledgements = p.extractAcknowledgements(doc)

	// Set priority label
	bulletin.PriorityLabel = bulletin.GetPriorityLabel()

	return bulletin, nil
}

// extractMetadata extracts dates and priority from the bulletin metadata table
func (p *Parser) extractMetadata(doc *goquery.Document, bulletin *SecurityBulletin) {
	// Look for the main metadata table (usually the first table)
	doc.Find("table").First().Find("tr").Each(func(i int, row *goquery.Selection) {
		cells := row.Find("td")
		if cells.Length() >= 3 {
			bulletinIDCell := cells.Eq(0)
			dateCell := cells.Eq(1)
			priorityCell := cells.Eq(2)

			// Check if this row contains our bulletin
			if strings.Contains(bulletinIDCell.Text(), bulletin.ID) {
				// Parse date
				dateText := strings.TrimSpace(dateCell.Text())
				bulletin.PublishedAt = parseDate(dateText)
				bulletin.UpdatedAt = bulletin.PublishedAt // Default to same as published

				// Parse priority
				priorityText := strings.TrimSpace(priorityCell.Text())
				if priority, err := strconv.Atoi(priorityText); err == nil {
					bulletin.Priority = priority
				}
			}
		}
	})

	// Also check for "Last updated on" text
	doc.Find("p, div, span").Each(func(i int, elem *goquery.Selection) {
		text := strings.TrimSpace(elem.Text())
		if strings.Contains(strings.ToLower(text), "last updated on") {
			dateStr := strings.TrimSpace(strings.Split(text, "Last updated on")[1])
			if updated := parseDate(dateStr); !updated.IsZero() {
				bulletin.UpdatedAt = updated
			}
		}
	})
}

// extractSummary extracts the summary section
func (p *Parser) extractSummary(doc *goquery.Document) string {
	// Look for Summary section
	var summary string

	doc.Find("h2, h3").Each(func(i int, heading *goquery.Selection) {
		if strings.Contains(strings.ToLower(heading.Text()), "summary") {
			// Get the next sibling elements until we hit another heading
			summary = extractTextUntilNextHeading(heading)
		}
	})

	if summary == "" {
		// Fallback: look for the first paragraph after the title
		doc.Find("h1").First().NextAllFiltered("p").First().Each(func(i int, p *goquery.Selection) {
			summary = strings.TrimSpace(p.Text())
		})
	}

	return cleanText(summary)
}

// extractVulnerabilities extracts vulnerability details from tables
func (p *Parser) extractVulnerabilities(doc *goquery.Document) []Vulnerability {
	var vulnerabilities []Vulnerability

	// Look for vulnerability details table
	doc.Find("h2, h3").Each(func(i int, heading *goquery.Selection) {
		if strings.Contains(strings.ToLower(heading.Text()), "vulnerability") ||
			strings.Contains(strings.ToLower(heading.Text()), "details") {

			// Find the next table after this heading
			heading.NextAllFiltered("table").First().Find("tr").Each(func(j int, row *goquery.Selection) {
				cells := row.Find("td")
				if cells.Length() >= 6 {
					vuln := Vulnerability{}

					// Common table structure: Type | Impact | Severity | Auth | Admin | CVSS | Vector | CVE
					if cells.Length() >= 8 {
						vuln.Type = cleanText(cells.Eq(0).Text())
						vuln.Impact = cleanText(cells.Eq(1).Text())
						vuln.Severity = cleanText(cells.Eq(2).Text())
						vuln.AuthRequired = strings.ToLower(cells.Eq(3).Text()) == "yes"
						vuln.AdminRequired = strings.ToLower(cells.Eq(4).Text()) == "yes"

						cvssText := cells.Eq(5).Text()
						if cvss, err := strconv.ParseFloat(cvssRegex.FindString(cvssText), 64); err == nil {
							vuln.CVSS = cvss
						}

						vuln.CVSSVector = cleanText(cells.Eq(6).Text())
						vuln.CVE = cveRegex.FindString(cells.Eq(7).Text())
					}

					if vuln.CVE != "" {
						vulnerabilities = append(vulnerabilities, vuln)
					}
				}
			})
		}
	})

	// Also scan the entire document for CVE references
	cves := cveRegex.FindAllString(doc.Text(), -1)
	for _, cve := range cves {
		found := false
		for _, existing := range vulnerabilities {
			if existing.CVE == cve {
				found = true
				break
			}
		}
		if !found {
			vulnerabilities = append(vulnerabilities, Vulnerability{
				CVE: cve,
			})
		}
	}

	return vulnerabilities
}

// extractAffectedVersions extracts affected product versions
func (p *Parser) extractAffectedVersions(doc *goquery.Document) []AffectedVersion {
	var affected []AffectedVersion

	// Look for Affected Versions section
	doc.Find("h2, h3").Each(func(i int, heading *goquery.Selection) {
		if strings.Contains(strings.ToLower(heading.Text()), "affected") {
			heading.NextAllFiltered("table").First().Find("tr").Each(func(j int, row *goquery.Selection) {
				cells := row.Find("td")
				if cells.Length() >= 3 {
					product := cleanText(cells.Eq(0).Text())
					versions := cleanText(cells.Eq(1).Text())

					if product != "" && versions != "" {
						priority := 2 // Default priority
						if cells.Length() >= 4 {
							if p, err := strconv.Atoi(cleanText(cells.Eq(2).Text())); err == nil {
								priority = p
							}
						}

						affected = append(affected, AffectedVersion{
							Product:   product,
							Versions:  []string{versions},
							Platforms: []string{"All"},
							Priority:  priority,
						})
					}
				}
			})
		}
	})

	return affected
}

// extractSolutions extracts solution information
func (p *Parser) extractSolutions(doc *goquery.Document) []Solution {
	var solutions []Solution

	// Look for Solution section
	doc.Find("h2, h3").Each(func(i int, heading *goquery.Selection) {
		if strings.Contains(strings.ToLower(heading.Text()), "solution") {
			heading.NextAllFiltered("table").First().Find("tr").Each(func(j int, row *goquery.Selection) {
				cells := row.Find("td")
				if cells.Length() >= 3 {
					product := cleanText(cells.Eq(0).Text())
					description := cleanText(cells.Eq(1).Text())

					if product != "" && description != "" {
						solutions = append(solutions, Solution{
							Product:     product,
							Type:        "Update",
							Description: description,
							Priority:    2,
						})
					}
				}
			})
		}
	})

	return solutions
}

// extractAcknowledgements extracts researcher acknowledgements
func (p *Parser) extractAcknowledgements(doc *goquery.Document) []string {
	var acknowledgements []string

	// Look for Acknowledgements section
	doc.Find("h2, h3").Each(func(i int, heading *goquery.Selection) {
		if strings.Contains(strings.ToLower(heading.Text()), "acknowledgement") {
			text := extractTextUntilNextHeading(heading)

			// Split by bullet points and clean up
			lines := strings.Split(text, "•")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				line = strings.TrimPrefix(line, "•")
				line = strings.TrimSpace(line)

				if line != "" && !strings.Contains(strings.ToLower(line), "adobe") {
					// Extract researcher name (everything before " -- ")
					if parts := strings.Split(line, " -- "); len(parts) > 0 {
						researcher := strings.TrimSpace(parts[0])
						if researcher != "" {
							acknowledgements = append(acknowledgements, researcher)
						}
					}
				}
			}
		}
	})

	return acknowledgements
}

// Helper functions

func getProductSlug(productName string) string {
	switch strings.ToLower(productName) {
	case "adobe experience manager", "aem":
		return "experience-manager"
	case "adobe commerce", "magento", "adobe commerce/magento":
		return "magento"
	case "adobe experience platform", "aep":
		return "experience-platform"
	default:
		return strings.ToLower(strings.ReplaceAll(productName, " ", "-"))
	}
}

func extractProductFromTitle(title string) string {
	title = strings.ToLower(title)
	if strings.Contains(title, "commerce") || strings.Contains(title, "magento") {
		return "Adobe Commerce"
	} else if strings.Contains(title, "experience manager") || strings.Contains(title, "aem") {
		return "Adobe Experience Manager"
	} else if strings.Contains(title, "experience platform") || strings.Contains(title, "aep") {
		return "Adobe Experience Platform"
	}
	return "Adobe Product"
}

func parseDate(dateStr string) time.Time {
	dateStr = strings.TrimSpace(dateStr)

	// Try common formats
	formats := []string{
		"01/02/2006",
		"1/2/2006",
		"02/01/2006",
		"2/1/2006",
		"Jan 2, 2006",
		"January 2, 2006",
		"2006-01-02",
		"Mon, 02 Jan 2006",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t
		}
	}

	// Try regex extraction for MM/DD/YYYY
	matches := dateRegex.FindStringSubmatch(dateStr)
	if len(matches) == 4 {
		month, _ := strconv.Atoi(matches[1])
		day, _ := strconv.Atoi(matches[2])
		year, _ := strconv.Atoi(matches[3])
		return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	}

	log.Printf("Warning: Could not parse date: %s", dateStr)
	return time.Time{}
}

func extractTextUntilNextHeading(elem *goquery.Selection) string {
	var text strings.Builder

	elem.NextAll().Each(func(i int, sibling *goquery.Selection) {
		// Stop at next heading
		if sibling.Is("h1, h2, h3, h4, h5, h6") {
			return
		}

		siblingText := strings.TrimSpace(sibling.Text())
		if siblingText != "" {
			text.WriteString(siblingText)
			text.WriteString(" ")
		}
	})

	return strings.TrimSpace(text.String())
}

func cleanText(text string) string {
	text = strings.TrimSpace(text)
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\t", " ")

	// Remove multiple spaces
	for strings.Contains(text, "  ") {
		text = strings.ReplaceAll(text, "  ", " ")
	}

	return text
}
