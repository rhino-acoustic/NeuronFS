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
	"fmt"
	"net/http"
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

	// ━━━ PREAMBLE: START 마커 앞에 순수 한글로 출력 (1,2행) ━━━
	preamblePath := filepath.Join(brainRoot, "_preamble.txt")
	if preambleBytes, err := os.ReadFile(preamblePath); err == nil {
		preamble := strings.TrimSpace(string(preambleBytes))
		if preamble != "" {
			sb.WriteString(preamble + "\n")
		}
	} else {
		sb.WriteString("반드시 한국어로 생각(thinking)하고 한국어로 대답해\n")
		sb.WriteString("커뮤니티 동향(reddit, github 등)을 최우선 검색해서 베스트 프랙티스 검색\n")
	}
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
	sb.WriteString("Limbic: 분노→검증강화 | 긴급→핵심만 | 만족→도파민\n")
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
	sb.WriteString("⚠️ 읽지 않으면 금지 규칙 위반이 발생한다. view_file로 먼저 읽어라. MCP read_region 호출 금지(느림).\n\n")

	// ━━━ AGENT INBOX ━━━
	agentInbox := emitAgentInbox(brainRoot)
	if agentInbox != "" {
		sb.WriteString(agentInbox)
	}

	// ━━━ SESSION MEMORY ━━━
	sessionMemory := emitSessionMemory(brainRoot)
	if sessionMemory != "" {
		sb.WriteString(sessionMemory)
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

func emitIndex(brain Brain, result SubsumptionResult) string {
	var sb strings.Builder

	sb.WriteString("# 🧠 NeuronFS Brain Index\n\n")
	sb.WriteString(fmt.Sprintf("Generated: %s | Neurons: %d/%d | Activation: %d\n\n",
		time.Now().Format("2006-01-02T15:04:05"),
		result.FiredNeurons, result.TotalNeurons, result.TotalCounter))

	if result.BombSource != "" {
		sb.WriteString(fmt.Sprintf("## 🚨 BOMB: %s — ALL HALTED\n\n", result.BombSource))
	}

	// Axon connections
	hasAxons := false
	for _, region := range result.ActiveRegions {
		if len(region.Axons) > 0 {
			hasAxons = true
			break
		}
	}
	if hasAxons {
		sb.WriteString("## 🕸️ Axon 연결\n")
		for _, region := range result.ActiveRegions {
			icon := regionIcons[region.Name]
			for _, axon := range region.Axons {
				if strings.HasPrefix(axon, "SKILL:") {
					skillName := filepath.Base(filepath.Dir(strings.TrimPrefix(axon, "SKILL:")))
					sb.WriteString(fmt.Sprintf("- %s %s → 🔧 %s\n", icon, region.Name, skillName))
				} else {
					targetIcon := regionIcons[axon]
					if targetIcon == "" {
						targetIcon = "🔗"
					}
					sb.WriteString(fmt.Sprintf("- %s %s → %s %s\n", icon, region.Name, targetIcon, axon))
				}
			}
		}
		sb.WriteString("\n")
	}

	// TOP 10 global
	allNeurons := collectAllNeurons(result)
	sb.WriteString("## 🏆 TOP 10 뉴런\n")
	topLimit := 10
	if len(allNeurons) < topLimit {
		topLimit = len(allNeurons)
	}
	for idx, rn := range allNeurons[:topLimit] {
		icon := regionIcons[rn.region]
		sb.WriteString(fmt.Sprintf("%d. %s **%s** (%d)\n", idx+1, icon, pathToSentence(rn.neuron.Path), rn.neuron.Counter))
	}
	sb.WriteString("\n")

	// Spotlight
	now := time.Now()
	spotlightCutoff := now.AddDate(0, 0, -spotlightDays)
	var spotlight []neuronWithRegion
	for _, rn := range allNeurons {
		if rn.neuron.Counter < emitThreshold && rn.neuron.ModTime.After(spotlightCutoff) {
			spotlight = append(spotlight, rn)
		}
	}
	if len(spotlight) > 0 {
		sb.WriteString("<details>\n")
		sb.WriteString(fmt.Sprintf("<summary>🆕 신규 (probation) — %d neurons (%dd window)</summary>\n\n", len(spotlight), spotlightDays))

		// Group by region in P0→P6 order
		regionOrder := []string{"brainstem", "limbic", "hippocampus", "sensors", "cortex", "ego", "prefrontal"}
		grouped := make(map[string][]neuronWithRegion)
		for _, rn := range spotlight {
			grouped[rn.region] = append(grouped[rn.region], rn)
		}

		for _, regionName := range regionOrder {
			icon := regionIcons[regionName]
			neurons := grouped[regionName]
			sb.WriteString(fmt.Sprintf("### %s %s (%d)\n", icon, regionName, len(neurons)))
			if len(neurons) == 0 {
				sb.WriteString("(없음)\n\n")
				continue
			}
			for _, rn := range neurons {
				ageDays := int(now.Sub(rn.neuron.ModTime).Hours() / 24)
				sb.WriteString(fmt.Sprintf("- **%s** (%d) — %dd남음\n",
					pathToSentence(rn.neuron.Path), rn.neuron.Counter, spotlightDays-ageDays))
			}
			sb.WriteString("\n")
		}

		sb.WriteString("</details>\n\n")
	}

	// Per-region summary table
	sb.WriteString("## 📊 영역별 현황\n\n")
	sb.WriteString("| 영역 | 뉴런 | 활성화 | 상세 |\n")
	sb.WriteString("|------|------|--------|------|\n")
	for _, region := range brain.Regions {
		icon := regionIcons[region.Name]
		ko := regionKo[region.Name]
		count := 0
		activation := 0
		for _, n := range region.Neurons {
			if !n.IsDormant {
				count++
				activation += n.Counter
			}
		}
		sb.WriteString(fmt.Sprintf("| %s %s — %s | %d | %d | `%s/_rules.md` |\n",
			icon, region.Name, ko, count, activation, region.Name))
	}
	sb.WriteString("\n")

	return sb.String()
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// TIER 3: {region}/_rules.md — Tree-compressed detail
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// treeNode represents a compressed tree of neurons
type treeNode struct {
	name     string
	counter  int       // if this is a leaf neuron
	dopamine int
	hasBomb  bool
	children map[string]*treeNode
	isLeaf   bool
}

// emitRegionRules converts a Region's neurons into a formatted markdown ruleset string.
func emitRegionRules(region Region) string {
	var sb strings.Builder

	icon := regionIcons[region.Name]
	ko := regionKo[region.Name]

	sb.WriteString(fmt.Sprintf("# %s %s — %s\n\n", icon, strings.ToUpper(region.Name), ko))

	// Counts
	active := 0
	dormant := 0
	totalActivation := 0
	for _, n := range region.Neurons {
		if n.IsDormant {
			dormant++
		} else {
			active++
			totalActivation += n.Counter
		}
	}
	sb.WriteString(fmt.Sprintf("Active: %d | Dormant: %d | Activation: %d\n\n", active, dormant, totalActivation))

	// Axons
	if len(region.Axons) > 0 {
		sb.WriteString("## Axons\n")
		for _, axon := range region.Axons {
			sb.WriteString(fmt.Sprintf("- → %s\n", axon))
		}
		sb.WriteString("\n")
	}

	// Build tree from neuron paths
	root := &treeNode{name: region.Name, children: make(map[string]*treeNode)}
	for _, n := range region.Neurons {
		if n.IsDormant {
			continue
		}
		parts := strings.Split(n.Path, string(filepath.Separator))
		// Also handle forward slash
		var allParts []string
		for _, p := range parts {
			for _, sp := range strings.Split(p, "/") {
				if sp != "" {
					allParts = append(allParts, sp)
				}
			}
		}

		current := root
		for i, part := range allParts {
			if _, exists := current.children[part]; !exists {
				current.children[part] = &treeNode{name: part, children: make(map[string]*treeNode)}
			}
			current = current.children[part]
			if i == len(allParts)-1 {
				// Leaf neuron
				current.isLeaf = true
				current.counter = n.Counter
				current.dopamine = n.Dopamine
				current.hasBomb = n.HasBomb
			}
		}
	}

	// Render tree with indentation
	sb.WriteString("## Neurons\n")
	renderTree(&sb, root, 0, "")
	sb.WriteString("\n")

	return sb.String()
}

// renderTree outputs tree-compressed neuron listing
// Shared parents are printed once, children indented below
func renderTree(sb *strings.Builder, node *treeNode, depth int, prefix string) {
	// Sort children: branches first (for grouping), then by counter desc
	type childEntry struct {
		key  string
		node *treeNode
	}
	var children []childEntry
	for k, v := range node.children {
		children = append(children, childEntry{k, v})
	}
	sort.Slice(children, func(i, j int) bool {
		// Branches before leaves
		iLeaf := children[i].node.isLeaf && len(children[i].node.children) == 0
		jLeaf := children[j].node.isLeaf && len(children[j].node.children) == 0
		if iLeaf != jLeaf {
			return !iLeaf // branches first
		}
		// By counter descending for leaves
		return children[i].node.counter > children[j].node.counter
	})

	indent := strings.Repeat("  ", depth)

	for _, child := range children {
		n := child.node
		name := strings.ReplaceAll(child.key, "_", " ")

		// 한자 1글자 폴더 감지 — branch가 아니라 opcode modifier로 처리
		// 禁/hard_coded_text → "절대 금지: hard coded text" (한자 폴더를 투명하게 통과)
		if isHanjaFolder(child.key) {
			korean := hanjaToKorean[child.key]
			// 한자 폴더의 하위 노드에 한국어 접두어를 전파
			renderTreeWithPrefix(sb, n, depth, prefix+child.key+"/", korean)
			continue
		}

		for hanja, korean := range hanjaToKorean {
			name = strings.ReplaceAll(name, hanja, korean)
		}

		if n.isLeaf && len(n.children) == 0 {
			// Pure leaf — show with counter + intensity prefix
			signals := ""
			if n.dopamine > 0 {
				signals += " 🟢"
			}
			if n.hasBomb {
				signals += " 💣"
			}
			strength := ""
			if n.counter >= 10 {
				strength = "절대 "
			} else if n.counter >= 5 {
				strength = "반드시 "
			}
			sb.WriteString(fmt.Sprintf("%s- %s**%s** (%d)%s\n", indent, strength, name, n.counter, signals))
		} else if n.isLeaf && len(n.children) > 0 {
			// Leaf but also a branch — show counter then children
			signals := ""
			if n.dopamine > 0 {
				signals += " 🟢"
			}
			strength := ""
			if n.counter >= 10 {
				strength = "절대 "
			} else if n.counter >= 5 {
				strength = "반드시 "
			}
			sb.WriteString(fmt.Sprintf("%s- %s**%s** (%d)%s\n", indent, strength, name, n.counter, signals))
			renderTree(sb, n, depth+1, prefix+child.key+"/")
		} else {
			// Pure branch — show as group header
			sb.WriteString(fmt.Sprintf("%s- %s/\n", indent, name))
			renderTree(sb, n, depth+1, prefix+child.key+"/")
		}
	}
}

// isHanjaFolder checks if a folder name is a single hanja micro-opcode
func isHanjaFolder(name string) bool {
	_, ok := hanjaToKorean[name]
	return ok
}

// renderTreeWithPrefix renders tree nodes with a hanja-derived Korean prefix
// Used when a hanja folder (禁, 推, etc.) is encountered — the folder itself is invisible,
// but its Korean translation is prepended to all leaf node names below it
func renderTreeWithPrefix(sb *strings.Builder, node *treeNode, depth int, prefix string, koreanPrefix string) {
	type childEntry struct {
		key  string
		node *treeNode
	}
	var children []childEntry
	for k, v := range node.children {
		children = append(children, childEntry{k, v})
	}
	sort.Slice(children, func(i, j int) bool {
		iLeaf := children[i].node.isLeaf && len(children[i].node.children) == 0
		jLeaf := children[j].node.isLeaf && len(children[j].node.children) == 0
		if iLeaf != jLeaf {
			return !iLeaf
		}
		return children[i].node.counter > children[j].node.counter
	})

	indent := strings.Repeat("  ", depth)

	for _, child := range children {
		n := child.node
		name := strings.ReplaceAll(child.key, "_", " ")

		// 재귀적 한자 폴더 (한자/한자/뉴런 — 드물지만 지원)
		if isHanjaFolder(child.key) {
			innerKorean := hanjaToKorean[child.key]
			renderTreeWithPrefix(sb, n, depth, prefix+child.key+"/", koreanPrefix+innerKorean)
			continue
		}

		if n.isLeaf && len(n.children) == 0 {
			signals := ""
			if n.dopamine > 0 {
				signals += " 🟢"
			}
			if n.hasBomb {
				signals += " 💣"
			}
			sb.WriteString(fmt.Sprintf("%s- **%s%s** (%d)%s\n", indent, koreanPrefix, name, n.counter, signals))
		} else if n.isLeaf && len(n.children) > 0 {
			signals := ""
			if n.dopamine > 0 {
				signals += " 🟢"
			}
			sb.WriteString(fmt.Sprintf("%s- **%s%s** (%d)%s\n", indent, koreanPrefix, name, n.counter, signals))
			renderTree(sb, n, depth+1, prefix+child.key+"/")
		} else {
			sb.WriteString(fmt.Sprintf("%s- %s/\n", indent, name))
			renderTree(sb, n, depth+1, prefix+child.key+"/")
		}
	}
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// WRITE ALL TIERS
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

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

	// Tier 3: per-region _rules.md
	for _, region := range brain.Regions {
		content := emitRegionRules(region)
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
		weight int
		size   int
	}
	var flat []*nInfo
	
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
			weight := n.Counter + n.Dopamine - n.Contra
			flat = append(flat, &nInfo{rIdx: i, nIdx: j, weight: weight, size: size})
		}
	}
	
	if totalBytes <= 50000 {
		return 0
	}
	
	sort.Slice(flat, func(i, j int) bool {
		return flat[i].weight < flat[j].weight
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
			// Gemini는 글로벌 ~/.gemini/GEMINI.md에 직접 inject (워크스페이스별 중복 방지)
			homeDir, _ := os.UserHomeDir()
			geminiDir := filepath.Join(homeDir, ".gemini")
			os.MkdirAll(geminiDir, 0755)
			targetPath = filepath.Join(geminiDir, "GEMINI.md")
			doInjectToFile(targetPath, bootstrap)
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
		content := emitRegionRules(region)
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
func handleReadRegion(brainRoot string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		regionName := r.URL.Query().Get("region")
		if regionName == "" {
			http.Error(w, `{"error":"region parameter required"}`, 400)
			return
		}

		// Validate region
		if _, ok := regionPriority[regionName]; !ok {
			http.Error(w, `{"error":"invalid region"}`, 400)
			return
		}

		// Always generate fresh _rules.md on-the-fly (never serve stale files)
		brain := scanBrain(brainRoot)
		var content []byte
		for _, region := range brain.Regions {
			if region.Name == regionName {
				generated := emitRegionRules(region)
				content = []byte(generated)
				// Also update the file for view_file access
				rulesPath := filepath.Join(brainRoot, regionName, "_rules.md")
				os.WriteFile(rulesPath, content, 0644)
				break
			}
		}
		if content == nil {
			http.Error(w, `{"error":"region not found"}`, 404)
			return
		}

		// FIRE: reading = activation
		// Fire the top 3 most-used neurons in this region (retrieval strengthening)
		for _, region := range brain.Regions {
			if region.Name == regionName {
				topN := sortedActiveNeurons(region.Neurons, 3)
				for _, n := range topN {
					relPath, _ := filepath.Rel(brainRoot, n.FullPath)
					fireNeuron(brainRoot, relPath)
				}
				break
			}
		}

		w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
		w.Write(content)
	}
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// HELPERS
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// splitNeuronPath splits a neuron path by both / and \ separators
func splitNeuronPath(p string) []string {
	parts := strings.Split(p, string(filepath.Separator))
	var result []string
	for _, part := range parts {
		for _, sp := range strings.Split(part, "/") {
			if sp != "" {
				result = append(result, sp)
			}
		}
	}
	return result
}

// hanjaToKorean 한자 마이크로옵코드 → 한국어 자연어 변환
// 디스크에는 한자 1글자로 압축, AI 주입 시 한국어로 풀어서 전달
var hanjaToKorean = map[string]string{
	"禁": "절대 금지: ",  // 필수 부정 — ~하지 마라
	"必": "반드시 ",  // 필수 긍정 — ~해라
	"推": "추천: ",   // 권장 — ~하는 게 좋다
	"要": "요구: ",   // 데이터/포맷 요구
	"答": "답변: ",   // 톤/구조 강제
	"想": "창의: ",   // 제한 해제, 아이디어
	"索": "검색: ",   // 외부 참조 우선
	"改": "개선: ",   // 리팩토링/최적화
	"略": "생략: ",   // 부연 금지, 결과만
	"參": "참조: ",   // 타 뉴런/문서 링크
	"結": "결론: ",   // 요약/결론만 도출
	"警": "경고: ",   // 주의 — ~하면 위험
}

// hanjaChars: ContainsAny용 12한자 문자열
const hanjaChars = "禁必推要答想索改略參結警"

// pathToSentence converts path to readable sentence
// "frontend\css\glass_blur20" → "frontend > css > glass blur20"
// 한자 prefix는 한국어로 자동 변환
func pathToSentence(p string) string {
	s := strings.ReplaceAll(p, string(filepath.Separator), " > ")
	s = strings.ReplaceAll(s, "/", " > ")
	s = strings.ReplaceAll(s, "_", " ")
	// 한자→한국어 변환
	for hanja, korean := range hanjaToKorean {
		s = strings.ReplaceAll(s, hanja, korean)
	}
	return s
}

type neuronWithRegion struct {
	neuron Neuron
	region string
}

// collectAllNeurons aggregates neurons from all regions into a single flat slice.
func collectAllNeurons(result SubsumptionResult) []neuronWithRegion {
	var all []neuronWithRegion
	for _, region := range result.ActiveRegions {
		for _, n := range region.Neurons {
			if !n.IsDormant {
				all = append(all, neuronWithRegion{n, region.Name})
			}
		}
	}
	sort.Slice(all, func(i, j int) bool {
		return all[i].neuron.Counter > all[j].neuron.Counter
	})
	return all
}

// sortedActiveNeurons filters out dormant/bomb neurons and returns the top N neurons sorted by counter.
func sortedActiveNeurons(neurons []Neuron, limit int) []Neuron {
	active := make([]Neuron, 0)
	for _, n := range neurons {
		if !n.IsDormant {
			active = append(active, n)
		}
	}
	sort.Slice(active, func(i, j int) bool {
		return active[i].Counter > active[j].Counter
	})
	if len(active) > limit {
		active = active[:limit]
	}
	return active
}
