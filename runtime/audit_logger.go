package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ============================================================================
// Module: Autonomous Action Audit Logger (Phase 40)
// Records every self-healing and autonomous action as structured JSONL
// so the agent can audit its own past decisions.
// ============================================================================

// AuditEntry represents a single autonomous action record
type AuditEntry struct {
	Timestamp string `json:"ts"`
	Actor     string `json:"actor"`
	Action    string `json:"action"`
	Target    string `json:"target"`
	Reason    string `json:"reason"`
	Success   bool   `json:"success"`
}

var auditMu sync.Mutex

// RecordAudit appends a structured audit entry to hippocampus/audit_trail/audit.jsonl
func RecordAudit(brainRoot, actor, action, target, reason string, success bool) {
	if brainRoot == "" {
		return
	}

	entry := AuditEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Actor:     actor,
		Action:    action,
		Target:    target,
		Reason:    reason,
		Success:   success,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return
	}

	auditDir := filepath.Join(brainRoot, "hippocampus", "audit_trail")
	_ = os.MkdirAll(auditDir, 0755)

	auditFile := filepath.Join(auditDir, "audit.jsonl")

	auditMu.Lock()
	defer auditMu.Unlock()

	f, err := os.OpenFile(auditFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	_, _ = f.Write(data)
	_, _ = f.WriteString("\n")

	fmt.Printf("[Audit] %s → %s: %s\n", actor, action, target)
}
