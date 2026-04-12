package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ============================================================================
// Module: Context Cache — Phase 56
// Pre-builds and caches brain context so agents don't reload every invocation.
// Reduces MCP overhead from 15s to 0s for repeat queries.
// ============================================================================

// ContextCache holds pre-built brain context
type ContextCache struct {
	mu        sync.RWMutex
	summary   string
	builtAt   time.Time
	ttl       time.Duration
	brainRoot string
}

// NewContextCache creates a cache with configurable TTL
func NewContextCache(brainRoot string, ttl time.Duration) *ContextCache {
	return &ContextCache{
		brainRoot: brainRoot,
		ttl:       ttl,
	}
}

// Get returns cached context, rebuilding if stale
func (cc *ContextCache) Get() string {
	cc.mu.RLock()
	if cc.summary != "" && time.Since(cc.builtAt) < cc.ttl {
		defer cc.mu.RUnlock()
		return cc.summary
	}
	cc.mu.RUnlock()

	return cc.rebuild()
}

// rebuild scans brain and rebuilds context summary
func (cc *ContextCache) rebuild() string {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	// Double-check after acquiring write lock
	if cc.summary != "" && time.Since(cc.builtAt) < cc.ttl {
		return cc.summary
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("[NeuronFS Brain Context — Built: %s]\n\n", time.Now().Format("15:04:05")))

	// Count neurons per region
	regions := []string{"brainstem", "limbic", "hippocampus", "sensors", "cortex", "ego", "prefrontal"}
	totalNeurons := 0

	for _, region := range regions {
		regionPath := filepath.Join(cc.brainRoot, region)
		count := countFiles(regionPath, ".neuron")
		totalNeurons += count
		sb.WriteString(fmt.Sprintf("  %s: %d neurons\n", region, count))
	}
	sb.WriteString(fmt.Sprintf("\nTotal: %d neurons\n", totalNeurons))

	// Read active rules
	rulesPath := filepath.Join(cc.brainRoot, "cortex", "dev", "_rules.md")
	if data, err := os.ReadFile(rulesPath); err == nil {
		rules := string(data)
		if len(rules) > 500 {
			rules = rules[:500] + "...(truncated)"
		}
		sb.WriteString(fmt.Sprintf("\n[Active Rules]\n%s\n", rules))
	}

	// Read recent corrections
	correctionsPath := filepath.Join(cc.brainRoot, "_inbox", "corrections.jsonl")
	if data, err := os.ReadFile(correctionsPath); err == nil {
		lines := strings.Split(strings.TrimSpace(string(data)), "\n")
		sb.WriteString(fmt.Sprintf("\n[Corrections: %d entries]\n", len(lines)))
		// Show last 3
		start := len(lines) - 3
		if start < 0 {
			start = 0
		}
		for _, line := range lines[start:] {
			if len(line) > 200 {
				line = line[:200] + "..."
			}
			sb.WriteString("  " + line + "\n")
		}
	}

	// Read EVOLUTION_TODO status
	todoPath := filepath.Join(cc.brainRoot, "cortex", "dev", "작업", "EVOLUTION_TODO.neuron")
	if data, err := os.ReadFile(todoPath); err == nil {
		content := string(data)
		completed := strings.Count(content, "[x]")
		pending := strings.Count(content, "[ ]")
		sb.WriteString(fmt.Sprintf("\n[Evolution: %d completed, %d pending]\n", completed, pending))
	}

	cc.summary = sb.String()
	cc.builtAt = time.Now()

	fmt.Printf("[컨텍스트캐시] 재빌드 완료: %d neurons, %s\n", totalNeurons, time.Now().Format("15:04:05"))

	return cc.summary
}

// Invalidate forces a rebuild on next Get()
func (cc *ContextCache) Invalidate() {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.summary = ""
}

// countFiles counts files with given extension recursively
func countFiles(dir, ext string) int {
	count := 0
	_ = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ext) {
			count++
		}
		return nil
	})
	return count
}
