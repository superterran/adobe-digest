package storage

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// PageCache handles caching of successfully fetched HTML pages
type PageCache struct {
	cacheDir string
}

// CachedPage represents a cached HTML page with metadata
type CachedPage struct {
	URL         string    `json:"url"`
	Content     string    `json:"content"`
	FetchedAt   time.Time `json:"fetched_at"`
	StatusCode  int       `json:"status_code"`
	ContentType string    `json:"content_type"`
	Size        int       `json:"size"`
}

// NewPageCache creates a new page cache instance
func NewPageCache(cacheDir string) *PageCache {
	return &PageCache{
		cacheDir: cacheDir,
	}
}

// Init creates the cache directory if it doesn't exist
func (pc *PageCache) Init() error {
	if err := os.MkdirAll(pc.cacheDir, 0755); err != nil {
		return fmt.Errorf("creating cache directory: %w", err)
	}
	return nil
}

// getCacheKey generates a cache key from URL
func (pc *PageCache) getCacheKey(url string) string {
	hash := md5.Sum([]byte(url))
	return fmt.Sprintf("%x.json", hash)
}

// getCachePath returns the full path to the cache file
func (pc *PageCache) getCachePath(url string) string {
	return filepath.Join(pc.cacheDir, pc.getCacheKey(url))
}

// Store saves a page to cache
func (pc *PageCache) Store(url, content string, statusCode int, contentType string) error {
	if err := pc.Init(); err != nil {
		return err
	}

	cached := CachedPage{
		URL:         url,
		Content:     content,
		FetchedAt:   time.Now(),
		StatusCode:  statusCode,
		ContentType: contentType,
		Size:        len(content),
	}

	data, err := json.Marshal(cached)
	if err != nil {
		return fmt.Errorf("marshaling cached page: %w", err)
	}

	cachePath := pc.getCachePath(url)
	if err := os.WriteFile(cachePath, data, 0644); err != nil {
		return fmt.Errorf("writing cache file: %w", err)
	}

	log.Printf("📦 Cached page: %s (%d bytes) -> %s", url, len(content), cachePath)
	return nil
}

// Get retrieves a page from cache
func (pc *PageCache) Get(url string) (*CachedPage, error) {
	cachePath := pc.getCachePath(url)

	data, err := os.ReadFile(cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // Cache miss
		}
		return nil, fmt.Errorf("reading cache file: %w", err)
	}

	var cached CachedPage
	if err := json.Unmarshal(data, &cached); err != nil {
		return nil, fmt.Errorf("unmarshaling cached page: %w", err)
	}

	return &cached, nil
}

// IsValid checks if a cached page is still valid (not expired)
func (pc *PageCache) IsValid(cached *CachedPage, maxAge time.Duration) bool {
	if cached == nil {
		return false
	}
	return time.Since(cached.FetchedAt) < maxAge
}

// GetOrFetch tries to get from cache first, falls back to fetcher function
func (pc *PageCache) GetOrFetch(url string, maxAge time.Duration, fetcher func(string) (string, int, string, error)) (string, error) {
	// Try cache first
	cached, err := pc.Get(url)
	if err != nil {
		log.Printf("⚠️  Cache read error for %s: %v", url, err)
	}

	if pc.IsValid(cached, maxAge) {
		log.Printf("🎯 Cache HIT: %s (fetched %v ago)", url, time.Since(cached.FetchedAt).Round(time.Second))
		return cached.Content, nil
	}

	// Cache miss or expired - fetch fresh
	if cached == nil {
		log.Printf("💨 Cache MISS: %s (not cached)", url)
	} else {
		log.Printf("⏰ Cache EXPIRED: %s (age: %v)", url, time.Since(cached.FetchedAt).Round(time.Second))
	}

	content, statusCode, contentType, err := fetcher(url)
	if err != nil {
		// If we have expired cache and fetch fails, use expired cache as fallback
		if cached != nil {
			log.Printf("🆘 Fetch failed, using STALE cache: %s", url)
			return cached.Content, nil
		}
		return "", fmt.Errorf("fetch failed and no cache available: %w", err)
	}

	// Store successful fetch
	if storeErr := pc.Store(url, content, statusCode, contentType); storeErr != nil {
		log.Printf("⚠️  Failed to cache page: %v", storeErr)
	}

	return content, nil
}

// ListCached returns information about all cached pages
func (pc *PageCache) ListCached() ([]CachedPage, error) {
	if _, err := os.Stat(pc.cacheDir); os.IsNotExist(err) {
		return []CachedPage{}, nil
	}

	files, err := os.ReadDir(pc.cacheDir)
	if err != nil {
		return nil, fmt.Errorf("reading cache directory: %w", err)
	}

	var cached []CachedPage
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			data, err := os.ReadFile(filepath.Join(pc.cacheDir, file.Name()))
			if err != nil {
				continue
			}

			var page CachedPage
			if err := json.Unmarshal(data, &page); err != nil {
				continue
			}

			cached = append(cached, page)
		}
	}

	return cached, nil
}

// Stats returns cache statistics
func (pc *PageCache) Stats() (map[string]interface{}, error) {
	cached, err := pc.ListCached()
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"total_pages":  len(cached),
		"total_size":   0,
		"oldest_fetch": time.Time{},
		"newest_fetch": time.Time{},
		"cache_dir":    pc.cacheDir,
	}

	for _, page := range cached {
		stats["total_size"] = stats["total_size"].(int) + page.Size

		if stats["oldest_fetch"].(time.Time).IsZero() || page.FetchedAt.Before(stats["oldest_fetch"].(time.Time)) {
			stats["oldest_fetch"] = page.FetchedAt
		}

		if stats["newest_fetch"].(time.Time).IsZero() || page.FetchedAt.After(stats["newest_fetch"].(time.Time)) {
			stats["newest_fetch"] = page.FetchedAt
		}
	}

	return stats, nil
}
