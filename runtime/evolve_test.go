package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ?Å‚îÅ??Evolve Engine Unit Tests (Groq-independent) ?Å‚îÅ??
// ?Å‚îÅ??TEST 32: truncate ?Å‚îÅ??func TestTruncate(t *testing.T) {
	if truncate("hello", 10) != "hello" {
		t.Fatal("short string should not be truncated")
	}
	result := truncate("hello world", 5)
	if result != "hello..." {
		t.Fatalf("expected 'hello...', got '%s'", result)
	}
	t.Logf("OK: truncate works correctly")
}

// ?Å‚îÅ??TEST 33: boolStr ?Å‚îÅ??func TestBoolStr(t *testing.T) {
	if boolStr(true, "yes", "no") != "yes" {
		t.Fatal("true case failed")
	}
	if boolStr(false, "yes", "no") != "no" {
		t.Fatal("false case failed")
	}
	t.Logf("OK: boolStr returns correct value")
}

// ?Å‚îÅ??TEST 34: actionIcon ?Å‚îÅ??func TestActionIcon(t *testing.T) {
	cases := map[string]string{
		"grow":   "?å±",
		"fire":   "?î•",
		"signal": "?ì°",
		"prune":  "?í§",
		"decay":  "?í§",
		"merge":  "?îó",
		"other":  "??,
	}
	for input, expected := range cases {
		if got := actionIcon(input); got != expected {
			t.Fatalf("actionIcon(%s) = %s, want %s", input, got, expected)
		}
	}
	t.Logf("OK: actionIcon maps all types correctly")
}

// ?Å‚îÅ??TEST 35: collectEpisodes ?Å‚îÅ??func TestCollectEpisodes_Empty(t *testing.T) {
	dir := setupTestBrain(t)
	episodes := collectEpisodes(dir)
	// Test brain may or may not have episodes ??just verify no crash
	t.Logf("OK: collectEpisodes returned %d episodes (no crash)", len(episodes))
}

// ?Å‚îÅ??TEST 36: collectEpisodes with data ?Å‚îÅ??func TestCollectEpisodes_WithData(t *testing.T) {
	dir := setupTestBrain(t)
	logDir := filepath.Join(dir, "hippocampus", "session_log")
	os.MkdirAll(logDir, 0755)
	os.WriteFile(filepath.Join(logDir, "memory1.neuron"), []byte("episode 1: test fired"), 0644)
	os.WriteFile(filepath.Join(logDir, "memory2.neuron"), []byte("episode 2: grow completed"), 0644)
	os.WriteFile(filepath.Join(logDir, "memory3.neuron"), []byte("episode 3: evolve dry run"), 0644)

	episodes := collectEpisodes(dir)
	if len(episodes) != 3 {
		t.Fatalf("expected 3 episodes, got %d", len(episodes))
	}
	// Should be sorted chronologically
	if !strings.Contains(episodes[0], "episode 1") {
		t.Fatalf("episodes not sorted: first = %s", episodes[0])
	}
	if !strings.Contains(episodes[2], "episode 3") {
		t.Fatalf("episodes not sorted: last = %s", episodes[2])
	}
	t.Logf("OK: collectEpisodes returns sorted episodes")
}

// ?Å‚îÅ??TEST 37: buildBrainSummary ?Å‚îÅ??func TestBuildBrainSummary(t *testing.T) {
	dir := setupTestBrain(t)
	brain := scanBrain(dir)
	result := runSubsumption(brain)
	summary := buildBrainSummary(brain, result)

	if summary == "" {
		t.Fatal("buildBrainSummary returned empty string")
	}
	if !strings.Contains(summary, "Total neurons") {
		t.Fatal("summary missing 'Total neurons'")
	}
	if !strings.Contains(summary, "cortex") {
		t.Fatal("summary missing 'cortex' region")
	}
	t.Logf("OK: buildBrainSummary produces valid summary (%d bytes)", len(summary))
}

// ?Å‚îÅ??TEST 38: buildEvolvePrompt ?Å‚îÅ??func TestBuildEvolvePrompt(t *testing.T) {
	dir := setupTestBrain(t)
	brain := scanBrain(dir)
	result := runSubsumption(brain)
	summary := buildBrainSummary(brain, result)
	episodes := []string{"test episode 1", "test episode 2"}

	prompt := buildEvolvePrompt(episodes, summary, result)

	if prompt == "" {
		t.Fatal("buildEvolvePrompt returned empty string")
	}
	// Check key sections exist
	checks := []string{
		"NeuronFS Evolution Engine",
		"Subsumption Architecture",
		"STRICT RULES",
		"Current Brain State",
		"Episode Log",
		"valid JSON",
	}
	for _, check := range checks {
		if !strings.Contains(prompt, check) {
			t.Fatalf("prompt missing section: %s", check)
		}
	}
	t.Logf("OK: buildEvolvePrompt produces complete prompt (%d bytes)", len(prompt))
}

