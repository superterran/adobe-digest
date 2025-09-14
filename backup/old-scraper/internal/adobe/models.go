package adobe

import (
	"time"
)

// SecurityBulletin represents a complete Adobe security bulletin
type SecurityBulletin struct {
	ID               string            `json:"id" yaml:"id"`                             // APSB25-88
	Title            string            `json:"title" yaml:"title"`                       // Security update available for Adobe Commerce
	Product          string            `json:"product" yaml:"product"`                   // Commerce, AEM, etc.
	URL              string            `json:"url" yaml:"url"`                           // Full URL to bulletin
	PublishedAt      time.Time         `json:"published_at" yaml:"published_at"`         // Original publication date
	UpdatedAt        time.Time         `json:"updated_at" yaml:"updated_at"`             // Last updated date
	Priority         int               `json:"priority" yaml:"priority"`                 // 1-4 (1=Critical, 4=Low)
	PriorityLabel    string            `json:"priority_label" yaml:"priority_label"`     // Critical, Important, etc.
	Summary          string            `json:"summary" yaml:"summary"`                   // Brief description
	Affected         []AffectedVersion `json:"affected" yaml:"affected"`                 // Affected products and versions
	Solutions        []Solution        `json:"solutions" yaml:"solutions"`               // Update recommendations
	Vulnerabilities  []Vulnerability   `json:"vulnerabilities" yaml:"vulnerabilities"`   // CVE details
	Acknowledgements []string          `json:"acknowledgements" yaml:"acknowledgements"` // Credited researchers
	Content          string            `json:"content" yaml:"content"`                   // Full HTML content for processing
}

// Vulnerability represents a specific security vulnerability (CVE)
type Vulnerability struct {
	CVE           string  `json:"cve" yaml:"cve"`                       // CVE-2025-54236
	CWE           string  `json:"cwe" yaml:"cwe"`                       // CWE-20
	CVSS          float64 `json:"cvss" yaml:"cvss"`                     // 9.1
	CVSSVector    string  `json:"cvss_vector" yaml:"cvss_vector"`       // CVSS:3.1/AV:N/AC:L/...
	Type          string  `json:"type" yaml:"type"`                     // Improper Input Validation
	Impact        string  `json:"impact" yaml:"impact"`                 // Security feature bypass
	Severity      string  `json:"severity" yaml:"severity"`             // Critical, High, Medium, Low
	AuthRequired  bool    `json:"auth_required" yaml:"auth_required"`   // Authentication required to exploit
	AdminRequired bool    `json:"admin_required" yaml:"admin_required"` // Admin privileges required
	Exploited     bool    `json:"exploited" yaml:"exploited"`           // Known exploits in wild
}

// AffectedVersion represents products and versions affected by vulnerabilities
type AffectedVersion struct {
	Product   string   `json:"product" yaml:"product"`     // Adobe Commerce, Adobe Commerce B2B, etc.
	Versions  []string `json:"versions" yaml:"versions"`   // Version ranges affected
	Platforms []string `json:"platforms" yaml:"platforms"` // All, Windows, macOS, etc.
	Priority  int      `json:"priority" yaml:"priority"`   // Priority rating for this product
}

// Solution represents update/mitigation information
type Solution struct {
	Product     string `json:"product" yaml:"product"`         // Product name
	Type        string `json:"type" yaml:"type"`               // Hotfix, Update, etc.
	Description string `json:"description" yaml:"description"` // Solution description
	Priority    int    `json:"priority" yaml:"priority"`       // Priority rating
	URL         string `json:"url" yaml:"url"`                 // Link to release notes/download
	Compatible  string `json:"compatible" yaml:"compatible"`   // Compatible versions
}

// BulletinSummary represents a brief summary from product listing pages
type BulletinSummary struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	PublishedAt time.Time `json:"published_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Product     string    `json:"product"`
}

// ProductConfig represents configuration for a specific Adobe product
type ProductConfig struct {
	Name        string `yaml:"name"`         // Internal name (commerce, experience-manager)
	DisplayName string `yaml:"display_name"` // Display name (Adobe Commerce/Magento)
	URL         string `yaml:"url"`          // Product security page URL
	Enabled     bool   `yaml:"enabled"`      // Whether to scrape this product
}

// ScraperConfig represents the main configuration for the scraper
type ScraperConfig struct {
	Products []ProductConfig `yaml:"products"`
	Scraper  struct {
		UserAgent       string `yaml:"user_agent"`
		RateLimit       string `yaml:"rate_limit"`
		Timeout         string `yaml:"timeout"`
		RetryAttempts   int    `yaml:"retry_attempts"`
		ConcurrentLimit int    `yaml:"concurrent_limit"`
	} `yaml:"scraper"`
	Output struct {
		ContentDir string `yaml:"content_dir"`
		RSSFile    string `yaml:"rss_file"`
		CacheFile  string `yaml:"cache_file"`
	} `yaml:"output"`
	RSS struct {
		Title       string `yaml:"title"`
		Description string `yaml:"description"`
		Link        string `yaml:"link"`
		MaxItems    int    `yaml:"max_items"`
	} `yaml:"rss"`
}

// Cache represents the scraper's cache data
type Cache struct {
	LastRun   time.Time                  `json:"last_run"`
	Bulletins map[string]BulletinSummary `json:"bulletins"` // bulletinID -> summary
	UpdatedAt time.Time                  `json:"updated_at"`
}

// GetPriorityLabel returns the human-readable priority label
func (sb *SecurityBulletin) GetPriorityLabel() string {
	switch sb.Priority {
	case 1:
		return "Critical"
	case 2:
		return "Important"
	case 3:
		return "Moderate"
	case 4:
		return "Low"
	default:
		return "Unknown"
	}
}

// GetMaxCVSS returns the highest CVSS score from vulnerabilities
func (sb *SecurityBulletin) GetMaxCVSS() float64 {
	var maxCVSS float64
	for _, vuln := range sb.Vulnerabilities {
		if vuln.CVSS > maxCVSS {
			maxCVSS = vuln.CVSS
		}
	}
	return maxCVSS
}

// GetCVEs returns a slice of all CVE identifiers
func (sb *SecurityBulletin) GetCVEs() []string {
	var cves []string
	for _, vuln := range sb.Vulnerabilities {
		if vuln.CVE != "" {
			cves = append(cves, vuln.CVE)
		}
	}
	return cves
}

// GetAffectedVersionStrings returns a slice of version strings for Hugo front matter
func (sb *SecurityBulletin) GetAffectedVersionStrings() []string {
	var versions []string
	for _, affected := range sb.Affected {
		versions = append(versions, affected.Versions...)
	}
	return versions
}

// HasCriticalVulnerabilities returns true if any vulnerability is marked as Critical
func (sb *SecurityBulletin) HasCriticalVulnerabilities() bool {
	for _, vuln := range sb.Vulnerabilities {
		if vuln.Severity == "Critical" {
			return true
		}
	}
	return false
}
