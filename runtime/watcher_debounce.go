package main

import (
	"fmt"
	"sync"
	"time"
)

// ============================================================================
// Module: FS Watcher Debounce — Phase 50
// Batches filesystem change notifications per folder (60s window).
// Suppresses repeated alerts for the same folder after 3 occurrences.
// ============================================================================

// DebouncedEvent represents a batched set of changes in one folder
type DebouncedEvent struct {
	Folder     string
	Count      int
	FirstSeen  time.Time
	LastSeen   time.Time
	Suppressed bool
}

// WatcherDebouncer manages debouncing of FS events
type WatcherDebouncer struct {
	mu          sync.Mutex
	events      map[string]*DebouncedEvent
	suppressMap map[string]int // folder → total alert count (across windows)
	window      time.Duration
	maxRepeats  int
	onFlush     func([]DebouncedEvent)
}

// NewWatcherDebouncer creates a debouncer with configurable window and repeat limit
func NewWatcherDebouncer(window time.Duration, maxRepeats int, onFlush func([]DebouncedEvent)) *WatcherDebouncer {
	wd := &WatcherDebouncer{
		events:      make(map[string]*DebouncedEvent),
		suppressMap: make(map[string]int),
		window:      window,
		maxRepeats:  maxRepeats,
		onFlush:     onFlush,
	}
	go wd.flushLoop()
	return wd
}

// Record adds a filesystem event to the debounce buffer
func (wd *WatcherDebouncer) Record(folder string) {
	wd.mu.Lock()
	defer wd.mu.Unlock()

	if ev, ok := wd.events[folder]; ok {
		ev.Count++
		ev.LastSeen = time.Now()
	} else {
		wd.events[folder] = &DebouncedEvent{
			Folder:    folder,
			Count:     1,
			FirstSeen: time.Now(),
			LastSeen:  time.Now(),
		}
	}
}

// flushLoop periodically flushes accumulated events
func (wd *WatcherDebouncer) flushLoop() {
	ticker := time.NewTicker(wd.window)
	defer ticker.Stop()

	for range ticker.C {
		wd.flush()
	}
}

// flush processes and clears the event buffer
func (wd *WatcherDebouncer) flush() {
	wd.mu.Lock()
	defer wd.mu.Unlock()

	if len(wd.events) == 0 {
		return
	}

	var batch []DebouncedEvent
	for folder, ev := range wd.events {
		wd.suppressMap[folder]++

		if wd.suppressMap[folder] > wd.maxRepeats {
			ev.Suppressed = true
			fmt.Printf("[디바운스] 억제됨: %s (%d회 초과)\n", folder, wd.maxRepeats)
		} else {
			batch = append(batch, *ev)
		}
	}

	// Clear current window
	wd.events = make(map[string]*DebouncedEvent)

	if len(batch) > 0 && wd.onFlush != nil {
		wd.onFlush(batch)
	}
}

// ResetSuppress clears suppression for a folder (e.g., after manual review)
func (wd *WatcherDebouncer) ResetSuppress(folder string) {
	wd.mu.Lock()
	defer wd.mu.Unlock()
	delete(wd.suppressMap, folder)
}

// Stats returns current debouncer state
func (wd *WatcherDebouncer) Stats() (pending int, suppressed int) {
	wd.mu.Lock()
	defer wd.mu.Unlock()
	pending = len(wd.events)
	for _, count := range wd.suppressMap {
		if count > wd.maxRepeats {
			suppressed++
		}
	}
	return
}
