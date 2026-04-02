// NeuronFS Tiered Emit System
//
// Tier 1: GEMINI.md   ??auto-loaded, ~500 tokens (bootstrap + brainstem TOP)
// Tier 2: _index.md   ??brain overview (AI reads at conversation start)
// Tier 3: _rules.md   ??per-region detail (AI reads on demand)
//
// KEY FEATURES:
//   - Tree-compressed output: shared parent paths are grouped
//   - Read = Fire: reading a region via API auto-increments relevant neurons
//   - Brain can grow to 1000+ neurons without exceeding token budget
//
// USAGE:
//   emitBootstrap()     ??content for GEMINI.md
//   emitIndex()         ??content for brain_v4/_index.md
//   emitRegionRules()   ??content for brain_v4/{region}/_rules.md
//   writeAllTiers()     ??writes all files at once

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

// ?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”??
// TIER 1: GEMINI.md Bootstrap (~500 tokens)
// ?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”??

func emitBootstrap(result SubsumptionResult, brainRoot string) string {
	var sb strings.Builder

	sb.WriteString("<!-- NEURONFS:START -->\n")
	sb.WriteString(fmt.Sprintf("<!-- Generated: %s -->\n", time.Now().Format("2006-01-02T15:04:05")))
	sb.WriteString("<!-- Axiom: Folder=Neuron | File=Trace | Path=Sentence -->\n")
	sb.WriteString(fmt.Sprintf("<!-- Active: %d/%d neurons | Total activation: %d -->\n\n",
		result.FiredNeurons, result.TotalNeurons, result.TotalCounter))

	if result.BombSource != "" {
		sb.WriteString(fmt.Sprintf("## ?š¨ CIRCUIT BREAKER: %s\n", result.BombSource))
		sb.WriteString("**ALL OPERATIONS HALTED. REPAIR REQUIRED.**\n\n")
		sb.WriteString("<!-- NEURONFS:END -->\n")
		return sb.String()
	}

	sb.WriteString("## NeuronFS Active Rules\n\n")

	// ?â”??PERSONA (from ego region neurons ??not hardcoded) ?â”??
	sb.WriteString("### ?­ ?˜ë¥´?Œë‚˜\n")
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

	// ?â”??SUBSUMPTION (1-liner) ?â”??
	sb.WriteString("### ?”— Subsumption Cascade\n")
	sb.WriteString("```\nbrainstem ?â†’ limbic ?â†’ hippocampus ?â†’ sensors ?â†’ cortex ?â†’ ego ?â†’ prefrontal\n  (P0)         (P1)       (P2)          (P3)       (P4)     (P5)      (P6)\n```\n")
	sb.WriteString("??? Pê°€ ?’ì? Pë¥???ƒ ?°ì„ . bomb?€ ?„ì²´ ?•ì?.\n\n")

	// ?â”??BRAINSTEM TOP 5 (always in context ??conscience) ?â”??
	sb.WriteString("### ???µì‹¬ì§€ì¹?TOP 5\n")
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

	// ?â”??GROWTH PROTOCOL (compact) ?â”??
	sb.WriteString("### ?Œ± ?ê? ?±ì¥ ?„ë¡œ? ì½œ\n\n")
	sb.WriteString("**??ê·œì¹™?€ AIê°€ ë°˜ë“œ???°ë¼???˜ëŠ” ?±ì¥ ë©”ì»¤?ˆì¦˜?´ë‹¤.**\n\n")

	inboxPath := filepath.Join(brainRoot, "_inbox", "corrections.jsonl")

	sb.WriteString("#### êµì • ê°ì? ??_inbox??ê¸°ë¡\n")
	sb.WriteString("?¬ìš©?ê? ?¤ìˆ˜ë¥?êµì •?˜ë©´ ì¦‰ì‹œ `run_command`ë¡??¤í–‰:\n")
	sb.WriteString(fmt.Sprintf("```\n[IO.File]::AppendAllText(\"%s\", '{\"type\":\"correction\",\"path\":\"cortex/[ì¹´í…Œê³ ë¦¬]/[ê·œì¹™ëª?\",\"text\":\"êµì • ?¬ìœ \",\"counter_add\":1}' + \"`n\")\n```\n\n", inboxPath))

	sb.WriteString("#### ì¹?°¬ ê°ì? ??_inbox???„íŒŒë¯?ê¸°ë¡\n")
	sb.WriteString(fmt.Sprintf("```\n[IO.File]::AppendAllText(\"%s\", '{\"type\":\"correction\",\"path\":\"[ê¸°ì¡´?´ëŸ°ê²½ë¡œ]\",\"text\":\"ì¹?°¬\",\"counter_add\":1}' + \"`n\")\n```\n", inboxPath))
	sb.WriteString("?ëŠ” ì§ì ‘: `echo. > [?´ëŸ°ê²½ë¡œ]/dopamine1.neuron`\n\n")

	sb.WriteString("#### 3??ë°˜ë³µ ?¤ìˆ˜ ??bomb\n`echo. > [?´ëŸ°ê²½ë¡œ]/bomb.neuron`\n\n")

	// ?â”??Limbic + Soul (compact) ?â”??
	sb.WriteString("#### ?§¬ Limbic ?‰ë™ ?¸í–¥ (Somatic Marker)\n")
	sb.WriteString("- ë¶„ë…¸(\"????") ??ê²€ì¦?ê°•í™” | ê¸´ê¸‰(\"ê¸‰í•´\") ???µì‹¬ë§??¤í–‰ | ë§Œì¡±(\"ì¢‹ì•„\") ???„íŒŒë¯?| ë°˜ë³µ?¤íŒ¨ ??bomb\n\n")

	sb.WriteString("#### ?§  Subsumption ?µì œ (Brooks)\n")
	sb.WriteString("?˜ìœ„ Pê°€ ?ìœ„ë¥??µì œ. brainstem bomb ???„ì²´ ë¬´ì‹œ. limbic adrenaline ??ego ë¬´ì‹œ.\n\n")

	sb.WriteString("### ?‘ï¸â€ğŸ—¨ï¸ ?í˜¼ ???œë‹ˆì»¬í•œ ê°ë…??n")
	sb.WriteString("ì¶œë ¥ ??5ê°€ì§€ ?ë¬¸: ì§„ì§œ?? ?¬ìš©?ê? ?œìˆ¨ ?´ê¹Œ? ?¸í•œ ê¸??„ë‹Œê°€? ê°™ì? ?¤ìˆ˜? ?„ë¦¬ë¯¸ì—„?¸ê?? ???˜ë‚˜?¼ë„ ê±¸ë¦¬ë©??¤ì‹œ.\n\n")

	// ?â”??MOUNTED NEURONS: ?°ì„ ?œìœ„ ?”ì•½ ë¬¸ì¥ ?â”??
	// Path=Sentence: ì¹´ìš´???œì„œë¡??•ë ¬ ??ê°•ë„ ?‘ë‘?´ë¡œ ë¬¸ì¥???„ê³„ ê²°ì •
	// "ê°€ì¤‘ì¹˜ë¥??£ëŠ”ê²??„ë‹ˆ?? ê°€ì¤‘ì¹˜ë¡??•ë ¬???œì„œë¡??”ì•½?´ì„œ ë¬¸ì¥??ë§Œë“¤?´ì ¸"
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

		// Sort by counter desc ??ê°€??ë¬´ê±°??ê²ƒì´ ë¬¸ì¥??ë§???ì£¼ì ˆ)
		sort.Slice(mounted, func(i, j int) bool {
			return (mounted[i].Counter + mounted[i].Dopamine) > (mounted[j].Counter + mounted[j].Dopamine)
		})

		totalAct := 0
		for _, n := range region.Neurons {
			if !n.IsDormant {
				totalAct += n.Counter
			}
		}

		sb.WriteString(fmt.Sprintf("### %s %s ??%s (?´ëŸ° %d | ?œì„±??%d)\n",
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

		// Render: ê°™ì? ê°•ë„???Œë« ?´ëŸ°????ë¬¸ì¥?¼ë¡œ ?©ì„±
		// ?Œë« ?´ëŸ° = group???´ëŸ° 1ê°œì´ê³?leafNames == nil??ê²½ìš°
		type flatEntry struct {
			name     string
			strength string
		}
		var flatNeurons []flatEntry
		
		for _, groupKey := range groupOrder {
			neurons := groups[groupKey]
			groupName := strings.ReplaceAll(groupKey, "_", " ")
			for hanja, korean := range hanjaToKorean {
				groupName = strings.ReplaceAll(groupName, hanja, korean)
			}

			// ê°•ë„: ê·¸ë£¹ ??ìµœë? ì¹´ìš´??ê¸°ì?
			maxIntensity := 0
			hasKanjiOpcode := false  // ?œì ë§ˆì´?¬ë¡œ?µì½”??ê°ì?
			for _, n := range neurons {
				if v := n.Counter + n.Dopamine; v > maxIntensity {
					maxIntensity = v
				}
				// ç¦?å¿???è­??ëŠ” ?œêµ­???±ê?(ê¸ˆì?/ë°˜ë“œ??ì¶”ì²œ/ê²½ê³ )ê°€ ?´ë? ê°•ë„ë¥??œí˜„?˜ë?ë¡??‘ë‘??ë¶ˆí•„??
				if strings.ContainsAny(n.Path, "ç¦å¿…?¨è?") || strings.Contains(n.Path, "ê¸ˆì?") || strings.Contains(n.Path, "?ˆë?ë¡?) {
					hasKanjiOpcode = true
				}
			}
			// ê·¸ë£¹ëª…ì— ?œì ?ëŠ” ?œêµ­???¤ì›Œ?œê? ?¬í•¨?˜ì–´ ?ˆìœ¼ë©??™ì¼
			if strings.ContainsAny(groupKey, "ç¦å¿…?¨è?") || strings.Contains(groupName, "ê¸ˆì?:") || strings.Contains(groupName, "ë°˜ë“œ??) {
				hasKanjiOpcode = true
			}
			strength := ""
			if !hasKanjiOpcode {
				if maxIntensity >= 10 {
					strength = "?µì‹¬: "
				} else if maxIntensity >= 5 {
					strength = "ì¤‘ìš”: "
				}
			}

			// ?´ëŸ°?¤ì˜ ë¦¬í”„ ?´ë¦„ ?˜ì§‘ (?™ì–´ë°˜ë³µ ?œê±°)
			var leafNames []string
			isOnlyFlat := len(neurons) == 1 // ê·¸ë£¹???´ëŸ°??1ê°œë¿??ê²½ìš°ë§??Œë«
			for _, n := range neurons {
				parts := splitNeuronPath(n.Path)
				leaf := strings.ReplaceAll(parts[len(parts)-1], "_", " ")
				for hanja, korean := range hanjaToKorean {
					leaf = strings.ReplaceAll(leaf, hanja, korean)
				}

				// ?™ì–´ë°˜ë³µ ë°©ì?: ê·¸ë£¹ëª…ê³¼ ë¦¬í”„ê°€ ê°™ì? ?´ëŸ°?€ ?¤í‚µ
				if leaf == groupName {
					if len(parts) == 1 && isOnlyFlat {
						// ì§„ì§œ ?Œë« ?´ëŸ° (?˜ìœ„ ?†ìŒ): ë°°ì¹˜ ?˜ì§‘
						leafNames = nil
						break
					}
					continue // ?˜ìœ„ ?´ëŸ°???ˆìœ¼ë¯€ë¡?ì¹´í…Œê³ ë¦¬ ?ì²´??ê±´ë„ˆ?€
				}

				signals := ""
				if n.Dopamine > 0 {
					signals += " ?Ÿ¢"
				}
				if n.HasBomb {
					signals += " ?’£"
				}
				leafNames = append(leafNames, leaf+signals)

				if (n.Counter + n.Dopamine) >= 10 {
					topAnchors = append(topAnchors, fmt.Sprintf("%s > %s", groupName, leaf))
				}
			}

			if leafNames == nil {
				// ?Œë« ?´ëŸ°: ë°°ì¹˜ë¡?ëª¨ì•„???˜ì¤‘????ì¤„ë¡œ ì¶œë ¥
				flatNeurons = append(flatNeurons, flatEntry{name: groupName, strength: strength})
			} else if len(leafNames) == 0 {
				continue
			} else if len(leafNames) <= 5 {
				sb.WriteString(fmt.Sprintf("%s%s: %s.\n", strength, groupName, strings.Join(leafNames, ", ")))
			} else {
				// ê¸?ëª©ë¡: 5ê°œì”© ì¤„ë°”ê¿?
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
		
		// ?Œë« ?´ëŸ°: ê°™ì? ê°•ë„?¼ë¦¬ ??ë¬¸ì¥?¼ë¡œ ?©ì„±
		if len(flatNeurons) > 0 {
			batchMap := make(map[string][]string)
			batchOrder := []string{}
			for _, f := range flatNeurons {
				if _, exists := batchMap[f.strength]; !exists {
					batchOrder = append(batchOrder, f.strength)
				}
				batchMap[f.strength] = append(batchMap[f.strength], f.name)
			}
			for _, s := range batchOrder {
				names := batchMap[s]
				if len(names) <= 7 {
					sb.WriteString(fmt.Sprintf("%s%s.\n", s, strings.Join(names, ", ")))
				} else {
					sb.WriteString(fmt.Sprintf("%s%s", s, names[0]))
					for i := 1; i < len(names); i++ {
						if i%7 == 0 {
							sb.WriteString(fmt.Sprintf(".\n(cont): %s", names[i]))
						} else {
							sb.WriteString(fmt.Sprintf(", %s", names[i]))
						}
					}
					sb.WriteString(".\n")
				}
			}
		}
		sb.WriteString("\n")
	}


	// NOTE: Sandbox rules are NOT injected into GEMINI.md.
	// They are read via /api/sandbox GET (or "?Œë“œë°•ìŠ¤ ?•ì¸" trigger).

	// ?â”??ANCHOR: Repeat top rules at bottom (Lost in the Middle ?Œí”¼) ?â”??
	// Group anchors by category ??prose sentence per group
	if len(topAnchors) > 0 {
		sb.WriteString("### ? ï¸ ë¦¬ë§ˆ?¸ë” (?ˆë? ê·œì¹™)\n")
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

	// ?â”??MODE SWITCH: ?‘ì—… ê°ì? ???´ë‹¹ ?ì—­ _rules.md ë¨¼ì? ?½ê¸° ?â”??
	sb.WriteString("### ?§  ?‘ì—… ëª¨ë“œ ?„í™˜ (?„ìˆ˜)\n\n")
	sb.WriteString("**?‘ì—… ?œì‘ ???´ë‹¹ ?ì—­??`_rules.md`ë¥?`view_file`ë¡?ë°˜ë“œ??ë¨¼ì? ?½ëŠ”??**\n\n")
	sb.WriteString("| ?‘ì—… ê°ì? | ?½ì„ ?Œì¼ |\n|-----------|----------|\n")
	sb.WriteString(fmt.Sprintf("| CSS/?”ì??UI | `%s\\cortex\\_rules.md` |\n", brainRoot))
	sb.WriteString(fmt.Sprintf("| ë°±ì—”??API/DB | `%s\\cortex\\_rules.md` |\n", brainRoot))
	sb.WriteString(fmt.Sprintf("| NAS/?Œì¼ë³µì‚¬ | `%s\\sensors\\_rules.md` |\n", brainRoot))
	sb.WriteString(fmt.Sprintf("| ë¸Œëœ??ë§ˆì???| `%s\\sensors\\_rules.md` |\n", brainRoot))
	sb.WriteString(fmt.Sprintf("| ?„ë¡œ?íŠ¸ ë°©í–¥ | `%s\\prefrontal\\_rules.md` |\n", brainRoot))
	sb.WriteString(fmt.Sprintf("| NeuronFS ?ì²´ | `%s\\cortex\\_rules.md` |\n", brainRoot))
	sb.WriteString(fmt.Sprintf("\n??ê²½ë¡œ: `%s`\n\n", brainRoot))

	// ?â”??AGENT INBOX: ?ì´?„íŠ¸ ê°??Œí†µ (?¸ì ??ê¸°ë°˜) ?â”??
	agentInbox := emitAgentInbox(brainRoot)
	if agentInbox != "" {
		sb.WriteString(agentInbox)
	}

	sb.WriteString("<!-- NEURONFS:END -->\n")
	return sb.String()
}

// ?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”??
// AGENT INBOX: ?ì´?„íŠ¸ ê°??Œí†µ (?¸ì ??ê¸°ë°˜)
// _agents/<name>/inbox/ ?¤ìº” ??GEMINI.md???”ì•½ ?½ì…
// ?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”??

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

		// ?œìŠ¤???”ë ‰? ë¦¬ ?¤í‚µ
		if agentName == "scripts" || agentName == "pm" || strings.HasPrefix(agentName, ".") {
			continue
		}

		inboxDir := filepath.Join(agentsDir, agentName, "inbox")
		inboxFiles, err := os.ReadDir(inboxDir)
		if err != nil {
			continue
		}

		// ?€?€ [ë³¼ë¥¨ ?¬ì¸???„í‚¤?ì²˜ (Volume Pointer Architecture)] ?€?€
		// ê°œë³„ ë©”ì‹œì§€ ?¤í”„ë¥??†ì•  ?„ë¡¬?„íŠ¸ ê¸¸ì´ë¥?O(1) ê³ ì •?œí‚¤ê³?
		// ?ì´?„íŠ¸ ?¤ìŠ¤ë¡?list_dir ?„êµ¬ë¥??¨ì„œ ?´ë”ë¥??ìƒ‰?˜ë„ë¡?? ë„.
		count := 0
		for _, f := range inboxFiles {
			if !f.IsDir() && strings.HasSuffix(f.Name(), ".md") && !strings.HasPrefix(f.Name(), "_") {
				count++
			}
		}

		if count > 0 {
			if !hasMessages {
				sb.WriteString("### ?“¬ ?ì´?„íŠ¸ ?˜ì‹ ??(ë³¼ë¥¨ ?¬ì¸??\n\n")
				sb.WriteString("> **AI ì§€ì¹?** ? í° ë³´í˜¸ë¥??„í•´ ?˜ì‹ ??ëª©ë¡???ëµ?˜ì—ˆ?µë‹ˆ?? ë³¸ì¸(Agent)???´ë¦„???¸ëª…?˜ì—ˆ?¤ë©´, ?„êµ¬(`list_dir` ?¹ì? `run_command` Terminal)ë¥??¬ìš©?˜ì—¬ ì§€?•ëœ ?”—?´ë”ë¥?ë¦¬ìŠ¤?…í•˜ê³?ìµœì‹  ë©”ì‹œì§€ ?´ìš©??ì§ì ‘ ?Œì•…?˜ì„¸??\n\n")
				hasMessages = true
			}
			sb.WriteString(fmt.Sprintf("- **[%s]** ë¯¸í™•??ë©”ì‹œì§€: %dê±?n  - ?”— ë³¼ë¥¨ ?´ë”: `%s`\n\n", agentName, count, inboxDir))
		}
	}

	return sb.String()
}



// ?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”??
// TIER 2: _index.md ??Brain overview
// ?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”??

func emitIndex(brain Brain, result SubsumptionResult) string {
	var sb strings.Builder

	sb.WriteString("# ?§  NeuronFS Brain Index\n\n")
	sb.WriteString(fmt.Sprintf("Generated: %s | Neurons: %d/%d | Activation: %d\n\n",
		time.Now().Format("2006-01-02T15:04:05"),
		result.FiredNeurons, result.TotalNeurons, result.TotalCounter))

	if result.BombSource != "" {
		sb.WriteString(fmt.Sprintf("## ?š¨ BOMB: %s ??ALL HALTED\n\n", result.BombSource))
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
		sb.WriteString("## ?•¸ï¸?Axon ?°ê²°\n")
		for _, region := range result.ActiveRegions {
			icon := regionIcons[region.Name]
			for _, axon := range region.Axons {
				if strings.HasPrefix(axon, "SKILL:") {
					skillName := filepath.Base(filepath.Dir(strings.TrimPrefix(axon, "SKILL:")))
					sb.WriteString(fmt.Sprintf("- %s %s ???”§ %s\n", icon, region.Name, skillName))
				} else {
					targetIcon := regionIcons[axon]
					if targetIcon == "" {
						targetIcon = "?”—"
					}
					sb.WriteString(fmt.Sprintf("- %s %s ??%s %s\n", icon, region.Name, targetIcon, axon))
				}
			}
		}
		sb.WriteString("\n")
	}

	// TOP 10 global
	allNeurons := collectAllNeurons(result)
	sb.WriteString("## ?† TOP 10 ?´ëŸ°\n")
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
		sb.WriteString(fmt.Sprintf("<summary>?†• ? ê·œ (probation) ??%d neurons (%dd window)</summary>\n\n", len(spotlight), spotlightDays))

		// Group by region in P0?’P6 order
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
				sb.WriteString("(?†ìŒ)\n\n")
				continue
			}
			for _, rn := range neurons {
				ageDays := int(now.Sub(rn.neuron.ModTime).Hours() / 24)
				sb.WriteString(fmt.Sprintf("- **%s** (%d) ??%dd?¨ìŒ\n",
					pathToSentence(rn.neuron.Path), rn.neuron.Counter, spotlightDays-ageDays))
			}
			sb.WriteString("\n")
		}

		sb.WriteString("</details>\n\n")
	}

	// Per-region summary table
	sb.WriteString("## ?“Š ?ì—­ë³??„í™©\n\n")
	sb.WriteString("| ?ì—­ | ?´ëŸ° | ?œì„±??| ?ì„¸ |\n")
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
		sb.WriteString(fmt.Sprintf("| %s %s ??%s | %d | %d | `%s/_rules.md` |\n",
			icon, region.Name, ko, count, activation, region.Name))
	}
	sb.WriteString("\n")

	return sb.String()
}

// ?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”??
// TIER 3: {region}/_rules.md ??Tree-compressed detail
// ?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”??

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

	sb.WriteString(fmt.Sprintf("# %s %s ??%s\n\n", icon, strings.ToUpper(region.Name), ko))

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
			sb.WriteString(fmt.Sprintf("- ??%s\n", axon))
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
		for hanja, korean := range hanjaToKorean {
			name = strings.ReplaceAll(name, hanja, korean)
		}

		if n.isLeaf && len(n.children) == 0 {
			// Pure leaf ??show with counter + intensity prefix
			signals := ""
			if n.dopamine > 0 {
				signals += " ?Ÿ¢"
			}
			if n.hasBomb {
				signals += " ?’£"
			}
			strength := ""
			if n.counter >= 10 {
				strength = "?ˆë? "
			} else if n.counter >= 5 {
				strength = "ë°˜ë“œ??"
			}
			sb.WriteString(fmt.Sprintf("%s- %s**%s** (%d)%s\n", indent, strength, name, n.counter, signals))
		} else if n.isLeaf && len(n.children) > 0 {
			// Leaf but also a branch ??show counter then children
			signals := ""
			if n.dopamine > 0 {
				signals += " ?Ÿ¢"
			}
			strength := ""
			if n.counter >= 10 {
				strength = "?ˆë? "
			} else if n.counter >= 5 {
				strength = "ë°˜ë“œ??"
			}
			sb.WriteString(fmt.Sprintf("%s- %s**%s** (%d)%s\n", indent, strength, name, n.counter, signals))
			renderTree(sb, n, depth+1, prefix+child.key+"/")
		} else {
			// Pure branch ??show as group header
			sb.WriteString(fmt.Sprintf("%s- %s/\n", indent, name))
			renderTree(sb, n, depth+1, prefix+child.key+"/")
		}
	}
}

// ?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”??
// WRITE ALL TIERS
// ?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”??

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

	fmt.Printf("[SYNC] ?»ï¸  3-tier emit complete: GEMINI.md + _index.md + 7x _rules.md (%d neurons, activation: %d)\n",
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

// ?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”??
// EMIT TARGETS ??Multi-editor support
// ?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”??

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

		fmt.Printf("[EMIT] ??%s ??%s\n", et.Name, targetPath)
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

	fmt.Printf("[SYNC] ?»ï¸  emit complete: %d target(s) + _index.md + 7x _rules.md (%d neurons, activation: %d)\n",
		len(targets), result.FiredNeurons, result.TotalCounter)
}

// doInjectToFile injects NeuronFS content into an existing file, preserving surrounding content
func doInjectToFile(filePath string, rules string) {
	existing, err := os.ReadFile(filePath)
	if err != nil {
		// File doesn't exist ??create with just the rules
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
		if after != "" {
			content = content[:startIdx] + rules + "\n" + after
		} else {
			content = content[:startIdx] + rules
		}
	} else {
		content = rules + "\n\n" + content
	}

	os.WriteFile(filePath, []byte(content), 0644)
}

// ?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”??
// READ = FIRE: API endpoint that reads + auto-activates
// ?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”??

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

// ?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”??
// HELPERS
// ?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”?â”??

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

// hanjaToKorean ?œì ë§ˆì´?¬ë¡œ?µì½”?????œêµ­???ì—°??ë³€??
// ?”ìŠ¤?¬ì—???œì 1ê¸€?ë¡œ ?•ì¶•, AI ì£¼ì… ???œêµ­?´ë¡œ ?€?´ì„œ ?„ë‹¬
var hanjaToKorean = map[string]string{
	"ç¦?: "?ˆë? ê¸ˆì?: ",  // ?„ìˆ˜ ë¶€????~?˜ì? ë§ˆë¼
	"å¿?: "ë°˜ë“œ??",  // ?„ìˆ˜ ê¸ì • ??~?´ë¼
	"??: "ì¶”ì²œ: ",   // ê¶Œì¥ ??~?˜ëŠ” ê²?ì¢‹ë‹¤
	"è­?: "ê²½ê³ : ",   // ì£¼ì˜ ??~?˜ë©´ ?„í—˜
}

// pathToSentence converts path to readable sentence
// "frontend\css\glass_blur20" ??"frontend > css > glass blur20"
// ?œì prefix???œêµ­?´ë¡œ ?ë™ ë³€??
func pathToSentence(p string) string {
	s := strings.ReplaceAll(p, string(filepath.Separator), " > ")
	s = strings.ReplaceAll(s, "/", " > ")
	s = strings.ReplaceAll(s, "_", " ")
	// ?œì?’í•œêµ?–´ ë³€??
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

