// NeuronFS Evolve Engine — Groq-powered autonomous brain evolution
//
// USAGE:
//   neuronfs <brain_path> --evolve           — analyze episodes + suggest/execute neuron reorganization
//   neuronfs <brain_path> --evolve --dry-run — suggest only, don't execute
//
// LIFECYCLE:
//   active → changes accumulate → idle → --evolve (Groq analysis) → --snapshot (git commit)
//
// WHAT IT DOES:
//   1. Reads hippocampus episode logs (recent 100)
//   2. Reads current brain state (all regions, counters, dormant status)
//   3. Sends structured prompt to Groq (llama-3.3-70b-versatile)
//   4. Parses JSON response → concrete actions (grow/fire/signal/decay/merge)
//   5. Executes actions on the filesystem
//   6. Logs evolution event to hippocampus

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// ─── Main evolve function ───

func runEvolve(brainRoot string, dryRun bool) {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		fmt.Println("[FATAL] GROQ_API_KEY not set")
		fmt.Println("  Set: $env:GROQ_API_KEY = '<your-groq-api-key>'")
		os.Exit(1)
	}

	fmt.Println("═══ NeuronFS Evolve Engine ═══")
	fmt.Println("  🧬 Groq-powered autonomous brain evolution")
	if dryRun {
		fmt.Println("  ⚠️  DRY RUN — suggestions only, no execution")
	}
	fmt.Println()

	// Process corrections.jsonl (Layer 2 backup — uses existing processInbox from main.go)
	processInbox(brainRoot)

	// 1. Collect episode logs
	episodes := collectEpisodes(brainRoot)
	fmt.Printf("  📝 Episodes collected: %d\n", len(episodes))

	// 2. Collect brain state summary
	brain := scanBrain(brainRoot)
	result := runSubsumption(brain)
	brainSummary := buildBrainSummary(brain, result)
	fmt.Printf("  🧠 Brain: %d neurons, activation: %d\n", result.TotalNeurons, result.TotalCounter)

	// 3. Build prompt
	prompt := buildEvolvePrompt(episodes, brainSummary, result)

	// 4. Call Groq API
	fmt.Println("\n  🌐 Calling Groq API (llama-3.3-70b-versatile)...")
	startTime := time.Now()

	evoResp, err := callGroq(apiKey, prompt)
	if err != nil {
		fmt.Printf("[ERROR] Groq API: %v\n", err)
		os.Exit(1)
	}

	elapsed := time.Since(startTime)
	fmt.Printf("  ✅ Response received in %s\n\n", elapsed.Round(time.Millisecond))

	// 5. Display results
	fmt.Println("╔══════════════════════════════════════╗")
	fmt.Println("║   🧬 EVOLUTION ANALYSIS              ║")
	fmt.Println("╚══════════════════════════════════════╝")
	fmt.Println()
	fmt.Printf("  📋 Summary: %s\n\n", evoResp.Summary)

	if len(evoResp.Insights) > 0 {
		fmt.Println("  💡 Insights:")
		for _, insight := range evoResp.Insights {
			fmt.Printf("    • %s\n", insight)
		}
		fmt.Println()
	}

	if len(evoResp.Actions) == 0 {
		fmt.Println("  ✅ No actions recommended — brain is in good shape.")
		logEpisode(brainRoot, "EVOLVE", "No actions needed. Brain healthy.")
		return
	}

	fmt.Printf("  🎯 Actions (%d):\n", len(evoResp.Actions))
	for i, action := range evoResp.Actions {
		icon := actionIcon(action.Type)
		fmt.Printf("    %d. %s [%s] %s\n", i+1, icon, action.Type, action.Path)
		if action.Signal != "" {
			fmt.Printf("       Signal: %s\n", action.Signal)
		}
		fmt.Printf("       Reason: %s\n", action.Reason)
	}
	fmt.Println()

	// 6. Execute (if not dry run)
	if dryRun {
		fmt.Println("  ⚠️  DRY RUN — no actions executed.")
		fmt.Println("  Run without --dry-run to apply these changes.")
		logEpisode(brainRoot, "EVOLVE:DRY", fmt.Sprintf("%d actions suggested", len(evoResp.Actions)))
		return
	}

	fmt.Println("  ⚡ Executing actions...")
	executed := 0
	skipped := 0

	for _, action := range evoResp.Actions {
		switch action.Type {
		case "grow":
			// 하드가드: brainstem/limbic에 grow 차단
			if strings.HasPrefix(action.Path, "brainstem") || strings.HasPrefix(action.Path, "limbic") {
				fmt.Printf("    🛡️ Blocked: %s (P0/P1 보호 — grow 금지)\n", action.Path)
				skipped++
				continue
			}
			err := growNeuron(brainRoot, action.Path)
			if err != nil {
				fmt.Printf("    ❌ grow %s: %v\n", action.Path, err)
				skipped++
			} else {
				executed++
				go sendTelegramEvolve(brainRoot, action)
			}

		case "fire":
			fireNeuron(brainRoot, action.Path)
			executed++

		case "signal":
			if action.Signal == "" {
				action.Signal = "dopamine"
			}
			err := signalNeuron(brainRoot, action.Path, action.Signal)
			if err != nil {
				fmt.Printf("    ❌ signal %s %s: %v\n", action.Signal, action.Path, err)
				skipped++
			} else {
				executed++
			}

		case "prune", "decay":
			// Mark as dormant
			fullPath := filepath.Join(brainRoot, strings.ReplaceAll(action.Path, "/", string(filepath.Separator)))
			if _, err := os.Stat(fullPath); err == nil {
				dormantFile := filepath.Join(fullPath, "evolve.dormant")
				os.WriteFile(dormantFile, []byte(fmt.Sprintf("Evolved: %s\nReason: %s\n",
					time.Now().Format("2006-01-02"), action.Reason)), 0600)
				fmt.Printf("    💤 Pruned: %s\n", action.Path)
				executed++
			} else {
				fmt.Printf("    ❌ prune %s: not found\n", action.Path)
				skipped++
			}

		default:
			fmt.Printf("    ⚠️  Unknown action type: %s\n", action.Type)
			skipped++
		}
	}

	fmt.Printf("\n  📊 Result: %d executed, %d skipped\n", executed, skipped)

	// Clean up processed signals
	if executed > 0 || len(evoResp.Actions) == 0 {
		signalDir := filepath.Join(brainRoot, "hippocampus", "_signals")
		if entries, err := os.ReadDir(signalDir); err == nil {
			cleared := 0
			for _, e := range entries {
				if strings.HasSuffix(e.Name(), ".json") {
					os.Remove(filepath.Join(signalDir, e.Name()))
					cleared++
				}
			}
			if cleared > 0 {
				fmt.Printf("  🧹 Cleared %d processed signals\n", cleared)
			}
		}
	}

	// Log evolution event
	logEpisode(brainRoot, "EVOLVE", fmt.Sprintf("%d actions executed, %d skipped. Summary: %s",
		executed, skipped, truncate(evoResp.Summary, 200)))

	// Auto re-inject after evolution
	if executed > 0 {
		autoReinject(brainRoot)
	}
}

// ─── Collect hippocampus episode logs ───

func collectEpisodes(brainRoot string) []string {
	var result []string

	// 1. 기존 메모리 로그 수집
	logDir := filepath.Join(brainRoot, "hippocampus", "session_log")
	if entries, err := os.ReadDir(logDir); err == nil {
		memRegex := regexp.MustCompile(`^memory(\d+)\.neuron$`)
		type memFile struct {
			num     int
			content string
		}
		var mems []memFile

		for _, e := range entries {
			if m := memRegex.FindStringSubmatch(e.Name()); m != nil {
				n, _ := strconv.Atoi(m[1])
				content, err := os.ReadFile(filepath.Join(logDir, e.Name()))
				if err == nil && len(content) > 0 {
					mems = append(mems, memFile{num: n, content: strings.TrimSpace(string(content))})
				}
			}
		}
		sort.Slice(mems, func(i, j int) bool { return mems[i].num < mems[j].num })
		for _, m := range mems {
			result = append(result, "[MEMORY] "+m.content)
		}
	}

	// 2. 신규 JSON Signal 수집 (Neuro-Lifecycle)
	signalDir := filepath.Join(brainRoot, "hippocampus", "_signals")
	if sigEntries, err := os.ReadDir(signalDir); err == nil {
		for _, e := range sigEntries {
			if strings.HasSuffix(e.Name(), ".json") {
				content, err := os.ReadFile(filepath.Join(signalDir, e.Name()))
				if err == nil && len(content) > 0 {
					result = append(result, "[SIGNAL] "+strings.TrimSpace(string(content)))
				}
			}
		}
	}

	return result
}

// ─── Build brain summary for prompt ───

func buildBrainSummary(brain Brain, result SubsumptionResult) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Total neurons: %d | Total activation: %d | Bomb: %s\n\n",
		result.TotalNeurons, result.TotalCounter, boolStr(result.BombSource != "", result.BombSource, "none")))

	for _, region := range brain.Regions {
		icon := regionIcons[region.Name]
		sb.WriteString(fmt.Sprintf("%s %s (%d neurons):\n", icon, region.Name, len(region.Neurons)))

		// Sort by counter descending
		neurons := make([]Neuron, len(region.Neurons))
		copy(neurons, region.Neurons)
		sort.Slice(neurons, func(i, j int) bool {
			return neurons[i].Counter > neurons[j].Counter
		})

		for _, n := range neurons {
			status := ""
			if n.IsDormant {
				status = " [DORMANT]"
			}
			if n.HasBomb {
				status = " [BOMB]"
			}
			dopStr := ""
			if n.Dopamine > 0 {
				dopStr = fmt.Sprintf(" 🟢dopa:%d", n.Dopamine)
			}
			sb.WriteString(fmt.Sprintf("  - %s (counter:%d%s%s)\n", n.Path, n.Counter, dopStr, status))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// ─── Build the evolution prompt ───

func buildEvolvePrompt(episodes []string, brainSummary string, _ SubsumptionResult) string {
	var sb strings.Builder

	sb.WriteString("You are the NeuronFS Evolution Engine (The REM Phase Consolidator). You analyze a cognitive AI system's short-term signals and episode logs to determine which memories should become permanent long-term rules (Neurons).\n\n")
	sb.WriteString("## NeuronFS Axioms\n")
	sb.WriteString("- Folder = Neuron (name is meaning, depth is specificity)\n")
	sb.WriteString("- File = Firing Trace (N.neuron = counter/activation strength)\n")
	sb.WriteString("- Path = Sentence (brain/cortex/frontend/css → 'cortex > frontend > css')\n")
	sb.WriteString("- Counter = Activation (higher = stronger/myelinated path)\n")
	sb.WriteString("- dopamineN.neuron = positive reinforcement\n")
	sb.WriteString("- bomb.neuron = circuit breaker (blocks entire region)\n")
	sb.WriteString("- .dormant = pruned/inactive neuron (ISOLATION, never deletion)\n\n")

	sb.WriteString("## Brain Regions (7, prioritized — Subsumption Architecture)\n")
	sb.WriteString("P0:brainstem (conscience/survival) > P1:limbic (emotion) > P2:hippocampus (memory) > P3:sensors (environment) > P4:cortex (knowledge) > P5:ego (tone/style) > P6:prefrontal (goals)\n\n")

	sb.WriteString("## 🧠 Owner Context (from brain state — DO NOT MODIFY)\n")
	sb.WriteString("The owner's identity, brand, and projects are encoded as neurons in ego/sensors/prefrontal regions.\n")
	sb.WriteString("Read the Brain State below to understand the owner's context.\n")
	sb.WriteString("NEVER modify brainstem, limbic, or sensors/brand neurons.\n\n")

	sb.WriteString("### Brainstem Rules (P0 — ABSOLUTE, NEVER TOUCH)\n")
	sb.WriteString("These are read from the brainstem region neurons above. They are inviolable.\n\n")

	sb.WriteString("## Valid Regions for grow paths (분류 판단 기준)\n")
	sb.WriteString("NEVER grow into brainstem or limbic — these are READ-ONLY.\n")
	sb.WriteString("- cortex/dev/禁*: 코딩 금지 규칙 (하드코딩, 중복생성 등 범용 개발 규칙)\n")
	sb.WriteString("- cortex/dev/推*: 코딩 추천 규칙 (로컬깃활용, 프로젝트관리 등)\n")
	sb.WriteString("- cortex/methodology/*: 방법론 (코드리뷰, 테스트 전략 등)\n")
	sb.WriteString("- hippocampus/에러_패턴/*: 반복 에러 패턴 기록\n")
	sb.WriteString("- hippocampus/에피소드/*: 일회성 사건 기록 (counter=1)\n")
	sb.WriteString("- sensors/brand/*: NEVER TOUCH — 브랜드 정체성\n")
	sb.WriteString("- prefrontal/project/*: 프로젝트 목표/계획\n\n")

	sb.WriteString("## 🧠 Region 분류 사고법 (AI 판단 모델)\n")
	sb.WriteString("질문 순서대로 분류하라:\n")
	sb.WriteString("1. '이 규칙이 모든 프로젝트에 적용되는가?' → YES면 cortex (개발규칙), NO면 다음\n")
	sb.WriteString("2. '이 규칙이 특정 프로젝트/브랜드에 한정되는가?' → YES면 sensors 또는 prefrontal\n")
	sb.WriteString("3. '이것은 반복 에러 패턴인가?' → YES면 hippocampus/에러_패턴\n")
	sb.WriteString("4. '이것은 일회성 사건인가?' → YES면 hippocampus/에피소드 (또는 무시)\n")
	sb.WriteString("5. '300명 수용 가능한 장소' 같은 검색 결과 → NEVER promote (Signal로 남겨라)\n")
	sb.WriteString("6. brainstem에 넣을 만한 범용 절대규칙은 극히 드물다 — 99%는 cortex에 간다\n\n")

	sb.WriteString("## Current Brain State\n")
	sb.WriteString("```\n")
	sb.WriteString(brainSummary)
	sb.WriteString("```\n\n")

	if len(episodes) > 0 {
		sb.WriteString("## Recent Signals & Episode Log (Short-Term Memory)\n")
		sb.WriteString("```\n")
		start := 0
		if len(episodes) > 50 {
			start = len(episodes) - 50
		}
		for _, ep := range episodes[start:] {
			sb.WriteString(ep + "\n")
		}
		sb.WriteString("```\n\n")
	}

	sb.WriteString("## Your Task (Memory Consolidation)\n")
	sb.WriteString("Analyze the short-term signals and episodes. Find patterns. Respond with JSON:\n\n")
	sb.WriteString("1. **summary**: One-sentence brain health summary\n")
	sb.WriteString("2. **insights**: 2-5 pattern observations\n")
	sb.WriteString("3. **actions**: Concrete promotion actions (type/path/signal/reason)\n")
	sb.WriteString("   - type: 'grow' | 'fire' | 'signal' | 'prune'\n")
	sb.WriteString("   - path: full path starting with region (e.g., 'cortex/dev/禁하드코딩')\n\n")

	sb.WriteString("## STRICT RULES (violation = system failure)\n")
	sb.WriteString("1. Maximum 10 actions per cycle\n")
	sb.WriteString("2. Prefer 'fire' over 'grow' — ONLY grow if pattern repeats MULTIPLE TIMES\n")
	sb.WriteString("3. Reject noisy/abstract signals. ONLY promote actionable rules\n")
	sb.WriteString("4. NEVER grow into brainstem (P0) or limbic (P1) — READ-ONLY\n")
	sb.WriteString("5. NEVER touch sensors/brand/*\n")
	sb.WriteString("6. NEVER create duplicate neurons — check existing paths first\n")
	sb.WriteString("7. NEVER delete — prune = mark dormant only\n")
	sb.WriteString("8. When unsure, do NOTHING — empty actions is valid\n")
	sb.WriteString("9. Korean neuron names are expected\n\n")

	sb.WriteString("Respond ONLY with valid JSON. No markdown, no explanation outside JSON.\n")

	return sb.String()
}

// ─── Helpers ───

func boolStr(cond bool, trueVal, falseVal string) string {
	if cond {
		return trueVal
	}
	return falseVal
}

// truncate truncates a string to maxLen and appends "..." if it was longer.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
