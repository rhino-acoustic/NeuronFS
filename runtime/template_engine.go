package main

import (
	"bytes"
	"embed"
	"text/template"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

// BootstrapSection holds the data for a single section of the bootstrap output.
// Each format* function in emit_bootstrap.go will gradually migrate to use these.
type BootstrapSection struct {
	// Persona — ego region top neurons
	Persona []string

	// RecentMemories — hippocampus session_log entries
	RecentMemories []string

	// AbsoluteRules — brainstem 禁/必 absolute ban/mandate entries
	AbsoluteRules []string

	// Growth + Soul section
	InboxPath       string
	LimbicSummary   string
	EmotionBehavior string
	RegionSummary   string
	BrainRoot       string

	// CortexBans — cortex 금지 1-liner
	CortexBans string

	// Top5Rules — 핵심지침 TOP 5 항목 (legacy, 유지)
	Top5Rules []RuleItem

	// 3-Tier Rule Structure (AgentIF 벤치마크 기반 개선)
	// ALWAYS: 매 응답마다 적용 (必 접두어 + 항상 활성)
	// WHEN:   조건부 트리거 (qorz/索 + 상황 의존)
	// NEVER:  절대 금지 (禁 접두어)
	AlwaysRules []TieredRule
	WhenRules   []TieredRule
	NeverRules  []TieredRule
}

// TieredRule represents a classified rule with trigger condition, rationale, and action.
type TieredRule struct {
	Label     string // 규칙 요약 (예: "한국어 사고/응답")
	Condition string // WHEN 규칙의 트리거 조건 (예: "코딩/기술 작업 시")
	Why       string // WHY: 이 규칙을 따라야 하는 이유
	How       string // HOW: 규칙을 따르는 구체적 방법
	Score     int    // 발화수 (counter + dopamine)
}

// RuleItem represents a single rendered rule in the TOP 5 section.
type RuleItem struct {
	Index     int
	Label     string
	Desc      string
	RegionTag string
	Score     string
}

var bootstrapTemplates *template.Template

func init() {
	bootstrapTemplates = template.Must(template.ParseFS(templateFS, "templates/*.tmpl"))
}

// renderSection executes a named template and returns the rendered string.
// If the template fails, it returns an empty string (fail-safe, never panic).
func renderSection(name string, data interface{}) string {
	var buf bytes.Buffer
	if err := bootstrapTemplates.ExecuteTemplate(&buf, name, data); err != nil {
		return "" // fail-safe: 템플릿 실패 시 빈 문자열 반환
	}
	return buf.String()
}
