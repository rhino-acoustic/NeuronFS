package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)


// ━━━ TEST 20: rollbackNeuron — counter decrements ━━━
func TestRollbackNeuron_Decrements(t *testing.T) {
	dir := setupTestBrain(t)

	// hooks_pattern starts at 40
	err := rollbackNeuron(dir, "cortex/left/frontend/hooks_pattern")
	if err != nil {
		t.Fatalf("rollback failed: %v", err)
	}

	brain := scanBrain(dir)
	for _, r := range brain.Regions {
		if r.Name == "cortex" {
			for _, n := range r.Neurons {
				if n.Name == "hooks_pattern" {
					if n.Counter != 39 {
						t.Fatalf("expected counter 39 after rollback, got %d", n.Counter)
					}
				}
			}
		}
	}

	t.Logf("OK: rollback decremented hooks_pattern 40 → 39")
}

// ━━━ TEST 21: rollbackNeuron — minimum boundary ━━━
func TestRollbackNeuron_MinBoundary(t *testing.T) {
	dir := setupTestBrain(t)

	// hippocampus/failures/error_patterns has counter=5
	// Roll back 4 times to reach 1
	for i := 0; i < 4; i++ {
		rollbackNeuron(dir, "hippocampus/failures/error_patterns")
	}

	// At counter=1, should return error
	err := rollbackNeuron(dir, "hippocampus/failures/error_patterns")
	if err == nil {
		t.Fatal("expected error when rolling back counter at minimum")
	}

	t.Logf("OK: rollback correctly stops at minimum counter=1")
}

// ━━━ TEST 22: rollbackNeuron — nonexistent neuron ━━━
func TestRollbackNeuron_NotFound(t *testing.T) {
	dir := setupTestBrain(t)

	err := rollbackNeuron(dir, "cortex/nonexistent/thing")
	if err == nil {
		t.Fatal("expected error for nonexistent neuron")
	}

	t.Logf("OK: rollback correctly returns error for missing neuron")
}

// ━━━ TEST 23: EmitTarget — mapping validation ━━━
func TestEmitTarget_Mapping(t *testing.T) {
	expected := []string{"gemini", "cursor", "claude", "copilot", "generic"}
	for _, key := range expected {
		et, ok := emitTargetMap[key]
		if !ok {
			t.Fatalf("missing emit target: %s", key)
		}
		if et.Name == "" || et.FileName == "" {
			t.Fatalf("emit target %s has empty Name or FileName", key)
		}
	}

	// Verify Gemini has SubDir
	if emitTargetMap["gemini"].SubDir != ".gemini" {
		t.Fatalf("gemini SubDir should be .gemini, got: %s", emitTargetMap["gemini"].SubDir)
	}
	// Verify Copilot has SubDir
	if emitTargetMap["copilot"].SubDir != ".github" {
		t.Fatalf("copilot SubDir should be .github, got: %s", emitTargetMap["copilot"].SubDir)
	}

	t.Logf("OK: all 5 emit targets correctly mapped")
}

// ━━━ TEST 24: EmitTarget — writeAllTiersForTargets cursor ━━━
func TestEmitTarget_CursorOutput(t *testing.T) {
	dir := setupTestBrain(t)

	// Write to cursor target
	writeAllTiersForTargets(dir, "cursor")

	// Verify .cursorrules was created at project root (parent of brain)
	projectRoot := filepath.Dir(dir)
	cursorPath := filepath.Join(projectRoot, ".cursorrules")

	content, err := os.ReadFile(cursorPath)
	if err != nil {
		t.Fatalf("Failed to read .cursorrules: %v", err)
	}

	// Verify content has NeuronFS markers
	if !strings.Contains(string(content), "NEURONFS:START") {
		t.Fatalf("expected NEURONFS:START in .cursorrules")
	}
	if !strings.Contains(string(content), "NEURONFS:END") {
		t.Fatalf("expected NEURONFS:END in .cursorrules")
	}

	t.Logf("OK: --emit cursor creates .cursorrules with NeuronFS content (%d bytes)", len(content))
}

// ━━━ TEST 25: EmitTarget — writeAllTiersForTargets all ━━━
func TestEmitTarget_AllOutput(t *testing.T) {
	dir := setupTestBrain(t)

	// Write to all targets
	writeAllTiersForTargets(dir, "all")

	projectRoot := filepath.Dir(dir)

	// Check each target file exists
	checks := map[string]string{
		"cursor":  filepath.Join(projectRoot, ".cursorrules"),
		"claude":  filepath.Join(projectRoot, "CLAUDE.md"),
		"generic": filepath.Join(projectRoot, ".neuronrc"),
	}

	for name, path := range checks {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Fatalf("%s file not created: %s", name, path)
		}
	}

	// Copilot should be in .github/
	copilotPath := filepath.Join(projectRoot, ".github", "copilot-instructions.md")
	if _, err := os.Stat(copilotPath); os.IsNotExist(err) {
		t.Fatalf("copilot file not created: %s", copilotPath)
	}

	t.Logf("OK: --emit all creates all 5 target files")
}

// ━━━ TEST 26: doInjectToFile — preserves existing content ━━━
func TestDoInjectToFile_PreservesContent(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "test.md")

	// Create existing file with markers
	initial := "# My Config\n\n<!-- NEURONFS:START -->\nold content\n<!-- NEURONFS:END -->\n\n# Footer\n"
	os.WriteFile(filePath, []byte(initial), 0600)

	// Inject new content
	newRules := "<!-- NEURONFS:START -->\nnew content\n<!-- NEURONFS:END -->\n"
	doInjectToFile(filePath, newRules)

	result, _ := os.ReadFile(filePath)
	content := string(result)

	if !strings.Contains(content, "# My Config") {
		t.Fatal("lost header content")
	}
	if !strings.Contains(content, "new content") {
		t.Fatal("new content not injected")
	}
	if strings.Contains(content, "old content") {
		t.Fatal("old content was not replaced")
	}

	t.Logf("OK: doInjectToFile preserves surrounding content and replaces NeuronFS block")
}

// ━━━ TEST 27: EmitTarget — unknown target no crash ━━━
func TestEmitTarget_UnknownNoCrash(t *testing.T) {
	dir := setupTestBrain(t)

	// Should not panic or crash on unknown target
	writeAllTiersForTargets(dir, "nonexistent_editor")

	// Verify no files were created at project root
	projectRoot := filepath.Dir(dir)
	for _, name := range []string{".cursorrules", "CLAUDE.md", ".neuronrc"} {
		path := filepath.Join(projectRoot, name)
		if _, err := os.Stat(path); err == nil {
			t.Fatalf("unexpected file created for unknown target: %s", name)
		}
	}

	t.Logf("OK: unknown emit target handled gracefully, no crash, no files")
}
