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
