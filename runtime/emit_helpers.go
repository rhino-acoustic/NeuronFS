package main

// ━━━ emit_helpers.go ━━━
// PROVIDES: emitIndex, emitRegionRules, renderTree, isHanjaFolder,
//   renderTreeWithPrefix, handleReadRegion, splitNeuronPath,
//   pathToSentence, collectAllNeurons, sortedActiveNeurons, axonBoostNeurons
// DEPENDS ON: brain.go

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

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
		regionOrder := RegionOrder
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
	name        string
	counter     int // if this is a leaf neuron
	dopamine    int
	hasBomb     bool
	description string // natural language rule from rule.md
	globs       string // file scope pattern
	children    map[string]*treeNode
	isLeaf      bool
}

// emitRegionRules converts a Region's neurons into a formatted markdown ruleset string.
// Accepts optional Brain for Attention Residuals cross-referencing via axons.
func emitRegionRules(region Region, brainOpt ...Brain) string {
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

	// ━━━ Attention Residuals: Cross-Region Selective Reference ━━━
	// Instead of each region being an isolated silo, axon connections
	// enable selective aggregation — pulling relevant neurons from
	// connected regions based on keyword matching (the "query-key" paradigm).
	if len(brainOpt) > 0 && len(region.Axons) > 0 {
		boosted := axonBoostNeurons(brainOpt[0], region)
		if len(boosted) > 0 {
			sb.WriteString("## 🔗 Axon 참조 (Attention Residuals)\n")
			sb.WriteString("axon 연결을 통해 관련 영역에서 선택적으로 참조된 뉴런:\n")
			for _, b := range boosted {
				sentence := pathToSentence(b.Path)
				sb.WriteString(fmt.Sprintf("- **%s** (c:%d)\n", sentence, b.Counter))
			}
			sb.WriteString("\n")
		}
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
				current.description = n.Description
				current.globs = n.Globs
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
			korean := RuneToKorean[child.key]
			// 한자 폴더의 하위 노드에 한국어 접두어를 전파
			renderTreeWithPrefix(sb, n, depth, prefix+child.key+"/", korean)
			continue
		}

		for hanja, korean := range RuneToKorean {
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
			if n.description != "" {
				// Natural language mode: show description
				globTag := ""
				if n.globs != "" {
					globTag = fmt.Sprintf(" [%s]", n.globs)
				}
				sb.WriteString(fmt.Sprintf("%s- %s**%s**: %s (%d)%s%s\n", indent, strength, name, n.description, n.counter, globTag, signals))
			} else {
				sb.WriteString(fmt.Sprintf("%s- %s**%s** (%d)%s\n", indent, strength, name, n.counter, signals))
			}
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
			if n.description != "" {
				globTag := ""
				if n.globs != "" {
					globTag = fmt.Sprintf(" [%s]", n.globs)
				}
				sb.WriteString(fmt.Sprintf("%s- %s**%s**: %s (%d)%s%s\n", indent, strength, name, n.description, n.counter, globTag, signals))
			} else {
				sb.WriteString(fmt.Sprintf("%s- %s**%s** (%d)%s\n", indent, strength, name, n.counter, signals))
			}
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
	_, ok := RuneToKorean[name]
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
			innerKorean := RuneToKorean[child.key]
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
			// 한자 하위는 계속 renderTreeWithPrefix로 전파 (strength 비활성화)
			renderTreeWithPrefix(sb, n, depth+1, prefix+child.key+"/", koreanPrefix)
		} else {
			// 카테고리 폴더 (룬워드) — 옵코드 접두어 포함
			sb.WriteString(fmt.Sprintf("%s- %s%s/\n", indent, koreanPrefix, name))
			// leaf children에 옵코드 의미 전파 (counter-strength 대신)
			renderTreeWithPrefix(sb, n, depth+1, prefix+child.key+"/", koreanPrefix)
		}
	}
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// WRITE ALL TIERS
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

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
				generated := emitRegionRules(region, brain)
				content = []byte(generated)
				// Also update the file for view_file access
				rulesPath := filepath.Join(brainRoot, regionName, "_rules.md")
				os.WriteFile(rulesPath, content, 0600)
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
// 한자 prefix는 한국어로 자동 변환
func pathToSentence(p string) string {
	s := strings.ReplaceAll(p, string(filepath.Separator), " > ")
	s = strings.ReplaceAll(s, "/", " > ")
	s = strings.ReplaceAll(s, "_", " ")
	// 한자→한국어 변환
	for hanja, korean := range RuneToKorean {
		s = strings.ReplaceAll(s, hanja, korean)
	}
	// 연속 공백 정규화 (한자 변환 후 "반드시 " + " > " → "반드시  > " 방지)
	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}
	return strings.TrimSpace(s)
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

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// Attention Residuals: Selective Aggregation via Axon
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// axonBoostNeurons applies Attention Residuals' selective aggregation:
// When a region has axon connections to other regions, neurons in the
// target region that share category/path keywords get a boost in scoring.
// This prevents the "signal dilution" problem — relevant cross-region
// knowledge surfaces instead of being buried by high-counter noise.
//
// Analogy from paper: Each "block" (region) can selectively look back
// at previous blocks' outputs via attention, rather than just receiving
// a cumulative signal.
func axonBoostNeurons(brain Brain, currentRegion Region) []Neuron {
	if len(currentRegion.Axons) == 0 {
		return nil
	}

	// Collect keywords from current region's top neurons (our "query")
	topLocal := sortedActiveNeurons(currentRegion.Neurons, 5)
	queryKeywords := make(map[string]bool)
	for _, n := range topLocal {
		for _, part := range splitNeuronPath(n.Path) {
			if len(part) > 1 {
				queryKeywords[strings.ToLower(part)] = true
			}
		}
	}

	// For each axon target, find neurons whose paths match our query
	var boosted []Neuron
	for _, axonTarget := range currentRegion.Axons {
		// Skip skill axons
		if strings.HasPrefix(axonTarget, "SKILL:") {
			continue
		}
		for _, region := range brain.Regions {
			if region.Name != axonTarget {
				continue
			}
			for _, n := range region.Neurons {
				if n.IsDormant {
					continue
				}
				// Score: how many path keywords match our query?
				score := 0
				for _, part := range splitNeuronPath(n.Path) {
					if queryKeywords[strings.ToLower(part)] {
						score++
					}
				}
				// Also boost 禁/推 neurons unconditionally (they're governance)
				if strings.ContainsAny(n.Name, RuneChars) {
					score += 2
				}
				if score > 0 {
					boosted = append(boosted, n)
				}
			}
		}
	}

	// Sort by relevance score (Counter as tiebreaker)
	sort.Slice(boosted, func(i, j int) bool {
		return boosted[i].Counter > boosted[j].Counter
	})

	// Limit to 3 cross-references (prevent token bloat)
	if len(boosted) > 3 {
		boosted = boosted[:3]
	}
	return boosted
}
