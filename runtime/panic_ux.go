// panic_ux.go — OOM Flatline Death Screen UX v2
//
// ENFP PRD v2: 20260401_083800_enfp_oom_flatline_ux_prd_v2.md
//
// 시스템 패닉(OOM 등) 시 Go 기본 스택 트레이스 출력을 차단하고,
// 생물학적 메타포 기반의 3단계 CLI 데스 스크린을 렌더링한다:
//
//	Step 1: 아밀로이드 플라크 발작 (400ms 붉은 글리치)
//	Step 2: 뇌파 정지 Flatline (1100ms EEG ASCII 모션)
//	Step 3: 부검 소견서 (CAUSE + FAULT NODE 1줄 파싱)
//
// v2 변경사항:
//   - \033[2J 대신 \033[H + 빈줄 (tmux 스크롤 이력 보호)
//   - 2차 패닉 안전장치 (os.Exit(2) + raw stderr)
//   - 이모지 ASCII 대체 (TERM=dumb / NO_COLOR)
//   - ANSI 3단계 폴백 체인 (Truecolor → 256색 → 8색)
//   - Quiet 모드 지원
//
// Usage:
//
//	defer RenderFlatlineOnPanic()  // ← main() 또는 goroutine 최상위에서 호출
package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"runtime/debug"
	"time"
)

// ─── Flatline Color Helpers (Terminal-aware, 3-stage fallback) ───

// flatColor wraps ANSI codes with terminal-aware fallback.
type flatColor struct {
	mode colorMode // reuses awakening.go's colorMode enum
}

func (c flatColor) red(s string) string {
	if c.mode == colorNone {
		return s
	}
	return "\033[31;1m" + s + "\033[0m"
}

func (c flatColor) redBlink(s string) string {
	switch c.mode {
	case colorNone:
		return s
	case colorBasic:
		// Blink often unsupported; bold only
		return "\033[31;1m" + s + "\033[0m"
	default:
		return "\033[31;5m" + s + "\033[0m"
	}
}

func (c flatColor) yellow(s string) string {
	if c.mode == colorNone {
		return s
	}
	return "\033[33m" + s + "\033[0m"
}

func (c flatColor) dimGray(s string) string {
	if c.mode == colorNone {
		return s
	}
	return "\033[90m" + s + "\033[0m"
}

func (c flatColor) rose(s string) string {
	switch c.mode {
	case colorTrueC:
		return "\033[38;2;244;63;94m" + s + "\033[0m" // #F43F5E
	case color256:
		return "\033[38;5;197m" + s + "\033[0m"
	case colorBasic:
		return "\033[31m" + s + "\033[0m" // standard red fallback
	default:
		return s
	}
}

// ─── Emoji Substitution (RISK-4) ───

// flatEmoji returns terminal-safe emoji or ASCII substitutes.
func flatEmoji(emoji, asciiAlt string) string {
	cm := detectColorMode()
	if cm == colorNone || os.Getenv("TERM") == "dumb" {
		return asciiAlt
	}
	return emoji
}

// ─── Soft Screen Clear (RISK-3: tmux scroll protection) ───

// softClear replaces \033[2J with cursor-home + blank lines.
// Preserves tmux/screen scrollback history.
func softClear() {
	fmt.Fprintf(os.Stderr, "\033[H") // cursor to home
	for i := 0; i < 10; i++ {
		fmt.Fprintf(os.Stderr, "\033[K\n") // clear line + newline
	}
	fmt.Fprintf(os.Stderr, "\033[H") // back to top
}

// ─── Flatline Configuration ───

// FlatlineQuiet controls whether the death screen is suppressed.
// Set to true when --quiet flag is active.
var FlatlineQuiet bool

// secureIntn replaces math/rand.Intn with crypto/rand
func secureIntn(n int) int {
	if n <= 0 {
		return 0
	}
	val, err := rand.Int(rand.Reader, big.NewInt(int64(n)))
	if err != nil {
		return 0
	}
	return int(val.Int64())
}

// ─── Main Entry Point ───

// RenderFlatlineOnPanic intercepts panics at the top level.
// Place `defer RenderFlatlineOnPanic()` at the very start of main()
// or any critical goroutine entry point.
//
// RULE: Zero standard stack trace leaks to stdout. (Zero-Garbage 원칙)
// RISK-5: 2차 패닉 안전장치 — inner defer/recover wraps all output.
func RenderFlatlineOnPanic() {
	r := recover()
	if r == nil {
		return
	}

	// Capture raw stack before any output
	rawStack := string(debug.Stack())

	// Extract structured cause & fault node
	cause := fmt.Sprintf("%v", r)
	faultNode := extractFaultNode(rawStack)

	// Quiet mode: skip visual death screen, just log and exit
	if FlatlineQuiet {
		dumpForensicLog(cause, faultNode, rawStack)
		fmt.Fprintf(os.Stderr, "[FATAL] %s at %s\n", cause, faultNode)
		os.Exit(1)
		return
	}

	// RISK-5: Wrap all rendering in secondary recover
	// If the death screen itself panics, fall back to raw output
	func() {
		defer func() {
			if r2 := recover(); r2 != nil {
				// 2차 패닉: 미학보다 안전 우선
				fmt.Fprintf(os.Stderr, "\n[FATAL] Death screen internal panic: %v\n", r2)
				fmt.Fprintf(os.Stderr, "[FATAL] Original panic: %s at %s\n", cause, faultNode)
				fmt.Fprintf(os.Stderr, "=== Raw Stack ===\n%s\n", rawStack)
				os.Exit(2)
			}
		}()

		cm := detectColorMode()
		clr := flatColor{mode: cm}

		// ── Step 1: 아밀로이드 플라크 발작 (T=0ms ~ 400ms) ──
		renderSeizure(clr)

		// ── Step 2: 뇌파 정지 Flatline (T=400ms ~ 1500ms) ──
		renderFlatlineEEG(clr)

		// ── Step 3: 부검 소견서 (T=1500ms~) ──
		renderNecrosisReport(clr, cause, faultNode)

		// Dump full trace to file (forensics), never to terminal
		dumpForensicLog(cause, faultNode, rawStack)
	}()

	os.Exit(1)
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// Step 1: Seizure — 400ms red glitch blink
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// glitchFragments generates pseudo-hex garbage reminiscent of memory corruption.
var glitchFragments = []string{
	"0xDEADBEEF", "0xBADCA5E", "0xFACEFEED", "0xC0FFEE",
	"0x0B57AC1E", "0xDEFEC8ED", "0xD15EA5E", "0xBAADF00D",
	"ERR_AMYLOID_PLAQUE_MAX", "HEAP_OVERFLOW", "SYNAPSE_SEVERED",
	"CORTICAL_BREACH", "NEURAL_DECAY", "MEMBRANE_RUPTURE",
}

func renderSeizure(clr flatColor) {
	// 5 rapid frames across 400ms (80ms/frame)
	for frame := 0; frame < 5; frame++ {
		softClear()
		fmt.Fprintf(os.Stderr, "\n%s\n", clr.redBlink("[!!] CORTICAL OVERFLOW [!!]"))

		lineCount := 2 + secureIntn(2)
		for l := 0; l < lineCount; l++ {
			frag1 := glitchFragments[secureIntn(len(glitchFragments))]
			frag2 := glitchFragments[secureIntn(len(glitchFragments))]
			addr := secureIntn(0xFFFFFF)
			fmt.Fprintf(os.Stderr, "%s\n", clr.red(fmt.Sprintf("%s %s x%06X", frag1, frag2, addr)))
		}
		time.Sleep(80 * time.Millisecond)
	}

	// Soft-clear the glitch debris (소각 — tmux-safe)
	softClear()
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// Step 2: EEG Flatline — 1100ms brain wave decay
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// EEG waveform frames: heart-rate-like spikes decaying to flatline.
var eegFrames = []string{
	// Frame 0: alive — strong peaks
	`[EEG]  /\  /\/\    /\  /\    /\/\    /\`,
	// Frame 1: weakening — smaller amplitude
	`[EEG]  /\  /\    _/\  /    \_   /\  _`,
	// Frame 2: dying — intermittent blips
	`[EEG]  _   _/\_    ___/\_    __   _  __`,
	// Frame 3: near-death — faint tremor
	`[EEG]  ___  _/\___    ____    __ __ __`,
	// Frame 4: flatline — complete silence
	`[EEG]  _  _  _  _  _  _  _  _  _  _  _`,
}

func renderFlatlineEEG(clr flatColor) {
	fmt.Fprintf(os.Stderr, "\n")

	// Each frame ~220ms → 5 frames = 1100ms
	for i, frame := range eegFrames {
		// Overwrite the same line
		fmt.Fprintf(os.Stderr, "\r\033[K")

		var colored string
		if i >= 3 {
			colored = clr.dimGray(frame)
		} else {
			colored = clr.yellow(frame)
		}
		if i == len(eegFrames)-1 {
			colored = clr.rose(frame)
		}

		fmt.Fprintf(os.Stderr, "%s", colored)
		time.Sleep(220 * time.Millisecond)
	}

	fmt.Fprintf(os.Stderr, "\n\n")
	time.Sleep(200 * time.Millisecond) // brief silence before the report
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// Step 3: Necrosis Report — parsed autopsy brief
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

func renderNecrosisReport(clr flatColor, cause, faultNode string) {
	sep := "==========================================="
	ts := time.Now().Format("20060102_1504")

	skull := flatEmoji("💀", "[DEAD]")
	bolt := flatEmoji("⚡", "[!!]")

	fmt.Fprintf(os.Stderr, "%s\n", clr.rose(sep))
	fmt.Fprintf(os.Stderr, "%s\n", clr.rose(fmt.Sprintf("[%s SYSTEM NECROSIS DETECTED]", skull)))
	fmt.Fprintf(os.Stderr, "%s\n", clr.rose(sep))
	fmt.Fprintf(os.Stderr, "%s\n", clr.rose(fmt.Sprintf("CAUSE       : %s", cause)))
	fmt.Fprintf(os.Stderr, "%s\n", clr.rose(fmt.Sprintf("FAULT NODE  : %s", faultNode)))
	fmt.Fprintf(os.Stderr, "%s\n", clr.rose(fmt.Sprintf("QUARANTINE  : Dead tissue isolated to branch 'quarantine/infection-%s'", ts)))
	fmt.Fprintf(os.Stderr, "%s\n", clr.rose("STATUS      : Main trunk securely preserved."))
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "%s\n", clr.rose(fmt.Sprintf("> %s [AUTO-DEFIBRILLATION] Revival in 3.0s...", bolt)))
	fmt.Fprintf(os.Stderr, "\n")
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// Forensic logging — full trace goes to file, never terminal
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

func dumpForensicLog(cause, faultNode, rawStack string) {
	ts := time.Now().Format("20060102_150405")
	logDir := "logs"
	os.MkdirAll(logDir, 0750)
	logPath := fmt.Sprintf("%s/necrosis_%s.log", logDir, ts)

	content := fmt.Sprintf("=== NeuronFS Necrosis Report ===\n"+
		"Time       : %s\n"+
		"Cause      : %s\n"+
		"Fault Node : %s\n\n"+
		"=== Full Stack Trace (forensic only) ===\n%s\n",
		time.Now().Format(time.RFC3339), cause, faultNode, rawStack)

	if err := os.WriteFile(logPath, []byte(content), 0600); err == nil {
		cm := detectColorMode()
		clr := flatColor{mode: cm}
		fmt.Fprintf(os.Stderr, "%s\n", clr.dimGray(fmt.Sprintf("[FORENSIC] Full trace → %s", logPath)))
	}
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// RenderFlatlineOnOOM — supervisor OOM 연동 (non-panic)
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// RenderFlatlineOnOOM renders the death screen sequence when the supervisor
// detects an out-of-memory condition via external process monitoring.
// Unlike RenderFlatlineOnPanic, this does NOT call os.Exit — the supervisor
// handles process lifecycle after rendering.
//
// Parameters:
//   - processName: the child process name (e.g. "neuronfs-api")
//   - memKB: detected memory usage in KB
//   - topLeaks: pre-formatted string of top memory leak callers
func RenderFlatlineOnOOM(processName string, memKB int64, topLeaks string) {
	if FlatlineQuiet {
		fmt.Fprintf(os.Stderr, "[OOM] %s exceeded memory limit (%d KB)\n", processName, memKB)
		return
	}

	// Safety wrapper — if rendering itself fails, degrade gracefully
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "\n[FATAL] OOM death screen internal error: %v\n", r)
		}
	}()

	cm := detectColorMode()
	clr := flatColor{mode: cm}

	// ── Step 1: Seizure (abbreviated — 3 frames × 60ms = 180ms) ──
	for frame := 0; frame < 3; frame++ {
		softClear()
		fmt.Fprintf(os.Stderr, "\n%s\n", clr.redBlink("[!!] AMYLOID PLAQUE OVERLOAD [!!]"))
		frag1 := glitchFragments[secureIntn(len(glitchFragments))]
		frag2 := glitchFragments[secureIntn(len(glitchFragments))]
		fmt.Fprintf(os.Stderr, "%s\n", clr.red(fmt.Sprintf("  %s  %s  PID:%s", frag1, frag2, processName)))
		time.Sleep(60 * time.Millisecond)
	}
	softClear()

	// ── Step 2: EEG Flatline (last 3 frames only — 660ms) ──
	fmt.Fprintf(os.Stderr, "\n")
	for i := 2; i < len(eegFrames); i++ {
		fmt.Fprintf(os.Stderr, "\r\033[K")
		var colored string
		if i >= 3 {
			colored = clr.dimGray(eegFrames[i])
		} else {
			colored = clr.yellow(eegFrames[i])
		}
		if i == len(eegFrames)-1 {
			colored = clr.rose(eegFrames[i])
		}
		fmt.Fprintf(os.Stderr, "%s", colored)
		time.Sleep(220 * time.Millisecond)
	}
	fmt.Fprintf(os.Stderr, "\n\n")

	// ── Step 3: OOM Necrosis Report ──
	sep := "==========================================="
	skull := flatEmoji("💀", "[DEAD]")
	bolt := flatEmoji("⚡", "[!!]")
	memMB := memKB / 1024

	fmt.Fprintf(os.Stderr, "%s\n", clr.rose(sep))
	fmt.Fprintf(os.Stderr, "%s\n", clr.rose(fmt.Sprintf("[%s SYNAPTIC OVERLOAD — OOM NECROSIS]", skull)))
	fmt.Fprintf(os.Stderr, "%s\n", clr.rose(sep))
	fmt.Fprintf(os.Stderr, "%s\n", clr.rose(fmt.Sprintf("PROCESS     : %s", processName)))
	fmt.Fprintf(os.Stderr, "%s\n", clr.rose(fmt.Sprintf("MEMORY      : %d MB (threshold: 50 MB)", memMB)))
	fmt.Fprintf(os.Stderr, "%s\n", clr.rose("ACTION      : Process terminated. Awaiting neurogenesis."))
	if topLeaks != "" {
		fmt.Fprintf(os.Stderr, "\n%s\n", clr.yellow("=== Top Memory Leak Sources ==="))
		fmt.Fprintf(os.Stderr, "%s\n", clr.dimGray(topLeaks))
	}
	fmt.Fprintf(os.Stderr, "\n%s\n\n", clr.rose(fmt.Sprintf("> %s [AUTO-DEFIBRILLATION] Supervisor will restart process...", bolt)))
}
