package main

import (
	"context"
	"fmt"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"strings"
)

// main is the entry point for the NeuronFS CLI and background daemon.
func main() {
	// ── Flatline Death Screen: catch any unrecoverable panic ──
	defer RenderFlatlineOnPanic()

	brainRoot := findBrainRoot()
	if brainRoot == "" {
		fmt.Println("[FATAL] brain directory not found")
		fmt.Println("Usage: neuronfs <brain_path> [--emit|--inject|--watch|--dashboard|--grow|--fire|--signal|--decay|--api|--harness]")
		os.Exit(1)
	}

	quietMode := false
	forceAwakening := false

	// Legacy global flag parse (commands should parse their own specific flags)
	for _, arg := range os.Args {
		if arg == "--quiet" || arg == "-q" {
			quietMode = true
		}
		if arg == "--awakening" {
			forceAwakening = true
		}
	}

	// ── VFS Cartridge Ignition (Phase 2) ──
	jlootCartridge := filepath.Join(brainRoot, "base.jloot")
	if _, err := os.Stat(jlootCartridge); os.IsNotExist(err) {
		jlootCartridge = "" // Run purely UpperDir if no cartridge
	}

	if err := MountCartridge(jlootCartridge, brainRoot); err != nil {
		fmt.Printf("\033[31m[FATAL] VFS Hardware Error: %v\033[0m\n", err)
		os.Exit(1)
	}

	// ── Awakening Sequence (first-run boot animation) ──
	// Propagate quiet mode to flatline handler too
	FlatlineQuiet = quietMode
	RunAwakening(context.Background(), AwakeningConfig{
		BrainRoot:      brainRoot,
		Quiet:          quietMode,
		ForceAwakening: forceAwakening,
	})

	// ── Strangler Fig CLI Router Injection ──
	router := NewRouter()
	router.Register(&HarnessCmd{})
	router.Register(&InitCmd{})
	router.Register(&InjectCmd{})
	router.Register(&EvolveCmd{})
	router.Register(&WatchCmd{})
	router.Register(&DashboardCmd{})
	router.Register(&APICmd{})
	router.Register(&HtmlCmd{})
	router.Register(&GrowCmd{})
	router.Register(&FireCmd{})
	router.Register(&DashboardV2Cmd{})
	router.Register(&SignalCmd{})
	router.Register(&DecayCmd{})
	router.Register(&SnapshotCmd{})
	router.Register(&RollbackCmd{})
	router.Register(&RollbackAllCmd{})
	router.Register(&ExportSvgCmd{})
	router.Register(&StatsCmd{})
	router.Register(&VacuumCmd{})
	router.Register(&McpCmd{})
	router.Register(&SupervisorCmd{})
	router.Register(&NeuronizeCmd{})
	router.Register(&PolarizeCmd{})
	router.Register(&SymlinkCmd{})
	router.Register(&ToolCmd{})
	router.Register(&DiagCmd{})
	router.Register(&EmitCmd{})

	// Check if any arguments match our new router
	routed := false
	for _, arg := range os.Args {
		// Only check flags, as router uses flag names currently
		if strings.HasPrefix(arg, "--") {
			ok, err := router.TryRoute(arg, brainRoot, os.Args)
			if ok {
				if err != nil {
					fmt.Printf("[FATAL] %s execution failed: %v\n", arg, err)
					os.Exit(1)
				}
				routed = true
				break
			}
		}
	}

	if routed {
		return // Command fully handled by new router
	}

	// If it reached here without routing, it's an unknown invocation format
	fmt.Printf("[FATAL] Unknown command. Could not route given arguments.\n")
	os.Exit(1)
}

// ─── Region priority (hardcoded — no folder prefix numbers) ───
// ━━━ Brain structs/scanner → brain.go ━━━
// MOVED: Neuron, Region, Brain, SubsumptionResult, scanBrain,
//        runSubsumption, findBrainRoot, regionPriority, counterRegex

func emitRules(result SubsumptionResult) string {
	// Find brainRoot from first active region
	brainRoot := ""
	for _, r := range result.ActiveRegions {
		if r.Path != "" {
			brainRoot = filepath.Dir(r.Path)
			break
		}
	}
	if brainRoot == "" {
		home := os.Getenv("USERPROFILE")
		if home != "" {
			brainRoot = filepath.Join(home, "NeuronFS", "brain_v4")
		} else {
			brainRoot = "brain_v4"
		}
	}
	return emitBootstrap(result, brainRoot)
}

// activationBar visualizes a neuron's activation counter as a discrete block bar.
func activationBar(counter int) string {
	if counter >= 90 {
		return "█████"
	} else if counter >= 50 {
		return "████░"
	} else if counter >= 20 {
		return "███░░"
	} else if counter >= 10 {
		return "██░░░"
	} else if counter >= 5 {
		return "█░░░░"
	}
	return "░░░░░"
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// INJECT: Write rules into GEMINI.md
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
func injectToGemini(brainRoot string, rules string) {
	// 테스트 격리: brainRoot 내부에만 쓴다
	if os.Getenv("NEURONFS_TEST_ISOLATION") != "" {
		safePath := filepath.Join(brainRoot, ".gemini", "GEMINI.md")
		os.MkdirAll(filepath.Dir(safePath), 0750)
		os.WriteFile(safePath, []byte(rules), 0600)
		fmt.Printf("[OK] Rules injected → %s (test isolation)\n", safePath)
		return
	}

	// 글로벌 단일 경로: USERPROFILE/.gemini/GEMINI.md (전체 덮어쓰기)
	home := os.Getenv("USERPROFILE")
	if home == "" {
		fmt.Println("[WARN] USERPROFILE not set, outputting to stdout:")
		fmt.Print(rules)
		return
	}

	geminiPath := filepath.Join(home, ".gemini", "GEMINI.md")
	os.MkdirAll(filepath.Dir(geminiPath), 0750)
	// 전체 덮어쓰기 — doInject 금지 (중복 누적 원인)
	if err := os.WriteFile(geminiPath, []byte(rules), 0600); err != nil {
		fmt.Printf("[ERROR] Cannot write %s: %v\n", geminiPath, err)
		return
	}
	fmt.Printf("[OK] Rules injected → %s\n", geminiPath)
}

// doInject executes the injection of aggregated rules into target AI configuration files.
func doInject(geminiPath string, rules string) {
	existing, err := os.ReadFile(geminiPath)
	if err != nil {
		fmt.Printf("[ERROR] Cannot read %s: %v\n", geminiPath, err)
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

	err = os.WriteFile(geminiPath, []byte(content), 0600)
	if err != nil {
		fmt.Printf("[ERROR] Cannot write %s: %v\n", geminiPath, err)
		return
	}

	// Count active neurons
	activeCount := strings.Count(rules, "- **")
	fmt.Printf("[OK] Rules injected → %s\n", geminiPath)
	fmt.Printf("[OK] %d neurons active\n", activeCount)
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// WATCH: fsnotify Event-Driven Monitor + auto-inject
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// ANSI escape codes for premium CLI aesthetics
const (
	ansiReset   = "\033[0m"
	ansiCyan    = "\033[36m"
	ansiMagenta = "\033[35m"
	ansiYellow  = "\033[33m"
	ansiGreen   = "\033[32m"
	ansiRed     = "\033[31m"
	ansiDimGray = "\033[90m"
	ansiWhite   = "\033[37m"
)

// ━━━ Watch → watch.go ━━━
// ━━━ Diagnostics → diag.go ━━━
// ━━━ Neuron CRUD → neuron_crud.go ━━━
// ━━━ Injection Pipeline → inject.go ━━━
// ━━━ Transcripts/Idle → transcript.go ━━━
