package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// ============================================================================
// Module: Procedural Memory Extractor (Phase 66 — Community Trend Integration)
// Scans audit.jsonl for recurring action patterns and materializes them
// as skill neurons in hippocampus/skills/auto_extracted/.
// This closes the loop: Agent acts → Audit logs → Patterns extracted →
// Skills formed → Agent uses skills to improve future actions.
// ============================================================================

// PatternFrequency tracks how often an actor+action pair appears
type PatternFrequency struct {
	Actor  string `json:"actor"`
	Action string `json:"action"`
	Count  int    `json:"count"`
	LastTs string `json:"last_ts"`
}

// ExtractProceduralMemory reads audit.jsonl, counts actor+action frequencies,
// and writes skill neurons for patterns that appear >= threshold times.
func ExtractProceduralMemory(brainRoot string, threshold int) int {
	if threshold < 2 {
		threshold = 3 // Default: pattern must appear at least 3 times
	}

	auditPath := filepath.Join(brainRoot, "hippocampus", "audit_trail", "audit.jsonl")
	file, err := os.Open(auditPath)
	if err != nil {
		return 0
	}
	defer file.Close()

	// Count actor+action pairs
	freq := make(map[string]*PatternFrequency)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var entry AuditEntry
		if json.Unmarshal([]byte(line), &entry) != nil {
			continue
		}

		key := entry.Actor + "|" + entry.Action
		if _, exists := freq[key]; !exists {
			freq[key] = &PatternFrequency{
				Actor:  entry.Actor,
				Action: entry.Action,
			}
		}
		freq[key].Count++
		freq[key].LastTs = entry.Timestamp
	}

	// Sort by frequency, descending
	var patterns []*PatternFrequency
	for _, pf := range freq {
		if pf.Count >= threshold {
			patterns = append(patterns, pf)
		}
	}
	sort.Slice(patterns, func(i, j int) bool {
		return patterns[i].Count > patterns[j].Count
	})

	if len(patterns) == 0 {
		return 0
	}

	// Write skill neurons
	skillDir := filepath.Join(brainRoot, "hippocampus", "skills", "auto_extracted")
	os.MkdirAll(skillDir, 0755)

	created := 0
	for _, p := range patterns {
		safeName := strings.ReplaceAll(p.Actor+"_"+p.Action, " ", "_")
		safeName = strings.ReplaceAll(safeName, "/", "_")
		neuronPath := filepath.Join(skillDir, safeName+".neuron")

		// Skip if already extracted (idempotent)
		if _, err := os.Stat(neuronPath); err == nil {
			continue
		}

		content := fmt.Sprintf(
			"# Auto-Extracted Procedural Memory\n\n"+
				"actor: %s\n"+
				"action: %s\n"+
				"frequency: %d\n"+
				"last_seen: %s\n"+
				"extracted_at: %s\n\n"+
				"## Pattern\n"+
				"The actor `%s` repeatedly performs `%s` (%d times observed).\n"+
				"This workflow has been auto-promoted to a skill neuron.\n",
			p.Actor, p.Action, p.Count, p.LastTs,
			time.Now().Format(time.RFC3339),
			p.Actor, p.Action, p.Count,
		)

		os.WriteFile(neuronPath, []byte(content), 0644)
		created++
		fmt.Printf("[ProceduralMemory] Skill extracted: %s → %s (%dx)\n", p.Actor, p.Action, p.Count)
	}

	return created
}
