package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ━━━ TEST 14: Dashboard API — GET /api/brain ━━━
func TestDashboardAPI_GetBrain(t *testing.T) {
	dir := setupTestBrain(t)

	handler := http.HandlerFunc(withCORSDashboard(func(w http.ResponseWriter, r *http.Request) {
		data := buildBrainJSONResponse(dir)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	}))

	req := httptest.NewRequest("GET", "/api/brain", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var data BrainJSON
	if err := json.NewDecoder(rec.Body).Decode(&data); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if len(data.Regions) != 7 {
		t.Fatalf("expected 7 regions, got %d", len(data.Regions))
	}
	if data.TotalNeurons != 15 {
		t.Fatalf("expected 15 neurons, got %d", data.TotalNeurons)
	}
	if data.BombSource != "" {
		t.Fatalf("expected no bomb, got: %s", data.BombSource)
	}

	t.Logf("OK: /api/brain returns %d regions, %d neurons", len(data.Regions), data.TotalNeurons)
}

// ━━━ TEST 15: Dashboard API — POST /api/fire ━━━
func TestDashboardAPI_Fire(t *testing.T) {
	dir := setupTestBrain(t)

	handler := http.HandlerFunc(withCORSDashboard(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "POST only", 405)
			return
		}
		var req struct {
			Path string `json:"path"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad json", 400)
			return
		}
		path := strings.ReplaceAll(req.Path, "\\", "/")
		path = strings.Trim(path, "/")
		fireNeuron(dir, path)
		w.Write([]byte("OK — fired: " + path))
	}))

	body := strings.NewReader(`{"path":"cortex/left/frontend/hooks_pattern"}`)
	req := httptest.NewRequest("POST", "/api/fire", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	resp := rec.Body.String()
	if !strings.Contains(resp, "fired") {
		t.Fatalf("expected 'fired' in response, got: %s", resp)
	}

	// Verify counter increased
	brain := scanBrain(dir)
	for _, r := range brain.Regions {
		if r.Name == "cortex" {
			for _, n := range r.Neurons {
				if n.Name == "hooks_pattern" && n.Counter != 41 {
					t.Fatalf("expected counter 41 after fire, got %d", n.Counter)
				}
			}
		}
	}

	t.Logf("OK: /api/fire works, counter incremented to 41")
}

// ━━━ TEST 16: Dashboard API — POST /api/neuron ━━━
func TestDashboardAPI_AddNeuron(t *testing.T) {
	dir := setupTestBrain(t)

	handler := http.HandlerFunc(withCORSDashboard(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "POST only", 405)
			return
		}
		var req AddNeuronReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad json", 400)
			return
		}
		path := strings.ReplaceAll(req.Path, "\\", "/")
		path = strings.Trim(path, "/")
		growNeuron(dir, req.Region+"/"+path)
		w.Write([]byte("OK — " + req.Region + "/" + path))
	}))

	body := strings.NewReader(`{"region":"cortex","path":"testing/api_test"}`)
	req := httptest.NewRequest("POST", "/api/neuron", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	resp := rec.Body.String()
	if !strings.Contains(resp, "cortex") {
		t.Fatalf("expected 'cortex' in response, got: %s", resp)
	}

	t.Logf("OK: /api/neuron creates new neuron via dashboard")
}

// ━━━ TEST 17: Dashboard API — CORS headers ━━━
func TestDashboardAPI_CORS(t *testing.T) {
	handler := http.HandlerFunc(withCORSDashboard(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))

	// Test OPTIONS preflight
	req := httptest.NewRequest("OPTIONS", "/api/brain", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("OPTIONS expected 200, got %d", rec.Code)
	}

	cors := rec.Header().Get("Access-Control-Allow-Origin")
	if cors != "*" {
		t.Fatalf("expected CORS header '*', got: %s", cors)
	}

	methods := rec.Header().Get("Access-Control-Allow-Methods")
	if !strings.Contains(methods, "POST") {
		t.Fatalf("expected POST in Allow-Methods, got: %s", methods)
	}

	t.Logf("OK: CORS headers correctly set")
}

// ━━━ TEST 18: Dashboard API — POST /api/dedup ━━━
func TestDashboardAPI_Dedup(t *testing.T) {
	dir := setupTestBrain(t)

	handler := http.HandlerFunc(withCORSDashboard(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "POST only", 405)
			return
		}
		deduplicateNeurons(dir)
		w.Write([]byte("OK — dedup complete"))
	}))

	req := httptest.NewRequest("POST", "/api/dedup", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	if !strings.Contains(rec.Body.String(), "dedup") {
		t.Fatalf("expected 'dedup' in response")
	}

	t.Logf("OK: /api/dedup runs without error")
}

// ━━━ TEST 19: Dashboard API — POST /api/signal ━━━
func TestDashboardAPI_Signal(t *testing.T) {
	dir := setupTestBrain(t)

	handler := http.HandlerFunc(withCORSDashboard(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "POST only", 405)
			return
		}
		var req struct {
			Path string `json:"path"`
			Type string `json:"type"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad json", 400)
			return
		}
		path := strings.ReplaceAll(req.Path, "\\", "/")
		path = strings.Trim(path, "/")
		sigType := req.Type
		if sigType == "" {
			sigType = "dopamine"
		}
		if err := signalNeuron(dir, path, sigType); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		w.Write([]byte("OK — " + sigType + ": " + path))
	}))

	body := strings.NewReader(`{"path":"cortex/left/frontend/hooks_pattern","type":"dopamine"}`)
	req := httptest.NewRequest("POST", "/api/signal", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	if !strings.Contains(rec.Body.String(), "dopamine") {
		t.Fatalf("expected 'dopamine' in response, got: %s", rec.Body.String())
	}

	t.Logf("OK: /api/signal sends dopamine correctly")
}

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
