package cache

import (
	"sync"
	"time"
)

// CacheItem represents a cached item with expiration
type CacheItem struct {
	Value      interface{}
	Expiration time.Time
}

// Cache is an in-memory cache with TTL support
type Cache struct {
	mu      sync.RWMutex
	items   map[string]CacheItem
	ttl     time.Duration
	hits    int64
	misses  int64
}

// NewCache creates a new cache with the specified TTL
func NewCache(ttl time.Duration) *Cache {
	c := &Cache{
		items: make(map[string]CacheItem),
		ttl:   ttl,
	}
	// Start cleanup routine
	go c.cleanup()
	return c
}

// cleanup removes expired items periodically
func (c *Cache) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, item := range c.items {
			if now.After(item.Expiration) {
				delete(c.items, key)
			}
		}
		c.mu.Unlock()
	}
}

// Get retrieves a value from cache
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	item, found := c.items[key]
	if !found {
		c.misses++
		return nil, false
	}
	
	if time.Now().After(item.Expiration) {
		delete(c.items, key)
		c.misses++
		return nil, false
	}
	
	c.hits++
	return item.Value, true
}

// Set stores a value in cache with TTL
func (c *Cache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.items[key] = CacheItem{
		Value:      value,
		Expiration: time.Now().Add(c.ttl),
	}
}

// Delete removes a key from cache
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// Clear removes all items from cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]CacheItem)
	c.hits = 0
	c.misses = 0
}

// Stats returns cache statistics
func (c *Cache) Stats() (hits, misses int64, size int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.hits, c.misses, len(c.items)
}

// HitRate returns the cache hit rate as a percentage
func (c *Cache) HitRate() float64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	total := c.hits + c.misses
	if total == 0 {
		return 0
	}
	return float64(c.hits) / float64(total) * 100
}

// Default cache instance with 5-minute TTL
var DefaultCache = NewCache(5 * time.Minute)

// ArticleCache is the main cache for articles
var ArticleCache = NewCache(5 * time.Minute)
