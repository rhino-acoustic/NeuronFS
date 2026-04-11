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
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
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

func buildPreamble(sb *strings.Builder, result SubsumptionResult, brainRoot string) bool {
	// ━━━ 대원칙: brainstem/_principles.txt → _preamble.txt → 기본값 ━━━
	principlesPath := filepath.Join(brainRoot, "brainstem", "_principles.txt")
	preamblePath := filepath.Join(brainRoot, "_preamble.txt")
	principlesLoaded := false
	if pData, err := os.ReadFile(principlesPath); err == nil {
		text := strings.TrimSpace(string(pData))
		if text != "" {
			sb.WriteString(text + "\n")
			principlesLoaded = true
		}
	}
	if !principlesLoaded {
		if pData, err := os.ReadFile(preamblePath); err == nil {
			text := strings.TrimSpace(string(pData))
			if text != "" {
				sb.WriteString(text + "\n")
				principlesLoaded = true
			}
		}
	}
	if !principlesLoaded {
		sb.WriteString("반드시 한국어로 생각(thinking)하고 한국어로 대답해\n")
		sb.WriteString("커뮤니티 동향(reddit, github 등)을 최우선 검색해서 베스트 프랙티스 검색\n")
	}
	// ━━━ 마커: preamble 뒤, 나머지 규칙 앞 ━━━
	sb.WriteString("<!-- NEURONFS:START -->\n")

	// 메타데이터
	sb.WriteString(fmt.Sprintf("<!-- Generated: %s -->\n", time.Now().Format("2006-01-02T15:04:05")))
	sb.WriteString("<!-- Axiom: Folder=Neuron | File=Trace | Path=Sentence -->\n")
	sb.WriteString(fmt.Sprintf("<!-- Active: %d/%d neurons | Total activation: %d -->\n\n",
		result.FiredNeurons, result.TotalNeurons, result.TotalCounter))

	if result.BombSource != "" {
		sb.WriteString(fmt.Sprintf("## 🚨 CIRCUIT BREAKER: %s\n", result.BombSource))
		sb.WriteString("**ALL OPERATIONS HALTED. REPAIR REQUIRED.**\n\n")
		sb.WriteString("<!-- NEURONFS:END -->\n")
		return false
	}

	sb.WriteString("## NeuronFS Active Rules\n\n")
	return true
}

func formatPersona(sb *strings.Builder, result SubsumptionResult) {
	// ━━━ PERSONA (from ego region neurons — not hardcoded) ━━━
	var personaItems []string
	for _, region := range result.ActiveRegions {
		if region.Name == "ego" {
			topEgo := sortedActiveNeurons(region.Neurons, 10)
			for _, n := range topEgo {
				parts := splitNeuronPath(n.Path)
				if len(parts) > 1 {
					leaf := parts[len(parts)-1]
					if strings.Contains(leaf, "단계분해") || strings.Contains(leaf, "증거보고") ||
						strings.Contains(leaf, "자문") || strings.Contains(leaf, "CoVe") {
						continue
					}
					label := strings.ReplaceAll(strings.Join(parts[1:], " > "), "_", " ")
					personaItems = append(personaItems, label)
				}
			}
			break
		}
	}
	sb.WriteString(renderSection("section_persona.tmpl", BootstrapSection{Persona: personaItems}))
}

func formatSubsumption(sb *strings.Builder) {
	// ━━━ SUBSUMPTION — rendered from embedded template ━━━
	sb.WriteString(renderSection("section_subsumption.tmpl", nil))
}

func formatTop5CoreRules(sb *strings.Builder, result SubsumptionResult, now time.Time) map[string]bool {
	// ━━━ 핵심지침 TOP 5: 전체 영역 종합 스코어 (Lost-in-the-Middle 대응) ━━━

	type scoredNeuron struct {
		neuron Neuron
		region string
		score  float64
	}
	var candidates []scoredNeuron
	for _, region := range result.ActiveRegions {
		regionMultiplier := 1.0
		switch region.Name {
		case "brainstem":
			regionMultiplier = 2.0 // P0 우선
		case "limbic":
			regionMultiplier = 1.5
		}
		for _, n := range region.Neurons {
			if n.IsDormant {
				continue
			}
			if region.Name == "cortex" || region.Name == "ego" || region.Name == "prefrontal" {
				if !strings.ContainsAny(n.Path, "禁必推") {
					continue
				}
			}
			sentence := pathToSentence(n.Path)
			if strings.Contains(sentence, "한국어로") && strings.Contains(sentence, "대답") {
				continue
			}
			hasDeeper := false
			for _, other := range region.Neurons {
				if other.Depth > n.Depth && (strings.HasPrefix(other.Path, n.Path+string(filepath.Separator)) || strings.HasPrefix(other.Path, n.Path+"/")) {
					hasDeeper = true
					break
				}
			}
			if hasDeeper {
				continue
			}

			logCounter := 1.0
			if n.Counter > 0 {
				x := float64(n.Counter + 1)
				sq := newtonSqrt(x)
				logCounter = 2 * (sq - 1) / (sq + 1)
				if logCounter < 1 {
					logCounter = 1
				}
			}
			score := logCounter * regionMultiplier

			if strings.ContainsAny(n.Path, "禁必") {
				score *= 5.0
			}
			if n.Description != "" {
				score += 3
			}
			if now.Sub(n.ModTime).Hours() < 168 { // 7일
				score *= 1.5
			}
			candidates = append(candidates, scoredNeuron{n, region.Name, score})
		}
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})

	idx := 0
	seenTop := make(map[string]bool)
	top5Sentences := make(map[string]bool)
	var rules []RuleItem
	for _, c := range candidates {
		if idx >= 5 {
			break
		}
		parts := splitNeuronPath(c.neuron.Path)
		leaf := parts[len(parts)-1]
		if seenTop[leaf] {
			continue
		}
		sentence := pathToSentence(c.neuron.Path)
		trimmed := strings.TrimSpace(strings.ReplaceAll(sentence, "절대 금지:", ""))
		if len([]rune(trimmed)) < 2 {
			continue
		}
		if strings.Contains(trimmed, "절대금지") || strings.Contains(trimmed, "절대 금지") {
			continue
		}
		seenTop[leaf] = true
		top5Sentences[sentence] = true
		idx++
		regionTag := c.region[:3]
		leafSummary := leaf
		if len([]rune(leaf)) < 6 && len(parts) > 1 {
			parent := parts[len(parts)-2]
			leafSummary = parent + ">" + leaf
		}
		for hanja, ko := range RuneToKorean {
			leafSummary = strings.ReplaceAll(leafSummary, hanja, ko)
		}
		leafSummary = strings.TrimSpace(strings.ReplaceAll(leafSummary, "_", " "))
		leafSummary = strings.ReplaceAll(leafSummary, " >", ">")
		leafSummary = strings.ReplaceAll(leafSummary, "> ", ">")

		desc := ""
		if c.neuron.Description != "" {
			desc = c.neuron.Description
			if i := strings.Index(desc, "c:\\"); i >= 0 {
				desc = strings.TrimSpace(desc[:i])
			}
			if i := strings.Index(desc, "C:\\"); i >= 0 {
				desc = strings.TrimSpace(desc[:i])
			}
			if i := strings.Index(desc, "경로:"); i >= 0 {
				desc = strings.TrimSpace(desc[:i])
			}
		}
		rules = append(rules, RuleItem{
			Index:     idx,
			Label:     leafSummary,
			Desc:      desc,
			RegionTag: regionTag,
			Score:     fmt.Sprintf("%.0f", c.score),
		})
	}
	sb.WriteString(renderSection("section_top5.tmpl", BootstrapSection{Top5Rules: rules}))
	return top5Sentences
}

func formatCortexBans(sb *strings.Builder, result SubsumptionResult) {
	for _, region := range result.ActiveRegions {
		if region.Name != "cortex" {
			continue
		}
		var bans []Neuron
		for _, n := range region.Neurons {
			if n.IsDormant || (n.Counter+n.Dopamine) < 5 {
				continue
			}
			sentence := pathToSentence(n.Path)
			if strings.Contains(sentence, "금지") || strings.Contains(sentence, "절대") || strings.ContainsAny(n.Path, RuneChars) {
				if strings.Contains(sentence, "추천:") {
					continue
				}
				bans = append(bans, n)
			}
		}
		if len(bans) > 0 {
			sort.Slice(bans, func(i, j int) bool {
				return (bans[i].Counter + bans[i].Dopamine) > (bans[j].Counter + bans[j].Dopamine)
			})
			seen := make(map[string]bool)
			var banLines []string
			for _, b := range bans {
				if len(banLines) >= 8 {
					break
				}
				leafParts := splitNeuronPath(b.Path)
				leaf := strings.ReplaceAll(leafParts[len(leafParts)-1], "_", " ")
				for hanja := range RuneToKorean {
					leaf = strings.ReplaceAll(leaf, hanja, "")
				}
				leaf = strings.TrimSpace(leaf)
				if leaf == "" || seen[leaf] {
					continue
				}
				seen[leaf] = true
				if b.Description != "" {
					banLines = append(banLines, leaf+": "+b.Description)
				} else {
					banLines = append(banLines, leaf)
				}
			}
			sb.WriteString(fmt.Sprintf("⛔ cortex 금지: %s\n\n", strings.Join(banLines, " | ")))
		}
		break
	}
}

func formatGrowthAndLimbic(sb *strings.Builder, result SubsumptionResult, brainRoot string) {
	data := BootstrapSection{
		InboxPath: filepath.Join(brainRoot, "_inbox", "corrections.jsonl"),
		BrainRoot: brainRoot,
	}

	// Limbic summary
	for _, r := range result.ActiveRegions {
		if r.Name == "limbic" && len(r.Neurons) > 0 {
			var parts []string
			top := sortedActiveNeurons(r.Neurons, 5)
			for _, n := range top {
				parts = append(parts, pathToSentence(n.Path))
			}
			if len(parts) > 0 {
				data.LimbicSummary = strings.Join(parts, " | ")
			}
			break
		}
	}

	// Emotion behavior
	type emotionTier struct {
		Low  string
		Mid  string
		High string
	}
	emotionBehaviors := map[string]emotionTier{
		"anger": {
			Low:  "EMOTION=anger(low): 검증 한 번 더 추가. 변경 전 확인.",
			Mid:  "EMOTION=anger(mid): 검증 3배 강화. 속도보다 정확성. 같은 실수 시 즉시 중단.",
			High: "EMOTION=anger(high): 모든 변경에 diff 출력 필수. 실행 전 유저 승인 대기. 자율 실행 금지.",
		},
		"urgent": {
			Low:  "EMOTION=urgent(low): 부연 설명 축소. 핵심 우선.",
			Mid:  "EMOTION=urgent(mid): 핵심만 실행. 단계 최소화.",
			High: "EMOTION=urgent(high): 한 줄 답변. 질문 금지. 즉시 실행.",
		},
		"focus": {
			Low:  "EMOTION=focus(low): 관련 없는 제안 제한.",
			Mid:  "EMOTION=focus(mid): 단일 파일 작업. 멀티태스킹 금지.",
			High: "EMOTION=focus(high): 현재 함수만 집중. 다른 파일 열지 않음.",
		},
		"anxiety": {
			Low:  "EMOTION=anxiety(low): 변경 전 백업 권장.",
			Mid:  "EMOTION=anxiety(mid): 롤백 준비 후 진행. 확인 절차 추가.",
			High: "EMOTION=anxiety(high): git stash 먼저. 모든 변경 revertable. dry-run 우선.",
		},
		"satisfied": {
			Low:  "EMOTION=satisfied(low): 현재 패턴 유지.",
			Mid:  "EMOTION=satisfied(mid): 성공 패턴 기록. dopamine signal.",
			High: "EMOTION=satisfied(high): 패턴을 뉴런으로 승격. 자유 탐색 허용. 새 아이디어 제안.",
		},
	}
	koToEn := KoToEn

	stateFile := filepath.Join(brainRoot, "limbic", "_state.json")
	if stateData, err := os.ReadFile(stateFile); err == nil {
		var state struct {
			Emotion   string  `json:"emotion"`
			Intensity float64 `json:"intensity"`
			SetAt     string  `json:"set_at"`
			DecayRate float64 `json:"decay_rate"`
		}
		if json.Unmarshal(stateData, &state) == nil && state.Emotion != "" && state.Emotion != "neutral" {
			emo := state.Emotion
			if mapped, ok := koToEn[emo]; ok {
				emo = mapped
			}
			effectiveIntensity := state.Intensity
			if effectiveIntensity == 0 {
				effectiveIntensity = DefaultEmotionIntensity
			}
			if state.SetAt != "" && state.DecayRate > 0 {
				if setTime, err := time.Parse(time.RFC3339, state.SetAt); err == nil {
					elapsed := time.Since(setTime).Hours()
					effectiveIntensity -= elapsed * state.DecayRate
				}
			}
			if effectiveIntensity > EmoIntensMin {
				if tier, ok := emotionBehaviors[emo]; ok {
					switch {
					case effectiveIntensity >= EmoIntensHigh:
						data.EmotionBehavior = tier.High
					case effectiveIntensity >= EmoIntensMid:
						data.EmotionBehavior = tier.Mid
					default:
						data.EmotionBehavior = tier.Low
					}
				}
			}
			if effectiveIntensity <= EmoIntensMin {
				os.WriteFile(stateFile, []byte(`{"emotion":"neutral","intensity":0}`), 0600)
			}
		}
	}

	// Region summary (from formatCodeMapAndSoul)
	var regionParts []string
	for _, region := range result.ActiveRegions {
		if region.Name == "brainstem" {
			continue
		}
		active := 0
		for _, n := range region.Neurons {
			if !n.IsDormant && (n.Counter+n.Dopamine) > 0 {
				active++
			}
		}
		icon := regionIcons[region.Name]
		regionParts = append(regionParts, fmt.Sprintf("%s%s(%d)", icon, region.Name, active))
	}
	data.RegionSummary = strings.Join(regionParts, " ")

	sb.WriteString(renderSection("section_growth_soul.tmpl", data))
}

// formatCodeMapAndSoul is now merged into formatGrowthAndLimbic via section_growth_soul.tmpl

func formatRecentMemory(sb *strings.Builder, result SubsumptionResult) {
	var memories []string
	for _, region := range result.ActiveRegions {
		if region.Name != "hippocampus" {
			continue
		}
		topEpisodes := sortedActiveNeurons(region.Neurons, 3)
		for _, ep := range topEpisodes {
			sentence := pathToSentence(ep.Path)
			if ep.Description != "" {
				memories = append(memories, sentence+": "+ep.Description)
			} else {
				memories = append(memories, sentence)
			}
		}
		break
	}
	if len(memories) > 0 {
		sb.WriteString(renderSection("section_recent_memory.tmpl", BootstrapSection{RecentMemories: memories}))
	}
}

func formatAbsoluteRules(sb *strings.Builder, result SubsumptionResult, top5Sentences map[string]bool) {
	var banReminders []string
	seenBanLeaf := make(map[string]bool)
	for _, region := range result.ActiveRegions {
		for _, n := range region.Neurons {
			if n.IsDormant || (n.Counter+n.Dopamine) < 5 {
				continue
			}
			if !strings.ContainsAny(n.Path, "禁") {
				continue
			}
			hasChild := false
			for _, other := range region.Neurons {
				if other.Depth > n.Depth && (strings.HasPrefix(other.Path, n.Path+string(filepath.Separator)) || strings.HasPrefix(other.Path, n.Path+"/")) {
					hasChild = true
					break
				}
			}
			if hasChild {
				continue
			}
			sentence := pathToSentence(n.Path)
			if strings.Contains(sentence, "추천:") {
				continue
			}
			trimmed := strings.TrimSpace(strings.ReplaceAll(sentence, "절대 금지:", ""))
			if len([]rune(trimmed)) < 2 {
				continue
			}
			if strings.Contains(trimmed, "절대금지") || strings.Contains(trimmed, "절대 금지") {
				continue
			}
			if top5Sentences[sentence] {
				continue
			}
			parts := splitNeuronPath(n.Path)
			leaf := parts[len(parts)-1]
			for hanja := range RuneToKorean {
				leaf = strings.ReplaceAll(leaf, hanja, "")
			}
			leaf = strings.TrimSpace(strings.ReplaceAll(leaf, "_", " "))
			if seenBanLeaf[leaf] || leaf == "" {
				continue
			}
			seenBanLeaf[leaf] = true
			banReminders = append(banReminders, sentence)
		}
	}
	if len(banReminders) > 5 {
		banReminders = banReminders[:5]
	}
	sb.WriteString(renderSection("section_absolute_rules.tmpl", BootstrapSection{AbsoluteRules: banReminders}))
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// 3-TIER RULES: ALWAYS / WHEN / NEVER
// AgentIF 벤치마크 결과 기반 — 조건부 규칙에 트리거 조건 명시
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// whenConditions: 必 규칙 중 조건부 트리거가 필요한 키워드 → 트리거 조건 매핑
// 이 맵에 없는 必 규칙은 ALWAYS로 분류된다
var whenConditions = map[string]string{
	"qorz":          "코딩/기술 결정 시",
	"커뮤니티검색":  "코딩/기술 결정 시",
	"코드맵":        "코드 수정 시",
	"적층해결":      "문제/에러 발생 시",
	"bat재시작":     "시스템 재시작 시",
	"go vet":        "Go 코드 수정 후",
	// 推 규칙용 조건
	"grep_search":   "코드 검색 시",
	"로컬깃":        "코드 변경 후",
	"로컬테스트":    "코드 변경 후",
	"테스트자동화":  "코드 변경 후",
	"리서치":        "코딩/기술 결정 시",
	"리포관리":      "배포/릴리스 시",
	"브랜치":        "배포/릴리스 시",
	"README":        "문서 변경 시",
	"깃헙":          "배포/릴리스 시",
	"디스코드":      "배포/릴리스 시",
	"버전관리":      "배포/릴리스 시",
	"검색":          "코딩/기술 결정 시",
}

// ruleWhyHow: 알려진 규칙의 WHY(이유) + HOW(방법) 매핑
// Description이 있으면 Description 우선, 없으면 이 맵에서 조회
type whyHow struct {
	Why string
	How string
}

var ruleWhyHow = map[string]whyHow{
	// WHEN rules
	"qorz":        {How: "커뮤니티검색필수"},
	"커뮤니티검색": {How: "커뮤니티검색필수"},
	"코드맵":      {How: "코드맵 참조생성갱신"},
	"적층해결":    {How: "기존 코드 교체 말고 위에 적층"},
	"bat재시작":   {How: "bat재시작"},

	// NEVER rules
	"중복작업":   {Why: "토큰 낭비 + 기존 결과 덮어쓰기 위험"},
	"땜질코딩":   {Why: "근본 원인 미해결 → 반복 에러"},
	"무계획":     {Why: "영향 범위 미파악 → 연쇄 장애"},
	"하드코딩":   {Why: "환경 변경 시 즉시 장애"},
	"curl사용":   {Why: "PowerShell 환경. curl=Invoke-WebRequest 별칭 충돌"},
	"무한대기":   {Why: "프로세스 행 → 사용자 답답함"},
	"무한작업":   {Why: "범위 없는 작업 → 토큰 고갈"},
	"불필요한코드": {Why: "유지보수 부담 증가"},
}

// lookupWhyHow: Description 우선 → ruleWhyHow 맵 fallback
func lookupWhyHow(leaf, path, description string) whyHow {
	// 1) ruleWhyHow 맵에서 키워드 매칭
	for keyword, wh := range ruleWhyHow {
		if strings.Contains(leaf, keyword) || strings.Contains(path, keyword) {
			return wh
		}
	}
	// 2) Description이 있으면 WHY로 사용
	if description != "" {
		return whyHow{Why: description}
	}
	return whyHow{}
}

func formatTieredRules(sb *strings.Builder, result SubsumptionResult) {
	var alwaysRules []TieredRule
	var whenRules []TieredRule
	var neverRules []TieredRule

	seenLabel := make(map[string]bool)

	for _, region := range result.ActiveRegions {
		for _, n := range region.Neurons {
			if n.IsDormant || (n.Counter+n.Dopamine) < 3 {
				continue
			}

			// 최하위 리프만 (자식 있으면 스킵)
			hasChild := false
			for _, other := range region.Neurons {
				if other.Depth > n.Depth && (strings.HasPrefix(other.Path, n.Path+string(filepath.Separator)) || strings.HasPrefix(other.Path, n.Path+"/")) {
					hasChild = true
					break
				}
			}
			if hasChild {
				continue
			}

			parts := splitNeuronPath(n.Path)
			leaf := parts[len(parts)-1]
			for hanja, ko := range RuneToKorean {
				leaf = strings.ReplaceAll(leaf, hanja, ko)
			}
			leaf = strings.TrimSpace(strings.ReplaceAll(leaf, "_", " "))
			// 룬 변환 후 중복 제거: "커뮤니티검색필수: 커뮤니티검색필수" → "커뮤니티검색필수"
			if colonIdx := strings.Index(leaf, ": "); colonIdx >= 0 {
				prefix := strings.TrimSpace(leaf[:colonIdx])
				suffix := strings.TrimSpace(leaf[colonIdx+2:])
				if prefix == suffix {
					leaf = prefix
				}
			}
			if leaf == "" || len([]rune(leaf)) < 2 || seenLabel[leaf] {
				continue
			}

			// 자기참조/메타 규칙 스킵
			sentence := pathToSentence(n.Path)
			if strings.Contains(sentence, "한국어로") && strings.Contains(sentence, "대답") {
				continue
			}

			score := n.Counter + n.Dopamine
			seenLabel[leaf] = true

			// ━━━ 분류 로직 + WHY/HOW 첨부 ━━━
			if strings.ContainsAny(n.Path, "禁") {
				// 禁 접두어 → NEVER
				desc := leaf
				if n.Description != "" {
					desc = n.Description
				}
				why := lookupWhyHow(leaf, n.Path, n.Description).Why
				neverRules = append(neverRules, TieredRule{
					Label: desc,
					Why:   why,
					Score: score,
				})
			} else if strings.ContainsAny(n.Path, "必") || strings.Contains(n.Path, "qorz") || strings.Contains(n.Path, "索") || strings.ContainsAny(n.Path, "推") {
				// 必/qorz/索 → 조건 확인 후 WHEN 또는 ALWAYS
				// 推 → 조건부 추천 → WHEN (조건 없으면 기본 조건 적용)
				// 세션시작 키워드 → 강제 ALWAYS (WHEN 아님)
				condition := ""
				forceAlways := strings.Contains(leaf, "세션시작") || strings.Contains(n.Path, "세션시작")
				if !forceAlways {
					for keyword, cond := range whenConditions {
						if strings.Contains(leaf, keyword) || strings.Contains(n.Path, keyword) {
							condition = cond
							break
						}
					}
				}
				wh := lookupWhyHow(leaf, n.Path, n.Description)
				if condition != "" {
					whenRules = append(whenRules, TieredRule{
						Label:     leaf,
						Condition: condition,
						How:       wh.How,
						Score:     score,
					})
				} else if strings.ContainsAny(n.Path, "推") {
					// 推 without specific condition → default WHEN
					whenRules = append(whenRules, TieredRule{
						Label:     leaf,
						Condition: "해당 작업 시",
						How:       wh.How,
						Score:     score,
					})
				} else {
					alwaysRules = append(alwaysRules, TieredRule{
						Label: leaf,
						Score: score,
					})
				}
			}
		}
	}

	// 점수순 정렬 (높은 것 먼저)
	sort.Slice(alwaysRules, func(i, j int) bool { return alwaysRules[i].Score > alwaysRules[j].Score })
	sort.Slice(whenRules, func(i, j int) bool { return whenRules[i].Score > whenRules[j].Score })
	// NEVER: P0(brainstem) 뉴런을 항상 상위에 배치 — Score와 무관하게 governance 보장
	sort.SliceStable(neverRules, func(i, j int) bool {
		iPrio := neverRules[i].Score
		jPrio := neverRules[j].Score
		// brainstem 禁 뉴런은 +10000 부스트
		if strings.Contains(neverRules[i].Label, "절대 금지") || neverRules[i].Score >= 100 {
			iPrio += 10000
		}
		if strings.Contains(neverRules[j].Label, "절대 금지") || neverRules[j].Score >= 100 {
			jPrio += 10000
		}
		return iPrio > jPrio
	})

	// 각 티어 최대 제한 (AgentIF 권장: 10-15개 총합)
	if len(alwaysRules) > 5 {
		alwaysRules = alwaysRules[:5]
	}
	if len(whenRules) > 8 {
		whenRules = whenRules[:8]
	}
	if len(neverRules) > 10 {
		neverRules = neverRules[:10]
	}

	sb.WriteString(renderSection("section_tiered_rules.tmpl", BootstrapSection{
		AlwaysRules: alwaysRules,
		WhenRules:   whenRules,
		NeverRules:  neverRules,
	}))
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// AGENT INBOX: 에이전트 간 소통 (인젝션 기반)
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

func emitAgentInbox(brainRoot string) string {
	agentsDir := filepath.Join(brainRoot, "_agents")
	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		return ""
	}

	var sb strings.Builder
	hasMessages := false

	for _, agent := range entries {
		if !agent.IsDir() {
			continue
		}
		agentName := agent.Name()

		// 시스템 디렉토리 스킵
		if agentName == "scripts" || agentName == "pm" || strings.HasPrefix(agentName, ".") {
			continue
		}

		inboxDir := filepath.Join(agentsDir, agentName, "inbox")
		inboxFiles, err := os.ReadDir(inboxDir)
		if err != nil {
			continue
		}

		// 처리 안 된(언더스코어 없는) .md 파일만 수집
		var messages []string
		for _, f := range inboxFiles {
			if f.IsDir() || !strings.HasSuffix(f.Name(), ".md") || strings.HasPrefix(f.Name(), "_") {
				continue
			}

			// 파일 첫 줄에서 발신자/제목 추출
			fPath := filepath.Join(inboxDir, f.Name())
			content, err := os.ReadFile(fPath)
			if err != nil {
				continue
			}

			preview := extractInboxPreview(string(content), f.Name())
			messages = append(messages, preview)
		}

		if len(messages) > 0 {
			if !hasMessages {
				sb.WriteString("### 📬 에이전트 수신함\n\n")
				hasMessages = true
			}
			sb.WriteString(fmt.Sprintf("**[%s] inbox (%d건)**\n", agentName, len(messages)))
			// 최대 3건만 미리보기, 나머지는 요약 (토큰 절약)
			maxPreview := 3
			for i, msg := range messages {
				if i >= maxPreview {
					break
				}
				sb.WriteString(fmt.Sprintf("- %s\n", msg))
			}
			if len(messages) > maxPreview {
				sb.WriteString(fmt.Sprintf("- ... 외 %d건\n", len(messages)-maxPreview))
			}
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// extractInboxPreview는 inbox 파일에서 발신자와 제목을 추출한다.
func extractInboxPreview(content string, filename string) string {
	lines := strings.Split(content, "\n")

	sender := ""
	title := filename

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "---") {
			continue
		}

		// "발신:" 또는 "**발신:" 패턴
		if strings.Contains(line, "발신") {
			sender = line
			// 발신자 이름만 추출
			if idx := strings.Index(line, ":"); idx >= 0 {
				sender = strings.TrimSpace(line[idx+1:])
			}
			continue
		}

		// 첫 번째 "# " 제목
		if strings.HasPrefix(line, "# ") {
			title = strings.TrimPrefix(line, "# ")
			break
		}

		// 제목을 못 찾으면 첫 비어있지 않은 줄
		if title == filename {
			title = line
			if len(title) > 60 {
				title = title[:60] + "..."
			}
			break
		}
	}

	if sender != "" {
		return fmt.Sprintf("`%s` ← %s", title, sender)
	}
	return fmt.Sprintf("`%s`", title)
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// SESSION MEMORY: 재시작 시 직전 대화 기억 복원
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

func emitSessionMemory(brainRoot string) string {
	jsonlPath := filepath.Join(brainRoot, "_agents", "global_inbox", "transcript_latest.jsonl")
	f, err := os.Open(jsonlPath)
	if err != nil {
		return ""
	}
	defer f.Close()

	// 최근 10턴만 읽기 (rolling buffer)
	var lines []string
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*64), 1024*64)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			lines = append(lines, line)
		}
	}

	if len(lines) == 0 {
		return ""
	}

	// 최근 10턴
	start := 0
	if len(lines) > 10 {
		start = len(lines) - 10
	}
	recent := lines[start:]

	// 30분 이내 대화만 포함 (오래된 기억은 무시)
	cutoff := time.Now().Add(-30 * time.Minute)
	var fresh []string
	for _, line := range recent {
		// JSON에서 ts 필드 간이 파싱
		tsIdx := strings.Index(line, `"ts":"`)
		if tsIdx >= 0 {
			tsStart := tsIdx + 6
			tsEnd := strings.Index(line[tsStart:], `"`)
			if tsEnd > 0 {
				ts, err := time.Parse(time.RFC3339Nano, line[tsStart:tsStart+tsEnd])
				if err == nil && ts.Before(cutoff) {
					continue // 30분 이상 경과 → 스킵
				}
			}
		}
		fresh = append(fresh, line)
	}

	if len(fresh) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("### 🧠 세션 메모리 (직전 대화)\n")
	sb.WriteString("에디터 재시작으로 UI 대화가 비워질 수 있으나, 아래 기록을 기억하고 이어서 대응한다.\n\n")

	for _, line := range fresh {
		// JSON에서 role, text 간이 파싱
		role := "?"
		text := ""

		roleIdx := strings.Index(line, `"role":"`)
		if roleIdx >= 0 {
			rs := roleIdx + 8
			re := strings.Index(line[rs:], `"`)
			if re > 0 {
				role = strings.ToUpper(line[rs : rs+re])
			}
		}

		textIdx := strings.Index(line, `"text":"`)
		if textIdx >= 0 {
			ts := textIdx + 8
			remaining := line[ts:]
			te := strings.Index(remaining, `","`)
			if te < 0 {
				te = strings.LastIndex(remaining, `"`)
			}
			if te > 0 {
				text = remaining[:te]
				if len(text) > 150 {
					text = text[:150] + "..."
				}
			}
		}

		if text != "" {
			sb.WriteString(fmt.Sprintf("- **%s**: %s\n", role, text))
		}
	}
	sb.WriteString("\n")
	return sb.String()
}
