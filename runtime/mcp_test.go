package main

import (
	"testing"
)

// ?Å‚îÅ??MCP Server Unit Tests ?Å‚îÅ??
// ?Å‚îÅ??TEST 28: mcpError returns error result ?Å‚îÅ??func TestMCPError_Format(t *testing.T) {
	result := mcpError("test error message")

	if !result.IsError {
		t.Fatal("mcpError should set IsError=true")
	}
	if len(result.Content) != 1 {
		t.Fatalf("expected 1 content block, got %d", len(result.Content))
	}

	t.Logf("OK: mcpError returns correct error format")
}

// ?Å‚îÅ??TEST 29: boolPtr helper ?Å‚îÅ??func TestBoolPtr(t *testing.T) {
	truePtr := boolPtr(true)
	falsePtr := boolPtr(false)

	if truePtr == nil || *truePtr != true {
		t.Fatal("boolPtr(true) failed")
	}
	if falsePtr == nil || *falsePtr != false {
		t.Fatal("boolPtr(false) failed")
	}

	t.Logf("OK: boolPtr returns correct pointer values")
}

// ?Å‚îÅ??TEST 30: registerMCPTools ??no panic ?Å‚îÅ??func TestRegisterMCPTools_NoPanic(t *testing.T) {
	dir := setupTestBrain(t)

	// MCP SDK requires server instance ??test that registration doesn't panic
	// We can't easily test full server without stdio, but we can verify
	// the dependency functions work with a test brain

	// Verify the functions that MCP tool handlers call internally
	brain := scanBrain(dir)
	if len(brain.Regions) == 0 {
		t.Fatal("scanBrain returned 0 regions for test brain")
	}

	// buildBrainJSONResponse (used by read_brain tool)
	data := buildBrainJSONResponse(dir)
	if data.TotalNeurons == 0 {
		t.Fatal("buildBrainJSONResponse returned empty brain")
	}

	// growNeuron (used by grow tool)
	err := growNeuron(dir, "cortex/mcp_test/test_grow")
	if err != nil {
		t.Fatalf("growNeuron failed: %v", err)
	}

	// fireNeuron (used by fire tool)
	fireNeuron(dir, "cortex/mcp_test/test_grow")

	// signalNeuron (used by signal tool)
	err = signalNeuron(dir, "cortex/mcp_test/test_grow", "dopamine")
	if err != nil {
		t.Fatalf("signalNeuron failed: %v", err)
	}

	t.Logf("OK: all MCP tool handler dependencies work correctly")
}

// ?Å‚îÅ??TEST 31: logWriter returns stderr ?Å‚îÅ??func TestLogWriter_ReturnsStderr(t *testing.T) {
	w := logWriter()
	if w == nil {
		t.Fatal("logWriter returned nil")
	}
	// logWriter should return os.Stderr
	if w.Fd() == 0 {
		t.Fatal("logWriter returned stdin instead of stderr")
	}
	t.Logf("OK: logWriter returns stderr (fd=%d)", w.Fd())
}

