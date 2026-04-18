package main

// ━━━ watch.go ━━━
// PROVIDES: runWatch
// DEPENDS ON: brain.go, emit_tiers.go, inject.go

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

func runWatch(brainRoot string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Printf("%s[TRAUMA] Failed to initialize neural watcher: %v%s\n", ansiRed, err, ansiReset)
		return
	}
	defer watcher.Close()

	// Recursively add all region directories to the watcher
	watchCount := 0
	for regionName := range regionPriority {
		regionPath := filepath.Join(brainRoot, regionName)
		if info, err := os.Stat(regionPath); err == nil && info.IsDir() {
			filepath.Walk(regionPath, func(path string, info os.FileInfo, err error) error {
				if err != nil || !info.IsDir() {
					return nil
				}
				baseName := filepath.Base(path)
				if strings.HasPrefix(baseName, ".") || baseName == "node_modules" ||
					baseName == "_transcripts" || baseName == "_archive" ||
					baseName == ".neuronfs_backup" || baseName == "_agents" {
					return filepath.SkipDir
				}
				if wErr := watcher.Add(path); wErr == nil {
					watchCount++
				}
				return nil
			})
		}
	}
	// Also watch _inbox for agent communication events
	inboxPath := filepath.Join(brainRoot, "_inbox")
	if info, err := os.Stat(inboxPath); err == nil && info.IsDir() {
		watcher.Add(inboxPath)
		watchCount++
	}

	fmt.Printf("%s[NEURON] Core Initialization Complete.%s\n", ansiCyan, ansiReset)
	fmt.Printf("%s[SYNAPSE] Watching %d neural pathways via fsnotify. Zero polling.%s\n", ansiMagenta, watchCount, ansiReset)
	fmt.Printf("%s  - Waiting for synaptic pulses...%s\n", ansiWhite, ansiReset)

	// Initial injection on startup
	processInbox(brainRoot)
	writeAllTiers(brainRoot)

	// Debounce timer: batch rapid filesystem events into a single re-scan
	var debounceTimer *time.Timer
	var debounceMu sync.Mutex
	debounceMs := 500 * time.Millisecond
	lastHash := ""

	triggerRescan := func() {
		start := time.Now()
		brain := scanBrain(brainRoot)
		result := runSubsumption(brain)
		hash := fmt.Sprintf("%d-%s-%d-%d",
			result.FiredNeurons, result.BombSource, result.TotalNeurons, result.TotalCounter)
		if hash != lastHash {
			lastHash = hash
			processInbox(brainRoot)
			writeAllTiers(brainRoot)
			elapsed := time.Since(start)
			if result.BombSource != "" {
				fmt.Printf("%s[TRAUMA] 💀 BOMB detected in %s — cascading shutdown%s\n",
					ansiRed, result.BombSource, ansiReset)
			} else {
				fmt.Printf("%s[PULSE] %d/%d neurons active | Δ activation: %d (%dms)%s\n",
					ansiGreen, result.FiredNeurons, result.TotalNeurons, result.TotalCounter,
					elapsed.Milliseconds(), ansiReset)
			}
		}
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			// .git 디렉토리 이벤트 무시 — git snapshot과 데드락 방지
			relPath, _ := filepath.Rel(brainRoot, event.Name)
			if strings.HasPrefix(relPath, ".git") || strings.Contains(relPath, string(filepath.Separator)+".git") {
				continue
			}
			// Log individual pulse events with biological naming
			ts := time.Now().Format("15:04:05")
			if event.Op&(fsnotify.Create|fsnotify.Write) != 0 {
				fmt.Printf("%s[%s] [PULSE] %s evolved.%s\n", ansiYellow, ts, relPath, ansiReset)
			} else if event.Op&fsnotify.Remove != 0 {
				fmt.Printf("%s[%s] [PRUNE] 데드 시냅스 제거: %s%s\n", ansiDimGray, ts, relPath, ansiReset)
			}

			// If a new directory was created, add it to the watcher (skip .git)
			if event.Op&fsnotify.Create != 0 {
				if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
					baseName := filepath.Base(event.Name)
					if !strings.HasPrefix(baseName, ".") {
						watcher.Add(event.Name)
					}
				}
			}

			// Debounce: reset timer on each event, fire rescan after debounceMs of silence
			debounceMu.Lock()
			if debounceTimer != nil {
				debounceTimer.Stop()
			}
			debounceTimer = time.AfterFunc(debounceMs, triggerRescan)
			debounceMu.Unlock()

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			fmt.Printf("%s[TRAUMA] Watcher error: %v%s\n", ansiRed, err, ansiReset)
		}
	}
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// DIAGNOSTICS
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
