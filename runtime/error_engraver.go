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

	// ── Phase 47: Dedup + Counter + Escalation ──
	// Check if a similar error already exists. If so, increment counter instead of creating new file.
	existingPath, repeatCount := findExistingErrorNeuron(targetDir, module, errMessage)

	if existingPath != "" {
		// EXISTING: increment repeat counter
		repeatCount++
		updateErrorCounter(existingPath, repeatCount)
		fmt.Printf("[Self-Healing] Repeat error #%d: %s\n", repeatCount, filepath.Base(existingPath))
		RecordAudit(brainRoot, "error_engraver", "repeat_increment", existingPath, fmt.Sprintf("count=%d", repeatCount), true)

		// Escalation: 3+ repeats → move to brainstem (P0 = highest priority)
		if repeatCount >= 3 && repeatCount < 5 {
			escalateToP0(brainRoot, existingPath, module, errMessage, repeatCount)
		}
		// Critical: 5+ repeats → BOMB (halt everything, force human review)
		if repeatCount >= 5 {
			bombPath := filepath.Join(brainRoot, "brainstem", "bomb.neuron")
			bombContent := fmt.Sprintf("# BOMB: %s failed %d times\n\nModule: %s\nError: %s\nRepeats: %d\n\n> Auto-triggered by Phase 47 escalation. Remove this file to resume.", module, repeatCount, module, errMessage, repeatCount)
			_ = os.WriteFile(bombPath, []byte(bombContent), 0644)
			fmt.Printf("[BOMB] 🔴 Error repeated %d times. System halted.\n", repeatCount)
			RecordAudit(brainRoot, "error_engraver", "BOMB_escalation", bombPath, fmt.Sprintf("%s failed %dx", module, repeatCount), true)
		}
		return
	}

	// NEW: first occurrence — create error neuron as before
	fileName := fmt.Sprintf("%s_%d.neuron", safeName, time.Now().Unix())
	filePath := filepath.Join(targetDir, fileName)

	content := fmt.Sprintf(`---
module: %s
timestamp: %s
status: auto-engraved
repeat_count: 1
---
# Runtime Error Detected

System autopilot trapped a runtime error.

## Origin Module
%s

## Error Message
%s

> 이 지식은 Phase 34에 의해 자가-각인되었습니다. FS Watcher가 이 파일을 스캔하고 bot1 에이전트 인박스로 할당하여 진화를 트리거합니다.`, module, time.Now().Format(time.RFC3339), module, errMessage)

	if writeErr := os.WriteFile(filePath, []byte(content), 0644); writeErr == nil {
		fmt.Printf("[Self-Healing] NEW error engraved: %s\n", fileName)
		// Phase 32: Webhook Broadcast integration
		TriggerWebhook("SYSTEM_PANIC_ENGRAVED", fmt.Sprintf("Module %s crashed: %s", module, errMessage), map[string]string{
			"module": module,
			"path":   filePath,
		})
		// Phase 40: Audit Trail
		RecordAudit(brainRoot, "error_engraver", "engrave", filePath, errMessage, true)
		// Phase 45: Self-Repair Suggestion
		go SuggestRepair(brainRoot, module, errMessage)
	} else {
		fmt.Printf("[Self-Healing] Failed to engrave error: %v\n", writeErr)
		RecordAudit(brainRoot, "error_engraver", "engrave_failed", filePath, writeErr.Error(), false)
	}
}

// findExistingErrorNeuron scans the error directory for a neuron with the same module.
// Returns its path and current repeat count if found.
func findExistingErrorNeuron(dir, module, errMsg string) (string, int) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", 0
	}
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".neuron" {
			continue
		}
		fp := filepath.Join(dir, e.Name())
		data, readErr := os.ReadFile(fp)
		if readErr != nil {
			continue
		}
		content := string(data)
		// Match by module name in frontmatter
		if strings.Contains(content, "module: "+module) {
			// Extract repeat_count
			count := 1
			for _, line := range strings.Split(content, "\n") {
				if strings.HasPrefix(strings.TrimSpace(line), "repeat_count:") {
					fmt.Sscanf(strings.TrimPrefix(strings.TrimSpace(line), "repeat_count:"), "%d", &count)
				}
			}
			return fp, count
		}
	}
	return "", 0
}

// updateErrorCounter updates the repeat_count in an existing error neuron's frontmatter.
func updateErrorCounter(path string, newCount int) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	content := string(data)
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "repeat_count:") {
			lines[i] = fmt.Sprintf("repeat_count: %d", newCount)
		}
	}
	_ = os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
}

// escalateToP0 copies the error neuron to brainstem/ for highest priority
func escalateToP0(brainRoot, srcPath, module, errMsg string, count int) {
	p0Dir := filepath.Join(brainRoot, "brainstem")
	_ = os.MkdirAll(p0Dir, 0755)

	destName := fmt.Sprintf("禁_repeat_%s.neuron", strings.ReplaceAll(module, "/", "_"))
	destPath := filepath.Join(p0Dir, destName)

	// Don't duplicate if already escalated
	if _, err := os.Stat(destPath); err == nil {
		return
	}

	escalation := fmt.Sprintf(`---
module: %s
repeat_count: %d
escalated: true
priority: P0
---
# ESCALATED: Repeated Error (%dx)

This error has occurred %d times and has been auto-escalated to brainstem (P0).

## Module
%s

## Error
%s

> Phase 47 auto-escalation. This means the error was NOT learned from.
> Fix the root cause, then remove this neuron.`, module, count, count, count, module, errMsg)

	if writeErr := os.WriteFile(destPath, []byte(escalation), 0644); writeErr == nil {
		fmt.Printf("[ESCALATION] Error promoted to P0 (brainstem): %s\n", destName)
		RecordAudit(brainRoot, "error_engraver", "escalate_P0", destPath, fmt.Sprintf("%s repeated %dx", module, count), true)
	}
}
