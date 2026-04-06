// api_handlers.go — REST API 핸들러 (CRUD, Config, System)
//
// PROVIDES: registerCRUDRoutes, registerConfigRoutes, registerSystemRoutes, rollbackAll
// DEPENDS:  neuron_crud.go (growNeuron, fireNeuron, signalNeuron, rollbackNeuron)
//           lifecycle.go (runDecay, deduplicateNeurons)
//           brain.go (scanBrain, runSubsumption)
//           inject.go (autoReinject)
//           security/ (BuildChain, VerifyChain, loadOrCreateHMACKey)

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// CRUD Routes: grow, fire, signal, decay, state
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

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
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// Config Routes: principles, emotion, sandbox
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

func registerConfigRoutes(mux *http.ServeMux, brainRoot string, withCORS func(http.HandlerFunc) http.HandlerFunc) {
	// GET/POST /api/principles
	mux.HandleFunc("/api/principles", withCORS(func(w http.ResponseWriter, r *http.Request) {
		principlesFile := filepath.Join(brainRoot, "brainstem", "_principles.txt")

		if r.Method == "GET" {
			result := map[string]interface{}{"principles": []string{}}
			data, err := os.ReadFile(principlesFile)
			if err != nil {
				data, err = os.ReadFile(filepath.Join(brainRoot, "_preamble.txt"))
			}
			if err == nil {
				text := strings.TrimSpace(string(data))
				if text != "" {
					lines := []string{}
					for _, line := range strings.Split(text, "\n") {
						line = strings.TrimSpace(line)
						if line != "" {
							lines = append(lines, line)
						}
					}
					result["principles"] = lines
				}
			}
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			json.NewEncoder(w).Encode(result)
			return
		}
		if r.Method != "POST" {
			http.Error(w, "GET or POST only", 405)
			return
		}
		var req struct {
			Principles []string `json:"principles"`
		}
		json.NewDecoder(r.Body).Decode(&req)

		var clean []string
		for _, p := range req.Principles {
			p = strings.TrimSpace(p)
			if p != "" && len(clean) < 2 {
				clean = append(clean, p)
			}
		}

		if len(clean) == 0 {
			os.Remove(principlesFile)
			autoReinject(brainRoot)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			json.NewEncoder(w).Encode(map[string]string{"status": "cleared"})
			return
		}

		os.MkdirAll(filepath.Dir(principlesFile), 0755)
		os.WriteFile(principlesFile, []byte(strings.Join(clean, "\n")), 0644)

		autoReinject(brainRoot)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":     "applied",
			"principles": clean,
		})
	}))

	// GET/POST /api/emotion — 감정 상태 머신
	mux.HandleFunc("/api/emotion", withCORS(func(w http.ResponseWriter, r *http.Request) {
		stateFile := filepath.Join(brainRoot, "limbic", "_state.json")

		if r.Method == "GET" {
			result := map[string]interface{}{
				"emotion":   "neutral",
				"intensity": 0,
				"trigger":   "",
				"since":     "",
			}
			data, err := os.ReadFile(stateFile)
			if err == nil {
				json.Unmarshal(data, &result)
			}
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			json.NewEncoder(w).Encode(result)
			return
		}
		if r.Method != "POST" {
			http.Error(w, "GET or POST only", 405)
			return
		}
		var req struct {
			Emotion   string  `json:"emotion"`
			Intensity float64 `json:"intensity"`
			Trigger   string  `json:"trigger"`
		}
		json.NewDecoder(r.Body).Decode(&req)

		valid := map[string]bool{"분노": true, "긴급": true, "만족": true, "불안": true, "집중": true, "neutral": true}
		if !valid[req.Emotion] {
			http.Error(w, "invalid emotion: "+req.Emotion, 400)
			return
		}

		if req.Intensity <= 0 {
			req.Intensity = 0.5
		}
		if req.Intensity > 1 {
			req.Intensity = 1
		}

		state := map[string]interface{}{
			"emotion":   req.Emotion,
			"intensity": req.Intensity,
			"trigger":   req.Trigger,
			"since":     time.Now().Format("2006-01-02T15:04:05"),
		}

		os.MkdirAll(filepath.Dir(stateFile), 0755)
		stateBytes, _ := json.MarshalIndent(state, "", "  ")
		os.WriteFile(stateFile, stateBytes, 0644)

		emotionNeuronPath := "limbic/" + req.Emotion
		fireNeuron(brainRoot, emotionNeuronPath)

		autoReinject(brainRoot)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(state)
	}))

	// GET/POST /api/sandbox
	mux.HandleFunc("/api/sandbox", withCORS(func(w http.ResponseWriter, r *http.Request) {
		sandboxFile := filepath.Join(brainRoot, "brainstem", "_sandbox.txt")

		if r.Method == "GET" {
			result := map[string]interface{}{"rules": []string{}}
			data, err := os.ReadFile(sandboxFile)
			if err == nil {
				text := strings.TrimSpace(string(data))
				if text != "" {
					rules := []string{}
					for _, line := range strings.Split(text, "\n") {
						line = strings.TrimSpace(line)
						if line != "" {
							rules = append(rules, line)
						}
					}
					result["rules"] = rules
				}
			}
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			json.NewEncoder(w).Encode(result)
			return
		}
		if r.Method != "POST" {
			http.Error(w, "GET or POST only", 405)
			return
		}
		var req struct {
			Text string `json:"text"`
		}
		json.NewDecoder(r.Body).Decode(&req)

		text := strings.TrimSpace(req.Text)
		if text == "" {
			os.Remove(sandboxFile)
			os.RemoveAll(filepath.Join(brainRoot, "brainstem", "_sandbox"))
			autoReinject(brainRoot)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			json.NewEncoder(w).Encode(map[string]string{"status": "cleared"})
			return
		}

		os.MkdirAll(filepath.Dir(sandboxFile), 0755)
		os.WriteFile(sandboxFile, []byte(text), 0644)

		sandboxDir := filepath.Join(brainRoot, "brainstem", "_sandbox")
		os.RemoveAll(sandboxDir)
		created := 0
		var createdPaths []string

		for _, line := range strings.Split(text, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			name := strings.ReplaceAll(line, " ", "_")
			name = strings.Map(func(r rune) rune {
				if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' || (r >= 0xAC00 && r <= 0xD7AF) || (r >= 0x3131 && r <= 0x318E) {
					return r
				}
				return '_'
			}, name)
			if name == "" {
				continue
			}
			neuronDir := filepath.Join(sandboxDir, name)
			os.MkdirAll(neuronDir, 0755)
			os.WriteFile(filepath.Join(neuronDir, "1.neuron"), []byte{}, 0644)
			createdPaths = append(createdPaths, "brainstem/_sandbox/"+name)
			created++
		}

		autoReinject(brainRoot)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "applied",
			"created": created,
			"paths":   createdPaths,
		})
	}))
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// System Routes: integrity, evolution, community, reports, codemap
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

func registerSystemRoutes(mux *http.ServeMux, brainRoot string, withCORS func(http.HandlerFunc) http.HandlerFunc) {
	// GET /api/integrity?region=cortex
	mux.HandleFunc("/api/integrity", withCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "GET only", 405)
			return
		}

		hmacKey := loadOrCreateHMACKey(brainRoot)
		targetRegion := r.URL.Query().Get("region")

		regions := []string{"brainstem", "limbic", "hippocampus", "sensors", "cortex", "ego", "prefrontal"}
		if targetRegion != "" {
			if _, ok := regionPriority[targetRegion]; !ok {
				http.Error(w, `{"error":"invalid region"}`, 400)
				return
			}
			regions = []string{targetRegion}
		}

		type regionResult struct {
			Region   string `json:"region"`
			Status   string `json:"status"`
			RootHash string `json:"root_hash"`
			Files    int    `json:"files"`
			Detail   string `json:"detail,omitempty"`
		}

		var results []regionResult
		allOK := true

		for _, reg := range regions {
			regionPath := filepath.Join(brainRoot, reg)
			chain, err := BuildChain(regionPath, hmacKey)
			if err != nil {
				if errors.Is(err, errNoFiles) {
					results = append(results, regionResult{Region: reg, Status: "empty", Files: 0})
				} else {
					results = append(results, regionResult{Region: reg, Status: "error", Detail: err.Error()})
					allOK = false
				}
				continue
			}

			valid, brokenAt, verr := VerifyChain(chain, regionPath)
			if valid {
				results = append(results, regionResult{
					Region:   reg,
					Status:   "intact",
					RootHash: chain.RootHash[:16] + "...",
					Files:    len(chain.Nodes),
				})
			} else {
				detail := "chain verification failed"
				if brokenAt != "" {
					detail = fmt.Sprintf("broken at: %s", brokenAt)
				}
				if verr != nil {
					detail = verr.Error()
				}
				results = append(results, regionResult{
					Region: reg,
					Status: "violated",
					Files:  len(chain.Nodes),
					Detail: detail,
				})
				allOK = false
			}
		}

		overall := "NOMINAL"
		if !allOK {
			overall = "VIOLATION_DETECTED"
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  overall,
			"regions": results,
		})
	}))

	// POST /api/community — 외부 커뮤니티 트렌드를 뉴런으로 수집
	mux.HandleFunc("/api/community", withCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "POST only", 405)
			return
		}
		var req struct {
			Source  string `json:"source"`
			Topic   string `json:"topic"`
			Insight string `json:"insight"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Topic == "" {
			http.Error(w, `{"error":"topic required"}`, 400)
			return
		}
		safeTopic := strings.ReplaceAll(req.Topic, " ", "_")
		safeTopic = strings.ReplaceAll(safeTopic, "/", "_")
		safeTopic = strings.ReplaceAll(safeTopic, "\\", "_")

		neuronPath := fmt.Sprintf("cortex/community/%s/%s", req.Source, safeTopic)
		growNeuron(brainRoot, neuronPath)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"path":   neuronPath,
		})
	}))

	// POST /api/report — stackable report queue
	mux.HandleFunc("/api/report", withCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "POST only", 405)
			return
		}
		var req struct {
			Message  string `json:"message"`
			Priority string `json:"priority"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		if req.Message == "" {
			http.Error(w, "message required", 400)
			return
		}
		if req.Priority == "" {
			req.Priority = "normal"
		}
		reportsDir := filepath.Join(brainRoot, "_inbox", "reports")
		os.MkdirAll(reportsDir, 0755)
		ts := fmt.Sprintf("%d", time.Now().UnixMilli())
		filename := fmt.Sprintf("%s_%s.report", ts, req.Priority)
		content := fmt.Sprintf("priority: %s\ntimestamp: %s\n\n%s\n", req.Priority, time.Now().Format("2006-01-02 15:04:05"), req.Message)
		os.WriteFile(filepath.Join(reportsDir, filename), []byte(content), 0644)

		entries, _ := os.ReadDir(reportsDir)
		pending := 0
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), ".report") {
				pending++
			}
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "confirmed", "pending": pending, "priority": req.Priority,
		})
	}))

	// GET /api/reports — list pending
	mux.HandleFunc("/api/reports", withCORS(func(w http.ResponseWriter, r *http.Request) {
		reportsDir := filepath.Join(brainRoot, "_inbox", "reports")
		entries, _ := os.ReadDir(reportsDir)
		var reports []map[string]string
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), ".report") {
				data, _ := os.ReadFile(filepath.Join(reportsDir, e.Name()))
				reports = append(reports, map[string]string{"name": e.Name(), "content": string(data)})
			}
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"pending": len(reports), "reports": reports})
	}))

	// GET /api/evolution — Git-based neural evolution timeline
	mux.HandleFunc("/api/evolution", withCORS(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if _, err := exec.LookPath("git"); err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{"error": "git not found", "events": []interface{}{}})
			return
		}

		cmd := exec.Command("git", "log", "--pretty=format:%H|%ai|%s", "--name-status", "-n", "50")
		cmd.Dir = brainRoot
		out, err := cmd.Output()
		if err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error(), "events": []interface{}{}})
			return
		}

		type EvolutionEvent struct {
			Hash      string `json:"hash"`
			Timestamp string `json:"timestamp"`
			Message   string `json:"message"`
			Action    string `json:"action"`
			Path      string `json:"path"`
			Region    string `json:"region"`
		}

		var events []EvolutionEvent
		var currentHash, currentTime, currentMsg string

		for _, line := range strings.Split(string(out), "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			if parts := strings.SplitN(line, "|", 3); len(parts) == 3 && len(parts[0]) >= 40 {
				currentHash = parts[0][:8]
				currentTime = parts[1]
				currentMsg = parts[2]
				continue
			}

			if len(line) >= 2 && (line[0] == 'A' || line[0] == 'M' || line[0] == 'D') && line[1] == '\t' {
				filePath := line[2:]
				if !strings.HasSuffix(filePath, ".neuron") && !strings.HasSuffix(filePath, ".contra") &&
					!strings.HasSuffix(filePath, ".axon") && !strings.HasSuffix(filePath, ".goal") {
					continue
				}

				action := "modified"
				switch line[0] {
				case 'A':
					action = "created"
				case 'D':
					action = "suppressed"
				}

				region := ""
				pathParts := strings.SplitN(filePath, "/", 2)
				if len(pathParts) > 0 {
					if _, ok := regionPriority[pathParts[0]]; ok {
						region = pathParts[0]
					}
				}

				events = append(events, EvolutionEvent{
					Hash: currentHash, Timestamp: currentTime, Message: currentMsg,
					Action: action, Path: filePath, Region: region,
				})
			}
		}

		// Unstaged changes as "live" brainwaves
		diffCmd := exec.Command("git", "status", "--porcelain")
		diffCmd.Dir = brainRoot
		diffOut, err := diffCmd.Output()
		if err == nil {
			for _, line := range strings.Split(string(diffOut), "\n") {
				line = strings.TrimSpace(line)
				if len(line) < 4 {
					continue
				}
				filePath := strings.TrimSpace(line[2:])
				if !strings.HasSuffix(filePath, ".neuron") && !strings.HasSuffix(filePath, ".contra") {
					continue
				}

				action := "modified"
				switch line[0] {
				case '?', 'A':
					action = "created"
				case 'D':
					action = "suppressed"
				}

				region := ""
				pathParts := strings.SplitN(filePath, "/", 2)
				if len(pathParts) > 0 {
					if _, ok := regionPriority[pathParts[0]]; ok {
						region = pathParts[0]
					}
				}

				events = append([]EvolutionEvent{{
					Hash: "unstaged", Timestamp: time.Now().Format("2006-01-02 15:04:05 -0700"),
					Message: "🧠 Active brainwave (uncommitted)",
					Action: action, Path: filePath, Region: region,
				}}, events...)
			}
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"events": events,
			"total":  len(events),
		})
	}))

	// GET /api/retrieve — Hebbian Retrieval
	mux.HandleFunc("/api/retrieve", handleRetrieve(brainRoot))

	// GET /api/codemap — Runtime file tree snapshot
	mux.HandleFunc("/api/codemap", withCORS(func(w http.ResponseWriter, r *http.Request) {
		touchActivity()
		runtimeDir := filepath.Join(filepath.Dir(brainRoot), "runtime")
		entries, err := os.ReadDir(runtimeDir)
		if err != nil {
			http.Error(w, `{"error":"cannot read runtime dir"}`, 500)
			return
		}

		type FileEntry struct {
			Name  string `json:"name"`
			Lines int    `json:"lines"`
			Role  string `json:"role"`
		}
		var files []FileEntry
		totalLines := 0
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
				continue
			}
			data, err := os.ReadFile(filepath.Join(runtimeDir, e.Name()))
			if err != nil {
				continue
			}
			lines := strings.Count(string(data), "\n") + 1
			totalLines += lines
			role := ""
			for _, line := range strings.SplitN(string(data), "\n", 20) {
				if strings.Contains(line, "PROVIDES:") {
					role = strings.TrimSpace(strings.SplitN(line, "PROVIDES:", 2)[1])
					if len(role) > 80 {
						role = role[:80]
					}
					break
				}
			}
			files = append(files, FileEntry{Name: e.Name(), Lines: lines, Role: role})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"generated":  time.Now().Format(time.RFC3339),
			"totalFiles": len(files),
			"totalLines": totalLines,
			"files":      files,
		})
	}))

	// Expose pprof
	mux.HandleFunc("/debug/pprof/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.DefaultServeMux.ServeHTTP(w, r)
	}))
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// rollbackAll: 글로벌 Git 롤백 (quarantine fallback)
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

func rollbackAll(brainRoot string) error {
	cmd := exec.Command("git", "reset", "--hard", "HEAD~1")
	cmd.Dir = brainRoot
	_, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("\033[33m[WARNING] Hard reset failed! Initiating Quarantine Protocol...\033[0m\n")
		qBranch := fmt.Sprintf("quarantine-%s", time.Now().Format("20060102-150405"))
		
		cmdBranch := exec.Command("git", "checkout", "-b", qBranch)
		cmdBranch.Dir = brainRoot
		cmdBranch.Run()
		
		cmdAdd := exec.Command("git", "add", ".")
		cmdAdd.Dir = brainRoot
		cmdAdd.Run()
		
		cmdCommit := exec.Command("git", "commit", "-m", "Auto-quarantine corrupted state")
		cmdCommit.Dir = brainRoot
		cmdCommit.Run()
		
		fmt.Printf("\033[35m[QUARANTINE] Corrupted state isolated to branch: %s\033[0m\n", qBranch)
		
		cmdCheckout := exec.Command("git", "checkout", "main")
		cmdCheckout.Dir = brainRoot
		cmdCheckout.Run()
		
		return fmt.Errorf("git reset failed, isolated to %s. err: %v", qBranch, err)
	}
	
	cmdClean := exec.Command("git", "clean", "-fd")
	cmdClean.Dir = brainRoot
	cmdClean.Run()
	
	return nil
}
