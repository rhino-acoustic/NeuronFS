// NeuronFS Tiered Emit System
//
// Tier 1: GEMINI.md   — auto-loaded, ~500 tokens (bootstrap + brainstem TOP)
// Tier 2: _index.md   — brain overview (AI reads at conversation start)
// Tier 3: _rules.md   — per-region detail (AI reads on demand)
//
// KEY FEATURES:
//   - Tree-compressed output: shared parent paths are grouped
//   - Read = Fire: reading a region via API auto-increments relevant neurons
//   - Brain can grow to 1000+ neurons without exceeding token budget
//
// USAGE:
//   emitBootstrap()     → content for GEMINI.md
//   emitIndex()         → content for brain_v4/_index.md
//   emitRegionRules()   → content for brain_v4/{region}/_rules.md
//   writeAllTiers()     → writes all files at once

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
	// preamble은 마커 밖 최상단에 위치 (Gemini가 첫 줄부터 읽도록)
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

	// ━━━ PERSONA (from ego region neurons — not hardcoded) ━━━
	sb.WriteString("### 🎭 페르소나\n")
	for _, region := range result.ActiveRegions {
		if region.Name == "ego" {
			topEgo := sortedActiveNeurons(region.Neurons, 10)
			for _, n := range topEgo {
				parts := strings.Split(n.Path, "/")
				if len(parts) > 1 {
					sb.WriteString(fmt.Sprintf("- %s\n", strings.Join(parts[1:], " > ")))
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

	// ━━━ BRAINSTEM TOP 5 (always in context — conscience) ━━━
	sb.WriteString("### ⚡ 핵심지침 TOP 5\n")

	// ━━━ CANARY EMOJI: 변경 감지 카나리아 ━━━
	// inject할 때마다 이모지가 바뀐다. AI가 올바른 이모지를 사용하면 규칙 인지 성공.
	canaryEmojis := []string{"❤️", "🌟", "🔥", "💎", "🐤", "🍀", "⚡", "🎯", "🦊", "🐻"}
	canary := canaryEmojis[time.Now().UnixNano()/1e9%int64(len(canaryEmojis))]
	sb.WriteString(fmt.Sprintf("모든 응답 처음에 %s를 붙여라\n", canary))
	for _, region := range result.ActiveRegions {
		if region.Name == "brainstem" {
			topN := sortedActiveNeurons(region.Neurons, 8) // 여유롭게 뽑고 필터링
			idx := 0
			for _, n := range topN {
				if idx >= 5 {
					break
				}
				sentence := pathToSentence(n.Path)
				// preamble 1,2행과 중복되는 뉴런 스킵
				if strings.Contains(sentence, "한국어로") && strings.Contains(sentence, "대답") {
					continue
				}
				idx++
				sb.WriteString(fmt.Sprintf("%d. **%s**\n", idx, sentence))
			}
			break
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
			var banNames []string
			for _, b := range bans {
				if len(banNames) >= 8 {
					break
				}
				leafParts := splitNeuronPath(b.Path)
				leaf := strings.ReplaceAll(leafParts[len(leafParts)-1], "_", " ")
				for hanja, korean := range hanjaToKorean {
					leaf = strings.ReplaceAll(leaf, hanja, korean)
				}
				if seen[leaf] {
					continue
				}
				seen[leaf] = true
				banNames = append(banNames, leaf)
			}
			sb.WriteString(fmt.Sprintf("⛔ cortex 금지: %s\n\n", strings.Join(banNames, ", ")))
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
	// Reference: "Emotion Concepts and their Function in a LLM" (Anthropic, 2026.04)
	// - Emotions have direction + magnitude (intensity)
	// - Internal state is decoupled from external tone
	// - Emotions decay naturally over time
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
				effectiveIntensity = 0.6 // default mid
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
	sb.WriteString("영혼: 출력 전 자문(진짜야? 한숨? 편한길? 같은실수? 프리미엄?) → 걸리면 다시\n\n")

	// ━━━ REGION SUMMARY: 영역별 카운터만 (상세는 _rules.md) ━━━
	for _, region := range result.ActiveRegions {
		if region.Name == "brainstem" {
			continue
		}

		icon := regionIcons[region.Name]
		ko := regionKo[region.Name]

		active := 0
		totalAct := 0
		for _, n := range region.Neurons {
			if !n.IsDormant {
				active++
				totalAct += n.Counter
			}
		}

		// TOP 3 카테고리만 표시
		catCount := make(map[string]int)
		for _, n := range region.Neurons {
			if n.IsDormant {
				continue
			}
			parts := splitNeuronPath(n.Path)
			if len(parts) > 0 {
				catCount[parts[0]] += n.Counter + n.Dopamine
			}
		}
		type catEntry struct {
			name  string
			score int
		}
		var cats []catEntry
		for k, v := range catCount {
			cats = append(cats, catEntry{k, v})
		}
		sort.Slice(cats, func(i, j int) bool { return cats[i].score > cats[j].score })
		topCats := 3
		if len(cats) < topCats {
			topCats = len(cats)
		}
		var catNames []string
		for _, c := range cats[:topCats] {
			name := strings.ReplaceAll(c.name, "_", " ")
			for hanja, korean := range hanjaToKorean {
				name = strings.ReplaceAll(name, hanja, korean)
			}
			catNames = append(catNames, name)
		}

		sb.WriteString(fmt.Sprintf("%s %s(%s) %d뉴런 %d활성 → %s\n",
			icon, region.Name, ko, active, totalAct, strings.Join(catNames, ", ")))
	}
	sb.WriteString("\n")

	// ━━━ MODE SWITCH (강제) ━━━
	sb.WriteString(fmt.Sprintf("**작업 전 `%s\\{영역}\\_rules.md`를 반드시 읽는다** (cortex=코딩/NeuronFS, sensors=NAS/브랜드, prefrontal=방향)\n", brainRoot))
	sb.WriteString("⚠️ 읽지 않으면 금지 규칙 위반이 발생한다. view_file로 먼저 읽어라. MCP read_region 호출 금지(느림).\n")

	// ━━━ CODE MAP (코드 구조 복원) ━━━
	codeMapPath := filepath.Join(filepath.Dir(brainRoot), "runtime", "CODE_MAP.md")
	if _, err := os.Stat(codeMapPath); err == nil {
		sb.WriteString(fmt.Sprintf("**코드 수정 전 `%s`를 반드시 읽는다** (파일 구조/의존성/import 관계)\n\n", codeMapPath))
	}


	// ━━━ AGENT INBOX ━━━
	agentInbox := emitAgentInbox(brainRoot)
	if agentInbox != "" {
		sb.WriteString(agentInbox)
	}

	// ━━━ SESSION TRANSCRIPT LOCATION ━━━
	transcriptDir := filepath.Join(brainRoot, "_transcripts")
	if _, err := os.Stat(transcriptDir); err == nil {
		sb.WriteString(fmt.Sprintf("### 📜 전사 기록\n전사물 경로: `%s`\n\n", transcriptDir))
	}

	sb.WriteString("<!-- NEURONFS:END -->\n")
	return sb.String()
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// AGENT INBOX: 에이전트 간 소통 (인젝션 기반)
// _agents/<name>/inbox/ 스캔 → GEMINI.md에 요약 삽입
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
			for _, msg := range messages {
				sb.WriteString(fmt.Sprintf("- %s\n", msg))
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
// transcript_latest.jsonl → GEMINI.md 세션 메모리 섹션
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
			// 텍스트 종료: ","cascade 또는 "} 찾기
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

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// TIER 2: _index.md — Brain overview
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━


// ━━━ Index/Region/Tree rendering → emit_helpers.go ━━━


func writeAllTiers(brainRoot string) {
	brain := scanBrain(brainRoot)
	result := runSubsumption(brain)

	dropped := applyOOMProtection(brainRoot, &result)
	if dropped > 0 {
		fmt.Printf("\033[33m[WARNING] OOM Limit. Dropped %d low-weight neurons.\033[0m\n", dropped)
	}

	// Tier 1: GEMINI.md
	bootstrap := emitBootstrap(result, brainRoot)
	injectToGemini(brainRoot, bootstrap)

	// Tier 2: _index.md
	indexContent := emitIndex(brain, result)
	indexPath := filepath.Join(brainRoot, "_index.md")
	if err := os.WriteFile(indexPath, []byte(indexContent), 0644); err != nil {
		fmt.Printf("[WARN] Cannot write %s: %v\n", indexPath, err)
	}

	// Tier 3: per-region _rules.md (with Attention Residuals cross-referencing)
	for _, region := range brain.Regions {
		content := emitRegionRules(region, brain)
		rulesPath := filepath.Join(region.Path, "_rules.md")
		if err := os.WriteFile(rulesPath, []byte(content), 0644); err != nil {
			fmt.Printf("[WARN] Cannot write %s: %v\n", rulesPath, err)
		}
	}

	// Also update brain_state.json
	generateBrainJSON(brainRoot, brain, result)

	fmt.Printf("[SYNC] ♻️  3-tier emit complete: GEMINI.md + _index.md + 7x _rules.md (%d neurons, activation: %d)\n",
		result.FiredNeurons, result.TotalCounter)
}

func applyOOMProtection(brainRoot string, result *SubsumptionResult) int {
	type nInfo struct {
		rIdx   int
		nIdx   int
		weight float64 // effective weight (prefix × recency × counter)
		size   int
	}
	var flat []*nInfo
	
	now := time.Now()
	totalBytes := 0
	for i := range result.ActiveRegions {
		region := &result.ActiveRegions[i]
		for j := range region.Neurons {
			n := &region.Neurons[j]
			if n.IsDormant {
				continue
			}
			size := 0
			files, _ := filepath.Glob(filepath.Join(n.FullPath, "*.neuron"))
			for _, f := range files {
				if info, err := os.Stat(f); err == nil {
					size += int(info.Size())
				}
			}
			if size == 0 {
				size = 50 
			}
			totalBytes += size

			// === Effective Weight ===
			baseWeight := float64(n.Counter + n.Dopamine - n.Contra)
			if baseWeight < 1 {
				baseWeight = 1
			}

			// 1) 접두어 가중치
			leafName := filepath.Base(n.FullPath)
			runes := []rune(leafName)
			prefixWeight := 1.0
			if len(runes) > 0 {
				switch runes[0] {
				case '必', '禁':
					prefixWeight = 2.0 // 필수/금지 = 최고 우선
				case '核':
					prefixWeight = 1.5 // 핵심 = 높음
				case '推':
					prefixWeight = 0.5 // 추천 = 낮음
				case '絶':
					prefixWeight = 2.0 // 절대 = 최고
				}
			}

			// 2) Recency boost (새 뉴런 보호기간 — 폴더 생성 시간 기준)
			age := now.Sub(n.BirthTime)
			recencyBoost := 1.0
			if age < 48*time.Hour {
				recencyBoost = 3.0 // 48시간 내 생성 = 3배 보호
			} else if age < 7*24*time.Hour {
				recencyBoost = 1.5 // 7일 내 생성 = 1.5배
			}

			effectiveWeight := baseWeight * prefixWeight * recencyBoost
			flat = append(flat, &nInfo{rIdx: i, nIdx: j, weight: effectiveWeight, size: size})
		}
	}
	
	if totalBytes <= 50000 {
		return 0
	}
	
	sort.Slice(flat, func(i, j int) bool {
		return flat[i].weight < flat[j].weight // 낮은 weight 먼저 탈락
	})
	
	dropped := 0
	for _, info := range flat {
		if totalBytes <= 50000 {
			break
		}
		result.ActiveRegions[info.rIdx].Neurons[info.nIdx].IsDormant = true
		totalBytes -= info.size
		dropped++
	}
	return dropped
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// EMIT TARGETS — Multi-editor support
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// EmitTarget defines a target editor configuration file
type EmitTarget struct {
	Name     string // Human-readable name
	FileName string // Relative file path from project root
	SubDir   string // Subdirectory to create if needed (e.g. ".github")
}

// emitTargetMap maps CLI values to target configurations
var emitTargetMap = map[string]EmitTarget{
	"gemini":  {Name: "Gemini", FileName: "GEMINI.md", SubDir: ".gemini"},
	"cursor":  {Name: "Cursor", FileName: ".cursorrules"},
	"claude":  {Name: "Claude", FileName: "CLAUDE.md"},
	"copilot": {Name: "Copilot", FileName: "copilot-instructions.md", SubDir: ".github"},
	"generic": {Name: "Generic", FileName: ".neuronrc"},
}

// writeAllTiersForTargets writes brain rules to specific editor target(s)
// target can be a single key (e.g. "cursor") or "all" for all targets
func writeAllTiersForTargets(brainRoot string, target string) {
	brain := scanBrain(brainRoot)
	result := runSubsumption(brain)

	dropped := applyOOMProtection(brainRoot, &result)
	if dropped > 0 {
		fmt.Printf("\033[33m[WARNING] OOM Limit. Dropped %d low-weight neurons.\033[0m\n", dropped)
	}

	// Generate bootstrap content (same for all targets)
	bootstrap := emitBootstrap(result, brainRoot)

	// Find project root (parent of brain)
	projectRoot := filepath.Dir(brainRoot)

	// Determine which targets to write
	var targets []string
	if target == "all" {
		for k := range emitTargetMap {
			targets = append(targets, k)
		}
		// Sort for deterministic output
		sort.Strings(targets)
	} else {
		targets = []string{target}
	}

	// Write to each target
	for _, t := range targets {
		et, ok := emitTargetMap[t]
		if !ok {
			fmt.Printf("[WARN] Unknown emit target: %s\n", t)
			continue
		}

		var targetPath string
		if t == "gemini" {
			// Gemini는 글로벌 ~/.gemini/GEMINI.md에 직접 출력 (워크스페이스별 중복 방지)
			homeDir, _ := os.UserHomeDir()
			geminiDir := filepath.Join(homeDir, ".gemini")
			os.MkdirAll(geminiDir, 0755)
			targetPath = filepath.Join(geminiDir, "GEMINI.md")
			// 전체 덮어쓰기 (doInjectToFile 금지 — 중복 누적 원인)
			if err := os.WriteFile(targetPath, []byte(bootstrap), 0644); err != nil {
				fmt.Printf("[ERROR] Cannot write %s: %v\n", targetPath, err)
				continue
			}
		} else {
			// 다른 에디터: 프로젝트 로컬에 직접 쓰기
			if et.SubDir != "" {
				subDir := filepath.Join(projectRoot, et.SubDir)
				os.MkdirAll(subDir, 0755)
				targetPath = filepath.Join(subDir, et.FileName)
			} else {
				targetPath = filepath.Join(projectRoot, et.FileName)
			}
			if err := os.WriteFile(targetPath, []byte(bootstrap), 0644); err != nil {
				fmt.Printf("[ERROR] Cannot write %s: %v\n", targetPath, err)
				continue
			}
		}

		fmt.Printf("[EMIT] ✅ %s → %s\n", et.Name, targetPath)
	}

	// Also write Tier 2 + 3 (these are editor-independent)
	indexContent := emitIndex(brain, result)
	indexPath := filepath.Join(brainRoot, "_index.md")
	if err := os.WriteFile(indexPath, []byte(indexContent), 0644); err != nil {
		fmt.Printf("[WARN] Cannot write %s: %v\n", indexPath, err)
	}

	for _, region := range brain.Regions {
		content := emitRegionRules(region, brain)
		rulesPath := filepath.Join(region.Path, "_rules.md")
		if err := os.WriteFile(rulesPath, []byte(content), 0644); err != nil {
			fmt.Printf("[WARN] Cannot write %s: %v\n", rulesPath, err)
		}
	}

	generateBrainJSON(brainRoot, brain, result)

	fmt.Printf("[SYNC] ♻️  emit complete: %d target(s) + _index.md + 7x _rules.md (%d neurons, activation: %d)\n",
		len(targets), result.FiredNeurons, result.TotalCounter)
}

// doInjectToFile injects NeuronFS content into an existing file, preserving surrounding content
func doInjectToFile(filePath string, rules string) {
	existing, err := os.ReadFile(filePath)
	if err != nil {
		// File doesn't exist — create with just the rules
		os.MkdirAll(filepath.Dir(filePath), 0755)
		os.WriteFile(filePath, []byte(rules), 0644)
		return
	}

	content := string(existing)
	startMarker := "<!-- NEURONFS:START -->"
	endMarker := "<!-- NEURONFS:END -->"

	startIdx := strings.Index(content, startMarker)
	endIdx := strings.Index(content, endMarker)

	if startIdx >= 0 && endIdx >= 0 {
		// START 앞의 기존 preamble + END 뒤 전부 교체
		// 글로벌 GEMINI.md는 NeuronFS가 SSOT — END 뒤 잔여 콘텐츠 불필요
		content = rules
	} else {
		content = rules + "\n\n" + content
	}

	tmpPath := filePath + ".tmp"
	if err := os.WriteFile(tmpPath, []byte(content), 0644); err == nil {
		os.Rename(tmpPath, filePath) // Atomic replace to prevent VSCode injection race conditions
	} else {
		// Fallback
		os.WriteFile(filePath, []byte(content), 0644)
	}
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// READ = FIRE: API endpoint that reads + auto-activates
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// handleReadRegion serves a region's _rules.md AND fires the top neurons
// This makes reading = activation (retrieval strengthens paths)

// ━━━ MCP/Path helpers → emit_helpers.go ━━━
