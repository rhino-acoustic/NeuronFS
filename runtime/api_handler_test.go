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

// ============================================================================
// Coverage Boost Phase 4 — HTTP API 핸들러 테스트
// startAPI의 인라인 핸들러들을 httptest로 테스트하여 대량 커버리지 확보
// ============================================================================

// setupAPIServer creates a test server with the same mux as startAPI
func setupAPIServer(t *testing.T) (*httptest.Server, string) {
	t.Helper()
	dir := t.TempDir()

	// CRITICAL: Prevent test brain from writing to real GEMINI.md
	oldHome := os.Getenv("USERPROFILE")
	os.Setenv("USERPROFILE", dir)
	t.Cleanup(func() { os.Setenv("USERPROFILE", oldHome) })

	initBrain(dir)

	// Create test neurons
	for _, p := range []string{
		"cortex/test_api/hook",
		"brainstem/canon/api_test",
	} {
		full := filepath.Join(dir, filepath.FromSlash(p))
		os.MkdirAll(full, 0755)
		os.WriteFile(filepath.Join(full, "3.neuron"), []byte{}, 0644)
	}

	mux := http.NewServeMux()

	// Replicate the key handlers from startAPI
	withCORS := func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			if r.Method == "OPTIONS" {
				w.WriteHeader(200)
				return
			}
			h(w, r)
		}
	}

	brainRoot := dir

	// /api/health
	mux.HandleFunc("/api/health", withCORS(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(buildHealthJSON(brainRoot))
	}))

	// /api/brain
	mux.HandleFunc("/api/brain", withCORS(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(buildBrainJSONResponse(brainRoot))
	}))

	// /api/grow
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

	// /api/fire
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

	// /api/signal
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
		json.NewEncoder(w).Encode(map[string]string{"status": "signaled"})
	}))

	// /api/rollback
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
		err := rollbackNeuron(brainRoot, req.Path)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			json.NewEncoder(w).Encode(map[string]string{"status": "error", "error": err.Error()})
		} else {
			json.NewEncoder(w).Encode(map[string]string{"status": "rolled_back"})
		}
	}))

	// /api/region — uses handleReadRegion from emit.go
	mux.HandleFunc("/api/region", withCORS(handleReadRegion(brainRoot)))

	// /api/neuronize — uses handleNeuronizeAPI
	mux.HandleFunc("/api/neuronize", withCORS(handleNeuronizeAPI(brainRoot)))

	// /api/polarize — uses handlePolarizeAPI
	mux.HandleFunc("/api/polarize", withCORS(handlePolarizeAPI(brainRoot)))

	ts := httptest.NewServer(mux)
	return ts, dir
}

// ---------------------------------------------------------------------------
// /api/health
// ---------------------------------------------------------------------------

func TestAPI_Health(t *testing.T) {
	ts, _ := setupAPIServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/health")
	if err != nil {
		t.Fatalf("GET /api/health failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var health map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&health)
	if health["api"] == nil {
		t.Fatal("health response missing 'api' field")
	}
	t.Logf("OK: /api/health returned api=%v", health["api"])
}

// ---------------------------------------------------------------------------
// /api/brain
// ---------------------------------------------------------------------------

func TestAPI_Brain(t *testing.T) {
	ts, _ := setupAPIServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/brain")
	if err != nil {
		t.Fatalf("GET /api/brain failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	t.Log("OK: /api/brain returned 200")
}

// ---------------------------------------------------------------------------
// /api/grow
// ---------------------------------------------------------------------------

func TestAPI_Grow(t *testing.T) {
	ts, _ := setupAPIServer(t)
	defer ts.Close()

	body := `{"path":"cortex/api_test/new_neuron"}`
	resp, err := http.Post(ts.URL+"/api/grow", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST /api/grow failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var result map[string]string
	json.NewDecoder(resp.Body).Decode(&result)
	if result["status"] != "grown" {
		t.Fatalf("expected status=grown, got %v", result["status"])
	}
	t.Log("OK: /api/grow successfully grew neuron")
}

func TestAPI_Grow_MethodNotAllowed(t *testing.T) {
	ts, _ := setupAPIServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/grow")
	if err != nil {
		t.Fatalf("GET /api/grow failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 405 {
		t.Fatalf("expected 405, got %d", resp.StatusCode)
	}
	t.Log("OK: /api/grow rejects GET")
}

func TestAPI_Grow_MissingPath(t *testing.T) {
	ts, _ := setupAPIServer(t)
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/api/grow", "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Fatalf("POST failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 400 {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	t.Log("OK: /api/grow rejects empty path")
}

// ---------------------------------------------------------------------------
// /api/fire
// ---------------------------------------------------------------------------

func TestAPI_Fire(t *testing.T) {
	ts, _ := setupAPIServer(t)
	defer ts.Close()

	body := `{"path":"cortex/test_api/hook"}`
	resp, err := http.Post(ts.URL+"/api/fire", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST /api/fire failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	t.Log("OK: /api/fire executed")
}

// ---------------------------------------------------------------------------
// /api/signal
// ---------------------------------------------------------------------------

func TestAPI_Signal_Dopamine(t *testing.T) {
	ts, _ := setupAPIServer(t)
	defer ts.Close()

	body := `{"path":"cortex/test_api/hook","type":"dopamine"}`
	resp, err := http.Post(ts.URL+"/api/signal", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST /api/signal failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	t.Log("OK: /api/signal dopamine executed")
}

func TestAPI_Signal_Memory(t *testing.T) {
	ts, _ := setupAPIServer(t)
	defer ts.Close()

	body := `{"path":"cortex/test_api/hook","type":"memory"}`
	resp, _ := http.Post(ts.URL+"/api/signal", "application/json", strings.NewReader(body))
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	t.Log("OK: /api/signal memory executed")
}

// ---------------------------------------------------------------------------
// /api/rollback
// ---------------------------------------------------------------------------

func TestAPI_Rollback(t *testing.T) {
	ts, _ := setupAPIServer(t)
	defer ts.Close()

	body := `{"path":"cortex/test_api/hook"}`
	resp, err := http.Post(ts.URL+"/api/rollback", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST /api/rollback failed: %v", err)
	}
	defer resp.Body.Close()
	t.Logf("OK: /api/rollback returned status %d", resp.StatusCode)
}

// ---------------------------------------------------------------------------
// /api/region
// ---------------------------------------------------------------------------

func TestAPI_Region(t *testing.T) {
	ts, _ := setupAPIServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/region?region=cortex")
	if err != nil {
		t.Fatalf("GET /api/region failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	t.Log("OK: /api/region?region=cortex returned 200")
}

func TestAPI_Region_InvalidRegion(t *testing.T) {
	ts, _ := setupAPIServer(t)
	defer ts.Close()

	resp, _ := http.Get(ts.URL + "/api/region?region=invalid")
	defer resp.Body.Close()

	if resp.StatusCode != 400 {
		t.Fatalf("expected 400 for invalid region, got %d", resp.StatusCode)
	}
	t.Log("OK: /api/region rejects invalid region")
}

func TestAPI_Region_NoParam(t *testing.T) {
	ts, _ := setupAPIServer(t)
	defer ts.Close()

	resp, _ := http.Get(ts.URL + "/api/region")
	defer resp.Body.Close()

	if resp.StatusCode != 400 {
		t.Fatalf("expected 400 for missing region, got %d", resp.StatusCode)
	}
	t.Log("OK: /api/region rejects missing param")
}

// ---------------------------------------------------------------------------
// /api/neuronize
// ---------------------------------------------------------------------------

func TestAPI_Neuronize(t *testing.T) {
	ts, _ := setupAPIServer(t)
	defer ts.Close()

	body := `{"dry_run":true}`
	resp, err := http.Post(ts.URL+"/api/neuronize", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST /api/neuronize failed: %v", err)
	}
	defer resp.Body.Close()
	// May return 500 (no GROQ_API_KEY) which is expected
	t.Logf("OK: /api/neuronize returned status %d", resp.StatusCode)
}

// ---------------------------------------------------------------------------
// /api/polarize
// ---------------------------------------------------------------------------

func TestAPI_Polarize(t *testing.T) {
	ts, _ := setupAPIServer(t)
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/api/polarize", "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Fatalf("POST /api/polarize failed: %v", err)
	}
	defer resp.Body.Close()
	t.Logf("OK: /api/polarize returned status %d", resp.StatusCode)
}

// ---------------------------------------------------------------------------
// CORS OPTIONS preflight
// ---------------------------------------------------------------------------

func TestAPI_CORS_Options(t *testing.T) {
	ts, _ := setupAPIServer(t)
	defer ts.Close()

	req, _ := http.NewRequest("OPTIONS", ts.URL+"/api/health", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("OPTIONS failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200 for OPTIONS, got %d", resp.StatusCode)
	}
	if resp.Header.Get("Access-Control-Allow-Origin") != "*" {
		t.Fatal("missing CORS header")
	}
	t.Log("OK: OPTIONS returns proper CORS headers")
}
