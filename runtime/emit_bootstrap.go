// emit_bootstrap.go — Tier 1 컨텐츠 생성
//
// PROVIDES: emitBootstrap, emitAgentInbox, extractInboxPreview, emitSessionMemory
// DEPENDS:  brain.go (SubsumptionResult, Neuron, Region)
//           emit_helpers.go (pathToSentence, splitNeuronPath, sortedActiveNeurons)
//
// emitBootstrap: SubsumptionResult → GEMINI.md 문자열
//   ├→ emitAgentInbox (에이전트 수신함 섹션)
//   ├→ extractInboxPreview (파일 preview 추출)
//   └→ emitSessionMemory (세션 메모리 섹션)

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// TIER 1: GEMINI.md Bootstrap (~500 tokens)
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

func emitBootstrap(result SubsumptionResult, brainRoot string) string {
	var sb strings.Builder

	if !buildPreamble(&sb, result, brainRoot) {
		return sb.String()
	}

	formatPersona(&sb, result)
	formatSubsumption(&sb)

	// ━━━ 3-TIER RULES (AgentIF 벤치마크 기반 개선) ━━━
	formatTieredRules(&sb, result)

	formatCortexBans(&sb, result)
	formatGrowthAndLimbic(&sb, result, brainRoot)
	formatRecentMemory(&sb, result)

	// ━━━ AGENT INBOX ━━━
	agentInbox := emitAgentInbox(brainRoot)
	if agentInbox != "" {
		sb.WriteString(agentInbox)
	}

	// ━━━ SESSION TRANSCRIPT LOCATION (절대경로 강제) ━━━
	transcriptDir := filepath.Join(brainRoot, "_transcripts")
	if _, err := os.Stat(transcriptDir); err == nil {
		absTranscript, _ := filepath.Abs(transcriptDir)
		sb.WriteString(fmt.Sprintf("### 📜 전사 기록\n전사물 경로: `%s`\n\n", absTranscript))
	}

	sb.WriteString("<!-- NEURONFS:END -->\n")
	return sb.String()
}

// ─── HELPER FUNCTIONS ───

// formatCodeMapAndSoul is now merged into formatGrowthAndLimbic via section_growth_soul.tmpl

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// 3-TIER RULES: ALWAYS / WHEN / NEVER
// AgentIF 벤치마크 결과 기반 — 조건부 규칙에 트리거 조건 명시
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// whenConditions: 必 규칙 중 조건부 트리거가 필요한 키워드 → 트리거 조건 매핑
// 이 맵에 없는 必 규칙은 ALWAYS로 분류된다
var whenConditions = map[string]string{
	"qorz":   "코딩/기술 결정 시",
	"커뮤니티검색": "코딩/기술 결정 시",
	"코드맵":    "코드 수정 시",
	"적층해결":   "문제/에러 발생 시",
	"bat재시작": "시스템 재시작 시",
	"go vet": "Go 코드 수정 후",
	"govet":  "Go 코드 수정 후",
	// 推 규칙용 조건
	"grep_search": "코드 검색 시",
	"로컬깃":         "코드 변경 후",
	"로컬테스트":       "코드 변경 후",
	"테스트자동화":      "코드 변경 후",
	"리서치":         "코딩/기술 결정 시",
	"리포관리":        "배포/릴리스 시",
	"브랜치":         "배포/릴리스 시",
	"README":      "문서 변경 시",
	"깃헙":          "배포/릴리스 시",
	"디스코드":        "배포/릴리스 시",
	"버전관리":        "배포/릴리스 시",
	"검색":          "코딩/기술 결정 시",
	"영향범위":        "코드 수정 전",
	"누락대조":        "코드 생성/갱신 시",
	"기능누락":        "코드 생성/갱신 시",
}

// ruleWhyHow: 알려진 규칙의 WHY(이유) + HOW(방법) 매핑
// Description이 있으면 Description 우선, 없으면 이 맵에서 조회
type whyHow struct {
	Why string
	How string
}

var ruleWhyHow = map[string]whyHow{
	// WHEN rules
	"qorz":   {How: "커뮤니티검색필수"},
	"커뮤니티검색": {How: "커뮤니티검색필수"},
	"코드맵":    {How: "코드맵 참조생성갱신"},
	"적층해결":   {How: "기존 코드 교체 말고 위에 적층"},
	"bat재시작": {How: "bat재시작"},

	// NEVER rules
	"중복작업":   {Why: "토큰 낭비 + 기존 결과 덮어쓰기 위험"},
	"땜질코딩":   {Why: "근본 원인 미해결 → 반복 에러"},
	"무계획":    {Why: "영향 범위 미파악 → 연쇄 장애"},
	"하드코딩":   {Why: "환경 변경 시 즉시 장애"},
	"curl사용": {Why: "PowerShell 환경. curl=Invoke-WebRequest 별칭 충돌"},
	"무한대기":   {Why: "프로세스 행 → 사용자 답답함"},
	"무한작업":   {Why: "범위 없는 작업 → 토큰 고갈"},
	"불필요한코드": {Why: "유지보수 부담 증가"},
}

// lookupWhyHow: Description 우선 → ruleWhyHow 맵 fallback

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// AGENT INBOX: 에이전트 간 소통 (인젝션 기반)
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// extractInboxPreview는 inbox 파일에서 발신자와 제목을 추출한다.

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// SESSION MEMORY: 재시작 시 직전 대화 기억 복원
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
