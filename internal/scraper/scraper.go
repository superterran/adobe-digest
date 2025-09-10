package scraper

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Config struct {
	RootDir     string
	ContentRoot string
	UserAgent   string
	HTTPTimeout time.Duration
	MaxLinks    int
}

type Scraper struct {
	cfg Config
	hc  *http.Client
}

func New(cfg Config) *Scraper {
	timeout := cfg.HTTPTimeout
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		// Keep defaults for TLS; disable HTTP/2 for stability
		TLSNextProto:          make(map[string]func(string, *tls.Conn) http.RoundTripper),
		ForceAttemptHTTP2:     false,
		DisableKeepAlives:     true, // avoid lingering connections
		MaxIdleConns:          0,
		IdleConnTimeout:       5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			d := &net.Dialer{Timeout: 10 * time.Second, KeepAlive: 0}
			// Force IPv4
			return d.DialContext(ctx, "tcp4", addr)
		},
	}
	return &Scraper{
		cfg: cfg,
		hc:  &http.Client{Timeout: timeout, Transport: tr},
	}
}

var cveRe = regexp.MustCompile(`CVE-\d{4}-\d{4,7}`)
var sevRe = regexp.MustCompile(`(?i)(Critical|Important|Moderate|Low)`) // Capture first severity word

// ScrapeAdobeList fetches a product bulletin listing page and processes item pages.
func (s *Scraper) ScrapeAdobeList(listURL, vendor, product string) (wrote int, skipped int, err error) {
	// Fetch list with retries and a longer per-request timeout
	fetchList := func() (*goquery.Document, error) {
		var lastErr error
		maxAttempts := 3
		for attempt := 1; attempt <= maxAttempts; attempt++ {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			req, _ := http.NewRequestWithContext(ctx, http.MethodGet, listURL, nil)
			if s.cfg.UserAgent != "" {
				req.Header.Set("User-Agent", s.cfg.UserAgent)
			}
			req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
			req.Header.Set("Accept-Language", "en-US,en;q=0.9")
			req.Header.Set("Upgrade-Insecure-Requests", "1")
			req.Header.Set("Connection", "close") // force h1 per-request
			start := time.Now()
			resp, err := s.hc.Do(req)
			if err != nil {
				lastErr = err
			} else {
				if resp.StatusCode == http.StatusOK {
					doc, e := goquery.NewDocumentFromReader(resp.Body)
					resp.Body.Close()
					if e != nil {
						return nil, fmt.Errorf("parse list: %w", e)
					}
					fmt.Fprintf(os.Stderr, "[list ok] %s (%.1fs)\n", listURL, time.Since(start).Seconds())
					return doc, nil
				}
				// retry on 429 and 5xx
				if resp.StatusCode == 429 || (resp.StatusCode >= 500 && resp.StatusCode <= 599) {
					lastErr = fmt.Errorf("list status %d", resp.StatusCode)
				} else {
					io.Copy(io.Discard, resp.Body)
					resp.Body.Close()
					return nil, fmt.Errorf("list status %d", resp.StatusCode)
				}
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
			}
			// backoff
			sleep := time.Duration(math.Pow(2, float64(attempt-1))) * time.Second
			time.Sleep(sleep)
		}
		return nil, fmt.Errorf("fetch list: %w", lastErr)
	}

	doc, err := fetchList()
	if err != nil {
		return 0, 0, err
	}

	max := s.cfg.MaxLinks
	if max <= 0 {
		max = 20
	}

	type job struct {
		url     string
		vendor  string
		product string
	}
	jobs := make([]job, 0, max)
	doc.Find("a").Each(func(i int, sel *goquery.Selection) {
		if len(jobs) >= max {
			return
		}
		href, ok := sel.Attr("href")
		if !ok {
			return
		}
		text := strings.TrimSpace(sel.Text())
		if strings.Contains(href, "/security/products/") && strings.Contains(strings.ToLower(text), "apsb") {
			u := href
			if strings.HasPrefix(href, "/") {
				u = originOf(listURL) + href
			}
			jobs = append(jobs, job{url: u, vendor: vendor, product: product})
		}
	})

	// Concurrency: process up to 5 at a time
	sem := make(chan struct{}, 5)
	results := make(chan struct{ w, s int }, len(jobs))

	for _, j := range jobs {
		sem <- struct{}{}
		go func(j job) {
			defer func() { <-sem }()
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			start := time.Now()
			w, s, e := s.processAdobeBulletinCtx(ctx, j.url, j.vendor, j.product)
			elapsed := time.Since(start)
			if e != nil {
				fmt.Fprintf(os.Stderr, "[warn] %s: %v (%.1fs)\n", j.url, e, elapsed.Seconds())
			} else {
				fmt.Fprintf(os.Stderr, "[ok] %s: wrote=%d skipped=%d (%.1fs)\n", j.url, w, s, elapsed.Seconds())
			}
			results <- struct{ w, s int }{w, s}
		}(j)
	}

	// Wait for all jobs
	for i := 0; i < len(jobs); i++ {
		r := <-results
		wrote += r.w
		skipped += r.s
	}

	return wrote, skipped, nil
}

func originOf(u string) string {
	// very small origin extractor
	// expects schema://host...
	i := strings.Index(u, "://")
	if i == -1 {
		return ""
	}
	j := strings.Index(u[i+3:], "/")
	if j == -1 {
		return u
	}
	return u[:i+3+j]
}

func (s *Scraper) processAdobeBulletinCtx(ctx context.Context, url, vendor, product string) (wrote int, skipped int, err error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if s.cfg.UserAgent != "" {
		req.Header.Set("User-Agent", s.cfg.UserAgent)
	}
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	resp, err := s.hc.Do(req)
	if err != nil {
		return 0, 0, fmt.Errorf("fetch bulletin: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, 0, fmt.Errorf("bulletin status %d", resp.StatusCode)
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, 0, err
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(bodyBytes)))
	if err != nil {
		return 0, 0, err
	}

	title := strings.TrimSpace(doc.Find("h1").First().Text())
	if title == "" {
		title = strings.TrimSpace(doc.Find("title").Text())
	}

	// publication date heuristics
	date := time.Now().UTC()
	if t, ok := doc.Find("time").Attr("datetime"); ok {
		if d, e := time.Parse(time.RFC3339, t); e == nil {
			date = d
		}
	}

	// Extract CVEs and severity from text
	text := doc.Text()
	cveSet := map[string]struct{}{}
	for _, m := range cveRe.FindAllString(text, -1) {
		cveSet[m] = struct{}{}
	}
	cves := make([]string, 0, len(cveSet))
	for k := range cveSet {
		cves = append(cves, k)
	}

	severity := "unknown"
	if m := sevRe.FindStringSubmatch(text); len(m) > 1 {
		severity = strings.Title(strings.ToLower(m[1]))
	}

	// write markdown
	if err := s.writeMarkdown(vendor, product, title, date, severity, cves, url); err != nil {
		return 0, 0, err
	}
	return 1, 0, nil
}

func (s *Scraper) writeMarkdown(vendor, product, title string, date time.Time, severity string, cves []string, sourceURL string) error {
	year := date.UTC().Year()
	base := slugify(fmt.Sprintf("%s-%s-%s", vendor, product, title))
	id := shortHash(fmt.Sprintf("%s|%s|%s", vendor, product, sourceURL))
	slug := fmt.Sprintf("%s-%s", base, id)

	dir := filepath.Join(s.cfg.ContentRoot, strings.ToLower(vendor), fmt.Sprintf("%d", year))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	fp := filepath.Join(dir, slug+".md")

	if _, err := os.Stat(fp); err == nil {
		// already exists
		return nil
	}

	fm := buildFrontMatter(map[string]interface{}{
		"title":      title,
		"date":       date.Format(time.RFC3339),
		"vendor":     vendor,
		"product":    product,
		"severity":   severity,
		"cves":       cves,
		"source_url": sourceURL,
		"draft":      false,
		"tags":       []string{vendor, product, "security-bulletin"},
	})

	body := strings.Join([]string{
		fmt.Sprintf("Source: %s", sourceURL),
		"",
		"This page is an automated capture of an Adobe security bulletin. Refer to the source for authoritative details.",
		"",
	}, "\n")

	content := fm + "\n" + body + "\n"
	return os.WriteFile(fp, []byte(content), 0o644)
}

func shortHash(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])[:12]
}

func slugify(s string) string {
	s = strings.ToLower(s)
	// replace non-alnum with dashes
	b := strings.Builder{}
	prevDash := false
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			prevDash = false
			continue
		}
		if !prevDash {
			b.WriteRune('-')
			prevDash = true
		}
	}
	out := b.String()
	out = strings.Trim(out, "-")
	out = strings.ReplaceAll(out, "--", "-")
	return out
}

func buildFrontMatter(m map[string]interface{}) string {
	// naive YAML for simple types
	var b strings.Builder
	b.WriteString("---\n")
	for k, v := range m {
		switch t := v.(type) {
		case string:
			b.WriteString(fmt.Sprintf("%s: %q\n", k, t))
		case bool:
			b.WriteString(fmt.Sprintf("%s: %v\n", k, t))
		case []string:
			b.WriteString(fmt.Sprintf("%s: [", k))
			for i, s := range t {
				if i > 0 {
					b.WriteString(", ")
				}
				b.WriteString(fmt.Sprintf("%q", s))
			}
			b.WriteString("]\n")
		default:
			b.WriteString(fmt.Sprintf("%s: %v\n", k, v))
		}
	}
	b.WriteString("---")
	return b.String()
}
