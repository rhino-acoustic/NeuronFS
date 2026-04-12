package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ============================================================================
// Module: Agent-to-Agent P2P Knowledge Synchronizer (Phase 36)
// Automatically merges local corrections.jsonl to the global swarm brain.
// ============================================================================

func startP2PSyncDaemon(brainRoot string) {
	if brainRoot == "" {
		return
	}
	
	fmt.Println("  🔄 SWARM ENGINE: P2P Knowledge Cross-over Daemon ONLINE")
	
	for {
		time.Sleep(5 * time.Minute)
		syncP2PCorrections(brainRoot)
	}
}

func syncP2PCorrections(brainRoot string) {
	agentsDir := filepath.Join(brainRoot, "_agents")
	globalInboxFile := filepath.Join(brainRoot, "_inbox", "corrections.jsonl")
	
	// Create global inbox directory if not exists
	_ = os.MkdirAll(filepath.Dir(globalInboxFile), 0755)
	
	// 1. Read existing global corrections to map hashes and avoid dups
	globalHashes := make(map[string]bool)
	content, err := os.ReadFile(globalInboxFile)
	if err == nil {
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				hash := hashLine(line)
				globalHashes[hash] = true
			}
		}
	}
	
	// 2. Scan all subdirectories in _agents for their own corrections.jsonl
	newEntries := 0
	var mergedLines []string
	
	_ = filepath.WalkDir(agentsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		
		if d.Name() == "corrections.jsonl" && path != globalInboxFile {
			localData, readErr := os.ReadFile(path)
			if readErr == nil {
				lines := strings.Split(string(localData), "\n")
				for _, line := range lines {
					cleanLine := strings.TrimSpace(line)
					if cleanLine != "" {
						h := hashLine(cleanLine)
						if !globalHashes[h] {
							globalHashes[h] = true
							mergedLines = append(mergedLines, cleanLine)
							newEntries++
						}
					}
				}
			}
		}
		return nil
	})
	
	// 3. Append to global
	if newEntries > 0 {
		f, err := os.OpenFile(globalInboxFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err == nil {
			defer f.Close()
			for _, ml := range mergedLines {
				_, _ = f.WriteString(ml + "\n")
			}
			fmt.Printf("[P2P Sync] Successfully merged %d new cross-over corrections to swarm.\n", newEntries)
			
			// Broadcast via global SSE
			if GlobalSSEBroker != nil {
				GlobalSSEBroker.Broadcast("info", fmt.Sprintf("[P2P Sync] %d개의 피어 지식이 전역 스웜으로 융합되었습니다.", newEntries))
			}
			// Phase 40: Audit Trail
			RecordAudit(brainRoot, "p2p_sync", "merge", globalInboxFile, fmt.Sprintf("%d new corrections merged", newEntries), true)
		}
	}
}

func hashLine(line string) string {
	h := sha256.Sum256([]byte(line))
	return hex.EncodeToString(h[:])
}
