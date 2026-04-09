// emit_bootstrap.go — Tier 1 컨텐츠 생성
//
// PROVIDES: emitBootstrap, emitAgentInbox, extractInboxPreview, emitSessionMemory
// DEPENDS:  brain.go (SubsumptionResult, Neuron, Region)
//           emit_helpers.go (pathToSentence, splitNeuronPath, sortedActiveNeurons, hanjaToKorean)
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

	now := time.Now()
	top5Sentences := formatTop5CoreRules(&sb, result, now)

	formatCortexBans(&sb, result)
	formatGrowthAndLimbic(&sb, result, brainRoot)
	formatCodeMapAndSoul(&sb, result, brainRoot)
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

	formatAbsoluteRules(&sb, result, top5Sentences)

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
	sb.WriteString("### 🎭 페르소나\n")
	for _, region := range result.ActiveRegions {
		if region.Name == "ego" {
			topEgo := sortedActiveNeurons(region.Neurons, 10)
			for _, n := range topEgo {
				parts := splitNeuronPath(n.Path)
				if len(parts) > 1 {
					// 행동양식 뉴런은 🔮 영혼에서 출력 → SSOT 위반 방지
					leaf := parts[len(parts)-1]
					if strings.Contains(leaf, "단계분해") || strings.Contains(leaf, "증거보고") ||
						strings.Contains(leaf, "자문") || strings.Contains(leaf, "CoVe") {
						continue
					}
					label := strings.ReplaceAll(strings.Join(parts[1:], " > "), "_", " ")
					sb.WriteString(fmt.Sprintf("- %s\n", label))
				}
			}
			break
		}
	}
	sb.WriteString("\n")
}

func formatSubsumption(sb *strings.Builder) {
	// ━━━ SUBSUMPTION (1-liner) ━━━
	sb.WriteString("### 🔗 Subsumption Cascade\n")
	sb.WriteString("```\nbrainstem ←→ limbic ←→ hippocampus ←→ sensors ←→ cortex ←→ ego ←→ prefrontal\n  (P0)         (P1)       (P2)          (P3)       (P4)     (P5)      (P6)\n```\n")
	sb.WriteString("낮은 P가 높은 P를 항상 우선. bomb은 전체 정지.\n\n")
}

func formatTop5CoreRules(sb *strings.Builder, result SubsumptionResult, now time.Time) map[string]bool {
	// ━━━ 핵심지침 TOP 5: 전체 영역 종합 스코어 (Lost-in-the-Middle 대응) ━━━
	sb.WriteString("### ⚡ 핵심지침 TOP 5\n")

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
		for hanja, ko := range hanjaToKorean {
			leafSummary = strings.ReplaceAll(leafSummary, hanja, ko)
		}
		leafSummary = strings.TrimSpace(strings.ReplaceAll(leafSummary, "_", " "))
		leafSummary = strings.ReplaceAll(leafSummary, " >", ">")
		leafSummary = strings.ReplaceAll(leafSummary, "> ", ">")
		if c.neuron.Description != "" {
			desc := c.neuron.Description
			if i := strings.Index(desc, "c:\\"); i >= 0 {
				desc = strings.TrimSpace(desc[:i])
			}
			if i := strings.Index(desc, "C:\\"); i >= 0 {
				desc = strings.TrimSpace(desc[:i])
			}
			if i := strings.Index(desc, "경로:"); i >= 0 {
				desc = strings.TrimSpace(desc[:i])
			}
			if desc != "" {
				sb.WriteString(fmt.Sprintf("%d. **%s**: %s [%s](s:%.0f)\n", idx, leafSummary, desc, regionTag, c.score))
			} else {
				sb.WriteString(fmt.Sprintf("%d. **%s** [%s](s:%.0f)\n", idx, leafSummary, regionTag, c.score))
			}
		} else {
			sb.WriteString(fmt.Sprintf("%d. **%s** [%s](s:%.0f)\n", idx, leafSummary, regionTag, c.score))
		}
	}
	sb.WriteString("\n")
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
			if strings.Contains(sentence, "금지") || strings.Contains(sentence, "절대") || strings.ContainsAny(n.Path, hanjaChars) {
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
				for hanja := range hanjaToKorean {
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
	sb.WriteString("### 🌱 자가 성장\n")
	inboxPath := filepath.Join(brainRoot, "_inbox", "corrections.jsonl")
	sb.WriteString("교정→`corrections.jsonl` 기록 | 칭찬→dopamine | 3회실패→bomb\n")
	sb.WriteString(fmt.Sprintf("경로: `%s`\n", inboxPath))
	for _, r := range result.ActiveRegions {
		if r.Name == "limbic" && len(r.Neurons) > 0 {
			var parts []string
			top := sortedActiveNeurons(r.Neurons, 5)
			for _, n := range top {
				parts = append(parts, pathToSentence(n.Path))
			}
			if len(parts) > 0 {
				sb.WriteString("Limbic: " + strings.Join(parts, " | ") + "\n")
			}
			break
		}
	}
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
					var behavior string
					switch {
					case effectiveIntensity >= EmoIntensHigh:
						behavior = tier.High
					case effectiveIntensity >= EmoIntensMid:
						behavior = tier.Mid
					default:
						behavior = tier.Low
					}
					sb.WriteString(behavior + "\n")
				}
			}
			if effectiveIntensity <= EmoIntensMin {
				os.WriteFile(stateFile, []byte(`{"emotion":"neutral","intensity":0}`), 0600)
			}
		}
	}
}

func formatCodeMapAndSoul(sb *strings.Builder, result SubsumptionResult, brainRoot string) {
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
	sb.WriteString("영역: " + strings.Join(regionParts, " ") + "\n\n")

	sb.WriteString(fmt.Sprintf("**작업 전 `%s\\{영역}\\_rules.md`를 반드시 읽는다** (cortex=코딩/NeuronFS, sensors=NAS/브랜드, prefrontal=방향)\n", brainRoot))
	sb.WriteString("⚠️ 읽지 않으면 금지 규칙 위반이 발생한다. view_file로 먼저 읽어라. MCP read_region 호출 금지(느림).\n")

	sb.WriteString("🗺️ 코드맵=뉴런 계층(cortex/dev/). 코드 수정 전 뉴런 읽기 필수. 플랫 뉴런 금지. `go vet ./...` 실행.\n\n")

	sb.WriteString("### 🔮 영혼\n")
	sb.WriteString("자문: 진짜야? 불충분? 편한길? 같은실수? 프리미엄? → 걸리면 다시\n")
	sb.WriteString("CoVe: 초안→검증질문→독립검증→수정본 | 실행후 증거보고(시뮬레이션 금지) | 복잡작업→단계분해\n\n")
}

func formatRecentMemory(sb *strings.Builder, result SubsumptionResult) {
	for _, region := range result.ActiveRegions {
		if region.Name != "hippocampus" {
			continue
		}
		topEpisodes := sortedActiveNeurons(region.Neurons, 3)
		if len(topEpisodes) > 0 {
			sb.WriteString("### 📝 최근 기억\n")
			for _, ep := range topEpisodes {
				sentence := pathToSentence(ep.Path)
				if ep.Description != "" {
					sb.WriteString(fmt.Sprintf("- %s: %s\n", sentence, ep.Description))
				} else {
					sb.WriteString(fmt.Sprintf("- %s\n", sentence))
				}
			}
			sb.WriteString("\n")
		}
		break
	}
}

func formatAbsoluteRules(sb *strings.Builder, result SubsumptionResult, top5Sentences map[string]bool) {
	sb.WriteString("### 🔒 절대 규칙 (재확인)\n")
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
			for hanja := range hanjaToKorean {
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
	for _, ban := range banReminders {
		sb.WriteString(fmt.Sprintf("- ⛔ %s\n", ban))
	}
	sb.WriteString("\n")
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
