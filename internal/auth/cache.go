package auth

import (
	"sync"
	"time"
)

// CacheEntry represents a cached value with expiration
type CacheEntry[T any] struct {
	Value     T
	ExpiresAt time.Time
}

// Cache provides a generic thread-safe in-memory cache with TTL support
type Cache[T any] struct {
	mu    sync.RWMutex
	items map[string]*CacheEntry[T]
}

// NewCache creates a new generic cache
func NewCache[T any]() *Cache[T] {
	return &Cache[T]{
		items: make(map[string]*CacheEntry[T]),
	}
}

// Get retrieves a value from the cache
// Returns (value, found) where found indicates if the key exists and is not expired
func (c *Cache[T]) Get(key string) (T, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.items[key]
	if !exists {
		var zero T
		return zero, false
	}

	if time.Now().After(entry.ExpiresAt) {
		var zero T
		return zero, false
	}

	return entry.Value, true
}

// Set stores a value in the cache with a specific expiration time
func (c *Cache[T]) Set(key string, value T, expiresAt time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = &CacheEntry[T]{
		Value:     value,
		ExpiresAt: expiresAt,
	}
}

// SetWithTTL stores a value in the cache with a TTL duration
func (c *Cache[T]) SetWithTTL(key string, value T, ttl time.Duration) {
	c.Set(key, value, time.Now().Add(ttl))
}

// Delete removes a key from the cache
func (c *Cache[T]) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
}

// GetAndDelete atomically retrieves and removes a value from the cache
// Returns (value, found) where found indicates if the key existed and was not expired
func (c *Cache[T]) GetAndDelete(key string) (T, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, exists := c.items[key]
	if !exists {
		var zero T
		return zero, false
	}

	if time.Now().After(entry.ExpiresAt) {
		delete(c.items, key)
		var zero T
		return zero, false
	}

	delete(c.items, key)
	return entry.Value, true
}

// InvalidatePrefix removes all keys that start with the given prefix
func (c *Cache[T]) InvalidatePrefix(prefix string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key := range c.items {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			delete(c.items, key)
		}
	}
}

// Clear removes all entries from the cache
func (c *Cache[T]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*CacheEntry[T])
}
