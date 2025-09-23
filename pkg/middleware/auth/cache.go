package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"
)

// CacheEntry represents a cached validation result
type CacheEntry struct {
	Result    *ValidationResult
	ExpiresAt time.Time
}

// IsExpired checks if the cache entry has expired
func (e *CacheEntry) IsExpired() bool {
	return time.Now().After(e.ExpiresAt)
}

// Cache provides in-memory caching for token validation results
type Cache struct {
	entries  map[string]*CacheEntry
	mutex    sync.RWMutex
	ttl      time.Duration
	maxSize  int
	enabled  bool
	cleanupTicker *time.Ticker
	stopCleanup   chan struct{}
}

// NewCache creates a new cache instance
func NewCache(enabled bool, ttl time.Duration, maxSize int) *Cache {
	if !enabled {
		return &Cache{enabled: false}
	}

	cache := &Cache{
		entries:     make(map[string]*CacheEntry),
		ttl:         ttl,
		maxSize:     maxSize,
		enabled:     true,
		stopCleanup: make(chan struct{}),
	}

	// Start cleanup goroutine
	cache.cleanupTicker = time.NewTicker(ttl / 2) // Clean up twice per TTL period
	go cache.cleanup()

	return cache
}

// Get retrieves a cached validation result
func (c *Cache) Get(token string) (*ValidationResult, bool) {
	if !c.enabled {
		return nil, false
	}

	c.mutex.RLock()
	defer c.mutex.RUnlock()

	key := c.tokenKey(token)
	entry, exists := c.entries[key]
	if !exists {
		return nil, false
	}

	if entry.IsExpired() {
		// Entry is expired, will be cleaned up later
		return nil, false
	}

	return entry.Result, true
}

// Set stores a validation result in the cache
func (c *Cache) Set(token string, result *ValidationResult) {
	if !c.enabled || result == nil {
		return
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	// If cache is at max size, remove oldest entries
	if len(c.entries) >= c.maxSize {
		c.evictOldest()
	}

	key := c.tokenKey(token)
	c.entries[key] = &CacheEntry{
		Result:    result,
		ExpiresAt: time.Now().Add(c.ttl),
	}
}

// Clear removes all entries from the cache
func (c *Cache) Clear() {
	if !c.enabled {
		return
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.entries = make(map[string]*CacheEntry)
}

// Size returns the number of entries in the cache
func (c *Cache) Size() int {
	if !c.enabled {
		return 0
	}

	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return len(c.entries)
}

// Stop stops the cache cleanup goroutine
func (c *Cache) Stop() {
	if !c.enabled || c.cleanupTicker == nil {
		return
	}

	c.cleanupTicker.Stop()
	close(c.stopCleanup)
}

// tokenKey generates a cache key from the token
// Uses SHA256 hash to avoid storing actual tokens in memory
func (c *Cache) tokenKey(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// evictOldest removes the oldest 10% of entries
func (c *Cache) evictOldest() {
	if len(c.entries) == 0 {
		return
	}

	// Calculate how many entries to evict (10% or at least 1)
	evictCount := len(c.entries) / 10
	if evictCount == 0 {
		evictCount = 1
	}

	// Find oldest entries by expiration time
	type keyTime struct {
		key       string
		expiresAt time.Time
	}

	var keyTimes []keyTime
	for key, entry := range c.entries {
		keyTimes = append(keyTimes, keyTime{
			key:       key,
			expiresAt: entry.ExpiresAt,
		})
	}

	// Sort by expiration time (oldest first)
	for i := 0; i < len(keyTimes)-1; i++ {
		for j := i + 1; j < len(keyTimes); j++ {
			if keyTimes[i].expiresAt.After(keyTimes[j].expiresAt) {
				keyTimes[i], keyTimes[j] = keyTimes[j], keyTimes[i]
			}
		}
	}

	// Remove oldest entries
	for i := 0; i < evictCount && i < len(keyTimes); i++ {
		delete(c.entries, keyTimes[i].key)
	}
}

// cleanup runs periodically to remove expired entries
func (c *Cache) cleanup() {
	for {
		select {
		case <-c.cleanupTicker.C:
			c.removeExpired()
		case <-c.stopCleanup:
			return
		}
	}
}

// removeExpired removes all expired entries from the cache
func (c *Cache) removeExpired() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()
	for key, entry := range c.entries {
		if now.After(entry.ExpiresAt) {
			delete(c.entries, key)
		}
	}
}