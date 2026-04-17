package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// ============================================================================
// Module: Sensory Network OS Event Watcher (Phase 31)
// Concept: Real-time file system monitoring with Debouncing to prevent
// OOM panics during massive git operations or IDE refactoring.
// ============================================================================

// debounceMap track recent events to prevent storm
var (
	debounceMap = make(map[string]time.Time)
	debounceMu  sync.Mutex
)

// debounceThreshold is the minimum time between identical file alerts
const debounceThreshold = 1500 * time.Millisecond

// startFSWatcherPool initializes the fsnotify watcher and begins scanning
// specific cortex memory regions for mutations
func startFSWatcherPool(brainRoot string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Printf("[Sensory] Error creating FS Watcher: %v\n", err)
		return
	}
	// Cannot defer Close here because it's a daemon
	
	// Listen for events in the background
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				
				// Focus only on Creates and Writes
				if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
					handleFileEvent(brainRoot, event.Name, event.Op.String())
				}
				
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Printf("[Sensory] Watcher error: %v\n", err)
			}
		}
	}()

	// We recursively add directories to the watcher. 
	// To save OS file descriptors, we only watch high-priority zones:
	zonesToWatch := []string{
		filepath.Join(brainRoot, "cortex", "dev", "작업"),
		filepath.Join(brainRoot, "cortex", "dev", "에러패턴확인"),
	}

	for _, zone := range zonesToWatch {
		addWatchRecursive(watcher, zone)
	}

	// ── debounceMap GC: 5분마다 오래된 항목 정리 (메모리 누수 방지) ──
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			debounceMu.Lock()
			now := time.Now()
			for k, v := range debounceMap {
				if now.Sub(v) > 5*time.Minute {
					delete(debounceMap, k)
				}
			}
			debounceMu.Unlock()
		}
	}()

	fmt.Println("[Sensory] FS Watcher daemon activated. OS changes are now synaptic triggers.")
}

func addWatchRecursive(watcher *fsnotify.Watcher, rootDir string) {
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && info.IsDir() {
			// Skip .git or hidden dirs
			if strings.HasPrefix(filepath.Base(path), ".") {
				return filepath.SkipDir
			}
			watcher.Add(path)
		}
		return nil
	})
	if err != nil {
		fmt.Printf("[Sensory] Warning: could not watch %s: %v\n", rootDir, err)
	}
}

func handleFileEvent(brainRoot, changedPath, opType string) {
	debounceMu.Lock()
	defer debounceMu.Unlock()

	lastTime, exists := debounceMap[changedPath]
	if exists && time.Since(lastTime) < debounceThreshold {
		// Suppress event (debounce)
		return
	}
	debounceMap[changedPath] = time.Now()

	// Launch event processor asynchronously so it doesn't block the watcher listener
	go processSensoryEvent(brainRoot, changedPath, opType)
}

func processSensoryEvent(brainRoot, changedPath, opType string) {
	// Must only process .neuron, .md or .go files
	ext := filepath.Ext(changedPath)
	if ext != ".neuron" && ext != ".md" && ext != ".go" && ext != ".html" {
		return
	}

	// Phase 37: V9 Temporal Log (4D Memory) snapshot packing
	RecordTemporalSnapshot(brainRoot, changedPath)

	// Dump alert to bot1 inbox
	inboxDir := filepath.Join(brainRoot, "_agents", "bot1", "inbox")
	os.MkdirAll(inboxDir, 0755)

	alertId := time.Now().UnixNano()
	filePath := filepath.Join(inboxDir, fmt.Sprintf("sensory_alert_%d.md", alertId))

	content := fmt.Sprintf("발신: OS Sensory Network (V7)\n# OS 레벨 감지 보고: 파싱/수정 이벤트\n\n- 대상 파일: `%s`\n- 액션: `%s`\n\n인간 설계자 혹은 타 프로세스가 파일을 조작했습니다. 즉각 확인 및 분석(Linting)이 필요합니다.\n", changedPath, opType)
	
	_ = os.WriteFile(filePath, []byte(content), 0644)

	if GlobalSSEBroker != nil {
		GlobalSSEBroker.Broadcast("info", fmt.Sprintf("[Sensory Alert] %s 감지됨 (%s)", filepath.Base(changedPath), opType))
	}

	// Phase 32: B2B Integration - Push Webhook when OS file changes
	TriggerWebhook("OS_SENSORY_EVENT", fmt.Sprintf("File %s has been %s", changedPath, opType), map[string]string{
		"path": changedPath,
		"op":   opType,
	})
}
