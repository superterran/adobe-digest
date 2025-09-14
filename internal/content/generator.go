package content

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/superterran/adobe-digest/internal/adobe"
)

// Generator handles the generation of Hugo content from security bulletins
type Generator struct {
	contentDir string
	templates  map[string]*template.Template
}

// NewGenerator creates a new content generator
func NewGenerator(contentDir string) (*Generator, error) {
	g := &Generator{
		contentDir: contentDir,
		templates:  make(map[string]*template.Template),
	}

	// Load templates
	if err := g.loadTemplates(); err != nil {
		return nil, fmt.Errorf("loading templates: %w", err)
	}

	return g, nil
}

// loadTemplates loads all content templates
func (g *Generator) loadTemplates() error {
	templateDir := "internal/content/templates"

	// Custom template functions
	funcMap := template.FuncMap{
		"lower":   strings.ToLower,
		"replace": strings.ReplaceAll,
		"cleanProductSlug": func(product string) string {
			slug := strings.ToLower(product)
			slug = strings.ReplaceAll(slug, "adobe ", "")
			slug = strings.ReplaceAll(slug, " ", "-")
			return slug
		},
	}

	// Load bulletin template
	bulletinTmpl, err := template.New("bulletin.md.tmpl").Funcs(funcMap).ParseFiles(filepath.Join(templateDir, "bulletin.md.tmpl"))
	if err != nil {
		return fmt.Errorf("loading bulletin template: %w", err)
	}
	g.templates["bulletin"] = bulletinTmpl

	// Load product index template
	productTmpl, err := template.New("product-index.md.tmpl").Funcs(funcMap).ParseFiles(filepath.Join(templateDir, "product-index.md.tmpl"))
	if err != nil {
		return fmt.Errorf("loading product index template: %w", err)
	}
	g.templates["product-index"] = productTmpl

	// Load bulletins index template
	indexTmpl, err := template.New("bulletins-index.md.tmpl").Funcs(funcMap).ParseFiles(filepath.Join(templateDir, "bulletins-index.md.tmpl"))
	if err != nil {
		return fmt.Errorf("loading bulletins index template: %w", err)
	}
	g.templates["bulletins-index"] = indexTmpl

	return nil
}

// GenerateBulletin creates a markdown file for a security bulletin
func (g *Generator) GenerateBulletin(bulletin *adobe.SecurityBulletin) error {
	productSlug := getProductSlug(bulletin.Product)
	fileName := strings.ToLower(bulletin.ID) + ".md"

	// Create product directory if it doesn't exist
	productDir := filepath.Join(g.contentDir, "bulletins", productSlug)
	if err := os.MkdirAll(productDir, 0755); err != nil {
		return fmt.Errorf("creating product directory: %w", err)
	}

	// Generate content
	filePath := filepath.Join(productDir, fileName)
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("creating bulletin file: %w", err)
	}
	defer file.Close()

	tmpl, exists := g.templates["bulletin"]
	if !exists {
		return fmt.Errorf("bulletin template not found")
	}

	if err := tmpl.Execute(file, bulletin); err != nil {
		return fmt.Errorf("executing bulletin template: %w", err)
	}

	return nil
}

// GenerateProductIndex creates an index file for a product's bulletins
func (g *Generator) GenerateProductIndex(product adobe.ProductConfig, bulletins []adobe.BulletinSummary) error {
	productSlug := getProductSlug(product.DisplayName)
	productDir := filepath.Join(g.contentDir, "bulletins", productSlug)

	if err := os.MkdirAll(productDir, 0755); err != nil {
		return fmt.Errorf("creating product directory: %w", err)
	}

	filePath := filepath.Join(productDir, "_index.md")
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("creating product index file: %w", err)
	}
	defer file.Close()

	data := struct {
		DisplayName string
		Bulletins   []BulletinSummaryWithURL
	}{
		DisplayName: product.DisplayName,
		Bulletins:   convertBulletinSummaries(bulletins, productSlug),
	}

	tmpl, exists := g.templates["product-index"]
	if !exists {
		return fmt.Errorf("product index template not found")
	}

	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("executing product index template: %w", err)
	}

	return nil
}

// GenerateBulletinsIndex creates the main bulletins index page
func (g *Generator) GenerateBulletinsIndex(recentBulletins []adobe.BulletinSummary) error {
	bulletinsDir := filepath.Join(g.contentDir, "bulletins")
	if err := os.MkdirAll(bulletinsDir, 0755); err != nil {
		return fmt.Errorf("creating bulletins directory: %w", err)
	}

	filePath := filepath.Join(bulletinsDir, "_index.md")
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("creating bulletins index file: %w", err)
	}
	defer file.Close()

	data := struct {
		RecentBulletins []BulletinSummaryWithLocalURL
	}{
		RecentBulletins: convertToLocalURLs(recentBulletins),
	}

	tmpl, exists := g.templates["bulletins-index"]
	if !exists {
		return fmt.Errorf("bulletins index template not found")
	}

	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("executing bulletins index template: %w", err)
	}

	return nil
}

// BulletinSummaryWithURL extends BulletinSummary with local URL
type BulletinSummaryWithURL struct {
	adobe.BulletinSummary
	LocalURL string
}

// BulletinSummaryWithLocalURL extends BulletinSummary with local URL and priority
type BulletinSummaryWithLocalURL struct {
	adobe.BulletinSummary
	LocalURL      string
	PriorityLabel string
}

// Helper functions

func getProductSlug(productName string) string {
	slug := strings.ToLower(productName)
	slug = strings.ReplaceAll(slug, "adobe ", "")
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "/", "-")
	return slug
}

func convertBulletinSummaries(summaries []adobe.BulletinSummary, productSlug string) []BulletinSummaryWithURL {
	var result []BulletinSummaryWithURL
	for _, summary := range summaries {
		result = append(result, BulletinSummaryWithURL{
			BulletinSummary: summary,
			LocalURL:        fmt.Sprintf("/bulletins/%s/%s/", productSlug, strings.ToLower(summary.ID)),
		})
	}
	return result
}

func convertToLocalURLs(summaries []adobe.BulletinSummary) []BulletinSummaryWithLocalURL {
	var result []BulletinSummaryWithLocalURL
	for _, summary := range summaries {
		productSlug := getProductSlug(summary.Product)
		priority := getPriorityFromTitle(summary.Title) // Simple heuristic

		result = append(result, BulletinSummaryWithLocalURL{
			BulletinSummary: summary,
			LocalURL:        fmt.Sprintf("/bulletins/%s/%s/", productSlug, strings.ToLower(summary.ID)),
			PriorityLabel:   priority,
		})
	}
	return result
}

func getPriorityFromTitle(title string) string {
	// Simple heuristic based on common priority patterns
	title = strings.ToLower(title)
	if strings.Contains(title, "critical") {
		return "Critical"
	} else if strings.Contains(title, "important") {
		return "Important"
	} else if strings.Contains(title, "moderate") {
		return "Moderate"
	} else if strings.Contains(title, "low") {
		return "Low"
	}
	return "Important" // Default fallback
}

// GenerateAll generates all content for the given bulletins and products
func (g *Generator) GenerateAll(bulletinsByProduct map[string][]adobe.SecurityBulletin, products []adobe.ProductConfig) error {
	var allSummaries []adobe.BulletinSummary

	// Generate individual bulletin pages and product indexes
	for productName, bulletins := range bulletinsByProduct {
		var productSummaries []adobe.BulletinSummary

		// Generate individual bulletin pages
		for _, bulletin := range bulletins {
			if err := g.GenerateBulletin(&bulletin); err != nil {
				return fmt.Errorf("generating bulletin %s: %w", bulletin.ID, err)
			}

			// Convert to summary for indexes
			summary := adobe.BulletinSummary{
				ID:          bulletin.ID,
				Title:       bulletin.Title,
				URL:         bulletin.URL,
				PublishedAt: bulletin.PublishedAt,
				UpdatedAt:   bulletin.UpdatedAt,
				Product:     bulletin.Product,
			}
			productSummaries = append(productSummaries, summary)
			allSummaries = append(allSummaries, summary)
		}

		// Find matching product config
		var productConfig adobe.ProductConfig
		for _, p := range products {
			if strings.EqualFold(p.DisplayName, productName) ||
				strings.Contains(strings.ToLower(productName), strings.ToLower(p.DisplayName)) {
				productConfig = p
				break
			}
		}

		// Generate product index
		if productConfig.DisplayName != "" {
			if err := g.GenerateProductIndex(productConfig, productSummaries); err != nil {
				return fmt.Errorf("generating product index for %s: %w", productName, err)
			}
		}
	}

	// Sort all summaries by date (newest first) and take top 10
	if len(allSummaries) > 1 {
		for i := 0; i < len(allSummaries)-1; i++ {
			for j := i + 1; j < len(allSummaries); j++ {
				if allSummaries[i].PublishedAt.Before(allSummaries[j].PublishedAt) {
					allSummaries[i], allSummaries[j] = allSummaries[j], allSummaries[i]
				}
			}
		}
	}

	// Take top 10 for recent bulletins
	recentCount := 10
	if len(allSummaries) < recentCount {
		recentCount = len(allSummaries)
	}
	recentBulletins := allSummaries[:recentCount]

	// Generate main bulletins index
	if err := g.GenerateBulletinsIndex(recentBulletins); err != nil {
		return fmt.Errorf("generating bulletins index: %w", err)
	}

	return nil
}
