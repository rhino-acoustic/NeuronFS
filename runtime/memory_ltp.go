package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// RunLTP (Long-Term Potentiation) defragments the nervous system by merging mature (aged)
// short-term .neuron files into the long-term semantic _rules.md memory.
// Once merged, local .neuron files are quarantined into _archive/ to prevent clutter.
func RunLTP(brainRoot string, logger func(string)) {
	if logger != nil {
		logger("🧠 [LTP] Memory Defragmentation Cycle Activated")
	}

	mergedCount := 0
	
	// Scan the brain hierarchical folders for any .neuron files older than 24h
	filepath.Walk(brainRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		
		// Skip archive, strict paths, or already processed
		if strings.Contains(path, "_archive") || strings.Contains(path, "_quarantine") || strings.Contains(path, "scratch") {
			return nil
		}

		// Only target .neuron short-term memory files
		if !strings.HasSuffix(info.Name(), ".neuron") {
			return nil
		}
		
		// Age threshold check: Only consolidate memories older than 12 hours
		if time.Since(info.ModTime()) > 12*time.Hour {
			dir := filepath.Dir(path)
			
			// If a _rules.md exists in the same directory or we can append
			// We only append to the first closest _rules.md upwards
			targetRules := locateRulesFile(dir, brainRoot)
			if targetRules != "" {
				content, readErr := os.ReadFile(path)
				if readErr == nil {
					// Append content to _rules.md
					f, err := os.OpenFile(targetRules, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
					if err == nil {
						fmt.Fprintf(f, "\n\n<!-- Consolidated via LTP from %s at %s -->\n", info.Name(), time.Now().Format(time.RFC3339))
						f.Write(content)
						f.Close()
						
						// Archive original neuron
						archiveDir := filepath.Join(dir, "_archive")
						os.MkdirAll(archiveDir, 0750)
						destPath := filepath.Join(archiveDir, info.Name())
						os.Rename(path, destPath)
						
						mergedCount++
					}
				}
			}
		}
		return nil
	})
	
	if logger != nil && mergedCount > 0 {
		logger(fmt.Sprintf("🧠 [LTP] Successfully consolidated %d neurons into Long-Term Memory.", mergedCount))
	} else if logger != nil {
		logger("🧠 [LTP] No mature neurons found for consolidation.")
	}
}

// locateRulesFile walks up the directory tree to find the nearest _rules.md, stopping at brainRoot.
func locateRulesFile(startDir, bound string) string {
	curr := startDir
	for {
		candidate := filepath.Join(curr, "_rules.md")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		// Try to create it if we are in a main region like cortex or brainstem but no _rules.md exists
		
		if curr == bound || curr == filepath.Dir(bound) || curr == "." || curr == "\\" || curr == "/" {
			break
		}
		curr = filepath.Dir(curr)
	}
	return ""
}
