package main

import (
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/superterran/adobe-digest/internal/scraper"
)

func main() {
	root, err := filepath.Abs(".")
	if err != nil {
		log.Fatal(err)
	}
	cfg := scraper.Config{
		RootDir:     root,
		ContentRoot: filepath.Join(root, "content", "bulletins"),
		UserAgent:   "adobe-digest-bot/0.1 (+https://adobedigest.com)",
		HTTPTimeout: 20 * time.Second,
		MaxLinks:    30,
	}

	s := scraper.New(cfg)

	total := 0
	wrote := 0
	skipped := 0

	// Adobe properties
	jobs := []func() (int, int, error){
		func() (int, int, error) {
			return s.ScrapeAdobeList("https://helpx.adobe.com/security/products/magento.html", "Adobe", "Adobe Commerce")
		},
		func() (int, int, error) {
			return s.ScrapeAdobeList("https://helpx.adobe.com/security/products/experience-manager.html", "Adobe", "AEM")
		},
		func() (int, int, error) {
			return s.ScrapeAdobeList("https://helpx.adobe.com/security/products/experience-platform.html", "Adobe", "AEP")
		},
	}

	for _, job := range jobs {
		w, s, err := job()
		if err != nil {
			log.Printf("warn: job error: %v", err)
		}
		wrote += w
		skipped += s
		total += w + s
	}

	fmt.Printf("Done. total=%d wrote=%d skipped=%d\n", total, wrote, skipped)
}
