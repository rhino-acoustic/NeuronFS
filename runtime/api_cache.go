package main

import (
	"strings"
	"sync"
	"time"
)

// APICacheItem represents a cached API response with optional region metadata
type APICacheItem struct {
	Data        []byte
	ContentType string
	ExpiresAt   time.Time
	Region      string // "brainstem", "cortex", "prefrontal", etc.
}

// APICacheManager manages API response caches with TTL and hierarchical invalidation
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
	m.SetWithRegion(key, "", data, contentType, ttl)
}

// SetWithRegion stores an item in the cache with region metadata for hierarchical invalidation
func (m *APICacheManager) SetWithRegion(key string, region string, data []byte, contentType string, ttl time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.items[key] = APICacheItem{
		Data:        data,
		ContentType: contentType,
		ExpiresAt:   time.Now().Add(ttl),
		Region:      region,
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
		if strings.HasPrefix(key, prefix) {
			delete(m.items, key)
		}
	}
}

// InvalidateHierarchical removes items based on NeuronFS hierarchy.
// If region X (Priority Pn) is invalidated, all items in region X and its lower-priority
// regions (Pn+1...P7) are also removed.
func (m *APICacheManager) InvalidateHierarchical(region string) {
	p, ok := RegionPriority[region]
	if !ok {
		// Unknown region? Clear all to be safe, or just clear the prefix.
		m.InvalidateAll()
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for key, item := range m.items {
		if item.Region == "" {
			// System-wide or unassigned items are cleared if P0 (brainstem) or P1 (limbic) changes.
			if p <= 1 {
				delete(m.items, key)
			}
			continue
		}

		itemP, exists := RegionPriority[item.Region]
		if !exists || itemP >= p {
			// Invalidate if item's priority is lower than or equal to the changed region.
			// (Lower priority = higher P number, e.g., P4 cortex vs P6 prefrontal)
			delete(m.items, key)
		}
	}
}
