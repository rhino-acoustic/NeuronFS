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
// Module: Runtime Error Auto-Engraver (Phase 34)
// Capture panics and errors and engrave them into the cortex as .neuron files.
// FS Watcher will detect them and trigger a Swarm Action automatically.
// ============================================================================

var (
	errorDebounceMap = make(map[string]time.Time)
	errorDebounceMu  sync.Mutex
)

// EngraveRuntimeError writes an error stack/pattern into NeuronFS cortex directory
// and triggers the OS FS Watcher to kick off self-healing feedback loops.
func EngraveRuntimeError(brainRoot string, module string, errMessage string) {
	if brainRoot == "" {
		return
	}

	errorDebounceMu.Lock()
	defer errorDebounceMu.Unlock()

	// Debounce exact same errors for 15 seconds to prevent filesystem spam
	hashKey := module + "_" + errMessage
	if lastTime, ok := errorDebounceMap[hashKey]; ok && time.Since(lastTime) < 15*time.Second {
		return
	}
	errorDebounceMap[hashKey] = time.Now()

	// Target Directory
	targetDir := filepath.Join(brainRoot, "cortex", "dev", "에러패턴확인", "런타임에러_자동수집")
	_ = os.MkdirAll(targetDir, 0755)

	// Clean the error message to make a safe filename
	safeName := strings.ReplaceAll(errMessage, " ", "_")
	safeName = strings.ReplaceAll(safeName, "/", "_")
	safeName = strings.ReplaceAll(safeName, "\\", "_")
	safeName = strings.ReplaceAll(safeName, ":", "")
	if len(safeName) > 40 {
		safeName = safeName[:40]
	}

	fileName := fmt.Sprintf("%s_%d.neuron", safeName, time.Now().Unix())
	filePath := filepath.Join(targetDir, fileName)

	content := fmt.Sprintf(`---
module: %s
timestamp: %s
status: auto-engraved
---
# Runtime Error Detected

System autopilot trapped a runtime error.

## Origin Module
%s

## Error Message
%s

> 이 지식은 Phase 34에 의해 자가-각인되었습니다. FS Watcher가 이 파일을 스캔하고 bot1 에이전트 인박스로 할당하여 진화를 트리거합니다.`, module, time.Now().Format(time.RFC3339), module, errMessage)

	if writeErr := os.WriteFile(filePath, []byte(content), 0644); writeErr == nil {
		fmt.Printf("[Self-Healing] Runtimer Error Auto-Engraved at: %s\n", fileName)
		// Phase 32: Webhook Broadcast integration
		TriggerWebhook("SYSTEM_PANIC_ENGRAVED", fmt.Sprintf("Module %s crashed: %s", module, errMessage), map[string]string{
			"module": module,
			"path":   filePath,
		})
		// Phase 40: Audit Trail
		RecordAudit(brainRoot, "error_engraver", "engrave", filePath, errMessage, true)
	} else {
		fmt.Printf("[Self-Healing] Failed to engrave error: %v\n", writeErr)
		RecordAudit(brainRoot, "error_engraver", "engrave_failed", filePath, writeErr.Error(), false)
	}
}
