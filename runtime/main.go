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

	mode := "diag"
	dryRun := false
	quietMode := false
	forceAwakening := false
	emitTarget := "" // --emit target: gemini, cursor, claude, copilot, generic, all
	for i, arg := range os.Args {
		switch arg {
		case "--emit":
			mode = "emit"
			// Check if next arg is an emit target (not a flag)
			if i+1 < len(os.Args) && !strings.HasPrefix(os.Args[i+1], "--") {
				candidate := strings.ToLower(os.Args[i+1])
				if candidate == "gemini" || candidate == "cursor" || candidate == "claude" || candidate == "copilot" || candidate == "generic" || candidate == "all" || candidate == "auto" {
					emitTarget = candidate
					mode = "emit-target" // file output mode
				}
			}
		case "--inject":
			mode = "inject"
		case "--watch":
			mode = "watch"
		case "--html":
			mode = "html"
		case "--dashboard":
			mode = "dashboard"
		case "--init":
			mode = "init"
		case "--grow":
			mode = "grow"
		case "--fire":
			mode = "fire"
		case "--signal":
			mode = "signal"
		case "--decay":
			mode = "decay"
		case "--api":
			mode = "api"
		case "--mcp":
			mode = "mcp"
		case "--snapshot":
			mode = "snapshot"
		case "--evolve":
			mode = "evolve"
		case "--rollback":
			mode = "rollback"
		case "--rollback-all":
			mode = "rollback-all"
		case "--stats":
			mode = "stats"
		case "--vacuum":
			mode = "vacuum"
		case "--webhook":
			// Standby for P9 (B2B Social Pressure / Slack Shaming Hook)
			if i+1 < len(os.Args) && !strings.HasPrefix(os.Args[i+1], "--") {
				_ = os.Args[i+1] // webhookUrl = os.Args[i+1]
			}
		case "--supervisor":
			mode = "supervisor"
		case "--neuronize":
			mode = "neuronize"
		case "--polarize":
			mode = "polarize"
		case "--symlink":
			mode = "symlink"
		case "--dry-run":
			dryRun = true
		case "--quiet", "-q":
			quietMode = true
		case "--awakening":
			forceAwakening = true
		case "--harness":
			mode = "harness"
		case "--tool":
			mode = "tool"
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

	// ── Legacy Switch-Case Fallback (Anti-Corruption Layer) ──
	switch mode {
	case "init":
		// Handled by router
	case "diag":
		brain := scanBrain(brainRoot)
		result := runSubsumption(brain)
		printDiag(brain, result)
	case "emit":
		brain := scanBrain(brainRoot)
		result := runSubsumption(brain)
		fmt.Print(emitRules(result))
	case "emit-target":
		processInbox(brainRoot)
		writeAllTiersForTargets(brainRoot, emitTarget)
	case "harness":
		// Handled by router
	case "stats":
		runStats(brainRoot)
	case "vacuum":
		runVacuum(brainRoot)
	case "mcp":
		// MCP Streamable HTTP server + background loops
		// HTTP transport: IDE 재시작에도 연결 유지
		go func() {
			mcpAPIPort := MCPPort
			fmt.Fprintf(os.Stderr, "[MCP] REST API on :%d (fallback)\n", mcpAPIPort)
			startAPI(brainRoot, mcpAPIPort)
		}()
		go runInjectionLoop(brainRoot)
		go runIdleLoop(brainRoot)
		startMCPHTTPServer(brainRoot, MCPStreamPort) // blocking: HTTP server
	case "supervisor":
		runSupervisor(brainRoot)
	case "tool":
		toolName := ""
		argsJson := ""
		for i, arg := range os.Args {
			if arg == "--tool" && i+2 < len(os.Args) {
				toolName = os.Args[i+1]
				argsJson = os.Args[i+2]
				break
			}
		}
		if toolName == "" {
			fmt.Println("[FATAL] Usage: neuronfs <brain> --tool <toolname> <args_json>")
			os.Exit(1)
		}
		runWorkerTool(brainRoot, toolName, argsJson)
	case "neuronize":
		runNeuronize(brainRoot, dryRun)
	case "polarize":
		runPolarize(brainRoot, dryRun)
	case "symlink":
		targetDir := ""
		for i, arg := range os.Args {
			if arg == "--symlink" && i+1 < len(os.Args) {
				targetDir = os.Args[i+1]
				break
			}
		}
		if targetDir == "" {
			fmt.Println("[FATAL] Usage: neuronfs <brain> --symlink <global_path>")
			os.Exit(1)
		}

		sharedDir := filepath.Join(brainRoot, ".neuronfs", "shared")
		os.MkdirAll(filepath.Dir(sharedDir), 0750)

		absTarget, _ := filepath.Abs(targetDir)
		err := os.Symlink(absTarget, sharedDir)
		if err != nil {
			out, e2 := SafeCombinedOutput(ExecTimeoutShell, "cmd", "/c", "mklink", "/J", sharedDir, absTarget)
			if e2 != nil {
				fmt.Printf("\033[31m[ERROR] Symlink/Junction failed: %v, out: %s\033[0m\n", e2, string(out))
			} else {
				fmt.Printf("\033[32m[OK] Created Junction %s -> %s\033[0m\n", sharedDir, absTarget)
			}
		} else {
			fmt.Printf("\033[32m[OK] Created symlink %s -> %s\033[0m\n", sharedDir, absTarget)
		}
	}
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
