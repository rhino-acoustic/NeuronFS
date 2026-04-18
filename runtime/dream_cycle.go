package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ============================================================================
// Module: Dream Cycle — Autonomous Memory Consolidation (Phase 41)
// Periodically runs background "sleep" maintenance:
//   1. Prune stale sensory alerts from agent inboxes
//   2. Summarize audit trail into daily digests
//   3. Clean temporal_log deltas older than 7 days
// ============================================================================

func startDreamCycleDaemon(brainRoot string) {
	if brainRoot == "" {
		return
	}

	svLog("  💤 DREAM ENGINE: Memory Consolidation Daemon ONLINE (30min cycle)")

	for {
		time.Sleep(30 * time.Minute)
		runDreamCycle(brainRoot)
	}
}

func runDreamCycle(brainRoot string) {
	svLog("[Dream] 💤 Starting memory consolidation cycle...")

	pruned := pruneSensoryAlerts(brainRoot)
	cleaned := cleanOldDeltas(brainRoot, 7)
	compacted := compactAuditTrail(brainRoot)

	summary := fmt.Sprintf("pruned %d alerts, cleaned %d deltas, compacted %d audit lines", pruned, cleaned, compacted)
	svLog(fmt.Sprintf("[Dream] ✅ Cycle complete: %s", summary))

	RecordAudit(brainRoot, "dream_cycle", "consolidate", "hippocampus", summary, true)

	if GlobalSSEBroker != nil {
		GlobalSSEBroker.Broadcast("info", fmt.Sprintf("[Dream] 💤 %s", summary))
	}
}

// pruneSensoryAlerts removes sensory_alert_*.md files older than 1 hour
func pruneSensoryAlerts(brainRoot string) int {
	pruned := 0
	agentsDir := filepath.Join(brainRoot, "_agents")
	cutoff := time.Now().Add(-1 * time.Hour)

	_ = filepath.Walk(agentsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if strings.HasPrefix(info.Name(), "sensory_alert_") && info.ModTime().Before(cutoff) {
			archiveDir := filepath.Join(filepath.Dir(path), "_archive")
			_ = os.MkdirAll(archiveDir, 0755)
			archivePath := filepath.Join(archiveDir, info.Name())
			if os.Rename(path, archivePath) == nil {
				pruned++
			}
		}
		return nil
	})
	return pruned
}

// cleanOldDeltas archives temporal_log deltas older than maxDays
func cleanOldDeltas(brainRoot string, maxDays int) int {
	cleaned := 0
	temporalDir := filepath.Join(brainRoot, "hippocampus", "temporal_log")
	archiveDir := filepath.Join(temporalDir, "_archive")
	cutoff := time.Now().AddDate(0, 0, -maxDays)

	_ = filepath.Walk(temporalDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if strings.HasSuffix(info.Name(), ".delta") && info.ModTime().Before(cutoff) {
			_ = os.MkdirAll(archiveDir, 0755)
			archivePath := filepath.Join(archiveDir, info.Name())
			if os.Rename(path, archivePath) == nil {
				cleaned++
			}
		}
		return nil
	})
	return cleaned
}

// compactAuditTrail rotates audit.jsonl when it exceeds 1000 lines
func compactAuditTrail(brainRoot string) int {
	auditFile := filepath.Join(brainRoot, "hippocampus", "audit_trail", "audit.jsonl")
	data, err := os.ReadFile(auditFile)
	if err != nil {
		return 0
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) <= 1000 {
		return 0
	}

	// Keep the last 500 lines, archive the rest
	archiveDir := filepath.Join(brainRoot, "hippocampus", "audit_trail", "_archive")
	_ = os.MkdirAll(archiveDir, 0755)

	archiveName := fmt.Sprintf("audit_%s.jsonl", time.Now().Format("20060102_150405"))
	archivePath := filepath.Join(archiveDir, archiveName)

	// Write old lines to archive
	oldLines := lines[:len(lines)-500]
	_ = os.WriteFile(archivePath, []byte(strings.Join(oldLines, "\n")), 0644)

	// Keep recent lines
	recentLines := lines[len(lines)-500:]
	_ = os.WriteFile(auditFile, []byte(strings.Join(recentLines, "\n")), 0644)

	return len(oldLines)
}
