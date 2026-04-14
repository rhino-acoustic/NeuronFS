// evolve_context.go — 컨텍스트 인식형 넛지 프롬프트 빌더
//
// 설계 원칙:
//   - Go 런타임은 인프라일 뿐, AI 에이전트(CLI)의 상관이 아니다
//   - 직전 대화 맥락 + 시스템 상태를 읽고, 부드러운 넛지만 생성
//   - "~해라" 명령이 아닌 "~봐봐", "~확인해볼래?" 수준의 제안
//   - 할 일이 없으면 아예 주입하지 않음
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// hlBuildContextualPrompt 는 시스템 상태를 수집한 뒤,
// 할 일이 있으면 부드러운 넛지 프롬프트를 반환한다.
// 할 일이 없으면 ("", false)를 반환하여 주입을 생략한다.
func hlBuildContextualPrompt(brainRoot string) (string, bool) {
	var nudges []string

	// ── 1. 직전 대화 맥락 읽기 ──
	recentContext := hlReadRecentConversation(brainRoot)

	// ── 2. 시스템 건강 체크 (Harness 실패, bomb 등) ──
	healthNudges := hlCheckHealthNudge(brainRoot)
	nudges = append(nudges, healthNudges...)

	// ── 3. corrections.jsonl 미처리 항목 ──
	corrNudge := hlCheckCorrectionsNudge(brainRoot)
	if corrNudge != "" {
		nudges = append(nudges, corrNudge)
	}

	// ── 4. 에러 로그 ──
	errNudge := hlCheckErrorNudge(brainRoot)
	if errNudge != "" {
		nudges = append(nudges, errNudge)
	}

	// 할 일이 없으면 주입 안 함
	if len(nudges) == 0 {
		return "", false
	}

	// 부드러운 넛지 프롬프트 조립
	var sb strings.Builder
	if recentContext != "" {
		sb.WriteString(recentContext)
		sb.WriteString("\n\n")
	}

	sb.WriteString("참고로 시스템을 보니까:\n")
	for _, n := range nudges {
		sb.WriteString(fmt.Sprintf("- %s\n", n))
	}
	sb.WriteString("\n시간 될 때 한번 봐줄래?")

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
			nudges = append(nudges, fmt.Sprintf("brainstem/%s 폴더가 없는 것 같은데 확인해볼래?", hanja))
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
			nudges = append(nudges, fmt.Sprintf("brainstem/%s에 뉴런이 하나도 없어 봐봐", hanja))
		}
	}

	// GEMINI.md 마커 체크
	home := os.Getenv("USERPROFILE")
	if home != "" {
		geminiPath := filepath.Join(home, ".gemini", "GEMINI.md")
		data, err := os.ReadFile(geminiPath)
		if err != nil {
			nudges = append(nudges, "GEMINI.md 파일 접근이 안 되는 것 같아 확인 부탁")
		} else {
			content := string(data)
			if !strings.Contains(content, "<!-- NEURONFS:START -->") || !strings.Contains(content, "<!-- NEURONFS:END -->") {
				nudges = append(nudges, "GEMINI.md에 NeuronFS 마커가 깨진 것 같으니 한번 봐줘")
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
				nudges = append(nudges, fmt.Sprintf("%s에 bomb이 떠 있어 — 긴급하니까 먼저 봐봐", rel))
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
	return fmt.Sprintf("corrections.jsonl에 %d건 쌓여있는데 시간 나면 정리해봐", len(lines))
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
		if strings.Contains(line, "ERROR") || strings.Contains(line, "panic") || strings.Contains(line, "FAIL") {
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
	return fmt.Sprintf("최근 30분 로그에 에러가 %d건 보이는데 한번 확인해볼래?", len(recentErrors))
}
