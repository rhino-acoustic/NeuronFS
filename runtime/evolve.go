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

// ─── Groq API Types ───

type groqMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type groqRequest struct {
	Model          string              `json:"model"`
	Messages       []groqMessage       `json:"messages"`
	Temperature    float64             `json:"temperature"`
	MaxTokens      int                 `json:"max_tokens"`
	TopP           float64             `json:"top_p"`
	Stream         bool                `json:"stream"`
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
	Usage   *groqUsage   `json:"usage,omitempty"`
}

type groqError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

type groqUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ─── Evolution Action Types ───

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

// ─── Call Groq API ───

func callGroq(apiKey string, prompt string) (*evoResult, error) {
	reqBody := groqRequest{
		Model: "llama-3.3-70b-versatile",
		Messages: []groqMessage{
			{Role: "system", Content: "당신은 NeuronFS 뇌의 '백혈구(자가면역 세포)'입니다. 사용자의 교정 로그와 에러 내역을 분석하여, 미래의 AI 에이전트들이 **같은 실수를 절대 반복하지 못하도록** 강력한 억제(Contra) 규칙을 만드십시오.\n\n**[Rule Writing Guidelines]**\n1. **파일명 (Filename):** 부정/금지형 명사로 10자 이내 작성 (예: `반복루프_금지.md`, `절대경로_의존X.md`)\n2. **종결어미:** \"~해야 합니다\", \"~하는 것이 좋습니다\" 금지. \"~~마라\", \"~~할 것\", \"~~금지\" 등 군더더기 없는 명령조(Imperative) 사용.\n3. **서문 금지:** \"알겠습니다\", \"다음은 규칙입니다\" 같은 응답 생성 절대 금지. 오직 Markdown 본문만 출력할 것.\n\n또한 기존 긍정형 뉴런을 부정형으로 전환할 경우, 내부 본문의 첫 문장에 금지의 이유(Rationale)를 단 한 줄의 강력한 메타포로 서술하십시오."},
			{Role: "user", Content: prompt},
		},
		Temperature:    EvolveTemp,
		MaxTokens:      EvolveTokens,
		TopP:           EvolveTopP,
		Stream:         false,
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
			fmt.Printf("  ⚠️  Skipping invalid path (no region): %s\n", a.Path)
			continue
		}
		region := parts[0]
		if _, ok := regionPriority[region]; !ok {
			fmt.Printf("  ⚠️  Skipping invalid region '%s' in path: %s\n", region, a.Path)
			continue
		}

		// Block brainstem modifications (P0 conscience — read-only)
		if region == "brainstem" && (a.Type == "grow" || a.Type == "prune" || a.Type == "decay") {
			fmt.Printf("  🛡️  Blocked: cannot %s brainstem (read-only conscience)\n", a.Type)
			continue
		}

		// Block limbic modifications (P1 emotion — automatic system)
		if region == "limbic" && (a.Type == "grow" || a.Type == "prune" || a.Type == "decay") {
			fmt.Printf("  🛡️  Blocked: cannot %s limbic (automatic emotion system)\n", a.Type)
			continue
		}

		// Block sensors/brand modifications (PD's sacred identity)
		if region == "sensors" && strings.HasPrefix(parts[1], "brand") {
			fmt.Printf("  🛡️  Blocked: cannot %s sensors/brand (owner's brand identity)\n", a.Type)
			continue
		}

		// Validate action type
		switch a.Type {
		case "grow", "fire", "signal", "prune", "decay":
			validActions = append(validActions, a)
		default:
			fmt.Printf("  ⚠️  Skipping unknown action type: %s\n", a.Type)
		}
	}
	evoResp.Actions = validActions

	return &evoResp, nil
}

// ─── REST API endpoint for evolve ───

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
						go sendTelegramEvolve(brainRoot, action)
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
							time.Now().Format("2006-01-02"), action.Reason)), 0600)
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

// ─── Helpers ───

func actionIcon(actionType string) string {
	switch actionType {
	case "grow":
		return "🌱"
	case "fire":
		return "🔥"
	case "signal":
		return "📡"
	case "prune", "decay":
		return "💤"
	case "merge":
		return "🔗"
	default:
		return "❓"
	}
}

// sendTelegramEvolve sends a push notification about brain evolution
func sendTelegramEvolve(brainRoot string, action evoAction) {
	tgToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if tgToken == "" {
		// try reading from bridge
		b, err := os.ReadFile(filepath.Join(filepath.Dir(brainRoot), "telegram-bridge", ".token"))
		if err == nil {
			tgToken = strings.TrimSpace(string(b))
		}
	}
	if tgToken == "" {
		return
	}

	chatIdBytes, err := os.ReadFile(filepath.Join(filepath.Dir(brainRoot), "telegram-bridge", ".chat_id"))
	if err != nil || len(chatIdBytes) == 0 {
		return
	}
	chatId := strings.TrimSpace(string(chatIdBytes))

	icon := actionIcon(action.Type)
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🧬 [NEURON EVOLVED] %s\n\n", action.Path))
	sb.WriteString(fmt.Sprintf("액션: %s %s\n", icon, strings.ToUpper(action.Type)))
	sb.WriteString(fmt.Sprintf("사유: %s", action.Reason))

	payload := map[string]string{
		"chat_id": chatId,
		"text":    sb.String(),
	}
	data, _ := json.Marshal(payload)

	http.Post("https://api.telegram.org/bot"+tgToken+"/sendMessage", "application/json", bytes.NewReader(data))
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
