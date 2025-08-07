package cache

import (
	"sync"
	"time"
)

// CacheEntry represents a cached item with expiration time
type CacheEntry struct {
	Value     interface{}
	ExpiresAt time.Time
}

// Cache interface defines the caching operations
type Cache interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, ttl time.Duration)
	Delete(key string)
	Clear()
}

// MemoryCache implements an in-memory cache with TTL support
type MemoryCache struct {
	items map[string]*CacheEntry
	mutex sync.RWMutex
}

// NewMemoryCache creates a new in-memory cache
func NewMemoryCache() *MemoryCache {
	cache := &MemoryCache{
		items: make(map[string]*CacheEntry),
	}
	
	// Start cleanup goroutine to remove expired entries
	go cache.cleanup()
	
	return cache
}

// Get retrieves a value from the cache
func (c *MemoryCache) Get(key string) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	entry, exists := c.items[key]
	if !exists {
		return nil, false
	}
	
	// Check if entry has expired
	if time.Now().After(entry.ExpiresAt) {
		// Don't delete here to avoid write lock in read operation
		return nil, false
	}
	
	return entry.Value, true
}

// Set stores a value in the cache with TTL
func (c *MemoryCache) Set(key string, value interface{}, ttl time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	c.items[key] = &CacheEntry{
		Value:     value,
		ExpiresAt: time.Now().Add(ttl),
	}
}

// Delete removes a key from the cache
func (c *MemoryCache) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	delete(c.items, key)
}

// Clear removes all entries from the cache
func (c *MemoryCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	c.items = make(map[string]*CacheEntry)
}

// cleanup periodically removes expired entries
func (c *MemoryCache) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			c.removeExpired()
		}
	}
}

// removeExpired removes all expired entries from the cache
func (c *MemoryCache) removeExpired() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	now := time.Now()
	for key, entry := range c.items {
		if now.After(entry.ExpiresAt) {
			delete(c.items, key)
		}
	}
}