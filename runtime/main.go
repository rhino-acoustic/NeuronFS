// NeuronFS Runtime v4.0 — Folder-as-Neuron Cognitive Engine
//
// AXIOMS:
//   1. Folder = Neuron (name is meaning, depth is specificity)
//   2. File = Firing Trace (N.neuron = counter, dopamineN = reward, bomb = pain)
//   3. Path = Sentence (brain/cortex/quality/no_hardcoded → "cortex > quality > no_hardcoded")
//   4. Counter = Activation (higher = stronger/myelinated path)
//   5. AI writes back (counter increment = experience growth)
//
// USAGE:
//   neuronfs <brain_path>              — diagnostics
//   neuronfs <brain_path> --emit       — output rules to stdout
//   neuronfs <brain_path> --emit <target> — emit to editor file (gemini/cursor/claude/copilot/generic/all)
//   neuronfs <brain_path> --inject     — write rules to GEMINI.md
//   neuronfs <brain_path> --watch      — watch + auto-inject
//   neuronfs <brain_path> --dashboard  — web dashboard on :9090
//   neuronfs <brain_path> --grow <path> — create new neuron
//   neuronfs <brain_path> --fire <path> — increment neuron counter
//   neuronfs <brain_path> --signal <type> <path> — add dopamine/bomb/memory
//   neuronfs <brain_path> --decay [days] — move inactive neurons to dormant
//   neuronfs <brain_path> --rollback <path> — decrement neuron counter (min=1)
//   neuronfs <brain_path> --api        — start REST API on :9090
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// ─── Region priority (hardcoded — no folder prefix numbers) ───
var regionPriority = map[string]int{
	"brainstem":   0,
	"limbic":      1,
	"hippocampus": 2,
	"sensors":     3,
	"cortex":      4,
	"ego":         5,
	"prefrontal":  6,
}

var regionIcons = map[string]string{
	"brainstem":   "🛡️",
	"limbic":      "💓",
	"hippocampus": "📝",
	"sensors":     "👁️",
	"cortex":      "🧠",
	"ego":         "🎭",
	"prefrontal":  "🎯",
}

var regionKo = map[string]string{
	"brainstem":   "양심/본능",
	"limbic":      "감정 필터",
	"hippocampus": "기록/기억",
	"sensors":     "환경 제약",
	"cortex":      "지식/기술",
	"ego":         "성향/톤",
	"prefrontal":  "목표/계획",
}

// ─── Neuron = a folder ───
type Neuron struct {
	Name      string    // folder name
	Path      string    // relative path from region root (e.g. "frontend/css/glass_blur20")
	FullPath  string    // absolute path
	Counter   int       // from N.neuron filename (correction count)
	Contra    int       // from N.contra filename (inhibition count)
	Dopamine  int       // from dopamineN.neuron filename (reward count)
	Intensity int       // Counter - Contra + Dopamine (net activation)
	Polarity  float64   // net / total (-1.0=pure inhibition, +1.0=pure excitation)
	HasBomb   bool      // bomb.neuron exists
	HasMemory bool      // memoryN.neuron exists
	HasGoal   bool      // .goal file exists (todo/objective)
	GoalText  string    // content of .goal file if present
	Geofence  string    // content of .geofence file if present
	IsDormant bool      // .dormant file exists
	Depth     int       // depth within region
	ModTime   time.Time // most recent .neuron file modification
}

// Emission thresholds
const (
	emitThreshold = 5 // min counter to appear in region listings
	spotlightDays = 7 // days a new neuron gets spotlight regardless of counter
)

// ─── Region ───
type Region struct {
	Name     string
	Priority int
	Path     string
	Neurons  []Neuron
	Axons    []string // .axon targets
	HasBomb  bool     // any neuron in this region has bomb
}

// ─── Brain ───
type Brain struct {
	Root    string
	Regions []Region
}

// ─── Subsumption Result ───
type SubsumptionResult struct {
	ActiveRegions  []Region
	BlockedRegions []string
	BombSource     string
	FiredNeurons   int
	TotalNeurons   int
	TotalCounter   int
}

// ─── Regex for trace files ───
var counterRegex = regexp.MustCompile(`^(\d+)\.neuron$`)
var dopamineRegex = regexp.MustCompile(`^dopamine(\d+)\.neuron$`)

// main is the entry point for the NeuronFS CLI and background daemon.
func main() {
	brainRoot := findBrainRoot()
	if brainRoot == "" {
		fmt.Println("[FATAL] brain directory not found")
		fmt.Println("Usage: neuronfs <brain_path> [--emit|--inject|--watch|--dashboard|--grow|--fire|--signal|--decay|--api]")
		os.Exit(1)
	}

	mode := "diag"
	port := 9090
	dryRun := false
	emitTarget := "" // --emit target: gemini, cursor, claude, copilot, generic, all
	for i, arg := range os.Args {
		switch arg {
		case "--emit":
			mode = "emit"
			// Check if next arg is an emit target (not a flag)
			if i+1 < len(os.Args) && !strings.HasPrefix(os.Args[i+1], "--") {
				candidate := strings.ToLower(os.Args[i+1])
				if candidate == "gemini" || candidate == "cursor" || candidate == "claude" || candidate == "copilot" || candidate == "generic" || candidate == "all" {
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
		case "--dry-run":
			dryRun = true
		}
	}

	switch mode {
	case "init":
		// --init requires a path argument
		initPath := ""
		for _, arg := range os.Args[1:] {
			if !strings.HasPrefix(arg, "--") {
				initPath = arg
				break
			}
		}
		if initPath == "" {
			initPath = filepath.Join(".", "brain_v4")
		}
		abs, _ := filepath.Abs(initPath)
		initBrain(abs)
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
	case "inject":
		processInbox(brainRoot)
		writeAllTiers(brainRoot)
	case "watch":
		fmt.Println("[NeuronFS] Watch mode — monitoring brain/ for changes...")
		runWatch(brainRoot)
	case "html":
		brain := scanBrain(brainRoot)
		result := runSubsumption(brain)
		generateBrainJSON(brainRoot, brain, result)
	case "dashboard":
		startDashboard(brainRoot, port)
	case "grow":
		neuronPath := getNonFlagArg(1) // brain_v4=0, path=1
		if neuronPath == "" {
			fmt.Println("[FATAL] Usage: neuronfs <brain> --grow <region/path/to/neuron>")
			fmt.Println("  Example: neuronfs brain_v4 --grow cortex/frontend/coding/no_console_log")
			os.Exit(1)
		}
		growNeuron(brainRoot, neuronPath)
	case "fire":
		neuronPath := getNonFlagArg(1)
		if neuronPath == "" {
			fmt.Println("[FATAL] Usage: neuronfs <brain> --fire <region/path/to/neuron>")
			os.Exit(1)
		}
		fireNeuron(brainRoot, neuronPath)
	case "signal":
		sigType := ""
		neuronPath := ""
		// Find sigType (the argument immediately after --signal)
		for i, arg := range os.Args {
			if arg == "--signal" && i+1 < len(os.Args) {
				sigType = os.Args[i+1]
				break
			}
		}
		// Find neuronPath (the first non-flag arg after brainRoot and sigType)
		neuronPath = getNonFlagArg(1) // brainRoot is 0, path is 1
		if sigType == "" || neuronPath == "" {
			fmt.Println("[FATAL] Usage: neuronfs <brain> --signal dopamine|bomb|memory <path>")
			os.Exit(1)
		}
		signalNeuron(brainRoot, neuronPath, sigType)
	case "decay":
		daysStr := "30" // Default decay days
		// Find days (the argument immediately after --decay)
		for i, arg := range os.Args {
			if arg == "--decay" && i+1 < len(os.Args) && !strings.HasPrefix(os.Args[i+1], "--") {
				daysStr = os.Args[i+1]
				break
			}
		}
		days, err := strconv.Atoi(daysStr)
		if err != nil || days <= 0 {
			fmt.Printf("[WARN] Invalid decay days '%s', using default 30 days.\n", daysStr)
			days = 30
		}
		runDecay(brainRoot, days)
	case "api":
		startAPI(brainRoot, port)
	case "snapshot":
		gitSnapshot(brainRoot)
	case "rollback":
		neuronPath := getNonFlagArg(1)
		if neuronPath == "" {
			fmt.Println("[FATAL] Usage: neuronfs <brain> --rollback <region/path/to/neuron>")
			os.Exit(1)
		}
		if err := rollbackNeuron(brainRoot, neuronPath); err != nil {
			fmt.Printf("[ERROR] rollback failed: %v\n", err)
			os.Exit(1)
		}
	case "stats":
		runStats(brainRoot)
	case "vacuum":
		runVacuum(brainRoot)
	case "evolve":
		runEvolve(brainRoot, dryRun)
	case "mcp":
		// MCP stdio server + background loops
		// CRITICAL: MCP stdio protocol requires stdout to be JSON-RPC only.
		// Redirect os.Stdout → os.Stderr so all fmt.Print* goes to stderr.
		// Preserve the real stdout for the MCP transport.
		realStdout := os.Stdout
		os.Stdout = os.Stderr

		// REST API on fallback port (9091) to avoid conflict with existing --api on 9090
		go func() {
			mcpAPIPort := port + 1 // 9091
			fmt.Fprintf(os.Stderr, "[MCP] REST API on :%d (fallback)\n", mcpAPIPort)
			startAPI(brainRoot, mcpAPIPort)
		}()
		go runInjectionLoop(brainRoot)
		go runIdleLoop(brainRoot)
		startMCPServerWithStdout(brainRoot, realStdout) // blocking: stdio loop
	case "supervisor":
		runSupervisor(brainRoot)
	}
}

// ─── Find brain root ───
func findBrainRoot() string {
	// First non-flag arg
	for _, arg := range os.Args[1:] {
		if !strings.HasPrefix(arg, "--") {
			abs, err := filepath.Abs(arg)
			if err == nil {
				if info, err := os.Stat(abs); err == nil && info.IsDir() {
					return abs
				}
			}
		}
	}
	// Fallback: look for brain/ or brain_v4/ nearby
	home := os.Getenv("USERPROFILE")
	candidates := []string{
		"brain_v4", "brain",
		filepath.Join("..", "brain_v4"), filepath.Join("..", "brain"),
	}
	if home != "" {
		candidates = append(candidates,
			filepath.Join(home, "NeuronFS", "brain_v4"),
			filepath.Join(home, "NeuronFS", "brain"),
		)
	}
	for _, c := range candidates {
		abs, _ := filepath.Abs(c)
		if info, err := os.Stat(abs); err == nil && info.IsDir() {
			return abs
		}
	}
	return ""
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// SCAN: Folder tree → Brain structure
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
func scanBrain(root string) Brain {
	brain := Brain{Root: root}

	entries, err := os.ReadDir(root)
	if err != nil {
		return brain
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		priority, ok := regionPriority[name]
		if !ok {
			continue
		}

		regionPath := filepath.Join(root, name)
		region := Region{
			Name:     name,
			Priority: priority,
			Path:     regionPath,
		}

		// Scan for .axon files at region root
		axonFiles, _ := filepath.Glob(filepath.Join(regionPath, "*.axon"))
		for _, af := range axonFiles {
			content, _ := os.ReadFile(af)
			target := strings.TrimSpace(string(content))
			target = strings.TrimPrefix(target, "TARGET: ")
			region.Axons = append(region.Axons, target)
		}

		// Scan flat neurons at region root (e.g., brainstem: 禁fallback.1.neuron)
		// Pattern: NeuronName.Counter.neuron or NeuronName.neuron
		flatNeuronRegex := regexp.MustCompile(`^(.+)\.(\d+)\.neuron$`)
		flatNeuronSimple := regexp.MustCompile(`^(.+)\.neuron$`)
		rootNeuronFiles, _ := filepath.Glob(filepath.Join(regionPath, "*.neuron"))
		neuronMap := make(map[string]*Neuron) // group by neuron name (Path)
		for _, nf := range rootNeuronFiles {
			fname := filepath.Base(nf)
			var neuronName string
			var counter int

			if m := flatNeuronRegex.FindStringSubmatch(fname); m != nil {
				neuronName = m[1]
				counter, _ = strconv.Atoi(m[2])
			} else if m := flatNeuronSimple.FindStringSubmatch(fname); m != nil {
				neuronName = m[1]
				if neuronName == "bomb" || strings.HasPrefix(neuronName, "dopamine") || strings.HasPrefix(neuronName, "memory") {
					continue
				}
			} else {
				continue
			}

			if existing, ok := neuronMap[neuronName]; ok {
				if counter > existing.Counter {
					existing.Counter = counter
				}
			} else {
				n := &Neuron{
					Name:     neuronName,
					Path:     neuronName,
					FullPath: filepath.Join(regionPath, neuronName),
					Depth:    0,
					Counter:  counter,
				}
				if fileInfo, err := os.Stat(nf); err == nil {
					n.ModTime = fileInfo.ModTime()
				} else {
					n.ModTime = time.Now()
				}
				if fname == "bomb.neuron" {
					n.HasBomb = true
					region.HasBomb = true
				} else if strings.HasPrefix(fname, "bomb_") {
					n.HasBomb = true
				}
				neuronMap[neuronName] = n
			}
		}

		// Walk for neuron folders — Axiom: Folder=Neuron, File=Trace
		filepath.Walk(regionPath, func(path string, info os.FileInfo, err error) error {
			if err != nil || !info.IsDir() || path == regionPath {
				return nil
			}

			baseName := filepath.Base(path)
			if strings.HasPrefix(baseName, "_") || strings.HasPrefix(baseName, ".") {
				if baseName == "_sandbox" {
					return nil
				}
				return filepath.SkipDir
			}

			relPath, _ := filepath.Rel(regionPath, path)
			depth := strings.Count(relPath, string(filepath.Separator))

			n, exists := neuronMap[relPath]
			if !exists {
				n = &Neuron{
					Name:     baseName,
					Path:     relPath,
					FullPath: path,
					Depth:    depth,
					ModTime:  info.ModTime(),
				}
				neuronMap[relPath] = n
			} else {
				n.FullPath = path
				n.Depth = depth
				if info.ModTime().After(n.ModTime) {
					n.ModTime = info.ModTime()
				}
			}

			neuronFiles, _ := filepath.Glob(filepath.Join(path, "*.neuron"))
			contraFiles, _ := filepath.Glob(filepath.Join(path, "*.contra"))
			neuronFiles = append(neuronFiles, contraFiles...)
			goalFiles, _ := filepath.Glob(filepath.Join(path, "*.goal"))

			for _, nf := range neuronFiles {
				fname := filepath.Base(nf)
				if nfInfo, err := os.Stat(nf); err == nil {
					if nfInfo.ModTime().After(n.ModTime) {
						n.ModTime = nfInfo.ModTime()
					}
				}

				if m := counterRegex.FindStringSubmatch(fname); m != nil {
					cnt, _ := strconv.Atoi(m[1])
					if cnt > n.Counter {
						n.Counter = cnt
					}
				}

				if m := dopamineRegex.FindStringSubmatch(fname); m != nil {
					cnt, _ := strconv.Atoi(m[1])
					n.Dopamine += cnt
				}

				if strings.HasSuffix(fname, ".contra") && region.Name != "brainstem" {
					base := strings.TrimSuffix(fname, ".contra")
					if cnt, err := strconv.Atoi(base); err == nil && cnt > n.Contra {
						n.Contra = cnt
					}
				}

				if fname == "bomb.neuron" {
					n.HasBomb = true
					region.HasBomb = true
				}
				if strings.HasPrefix(fname, "memory") {
					n.HasMemory = true
				}
			}

			if len(goalFiles) > 0 {
				n.HasGoal = true
				if content, err := os.ReadFile(goalFiles[0]); err == nil && len(content) > 0 {
					n.GoalText = strings.TrimSpace(string(content))
				}
			}

			geofenceFiles, _ := filepath.Glob(filepath.Join(path, "*.geofence"))
			if len(geofenceFiles) > 0 {
				if content, err := os.ReadFile(geofenceFiles[0]); err == nil && len(content) > 0 {
					n.Geofence = strings.TrimSpace(string(content))
				}
			}

			dormantFiles, _ := filepath.Glob(filepath.Join(path, "*.dormant"))
			if len(dormantFiles) > 0 {
				n.IsDormant = true
			}

			return nil
		})

		// Finalize compute elements and ensure deterministic order
		var paths []string
		for path := range neuronMap {
			paths = append(paths, path)
		}
		sort.Strings(paths)

		for _, path := range paths {
			n := neuronMap[path]
			n.Intensity = n.Counter - n.Contra + n.Dopamine
			totalSignals := n.Counter + n.Contra + n.Dopamine
			if totalSignals > 0 {
				n.Polarity = float64(n.Counter+n.Dopamine-n.Contra) / float64(totalSignals)
			} else {
				n.Polarity = 0.5
			}
			region.Neurons = append(region.Neurons, *n)
		}

		brain.Regions = append(brain.Regions, region)
	}

	// Sort regions by priority
	sort.Slice(brain.Regions, func(i, j int) bool {
		return brain.Regions[i].Priority < brain.Regions[j].Priority
	})

	return brain
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// SUBSUMPTION: Priority cascade + bomb detection
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
func runSubsumption(brain Brain) SubsumptionResult {
	result := SubsumptionResult{}
	blocked := false

	for _, region := range brain.Regions {
		activeNeurons := 0
		for _, n := range region.Neurons {
			if !n.IsDormant {
				activeNeurons++
				result.TotalCounter += n.Counter
			}
		}
		result.TotalNeurons += activeNeurons

		if blocked {
			result.BlockedRegions = append(result.BlockedRegions, region.Name)
			continue
		}

		if region.HasBomb {
			result.BombSource = region.Name
			result.BlockedRegions = append(result.BlockedRegions, region.Name+" [BOMB]")
			blocked = true

			// Physical Hook Trigger (e.g., ring alarm or strict kill process)
			// Triggered when a bomb is found in the geofenced context.
			triggerPhysicalHook(region.Name)
			continue
		}

		result.ActiveRegions = append(result.ActiveRegions, region)
		result.FiredNeurons += activeNeurons
	}

	return result
}

// emitRules is a compatibility wrapper for the tiered system.
// Returns bootstrap content (Tier 1) for GEMINI.md injection.
// The full tier system (index + per-region) is handled by writeAllTiers.
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
	// Walk up to find .gemini/GEMINI.md
	dir := filepath.Dir(brainRoot)
	for i := 0; i < 3; i++ {
		geminiPath := filepath.Join(dir, ".gemini", "GEMINI.md")
		if _, err := os.Stat(geminiPath); err == nil {
			doInject(geminiPath, rules)
			return
		}
		dir = filepath.Dir(dir)
	}

	// Try USERPROFILE
	home := os.Getenv("USERPROFILE")
	if home != "" {
		geminiPath := filepath.Join(home, ".gemini", "GEMINI.md")
		if _, err := os.Stat(geminiPath); err == nil {
			doInject(geminiPath, rules)
			return
		}
	}

	fmt.Println("[WARN] GEMINI.md not found, outputting to stdout:")
	fmt.Print(rules)
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

	err = os.WriteFile(geminiPath, []byte(content), 0644)
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
// WATCH: Monitor + auto-inject
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
func runWatch(brainRoot string) {
	lastHash := ""
	for {
		brain := scanBrain(brainRoot)
		result := runSubsumption(brain)
		hash := fmt.Sprintf("%d-%s-%d-%d",
			result.FiredNeurons, result.BombSource, result.TotalNeurons, result.TotalCounter)
		if hash != lastHash {
			lastHash = hash
			writeAllTiers(brainRoot)
			if result.BombSource != "" {
				fmt.Printf("[%s] 💀 BOMB in %s\n", time.Now().Format("15:04:05"), result.BombSource)
			} else {
				fmt.Printf("[%s] ✅ %d/%d neurons | activation: %d | brain_state.json updated\n",
					time.Now().Format("15:04:05"), result.FiredNeurons, result.TotalNeurons, result.TotalCounter)
			}
		}
		time.Sleep(2 * time.Second)
	}
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// DIAGNOSTICS
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

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// GROWTH ENGINE: Mechanical neuron creation & mutation
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// tokenize splits a snake_case neuron name into stemmed lowercase tokens
// "no_console_logging" → {"no", "console", "log"}
func tokenize(name string) []string {
	// 밑줄과 공백 모두 분리자로 처리
	normalized := strings.ReplaceAll(strings.ToLower(name), "_", " ")
	parts := strings.Fields(normalized) // Fields는 연속 공백도 처리
	var tokens []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		// Simple English stemming for merge matching
		p = stem(p)
		tokens = append(tokens, p)
	}
	return tokens
}

// stem applies minimal suffix stripping for merge matching
// Not a full Porter stemmer — just handles common AI naming patterns
func stem(word string) string {
	// Order matters: check longer suffixes first
	suffixes := []string{"ation", "ting", "ning", "ding", "ring", "sing", "ling", "ping", "ging", "ing", "ied", "ies", "ness", "ment", "able", "ible", "ful", "less", "ous", "ive", "ed"}
	for _, s := range suffixes {
		if len(word) > len(s)+2 && strings.HasSuffix(word, s) {
			return word[:len(word)-len(s)]
		}
	}
	// Trailing 's' (plural) — only if word is 4+ chars
	if len(word) >= 4 && strings.HasSuffix(word, "s") && !strings.HasSuffix(word, "ss") {
		return word[:len(word)-1]
	}
	return word
}

// jaccardSimilarity computes |A∩B| / |A∪B| between two token sets
func jaccardSimilarity(a, b []string) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	setA := make(map[string]bool)
	for _, t := range a {
		setA[t] = true
	}
	setB := make(map[string]bool)
	for _, t := range b {
		setB[t] = true
	}
	intersection := 0
	for t := range setA {
		if setB[t] {
			intersection++
		}
	}
	union := len(setA)
	for t := range setB {
		if !setA[t] {
			union++
		}
	}
	if union == 0 {
		return 0
	}
	return float64(intersection) / float64(union)
}

// growNeuron creates a new neuron folder with 1.neuron
// If a similar neuron already exists (Jaccard similarity >= 0.6), fire that instead (consolidation)
// Usage: neuronfs brain_v4 --grow cortex/frontend/coding/no_console_log
// Returns error instead of os.Exit so REST API won't crash
func growNeuron(brainRoot string, neuronPath string) error {
	// Normalize path separators
	neuronPath = strings.ReplaceAll(neuronPath, "/", string(filepath.Separator))
	fullPath := filepath.Join(brainRoot, neuronPath)

	// Validate region
	parts := strings.SplitN(neuronPath, string(filepath.Separator), 2)
	region := parts[0]
	if _, ok := regionPriority[region]; !ok {
		err := fmt.Errorf("invalid region: %s (valid: brainstem,limbic,hippocampus,sensors,cortex,ego,prefrontal)", region)
		fmt.Printf("[FATAL] %v\n", err)
		return err
	}

	// Check if neuron already exists (exact match)
	if info, err := os.Stat(fullPath); err == nil && info.IsDir() {
		fmt.Printf("[SKIP] Neuron already exists: %s\n", neuronPath)
		return nil
	}

	// ── Synaptic Consolidation: merge similar neurons ──
	// Tokenize the new neuron's leaf name
	leafName := filepath.Base(neuronPath)
	newTokens := tokenize(leafName)
	if len(newTokens) > 0 {
		// Walk the same region to find similar neurons
		regionPath := filepath.Join(brainRoot, region)
		bestMatch := ""
		bestSimilarity := 0.0

		filepath.Walk(regionPath, func(path string, info os.FileInfo, err error) error {
			if err != nil || !info.IsDir() || path == regionPath {
				return nil
			}
			// Check if this is a neuron folder (has .neuron files)
			neuronFiles, _ := filepath.Glob(filepath.Join(path, "*.neuron"))
			if len(neuronFiles) == 0 {
				return nil
			}
			existingLeaf := filepath.Base(path)
			existingTokens := tokenize(existingLeaf)
			sim := jaccardSimilarity(newTokens, existingTokens)
			if sim > bestSimilarity {
				bestSimilarity = sim
				rel, _ := filepath.Rel(brainRoot, path)
				bestMatch = rel
			}
			return nil
		})

		if bestSimilarity >= 0.6 && bestMatch != "" {
			fmt.Printf("[MERGE] 🔗 '%s' ≈ '%s' (%.0f%% similar) → firing existing\n",
				neuronPath, bestMatch, bestSimilarity*100)
			fireNeuron(brainRoot, bestMatch)
			logEpisode(brainRoot, "MERGE", fmt.Sprintf("%s → %s (%.0f%%)", neuronPath, bestMatch, bestSimilarity*100))
			return nil
		}
	}

	// Create folder
	if err := os.MkdirAll(fullPath, 0755); err != nil {
		fmt.Printf("[ERROR] mkdir: %v\n", err)
		return err
	}

	// Create 1.neuron
	neuronFile := filepath.Join(fullPath, "1.neuron")
	if err := os.WriteFile(neuronFile, []byte{}, 0644); err != nil {
		fmt.Printf("[ERROR] create trace: %v\n", err)
		return err
	}

	fmt.Printf("[GROW] ✅ %s → 1.neuron\n", neuronPath)

	// Log to hippocampus
	logEpisode(brainRoot, "GROW", neuronPath)

	// Mark dirty — periodic loop will handle injection
	markBrainDirty()
	return nil
}

// fireNeuron increments the counter of an existing neuron
// Usage: neuronfs brain_v4 --fire cortex/frontend/coding/no_console_log
func fireNeuron(brainRoot string, neuronPath string) {
	neuronPath = strings.ReplaceAll(neuronPath, "/", string(filepath.Separator))
	fullPath := filepath.Join(brainRoot, neuronPath)

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		fmt.Printf("[WARN] Neuron not found: %s — auto-growing...\n", neuronPath)
		growNeuron(brainRoot, neuronPath)
		return
	}

	// Find current counter
	currentCounter := 0
	currentFile := ""
	entries, _ := os.ReadDir(fullPath)
	for _, e := range entries {
		if m := counterRegex.FindStringSubmatch(e.Name()); m != nil {
			n, _ := strconv.Atoi(m[1])
			if n > currentCounter {
				currentCounter = n
				currentFile = filepath.Join(fullPath, e.Name())
			}
		}
	}

	newCounter := currentCounter + 1

	// Delete old counter file
	if currentFile != "" {
		if err := os.Remove(currentFile); err != nil {
			fmt.Fprintf(os.Stderr, "[WARN] fire: old counter cleanup: %v\n", err)
		}
	}

	// Create new counter file
	newFile := filepath.Join(fullPath, fmt.Sprintf("%d.neuron", newCounter))
	if err := os.WriteFile(newFile, []byte{}, 0644); err != nil {
		fmt.Printf("[ERROR] fire: %v\n", err)
		return
	}

	fmt.Printf("[FIRE] 🔥 %s → %d → %d\n", neuronPath, currentCounter, newCounter)

	logEpisode(brainRoot, "FIRE", fmt.Sprintf("%s (%d→%d)", neuronPath, currentCounter, newCounter))
	markBrainDirty()
}

// rollbackNeuron decrements the counter of an existing neuron
// Minimum counter is 1 (won't go below). Returns error for API usage.
// Usage: neuronfs brain_v4 --rollback cortex/frontend/coding/no_console_log
func rollbackNeuron(brainRoot string, neuronPath string) error {
	neuronPath = strings.ReplaceAll(neuronPath, "/", string(filepath.Separator))
	fullPath := filepath.Join(brainRoot, neuronPath)

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		err := fmt.Errorf("neuron not found: %s", neuronPath)
		fmt.Printf("[ROLLBACK] ❌ %v\n", err)
		return err
	}

	// Find current counter
	currentCounter := 0
	currentFile := ""
	entries, _ := os.ReadDir(fullPath)
	for _, e := range entries {
		if m := counterRegex.FindStringSubmatch(e.Name()); m != nil {
			n, _ := strconv.Atoi(m[1])
			if n > currentCounter {
				currentCounter = n
				currentFile = filepath.Join(fullPath, e.Name())
			}
		}
	}

	if currentCounter <= 1 {
		fmt.Printf("[ROLLBACK] ⚠️ %s counter already at minimum (%d)\n", neuronPath, currentCounter)
		return fmt.Errorf("counter at minimum: %d", currentCounter)
	}

	newCounter := currentCounter - 1

	// Delete old counter file
	if currentFile != "" {
		if err := os.Remove(currentFile); err != nil {
			fmt.Fprintf(os.Stderr, "[WARN] rollback: old counter cleanup: %v\n", err)
		}
	}

	// Create new counter file
	newFile := filepath.Join(fullPath, fmt.Sprintf("%d.neuron", newCounter))
	if err := os.WriteFile(newFile, []byte{}, 0644); err != nil {
		fmt.Printf("[ERROR] rollback: %v\n", err)
		return err
	}

	fmt.Printf("[ROLLBACK] ⏪ %s → %d → %d\n", neuronPath, currentCounter, newCounter)

	logEpisode(brainRoot, "ROLLBACK", fmt.Sprintf("%s (%d→%d)", neuronPath, currentCounter, newCounter))
	markBrainDirty()
	return nil
}

// signalNeuron adds dopamine/bomb/memory signal to a neuron
// Returns error instead of os.Exit so REST API won't crash
func signalNeuron(brainRoot string, neuronPath string, sigType string) error {
	neuronPath = strings.ReplaceAll(neuronPath, "/", string(filepath.Separator))
	fullPath := filepath.Join(brainRoot, neuronPath)

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		err := fmt.Errorf("neuron not found: %s", neuronPath)
		fmt.Printf("[FATAL] %v\n", err)
		return err
	}

	switch sigType {
	case "dopamine":
		nextDopa := 1
		entries, _ := os.ReadDir(fullPath)
		for _, e := range entries {
			if m := dopamineRegex.FindStringSubmatch(e.Name()); m != nil {
				n, _ := strconv.Atoi(m[1])
				if n >= nextDopa {
					nextDopa = n + 1
				}
			}
		}
		df := filepath.Join(fullPath, fmt.Sprintf("dopamine%d.neuron", nextDopa))
		os.WriteFile(df, []byte{}, 0644)
		fmt.Printf("[SIGNAL] 🟢 dopamine%d → %s\n", nextDopa, neuronPath)

	case "bomb":
		bf := filepath.Join(fullPath, "bomb.neuron")
		os.WriteFile(bf, []byte{}, 0644)
		fmt.Printf("[SIGNAL] 💣 BOMB → %s\n", neuronPath)

	case "memory":
		nextMem := 1
		memRegex := regexp.MustCompile(`^memory(\d+)\.neuron$`)
		entries, _ := os.ReadDir(fullPath)
		for _, e := range entries {
			if m := memRegex.FindStringSubmatch(e.Name()); m != nil {
				n, _ := strconv.Atoi(m[1])
				if n >= nextMem {
					nextMem = n + 1
				}
			}
		}
		mf := filepath.Join(fullPath, fmt.Sprintf("memory%d.neuron", nextMem))
		os.WriteFile(mf, []byte{}, 0644)
		fmt.Printf("[SIGNAL] 📝 memory%d → %s\n", nextMem, neuronPath)

	default:
		err := fmt.Errorf("unknown signal type: %s (use dopamine|bomb|memory)", sigType)
		fmt.Printf("[FATAL] %v\n", err)
		return err
	}

	logEpisode(brainRoot, "SIGNAL:"+sigType, neuronPath)
	markBrainDirty()
	return nil
}

// runDecay moves neurons untouched for N days to dormant state
// Usage: neuronfs brain_v4 --decay 30
func runDecay(brainRoot string, days int) {
	cutoff := time.Now().AddDate(0, 0, -days)
	decayed := 0
	total := 0

	for _, regionName := range []string{"brainstem", "limbic", "hippocampus", "sensors", "cortex", "ego", "prefrontal"} {
		regionPath := filepath.Join(brainRoot, regionName)
		if _, err := os.Stat(regionPath); os.IsNotExist(err) {
			continue
		}

		filepath.Walk(regionPath, func(path string, info os.FileInfo, err error) error {
			if err != nil || !info.IsDir() || path == regionPath {
				return nil
			}

			// Check if this is a neuron folder (has .neuron files)
			neuronFiles, _ := filepath.Glob(filepath.Join(path, "*.neuron"))
			if len(neuronFiles) == 0 {
				return nil
			}
			total++

			// Skip if already dormant
			dormantFiles, _ := filepath.Glob(filepath.Join(path, "*.dormant"))
			if len(dormantFiles) > 0 {
				return nil
			}

			// Find the most recent .neuron file modification time
			var newestMod time.Time
			for _, nf := range neuronFiles {
				fi, err := os.Stat(nf)
				if err == nil && fi.ModTime().After(newestMod) {
					newestMod = fi.ModTime()
				}
			}

			if !newestMod.IsZero() && newestMod.Before(cutoff) {
				// Mark as dormant
				df := filepath.Join(path, "decay.dormant")
				os.WriteFile(df, []byte(fmt.Sprintf("Decayed: %s\nLast active: %s\nThreshold: %d days\n",
					time.Now().Format("2006-01-02"),
					newestMod.Format("2006-01-02"),
					days)), 0644)

				relPath, _ := filepath.Rel(brainRoot, path)
				ageDays := int(time.Since(newestMod).Hours() / 24)
				fmt.Printf("[DECAY] 💤 %s (inactive %d days)\n", relPath, ageDays)
				decayed++
			}

			return nil
		})
	}

	fmt.Printf("[DECAY] Scanned %d neurons, decayed %d (threshold: %d days)\n", total, decayed, days)

	if decayed > 0 {
		logEpisode(brainRoot, "DECAY", fmt.Sprintf("%d neurons dormant (>%d days)", decayed, days))
		markBrainDirty()
	}
}

// logEpisode records an event in hippocampus/session_log
// Circular buffer: keeps only the most recent 100 episodes
const maxEpisodes = 100

// logEpisode writes an event log to the hippocampus memory store.
func logEpisode(brainRoot string, event string, detail string) {
	logDir := filepath.Join(brainRoot, "hippocampus", "session_log")
	os.MkdirAll(logDir, 0755)

	memRegex := regexp.MustCompile(`^memory(\d+)\.neuron$`)
	entries, _ := os.ReadDir(logDir)

	// Collect all memory files with their numbers
	type memEntry struct {
		num  int
		name string
	}
	var mems []memEntry
	for _, e := range entries {
		if m := memRegex.FindStringSubmatch(e.Name()); m != nil {
			n, _ := strconv.Atoi(m[1])
			mems = append(mems, memEntry{num: n, name: e.Name()})
		}
	}

	// Sort by number ascending
	sort.Slice(mems, func(i, j int) bool { return mems[i].num < mems[j].num })

	// Evict oldest if at limit
	if len(mems) >= maxEpisodes {
		evictCount := len(mems) - maxEpisodes + 1
		for i := 0; i < evictCount; i++ {
			os.Remove(filepath.Join(logDir, mems[i].name))
		}
		fmt.Printf("[MEMORY] 🗑️ Evicted %d old episodes (circular buffer %d)\n", evictCount, maxEpisodes)
	}

	// Find next number
	nextN := 1
	if len(mems) > 0 {
		nextN = mems[len(mems)-1].num + 1
	}

	content := fmt.Sprintf("%s | %s | %s\n", time.Now().Format("2006-01-02T15:04:05"), event, detail)
	memFile := filepath.Join(logDir, fmt.Sprintf("memory%d.neuron", nextN))
	os.WriteFile(memFile, []byte(content), 0644)
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// DIRTY FLAG + BATCH INJECTION
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
var (
	brainDirty    bool
	brainDirtyMu  sync.Mutex
	lastMountHash string
	triggerChan   = make(chan struct{}, 1)
)

// markBrainDirty signals that the brain state has changed and needs an event broadcast.
func markBrainDirty() {
	brainDirtyMu.Lock()
	brainDirty = true
	brainDirtyMu.Unlock()

	// Non-blocking trigger
	select {
	case triggerChan <- struct{}{}:
	default:
	}
}

// consumeDirty checks and clears the brain's dirty state flag.
func consumeDirty() bool {
	brainDirtyMu.Lock()
	defer brainDirtyMu.Unlock()
	if brainDirty {
		brainDirty = false
		return true
	}
	return false
}

// computeMountHash returns a hash of the current mount set (neuron IDs + counters that are mounted)
func computeMountHash(brainRoot string) string {
	brain := scanBrain(brainRoot)
	result := runSubsumption(brain)
	var parts []string
	for _, region := range result.ActiveRegions {
		for _, n := range region.Neurons {
			if n.IsDormant {
				continue
			}
			// Include neurons that would be mounted
			if n.Counter >= emitThreshold || time.Since(n.ModTime).Hours() < float64(spotlightDays*24) {
				parts = append(parts, fmt.Sprintf("%s:%d", n.Path, n.Counter))
			}
		}
	}
	sort.Strings(parts)
	return fmt.Sprintf("%x", len(parts)) + "|" + strings.Join(parts, ",")
}

// autoReinject checks mount set hash and only writes if changed
func autoReinject(brainRoot string) {
	newHash := computeMountHash(brainRoot)
	if newHash == lastMountHash {
		return // mount set unchanged, skip injection
	}
	lastMountHash = newHash
	writeAllTiers(brainRoot)
	fmt.Printf("[INJECT] ♻️  Mount set changed → GEMINI.md updated\n")
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// INBOX PROCESSOR: AI tool call → _inbox → neurons
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// inboxEntry represents a correction or insight from AI or auto-accept
type inboxEntry struct {
	Ts         string `json:"ts"`
	Type       string `json:"type"` // "correction" | "insight"
	Text       string `json:"text"`
	Source     string `json:"source"`      // "ai" | "auto-accept"
	Path       string `json:"path"`        // optional: pre-computed neuron path
	CounterAdd int    `json:"counter_add"` // optional: how much to add
	Author     string `json:"author"`      // optional: explicit author mapping
}

// processInbox reads _inbox/corrections.jsonl, creates/fires neurons, then clears
func processInbox(brainRoot string) {
	inboxPath := filepath.Join(brainRoot, "_inbox", "corrections.jsonl")
	data, err := os.ReadFile(inboxPath)
	if err != nil {
		return // no inbox file = nothing to process
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
		return
	}

	processed := 0
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var entry inboxEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			fmt.Printf("[INBOX] ⚠️ parse error: %s\n", line)
			continue
		}

		// Determine neuron path
		neuronPath := entry.Path
		if neuronPath == "" {
			// Auto-generate path from text if not provided
			// Simple heuristic: cortex/_inbox_pending/<sanitized_text>
			sanitized := strings.Map(func(r rune) rune {
				if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') ||
					r == '_' || r == '-' ||
					(r >= 0xAC00 && r <= 0xD7AF) || // 한글 음절
					(r >= 0x3131 && r <= 0x318E) || // 한글 자모
					(r >= 0x4E00 && r <= 0x9FFF) { // 한자 CJK
					return r
				}
				return '_'
			}, strings.ReplaceAll(entry.Text, " ", "_"))
			if len(sanitized) > 60 {
				sanitized = sanitized[:60]
			}
			neuronPath = "hippocampus/_inbox_pending/" + sanitized
		}

		// Security: Basic Prompt Injection & Path Traversal Defense
		if strings.Contains(neuronPath, "..") || strings.Contains(neuronPath, `\`) ||
			strings.Contains(neuronPath, "$") || strings.Contains(neuronPath, "&") ||
			strings.Contains(neuronPath, "|") || strings.Contains(neuronPath, ">") {
			fmt.Printf("[SECURITY] 🛡️ Injection blocked: %s\n", neuronPath)
			continue
		}

		// 기계적 칭찬(Dopamine Inflation) 필터링
		isPraise := false
		if entry.Type == "correction" && entry.Text == "PD칭찬" {
			isPraise = true
		}
		praiseRegex := regexp.MustCompile(`(?i)(칭찬|잘\s*쓰셨습니다|좋아|훌륭|완벽|최고)`)
		if praiseRegex.MatchString(entry.Text) || strings.Contains(strings.ToLower(neuronPath), "dopamine") {
			isPraise = true
		}

		if isPraise {
			authorId := entry.Author
			if authorId == "" {
				authorId = entry.Source
			}
			authorId = strings.ToLower(authorId)

			if authorId != "pm" && authorId != "basement_admin" && !strings.Contains(authorId, "pd") {
				fmt.Printf("[INBOX] 🛡️ 도파민 인플레이션 차단 (침해자: %s): %s\n", authorId, entry.Text)
				continue
			}
			// PM 칭찬은 바로 도파민 발화
			fullPath := filepath.Join(brainRoot, strings.ReplaceAll(neuronPath, "/", string(filepath.Separator)))
			_ = os.MkdirAll(fullPath, 0755)
			_ = signalNeuron(brainRoot, neuronPath, "dopamine")
			fmt.Printf("[INBOX] 🟢 PM 칭찬 확인 — 도파민 배포: %s\n", neuronPath)
			processed++
			continue
		}

		// Determine action
		counterAdd := entry.CounterAdd
		if counterAdd <= 0 {
			counterAdd = 1
		}

		// Check if neuron exists → fire, else → grow
		fullPath := filepath.Join(brainRoot, strings.ReplaceAll(neuronPath, "/", string(filepath.Separator)))
		if info, err := os.Stat(fullPath); err == nil && info.IsDir() {
			// Exists → fire N times
			for i := 0; i < counterAdd; i++ {
				fireNeuron(brainRoot, neuronPath)
			}
			fmt.Printf("[INBOX] 🔥 fire %s (×%d)\n", neuronPath, counterAdd)
		} else {
			// New → grow
			if err := growNeuron(brainRoot, neuronPath); err != nil {
				fmt.Printf("[INBOX] ❌ grow failed: %s — %v\n", neuronPath, err)
				continue
			}
			// Fire additional times if counter_add > 1
			for i := 1; i < counterAdd; i++ {
				fireNeuron(brainRoot, neuronPath)
			}
			fmt.Printf("[INBOX] 🌱 grow %s (counter=%d)\n", neuronPath, counterAdd)
		}
		processed++
	}

	if processed > 0 {
		// Clear inbox
		os.WriteFile(inboxPath, []byte{}, 0644)
		markBrainDirty()
		fmt.Printf("[INBOX] ✅ %d entries processed, inbox cleared\n", processed)
	}
}

// runInjectionLoop uses fsnotify and channels for event-driven, real-time updates
func runInjectionLoop(brainRoot string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Printf("[ERROR] fsnotify: %v\n", err)
		return
	}
	defer watcher.Close()

	inboxDir := filepath.Join(brainRoot, "_inbox")
	os.MkdirAll(inboxDir, 0755)
	if err := watcher.Add(inboxDir); err != nil {
		fmt.Printf("[ERROR] fsnotify watch _inbox: %v\n", err)
	}

	debounceDuration := 100 * time.Millisecond
	var timer *time.Timer

	executeUpdate := func() {
		processInbox(brainRoot)
		if consumeDirty() {
			autoReinject(brainRoot)
		}
	}

	queueUpdate := func() {
		if timer != nil {
			timer.Stop()
		}
		timer = time.AfterFunc(debounceDuration, executeUpdate)
	}

	// Initial check
	queueUpdate()

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
				queueUpdate()
			}
		case <-triggerChan:
			queueUpdate()
		case <-time.After(5 * time.Minute):
			// fallback heartbeat
			queueUpdate()
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			fmt.Printf("[ERROR] watcher: %v\n", err)
		}
	}
}

// gitSnapshot takes a single git snapshot of the brain state
// Called only during idle via --snapshot flag (not on every fire/grow)
// Lifecycle: active → changes accumulate → idle → groq analysis → snapshot
func gitSnapshot(brainRoot string) {
	// Check if git is available
	if _, err := exec.LookPath("git"); err != nil {
		fmt.Println("[GIT] git not found, skipping snapshot")
		return
	}

	// Auto-init if not a git repo
	gitDir := filepath.Join(brainRoot, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		cmd := exec.Command("git", "init")
		cmd.Dir = brainRoot
		if err := cmd.Run(); err != nil {
			fmt.Printf("[GIT] ⚠️ init failed: %v\n", err)
			return
		}
		gitignore := filepath.Join(brainRoot, ".gitignore")
		os.WriteFile(gitignore, []byte("*.dormant\n"), 0644)
		fmt.Printf("[GIT] 📂 Initialized git repo in %s\n", brainRoot)
	}

	// Check for changes
	statusCmd := exec.Command("git", "status", "--porcelain")
	statusCmd.Dir = brainRoot
	out, err := statusCmd.Output()
	if err != nil || len(out) == 0 {
		fmt.Println("[GIT] No changes to snapshot")
		return
	}

	// Stage all
	addCmd := exec.Command("git", "add", "-A")
	addCmd.Dir = brainRoot
	if err := addCmd.Run(); err != nil {
		fmt.Printf("[GIT] ⚠️ add failed: %v\n", err)
		return
	}

	// Build commit message from current brain state
	brain := scanBrain(brainRoot)
	result := runSubsumption(brain)
	changes := strings.Count(string(out), "\n")
	timestamp := time.Now().Format("01-02 15:04")
	msg := fmt.Sprintf("[%s] %d neurons, act:%d, Δ%d files",
		timestamp, result.TotalNeurons, result.TotalCounter, changes)

	commitCmd := exec.Command("git", "commit", "-m", msg, "--no-verify")
	commitCmd.Dir = brainRoot
	if err := commitCmd.Run(); err != nil {
		return
	}
	fmt.Printf("[GIT] 📸 %s\n", msg)

	// ── git diff 진화판정: 뉴런 순감소이면 자동 rollback ──
	diffCmd := exec.Command("git", "diff", "HEAD~1", "--stat")
	diffCmd.Dir = brainRoot
	diffOut, err := diffCmd.Output()
	if err == nil {
		diffStr := string(diffOut)
		deletions := strings.Count(diffStr, "deletion")
		insertions := strings.Count(diffStr, "insertion")
		if deletions > insertions*2 && deletions > 5 {
			// 삭제가 삽입의 2배 이상이고 5건 초과이면 퇴화로 판정
			fmt.Printf("[GIT] ⚠️ 퇴화 감지 (삭제 %d > 삽입 %d×2) — 자동 rollback\n", deletions, insertions)
			revertCmd := exec.Command("git", "revert", "HEAD", "--no-edit")
			revertCmd.Dir = brainRoot
			if err := revertCmd.Run(); err != nil {
				fmt.Printf("[GIT] ❌ rollback 실패: %v\n", err)
			} else {
				fmt.Println("[GIT] ✅ 퇴화 commit이 revert 되었습니다")
			}
		} else {
			fmt.Printf("[GIT] ✅ 진화 판정 통과 (ins:%d, del:%d)\n", insertions, deletions)
		}
	}
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// IDLE ENGINE: Auto evolve → snapshot → NAS sync
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

const (
	idleThresholdMinutes = 5  // minutes of no API activity → trigger idle cycle
	idleCheckInterval    = 30 // seconds between idle checks
)

var (
	lastAPIActivity   time.Time
	lastAPIActivityMu sync.Mutex
	idleEvolveRunning bool

	// Heartbeat control
	heartbeatEnabled  = true // 기본 ON
	heartbeatInterval = 10   // 초 (켜져 있을 때 즉각 반응)
	heartbeatCooldown = 3    // 분 (주입 후 쿨다운)
	heartbeatMu       sync.Mutex
)

// touchActivity updates the system's last recorded activity timestamp.
func touchActivity() {
	lastAPIActivityMu.Lock()
	lastAPIActivity = time.Now()
	lastAPIActivityMu.Unlock()
}

// getLastActivity returns the latest timestamp among multiple tracked activity records.
func getLastActivity() time.Time {
	lastAPIActivityMu.Lock()
	defer lastAPIActivityMu.Unlock()
	return lastAPIActivity
}

// runIdleLoop runs in a goroutine, checking for idle state periodically.
// When idle is detected: evolve (if GROQ_API_KEY set) → snapshot → NAS sync
func runIdleLoop(brainRoot string) {
	lastEvolveTime := time.Now()

	for {
		time.Sleep(time.Duration(idleCheckInterval) * time.Second)

		lastAct := getLastActivity()
		idleDuration := time.Since(lastAct)

		// Need at least idleThresholdMinutes of idle AND at least 30 min since last evolve
		if idleDuration < time.Duration(idleThresholdMinutes)*time.Minute {
			continue
		}
		if time.Since(lastEvolveTime) < 30*time.Minute {
			continue
		}
		if idleEvolveRunning {
			continue
		}

		idleEvolveRunning = true
		fmt.Printf("\n[IDLE] 💤 %s idle detected — starting autonomous cycle...\n", idleDuration.Round(time.Second))

		// 1. Evolve (if GROQ_API_KEY available)
		apiKey := os.Getenv("GROQ_API_KEY")
		if apiKey != "" {
			fmt.Println("[IDLE] 🧬 Running Groq evolve...")
			runEvolve(brainRoot, false)
		} else {
			fmt.Println("[IDLE] ⚠️  GROQ_API_KEY not set — skipping evolve")
		}

		// 2. Auto-decay (mark neurons untouched for 30+ days as dormant)
		fmt.Println("[IDLE] 💤 Running auto-decay (30 days)...")
		runDecay(brainRoot, 30)

		// 3. Dedup (merge semantically similar neurons, Jaccard >= 0.6)
		fmt.Println("[IDLE] 🔀 Running dedup (Jaccard similarity)...")
		deduplicateNeurons(brainRoot)

		// 4. Growth tracking (뇌 성장 이력 추적)
		brain := scanBrain(brainRoot)
		result := runSubsumption(brain)
		growthLogDir := filepath.Join(brainRoot, "hippocampus", "session_log")
		os.MkdirAll(growthLogDir, 0755)
		growthLogFile := filepath.Join(growthLogDir, "growth.log")
		entry := fmt.Sprintf("%s: neurons=%d, activation=%d, regions=%d\n",
			time.Now().Format("2006-01-02_15:04"), result.TotalNeurons, result.TotalCounter, len(result.ActiveRegions))
		f, _ := os.OpenFile(growthLogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if f != nil {
			f.WriteString(entry)
			f.Close()
		}
		fmt.Printf("[GROWTH] 📈 %s", entry)

		// 5. Git snapshot
		fmt.Println("[IDLE] 📸 Git snapshot...")
		gitSnapshot(brainRoot)

		// 3. NAS sync (if Z: available)
		nasTarget := `Z:\VOL1\VGVR\BRAIN\LW\system\neurons\brain_v4`
		if _, err := os.Stat(nasTarget); err == nil {
			fmt.Println("[IDLE] 📡 NAS sync...")
			syncCmd := exec.Command("robocopy", brainRoot, nasTarget, "/MIR", "/XD", ".git", "/XF", "*.dormant", "/NFL", "/NDL", "/NP", "/NJH", "/NJS")
			if out, err := syncCmd.CombinedOutput(); err != nil {
				// robocopy exit code 1 = files copied (success), only >=8 is error
				exitCode := syncCmd.ProcessState.ExitCode()
				if exitCode >= 8 {
					fmt.Printf("[IDLE] ❌ NAS sync error (exit %d): %s\n", exitCode, string(out))
				} else {
					fmt.Printf("[IDLE] ✅ NAS synced (exit %d)\n", exitCode)
				}
			} else {
				fmt.Println("[IDLE] ✅ NAS synced (no changes)")
			}
		} else {
			fmt.Println("[IDLE] ⚠️  NAS Z: not available — skipping sync")
		}

		lastEvolveTime = time.Now()
		idleEvolveRunning = false
		fmt.Printf("[IDLE] ✅ Autonomous cycle complete at %s\n\n", lastEvolveTime.Format("15:04:05"))
	}
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// DEDUP: 중복 뉴런 폴더 병합 (카운터 합산)
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// deduplicateNeurons scans brain for semantically similar neuron folders
// and merges them: keeps the deeper/higher-counter one, sums counters +1 bonus
// Uses Jaccard similarity (>= 0.6) on tokenized+stemmed folder names
func deduplicateNeurons(brainRoot string) {
	brain := scanBrain(brainRoot)

	type neuronRef struct {
		fullPath string
		counter  int
		region   string
		relPath  string
		tokens   []string
		depth    int
	}

	// Collect all active leaf neurons
	var allRefs []neuronRef
	for _, region := range brain.Regions {
		if region.Name == "brainstem" {
			continue // brainstem은 읽기 전용 — 건드리지 않음
		}
		for _, n := range region.Neurons {
			if n.IsDormant {
				continue
			}
			leafName := filepath.Base(n.FullPath)
			tokens := tokenize(leafName)
			allRefs = append(allRefs, neuronRef{
				fullPath: n.FullPath,
				counter:  n.Counter,
				region:   region.Name,
				relPath:  n.Path,
				tokens:   tokens,
				depth:    n.Depth,
			})
		}
	}

	// Compare all pairs (O(n²) — 200 뉴런이면 ~20,000 비교, 무시할 수준)
	merged := make(map[int]bool) // index of already-merged victims
	mergeCount := 0

	for i := 0; i < len(allRefs); i++ {
		if merged[i] {
			continue
		}
		for j := i + 1; j < len(allRefs); j++ {
			if merged[j] {
				continue
			}

			sim := jaccardSimilarity(allRefs[i].tokens, allRefs[j].tokens)
			if sim < 0.6 {
				continue
			}

			// 유사도 0.6 이상 — 병합 대상
			// 생존자: 더 깊거나 카운터가 높은 쪽
			survivor := &allRefs[i]
			victim := &allRefs[j]
			if victim.depth > survivor.depth || (victim.depth == survivor.depth && victim.counter > survivor.counter) {
				survivor, victim = victim, survivor
				// swap indices for merged tracking
				if survivor == &allRefs[j] {
					merged[i] = true
				}
			} else {
				merged[j] = true
			}

			// 카운터 합산 + 보너스 (+1)
			totalCounter := survivor.counter + victim.counter + 1
			fmt.Printf("[DEDUP] 🔀 병합 (sim=%.2f): %s/%s (%d) ← %s/%s (%d) → %d\n",
				sim,
				survivor.region, filepath.Base(survivor.fullPath), survivor.counter,
				victim.region, filepath.Base(victim.fullPath), victim.counter,
				totalCounter)

			// 생존자 카운터 갱신
			surviveFiles, _ := filepath.Glob(filepath.Join(survivor.fullPath, "*.neuron"))
			for _, f := range surviveFiles {
				base := filepath.Base(f)
				if counterRegex.MatchString(base) {
					os.Remove(f)
				}
			}
			newCounterFile := filepath.Join(survivor.fullPath, fmt.Sprintf("%d.neuron", totalCounter))
			os.WriteFile(newCounterFile, []byte(""), 0644)

			// victim의 dopamine/memory 시그널을 생존자로 이동
			victimFiles, _ := filepath.Glob(filepath.Join(victim.fullPath, "*.neuron"))
			for _, f := range victimFiles {
				base := filepath.Base(f)
				if strings.HasPrefix(base, "dopamine") || strings.HasPrefix(base, "memory") {
					destFile := filepath.Join(survivor.fullPath, base)
					if _, err := os.Stat(destFile); os.IsNotExist(err) {
						os.Rename(f, destFile)
					}
				}
			}

			// victim 폴더 삭제
			os.RemoveAll(victim.fullPath)
			survivor.counter = totalCounter
			mergeCount++
		}
	}

	if mergeCount > 0 {
		fmt.Printf("[DEDUP] ✅ %d 건 중복 뉴런 병합 완료 (카운터 합산+보너스)\n", mergeCount)
		writeAllTiers(brainRoot)
	} else {
		fmt.Println("[DEDUP] ✓ 중복 뉴런 없음")
	}
}

// runHeartbeatLoop drives the conscious mind (AI) forward without human input.
// 유휴 판정: auto-accept의 전사 로그(_transcripts/) mtime을 확인
// → 전사가 멈추면 AI가 정지한 것 → Todo 주입
func runHeartbeatLoop(brainRoot string) {
	pulseScript := filepath.Join(filepath.Dir(brainRoot), "runtime", "pulse.mjs")
	todoDir := filepath.Join(brainRoot, "prefrontal", "todo")

	// Antigravity session logs — home dir, not brainRoot parent
	homeDir, _ := os.UserHomeDir()
	antigravityBrain := filepath.Join(homeDir, ".gemini", "antigravity", "brain")
	var lastLogFile string
	var lastLogSize int64

	lastInjection := time.Time{} // 마지막 주입 시각

	for {
		heartbeatMu.Lock()
		enabled := heartbeatEnabled
		interval := heartbeatInterval
		cooldown := time.Duration(heartbeatCooldown) * time.Minute
		heartbeatMu.Unlock()

		time.Sleep(time.Duration(interval) * time.Second)

		if !enabled {
			continue
		}

		// 쿨다운: 주입 후 N분 이내면 스킵
		if !lastInjection.IsZero() && time.Since(lastInjection) < cooldown {
			continue
		}

		// ── Priority 0: Memory Observer (전사 기반 뉴런화) ──
		var nextPrompt string
		if entries, err := os.ReadDir(antigravityBrain); err == nil {
			// Find latest session overview.txt
			var latestFile string
			var latestTime int64
			for _, s := range entries {
				if !s.IsDir() {
					continue
				}
				op := filepath.Join(antigravityBrain, s.Name(), ".system_generated", "logs", "overview.txt")
				if info, err := os.Stat(op); err == nil {
					if info.ModTime().UnixMilli() > latestTime {
						latestTime = info.ModTime().UnixMilli()
						latestFile = op
					}
				}
			}

			if latestFile != "" {
				info, _ := os.Stat(latestFile)
				currentSize := info.Size()
				const chunkSize int64 = 3000

				// Reset if file changed
				if lastLogFile != latestFile {
					lastLogFile = latestFile
					lastLogSize = max64(0, currentSize-chunkSize)
				}

				if currentSize-lastLogSize >= chunkSize {
					f, err := os.Open(latestFile)
					if err == nil {
						buf := make([]byte, currentSize-lastLogSize)
						f.ReadAt(buf, lastLogSize)
						f.Close()
						lastLogSize = currentSize

						recentLogs := string(buf)
						nextPrompt = fmt.Sprintf(`[MEMORY_OBSERVER %s] 아래는 최근 시스템 대화 로그 (%d바이트)이다.
---
%s
---
위 로그에서 뉴런화되지 않은 중요 아키텍처 결정(암묵적 룰, 해결책)을 찾아라.
발견되면 [Folder-as-Neuron] 온톨로지에 맞춰:
1. mkdir brain_v4/[region]/[카테고리]/[행동_강령]
2. 최종 리프 폴더에만 touch 1.neuron
새로운 룰이 없으면 "로그 스캔 완료: 추출할 뉴런 없음"이라 보고하라.`,
							time.Now().Format("15:04"), len(buf), recentLogs)
						fmt.Printf("[HEARTBEAT] 📡 MEMORY_OBSERVER: %d bytes from session log\n", len(buf))
					}
				} else {
					fmt.Printf("[HEARTBEAT] 📡 로그 청크 미달 (%d/%d bytes)\n", currentSize-lastLogSize, chunkSize)
				}
			}
		}

		// ── Priority 1: Pending reports (적층형 보고 큐) ──
		if nextPrompt == "" {
		reportsDir := filepath.Join(brainRoot, "_inbox", "reports")
		if reportEntries, err := os.ReadDir(reportsDir); err == nil && len(reportEntries) > 0 {
			// Find highest priority: urgent > normal > low
			priorityOrder := []string{"urgent", "normal", "low"}
			for _, prio := range priorityOrder {
				for _, e := range reportEntries {
					if strings.HasSuffix(e.Name(), ".report") && strings.Contains(e.Name(), prio) {
						data, _ := os.ReadFile(filepath.Join(reportsDir, e.Name()))
						// Move to done
						doneDir := filepath.Join(brainRoot, "_inbox", "reports_done")
						os.MkdirAll(doneDir, 0755)
						os.Rename(filepath.Join(reportsDir, e.Name()), filepath.Join(doneDir, e.Name()))

						remaining := 0
						if updated, _ := os.ReadDir(reportsDir); updated != nil {
							for _, u := range updated {
								if strings.HasSuffix(u.Name(), ".report") {
									remaining++
								}
							}
						}

						nextPrompt = fmt.Sprintf("[HEARTBEAT %s] 사용자 보고 처리 요청 (우선순위: %s, 대기열 잔여: %d건)\n\n--- 보고 내용 ---\n%s\n--- 끝 ---\n\n위 보고 내용을 분석하고 즉시 실행하세요. 완료 후 결과를 보고하세요.",
							time.Now().Format("15:04"), prio, remaining, string(data))
						break
					}
				}
				if nextPrompt != "" {
					break
				}
			}
		}
		}

		// ── Priority 2: System health check (5분마다 1회) ──
		if nextPrompt == "" {
			// Check critical system state
			var issues []string

			// Check supervisor log freshness
			svLog := filepath.Join(filepath.Dir(brainRoot), "logs", "supervisor.log")
			if info, err := os.Stat(svLog); err != nil {
				issues = append(issues, "supervisor.log 없음 — supervisor 미실행 가능성")
			} else if time.Since(info.ModTime()) > 5*time.Minute {
				issues = append(issues, fmt.Sprintf("supervisor.log %s 이후 갱신 없음", info.ModTime().Format("15:04")))
			}

			// Check corrections inbox
			cPath := filepath.Join(brainRoot, "_inbox", "corrections.jsonl")
			if info, err := os.Stat(cPath); err == nil && info.Size() > 0 {
				issues = append(issues, fmt.Sprintf("미처리 교정 %d바이트 대기 중", info.Size()))
			}

			if len(issues) > 0 {
				nextPrompt = fmt.Sprintf("[WATCHDOG %s] 시스템 상태 점검 결과:\n\n%s\n\n각 항목을 확인하고 조치하세요.",
					time.Now().Format("15:04"), strings.Join(issues, "\n- "))
			}
		}

		// ── Priority 3: Todo fallback ──
		if nextPrompt == "" {
			filepath.Walk(todoDir, func(path string, info os.FileInfo, err error) error {
				if err != nil || !info.IsDir() || path == todoDir {
					return nil
				}
				if _, err := os.Stat(filepath.Join(path, "decay.dormant")); err == nil {
					return nil
				}
				taskName := filepath.Base(path)
				nextPrompt = fmt.Sprintf("[HEARTBEAT %s] 유휴 상태 감지. prefrontal/todo에서 작업 발견: '%s'\n\n이 작업을 분석하고 실행 계획을 세운 뒤 즉시 착수하세요. 완료 후 결과를 보고하세요.",
					time.Now().Format("15:04"), taskName)
				return filepath.SkipDir
			})
		}

		if nextPrompt != "" {
			fmt.Printf("[HEARTBEAT] ⚡ → bot1 주입: %s\n", nextPrompt[:min(80, len(nextPrompt))])
			cmd := exec.Command("node", pulseScript, nextPrompt, "bot1")
			if err := cmd.Run(); err != nil {
				fmt.Printf("[HEARTBEAT] ⚠️ CDP injection failed: %v\n", err)
			}
			touchActivity()
			lastInjection = time.Now()
		}
	}
}

func max64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// REST API: Programmatic growth for n8n/dashboard/webhooks
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
func startAPI(brainRoot string, port int) {
	mux := http.NewServeMux()

	// Initialize activity tracker
	touchActivity()

	// CORS middleware with activity tracking
	withCORS := func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			if r.Method == "OPTIONS" {
				w.WriteHeader(200)
				return
			}
			// Track activity (skip dashboard polling to avoid resetting idle timer)
			if r.URL.Path != "/api/brain" && r.URL.Path != "/api/state" && r.URL.Path != "/favicon.ico" {
				touchActivity()
			}
			h(w, r)
		}
	}

	// POST /api/grow  {"path": "cortex/frontend/coding/no_console_log"}
	mux.HandleFunc("/api/grow", withCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "POST only", 405)
			return
		}
		var req struct {
			Path string `json:"path"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Path == "" {
			http.Error(w, `{"error":"path required"}`, 400)
			return
		}
		growNeuron(brainRoot, req.Path)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "grown", "path": req.Path})
	}))

	// POST /api/fire  {"path": "cortex/frontend/coding/no_console_log"}
	mux.HandleFunc("/api/fire", withCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "POST only", 405)
			return
		}
		var req struct {
			Path string `json:"path"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Path == "" {
			http.Error(w, `{"error":"path required"}`, 400)
			return
		}
		fireNeuron(brainRoot, req.Path)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "fired", "path": req.Path})
	}))

	// POST /api/signal  {"path": "...", "type": "dopamine|bomb|memory"}
	mux.HandleFunc("/api/signal", withCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "POST only", 405)
			return
		}
		var req struct {
			Path string `json:"path"`
			Type string `json:"type"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Path == "" || req.Type == "" {
			http.Error(w, `{"error":"path and type required"}`, 400)
			return
		}
		signalNeuron(brainRoot, req.Path, req.Type)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "signaled", "path": req.Path, "type": req.Type})
	}))

	// POST /api/decay  {"days": 30}
	mux.HandleFunc("/api/decay", withCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "POST only", 405)
			return
		}
		var req struct {
			Days int `json:"days"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		if req.Days <= 0 {
			req.Days = 30
		}
		runDecay(brainRoot, req.Days)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "decay_complete", "days": req.Days})
	}))

	// GET /api/state — current brain state JSON
	mux.HandleFunc("/api/state", withCORS(func(w http.ResponseWriter, r *http.Request) {
		stateFile := filepath.Join(brainRoot, "..", "brain_state.json")
		abs, _ := filepath.Abs(stateFile)
		data, err := os.ReadFile(abs)
		if err != nil {
			http.Error(w, `{"error":"brain_state.json not found"}`, 404)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	}))

	// POST /api/evolve  {"dry_run": false}
	mux.HandleFunc("/api/evolve", withCORS(handleEvolveAPI(brainRoot)))

	// POST /api/dedup — 중복 뉴런 Jaccard 병합
	mux.HandleFunc("/api/dedup", withCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "POST only", 405)
			return
		}
		deduplicateNeurons(brainRoot)
		brain := scanBrain(brainRoot)
		result := runSubsumption(brain)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":     "ok",
			"neurons":    result.TotalNeurons,
			"activation": result.TotalCounter,
		})
	}))

	// GET /api/read?region=cortex — read region rules + auto-fire top neurons (RAG retrieval)
	mux.HandleFunc("/api/read", withCORS(handleReadRegion(brainRoot)))

	// POST /api/inject — Re-scan brain + inject into GEMINI.md
	mux.HandleFunc("/api/inject", withCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "POST only", 405)
			return
		}
		autoReinject(brainRoot)
		brain := scanBrain(brainRoot)
		result := runSubsumption(brain)
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "Injected %d neurons, activation: %d", result.TotalNeurons, result.TotalCounter)
	}))

	// POST /api/sandbox — Live test: write text → stored as file, empty → deleted
	// Uses _sandbox.txt (raw text) instead of folder names to preserve emojis/special chars
	mux.HandleFunc("/api/sandbox", withCORS(func(w http.ResponseWriter, r *http.Request) {
		sandboxFile := filepath.Join(brainRoot, "brainstem", "_sandbox.txt")

		if r.Method == "GET" {
			// Return current sandbox content from text file
			result := map[string]interface{}{"rules": []string{}}
			data, err := os.ReadFile(sandboxFile)
			if err == nil {
				text := strings.TrimSpace(string(data))
				if text != "" {
					rules := []string{}
					for _, line := range strings.Split(text, "\n") {
						line = strings.TrimSpace(line)
						if line != "" {
							rules = append(rules, line)
						}
					}
					result["rules"] = rules
				}
			}
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			json.NewEncoder(w).Encode(result)
			return
		}
		if r.Method != "POST" {
			http.Error(w, "GET or POST only", 405)
			return
		}
		var req struct {
			Text string `json:"text"`
		}
		json.NewDecoder(r.Body).Decode(&req)

		text := strings.TrimSpace(req.Text)
		if text == "" {
			// Empty = delete sandbox, reinject
			os.Remove(sandboxFile)
			// Also clean up legacy folder-based sandbox
			os.RemoveAll(filepath.Join(brainRoot, "brainstem", "_sandbox"))
			autoReinject(brainRoot)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			json.NewEncoder(w).Encode(map[string]string{"status": "cleared"})
			return
		}

		// 1. Write raw text to _sandbox.txt (preserves emojis for API GET)
		os.MkdirAll(filepath.Dir(sandboxFile), 0755)
		os.WriteFile(sandboxFile, []byte(text), 0644)

		// 2. Create neuron folders in _sandbox/ (for brain scan + dashboard)
		sandboxDir := filepath.Join(brainRoot, "brainstem", "_sandbox")
		os.RemoveAll(sandboxDir)
		created := 0
		var createdPaths []string

		for _, line := range strings.Split(text, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			// Sanitize for folder name: keep Korean, alphanumeric, underscore, dash
			name := strings.ReplaceAll(line, " ", "_")
			name = strings.Map(func(r rune) rune {
				if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' || (r >= 0xAC00 && r <= 0xD7AF) || (r >= 0x3131 && r <= 0x318E) {
					return r
				}
				return '_'
			}, name)
			if name == "" {
				continue
			}
			neuronDir := filepath.Join(sandboxDir, name)
			os.MkdirAll(neuronDir, 0755)
			os.WriteFile(filepath.Join(neuronDir, "1.neuron"), []byte{}, 0644)
			createdPaths = append(createdPaths, "brainstem/_sandbox/"+name)
			created++
		}

		autoReinject(brainRoot)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "applied",
			"created": created,
			"paths":   createdPaths,
		})
	}))

	// POST /api/rollback {\"path\": \"cortex/...\"} — decrement neuron counter (min=1)
	mux.HandleFunc("/api/rollback", withCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "POST only", 405)
			return
		}
		var req struct {
			Path string `json:"path"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Path == "" {
			http.Error(w, `{"error":"path required"}`, 400)
			return
		}
		if err := rollbackNeuron(brainRoot, req.Path); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(400)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error(), "path": req.Path})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "rolled_back", "path": req.Path})
	}))

	// GET / — Dashboard HTML (same as --dashboard mode)
	// Static files: 3D dashboard, brain.obj, brain_state.json
	neuronfsRoot := filepath.Dir(brainRoot) // NeuronFS/ directory (parent of brain_v4)
	mux.HandleFunc("/3d", withCORS(func(w http.ResponseWriter, r *http.Request) {
		htmlPath := filepath.Join(neuronfsRoot, "brain_dashboard.html")
		data, err := os.ReadFile(htmlPath)
		if err != nil {
			http.Error(w, "brain_dashboard.html not found", 404)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(data)
	}))
	mux.HandleFunc("/brain.obj", withCORS(func(w http.ResponseWriter, r *http.Request) {
		objPath := filepath.Join(neuronfsRoot, "brain.obj")
		data, err := os.ReadFile(objPath)
		if err != nil {
			http.Error(w, "brain.obj not found", 404)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		w.Write(data)
	}))
	mux.HandleFunc("/brain_state.json", withCORS(func(w http.ResponseWriter, r *http.Request) {
		jsonPath := filepath.Join(neuronfsRoot, "brain_state.json")
		data, err := os.ReadFile(jsonPath)
		if err != nil {
			http.Error(w, "brain_state.json not found", 404)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	}))

	// GET / — Unified Dashboard (3D + management)
	mux.HandleFunc("/", withCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			if r.URL.Path == "/favicon.ico" || r.URL.Path == "/manifest.json" {
				w.WriteHeader(204)
				return
			}
			http.NotFound(w, r)
			return
		}
		htmlPath := filepath.Join(neuronfsRoot, "brain_dashboard.html")
		data, err := os.ReadFile(htmlPath)
		if err != nil {
			// Fallback to embedded card dashboard
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprint(w, dashboardHTML)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(data)
	}))

	// GET /cards — Card-only dashboard (legacy)
	mux.HandleFunc("/cards", withCORS(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, dashboardHTML)
	}))

	// POST /api/community — 외부 커뮤니티 트렌드를 뉴런으로 수집
	// Body: {"source":"github|reddit|hackernews","topic":"AI memory","insight":"..."}
	mux.HandleFunc("/api/community", withCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "POST only", 405)
			return
		}
		var req struct {
			Source  string `json:"source"`
			Topic   string `json:"topic"`
			Insight string `json:"insight"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Topic == "" {
			http.Error(w, `{"error":"topic required"}`, 400)
			return
		}
		// 안전한 경로 생성
		safeTopic := strings.ReplaceAll(req.Topic, " ", "_")
		safeTopic = strings.ReplaceAll(safeTopic, "/", "_")
		safeTopic = strings.ReplaceAll(safeTopic, "\\", "_")

		neuronPath := filepath.Join(brainRoot, "cortex", "community", req.Source, safeTopic)
		os.MkdirAll(neuronPath, 0755)

		// 카운터 파일 생성/증가
		files, _ := filepath.Glob(filepath.Join(neuronPath, "*.neuron"))
		counter := len(files) + 1
		counterFile := filepath.Join(neuronPath, fmt.Sprintf("%d.neuron", counter))
		os.WriteFile(counterFile, []byte(req.Insight), 0644)

		fmt.Printf("[COMMUNITY] 📡 %s/%s → counter %d\n", req.Source, safeTopic, counter)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "ok",
			"path":    fmt.Sprintf("cortex/community/%s/%s", req.Source, safeTopic),
			"counter": counter,
		})
	}))

	// GET /api/health — system process health check
	mux.HandleFunc("/api/health", withCORS(func(w http.ResponseWriter, r *http.Request) {
		health := buildHealthJSON(brainRoot)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(health)
	}))

	// GET /api/brain — full brain state for dashboard (compatible with dashboard.go format)
	mux.HandleFunc("/api/brain", withCORS(func(w http.ResponseWriter, r *http.Request) {
		data := buildBrainJSONResponse(brainRoot)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	}))

	// GET/POST /api/heartbeat — heartbeat 토글 + 상태 조회
	mux.HandleFunc("/api/heartbeat", withCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			var req struct {
				Enabled  *bool `json:"enabled"`
				Interval *int  `json:"interval"`
				Cooldown *int  `json:"cooldown"`
			}
			json.NewDecoder(r.Body).Decode(&req)

			heartbeatMu.Lock()
			if req.Enabled != nil {
				heartbeatEnabled = *req.Enabled
			}
			if req.Interval != nil && *req.Interval >= 5 {
				heartbeatInterval = *req.Interval
			}
			if req.Cooldown != nil && *req.Cooldown >= 1 {
				heartbeatCooldown = *req.Cooldown
			}
			state := map[string]interface{}{
				"enabled":  heartbeatEnabled,
				"interval": heartbeatInterval,
				"cooldown": heartbeatCooldown,
			}
			heartbeatMu.Unlock()

			action := "OFF"
			if heartbeatEnabled {
				action = "ON"
			}
			fmt.Printf("[HEARTBEAT] ⚡ %s (interval=%ds, cooldown=%dm)\n", action, heartbeatInterval, heartbeatCooldown)

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(state)
			return
		}
		// GET
		heartbeatMu.Lock()
		state := map[string]interface{}{
			"enabled":  heartbeatEnabled,
			"interval": heartbeatInterval,
			"cooldown": heartbeatCooldown,
		}
		heartbeatMu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(state)
	}))

	// Start injection loop (inbox processing + auto-reinject) in background
	go runInjectionLoop(brainRoot)

	// Start idle engine in background
	go runIdleLoop(brainRoot)

	// Start heartbeat in background
	go runHeartbeatLoop(brainRoot)

	fmt.Printf("[NeuronFS API] 🧠 Serving on http://localhost:%d\n", port)
	fmt.Printf("  GET  /                           — Dashboard (카드 UI)\n")
	fmt.Printf("  GET  /3d                         — 3D Brain Topology\n")
	fmt.Printf("  POST /api/grow    {path}          — Create neuron\n")
	fmt.Printf("  POST /api/fire    {path}          — Increment counter\n")
	fmt.Printf("  POST /api/signal  {path, type}    — Add dopamine/bomb/memory\n")
	fmt.Printf("  POST /api/decay   {days}          — Dormant sweep\n")
	fmt.Printf("  POST /api/evolve  {dry_run}       — Groq autonomous evolution\n")
	fmt.Printf("  GET  /api/read    ?region=cortex  — Read region rules (RAG, auto-fire)\n")
	fmt.Printf("  GET  /api/state                   — Brain state JSON\n")
	fmt.Printf("  GET  /api/brain                   — Full brain state for dashboard\n")
	fmt.Printf("  GET  /api/heartbeat               — Heartbeat status\n")
	fmt.Printf("  POST /api/heartbeat {enabled,interval,cooldown} — Toggle heartbeat\n")
	fmt.Printf("  💓 HEARTBEAT: %s (interval=%ds, cooldown=%dm)\n", func() string {
		if heartbeatEnabled {
			return "ON"
		}
		return "OFF"
	}(), heartbeatInterval, heartbeatCooldown)
	fmt.Printf("  🔄 IDLE ENGINE: auto evolve/snapshot/NAS every %dm idle\n", idleThresholdMinutes)
	// POST /api/report — stackable report queue
	mux.HandleFunc("/api/report", withCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "POST only", 405)
			return
		}
		var req struct {
			Message  string `json:"message"`
			Priority string `json:"priority"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		if req.Message == "" {
			http.Error(w, "message required", 400)
			return
		}
		if req.Priority == "" {
			req.Priority = "normal"
		}
		reportsDir := filepath.Join(brainRoot, "_inbox", "reports")
		os.MkdirAll(reportsDir, 0755)
		ts := fmt.Sprintf("%d", time.Now().UnixMilli())
		filename := fmt.Sprintf("%s_%s.report", ts, req.Priority)
		content := fmt.Sprintf("priority: %s\ntimestamp: %s\n\n%s\n", req.Priority, time.Now().Format("2006-01-02 15:04:05"), req.Message)
		os.WriteFile(filepath.Join(reportsDir, filename), []byte(content), 0644)

		entries, _ := os.ReadDir(reportsDir)
		pending := 0
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), ".report") {
				pending++
			}
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "confirmed", "pending": pending, "priority": req.Priority,
			"ack": "새로운 보고가 확인되었습니다. 사용자의 요청 처리 후 팔로업합니다.",
		})
	}))

	// GET /api/reports — list pending
	mux.HandleFunc("/api/reports", withCORS(func(w http.ResponseWriter, r *http.Request) {
		reportsDir := filepath.Join(brainRoot, "_inbox", "reports")
		entries, _ := os.ReadDir(reportsDir)
		var reports []map[string]string
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), ".report") {
				data, _ := os.ReadFile(filepath.Join(reportsDir, e.Name()))
				reports = append(reports, map[string]string{"name": e.Name(), "content": string(data)})
			}
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"pending": len(reports), "reports": reports})
	}))

	fmt.Printf("  POST /api/report  {message,priority} — Stackable report queue\n")
	fmt.Printf("  GET  /api/reports                — List pending reports\n")
	http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
}
