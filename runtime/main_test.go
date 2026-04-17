package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// helper: create temp brain structure matching actual scanBrain() expectations
// scanBrain() looks for exact folder names: brainstem, limbic, hippocampus, sensors, cortex, ego, prefrontal
// Neurons = folders containing N.neuron files (counter = N)
func setupTestBrain(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// VFS 초기화 (테스트에서는 Upper-only 모드)
	initVFS(dir)

	neuronDirs := []struct {
		path    string
		counter int
	}{
		{"brainstem/必/절대_폴백_금지", 103},
		{"brainstem/canon/no_simulation", 100},
		{"brainstem/reflexes/self_debug", 100},
		{"limbic/emotion_parser/긴급_상황_감지", 50},
		{"limbic/neurotransmitters/도파민_보상", 30},
		{"hippocampus/rewards/dopamine_log", 10},
		{"hippocampus/failures/error_patterns", 5},
		{"sensors/workspace/nas_write_cmd", 20},
		{"sensors/design/sandstone", 15},
		{"cortex/left/frontend/hooks_pattern", 40},
		{"cortex/left/frontend/accent_blue", 25},
		{"ego/tone/expert_concise", 60},
		{"ego/philosophy/transistor_gate", 45},
		{"prefrontal/active/current_sprint", 10},
		{"prefrontal/vision/long_term", 8},
	}

	for _, nd := range neuronDirs {
		fullDir := filepath.Join(dir, nd.path)
		os.MkdirAll(fullDir, 0750)
		counterFile := filepath.Join(fullDir, fmt.Sprintf("%d.neuron", nd.counter))
		os.WriteFile(counterFile, []byte{}, 0600)
	}

	os.WriteFile(filepath.Join(dir, "brainstem", "connect_limbic.axon"),
		[]byte("TARGET: limbic"), 0600)
	os.WriteFile(filepath.Join(dir, "limbic", "connect_brainstem.axon"),
		[]byte("TARGET: brainstem"), 0600)

	return dir
}

// ━━━ TEST 1: Normal — all 7 regions active ━━━
func TestNormal_AllRegionsActive(t *testing.T) {
	brain := scanBrain(setupTestBrain(t))
	result := runSubsumption(brain)

	if result.BombSource != "" {
		t.Fatalf("expected no bomb, got: %s", result.BombSource)
	}
	if len(result.ActiveRegions) != 7 {
		t.Fatalf("expected 7 active regions, got %d", len(result.ActiveRegions))
	}
	if result.FiredNeurons != result.TotalNeurons {
		t.Fatalf("expected all neurons fired: %d/%d", result.FiredNeurons, result.TotalNeurons)
	}
	if len(result.BlockedRegions) != 0 {
		t.Fatalf("expected 0 blocked, got %d", len(result.BlockedRegions))
	}

	t.Logf("OK: %d/%d neurons active, 7 regions", result.FiredNeurons, result.TotalNeurons)
}

// ━━━ TEST 2: P0 bomb blocks everything ━━━
func TestP0Bomb_AllBlocked(t *testing.T) {
	dir := setupTestBrain(t)

	bombDir := filepath.Join(dir, "brainstem", "必", "절대_폴백_금지")
	os.WriteFile(filepath.Join(bombDir, "bomb.neuron"), []byte{}, 0600)

	brain := scanBrain(dir)
	result := runSubsumption(brain)

	if result.BombSource == "" {
		t.Fatal("expected bomb detection, got none")
	}
	if !strings.Contains(result.BombSource, "brainstem") {
		t.Fatalf("bomb should be in brainstem, got: %s", result.BombSource)
	}
	if result.FiredNeurons != 0 {
		t.Fatalf("expected 0 fired neurons, got %d", result.FiredNeurons)
	}
	if len(result.ActiveRegions) != 0 {
		t.Fatalf("expected 0 active regions, got %d", len(result.ActiveRegions))
	}

	t.Logf("OK: bomb in brainstem, all %d neurons blocked", result.TotalNeurons)
}

// ━━━ TEST 3: limbic bomb — brainstem survives ━━━
func TestLimbicBomb_BrainstemSurvives(t *testing.T) {
	dir := setupTestBrain(t)

	bombDir := filepath.Join(dir, "limbic", "emotion_parser", "긴급_상황_감지")
	os.WriteFile(filepath.Join(bombDir, "bomb.neuron"), []byte{}, 0600)

	brain := scanBrain(dir)
	result := runSubsumption(brain)

	if result.BombSource == "" {
		t.Fatal("expected bomb detection")
	}
	if len(result.ActiveRegions) != 1 {
		t.Fatalf("expected 1 active region (brainstem), got %d", len(result.ActiveRegions))
	}
	if result.ActiveRegions[0].Name != "brainstem" {
		t.Fatalf("expected brainstem active, got: %s", result.ActiveRegions[0].Name)
	}
	if len(result.BlockedRegions) != 6 {
		t.Fatalf("expected 6 blocked regions, got %d", len(result.BlockedRegions))
	}
	// Folder=Neuron 원칙: brainstem 활성 뉴런 수를 동적 확인
	bsNeurons := 0
	for _, r := range result.ActiveRegions {
		if r.Name == "brainstem" {
			for _, n := range r.Neurons {
				if !n.IsDormant {
					bsNeurons++
				}
			}
		}
	}
	if result.FiredNeurons != bsNeurons {
		t.Fatalf("expected %d brainstem neurons fired, got %d", bsNeurons, result.FiredNeurons)
	}

	t.Logf("OK: brainstem alive (%d neurons), limbic~prefrontal blocked (%d regions)",
		result.FiredNeurons, len(result.BlockedRegions))
}

// ━━━ TEST 4: growNeuron creates new neuron ━━━
func TestGrowNeuron_CountIncreases(t *testing.T) {
	dir := setupTestBrain(t)

	brain1 := scanBrain(dir)
	result1 := runSubsumption(brain1)
	before := result1.TotalNeurons

	err := growNeuron(dir, "cortex/left/frontend/new_rule")
	if err != nil {
		t.Fatalf("growNeuron failed: %v", err)
	}

	brain2 := scanBrain(dir)
	result2 := runSubsumption(brain2)
	after := result2.TotalNeurons

	// growNeuron also calls logEpisode -> hippocampus/session_log/memoryN.neuron
	if after <= before {
		t.Fatalf("expected count to increase from %d, got %d", before, after)
	}

	newPath := filepath.Join(dir, "cortex", "left", "frontend", "new_rule", "1.neuron")
	if _, err := os.Stat(newPath); os.IsNotExist(err) {
		t.Fatal("expected 1.neuron to exist in new neuron folder")
	}

	t.Logf("OK: %d -> %d neurons, new_rule created", before, after)
}

// ━━━ TEST 5: delete neuron counter file ━━━
func TestDeleteNeuron_CountDecreases(t *testing.T) {
	dir := setupTestBrain(t)

	brain1 := scanBrain(dir)
	result1 := runSubsumption(brain1)
	before := result1.TotalNeurons

	// Remove the entire hooks_pattern folder to actually delete the neuron
	// (Axiom: Folder=Neuron, removing just the trace file doesn't destroy the neuron)
	os.RemoveAll(filepath.Join(dir, "cortex", "left", "frontend", "hooks_pattern"))

	brain2 := scanBrain(dir)
	result2 := runSubsumption(brain2)
	after := result2.TotalNeurons

	if after != before-1 {
		t.Fatalf("expected count %d -> %d, got %d", before, before-1, after)
	}

	t.Logf("OK: %d -> %d neurons, hooks_pattern gone", before, after)
}

// ━━━ TEST 6: emitBootstrap format ━━━
func TestEmitFormat_MarkersAndOrder(t *testing.T) {
	dir := setupTestBrain(t)
	brain := scanBrain(dir)
	result := runSubsumption(brain)
	rules := emitBootstrap(result, dir)

	if !strings.Contains(rules, "<!-- NEURONFS:START -->") {
		t.Fatal("missing start marker")
	}
	if !strings.Contains(rules, "<!-- NEURONFS:END -->") {
		t.Fatal("missing end marker")
	}
	if !strings.Contains(rules, "ALWAYS") {
		t.Fatal("missing ALWAYS section")
	}
	if !strings.Contains(rules, "절대 폴백 금지") {
		t.Logf("GENERATED RULES:\n%s\n", rules)
		t.Fatal("expected brainstem '절대 폴백 금지' neuron in ALWAYS")
	}

	t.Logf("OK: markers present, brainstem ALWAYS rendered")
}

// ━━━ TEST 7: bomb and restore recovery flow ━━━
func TestRecoveryFlow_BombAndRestore(t *testing.T) {
	dir := setupTestBrain(t)

	bombDir := filepath.Join(dir, "brainstem", "必", "절대_폴백_금지")
	bombFile := filepath.Join(bombDir, "bomb.neuron")
	os.WriteFile(bombFile, []byte{}, 0600)

	brainA := scanBrain(dir)
	resultA := runSubsumption(brainA)

	if resultA.BombSource == "" {
		t.Fatal("Phase A: expected bomb")
	}
	if resultA.FiredNeurons != 0 {
		t.Fatalf("Phase A: expected 0 fired, got %d", resultA.FiredNeurons)
	}

	os.Remove(bombFile)

	brainB := scanBrain(dir)
	resultB := runSubsumption(brainB)

	if resultB.BombSource != "" {
		t.Fatalf("Phase B: bomb should be gone, got: %s", resultB.BombSource)
	}
	if resultB.FiredNeurons != resultB.TotalNeurons {
		t.Fatalf("Phase B: expected all neurons, got %d/%d", resultB.FiredNeurons, resultB.TotalNeurons)
	}
	if len(resultB.ActiveRegions) != 7 {
		t.Fatalf("Phase B: expected 7 active, got %d", len(resultB.ActiveRegions))
	}

	t.Logf("OK: Phase A blocked (%d total), Phase B recovered (%d/%d)",
		resultA.TotalNeurons, resultB.FiredNeurons, resultB.TotalNeurons)
}

// ━━━ TEST 8: Axon crosslinks ━━━
func TestAxonCrosslinks(t *testing.T) {
	dir := setupTestBrain(t)
	brain := scanBrain(dir)

	totalAxons := 0
	for _, r := range brain.Regions {
		totalAxons += len(r.Axons)
	}

	if totalAxons < 2 {
		t.Fatalf("expected at least 2 axons, got %d", totalAxons)
	}

	found := false
	for _, r := range brain.Regions {
		if r.Name == "brainstem" {
			for _, a := range r.Axons {
				if strings.Contains(a, "limbic") {
					found = true
				}
			}
		}
	}
	if !found {
		t.Fatal("expected brainstem->limbic axon")
	}

	t.Logf("OK: %d axons, brainstem->limbic link verified", totalAxons)
}

// ━━━ TEST 9: invalid folders ignored ━━━
func TestInvalidFolders_Ignored(t *testing.T) {
	dir := setupTestBrain(t)

	os.MkdirAll(filepath.Join(dir, "random_stuff"), 0750)
	os.MkdirAll(filepath.Join(dir, "notes"), 0750)
	os.WriteFile(filepath.Join(dir, "random_stuff", "test.neuron"), []byte{}, 0600)

	brain := scanBrain(dir)

	for _, r := range brain.Regions {
		if r.Name == "random_stuff" || r.Name == "notes" {
			t.Fatalf("should ignore non-region folder: %s", r.Name)
		}
	}

	if len(brain.Regions) != 7 {
		t.Fatalf("expected 7 regions, got %d", len(brain.Regions))
	}

	t.Logf("OK: invalid folders ignored, 7 valid regions")
}

// ━━━ TEST 10: fireNeuron counter increment ━━━
func TestFireNeuron_CounterIncrement(t *testing.T) {
	dir := setupTestBrain(t)

	brain1 := scanBrain(dir)
	var counterBefore int
	for _, r := range brain1.Regions {
		if r.Name == "cortex" {
			for _, n := range r.Neurons {
				if n.Name == "hooks_pattern" {
					counterBefore = n.Counter
				}
			}
		}
	}

	if counterBefore != 40 {
		t.Fatalf("expected initial counter 40, got %d", counterBefore)
	}

	fireNeuron(dir, "cortex/left/frontend/hooks_pattern")

	brain2 := scanBrain(dir)
	var counterAfter int
	for _, r := range brain2.Regions {
		if r.Name == "cortex" {
			for _, n := range r.Neurons {
				if n.Name == "hooks_pattern" {
					counterAfter = n.Counter
				}
			}
		}
	}

	if counterAfter != 41 {
		t.Fatalf("expected counter 41 after fire, got %d", counterAfter)
	}

	if _, err := os.Stat(filepath.Join(dir, "cortex", "left", "frontend", "hooks_pattern", "40.neuron")); !os.IsNotExist(err) {
		t.Fatal("old 40.neuron should be deleted")
	}
	if _, err := os.Stat(filepath.Join(dir, "cortex", "left", "frontend", "hooks_pattern", "41.neuron")); os.IsNotExist(err) {
		t.Fatal("new 41.neuron should exist")
	}

	t.Logf("OK: hooks_pattern counter %d -> %d", counterBefore, counterAfter)
}

// ━━━ TEST 11: signalNeuron dopamine ━━━
func TestSignalDopamine(t *testing.T) {
	dir := setupTestBrain(t)

	err := signalNeuron(dir, "cortex/left/frontend/hooks_pattern", "dopamine")
	if err != nil {
		t.Fatalf("signalNeuron failed: %v", err)
	}

	dopaFile := filepath.Join(dir, "cortex", "left", "frontend", "hooks_pattern", "dopamine1.neuron")
	if _, err := os.Stat(dopaFile); os.IsNotExist(err) {
		t.Fatal("expected dopamine1.neuron to exist")
	}

	err = signalNeuron(dir, "cortex/left/frontend/hooks_pattern", "dopamine")
	if err != nil {
		t.Fatalf("second signalNeuron failed: %v", err)
	}

	dopaFile2 := filepath.Join(dir, "cortex", "left", "frontend", "hooks_pattern", "dopamine2.neuron")
	if _, err := os.Stat(dopaFile2); os.IsNotExist(err) {
		t.Fatal("expected dopamine2.neuron to exist")
	}

	t.Logf("OK: dopamine signals created correctly")
}

// ━━━ TEST 12: Jaccard similarity ━━━
func TestJaccardSimilarity(t *testing.T) {
	tests := []struct {
		a, b     []string
		expected float64
	}{
		{[]string{"no", "console", "log"}, []string{"no", "console", "log"}, 1.0},
		{[]string{"no", "console", "log"}, []string{"no", "console"}, 0.6666},
		{[]string{"a", "b"}, []string{"c", "d"}, 0.0},
		{[]string{}, []string{"a"}, 0.0},
	}

	for _, tt := range tests {
		result := jaccardSimilarity(tt.a, tt.b)
		diff := result - tt.expected
		if diff < 0 {
			diff = -diff
		}
		if diff > 0.01 {
			t.Errorf("jaccardSimilarity(%v, %v) = %.4f, expected %.4f", tt.a, tt.b, result, tt.expected)
		}
	}

	t.Logf("OK: Jaccard similarity calculations correct")
}

// ━━━ TEST 13: tokenize + stem ━━━
func TestTokenizeAndStem(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"no_console_logging", []string{"no", "console", "log"}},
		{"verify_before_deliver", []string{"verify", "before", "deliver"}},
		{"self_debug", []string{"self", "debug"}},
	}

	for _, tt := range tests {
		result := tokenize(tt.input)
		if len(result) != len(tt.expected) {
			t.Errorf("tokenize(%q) = %v, expected %v", tt.input, result, tt.expected)
			continue
		}
		for i, r := range result {
			if r != tt.expected[i] {
				t.Errorf("tokenize(%q)[%d] = %q, expected %q", tt.input, i, r, tt.expected[i])
			}
		}
	}

	t.Logf("OK: tokenize + stem working correctly")
}
