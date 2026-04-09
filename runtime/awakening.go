// awakening.go — NeuronFS CLI Awakening Sequence v2
//
// ENFP PRD: 20260401_083600_enfp_cli_awakening_ux_prd_v2.md
//
// 초보 사용자가 처음 `neuronfs`를 실행할 때 바이오해커 랩이 깨어나는 듯한
// 3단계 ASCII 모션 시퀀스를 렌더링한다.
//
//	Step 1: 뇌간 세포 점화 (Brainstem Ignition) — T=0~800ms
//	Step 2: 시냅스 링크 (First Breath) — T=800~1500ms
//	Step 3: 각성 완료 (Full Consciousness) — T=1500~2500ms
//
// 외부 의존성: 0 (Go stdlib only)
// 반복실행: .neuronfs_init 마커 → 축약모드
// CI: 자동 quiet 모드
// ANSI: 3단계 폴백 (Truecolor → 256색 → 8색)
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ─── Awakening Configuration ───

// AwakeningConfig controls the awakening sequence behavior.
type AwakeningConfig struct {
	BrainRoot      string // brain_v4 path
	Quiet          bool   // --quiet / -q → skip entirely
	ForceAwakening bool   // --awakening → force full sequence
	NeuronCount    int    // pre-scanned neuron count
	PlaqueFaults   int    // detected amyloid plaques (bomb neurons)
}

// ─── Terminal Capability Detection ───

// colorMode represents terminal color capability.
type colorMode int

const (
	colorNone  colorMode = 0 // NO_COLOR or TERM=dumb
	colorBasic colorMode = 1 // 8-color ANSI
	color256   colorMode = 2 // 256-color
	colorTrueC colorMode = 3 // Truecolor (24-bit)
)

// detectColorMode probes environment variables to determine ANSI color capability.
func detectColorMode() colorMode {
	// NO_COLOR takes absolute priority (https://no-color.org/)
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		return colorNone
	}

	term := os.Getenv("TERM")
	if term == "dumb" || term == "" {
		// Windows Terminal / cmd usually have no TERM but support Truecolor
		// Check for Windows Terminal via WT_SESSION
		if os.Getenv("WT_SESSION") != "" {
			return colorTrueC
		}
		if term == "dumb" {
			return colorBasic
		}
		// Assume basic on unknown
		return colorBasic
	}

	// COLORTERM is the most reliable Truecolor indicator
	ct := os.Getenv("COLORTERM")
	if ct == "truecolor" || ct == "24bit" {
		return colorTrueC
	}

	if strings.Contains(term, "256color") {
		return color256
	}

	return colorBasic
}

// isCIEnvironment returns true if running inside CI/CD.
func isCIEnvironment() bool {
	envVars := []string{"CI", "GITHUB_ACTIONS", "JENKINS_URL", "GITLAB_CI", "CIRCLECI", "TRAVIS"}
	for _, e := range envVars {
		if os.Getenv(e) != "" {
			return true
		}
	}
	return os.Getenv("TERM") == "dumb"
}

// ─── ANSI Color Helpers with Fallback Chain ───

// awakColor wraps ANSI color codes with terminal-aware fallback.
type awakColor struct {
	mode colorMode
}

func (c awakColor) gray(s string) string {
	switch c.mode {
	case colorNone:
		return s
	default:
		return "\033[90m" + s + "\033[0m"
	}
}

func (c awakColor) cyan(s string) string {
	switch c.mode {
	case colorNone:
		return s
	default:
		return "\033[36m" + s + "\033[0m"
	}
}

func (c awakColor) blue(s string) string {
	switch c.mode {
	case colorNone:
		return s
	default:
		return "\033[34m" + s + "\033[0m"
	}
}

func (c awakColor) rose(s string) string {
	switch c.mode {
	case colorTrueC:
		return "\033[38;2;255;0;102m" + s + "\033[0m"
	case color256:
		return "\033[38;5;197m" + s + "\033[0m" // closest 256-color approximation
	case colorBasic:
		return "\033[35;1m" + s + "\033[0m" // bold magenta fallback
	default:
		return s
	}
}

func (c awakColor) green(s string) string {
	switch c.mode {
	case colorNone:
		return s
	default:
		return "\033[32m" + s + "\033[0m"
	}
}

func (c awakColor) alive(s string) string {
	switch c.mode {
	case colorTrueC, color256:
		// Try blink, fallback-safe
		return "\033[5;1m" + s + "\033[0m"
	case colorBasic:
		// Bold + underline fallback for terminals that ignore blink
		return "\033[1;4m" + s + "\033[0m"
	default:
		return "[" + s + "]"
	}
}

// ─── Marker File Management ───

const markerFileName = ".neuronfs_init"

func markerPath(brainRoot string) string {
	return filepath.Join(brainRoot, markerFileName)
}

func isFirstRun(brainRoot string) bool {
	_, err := os.Stat(markerPath(brainRoot))
	return os.IsNotExist(err)
}

func writeMarker(brainRoot string) {
	content := fmt.Sprintf("initialized: %s\n", time.Now().Format(time.RFC3339))
	os.WriteFile(markerPath(brainRoot), []byte(content), 0600)
}

// ─── Main Entry Point ───

// RunAwakening executes the CLI awakening sequence.
// Should be called at the very top of main(), before daemon loops.
func RunAwakening(ctx context.Context, cfg AwakeningConfig) {
	// CI auto-quiet
	if isCIEnvironment() && !cfg.ForceAwakening {
		cfg.Quiet = true
	}

	// Quiet mode: zero output, immediate return
	if cfg.Quiet {
		if cfg.BrainRoot != "" && isFirstRun(cfg.BrainRoot) {
			writeMarker(cfg.BrainRoot)
		}
		return
	}

	firstRun := cfg.BrainRoot != "" && isFirstRun(cfg.BrainRoot)

	// Second run: abbreviated mode (unless forced)
	if !firstRun && !cfg.ForceAwakening {
		runAbbreviatedBoot(cfg)
		return
	}

	// Full awakening sequence
	cm := detectColorMode()
	clr := awakColor{mode: cm}

	// Step 1: Brainstem Ignition (T=0ms ~ 800ms)
	if err := stepBrainstemIgnition(ctx, clr); err != nil {
		return
	}

	// Step 2: Synapse Link (T=800ms ~ 1500ms)
	if err := stepSynapseLink(ctx, clr); err != nil {
		return
	}

	// Step 3: Full Consciousness (T=1500ms ~ 2500ms)
	if err := stepFullConsciousness(ctx, clr, cfg); err != nil {
		return
	}

	// Write marker for subsequent runs
	if cfg.BrainRoot != "" {
		writeMarker(cfg.BrainRoot)
	}
}

// ─── Step 1: Brainstem Ignition ───

func stepBrainstemIgnition(ctx context.Context, clr awakColor) error {
	spinnerFrames := []string{"⠋", "⠙", "⠹"}
	messages := []struct {
		ts  string
		msg string
	}{
		{"0.00s", "Injecting vital signs..."},
		{"0.45s", "Cerebrospinal fluid stabilized."},
		{"0.80s", "Waking up the cortex..."},
	}

	for i, m := range messages {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		spinner := spinnerFrames[i%len(spinnerFrames)]
		line := fmt.Sprintf("[%s] %s %s", m.ts, spinner, m.msg)
		fmt.Fprintf(os.Stderr, "%s\n", clr.gray(line))
		time.Sleep(270 * time.Millisecond) // 3 lines across ~800ms
	}
	return nil
}

// ─── Step 2: Synapse Link ───

func stepSynapseLink(ctx context.Context, clr awakColor) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Phase 1: stalled at 64% (tension)
	bar64 := "[ ██████████░░░░░░ ]"
	fmt.Fprintf(os.Stderr, "%s\n",
		clr.blue(fmt.Sprintf("[1.00s] ⠼ SYNAPSE LINK  %s 64%% - Frontal lobe locked", bar64)))
	time.Sleep(350 * time.Millisecond) // tension hold

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Phase 2: snap to 100% (release)
	bar100 := "[ ████████████████ ]"
	fmt.Fprintf(os.Stderr, "%s\n",
		clr.cyan(fmt.Sprintf("[1.30s] ⠧ SYNAPSE LINK  %s 100%% - Synaptic crossover OK", bar100)))
	time.Sleep(350 * time.Millisecond)

	return nil
}

// ─── Step 3: Full Consciousness ───

var neuronFSAsciiArt = `       _   _                             _____ ____
      | \ | | ___ _   _ _ __ ___  _ __  |  ___/ ___|
      |  \| |/ _ \ | | | '__/ _ \| '_ \ | |_  \___ \
      | |\  |  __/ |_| | | | (_) | | | ||  _|  ___) |
      |_| \_|\___|\__,_|_|  \___/|_| |_||_|   |____/`

func stepFullConsciousness(ctx context.Context, clr awakColor, cfg AwakeningConfig) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Brief clear beat
	time.Sleep(100 * time.Millisecond)

	// ASCII art
	fmt.Fprintf(os.Stderr, "\n%s\n\n", clr.rose(neuronFSAsciiArt))

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Workspace line
	workspace := cfg.BrainRoot
	if workspace == "" {
		workspace = "unknown"
	}
	fmt.Fprintf(os.Stderr, "  > 🧠 [SYSTEM] Workspace %s is now %s.\n",
		workspace, clr.alive("ALIVE"))

	// Neuron stats
	fmt.Fprintf(os.Stderr, "  > 🧬 %s\n",
		clr.green(fmt.Sprintf("%d neurons found. %d amyloid plaques.", cfg.NeuronCount, cfg.PlaqueFaults)))

	fmt.Fprintf(os.Stderr, "  > ⚡ Ready to mutate. Waiting for cortical input...\n")
	fmt.Fprintf(os.Stderr, "\n")

	time.Sleep(500 * time.Millisecond) // let it sink in

	return nil
}

// ─── Abbreviated Boot (2nd+ runs) ───

func runAbbreviatedBoot(cfg AwakeningConfig) {
	cm := detectColorMode()
	clr := awakColor{mode: cm}

	status := fmt.Sprintf("🧠 NeuronFS online · %d neurons · %d plaques",
		cfg.NeuronCount, cfg.PlaqueFaults)
	fmt.Fprintf(os.Stderr, "%s\n", clr.green(status))
}
