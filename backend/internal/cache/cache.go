package cache

import (
	"strings"
	"sync"
	"time"
)

// entry holds a cached value with expiration.
type entry struct {
	value     interface{}
	expiresAt time.Time
}

// Cache is a thread-safe in-memory TTL cache.
type Cache struct {
	mu      sync.RWMutex
	entries map[string]*entry
	ttl     time.Duration
	stop    chan struct{}
}

// New creates a cache with the given default TTL.
// A background goroutine purges expired entries every 2 minutes.
func New(ttl time.Duration) *Cache {
	c := &Cache{
		entries: make(map[string]*entry),
		ttl:     ttl,
		stop:    make(chan struct{}),
	}
	go c.cleanup()
	return c
}

// Get retrieves a cached value. Returns (value, true) on hit, (nil, false) on miss/expired.
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	e, ok := c.entries[key]
	if !ok || time.Now().After(e.expiresAt) {
		c.mu.RUnlock()
		return nil, false
	}
	val := e.value
	c.mu.RUnlock()
	return val, true
}

// Set stores a value with the default TTL.
func (c *Cache) Set(key string, value interface{}) {
	c.SetWithTTL(key, value, c.ttl)
}

// SetWithTTL stores a value with a custom TTL (clamped to at least 1s when positive).
// A zero or negative ttl uses the cache default.
func (c *Cache) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	if ttl <= 0 {
		ttl = c.ttl
	}
	c.mu.Lock()
	c.entries[key] = &entry{value: value, expiresAt: time.Now().Add(ttl)}
	c.mu.Unlock()
}

// Delete removes a single key.
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	delete(c.entries, key)
	c.mu.Unlock()
}

// DeletePrefix removes all keys matching the given prefix.
func (c *Cache) DeletePrefix(prefix string) {
	c.mu.Lock()
	for k := range c.entries {
		if strings.HasPrefix(k, prefix) {
			delete(c.entries, k)
		}
	}
	c.mu.Unlock()
}

// DeleteWhere removes keys for which pred returns true.
func (c *Cache) DeleteWhere(pred func(key string) bool) {
	if pred == nil {
		return
	}
	c.mu.Lock()
	for k := range c.entries {
		if pred(k) {
			delete(c.entries, k)
		}
	}
	c.mu.Unlock()
}

// Flush removes all entries.
func (c *Cache) Flush() {
	c.mu.Lock()
	c.entries = make(map[string]*entry)
	c.mu.Unlock()
}

// Len returns the number of entries (including not-yet-purged expired ones).
func (c *Cache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}

// cleanup runs every 2 minutes to remove expired entries.
func (c *Cache) cleanup() {
	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			now := time.Now()
			c.mu.Lock()
			for k, e := range c.entries {
				if now.After(e.expiresAt) {
					delete(c.entries, k)
				}
			}
			c.mu.Unlock()
		case <-c.stop:
			return
		}
	}
}

// Stop halts the background cleanup goroutine.
func (c *Cache) Stop() {
	close(c.stop)
}
