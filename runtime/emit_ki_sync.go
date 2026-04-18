package main

// emit_ki_sync.go — NeuronFS emit → Antigravity KI 하이브리드 동기화 (적층)
// PROVIDES: syncKnowledgeItem
// DEPENDS ON: emit_tiers.go (writeAllTiersForTargets 에서 호출)
//
// 하이브리드 원칙:
//   1. 기존 Antigravity KI (sky_engine, vegavery 등)는 절대 건드리지 않음
//   2. neuronfs_session_progress KI만 관리
//   3. metadata.json이 이미 있으면 기존 필드 보존 + summary/lastAccessed만 갱신
//   4. artifacts는 추가만 (기존 파일 삭제 안 함)

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// syncKnowledgeItem updates the Antigravity Knowledge Item with current brain state.
// Called from writeAllTiersForTargets after every emit.
// Hybrid: preserves existing KI fields, only updates NeuronFS-specific data.
func syncKnowledgeItem(brainRoot string, result SubsumptionResult) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}

	kiDir := filepath.Join(homeDir, ".gemini", "antigravity", "knowledge", "neuronfs_session_progress")
	artDir := filepath.Join(kiDir, "artifacts")
	os.MkdirAll(artDir, 0750)

	// ── 1. session_progress.md → KI artifacts 복사 ──
	progressSrc := filepath.Join(brainRoot, "hippocampus", "전사분석", "_진행상황", "session_progress.md")
	if fileExists(progressSrc) {
		data, err := os.ReadFile(progressSrc)
		if err == nil {
			os.WriteFile(filepath.Join(artDir, "session_progress.md"), data, 0600)
		}
	}

	// ── 2. TODO 큐 현황 → KI artifacts 복사 ──
	todoSrc := filepath.Join(brainRoot, "brainstem", "必사용자요청_TODO적재", "rule.md")
	if fileExists(todoSrc) {
		data, err := os.ReadFile(todoSrc)
		if err == nil {
			os.WriteFile(filepath.Join(artDir, "todo_rule.md"), data, 0600)
		}
	}

	// ── 3. 하이브리드 metadata.json 갱신 ──
	metaPath := filepath.Join(kiDir, "metadata.json")

	// 기존 metadata 읽기 (있으면 보존)
	existing := make(map[string]interface{})
	if data, err := os.ReadFile(metaPath); err == nil {
		json.Unmarshal(data, &existing)
	}

	// NeuronFS 내용만 갱신 (기존 created, references 등 보존)
	existing["title"] = "NeuronFS Session Progress & TODO Queue"
	existing["summary"] = fmt.Sprintf(
		"NeuronFS SSOT. Neurons: %d, Activation: %d, Regions: %d. "+
			"8-step hourly cron (git→backup→prune→context→categorize→log→verify→dedup). "+
			"Role: Antigravity=foreground, Gemini CLI=background, Cron=maintenance. "+
			"TODO queue in session_progress.md. Last emit: %s.",
		result.TotalNeurons, result.TotalCounter, len(result.ActiveRegions),
		time.Now().Format("2006-01-02 15:04"))
	existing["lastAccessed"] = time.Now().Format(time.RFC3339)

	// created가 없으면 설정 (최초 생성 시)
	if _, ok := existing["created"]; !ok {
		existing["created"] = time.Now().Format(time.RFC3339)
	}

	// references가 없으면 설정
	if _, ok := existing["references"]; !ok {
		existing["references"] = []map[string]string{
			{
				"type":  "brain",
				"path":  brainRoot,
				"title": "NeuronFS brain_v4",
			},
		}
	}

	metaData, _ := json.MarshalIndent(existing, "", "  ")
	os.WriteFile(metaPath, metaData, 0600)

	fmt.Println("[EMIT] 🧠 KI 하이브리드 동기화 완료")
}
