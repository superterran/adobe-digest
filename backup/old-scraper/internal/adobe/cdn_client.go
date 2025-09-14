package adobe

import (
	"crypto/tls"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"time"
)

// Client represents an HTTP client configured for Adobe scraping with CDN evasion
type Client struct {
	httpClient  *http.Client
	userAgents  []string
	rateLimit   time.Duration
	lastRequest time.Time
}

// NewClient creates a new Adobe HTTP client with CDN evasion techniques
func NewClientWithCDNEvasion(timeout time.Duration, rateLimit time.Duration) *Client {
	// Realistic browser user agents - rotate to avoid detection
	userAgents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:122.0) Gecko/20100101 Firefox/122.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:122.0) Gecko/20100101 Firefox/122.0",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36",
	}

	// Create transport with CDN-friendly settings
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   20 * time.Second, // Longer timeout for CDN
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   15 * time.Second,
		ResponseHeaderTimeout: 60 * time.Second, // Much longer for CDN delays
		ExpectContinueTimeout: 2 * time.Second,
		MaxIdleConns:          5, // Fewer connections to appear less bot-like
		MaxIdleConnsPerHost:   1, // Only 1 connection per host
		IdleConnTimeout:       120 * time.Second,
		DisableCompression:    false,
		ForceAttemptHTTP2:     false, // HTTP/1.1 only to avoid stream issues

		// Mimic browser TLS fingerprint
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
			MaxVersion: tls.VersionTLS13,
			// Use common browser cipher suites
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256,
				tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			},
			PreferServerCipherSuites: false,
			InsecureSkipVerify:       false,
		},
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Follow redirects but limit to 10 and preserve headers
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			// Copy headers from original request
			for key, values := range via[0].Header {
				for _, value := range values {
					req.Header.Add(key, value)
				}
			}
			return nil
		},
	}

	return &Client{
		httpClient: client,
		userAgents: userAgents,
		rateLimit:  rateLimit,
	}
}

// Get performs an HTTP GET request with CDN evasion techniques
func (c *Client) Get(url string) (*http.Response, error) {
	// Enhanced rate limiting with jitter
	if !c.lastRequest.IsZero() {
		elapsed := time.Since(c.lastRequest)
		minDelay := c.rateLimit + time.Duration(rand.Intn(3000))*time.Millisecond // 0-3s jitter
		if elapsed < minDelay {
			time.Sleep(minDelay - elapsed)
		}
	}
	c.lastRequest = time.Now()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	// Add comprehensive browser headers to fool CDN
	c.addRealisticHeaders(req)

	fmt.Printf("🌐 Making CDN-evading request to: %s\n", url)
	fmt.Printf("🎭 Using User-Agent: %s\n", req.Header.Get("User-Agent")[:50]+"...")

	start := time.Now()
	resp, err := c.httpClient.Do(req)
	elapsed := time.Since(start)

	fmt.Printf("⏱️  Request completed in %v\n", elapsed)

	if err != nil {
		fmt.Printf("❌ Request failed: %v\n", err)
		return nil, fmt.Errorf("performing request: %w", err)
	}

	if resp == nil {
		fmt.Printf("❌ Response is nil\n")
		return nil, fmt.Errorf("nil response received")
	}

	fmt.Printf("📊 Response status: %s\n", resp.Status)
	fmt.Printf("🔧 Content-Type: %s\n", resp.Header.Get("Content-Type"))
	fmt.Printf("📏 Content-Length: %s\n", resp.Header.Get("Content-Length"))

	// Check for CDN blocking indicators
	if akamai := resp.Header.Get("akamai-grn"); akamai != "" {
		fmt.Printf("☁️  Akamai edge: %s\n", akamai)
	}

	if resp.StatusCode == 403 {
		fmt.Printf("🚫 CDN blocking detected (403 Forbidden)\n")
	} else if resp.StatusCode == 503 {
		fmt.Printf("🚫 CDN rate limiting detected (503 Service Unavailable)\n")
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		resp.Body.Close()
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	return resp, nil
}

// addRealisticHeaders adds comprehensive browser headers to fool CDNs
func (c *Client) addRealisticHeaders(req *http.Request) {
	// Rotate user agent
	userAgent := c.userAgents[rand.Intn(len(c.userAgents))]
	req.Header.Set("User-Agent", userAgent)

	// Add realistic browser headers in correct order
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("DNT", "1")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	// Modern browser security headers
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Sec-CH-UA", `"Not_A Brand";v="8", "Chromium";v="120", "Google Chrome";v="120"`)
	req.Header.Set("Sec-CH-UA-Mobile", "?0")
	req.Header.Set("Sec-CH-UA-Platform", `"Windows"`)

	// Cache control
	req.Header.Set("Cache-Control", "max-age=0")

	// Add referer for non-root requests (simulate browsing behavior)
	if !strings.HasSuffix(req.URL.Path, "/") && req.URL.Path != "/security.html" {
		req.Header.Set("Referer", "https://helpx.adobe.com/security.html")
	}
}

// GetWithRetry performs an HTTP GET request with retry logic and exponential backoff
func (c *Client) GetWithRetry(url string, maxRetries int) (*http.Response, error) {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff with jitter for CDN retry
			backoff := time.Duration(attempt*attempt) * time.Second
			jitter := time.Duration(rand.Intn(2000)) * time.Millisecond
			sleepTime := backoff + jitter

			fmt.Printf("🔄 Retry attempt %d/%d after %v\n", attempt, maxRetries, sleepTime)
			time.Sleep(sleepTime)
		}

		resp, err := c.Get(url)
		if err == nil {
			return resp, nil
		}

		lastErr = err
		fmt.Printf("⚠️  Attempt %d failed: %v\n", attempt+1, err)

		// Don't retry on certain errors
		if strings.Contains(err.Error(), "403") || strings.Contains(err.Error(), "blocked") {
			fmt.Printf("🚫 CDN blocking detected, not retrying\n")
			break
		}
	}

	return nil, fmt.Errorf("failed after %d attempts: %w", maxRetries+1, lastErr)
}

// ReadBody safely reads and closes the response body
func (c *Client) ReadBody(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	fmt.Printf("📄 Read %d bytes from response\n", len(body))
	return body, nil
}

// GetBodyWithRetry fetches a URL and returns the response body as string
func (c *Client) GetBodyWithRetry(url string, maxRetries int) (string, error) {
	resp, err := c.GetWithRetry(url, maxRetries)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response body: %w", err)
	}

	return string(body), nil
}
