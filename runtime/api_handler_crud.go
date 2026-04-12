package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func registerCRUDRoutes(mux *http.ServeMux, brainRoot string, withCORS func(http.HandlerFunc) http.HandlerFunc) {
	// POST /api/grow  {"path": "cortex/frontend/coding/no_console_log"}
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

	// POST /api/fire  {"path": "cortex/frontend/coding/no_console_log"}
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

	// POST /api/signal  {"path": "...", "type": "dopamine|bomb|memory"}
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
		json.NewEncoder(w).Encode(map[string]string{"status": "signaled", "path": req.Path, "type": req.Type})
	}))

	// POST /api/decay  {"days": 30}
	mux.HandleFunc("/api/decay", withCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "POST only", 405)
			return
		}
		var req struct {
			Days int `json:"days"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		if req.Days <= 0 {
			req.Days = 30
		}
		runDecay(brainRoot, req.Days)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "decay_complete", "days": req.Days})
	}))

	// GET /api/state — current brain state JSON
	mux.HandleFunc("/api/state", withCORS(func(w http.ResponseWriter, r *http.Request) {
		stateFile := filepath.Join(brainRoot, "..", "brain_state.json")
		abs, _ := filepath.Abs(stateFile)
		data, err := os.ReadFile(abs)
		if err != nil {
			http.Error(w, `{"error":"brain_state.json not found"}`, 404)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	}))

	// POST /api/evolve  {"dry_run": false}
	mux.HandleFunc("/api/evolve", withCORS(handleEvolveAPI(brainRoot)))

	// POST /api/neuronize {"dry_run": true}
	mux.HandleFunc("/api/neuronize", withCORS(handleNeuronizeAPI(brainRoot)))

	// POST /api/polarize
	mux.HandleFunc("/api/polarize", withCORS(handlePolarizeAPI(brainRoot)))

	// POST /api/dedup — 중복 뉴런 Jaccard 병합
	mux.HandleFunc("/api/dedup", withCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "POST only", 405)
			return
		}
		deduplicateNeurons(brainRoot)
		brain := scanBrain(brainRoot)
		result := runSubsumption(brain)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":     "ok",
			"neurons":    result.TotalNeurons,
			"activation": result.TotalCounter,
		})
	}))

	// GET /api/read?region=cortex — read region rules + auto-fire top neurons
	mux.HandleFunc("/api/read", withCORS(handleReadRegion(brainRoot)))

	// POST /api/inject — Re-scan brain + inject into GEMINI.md
	mux.HandleFunc("/api/inject", withCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "POST only", 405)
			return
		}
		autoReinject(brainRoot)
		brain := scanBrain(brainRoot)
		result := runSubsumption(brain)
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "Injected %d neurons, activation: %d", result.TotalNeurons, result.TotalCounter)
	}))

	// POST /api/rollback {"path": "cortex/..."}
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
		if err := rollbackNeuron(brainRoot, req.Path); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(400)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error(), "path": req.Path})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "rolled_back", "path": req.Path})
	}))

	// POST /api/rollback/all — Full system rollback via Git
	mux.HandleFunc("/api/rollback/all", withCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "POST only", 405)
			return
		}
		if err := rollbackAll(brainRoot); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(500)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error(), "status": "failed"})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "rolled_back", "message": "Global git rollback executed successfully (P0 included)."})
	}))

	// GET /api/health — system process health check
	mux.HandleFunc("/api/health", withCORS(func(w http.ResponseWriter, r *http.Request) {
		health := buildHealthJSON(brainRoot)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(health)
	}))

	// GET /api/brain — full brain state for dashboard
	mux.HandleFunc("/api/brain", withCORS(func(w http.ResponseWriter, r *http.Request) {
		data := buildBrainJSONResponse(brainRoot)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	}))

	// GET /api/usage — API usage stats + system metrics for dashboard
	mux.HandleFunc("/api/usage", withCORS(func(w http.ResponseWriter, r *http.Request) {
		// Groq usage from neuronize.go atomic counters
		groq := GetGroqUsage()

		// Emotion state
		emotion := map[string]interface{}{"emotion": "neutral", "intensity": 0}
		stateFile := filepath.Join(brainRoot, "limbic", "_state.json")
		if data, err := os.ReadFile(stateFile); err == nil {
			json.Unmarshal(data, &emotion)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"groq":    groq,
			"emotion": emotion,
			"uptime":  time.Since(startTime).Round(time.Second).String(),
		})
	}))

	// GET /api/skills — list all learned skills
	// POST /api/skills — learn a new skill {category, name, pattern, source}
	mux.HandleFunc("/api/skills", withCORS(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "POST" {
			var req struct {
				Category string `json:"category"`
				Name     string `json:"name"`
				Pattern  string `json:"pattern"`
				Source   string `json:"source"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
				http.Error(w, `{"error":"name required"}`, 400)
				return
			}
			if req.Category == "" {
				req.Category = "general"
			}
			if err := LearnSkill(brainRoot, req.Category, req.Name, req.Pattern, req.Source); err != nil {
				http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), 500)
				return
			}
			json.NewEncoder(w).Encode(map[string]string{"status": "learned", "category": req.Category, "name": req.Name})
			return
		}
		// GET
		skills, _ := RecallAllSkills(brainRoot)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"total":  len(skills),
			"skills": skills,
		})
	}))
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// Config Routes: principles, emotion, sandbox
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

