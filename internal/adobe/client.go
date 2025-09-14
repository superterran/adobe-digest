package adobe

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client represents an HTTP client configured for Adobe scraping
type Client struct {
	httpClient  *http.Client
	userAgent   string
	rateLimit   time.Duration
	lastRequest time.Time
}

// NewClient creates a new Adobe HTTP client with rate limiting
func NewClient(userAgent string, timeout time.Duration, rateLimit time.Duration) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				MaxIdleConnsPerHost: 2,
				IdleConnTimeout:     30 * time.Second,
			},
		},
		userAgent: userAgent,
		rateLimit: rateLimit,
	}
}

// Get performs an HTTP GET request with rate limiting
func (c *Client) Get(url string) (*http.Response, error) {
	// Rate limiting
	if c.rateLimit > 0 {
		elapsed := time.Since(c.lastRequest)
		if elapsed < c.rateLimit {
			time.Sleep(c.rateLimit - elapsed)
		}
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	// Set headers
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	c.lastRequest = time.Now()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("performing request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	return resp, nil
}

// GetWithRetry performs an HTTP GET request with retry logic
func (c *Client) GetWithRetry(url string, maxRetries int) (*http.Response, error) {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			backoff := time.Duration(attempt) * time.Second
			time.Sleep(backoff)
		}

		resp, err := c.Get(url)
		if err == nil {
			return resp, nil
		}

		lastErr = err
	}

	return nil, fmt.Errorf("failed after %d attempts: %w", maxRetries+1, lastErr)
}

// GetBody performs an HTTP GET request and returns the response body as string
func (c *Client) GetBody(url string) (string, error) {
	resp, err := c.Get(url)
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

// GetBodyWithRetry performs an HTTP GET request with retry and returns the response body as string
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

// Close cleans up the client resources
func (c *Client) Close() {
	if c.httpClient != nil {
		c.httpClient.CloseIdleConnections()
	}
}
