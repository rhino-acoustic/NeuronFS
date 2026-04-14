// evolve_context.go — 컨텍스트 인식형 넛지 프롬프트 빌더
//
// PROVIDES: hlBuildContextualPrompt, hlReadRecentConversation, hlCheckHealthNudge, hlCheckCorrectionsNudge, hlCheckErrorNudge
// DEPENDS ON: api_handler_system.go (hlCDPInject — nudge delivery)
//
// 설계 원칙:
//   - Go 런타임은 인프라일 뿐, AI 에이전트(CLI)의 상관이 아니다
//   - 넛지는 "직전 대화에서 뭘 했는지 요약 + 그 결과 확인해" 수준
//   - 뭘 해라마라 직접 지시하지 않음 (월권 금지)
//   - 시스템 이상은 참고 정보로만 첨부
//   - 할 일이 없으면 아예 주입하지 않음
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// hlBuildContextualPrompt 는 직전 대화를 요약하고,
// 후속 확인 방향만 제시하는 넛지를 생성한다.
// 할 일이 없으면 ("", false)를 반환하여 주입을 생략한다.
func hlBuildContextualPrompt(brainRoot string) (string, bool) {
	// ── 1. 직전 대화 맥락 읽기 ──
	recentContext := hlReadRecentConversation(brainRoot)
	if recentContext == "" {
		// 직전 대화가 없으면 넛지 자체가 무의미
		return "", false
	}

	// ── 2. 시스템 이상 참고 정보 (있을 때만) ──
	var refs []string
	healthRefs := hlCheckHealthNudge(brainRoot)
	refs = append(refs, healthRefs...)
	corrRef := hlCheckCorrectionsNudge(brainRoot)
	if corrRef != "" {
		refs = append(refs, corrRef)
	}
	errRef := hlCheckErrorNudge(brainRoot)
	if errRef != "" {
		refs = append(refs, errRef)
	}

	// 넛지 조립: 직전 대화 요약 + 확인 방향 + (참고 정보)
	var sb strings.Builder
	sb.WriteString(recentContext)
	sb.WriteString("\n\n")
	sb.WriteString("직전 대화에서 작업한 부분이 개선되었는지 확인해.")

	if len(refs) > 0 {
		sb.WriteString("\n\n참고:\n")
		for _, r := range refs {
			sb.WriteString(fmt.Sprintf("- %s\n", r))
		}
	}

	return sb.String(), true
}

// hlReadRecentConversation 은 직전 대화의 마지막 몇 줄을 읽어서 맥락을 제공한다.
func hlReadRecentConversation(brainRoot string) string {
	jsonlFile := filepath.Join(brainRoot, "_agents", "global_inbox", "transcript_latest.jsonl")
	data, err := os.ReadFile(jsonlFile)
	if err != nil || len(data) == 0 {
		return ""
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	// 최근 3줄만
	if len(lines) > 3 {
		lines = lines[len(lines)-3:]
	}

	var sb strings.Builder
	sb.WriteString("직전 대화 요약:\n")
	for _, line := range lines {
		// 간략화 — JSON에서 role과 text만 추출
		if roleIdx := strings.Index(line, `"role":"`); roleIdx >= 0 {
			rolePart := line[roleIdx+8:]
			if endQ := strings.Index(rolePart, `"`); endQ > 0 {
				role := rolePart[:endQ]
				textPart := ""
				if textIdx := strings.Index(line, `"text":"`); textIdx >= 0 {
					tp := line[textIdx+8:]
					if endT := strings.LastIndex(tp, `"`); endT > 0 {
						textPart = tp[:endT]
					}
				}
				// 텍스트 100자 제한
				if len([]rune(textPart)) > 100 {
					textPart = string([]rune(textPart)[:100]) + "..."
				}
				sb.WriteString(fmt.Sprintf("  %s: %s\n", strings.ToUpper(role), textPart))
			}
		}
	}
	return sb.String()
}

// hlCheckHealthNudge 는 시스템 건강 이상을 부드러운 넛지로 반환
func hlCheckHealthNudge(brainRoot string) []string {
	var nudges []string

	// brainstem 필수 구조 확인
	bsPath := filepath.Join(brainRoot, "brainstem")
	for _, hanja := range []string{"必", "禁"} {
		hp := filepath.Join(bsPath, hanja)
		if _, err := os.Stat(hp); err != nil {
			nudges = append(nudges, fmt.Sprintf("brainstem/%s 폴더 누락", hanja))
			continue
		}
		neuronCount := 0
		filepath.Walk(hp, func(path string, info os.FileInfo, err error) error {
			if err != nil || info == nil {
				return nil
			}
			if !info.IsDir() && strings.HasSuffix(info.Name(), ".neuron") {
				neuronCount++
			}
			return nil
		})
		if neuronCount == 0 {
			nudges = append(nudges, fmt.Sprintf("brainstem/%s 뉴런 0개", hanja))
		}
	}

	// GEMINI.md 마커 체크
	home := os.Getenv("USERPROFILE")
	if home != "" {
		geminiPath := filepath.Join(home, ".gemini", "GEMINI.md")
		data, err := os.ReadFile(geminiPath)
		if err != nil {
			nudges = append(nudges, "GEMINI.md 접근 불가")
		} else {
			content := string(data)
			if !strings.Contains(content, "<!-- NEURONFS:START -->") || !strings.Contains(content, "<!-- NEURONFS:END -->") {
				nudges = append(nudges, "GEMINI.md NeuronFS 마커 손상")
			}
		}
	}

	// bomb 확인
	regionOrder := []string{"brainstem", "limbic", "hippocampus", "sensors", "cortex", "ego", "prefrontal"}
	for _, r := range regionOrder {
		filepath.Walk(filepath.Join(brainRoot, r), func(path string, info os.FileInfo, err error) error {
			if err != nil || info == nil {
				return nil
			}
			if info.Name() == "bomb.neuron" {
				rel, _ := filepath.Rel(brainRoot, path)
				nudges = append(nudges, fmt.Sprintf("%s bomb 활성", rel))
			}
			return nil
		})
	}

	return nudges
}

// hlCheckCorrectionsNudge 는 미처리 교정이 있으면 넛지
func hlCheckCorrectionsNudge(brainRoot string) string {
	corrPath := filepath.Join(brainRoot, "_inbox", "corrections.jsonl")
	data, err := os.ReadFile(corrPath)
	if err != nil || len(data) == 0 {
		return ""
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) == 0 {
		return ""
	}
	return fmt.Sprintf("corrections.jsonl 미처리 %d건", len(lines))
}

// hlCheckErrorNudge 는 최근 에러 로그가 있으면 넛지
func hlCheckErrorNudge(brainRoot string) string {
	nfsRoot := filepath.Dir(brainRoot)
	logPath := filepath.Join(nfsRoot, "logs", "supervisor.log")
	data, err := os.ReadFile(logPath)
	if err != nil {
		return ""
	}

	lines := strings.Split(string(data), "\n")
	var recentErrors []string
	cutoff := time.Now().Add(-30 * time.Minute).Format("15:04")

	for _, line := range lines {
		if strings.Contains(line, "ERROR") || strings.Contains(line, "panic") {
			if len(line) > 9 && line[0] == '[' {
				end := strings.Index(line, "]")
				if end > 0 && end < 10 {
					logTime := line[1:end]
					if logTime >= cutoff {
						recentErrors = append(recentErrors, strings.TrimSpace(line))
					}
				}
			}
		}
	}

	if len(recentErrors) == 0 {
		return ""
	}
	return fmt.Sprintf("최근 30분 에러 %d건", len(recentErrors))
}
