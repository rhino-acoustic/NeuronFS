package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ━━━ activationBar ━━━
func TestActivationBar_AllTiers(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{100, "█████"},
		{90, "█████"},
		{50, "████░"},
		{75, "████░"},
		{20, "███░░"},
		{35, "███░░"},
		{10, "██░░░"},
		{15, "██░░░"},
		{5, "█░░░░"},
		{7, "█░░░░"},
		{0, "░░░░░"},
		{3, "░░░░░"},
	}
	for _, tt := range tests {
		result := activationBar(tt.input)
		if result != tt.expected {
			t.Errorf("activationBar(%d) = %q, want %q", tt.input, result, tt.expected)
		}
	}
	t.Logf("OK: activationBar all 6 tiers verified")
}

// ━━━ countNeuronFiles ━━━
func TestCountNeuronFiles(t *testing.T) {
	dir := t.TempDir()

	// Create some .neuron files at various depths
	os.MkdirAll(filepath.Join(dir, "a", "b"), 0755)
	os.WriteFile(filepath.Join(dir, "1.neuron"), []byte{}, 0644)
	os.WriteFile(filepath.Join(dir, "a", "2.neuron"), []byte{}, 0644)
	os.WriteFile(filepath.Join(dir, "a", "b", "3.neuron"), []byte{}, 0644)
	os.WriteFile(filepath.Join(dir, "a", "b", "bomb.neuron"), []byte{}, 0644)
	os.WriteFile(filepath.Join(dir, "a", "notaneuron.txt"), []byte{}, 0644)

	count := countNeuronFiles(dir)
	if count != 4 {
		t.Fatalf("expected 4 .neuron files, got %d", count)
	}
	t.Logf("OK: countNeuronFiles found %d .neuron files", count)
}

// ━━━ signalNeuron — bomb signal ━━━
func TestSignalNeuron_Bomb(t *testing.T) {
	dir := setupTestBrain(t)

	err := signalNeuron(dir, "cortex/left/frontend/hooks_pattern", "bomb")
	if err != nil {
		t.Fatalf("signalNeuron bomb failed: %v", err)
	}

	bombFile := filepath.Join(dir, "cortex", "left", "frontend", "hooks_pattern", "bomb.neuron")
	if _, err := os.Stat(bombFile); os.IsNotExist(err) {
		t.Fatal("expected bomb.neuron to exist")
	}
	t.Logf("OK: bomb signal created correctly")
}

// ━━━ signalNeuron — memory signal ━━━
func TestSignalNeuron_Memory(t *testing.T) {
	dir := setupTestBrain(t)

	err := signalNeuron(dir, "cortex/left/frontend/hooks_pattern", "memory")
	if err != nil {
		t.Fatalf("signalNeuron memory failed: %v", err)
	}

	memFile := filepath.Join(dir, "cortex", "left", "frontend", "hooks_pattern", "memory1.neuron")
	if _, err := os.Stat(memFile); os.IsNotExist(err) {
		t.Fatal("expected memory1.neuron to exist")
	}

	// Second memory signal
	err = signalNeuron(dir, "cortex/left/frontend/hooks_pattern", "memory")
	if err != nil {
		t.Fatalf("second signalNeuron memory failed: %v", err)
	}

	memFile2 := filepath.Join(dir, "cortex", "left", "frontend", "hooks_pattern", "memory2.neuron")
	if _, err := os.Stat(memFile2); os.IsNotExist(err) {
		t.Fatal("expected memory2.neuron to exist")
	}
	t.Logf("OK: memory signals created correctly")
}

// ━━━ signalNeuron — unknown type ━━━
func TestSignalNeuron_UnknownType(t *testing.T) {
	dir := setupTestBrain(t)

	err := signalNeuron(dir, "cortex/left/frontend/hooks_pattern", "invalid_type")
	if err == nil {
		t.Fatal("expected error for unknown signal type")
	}
	if !strings.Contains(err.Error(), "unknown signal type") {
		t.Fatalf("expected 'unknown signal type' error, got: %v", err)
	}
	t.Logf("OK: unknown signal type returns error: %v", err)
}

// ━━━ signalNeuron — nonexistent neuron ━━━
func TestSignalNeuron_NotFound(t *testing.T) {
	dir := setupTestBrain(t)

	err := signalNeuron(dir, "cortex/nonexistent/path", "dopamine")
	if err == nil {
		t.Fatal("expected error for nonexistent neuron")
	}
	t.Logf("OK: signal to nonexistent neuron returns error: %v", err)
}

// ━━━ emitRules ━━━
func TestEmitRules_Format(t *testing.T) {
	dir := setupTestBrain(t)
	brain := scanBrain(dir)
	result := runSubsumption(brain)

	rules := emitRules(result)

	if !strings.Contains(rules, "NEURONFS:START") {
		t.Fatal("expected NEURONFS:START marker")
	}
	if !strings.Contains(rules, "NEURONFS:END") {
		t.Fatal("expected NEURONFS:END marker")
	}
	if !strings.Contains(rules, "NeuronFS Active Rules") {
		t.Fatal("expected 'NeuronFS Active Rules' header")
	}
	t.Logf("OK: emitRules produces valid bootstrap (%d bytes)", len(rules))
}

// ━━━ runStats ━━━
func TestRunStats_NoCrash(t *testing.T) {
	dir := setupTestBrain(t)
	// runStats just prints to stdout — verify it doesn't crash
	runStats(dir)
	t.Logf("OK: runStats completed without crash")
}

// ━━━ runVacuum ━━━
func TestRunVacuum_NoCrash(t *testing.T) {
	dir := setupTestBrain(t)
	// runVacuum is a placeholder — verify it doesn't crash
	runVacuum(dir)
	t.Logf("OK: runVacuum completed without crash")
}

// ━━━ runDecay ━━━
func TestRunDecay_NoCrash(t *testing.T) {
	dir := setupTestBrain(t)
	// runDecay with 0 days should process all neurons
	runDecay(dir, 0)
	t.Logf("OK: runDecay completed without crash")
}

// ━━━ printDiag ━━━
func TestPrintDiag_NoCrash(t *testing.T) {
	dir := setupTestBrain(t)
	brain := scanBrain(dir)
	result := runSubsumption(brain)
	// printDiag just prints to stdout
	printDiag(brain, result)
	t.Logf("OK: printDiag completed without crash")
}

// ━━━ printDiag with bomb ━━━
func TestPrintDiag_WithBomb(t *testing.T) {
	dir := setupTestBrain(t)
	bombDir := filepath.Join(dir, "brainstem", "canon", "never_use_fallback")
	os.WriteFile(filepath.Join(bombDir, "bomb.neuron"), []byte{}, 0644)

	brain := scanBrain(dir)
	result := runSubsumption(brain)
	printDiag(brain, result)
	t.Logf("OK: printDiag with bomb completed without crash")
}

// ━━━ deduplicateNeurons ━━━
func TestDeduplicateNeurons_NoCrash(t *testing.T) {
	dir := setupTestBrain(t)
	deduplicateNeurons(dir)
	t.Logf("OK: deduplicateNeurons completed without crash")
}

// ━━━ buildHealthJSON ━━━
func TestBuildHealthJSON(t *testing.T) {
	dir := setupTestBrain(t)
	health := buildHealthJSON(dir)

	if !health.API {
		t.Fatal("expected API=true")
	}
	if len(health.Processes) == 0 {
		t.Fatal("expected at least 1 process")
	}
	if health.OS == "" {
		t.Fatal("expected OS to be set")
	}
	t.Logf("OK: buildHealthJSON returns %d processes, %d neuron files", len(health.Processes), health.NeuronFile)
}
