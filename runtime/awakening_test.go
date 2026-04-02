package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// captureStderr redirects os.Stderr to a buffer during fn execution.
func captureStderr(fn func()) string {
	r, w, _ := os.Pipe()
	old := os.Stderr
	os.Stderr = w

	fn()

	w.Close()
	os.Stderr = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String()
}

// TestAwakening_FullSequence verifies Step1???? execute in order
// with total duration roughly within 2.5s ± 500ms.
func TestAwakening_FullSequence(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := AwakeningConfig{
		BrainRoot:      tmpDir,
		Quiet:          false,
		ForceAwakening: true, // force full even on repeated test runs
		NeuronCount:    134,
		PlaqueFaults:   0,
	}

	start := time.Now()
	output := captureStderr(func() {
		RunAwakening(context.Background(), cfg)
	})
	elapsed := time.Since(start)

	// Verify all 3 steps produced output
	if len(output) == 0 {
		t.Fatal("expected awakening output, got empty string")
	}

	// Step 1 markers
	if !containsAny(output, "vital signs", "Cerebrospinal", "cortex") {
		t.Error("Step 1 (Brainstem Ignition) markers missing from output")
	}

	// Step 2 markers
	if !containsAny(output, "SYNAPSE LINK", "64%", "100%") {
		t.Error("Step 2 (Synapse Link) markers missing from output")
	}

	// Step 3 markers
	if !containsAny(output, "NeuronFS", "ALIVE", "134 neurons") {
		t.Error("Step 3 (Full Consciousness) markers missing from output")
	}

	// Duration check: should be ~2.5s ± 500ms
	if elapsed < 1500*time.Millisecond {
		t.Errorf("sequence too fast: %v (expected ~2.5s)", elapsed)
	}
	if elapsed > 4000*time.Millisecond {
		t.Errorf("sequence too slow: %v (expected ~2.5s)", elapsed)
	}

	// Verify .neuronfs_init marker was written
	markerFile := filepath.Join(tmpDir, ".neuronfs_init")
	if _, err := os.Stat(markerFile); os.IsNotExist(err) {
		t.Error(".neuronfs_init marker file not created after full sequence")
	}

	t.Logf("Full sequence completed in %v", elapsed)
}

// TestAwakening_QuietMode verifies zero output when Quiet=true.
func TestAwakening_QuietMode(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := AwakeningConfig{
		BrainRoot:   tmpDir,
		Quiet:       true,
		NeuronCount: 50,
	}

	output := captureStderr(func() {
		RunAwakening(context.Background(), cfg)
	})

	if len(output) != 0 {
		t.Errorf("quiet mode should produce 0 bytes of output, got %d bytes: %s", len(output), output)
	}

	// Still should write the marker
	markerFile := filepath.Join(tmpDir, ".neuronfs_init")
	if _, err := os.Stat(markerFile); os.IsNotExist(err) {
		t.Error(".neuronfs_init marker should be written even in quiet mode")
	}
}

// TestAwakening_SecondRun verifies abbreviated mode when .neuronfs_init exists.
func TestAwakening_SecondRun(t *testing.T) {
	tmpDir := t.TempDir()

	// Pre-create marker (simulate previous run)
	markerFile := filepath.Join(tmpDir, ".neuronfs_init")
	os.WriteFile(markerFile, []byte("initialized: 2026-04-01T00:00:00Z\n"), 0644)

	cfg := AwakeningConfig{
		BrainRoot:   tmpDir,
		NeuronCount: 200,
	}

	start := time.Now()
	output := captureStderr(func() {
		RunAwakening(context.Background(), cfg)
	})
	elapsed := time.Since(start)

	// Should NOT contain ASCII art (check art-specific box characters, not "NeuronFS" which appears in abbreviated line too)
	if containsAny(output, "|____/", "| \\_", "\\___|") {
		t.Error("second run should NOT display ASCII art")
	}

	// Should contain abbreviated status
	if !containsAny(output, "200 neurons", "online") {
		t.Error("second run should show abbreviated status line")
	}

	// Should be fast (< 500ms)
	if elapsed > 500*time.Millisecond {
		t.Errorf("abbreviated mode too slow: %v (expected < 500ms)", elapsed)
	}

	t.Logf("Abbreviated output: %s", output)
}

// TestAwakening_ContextCancel verifies clean shutdown when context is cancelled.
func TestAwakening_ContextCancel(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := AwakeningConfig{
		BrainRoot:      tmpDir,
		ForceAwakening: true,
		NeuronCount:    100,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	start := time.Now()
	captureStderr(func() {
		RunAwakening(ctx, cfg)
	})
	elapsed := time.Since(start)

	// Should terminate near the 500ms cancel point, not run the full 2.5s
	if elapsed > 1500*time.Millisecond {
		t.Errorf("context cancellation did not abort in time: %v", elapsed)
	}

	t.Logf("Context-cancelled in %v", elapsed)
}

// TestANSI_TruecolorFallback verifies the 3-stage ANSI color fallback chain.
func TestANSI_TruecolorFallback(t *testing.T) {
	// Test Truecolor mode
	tc := awakColor{mode: colorTrueC}
	roseTC := tc.rose("test")
	if !containsAny(roseTC, "38;2;255;0;102") {
		t.Errorf("truecolor rose should contain 24-bit escape, got: %q", roseTC)
	}

	// Test 256-color fallback
	c256 := awakColor{mode: color256}
	rose256 := c256.rose("test")
	if !containsAny(rose256, "38;5;197") {
		t.Errorf("256-color rose should contain 256-color escape, got: %q", rose256)
	}

	// Test basic (8-color) fallback
	basic := awakColor{mode: colorBasic}
	roseBasic := basic.rose("test")
	if !containsAny(roseBasic, "35;1") {
		t.Errorf("basic rose should use bold magenta fallback, got: %q", roseBasic)
	}

	// Test NO_COLOR mode
	none := awakColor{mode: colorNone}
	roseNone := none.rose("test")
	if roseNone != "test" {
		t.Errorf("NO_COLOR mode should return plain text, got: %q", roseNone)
	}

	t.Log("All 4 ANSI fallback tiers verified")
}

// containsAny returns true if s contains any of the needles.
func containsAny(s string, needles ...string) bool {
	for _, n := range needles {
		if len(n) > 0 && len(s) > 0 {
			found := false
			for i := 0; i <= len(s)-len(n); i++ {
				if s[i:i+len(n)] == n {
					found = true
					break
				}
			}
			if found {
				return true
			}
		}
	}
	return false
}

