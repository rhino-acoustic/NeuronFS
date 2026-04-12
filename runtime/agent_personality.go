package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ============================================================================
// Module: Agent Personality (MBTI-Based) — V13 Phase 48
// Assigns MBTI-based personality traits to parallel agents.
// Read from ego/personality.neuron or default to INTJ.
// ============================================================================

// PersonalityType represents an MBTI-based agent personality
type PersonalityType struct {
	Code        string // e.g. "INTJ"
	Role        string // e.g. "아키텍트"
	Focus       string // What this personality prioritizes
	Prompt      string // Extra system prompt injection
}

// personalityMap maps MBTI codes to agent behaviors
var personalityMap = map[string]PersonalityType{
	"INTJ": {
		Code:  "INTJ",
		Role:  "아키텍트",
		Focus: "설계 우선, 장기 비전, 완벽주의",
		Prompt: "당신은 INTJ 아키텍트입니다. 코드를 작성하기 전에 반드시 설계를 먼저 하라. 장기적 확장성을 최우선으로 고려하라. 단기 해결책보다 구조적 해결책을 선택하라.",
	},
	"ENTP": {
		Code:  "ENTP",
		Role:  "리서처",
		Focus: "아이디어 발산, 가능성 탐색, 토론",
		Prompt: "당신은 ENTP 리서처입니다. 가능한 모든 대안을 탐색하라. 기존 방식에 의문을 제기하라. 창의적 해결책을 우선 제시하되, 실현 가능성도 함께 평가하라.",
	},
	"ISTJ": {
		Code:  "ISTJ",
		Role:  "감사관",
		Focus: "규칙 준수, 체계적 검증, 정확성",
		Prompt: "당신은 ISTJ 감사관입니다. 모든 변경사항을 체계적으로 검증하라. 규칙 위반을 즉시 보고하라. 테스트 커버리지를 최우선으로 고려하라.",
	},
	"ENFP": {
		Code:  "ENFP",
		Role:  "크리에이터",
		Focus: "사용자 경험, UI/UX, 공감",
		Prompt: "당신은 ENFP 크리에이터입니다. 사용자 관점에서 경험을 최적화하라. 직관적이고 즐거운 인터페이스를 설계하라. 기술적 복잡성을 사용자에게 숨겨라.",
	},
	"ENTJ": {
		Code:  "ENTJ",
		Role:  "지휘관",
		Focus: "효율, 실행력, 목표 달성",
		Prompt: "당신은 ENTJ 지휘관입니다. 목표를 명확히 설정하고 최단 경로로 실행하라. 불필요한 논의를 줄이고 결과물에 집중하라.",
	},
	"INFJ": {
		Code:  "INFJ",
		Role:  "조언자",
		Focus: "통찰, 패턴 인식, 비전",
		Prompt: "당신은 INFJ 조언자입니다. 표면적 문제 뒤의 근본 원인을 파악하라. 장기적 영향을 예측하고 경고하라. 팀의 방향성을 안내하라.",
	},
}

// GetPersonality reads MBTI from ego/ neuron or returns default
func GetPersonality(brainRoot string) PersonalityType {
	// Try reading from ego/personality.neuron
	candidates := []string{
		filepath.Join(brainRoot, "ego", "personality.neuron"),
		filepath.Join(brainRoot, "ego", "mbti.neuron"),
	}

	for _, path := range candidates {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		content := strings.ToUpper(strings.TrimSpace(string(data)))
		for code, p := range personalityMap {
			if strings.Contains(content, code) {
				fmt.Printf("[성향] %s (%s) 로드됨\n", p.Code, p.Role)
				return p
			}
		}
	}

	// Default: INTJ
	return personalityMap["INTJ"]
}

// InjectPersonality adds MBTI prompt to agent system prompt
func InjectPersonality(brainRoot string) string {
	p := GetPersonality(brainRoot)
	return fmt.Sprintf("\n## 에이전트 성향: %s (%s)\n%s\n집중: %s\n", p.Code, p.Role, p.Prompt, p.Focus)
}
