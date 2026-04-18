package main

// 자동 헬스 리포트 — Verification-on-Resume 패턴
// 매시간 크론에서 실행. 결과를 brainstem/_health.md에 기록.
// AI가 새 세션을 시작하면 이 파일을 읽어 시스템 상태를 즉시 파악.
// PROVIDES: runHealthReport
// DEPENDS ON: transcript_categorizer.go (cron loop)

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func runHealthReport(brainRoot, nfsRoot string) {
	svLog("[HEALTH] 🔍 자동 헬스 체크 시작")

	var checks []string
	var fails int

	// 1. 뉴런 수
	neuronCount := countNeuronFiles(brainRoot)
	if neuronCount < 100 {
		checks = append(checks, fmt.Sprintf("❌ NEURONS: %d (최소 100 필요!)", neuronCount))
		fails++
	} else {
		checks = append(checks, fmt.Sprintf("✅ NEURONS: %d", neuronCount))
	}

	// 2. _rules.md 8개 영역
	regions := []string{"brainstem", "cortex", "ego", "hippocampus", "limbic", "prefrontal", "sensors", "shared"}
	missingRules := 0
	for _, r := range regions {
		if !fileExists(filepath.Join(brainRoot, r, "_rules.md")) {
			missingRules++
		}
	}
	if missingRules > 0 {
		checks = append(checks, fmt.Sprintf("❌ RULES: %d/%d 누락", missingRules, len(regions)))
		fails++
	} else {
		checks = append(checks, "✅ RULES: 8/8 영역")
	}

	// 3. GEMINI.md
	home, _ := os.UserHomeDir()
	geminiPath := filepath.Join(home, ".gemini", "GEMINI.md")
	if !fileExists(geminiPath) {
		checks = append(checks, "❌ GEMINI.md: 누락")
		fails++
	} else {
		data, _ := os.ReadFile(geminiPath)
		s := string(data)
		hasIdentity := strings.Contains(s, "<identity>")
		hasNeuronFS := strings.Contains(s, "NEURONFS:START")
		if !hasIdentity || !hasNeuronFS {
			checks = append(checks, "❌ GEMINI.md: identity/NeuronFS 블록 누락")
			fails++
		} else {
			checks = append(checks, "✅ GEMINI.md: identity + NeuronFS 정상")
		}
	}

	// 4. 코드맵
	codemapDir := filepath.Join(brainRoot, "cortex", "dev", "_codemap")
	codemapCount := 0
	if entries, err := os.ReadDir(codemapDir); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				codemapCount++
			}
		}
	}
	if codemapCount < 10 {
		checks = append(checks, fmt.Sprintf("⚠️ CODEMAP: %d (최소 10 필요)", codemapCount))
	} else {
		checks = append(checks, fmt.Sprintf("✅ CODEMAP: %d 뉴런", codemapCount))
	}

	// 5. Hook
	settingsPath := filepath.Join(nfsRoot, ".gemini", "settings.json")
	if !fileExists(settingsPath) {
		checks = append(checks, "❌ HOOKS: settings.json 없음")
		fails++
	} else {
		checks = append(checks, "✅ HOOKS: settings.json 존재")
	}

	// 6. git 추적
	out, err := exec.Command("git", "-C", nfsRoot, "ls-files", "brain_v4/").Output()
	gitTracked := 0
	if err == nil {
		s := strings.TrimSpace(string(out))
		if s != "" {
			gitTracked = len(strings.Split(s, "\n"))
		}
	}
	if gitTracked < 100 {
		checks = append(checks, fmt.Sprintf("❌ GIT: brain_v4 %d files tracked (위험!)", gitTracked))
		fails++
	} else {
		checks = append(checks, fmt.Sprintf("✅ GIT: brain_v4 %d files tracked", gitTracked))
	}

	// 7. corrections.jsonl
	corrPath := filepath.Join(brainRoot, "_inbox", "corrections.jsonl")
	if fileExists(corrPath) {
		data, _ := os.ReadFile(corrPath)
		s := strings.TrimSpace(string(data))
		lines := 0
		if s != "" {
			lines = len(strings.Split(s, "\n"))
		}
		checks = append(checks, fmt.Sprintf("✅ SELF-HEAL: %d corrections", lines))
	}

	// 8. hippocampus memory
	memDir := filepath.Join(brainRoot, "hippocampus", "session_log")
	memCount := 0
	if entries, err := os.ReadDir(memDir); err == nil {
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), ".neuron") {
				memCount++
			}
		}
	}
	checks = append(checks, fmt.Sprintf("✅ MEMORY: %d 에피소드", memCount))

	// 결과 집계
	status := "🟢 HEALTHY"
	if fails > 0 {
		status = fmt.Sprintf("🔴 UNHEALTHY (%d fails)", fails)
	}

	// brainstem/_health.md에 기록 (AI가 매 세션 읽음)
	report := fmt.Sprintf(`# NeuronFS Health Report
> Auto-generated: %s
> Status: %s

## Checks
%s

## Quick Stats
- Neurons: %d
- Codemap: %d
- Git tracked: %d
- Memory: %d episodes
- Fails: %d
`, time.Now().Format("2006-01-02 15:04:05"),
		status,
		"- "+strings.Join(checks, "\n- "),
		neuronCount, codemapCount, gitTracked, memCount, fails)

	healthPath := filepath.Join(brainRoot, "brainstem", "_health.md")
	os.WriteFile(healthPath, []byte(report), 0600)

	svLog(fmt.Sprintf("[HEALTH] %s (neurons=%d, codemap=%d, git=%d, fails=%d)",
		status, neuronCount, codemapCount, gitTracked, fails))
}
