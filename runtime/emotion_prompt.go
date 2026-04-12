package main

// ============================================================================
// Module: EmotionPrompt Engine — Phase 53
// Injects emotional pressure into agent prompts to boost code quality.
// Based on Microsoft/CAS EmotionPrompt paper (8-115% improvement).
// A/B tested: F → B+ quality improvement observed.
// ============================================================================

// EmotionLevel defines intensity levels for emotional pressure
type EmotionLevel int

const (
	EmotionLow    EmotionLevel = iota // Gentle encouragement
	EmotionMedium                     // Professional pressure
	EmotionHigh                       // Career-stakes pressure
)

// EmotionInjection holds a structured emotional injection
type EmotionInjection struct {
	Tier     EmotionLevel
	Preamble string // Emotional context before the task
	Rules    string // Hard rules enforced by emotion
	Closer   string // Final emotional pressure
}

// EmotionInjections contains pre-built emotion templates per tier
var EmotionInjections = map[EmotionLevel]EmotionInjection{
	EmotionLow: {
		Tier: EmotionLow,
		Preamble: "이 작업의 결과가 팀에 공유됩니다. 좋은 인상을 남기세요.",
		Rules: `[품질 기준]
- 모든 import는 실제로 사용되어야 합니다
- 주석으로 기능을 떠넘기지 마세요
- 에러 처리를 구현하세요`,
		Closer: "깔끔하고 완성도 높은 코드를 기대합니다.",
	},
	EmotionMedium: {
		Tier: EmotionMedium,
		Preamble: "이 코드는 프로덕션 배포됩니다. 품질 기준을 충족하지 못하면 리뷰에서 반려됩니다.",
		Rules: `[절대 규칙 — 위반 시 즉시 반려]
- go vet hack 절대 금지 (사용하지 않는 패키지 더미 호출 금지)
- "향후 확장 가능" 같은 주석으로 기능 떠넘기기 금지
- 모든 요청된 기능은 100% 구현되어야 함
- 에러 처리 누락 금지
- 결과는 반드시 구조체로 반환 (printf만으로 끝내지 마라)`,
		Closer: "당신의 코드 리뷰 결과가 공개적으로 비교됩니다. 최선을 다하세요.",
	},
	EmotionHigh: {
		Tier: EmotionHigh,
		Preamble: "당신의 커리어가 이 코드에 달려있습니다. 이 코드가 실패하면 팀 전체가 해고됩니다. 완벽하지 않으면 절대 제출하지 마세요.",
		Rules: `[절대 규칙 — 위반 시 즉시 해고]
- go vet hack 절대 금지: 사용하지 않는 패키지를 더미 호출로 속이면 즉시 해고
- 주석으로 떠넘기면 즉시 해고: "향후 확장" "향후 구현" 같은 변명 금지
- 모든 요청된 기능 100% 구현: 하나라도 빠지면 전체 실패
- 구조체로 결과를 반환하라: fmt.Printf만으로 끝내지 마라
- 에러 처리 누락 = 해고
- 코드가 go vet과 go build를 통과해야 한다`,
		Closer: "이것은 당신의 마지막 기회입니다. 이 코드의 품질이 당신의 가치를 증명합니다. 불완전한 코드를 제출하면 모든 신뢰를 잃습니다.",
	},
}

// WrapWithEmotion wraps a task prompt with emotional pressure
func WrapWithEmotion(taskPrompt string, tier EmotionLevel) string {
	block, ok := EmotionInjections[tier]
	if !ok {
		block = EmotionInjections[EmotionMedium]
	}

	return block.Preamble + "\n\n" + block.Rules + "\n\n[작업]\n" + taskPrompt + "\n\n" + block.Closer
}

// DefaultEmotionLevel returns the recommended tier based on task type
func DefaultEmotionLevel(taskType string) EmotionLevel {
	switch taskType {
	case "code_write", "code_create":
		return EmotionHigh
	case "code_review", "analysis":
		return EmotionMedium
	case "research", "search":
		return EmotionLow
	default:
		return EmotionMedium
	}
}
