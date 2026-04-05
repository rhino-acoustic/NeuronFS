package main

// ━━━ diag.go ━━━
// PROVIDES: printDiag, generateBrainJSON, getNonFlagArg, refreshCodeMap
// DEPENDS ON: brain.go

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
func printDiag(brain Brain, result SubsumptionResult) {
	fmt.Println("═══ NeuronFS v4.0 — Folder-as-Neuron Engine ═══")
	fmt.Printf("  Brain: %s\n", brain.Root)
	fmt.Printf("  Axiom: Folder=Neuron | File=Trace | Path=Sentence\n\n")

	for _, region := range brain.Regions {
		icon := regionIcons[region.Name]
		ko := regionKo[region.Name]
		bomb := " "
		if region.HasBomb {
			bomb = "💀"
		}

		totalCounter := 0
		for _, n := range region.Neurons {
			totalCounter += n.Counter
		}

		fmt.Printf("  %s %s %-13s %s | neurons: %3d | activation: %5d | axons: %d\n",
			bomb, icon, region.Name, ko, len(region.Neurons), totalCounter, len(region.Axons))
	}

	fmt.Println()
	if result.BombSource != "" {
		fmt.Printf("  💀 BOMB: %s\n", result.BombSource)
	}
	fmt.Println("  Active:")
	for _, r := range result.ActiveRegions {
		fmt.Printf("    + %s (%d neurons)\n", r.Name, len(r.Neurons))
	}
	if len(result.BlockedRegions) > 0 {
		fmt.Println("  Blocked:")
		for _, r := range result.BlockedRegions {
			fmt.Printf("    - %s\n", r)
		}
	}

	// Shadow Context (Dashboard UI Panel logic inline for CLI)
	diffCmd := exec.Command("git", "status", "--porcelain")
	diffCmd.Dir = brain.Root
	out, err := diffCmd.Output()
	var shadowFiles []string
	if err == nil {
		lines := strings.Split(string(out), "\n")
		for _, idx := range lines {
			idx = strings.TrimSpace(idx)
			if len(idx) > 2 {
				shadowFiles = append(shadowFiles, idx[2:])
			}
		}
	}
	
	fmt.Printf("\n  [Shadow Context: %d active working files]\n", len(shadowFiles))
	limit := 5
	if len(shadowFiles) < 5 { limit = len(shadowFiles) }
	for i:=0; i<limit; i++ {
		fmt.Printf("    * %s\n", strings.TrimSpace(shadowFiles[i]))
	}
	if len(shadowFiles) > 5 {
		fmt.Printf("    * ... (%d more)\n", len(shadowFiles)-5)
	}

	fmt.Printf("\n  Total: %d/%d neurons | Activation: %d | Status: ",
		result.FiredNeurons, result.TotalNeurons, result.TotalCounter)
	if result.BombSource == "" {
		fmt.Println("NOMINAL")
	} else {
		fmt.Println("CIRCUIT BREAKER ACTIVE")
	}
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// JSON OUTPUT: Pure data for dashboard consumption
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
func generateBrainJSON(brainRoot string, _ Brain, result SubsumptionResult) {
	type JsNeuron struct {
		Path      string  `json:"path"`
		Counter   int     `json:"counter"`
		Contra    int     `json:"contra,omitempty"`
		Depth     int     `json:"depth"`
		Dopamine  int     `json:"dopamine"`
		Intensity int     `json:"intensity"`
		Polarity  float64 `json:"polarity"`
		HasBomb   bool    `json:"hasBomb,omitempty"`
		Memory    bool    `json:"memory,omitempty"`
		Dormant   bool    `json:"dormant,omitempty"`
	}
	type JsRegion struct {
		Name     string     `json:"name"`
		Priority int        `json:"priority"`
		Icon     string     `json:"icon"`
		Ko       string     `json:"ko"`
		Axons    []string   `json:"axons"`
		Neurons  []JsNeuron `json:"neurons"`
	}
	type JsEdge struct {
		From string `json:"from"`
		To   string `json:"to"`
	}
	type BrainState struct {
		Generated    string     `json:"generated"`
		BrainPath    string     `json:"brainPath"`
		TotalNeurons int        `json:"totalNeurons"`
		FiredNeurons int        `json:"firedNeurons"`
		TotalCounter int        `json:"totalCounter"`
		BombSource   string     `json:"bombSource,omitempty"`
		Regions      []JsRegion `json:"regions"`
		Edges        []JsEdge   `json:"edges"`
	}

	state := BrainState{
		Generated:    time.Now().Format("2006-01-02T15:04:05"),
		BrainPath:    brainRoot,
		TotalNeurons: result.TotalNeurons,
		FiredNeurons: result.FiredNeurons,
		TotalCounter: result.TotalCounter,
		BombSource:   result.BombSource,
	}

	for _, r := range result.ActiveRegions {
		jr := JsRegion{
			Name:     r.Name,
			Priority: r.Priority,
			Icon:     regionIcons[r.Name],
			Ko:       regionKo[r.Name],
			Axons:    r.Axons,
		}
		for _, n := range r.Neurons {
			jr.Neurons = append(jr.Neurons, JsNeuron{
				Path: n.Path, Counter: n.Counter, Contra: n.Contra, Depth: n.Depth,
				Dopamine: n.Dopamine, Intensity: n.Intensity, Polarity: n.Polarity,
				HasBomb: n.HasBomb, Memory: n.HasMemory, Dormant: n.IsDormant,
			})
		}
		state.Regions = append(state.Regions, jr)

		for _, axon := range r.Axons {
			target := axon
			if strings.HasPrefix(axon, "SKILL:") {
				target = "skill:" + filepath.Base(filepath.Dir(strings.TrimPrefix(axon, "SKILL:")))
			}
			state.Edges = append(state.Edges, JsEdge{From: r.Name, To: target})
		}
	}

	jsonBytes, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		fmt.Printf("[ERROR] JSON marshal: %v\n", err)
		return
	}

	outPath := filepath.Join(brainRoot, "..", "brain_state.json")
	abs, _ := filepath.Abs(outPath)
	if err := os.WriteFile(abs, jsonBytes, 0644); err != nil {
		fmt.Printf("[ERROR] Write: %v\n", err)
		return
	}
	// fmt.Printf("[OK] Brain state → %s\n", abs) // Suppress for autoReinject
}

// initBrain is defined in init.go
// startDashboard is defined in dashboard.go

// getNonFlagArg returns the Nth non-flag argument (0-indexed)
func getNonFlagArg(n int) string {
	idx := 0
	skipNext := false
	for _, arg := range os.Args[1:] { // Start from the first argument after the program name
		if skipNext {
			skipNext = false
			continue
		}
		// These flags consume the next argument as their value
		if arg == "--signal" || arg == "--decay" {
			skipNext = true
			continue
		}
		if strings.HasPrefix(arg, "--") {
			continue // Skip other flags
		}
		if idx == n {
			return arg
		}
		idx++
	}
	return ""
}

// ━━━ Similarity Engine → similarity.go ━━━
// MOVED: tokenize, stem, jaccardSimilarity, hybridSimilarity,
//        cosineTokens, levenshteinDistance, extractPrefix,
//        newtonSqrt, maxInt, minInt


// ━━━ Neuron CRUD → neuron_crud.go ━━━
// ━━━ Injection Pipeline → inject.go ━━━
// ━━━ Transcripts/Idle → transcript.go ━━━

// refreshCodeMap regenerates runtime/CODE_MAP.md with current file stats.
// Called by IDLE loop to keep code structure docs in sync.
func refreshCodeMap(brainRoot string) {
	runtimeDir := filepath.Join(filepath.Dir(brainRoot), "runtime")
	codeMapPath := filepath.Join(runtimeDir, "CODE_MAP.md")

	entries, err := os.ReadDir(runtimeDir)
	if err != nil {
		return
	}

	var sb strings.Builder
	sb.WriteString("# NeuronFS Runtime — CODE_MAP\n")
	sb.WriteString(fmt.Sprintf("<!-- AUTO-GENERATED by IDLE loop: %s -->\n\n", time.Now().Format("2006-01-02T15:04")))
	sb.WriteString("## File Summary\n\n")
	sb.WriteString("| File | Lines | Role |\n")
	sb.WriteString("|------|-------|------|\n")

	totalLines := 0
	totalFiles := 0
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		fpath := filepath.Join(runtimeDir, e.Name())
		data, err := os.ReadFile(fpath)
		if err != nil {
			continue
		}
		lines := strings.Count(string(data), "\n") + 1
		totalLines += lines
		totalFiles++

		// Extract role from header comment (// PROVIDES: ...)
		role := ""
		for _, line := range strings.SplitN(string(data), "\n", 20) {
			if strings.Contains(line, "PROVIDES:") {
				role = strings.TrimSpace(strings.SplitN(line, "PROVIDES:", 2)[1])
				if len(role) > 60 {
					role = role[:60] + "..."
				}
				break
			}
		}
		sb.WriteString(fmt.Sprintf("| %s | %d | %s |\n", e.Name(), lines, role))
	}
	sb.WriteString(fmt.Sprintf("\n**Total: %d files, %d lines**\n", totalFiles, totalLines))
	sb.WriteString("\n## Critical Rules\n")
	sb.WriteString("1. ALL files are `package main` — no sub-packages\n")
	sb.WriteString("2. Always verify: `go vet ./...` → `go build .` after ANY change\n")
	sb.WriteString("3. main.go is the orchestrator (CLI dispatch only)\n")

	os.WriteFile(codeMapPath, []byte(sb.String()), 0644)
	fmt.Printf("[IDLE] 📋 CODE_MAP.md refreshed (%d files, %d lines)\n", totalFiles, totalLines)
}