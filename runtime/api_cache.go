package main

import (
	"sync"
	"time"
)

// APICacheItem represents a cached API response
type APICacheItem struct {
	Data        []byte
	ContentType string
	ExpiresAt   time.Time
}

// APICacheManager manages API response caches with TTL and invalidation
type APICacheManager struct {
	mu    sync.RWMutex
	items map[string]APICacheItem
}

// GlobalAPICache is the global instance of the cache manager
var GlobalAPICache = &APICacheManager{
	items: make(map[string]APICacheItem),
}

// Get retrieves a cached item if it exists and is not expired
func (m *APICacheManager) Get(key string) ([]byte, string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	item, exists := m.items[key]
	if !exists {
		return nil, "", false
	}

	if time.Now().After(item.ExpiresAt) {
		return nil, "", false
	}

	return item.Data, item.ContentType, true
}

// Set stores an item in the cache
func (m *APICacheManager) Set(key string, data []byte, contentType string, ttl time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.items[key] = APICacheItem{
		Data:        data,
		ContentType: contentType,
		ExpiresAt:   time.Now().Add(ttl),
	}
}

// Invalidate removes a specific key from the cache
func (m *APICacheManager) Invalidate(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.items, key)
}

// InvalidateAll clears all cached items
func (m *APICacheManager) InvalidateAll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.items = make(map[string]APICacheItem)
}

// InvalidatePrefix removes all keys that start with the given prefix
func (m *APICacheManager) InvalidatePrefix(prefix string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for key := range m.items {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			delete(m.items, key)
		}
	}
}
