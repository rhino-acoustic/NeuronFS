package main

// ━━━ api_server.go ━━━
// Module: REST API Server + Rollback
//
// PROVIDES:
//   startAPI, rollbackAll
//
// CONSUMED BY:
//   main.go → main() starts API on --supervisor mode
//
// DEPENDS ON:
//   main.go         → scanBrain(), runSubsumption(), growNeuron(), fireNeuron()
//   main.go         → signalNeuron(), touchActivity(), markBrainDirty()
//   lifecycle.go    → deduplicateNeurons()
//   emit.go         → writeAllTiers()

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
// REST API: Programmatic growth for n8n/dashboard/webhooks
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
func startAPI(brainRoot string, port int) {
	mux := http.NewServeMux()

	// Initialize activity tracker
	touchActivity()

	// CORS middleware with activity tracking
	withCORS := func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			if r.Method == "OPTIONS" {
				w.WriteHeader(200)
				return
			}
			// Track activity (skip dashboard polling to avoid resetting idle timer)
			if r.URL.Path != "/api/brain" && r.URL.Path != "/api/state" && r.URL.Path != "/favicon.ico" {
				touchActivity()
			}
			h(w, r)
		}
	}

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

	// POST /api/neuronize {"dry_run": true} — Groq 기반 contra 뉴런 자동 생성
	mux.HandleFunc("/api/neuronize", withCORS(handleNeuronizeAPI(brainRoot)))

	// POST /api/polarize — 긍정형→부정형 전환 대상 조회
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

	// GET /api/read?region=cortex — read region rules + auto-fire top neurons (RAG retrieval)
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

	// GET/POST /api/principles — 대원칙 설정 (brainstem TOP 규칙)
	// _principles.txt에 최대 2줄 저장. emit 시 GEMINI.md 최상단에 주입됨
	mux.HandleFunc("/api/principles", withCORS(func(w http.ResponseWriter, r *http.Request) {
		principlesFile := filepath.Join(brainRoot, "brainstem", "_principles.txt")

		if r.Method == "GET" {
			result := map[string]interface{}{"principles": []string{}}
			data, err := os.ReadFile(principlesFile)
			if err != nil {
				// fallback: _preamble.txt (이전 형식)
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

		// 최대 2줄만 허용
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

		// 대원칙을 brainstem 뉴런으로도 생성 (가중치 부여)
		for _, p := range clean {
			name := strings.ReplaceAll(p, " ", "_")
			name = strings.Map(func(r rune) rune {
				if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' || (r >= 0xAC00 && r <= 0xD7AF) || (r >= 0x3131 && r <= 0x318E) || r == 0x7981 || r == 0x5FC5 {
					return r
				}
				return '_'
			}, name)
			neuronDir := filepath.Join(brainRoot, "brainstem", name)
			if _, err := os.Stat(neuronDir); os.IsNotExist(err) {
				os.MkdirAll(neuronDir, 0755)
				os.WriteFile(filepath.Join(neuronDir, "50.neuron"), []byte{}, 0644) // 높은 초기 가중치
			}
		}

		autoReinject(brainRoot)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":     "applied",
			"principles": clean,
		})
	}))

	// GET/POST /api/emotion — 감정 상태 머신 (limbic/_state.json)
	// 감정: 분노/긴급/만족/불안/집중/neutral. emit 시 해당 감정의 하위 행동 뉴런을 강화
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

		// 감정 뉴런 발화 (해당 감정 카운터 증가)
		emotionNeuronPath := "limbic/" + req.Emotion
		fireNeuron(brainRoot, emotionNeuronPath)

		autoReinject(brainRoot)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(state)
	}))

	// Uses _sandbox.txt (raw text) instead of folder names to preserve emojis/special chars
	mux.HandleFunc("/api/sandbox", withCORS(func(w http.ResponseWriter, r *http.Request) {
		sandboxFile := filepath.Join(brainRoot, "brainstem", "_sandbox.txt")

		if r.Method == "GET" {
			// Return current sandbox content from text file
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
			// Empty = delete sandbox, reinject
			os.Remove(sandboxFile)
			// Also clean up legacy folder-based sandbox
			os.RemoveAll(filepath.Join(brainRoot, "brainstem", "_sandbox"))
			autoReinject(brainRoot)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			json.NewEncoder(w).Encode(map[string]string{"status": "cleared"})
			return
		}

		// 1. Write raw text to _sandbox.txt (preserves emojis for API GET)
		os.MkdirAll(filepath.Dir(sandboxFile), 0755)
		os.WriteFile(sandboxFile, []byte(text), 0644)

		// 2. Create neuron folders in _sandbox/ (for brain scan + dashboard)
		sandboxDir := filepath.Join(brainRoot, "brainstem", "_sandbox")
		os.RemoveAll(sandboxDir)
		created := 0
		var createdPaths []string

		for _, line := range strings.Split(text, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			// Sanitize for folder name: keep Korean, alphanumeric, underscore, dash
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

	// POST /api/rollback {\"path\": \"cortex/...\"} — decrement neuron counter (min=1)
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

	// POST /api/rollback/all — Full system rollback via Git (brainstem included)
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

	// GET /api/integrity?region=cortex — Merkle Hash Chain 무결성 검증
	// 전체: GET /api/integrity (7개 영역 모두 검증)
	// 단일: GET /api/integrity?region=cortex
	mux.HandleFunc("/api/integrity", withCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "GET only", 405)
			return
		}

		// HMAC 키 로드 (없으면 자동 생성)
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
			Status   string `json:"status"`    // "intact" | "violated" | "empty"
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

			// 검증: chain 재구축하여 비교
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

	// GET / — Dashboard HTML (same as --dashboard mode)
	// Static files: 3D dashboard, brain.obj, brain_state.json
	neuronfsRoot := filepath.Dir(brainRoot) // NeuronFS/ directory (parent of brain_v4)
	mux.HandleFunc("/3d", withCORS(func(w http.ResponseWriter, r *http.Request) {
		htmlPath := filepath.Join(neuronfsRoot, "brain_dashboard.html")
		data, err := os.ReadFile(htmlPath)
		if err != nil {
			http.Error(w, "brain_dashboard.html not found", 404)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(data)
	}))
	mux.HandleFunc("/brain.obj", withCORS(func(w http.ResponseWriter, r *http.Request) {
		objPath := filepath.Join(neuronfsRoot, "brain.obj")
		data, err := os.ReadFile(objPath)
		if err != nil {
			http.Error(w, "brain.obj not found", 404)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		w.Write(data)
	}))
	mux.HandleFunc("/brain_state.json", withCORS(func(w http.ResponseWriter, r *http.Request) {
		jsonPath := filepath.Join(neuronfsRoot, "brain_state.json")
		data, err := os.ReadFile(jsonPath)
		if err != nil {
			http.Error(w, "brain_state.json not found", 404)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	}))

	// GET /v2 — Dashboard V2 (experimental)
	mux.HandleFunc("/v2", withCORS(func(w http.ResponseWriter, r *http.Request) {
		v2Path := filepath.Join(neuronfsRoot, "dashboard_v2.html")
		data, err := os.ReadFile(v2Path)
		if err != nil {
			data, err = os.ReadFile(`C:\Users\BASEMENT_ADMIN\NeuronFS\dashboard_v2.html`)
		}
		if err != nil {
			http.Error(w, "dashboard_v2.html not found", 404)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(data)
	}))
	// GET / — Main Dashboard (brain_dashboard.html)
	mux.HandleFunc("/", withCORS(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		if r.URL.Path == "/favicon.ico" || r.URL.Path == "/manifest.json" {
			w.WriteHeader(204)
			return
		}
		htmlPath := filepath.Join(neuronfsRoot, "brain_dashboard.html")
		data, err := os.ReadFile(htmlPath)
		if err != nil {
			data, err = os.ReadFile(`C:\Users\BASEMENT_ADMIN\NeuronFS\brain_dashboard.html`)
		}
		if err != nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprint(w, dashboardHTML)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(data)
	}))

	// GET /cards — Card-only dashboard (legacy)
	mux.HandleFunc("/cards", withCORS(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, dashboardHTML)
	}))

	// POST /api/community — 외부 커뮤니티 트렌드를 뉴런으로 수집
	// Body: {"source":"github|reddit|hackernews","topic":"AI memory","insight":"..."}
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
		// 안전한 경로 생성
		safeTopic := strings.ReplaceAll(req.Topic, " ", "_")
		safeTopic = strings.ReplaceAll(safeTopic, "/", "_")
		safeTopic = strings.ReplaceAll(safeTopic, "\\", "_")

		neuronPath := filepath.Join(brainRoot, "cortex", "community", req.Source, safeTopic)
		os.MkdirAll(neuronPath, 0755)

		// 카운터 파일 생성/증가
		files, _ := filepath.Glob(filepath.Join(neuronPath, "*.neuron"))
		counter := len(files) + 1
		counterFile := filepath.Join(neuronPath, fmt.Sprintf("%d.neuron", counter))
		os.WriteFile(counterFile, []byte(req.Insight), 0644)

		fmt.Printf("[COMMUNITY] 📡 %s/%s → counter %d\n", req.Source, safeTopic, counter)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "ok",
			"path":    fmt.Sprintf("cortex/community/%s/%s", req.Source, safeTopic),
			"counter": counter,
		})
	}))

	// GET /api/health — system process health check
	mux.HandleFunc("/api/health", withCORS(func(w http.ResponseWriter, r *http.Request) {
		health := buildHealthJSON(brainRoot)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(health)
	}))

	// GET /api/brain — full brain state for dashboard (compatible with dashboard.go format)
	mux.HandleFunc("/api/brain", withCORS(func(w http.ResponseWriter, r *http.Request) {
		data := buildBrainJSONResponse(brainRoot)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	}))



	// Start injection loop (inbox processing + auto-reinject) in background
	go runInjectionLoop(brainRoot)

	// Start idle engine in background
	go runIdleLoop(brainRoot)


	fmt.Printf("  🔄 IDLE ENGINE: auto evolve/snapshot/NAS every %dm idle\n", idleThresholdMinutes)
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
			"ack": "새로운 보고가 확인되었습니다. 사용자의 요청 처리 후 팔로업합니다.",
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

		// Check git availability
		if _, err := exec.LookPath("git"); err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{"error": "git not found", "events": []interface{}{}})
			return
		}

		// Get recent git log with file status (last 50 commits)
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
			Action    string `json:"action"`   // created, modified, suppressed
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

			// Commit header line: hash|datetime|message
			if parts := strings.SplitN(line, "|", 3); len(parts) == 3 && len(parts[0]) >= 40 {
				currentHash = parts[0][:8]
				currentTime = parts[1]
				currentMsg = parts[2]
				continue
			}

			// File status line: A/M/D\tpath
			if len(line) >= 2 && (line[0] == 'A' || line[0] == 'M' || line[0] == 'D') && line[1] == '\t' {
				filePath := line[2:]
				// Only track neuron-relevant files
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

				// Extract region from path
				region := ""
				pathParts := strings.SplitN(filePath, "/", 2)
				if len(pathParts) > 0 {
					if _, ok := regionPriority[pathParts[0]]; ok {
						region = pathParts[0]
					}
				}

				events = append(events, EvolutionEvent{
					Hash:      currentHash,
					Timestamp: currentTime,
					Message:   currentMsg,
					Action:    action,
					Path:      filePath,
					Region:    region,
				})
			}
		}

		// Also include unstaged changes as "live" brainwaves
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
					Hash:      "unstaged",
					Timestamp: time.Now().Format("2006-01-02 15:04:05 -0700"),
					Message:   "🧠 Active brainwave (uncommitted)",
					Action:    action,
					Path:      filePath,
					Region:    region,
				}}, events...)
			}
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"events": events,
			"total":  len(events),
		})
	}))

	fmt.Printf("  POST /api/report  {message,priority} — Stackable report queue\n")
	fmt.Printf("  GET  /api/reports                — List pending reports\n")
	fmt.Printf("  GET  /api/evolution              — Git-based neural evolution timeline\n")
	fmt.Printf("  GET  /api/retrieve               — Hebbian Retrieval & LLM Router (Phase 8)\n")
	mux.HandleFunc("/api/retrieve", handleRetrieve(brainRoot))

	// Code Map API — runtime file tree snapshot for dashboard
	fmt.Printf("  GET  /api/codemap                — Runtime file tree snapshot\n")
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

		result := map[string]interface{}{
			"generated":  time.Now().Format(time.RFC3339),
			"totalFiles": len(files),
			"totalLines": totalLines,
			"files":      files,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}))
	// Expose pprof
	mux.HandleFunc("/debug/pprof/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.DefaultServeMux.ServeHTTP(w, r)
	}))
	
	http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
}

// rollbackAll executes a global git rollback to restore the system state (P0 included).
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
