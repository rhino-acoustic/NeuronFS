// RULE: ZERO EXTERNAL DEPENDENCY PRESERVED
// supervisor.go — NeuronFS 네이티브 프로세스 매니저
//
// watchdog.ps1 + 프로세스 관리를 Go 바이너리로 통합.
//
// Usage: neuronfs <brain_path> --supervisor
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func svWriteMetrics(children []*ChildSpec, nfsRoot string) {
	type svcEntry struct {
		Name     string `json:"name"`
		Status   string `json:"status"`
		Restarts int    `json:"restarts"`
	}
	svcs := make([]svcEntry, 0)
	for _, c := range children {
		c.mu.Lock()
		st := "DOWN"
		if c.running {
			st = "UP"
		}
		svcs = append(svcs, svcEntry{Name: c.Name, Status: st, Restarts: c.restartCount})
		c.mu.Unlock()
	}
	data := map[string]interface{}{
		"ts":         time.Now().Format(time.RFC3339),
		"bootTime":   svBootTime.Format(time.RFC3339),
		"uptimeMs":   time.Since(svBootTime).Milliseconds(),
		"checkCount": svCheckCount,
		"services":   svcs,
		"engine":     "go-supervisor",
	}
	jsonData, _ := json.MarshalIndent(data, "", "  ")
	os.WriteFile(filepath.Join(nfsRoot, "logs", "watchdog_metrics.json"), jsonData, 0600)
}

func svStatusReport(children []*ChildSpec, nfsRoot string) {
	uptime := time.Since(svBootTime).Minutes()
	report := fmt.Sprintf("📊 [NeuronFS Supervisor]\n⏱️ 가동: %.0f분 | 체크: %d회\n", uptime, svCheckCount)
	for _, c := range children {
		c.mu.Lock()
		st := "🟢 UP"
		if !c.running {
			st = "🔴 DOWN"
		}
		report += fmt.Sprintf("%s %s restarts=%d\n", st, c.Name, c.restartCount)
		c.mu.Unlock()
	}
	svLog(report)
	hlTgSend(hlTgChatID, report)
	svWriteMetrics(children, nfsRoot)
}
