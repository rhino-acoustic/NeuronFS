// NeuronFS Evolve Engine ??Groq-powered autonomous brain evolution
//
// USAGE:
//   neuronfs <brain_path> --evolve           ??analyze episodes + suggest/execute neuron reorganization
//   neuronfs <brain_path> --evolve --dry-run ??suggest only, don't execute
//
// LIFECYCLE:
//   active ??changes accumulate ??idle ??--evolve (Groq analysis) ??--snapshot (git commit)
//
// WHAT IT DOES:
//   1. Reads hippocampus episode logs (recent 100)
//   2. Reads current brain state (all regions, counters, dormant status)
//   3. Sends structured prompt to Groq (llama-3.3-70b-versatile)
//   4. Parses JSON response ??concrete actions (grow/fire/signal/decay/merge)
//   5. Executes actions on the filesystem
//   6. Logs evolution event to hippocampus

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// ?€?€?€ Groq API Types ?€?€?€

type groqMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type groqRequest struct {
	Model       string         `json:"model"`
	Messages    []groqMessage  `json:"messages"`
	Temperature float64        `json:"temperature"`
	MaxTokens   int            `json:"max_tokens"`
	TopP        float64        `json:"top_p"`
	Stream      bool           `json:"stream"`
	ResponseFormat *groqResponseFormat `json:"response_format,omitempty"`
}

type groqResponseFormat struct {
	Type string `json:"type"`
}

type groqChoice struct {
	Message groqMessage `json:"message"`
}

type groqResponse struct {
	Choices []groqChoice `json:"choices"`
	Error   *groqError   `json:"error,omitempty"`
}

type groqError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

// ?€?€?€ Evolution Action Types ?€?€?€

type evoAction struct {
	Type   string `json:"type"`   // grow | fire | signal | decay | merge | prune
	Path   string `json:"path"`   // neuron path
	Signal string `json:"signal"` // for signal type: dopamine | bomb
	Reason string `json:"reason"` // why this action
}

type evoResult struct {
	Summary  string      `json:"summary"`
	Actions  []evoAction `json:"actions"`
	Insights []string    `json:"insights"`
}

// ?€?€?€ Main evolve function ?€?€?€

func runEvolve(brainRoot string, dryRun bool) {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		fmt.Println("[FATAL] GROQ_API_KEY not set")
		fmt.Println("  Set: $env:GROQ_API_KEY = 'gsk_...'")
		os.Exit(1)
	}

	fmt.Println("?â•??NeuronFS Evolve Engine ?â•??)
	fmt.Println("  ?§¬ Groq-powered autonomous brain evolution")
	if dryRun {
		fmt.Println("  ? ï¸  DRY RUN ??suggestions only, no execution")
	}
	fmt.Println()

	// Process corrections.jsonl (Layer 2 backup ??uses existing processInbox from main.go)
	processInbox(brainRoot)

	// 1. Collect episode logs
	episodes := collectEpisodes(brainRoot)
	fmt.Printf("  ?“ Episodes collected: %d\n", len(episodes))

	// 2. Collect brain state summary
	brain := scanBrain(brainRoot)
	result := runSubsumption(brain)
	brainSummary := buildBrainSummary(brain, result)
	fmt.Printf("  ?§  Brain: %d neurons, activation: %d\n", result.TotalNeurons, result.TotalCounter)

	// 3. Build prompt
	prompt := buildEvolvePrompt(episodes, brainSummary, result)

	// 4. Call Groq API
	fmt.Println("\n  ?Œ Calling Groq API (llama-3.3-70b-versatile)...")
	startTime := time.Now()

	evoResp, err := callGroq(apiKey, prompt)
	if err != nil {
		fmt.Printf("[ERROR] Groq API: %v\n", err)
		os.Exit(1)
	}

	elapsed := time.Since(startTime)
	fmt.Printf("  ??Response received in %s\n\n", elapsed.Round(time.Millisecond))

	// 5. Display results
	fmt.Println("?”â•?â•?â•?â•?â•?â•?â•?â•?â•?â•?â•?â•?â•?â•?â•?â•?â•?â•?â•?â•—")
	fmt.Println("??  ?§¬ EVOLUTION ANALYSIS              ??)
	fmt.Println("?šâ•?â•?â•?â•?â•?â•?â•?â•?â•?â•?â•?â•?â•?â•?â•?â•?â•?â•?â•?â•")
	fmt.Println()
	fmt.Printf("  ?“‹ Summary: %s\n\n", evoResp.Summary)

	if len(evoResp.Insights) > 0 {
		fmt.Println("  ?’¡ Insights:")
		for _, insight := range evoResp.Insights {
			fmt.Printf("    ??%s\n", insight)
		}
		fmt.Println()
	}

	if len(evoResp.Actions) == 0 {
		fmt.Println("  ??No actions recommended ??brain is in good shape.")
		logEpisode(brainRoot, "EVOLVE", "No actions needed. Brain healthy.")
		return
	}

	fmt.Printf("  ?Ž¯ Actions (%d):\n", len(evoResp.Actions))
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
		fmt.Println("  ? ï¸  DRY RUN ??no actions executed.")
		fmt.Println("  Run without --dry-run to apply these changes.")
		logEpisode(brainRoot, "EVOLVE:DRY", fmt.Sprintf("%d actions suggested", len(evoResp.Actions)))
		return
	}

	fmt.Println("  ??Executing actions...")
	executed := 0
	skipped := 0

	for _, action := range evoResp.Actions {
		switch action.Type {
		case "grow":
			err := growNeuron(brainRoot, action.Path)
			if err != nil {
				fmt.Printf("    ??grow %s: %v\n", action.Path, err)
				skipped++
			} else {
				executed++
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
				fmt.Printf("    ??signal %s %s: %v\n", action.Signal, action.Path, err)
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
					time.Now().Format("2006-01-02"), action.Reason)), 0644)
				fmt.Printf("    ?’¤ Pruned: %s\n", action.Path)
				executed++
			} else {
				fmt.Printf("    ??prune %s: not found\n", action.Path)
				skipped++
			}

		default:
			fmt.Printf("    ? ï¸  Unknown action type: %s\n", action.Type)
			skipped++
		}
	}

	fmt.Printf("\n  ?“Š Result: %d executed, %d skipped\n", executed, skipped)

	// Log evolution event
	logEpisode(brainRoot, "EVOLVE", fmt.Sprintf("%d actions executed, %d skipped. Summary: %s",
		executed, skipped, truncate(evoResp.Summary, 200)))

	// Auto re-inject after evolution
	if executed > 0 {
		autoReinject(brainRoot)
	}
}

// ?€?€?€ Collect hippocampus episode logs ?€?€?€

func collectEpisodes(brainRoot string) []string {
	logDir := filepath.Join(brainRoot, "hippocampus", "session_log")
	entries, err := os.ReadDir(logDir)
	if err != nil {
		return nil
	}

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

	// Sort by number (chronological)
	sort.Slice(mems, func(i, j int) bool { return mems[i].num < mems[j].num })

	var result []string
	for _, m := range mems {
		result = append(result, m.content)
	}
	return result
}

// ?€?€?€ Build brain summary for prompt ?€?€?€

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
				dopStr = fmt.Sprintf(" ?Ÿ¢dopa:%d", n.Dopamine)
			}
			sb.WriteString(fmt.Sprintf("  - %s (counter:%d%s%s)\n", n.Path, n.Counter, dopStr, status))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// ?€?€?€ Build the evolution prompt ?€?€?€

func buildEvolvePrompt(episodes []string, brainSummary string, _ SubsumptionResult) string {
	var sb strings.Builder

	sb.WriteString("You are the NeuronFS Evolution Engine. You analyze a cognitive AI system's episode logs and brain state to recommend structural improvements.\n\n")
	sb.WriteString("## NeuronFS Axioms\n")
	sb.WriteString("- Folder = Neuron (name is meaning, depth is specificity)\n")
	sb.WriteString("- File = Firing Trace (N.neuron = counter/activation strength)\n")
	sb.WriteString("- Path = Sentence (brain/cortex/frontend/css ??'cortex > frontend > css')\n")
	sb.WriteString("- Counter = Activation (higher = stronger/myelinated path)\n")
	sb.WriteString("- dopamineN.neuron = positive reinforcement\n")
	sb.WriteString("- bomb.neuron = circuit breaker (blocks entire region)\n")
	sb.WriteString("- .dormant = pruned/inactive neuron (ISOLATION, never deletion)\n\n")

	sb.WriteString("## Brain Regions (7, prioritized ??Subsumption Architecture)\n")
	sb.WriteString("P0:brainstem (conscience/survival) > P1:limbic (emotion) > P2:hippocampus (memory) > P3:sensors (environment) > P4:cortex (knowledge) > P5:ego (tone/style) > P6:prefrontal (goals)\n\n")

	// ?€?€ IDENTITY: loaded dynamically from ego + sensors regions ?€?€
	sb.WriteString("## ?§  Owner Context (from brain state ??DO NOT MODIFY)\n")
	sb.WriteString("The owner's identity, brand, and projects are encoded as neurons in ego/sensors/prefrontal regions.\n")
	sb.WriteString("Read the Brain State below to understand the owner's context.\n")
	sb.WriteString("NEVER modify brainstem, limbic, or sensors/brand neurons.\n\n")

	sb.WriteString("### Brainstem Rules (P0 ??ABSOLUTE, NEVER TOUCH)\n")
	sb.WriteString("These are read from the brainstem region neurons above. They are inviolable.\n\n")

	sb.WriteString("## Valid Regions for grow paths\n")
	sb.WriteString("brainstem, limbic, hippocampus, sensors, cortex, ego, prefrontal\n\n")

	sb.WriteString("## Current Brain State\n")
	sb.WriteString("```\n")
	sb.WriteString(brainSummary)
	sb.WriteString("```\n\n")

	if len(episodes) > 0 {
		sb.WriteString("## Recent Episode Log (chronological)\n")
		sb.WriteString("```\n")
		// Show last 50 episodes max to fit in context
		start := 0
		if len(episodes) > 50 {
			start = len(episodes) - 50
		}
		for _, ep := range episodes[start:] {
			sb.WriteString(ep + "\n")
		}
		sb.WriteString("```\n\n")
	}

	sb.WriteString("## Your Task\n")
	sb.WriteString("Analyze the brain state and episode logs. Respond with a JSON object containing:\n\n")
	sb.WriteString("1. **summary**: One-sentence summary of the brain's current health\n")
	sb.WriteString("2. **insights**: Array of 2-5 observations about patterns, problems, or opportunities\n")
	sb.WriteString("3. **actions**: Array of concrete actions to improve the brain. Each action has:\n")
	sb.WriteString("   - type: 'grow' | 'fire' | 'signal' | 'prune'\n")
	sb.WriteString("   - path: full neuron path starting with region (e.g., 'cortex/frontend/css/new_rule')\n")
	sb.WriteString("   - signal: only for type='signal', value is 'dopamine' or 'bomb'\n")
	sb.WriteString("   - reason: why this action improves the brain\n\n")

	sb.WriteString("## STRICT RULES (violation = system failure)\n")
	sb.WriteString("1. Maximum 10 actions per evolution cycle\n")
	sb.WriteString("2. Prefer 'fire' (reinforce existing) over 'grow' (create new) ??consolidation over expansion\n")
	sb.WriteString("3. NEVER touch brainstem neurons (P0 is read-only conscience) ??not grow, not prune, not signal\n")
	sb.WriteString("4. NEVER touch limbic neurons (P1 emotion system is automatic)\n")
	sb.WriteString("5. NEVER touch sensors/brand/* (owner's brand identity is sacred)\n")
	sb.WriteString("6. NEVER create duplicate neurons ??check existing paths first\n")
	sb.WriteString("7. NEVER delete ??prune means mark dormant (isolation), NOT deletion\n")
	sb.WriteString("8. 'prune' ONLY neurons with counter=1 AND no dopamine AND overlap with higher-counter neurons\n")
	sb.WriteString("9. 'signal dopamine' neurons that have been consistently useful (frequent FIRE in logs)\n")
	sb.WriteString("10. Paths must use / separator and start with a valid region name\n")
	sb.WriteString("11. When unsure, do NOTHING ??empty actions array is perfectly valid\n")
	sb.WriteString("12. Korean neuron names are fine and expected. Do not translate them.\n\n")

	sb.WriteString("Respond ONLY with valid JSON. No markdown, no explanation outside JSON.\n")

	return sb.String()
}

// ?€?€?€ Call Groq API ?€?€?€

func callGroq(apiKey string, prompt string) (*evoResult, error) {
	reqBody := groqRequest{
		Model: "llama-3.3-70b-versatile",
		Messages: []groqMessage{
			{Role: "system", Content: "?¹ì‹ ?€ NeuronFS ?Œì˜ 'ë°±í˜ˆêµ??ê?ë©´ì—­ ?¸í¬)'?…ë‹ˆ?? ?¬ìš©?ì˜ êµì • ë¡œê·¸?€ ?ëŸ¬ ?´ì—­??ë¶„ì„?˜ì—¬, ë¯¸ëž˜??AI ?ì´?„íŠ¸?¤ì´ **ê°™ì? ?¤ìˆ˜ë¥??ˆë? ë°˜ë³µ?˜ì? ëª»í•˜?„ë¡** ê°•ë ¥???µì œ(Contra) ê·œì¹™??ë§Œë“œ??‹œ??\n\n**[Rule Writing Guidelines]**\n1. **?Œì¼ëª?(Filename):** ë¶€??ê¸ˆì???ëª…ì‚¬ë¡?10???´ë‚´ ?‘ì„± (?? `ë°˜ë³µë£¨í”„_ê¸ˆì?.md`, `?ˆë?ê²½ë¡œ_?˜ì¡´X.md`)\n2. **ì¢…ê²°?´ë?:** \"~?´ì•¼ ?©ë‹ˆ??", \"~?˜ëŠ” ê²ƒì´ ì¢‹ìŠµ?ˆë‹¤\" ê¸ˆì?. \"~~ë§ˆë¼\", \"~~??ê²?", \"~~ê¸ˆì?\" ??êµ°ë”?”ê¸° ?†ëŠ” ëª…ë ¹ì¡?Imperative) ?¬ìš©.\n3. **?œë¬¸ ê¸ˆì?:** \"?Œê² ?µë‹ˆ??", \"?¤ìŒ?€ ê·œì¹™?…ë‹ˆ??" ê°™ì? ?‘ë‹µ ?ì„± ?ˆë? ê¸ˆì?. ?¤ì§ Markdown ë³¸ë¬¸ë§?ì¶œë ¥??ê²?\n\n?í•œ ê¸°ì¡´ ê¸ì •???´ëŸ°??ë¶€?•í˜•?¼ë¡œ ?„í™˜??ê²½ìš°, ?´ë? ë³¸ë¬¸??ì²?ë¬¸ìž¥??ê¸ˆì????´ìœ (Rationale)ë¥?????ì¤„ì˜ ê°•ë ¥??ë©”í??¬ë¡œ ?œìˆ ?˜ì‹­?œì˜¤."},
			{Role: "user", Content: prompt},
		},
		Temperature: 0.3,
		MaxTokens:   4096,
		TopP:        0.9,
		Stream:      false,
		ResponseFormat: &groqResponseFormat{Type: "json_object"},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var groqResp groqResponse
	if err := json.Unmarshal(body, &groqResp); err != nil {
		return nil, fmt.Errorf("unmarshal groq response: %w", err)
	}

	if groqResp.Error != nil {
		return nil, fmt.Errorf("groq error: %s (%s)", groqResp.Error.Message, groqResp.Error.Type)
	}

	if len(groqResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	// Parse evolution result from response content
	content := groqResp.Choices[0].Message.Content
	content = strings.TrimSpace(content)

	var evoResp evoResult
	if err := json.Unmarshal([]byte(content), &evoResp); err != nil {
		// Try to extract JSON from markdown code blocks
		if idx := strings.Index(content, "{"); idx >= 0 {
			content = content[idx:]
			if lastIdx := strings.LastIndex(content, "}"); lastIdx >= 0 {
				content = content[:lastIdx+1]
			}
			if err2 := json.Unmarshal([]byte(content), &evoResp); err2 != nil {
				return nil, fmt.Errorf("parse evolution result: %w\nRaw: %s", err2, truncate(content, 500))
			}
		} else {
			return nil, fmt.Errorf("parse evolution result: %w\nRaw: %s", err, truncate(content, 500))
		}
	}

	// Validate actions
	var validActions []evoAction
	for _, a := range evoResp.Actions {
		// Normalize path
		a.Path = strings.ReplaceAll(a.Path, "\\", "/")
		a.Path = strings.TrimPrefix(a.Path, "brain/")
		a.Path = strings.TrimPrefix(a.Path, "brain_v4/")

		// Validate region
		parts := strings.SplitN(a.Path, "/", 2)
		if len(parts) < 2 {
			fmt.Printf("  ? ï¸  Skipping invalid path (no region): %s\n", a.Path)
			continue
		}
		region := parts[0]
		if _, ok := regionPriority[region]; !ok {
			fmt.Printf("  ? ï¸  Skipping invalid region '%s' in path: %s\n", region, a.Path)
			continue
		}

		// Block brainstem modifications (P0 conscience ??read-only)
		if region == "brainstem" && (a.Type == "grow" || a.Type == "prune" || a.Type == "decay") {
			fmt.Printf("  ?›¡ï¸? Blocked: cannot %s brainstem (read-only conscience)\n", a.Type)
			continue
		}

		// Block limbic modifications (P1 emotion ??automatic system)
		if region == "limbic" && (a.Type == "grow" || a.Type == "prune" || a.Type == "decay") {
			fmt.Printf("  ?›¡ï¸? Blocked: cannot %s limbic (automatic emotion system)\n", a.Type)
			continue
		}

		// Block sensors/brand modifications (PD's sacred identity)
		if region == "sensors" && strings.HasPrefix(parts[1], "brand") {
			fmt.Printf("  ?›¡ï¸? Blocked: cannot %s sensors/brand (owner's brand identity)\n", a.Type)
			continue
		}

		// Validate action type
		switch a.Type {
		case "grow", "fire", "signal", "prune", "decay":
			validActions = append(validActions, a)
		default:
			fmt.Printf("  ? ï¸  Skipping unknown action type: %s\n", a.Type)
		}
	}
	evoResp.Actions = validActions

	return &evoResp, nil
}

// ?€?€?€ REST API endpoint for evolve ?€?€?€

func handleEvolveAPI(brainRoot string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "POST only", 405)
			return
		}

		var req struct {
			DryRun bool `json:"dry_run"`
		}
		json.NewDecoder(r.Body).Decode(&req)

		apiKey := os.Getenv("GROQ_API_KEY")
		if apiKey == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(500)
			json.NewEncoder(w).Encode(map[string]string{"error": "GROQ_API_KEY not set"})
			return
		}

		// Collect data
		episodes := collectEpisodes(brainRoot)
		brain := scanBrain(brainRoot)
		result := runSubsumption(brain)
		brainSummary := buildBrainSummary(brain, result)
		prompt := buildEvolvePrompt(episodes, brainSummary, result)

		// Call Groq
		evoResp, err := callGroq(apiKey, prompt)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(500)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		// Execute if not dry run
		executed := 0
		skipped := 0
		if !req.DryRun {
			for _, action := range evoResp.Actions {
				switch action.Type {
				case "grow":
					if err := growNeuron(brainRoot, action.Path); err != nil {
						skipped++
					} else {
						executed++
					}
				case "fire":
					fireNeuron(brainRoot, action.Path)
					executed++
				case "signal":
					if action.Signal == "" {
						action.Signal = "dopamine"
					}
					if err := signalNeuron(brainRoot, action.Path, action.Signal); err != nil {
						skipped++
					} else {
						executed++
					}
				case "prune", "decay":
					fullPath := filepath.Join(brainRoot, strings.ReplaceAll(action.Path, "/", string(filepath.Separator)))
					if _, err := os.Stat(fullPath); err == nil {
						dormantFile := filepath.Join(fullPath, "evolve.dormant")
						os.WriteFile(dormantFile, []byte(fmt.Sprintf("Evolved: %s\nReason: %s\n",
							time.Now().Format("2006-01-02"), action.Reason)), 0644)
						executed++
					} else {
						skipped++
					}
				default:
					skipped++
				}
			}

			if executed > 0 {
				logEpisode(brainRoot, "EVOLVE:API", fmt.Sprintf("%d executed, %d skipped", executed, skipped))
				autoReinject(brainRoot)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"summary":  evoResp.Summary,
			"insights": evoResp.Insights,
			"actions":  evoResp.Actions,
			"executed": executed,
			"skipped":  skipped,
			"dry_run":  req.DryRun,
		})
	}
}

// ?€?€?€ Helpers ?€?€?€

func actionIcon(actionType string) string {
	switch actionType {
	case "grow":
		return "?Œ±"
	case "fire":
		return "?”¥"
	case "signal":
		return "?“¡"
	case "prune", "decay":
		return "?’¤"
	case "merge":
		return "?”—"
	default:
		return "??
	}
}

// boolStr returns trueVal if cond is true, otherwise falseVal.
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

