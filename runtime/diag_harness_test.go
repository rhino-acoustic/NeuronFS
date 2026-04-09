package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunHarness_AllPass(t *testing.T) {
	dir := t.TempDir()

	// 7개 영역 생성
	for _, r := range RegionOrder {
		os.MkdirAll(filepath.Join(dir, r), 0750)
		os.WriteFile(filepath.Join(dir, r, "_rules.md"), []byte("# rules"), 0600)
	}

	// brainstem/必, brainstem/禁 구조 생성
	mustDir := filepath.Join(dir, "brainstem", "必")
	banDir := filepath.Join(dir, "brainstem", "禁")
	os.MkdirAll(filepath.Join(mustDir, "test_neuron"), 0750)
	os.MkdirAll(filepath.Join(banDir, "test_ban"), 0750)

	// GEMINI.md 마커 생성
	home := os.Getenv("USERPROFILE")
	if home == "" {
		t.Skip("USERPROFILE not set")
	}
	geminiDir := filepath.Join(home, ".gemini")
	os.MkdirAll(geminiDir, 0750)
	geminiPath := filepath.Join(geminiDir, "GEMINI.md")
	// 기존 파일 백업
	existing, _ := os.ReadFile(geminiPath)
	defer os.WriteFile(geminiPath, existing, 0600)

	os.WriteFile(geminiPath, []byte("<!-- NEURONFS:START -->\ntest\n<!-- NEURONFS:END -->"), 0600)

	var logs []string
	logger := func(msg string) { logs = append(logs, msg) }

	RunHarness(dir, logger)

	// 로그에 PASS가 있어야 함
	found := false
	for _, l := range logs {
		if len(l) > 0 && l[0] == 0xe2 { // ✅ UTF-8 prefix
			found = true
		}
	}
	if !found {
		t.Errorf("harness did not pass, logs: %v", logs)
	}
}

func TestRunHarness_NilLogger(t *testing.T) {
	dir := t.TempDir()
	for _, r := range RegionOrder {
		os.MkdirAll(filepath.Join(dir, r), 0750)
	}
	// nil logger로 패닉 없이 실행
	RunHarness(dir, nil)
}

func TestRunHarness_MissingRegion(t *testing.T) {
	dir := t.TempDir()
	// 일부 영역만 생성 — 실패 기대
	os.MkdirAll(filepath.Join(dir, "brainstem"), 0750)

	var logs []string
	RunHarness(dir, func(msg string) { logs = append(logs, msg) })

	// 실패 로그가 있어야 함
	if len(logs) < 2 {
		t.Errorf("expected failure logs for missing regions, got %d", len(logs))
	}
}
