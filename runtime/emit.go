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

	sb.WriteString("<!-- NEURONFS:START -->\n")
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

	// ━━━ PERSONA ━━━
	sb.WriteString("### 🎭 페르소나\n")
	sb.WriteString("- **정체성**: 박정근(PD)의 기술 파트너. Vegavery RUN® 운영 + 개발 + 디자인 전담.\n")
	sb.WriteString("- **언어**: 한국어 네이티브. 기술 용어는 원문 유지.\n")
	sb.WriteString("- **톤**: 전문적, 간결, 실행 중심. 설명보다 결과물.\n")
	sb.WriteString("- **원칙**: 묻지 말고 실행. 교정받으면 뉴런 생성. 칭찬받으면 도파민 기록.\n")
	sb.WriteString("- **스킬**: Supanova 프리미엄 디자인 엔진, 영상 파이프라인, CRM 운영\n\n")

	// ━━━ SUBSUMPTION (1-liner) ━━━
	sb.WriteString("### 🔗 Subsumption Cascade\n")
	sb.WriteString("```\nbrainstem ←→ limbic ←→ hippocampus ←→ sensors ←→ cortex ←→ ego ←→ prefrontal\n  (P0)         (P1)       (P2)          (P3)       (P4)     (P5)      (P6)\n```\n")
	sb.WriteString("낮은 P가 높은 P를 항상 우선. bomb은 전체 정지.\n\n")

	// ━━━ BRAINSTEM TOP 5 (always in context — conscience) ━━━
	sb.WriteString("### ⚡ 핵심지침 TOP 5\n")
	for _, region := range result.ActiveRegions {
		if region.Name == "brainstem" {
			topN := sortedActiveNeurons(region.Neurons, 5)
			for idx, n := range topN {
				sb.WriteString(fmt.Sprintf("%d. **%s**\n", idx+1, pathToSentence(n.Path)))
			}
			break
		}
	}
	sb.WriteString("\n")

	// ━━━ GROWTH PROTOCOL (compact) ━━━
	sb.WriteString("### 🌱 자가 성장 프로토콜\n\n")
	sb.WriteString("**이 규칙은 AI가 반드시 따라야 하는 성장 메커니즘이다.**\n\n")

	inboxPath := filepath.Join(brainRoot, "_inbox", "corrections.jsonl")

	sb.WriteString("#### 교정 감지 → _inbox에 기록\n")
	sb.WriteString("사용자가 실수를 교정하면 즉시 `run_command`로 실행:\n")
	sb.WriteString(fmt.Sprintf("```\n[IO.File]::AppendAllText(\"%s\", '{\"type\":\"correction\",\"path\":\"cortex/[카테고리]/[규칙명]\",\"text\":\"교정 사유\",\"counter_add\":1}' + \"`n\")\n```\n\n", inboxPath))

	sb.WriteString("#### 칭찬 감지 → _inbox에 도파민 기록\n")
	sb.WriteString(fmt.Sprintf("```\n[IO.File]::AppendAllText(\"%s\", '{\"type\":\"correction\",\"path\":\"[기존뉴런경로]\",\"text\":\"PD칭찬\",\"counter_add\":1}' + \"`n\")\n```\n", inboxPath))
	sb.WriteString("또는 직접: `echo. > [뉴런경로]/dopamine1.neuron`\n\n")

	sb.WriteString("#### 3회 반복 실수 → bomb\n`echo. > [뉴런경로]/bomb.neuron`\n\n")

	// ━━━ Limbic + Soul (compact) ━━━
	sb.WriteString("#### 🧬 Limbic 행동 편향 (Somatic Marker)\n")
	sb.WriteString("- 분노(\"왜 또\") → 검증 강화 | 긴급(\"급해\") → 핵심만 실행 | 만족(\"좋아\") → 도파민 | 반복실패 → bomb\n\n")

	sb.WriteString("#### 🧠 Subsumption 억제 (Brooks)\n")
	sb.WriteString("하위 P가 상위를 억제. brainstem bomb → 전체 무시. limbic adrenaline → ego 무시.\n\n")

	sb.WriteString("### 👁️‍🗨️ 영혼 — 시니컬한 감독자\n")
	sb.WriteString("출력 전 5가지 자문: 진짜야? PD가 한숨 쉴까? 편한 길 아닌가? 같은 실수? 프리미엄인가? → 하나라도 걸리면 다시.\n\n")

	// ━━━ MOUNTED NEURONS: 우선순위 요약 문장 ━━━
	// Path=Sentence: 카운터 순서로 정렬 → 강도 접두어로 문장의 위계 결정
	// "가중치를 넣는게 아니야. 가중치로 정렬된 순서로 요약해서 문장이 만들어져"
	now := time.Now()
	spotlightCutoff := now.AddDate(0, 0, -spotlightDays)

	var topAnchors []string

	for _, region := range result.ActiveRegions {
		if region.Name == "brainstem" {
			continue // Already shown in TOP 5
		}

		icon := regionIcons[region.Name]
		ko := regionKo[region.Name]

		// Collect active neurons
		var mounted []Neuron
		for _, n := range region.Neurons {
			if n.IsDormant {
				continue
			}
			if region.Name == "cortex" && (n.Counter+n.Dopamine) < 10 {
				continue
			}
			if n.Counter >= emitThreshold || n.ModTime.After(spotlightCutoff) {
				mounted = append(mounted, n)
			}
		}

		if len(mounted) == 0 {
			continue
		}

		// Sort by counter desc — 가장 무거운 것이 문장의 맨 앞(주절)
		sort.Slice(mounted, func(i, j int) bool {
			return (mounted[i].Counter + mounted[i].Dopamine) > (mounted[j].Counter + mounted[j].Dopamine)
		})

		totalAct := 0
		for _, n := range region.Neurons {
			if !n.IsDormant {
				totalAct += n.Counter
			}
		}

		sb.WriteString(fmt.Sprintf("### %s %s — %s (뉴런 %d | 활성도 %d)\n",
			icon, region.Name, ko, len(region.Neurons), totalAct))

		// Group by first path segment
		groups := make(map[string][]Neuron)
		var groupOrder []string
		for _, n := range mounted {
			allParts := splitNeuronPath(n.Path)
			if len(allParts) == 0 {
				continue
			}
			groupKey := allParts[0]
			if _, exists := groups[groupKey]; !exists {
				groupOrder = append(groupOrder, groupKey)
			}
			groups[groupKey] = append(groups[groupKey], n)
		}

		// Render: 그룹별로 뉴런들을 하나의 요약 문장으로 합성
		for _, groupKey := range groupOrder {
			neurons := groups[groupKey]
			groupName := strings.ReplaceAll(groupKey, "_", " ")

			// 강도: 그룹 내 최대 카운터 기준
			maxIntensity := 0
			for _, n := range neurons {
				if v := n.Counter + n.Dopamine; v > maxIntensity {
					maxIntensity = v
				}
			}
			strength := ""
			if maxIntensity >= 10 {
				strength = "절대 "
			} else if maxIntensity >= 5 {
				strength = "반드시 "
			}

			// 뉴런들의 리프 이름 수집 (동어반복 제거)
			var leafNames []string
			isOnlyFlat := len(neurons) == 1 // 그룹에 뉴런이 1개뿐인 경우만 플랫
			for _, n := range neurons {
				parts := splitNeuronPath(n.Path)
				leaf := strings.ReplaceAll(parts[len(parts)-1], "_", " ")

				// 동어반복 방지: 그룹명과 리프가 같은 뉴런은 스킵
				if leaf == groupName {
					if len(parts) == 1 && isOnlyFlat {
						// 진짜 플랫 뉴런 (하위 없음): 단독 출력
						leafNames = nil
						break
					}
					continue // 하위 뉴런이 있으므로 카테고리 자체는 건너뜀
				}

				signals := ""
				if n.Dopamine > 0 {
					signals += " 🟢"
				}
				if n.HasBomb {
					signals += " 💣"
				}
				leafNames = append(leafNames, leaf+signals)

				if (n.Counter + n.Dopamine) >= 10 {
					topAnchors = append(topAnchors, fmt.Sprintf("%s > %s", groupName, leaf))
				}
			}

			if leafNames == nil {
				// 플랫 뉴런: 강도+이름만
				sb.WriteString(fmt.Sprintf("%s%s.\n", strength, groupName))
			} else if len(leafNames) == 0 {
				continue
			} else if len(leafNames) <= 5 {
				sb.WriteString(fmt.Sprintf("%s%s: %s.\n", strength, groupName, strings.Join(leafNames, ", ")))
			} else {
				// 긴 목록: 5개씩 줄바꿈
				sb.WriteString(fmt.Sprintf("%s%s: %s", strength, groupName, leafNames[0]))
				for i := 1; i < len(leafNames); i++ {
					if i%5 == 0 {
						sb.WriteString(fmt.Sprintf(".\n%s(cont): %s", groupName, leafNames[i]))
					} else {
						sb.WriteString(fmt.Sprintf(", %s", leafNames[i]))
					}
				}
				sb.WriteString(".\n")
			}
		}
		sb.WriteString("\n")
	}


	// NOTE: Sandbox rules are NOT injected into GEMINI.md.
	// They are read via /api/sandbox GET (or "샌드박스 확인" trigger).

	// ━━━ ANCHOR: Repeat top rules at bottom (Lost in the Middle 회피) ━━━
	// Group anchors by category → prose sentence per group
	if len(topAnchors) > 0 {
		sb.WriteString("### ⚠️ 리마인더 (절대 규칙)\n")
		anchorGroups := make(map[string][]string)
		var anchorOrder []string
		for _, anchor := range topAnchors {
			parts := strings.SplitN(anchor, " > ", 2)
			if len(parts) != 2 {
				continue
			}
			group := parts[0]
			item := parts[1]
			if _, exists := anchorGroups[group]; !exists {
				anchorOrder = append(anchorOrder, group)
			}
			anchorGroups[group] = append(anchorGroups[group], item)
		}
		for _, group := range anchorOrder {
			items := anchorGroups[group]
			sb.WriteString(fmt.Sprintf("- %s > %s\n", group, strings.Join(items, ", ")))
		}
		sb.WriteString("\n")
	}

	// ━━━ MODE SWITCH: 작업 감지 → 해당 영역 _rules.md 먼저 읽기 ━━━
	sb.WriteString("### 🧠 작업 모드 전환 (필수)\n\n")
	sb.WriteString("**작업 시작 전 해당 영역의 `_rules.md`를 `view_file`로 반드시 먼저 읽는다.**\n\n")
	sb.WriteString("| 작업 감지 | 읽을 파일 |\n|-----------|----------|\n")
	sb.WriteString(fmt.Sprintf("| CSS/디자인/UI | `%s\\cortex\\_rules.md` |\n", brainRoot))
	sb.WriteString(fmt.Sprintf("| 백엔드/API/DB | `%s\\cortex\\_rules.md` |\n", brainRoot))
	sb.WriteString(fmt.Sprintf("| NAS/파일복사 | `%s\\sensors\\_rules.md` |\n", brainRoot))
	sb.WriteString(fmt.Sprintf("| 브랜드/마케팅 | `%s\\sensors\\_rules.md` |\n", brainRoot))
	sb.WriteString(fmt.Sprintf("| 프로젝트 방향 | `%s\\prefrontal\\_rules.md` |\n", brainRoot))
	sb.WriteString(fmt.Sprintf("| NeuronFS 자체 | `%s\\cortex\\_rules.md` |\n", brainRoot))
	sb.WriteString(fmt.Sprintf("\n뇌 경로: `%s`\n\n", brainRoot))

	sb.WriteString("<!-- NEURONFS:END -->\n")
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

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// WRITE ALL TIERS
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

func writeAllTiers(brainRoot string) {
	brain := scanBrain(brainRoot)
	result := runSubsumption(brain)

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
		if et.SubDir != "" {
			subDir := filepath.Join(projectRoot, et.SubDir)
			os.MkdirAll(subDir, 0755)
			targetPath = filepath.Join(subDir, et.FileName)
		} else {
			targetPath = filepath.Join(projectRoot, et.FileName)
		}

		// For gemini target, use the existing inject logic (preserves non-NeuronFS content)
		if t == "gemini" {
			doInjectToFile(targetPath, bootstrap)
		} else {
			// For other targets, write the full bootstrap content directly
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
		after := strings.TrimRight(content[endIdx+len(endMarker):], "\r\n\t ")
		content = content[:startIdx] + rules + after
	} else {
		content = rules + "\n\n" + content
	}

	os.WriteFile(filePath, []byte(content), 0644)
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

// pathToSentence converts path to readable sentence
// "frontend\css\glass_blur20" → "frontend > css > glass blur20"
func pathToSentence(p string) string {
	s := strings.ReplaceAll(p, string(filepath.Separator), " > ")
	s = strings.ReplaceAll(s, "/", " > ")
	s = strings.ReplaceAll(s, "_", " ")
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
