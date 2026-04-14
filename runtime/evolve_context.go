// evolve_context.go — 컨텍스트 인식형 동적 진화 프롬프트 빌더
// 정적 마스터 프롬프트를 대체하여, 실제 시스템 상태를 수집한 뒤
// 할 일이 있을 때만 구체적인 지시를 조립하여 주입한다.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// hlBuildContextualPrompt gathers real system state and produces a targeted prompt.
// Returns ("", false) if there's nothing meaningful to inject.
func hlBuildContextualPrompt(brainRoot string) (string, bool) {
	var findings []string

	// ── 1. Harness 결과 확인 ──
	harnessFindings := hlCheckHarnessState(brainRoot)
	if len(harnessFindings) > 0 {
		findings = append(findings, fmt.Sprintf("[긴급: Harness 실패 %d건]\n%s", len(harnessFindings), strings.Join(harnessFindings, "\n")))
	}

	// ── 2. corrections.jsonl 미처리 교정 항목 확인 ──
	corrFindings := hlCheckCorrections(brainRoot)
	if corrFindings != "" {
		findings = append(findings, corrFindings)
	}

	// ── 3. bomb 상태 확인 ──
	bombFindings := hlCheckBombs(brainRoot)
	if bombFindings != "" {
		findings = append(findings, bombFindings)
	}

	// ── 4. 최근 에러 로그 확인 ──
	errFindings := hlCheckRecentErrors(brainRoot)
	if errFindings != "" {
		findings = append(findings, errFindings)
	}

	// 할 일이 없으면 주입하지 않음
	if len(findings) == 0 {
		return "", false
	}

	// 동적 프롬프트 조립
	var sb strings.Builder
	sb.WriteString("[NeuronFS 자율 진화: 컨텍스트 기반 지시]\n")
	sb.WriteString("시스템 스캔 결과 아래 항목이 발견되었습니다. 우선순위대로 처리하라.\n\n")

	for i, f := range findings {
		sb.WriteString(fmt.Sprintf("%d. %s\n\n", i+1, f))
	}

	sb.WriteString("[원칙]\n")
	sb.WriteString("- 기존 코드 적층(Strangler Fig), 삭제 금지\n")
	sb.WriteString("- go vet → go build 검증 후 커밋\n")
	sb.WriteString("- 반드시 한국어로 대답\n")
	sb.WriteString("- 완료 후 결과를 간결하게 보고하라\n")

	return sb.String(), true
}

// hlCheckHarnessState runs a lightweight harness scan inline
func hlCheckHarnessState(brainRoot string) []string {
	var fails []string
	bsPath := filepath.Join(brainRoot, "brainstem")

	for _, hanja := range []string{"必", "禁"} {
		hp := filepath.Join(bsPath, hanja)
		if _, err := os.Stat(hp); err != nil {
			fails = append(fails, fmt.Sprintf("brainstem/%s 폴더 누락", hanja))
			continue
		}
		entries, _ := os.ReadDir(hp)
		neuronCount := 0
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".neuron") {
				neuronCount++
			}
		}
		if neuronCount == 0 {
			fails = append(fails, fmt.Sprintf("brainstem/%s 뉴런 0개", hanja))
		}
	}

	// GEMINI.md 마커 체크
	home := os.Getenv("USERPROFILE")
	if home != "" {
		geminiPath := filepath.Join(home, ".gemini", "GEMINI.md")
		data, err := os.ReadFile(geminiPath)
		if err != nil {
			fails = append(fails, "GEMINI.md 없음")
		} else {
			content := string(data)
			if !strings.Contains(content, "<!-- NEURONFS:START -->") || !strings.Contains(content, "<!-- NEURONFS:END -->") {
				fails = append(fails, "GEMINI.md NeuronFS 마커 손상")
			}
		}
	}
	return fails
}

// hlCheckCorrections looks for unprocessed correction entries
func hlCheckCorrections(brainRoot string) string {
	corrPath := filepath.Join(brainRoot, "_inbox", "corrections.jsonl")
	data, err := os.ReadFile(corrPath)
	if err != nil || len(data) == 0 {
		return ""
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) == 0 {
		return ""
	}
	// 최근 5개만 요약
	recent := lines
	if len(recent) > 5 {
		recent = recent[len(recent)-5:]
	}
	return fmt.Sprintf("[미처리 교정 %d건] 최근:\n%s", len(lines), strings.Join(recent, "\n"))
}

// hlCheckBombs scans for active bomb.neuron files
func hlCheckBombs(brainRoot string) string {
	bombCount := 0
	var bombPaths []string
	regionOrder := []string{"brainstem", "limbic", "hippocampus", "sensors", "cortex", "ego", "prefrontal"}
	for _, r := range regionOrder {
		filepath.Walk(filepath.Join(brainRoot, r), func(path string, info os.FileInfo, err error) error {
			if err != nil || info == nil {
				return nil
			}
			if info.Name() == "bomb.neuron" {
				bombCount++
				rel, _ := filepath.Rel(brainRoot, path)
				bombPaths = append(bombPaths, rel)
			}
			return nil
		})
	}
	if bombCount == 0 {
		return ""
	}
	return fmt.Sprintf("[🚨 BOMB 활성 %d건] 경로: %s → 즉시 해제 필요", bombCount, strings.Join(bombPaths, ", "))
}

// hlCheckRecentErrors scans supervisor log for recent errors
func hlCheckRecentErrors(brainRoot string) string {
	nfsRoot := filepath.Dir(brainRoot)
	logPath := filepath.Join(nfsRoot, "logs", "supervisor.log")
	data, err := os.ReadFile(logPath)
	if err != nil {
		return ""
	}
	lines := strings.Split(string(data), "\n")
	var recentErrors []string
	cutoff := time.Now().Add(-10 * time.Minute).Format("15:04")

	for _, line := range lines {
		if len(line) < 10 {
			continue
		}
		// 로그 형식: [HH:MM:SS] ...
		if strings.Contains(line, "ERROR") || strings.Contains(line, "panic") || strings.Contains(line, "FAIL") {
			// 최근 10분 내 로그만
			if len(line) > 9 {
				logTime := ""
				if line[0] == '[' {
					end := strings.Index(line, "]")
					if end > 0 && end < 10 {
						logTime = line[1:end]
					}
				}
				if logTime >= cutoff || logTime == "" {
					recentErrors = append(recentErrors, strings.TrimSpace(line))
				}
			}
		}
	}

	if len(recentErrors) == 0 {
		return ""
	}
	if len(recentErrors) > 5 {
		recentErrors = recentErrors[len(recentErrors)-5:]
	}
	return fmt.Sprintf("[최근 에러 %d건]\n%s", len(recentErrors), strings.Join(recentErrors, "\n"))
}
