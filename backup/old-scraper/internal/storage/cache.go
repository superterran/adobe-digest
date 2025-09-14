package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/superterran/adobe-digest/internal/adobe"
)

// Cache handles persistent storage of scraper state
type Cache struct {
	filePath string
	data     *adobe.Cache
}

// NewCache creates a new cache instance
func NewCache(filePath string) *Cache {
	return &Cache{
		filePath: filePath,
		data: &adobe.Cache{
			Bulletins: make(map[string]adobe.BulletinSummary),
		},
	}
}

// Load reads the cache from disk
func (c *Cache) Load() error {
	if _, err := os.Stat(c.filePath); os.IsNotExist(err) {
		// Cache file doesn't exist, start with empty cache
		c.data.UpdatedAt = time.Now()
		return nil
	}

	data, err := os.ReadFile(c.filePath)
	if err != nil {
		return fmt.Errorf("reading cache file: %w", err)
	}

	if err := json.Unmarshal(data, c.data); err != nil {
		return fmt.Errorf("unmarshaling cache data: %w", err)
	}

	return nil
}

// Save writes the cache to disk
func (c *Cache) Save() error {
	c.data.UpdatedAt = time.Now()

	data, err := json.MarshalIndent(c.data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling cache data: %w", err)
	}

	if err := os.WriteFile(c.filePath, data, 0644); err != nil {
		return fmt.Errorf("writing cache file: %w", err)
	}

	return nil
}

// GetBulletin retrieves a cached bulletin summary
func (c *Cache) GetBulletin(id string) (adobe.BulletinSummary, bool) {
	bulletin, exists := c.data.Bulletins[id]
	return bulletin, exists
}

// SetBulletin stores a bulletin summary in cache
func (c *Cache) SetBulletin(summary adobe.BulletinSummary) {
	c.data.Bulletins[summary.ID] = summary
}

// IsNewOrUpdated checks if a bulletin is new or has been updated
func (c *Cache) IsNewOrUpdated(summary adobe.BulletinSummary) bool {
	cached, exists := c.data.Bulletins[summary.ID]
	if !exists {
		return true // New bulletin
	}

	// Check if update date has changed
	return summary.UpdatedAt.After(cached.UpdatedAt)
}

// GetLastRun returns the last successful scraper run time
func (c *Cache) GetLastRun() time.Time {
	return c.data.LastRun
}

// SetLastRun updates the last successful scraper run time
func (c *Cache) SetLastRun(t time.Time) {
	c.data.LastRun = t
}

// GetAllBulletins returns all cached bulletin summaries
func (c *Cache) GetAllBulletins() []adobe.BulletinSummary {
	var bulletins []adobe.BulletinSummary
	for _, summary := range c.data.Bulletins {
		bulletins = append(bulletins, summary)
	}
	return bulletins
}

// GetBulletinsByProduct returns cached bulletins for a specific product
func (c *Cache) GetBulletinsByProduct(product string) []adobe.BulletinSummary {
	var bulletins []adobe.BulletinSummary
	for _, summary := range c.data.Bulletins {
		if summary.Product == product {
			bulletins = append(bulletins, summary)
		}
	}
	return bulletins
}

// RemoveOldBulletins removes bulletins older than the specified duration
func (c *Cache) RemoveOldBulletins(maxAge time.Duration) int {
	cutoff := time.Now().Add(-maxAge)
	removed := 0

	for id, summary := range c.data.Bulletins {
		if summary.PublishedAt.Before(cutoff) {
			delete(c.data.Bulletins, id)
			removed++
		}
	}

	return removed
}

// GetStats returns cache statistics
func (c *Cache) GetStats() CacheStats {
	now := time.Now()
	stats := CacheStats{
		TotalBulletins: len(c.data.Bulletins),
		LastUpdated:    c.data.UpdatedAt,
		LastRun:        c.data.LastRun,
		ByProduct:      make(map[string]int),
	}

	// Count by product
	for _, summary := range c.data.Bulletins {
		stats.ByProduct[summary.Product]++
	}

	// Count recent bulletins (last 30 days)
	thirtyDaysAgo := now.AddDate(0, 0, -30)
	for _, summary := range c.data.Bulletins {
		if summary.PublishedAt.After(thirtyDaysAgo) {
			stats.RecentBulletins++
		}
	}

	return stats
}

// CacheStats represents cache statistics
type CacheStats struct {
	TotalBulletins  int            `json:"total_bulletins"`
	RecentBulletins int            `json:"recent_bulletins"` // Last 30 days
	LastUpdated     time.Time      `json:"last_updated"`
	LastRun         time.Time      `json:"last_run"`
	ByProduct       map[string]int `json:"by_product"`
}
