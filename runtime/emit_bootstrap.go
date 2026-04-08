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
		return sb.String()
	}

	sb.WriteString("## NeuronFS Active Rules\n\n")

	// ━━━ PRIMACY ZONE: 핵심지침을 가장 먼저 (LLM이 가장 주의하는 위치) ━━━

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

	// ━━━ SUBSUMPTION (1-liner) ━━━
	sb.WriteString("### 🔗 Subsumption Cascade\n")
	sb.WriteString("```\nbrainstem ←→ limbic ←→ hippocampus ←→ sensors ←→ cortex ←→ ego ←→ prefrontal\n  (P0)         (P1)       (P2)          (P3)       (P4)     (P5)      (P6)\n```\n")
	sb.WriteString("낮은 P가 높은 P를 항상 우선. bomb은 전체 정지.\n\n")

	// ━━━ 핵심지침 TOP 5: 전체 영역 종합 스코어 (Lost-in-the-Middle 대응) ━━━
	sb.WriteString("### ⚡ 핵심지침 TOP 5\n")
	// NOTE: 접두사 규칙은 🎭 페르소나의 ego 뉴런(🐤접두사)에서 관리.
	// canary emoji 랜덤 변경은 ego 뉴런과 충돌하므로 제거됨.

	// 전체 영역에서 종합 스코어로 TOP 5 산출
	type scoredNeuron struct {
		neuron Neuron
		region string
		score  float64
	}
	now := time.Now()
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
			// TOP 5는 거버넌스 원칙만: cortex/ego/prefrontal은 禁/必/推만 후보
			if region.Name == "cortex" || region.Name == "ego" || region.Name == "prefrontal" {
				if !strings.ContainsAny(n.Path, "禁必推") {
					continue
				}
			}
			sentence := pathToSentence(n.Path)
			// preamble과 중복 스킵
			if strings.Contains(sentence, "한국어로") && strings.Contains(sentence, "대답") {
				continue
			}
			// 하위 폴더 존재 시 상위 스킵
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

			// === 종합 스코어 산출 (로그 스케일: counter 2000 vs 100의 차이 축소) ===
			// log(counter+1)로 큰 counter가 스코어를 지배하는 것을 방지
			logCounter := 1.0
			if n.Counter > 0 {
				// 자연로그 근사: newtonSqrt 기반 (math import 충돌 방지)
				// ln(x) ≈ 2 * (sqrt(x) - 1) / (sqrt(x) + 1) — 충분한 정밀도
				x := float64(n.Counter + 1)
				sq := newtonSqrt(x)
				logCounter = 2 * (sq - 1) / (sq + 1)
				if logCounter < 1 {
					logCounter = 1
				}
			}
			score := logCounter * regionMultiplier

			// 禁/必 거버넌스 뉴런 부스트 (×5 — 로그 스케일에서 더 영향력)
			if strings.ContainsAny(n.Path, "禁必") {
				score *= 5.0
			}

			// description 있으면 +3 (의미 전달 가능, 과도한 부스트 방지)
			if n.Description != "" {
				score += 3
			}

			// recency 부스트: 최근 7일 이내 수정 ×1.5
			if now.Sub(n.ModTime).Hours() < 168 { // 7일
				score *= 1.5
			}

			candidates = append(candidates, scoredNeuron{n, region.Name, score})
		}
	}

	// 스코어 내림차순 정렬
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})

	// 중복 제거 (leaf 이름 기준) + TOP 5 선정
	idx := 0
	seenTop := make(map[string]bool)
	top5Sentences := make(map[string]bool) // 끝 🔒에서 중복 방지용
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
		// 의미없는 짧은 뉴런 필터 ("절대 금지:" 만 있는 것 등)
		trimmed := strings.TrimSpace(strings.ReplaceAll(sentence, "절대 금지:", ""))
		if len([]rune(trimmed)) < 2 {
			continue
		}
		// 동어반복 필터: "절대 금지: 절대금지" 같은 것
		if strings.Contains(trimmed, "절대금지") || strings.Contains(trimmed, "절대 금지") {
			continue
		}
		seenTop[leaf] = true
		top5Sentences[sentence] = true
		idx++
		regionTag := c.region[:3]
		// 2-layer 표시: bold(짧은 요약) + description(상세)
		// 짧은 요약 = 전체 그림/토큰 절약, description = 의미 불확실 시 참조
		leafSummary := leaf
		// leaf가 짧으면 부모 맥락 포함 (적층해결 → 쉬프트금지>적층해결)
		if len([]rune(leaf)) < 6 && len(parts) > 1 {
			parent := parts[len(parts)-2]
			leafSummary = parent + ">" + leaf
		}
		for hanja, ko := range hanjaToKorean {
			leafSummary = strings.ReplaceAll(leafSummary, hanja, ko)
		}
		leafSummary = strings.TrimSpace(strings.ReplaceAll(leafSummary, "_", " "))
		// > 앞뒤 공백 정리 (반드시 >적층해결 → 반드시>적층해결)
		leafSummary = strings.ReplaceAll(leafSummary, " >", ">")
		leafSummary = strings.ReplaceAll(leafSummary, "> ", ">")
		if c.neuron.Description != "" {
			desc := c.neuron.Description
			if idx := strings.Index(desc, "c:\\"); idx >= 0 {
				desc = strings.TrimSpace(desc[:idx])
			}
			if idx := strings.Index(desc, "C:\\"); idx >= 0 {
				desc = strings.TrimSpace(desc[:idx])
			}
			if idx := strings.Index(desc, "경로:"); idx >= 0 {
				desc = strings.TrimSpace(desc[:idx])
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

	// ━━━ CORTEX TOP 금지: 중복 제거 + 자동 추출 ━━━
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
				// "추천:"은 금지가 아님 → 제외
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
			// 중복 제거 (leaf 이름 기준) + TOP 8
			seen := make(map[string]bool)
			var banLines []string
			for _, b := range bans {
				if len(banLines) >= 8 {
					break
				}
				leafParts := splitNeuronPath(b.Path)
				leaf := strings.ReplaceAll(leafParts[len(leafParts)-1], "_", " ")
				// 한자 제거 (헤더에서 이미 "금지"라고 명시)
				for hanja := range hanjaToKorean {
					leaf = strings.ReplaceAll(leaf, hanja, "")
				}
				leaf = strings.TrimSpace(leaf)
				if leaf == "" || seen[leaf] {
					continue
				}
				seen[leaf] = true
				// description이 있으면 풀어쓰기
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

	// ━━━ GROWTH PROTOCOL (ultra-compact) ━━━
	sb.WriteString("### 🌱 자가 성장\n")
	inboxPath := filepath.Join(brainRoot, "_inbox", "corrections.jsonl")
	sb.WriteString(fmt.Sprintf("교정→`corrections.jsonl` 기록 | 칭찬→dopamine | 3회실패→bomb\n"))
	sb.WriteString(fmt.Sprintf("경로: `%s`\n", inboxPath))
	// Limbic: 실제 뉴런 기반 동적 렌더링 (하드코딩 금지)
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
	// ── Emotion State Machine (Anthropic emotion-vector inspired) ──
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
	// Legacy Korean key fallback
	koToEn := map[string]string{"분노": "anger", "긴급": "urgent", "만족": "satisfied", "불안": "anxiety", "집중": "focus"}

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
			// Compute effective intensity with time-based decay
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
			if effectiveIntensity > 0.1 {
				if tier, ok := emotionBehaviors[emo]; ok {
					var behavior string
					switch {
					case effectiveIntensity >= 0.7:
						behavior = tier.High
					case effectiveIntensity >= 0.4:
						behavior = tier.Mid
					default:
						behavior = tier.Low
					}
					sb.WriteString(behavior + "\n")
				}
			}
			// Auto-reset if decayed below threshold
			if effectiveIntensity <= 0.1 {
				os.WriteFile(stateFile, []byte(`{"emotion":"neutral","intensity":0}`), 0644)
			}
		}
	}
	// (영혼 자문은 🔮 영혼 섹션에서 한 번만 출력 — 중복 제거됨)


	// ━━━ REGION SUMMARY: 한 줄 요약 (상세는 _rules.md) ━━━
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

	// ━━━ MODE SWITCH (강제) ━━━
	sb.WriteString(fmt.Sprintf("**작업 전 `%s\\{영역}\\_rules.md`를 반드시 읽는다** (cortex=코딩/NeuronFS, sensors=NAS/브랜드, prefrontal=방향)\n", brainRoot))
	sb.WriteString("⚠️ 읽지 않으면 금지 규칙 위반이 발생한다. view_file로 먼저 읽어라. MCP read_region 호출 금지(느림).\n")

	// ━━━ CODE MAP (1줄 축약) ━━━
	sb.WriteString(fmt.Sprintf("🗺️ 코드맵=뉴런 계층(cortex/dev/). 코드 수정 전 뉴런 읽기 필수. 플랫 뉴런 금지. `go vet ./...` 실행.\n\n"))

	// ━━━ 영혼 (중복 제거 — 이 한 곳에서만 출력) ━━━
	sb.WriteString("### 🔮 영혼\n")
	sb.WriteString("자문: 진짜야? 불충분? 편한길? 같은실수? 프리미엄? → 걸리면 다시\n")
	sb.WriteString("CoVe: 초안→검증질문→독립검증→수정본 | 실행후 증거보고(시뮬레이션 금지) | 복잡작업→단계분해\n\n")

	// ━━━ HIPPOCAMPUS: 최근 에피소드 (컨텍스트 힌트) ━━━
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

	// ━━━ RECENCY ANCHOR: 끝에 핵심 禁 규칙 재반복 (Lost-in-the-Middle 대응) ━━━
	// LLM은 맨 앞(primacy)과 맨 끝(recency)에 가장 주의를 집중한다.
	sb.WriteString("### 🔒 절대 규칙 (재확인)\n")
	var banReminders []string
	seenBanLeaf := make(map[string]bool) // leaf 기준 중복 제거
	for _, region := range result.ActiveRegions {
		for _, n := range region.Neurons {
			if n.IsDormant || (n.Counter+n.Dopamine) < 5 {
				continue
			}
			if !strings.ContainsAny(n.Path, "禁") {
				continue
			}
			// 하위 폴더 존재 시 상위 스킵 (중간 경로 제거)
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
			// 의미없는 짧은 sentence 제외
			trimmed := strings.TrimSpace(strings.ReplaceAll(sentence, "절대 금지:", ""))
			if len([]rune(trimmed)) < 2 {
				continue
			}
			// 동어반복 필터
			if strings.Contains(trimmed, "절대금지") || strings.Contains(trimmed, "절대 금지") {
				continue
			}
			// TOP 5에 이미 나온 것 스킵
			if top5Sentences[sentence] {
				continue
			}
			// cortex ban에 이미 나온 leaf 스킵
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

	sb.WriteString("<!-- NEURONFS:END -->\n")
	return sb.String()
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
