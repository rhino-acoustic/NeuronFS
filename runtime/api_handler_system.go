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
	"sync"
	"time"
)

type LogMessage struct {
	Level   string `json:"level"`
	Message string `json:"message"`
}

type EventMessage map[string]interface{}

type SSEBroker struct {
	mu      sync.Mutex
	clients map[chan EventMessage]bool
}

func (b *SSEBroker) AddClient(c chan EventMessage) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.clients[c] = true
}

func (b *SSEBroker) RemoveClient(c chan EventMessage) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.clients, c)
}

func (b *SSEBroker) Broadcast(level, msg string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	logObj := EventMessage{"type": "log", "level": level, "message": msg}
	for c := range b.clients {
		select {
		case c <- logObj:
		default:
		}
	}
}

func (b *SSEBroker) BroadcastEvent(data EventMessage) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for c := range b.clients {
		select {
		case c <- data:
		default:
		}
	}
}

func (b *SSEBroker) Broadcastf(level, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	b.Broadcast(level, msg)
}

var GlobalSSEBroker = &SSEBroker{clients: make(map[chan EventMessage]bool)}

func registerSystemRoutes(mux *http.ServeMux, brainRoot string, withCORS func(http.HandlerFunc) http.HandlerFunc) {
	// GET /api/integrity?region=cortex
	mux.HandleFunc("/api/integrity", withCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "GET only", 405)
			return
		}

		hmacKey := loadOrCreateHMACKey(brainRoot)
		targetRegion := r.URL.Query().Get("region")

		regions := RegionOrder
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
		os.MkdirAll(reportsDir, 0750)
		ts := fmt.Sprintf("%d", time.Now().UnixMilli())
		filename := fmt.Sprintf("%s_%s.report", ts, req.Priority)
		content := fmt.Sprintf("priority: %s\ntimestamp: %s\n\n%s\n", req.Priority, time.Now().Format("2006-01-02 15:04:05"), req.Message)
		os.WriteFile(filepath.Join(reportsDir, filename), []byte(content), 0600)

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

		out, err := SafeOutputDir(ExecTimeoutGit, brainRoot, "git", "log", "--pretty=format:%H|%ai|%s", "--name-status", "-n", "50")
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
		diffOut, err := SafeOutputDir(ExecTimeoutQuery, brainRoot, "git", "status", "--porcelain")
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
					Action:  action, Path: filePath, Region: region,
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

	// GET /api/stream — SSE Telemetry Stream
	mux.HandleFunc("/api/stream", withCORS(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
			return
		}

		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		clientChan := make(chan EventMessage, 100)
		GlobalSSEBroker.AddClient(clientChan)
		defer GlobalSSEBroker.RemoveClient(clientChan)

		for {
			select {
			case <-r.Context().Done():
				return
			case eventData := <-clientChan:
				jsonBytes, _ := json.Marshal(eventData)
				fmt.Fprintf(w, "data: %s\n\n", string(jsonBytes))
				flusher.Flush()
			case <-ticker.C:
				brain := scanBrain(brainRoot)
				totalNeurons := 0
				totalActivation := 0
				for _, r := range brain.Regions {
					totalNeurons += len(r.Neurons)
					for _, n := range r.Neurons {
						totalActivation += n.Counter
					}
				}

				data := map[string]interface{}{
					"type":            "stat",
					"ts":              time.Now().Format(time.RFC3339),
					"totalNeurons":    totalNeurons,
					"totalActivation": totalActivation,
				}
				jsonBytes, _ := json.Marshal(data)
				fmt.Fprintf(w, "data: %s\n\n", string(jsonBytes))
				flusher.Flush()
			}
		}
	}))

	// GET /api/ops — 통합 운영 상태 (watchdog metrics + brain state)
	mux.HandleFunc("/api/ops", withCORS(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// watchdog 메트릭 읽기
		metricsPath := filepath.Join(filepath.Dir(brainRoot), "logs", "watchdog_metrics.json")
		var watchdogData interface{}
		if data, err := os.ReadFile(metricsPath); err == nil {
			json.Unmarshal(data, &watchdogData)
		}

		// brain 상태
		brain := scanBrain(brainRoot)
		totalNeurons := 0
		totalActivation := 0
		regionSummary := make([]map[string]interface{}, 0)
		for _, r := range brain.Regions {
			totalNeurons += len(r.Neurons)
			regionAct := 0
			for _, n := range r.Neurons {
				totalActivation += n.Counter
				regionAct += n.Counter
			}
			regionSummary = append(regionSummary, map[string]interface{}{
				"name":       r.Name,
				"neurons":    len(r.Neurons),
				"activation": regionAct,
			})
		}

		// 로그 파일 크기
		logsDir := filepath.Join(filepath.Dir(brainRoot), "logs")
		logEntries, _ := os.ReadDir(logsDir)
		logFiles := make([]map[string]interface{}, 0)
		for _, e := range logEntries {
			if e.IsDir() {
				continue
			}
			info, err := e.Info()
			if err != nil {
				continue
			}
			logFiles = append(logFiles, map[string]interface{}{
				"name":     e.Name(),
				"size":     info.Size(),
				"modified": info.ModTime().Format(time.RFC3339),
			})
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"ts":       time.Now().Format(time.RFC3339),
			"watchdog": watchdogData,
			"brain": map[string]interface{}{
				"totalNeurons":    totalNeurons,
				"totalActivation": totalActivation,
				"regions":         regionSummary,
			},
			"logs": logFiles,
		})
	}))

	// GET /ops — 운영 대시보드 HTML
	mux.HandleFunc("/ops", withCORS(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(opsDashboardHTML))
	}))

	// GET/POST /api/autopilot — 자율주행 모드 온오프
	mux.HandleFunc("/api/autopilot", withCORS(func(w http.ResponseWriter, r *http.Request) {
		nfsRoot := filepath.Dir(brainRoot)
		flagFile := filepath.Join(nfsRoot, "telegram-bridge", ".auto_evolve_disabled")
		w.Header().Set("Content-Type", "application/json")

		if r.Method == "POST" {
			var req struct {
				Enabled bool `json:"enabled"`
			}
			json.NewDecoder(r.Body).Decode(&req)
			if req.Enabled {
				os.Remove(flagFile)
			} else {
				os.WriteFile(flagFile, []byte("1"), 0600)
			}
		}

		enabled := !fileExists(flagFile)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"autopilot": enabled,
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
