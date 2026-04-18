package main

// emit_ki_sync.go — NeuronFS emit → Antigravity KI 자동 동기화 (적층)
// PROVIDES: syncKnowledgeItem
// DEPENDS ON: emit_tiers.go (writeAllTiersForTargets 에서 호출)

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// syncKnowledgeItem updates the Antigravity Knowledge Item with current brain state.
// Called from writeAllTiersForTargets after every emit.
// This makes KI a derivative of NeuronFS — not a separate system.
func syncKnowledgeItem(brainRoot string, result SubsumptionResult) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}

	kiDir := filepath.Join(homeDir, ".gemini", "antigravity", "knowledge", "neuronfs_session_progress")
	artDir := filepath.Join(kiDir, "artifacts")
	os.MkdirAll(artDir, 0750)

	// 1. session_progress.md가 있으면 KI artifacts로 복사
	progressSrc := filepath.Join(brainRoot, "hippocampus", "전사분석", "_진행상황", "session_progress.md")
	if fileExists(progressSrc) {
		data, err := os.ReadFile(progressSrc)
		if err == nil {
			os.WriteFile(filepath.Join(artDir, "session_progress.md"), data, 0600)
		}
	}

	// 2. metadata.json 갱신 (뉴런 수, 활성화 등 최신 상태 반영)
	meta := map[string]interface{}{
		"title": "NeuronFS Session Progress & TODO Queue",
		"summary": fmt.Sprintf(
			"SSOT for NeuronFS session progress. Neurons: %d, Activation: %d, Regions: %d. "+
				"Contains completed items, pending TODO queue, 8-step hourly cron cycle, "+
				"and role assignments. Must be read at session start. Last emit: %s.",
			result.TotalNeurons, result.TotalCounter, len(result.ActiveRegions),
			time.Now().Format("2006-01-02 15:04")),
		"created":      "2026-04-18T09:30:00+09:00",
		"lastAccessed": time.Now().Format(time.RFC3339),
		"references": []map[string]string{
			{
				"type":  "brain",
				"path":  brainRoot,
				"title": "NeuronFS brain_v4",
			},
		},
	}

	metaData, _ := json.MarshalIndent(meta, "", "  ")
	os.WriteFile(filepath.Join(kiDir, "metadata.json"), metaData, 0600)

	fmt.Println("[EMIT] 🧠 KI 동기화 완료 (Antigravity knowledge)")
}
