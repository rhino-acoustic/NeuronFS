package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// ─── Health check models ───

type ProcessHealth struct {
	Name    string `json:"name"`
	Role    string `json:"role"`
	Running bool   `json:"running"`
	PID     string `json:"pid,omitempty"`
}

type HealthJSON struct {
	API        bool            `json:"api"`
	Processes  []ProcessHealth `json:"processes"`
	OS         string          `json:"os"`
	BrainRoot  string          `json:"brainRoot"`
	NeuronFile int             `json:"neuronFiles"`
}

// ─── JSON models for API ───

type NeuronJSON struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	Counter   int    `json:"counter"`
	Contra    int    `json:"contra"`
	Dopamine  int    `json:"dopamine"`
	HasBomb   bool   `json:"hasBomb"`
	HasMemory bool   `json:"hasMemory"`
	IsDormant bool   `json:"isDormant"`
	Depth     int    `json:"depth"`
	ModTime   int64  `json:"modTime"`
}

type RegionJSON struct {
	Name     string       `json:"name"`
	Icon     string       `json:"icon"`
	Ko       string       `json:"ko"`
	Priority int          `json:"priority"`
	HasBomb  bool         `json:"hasBomb"`
	Neurons  []NeuronJSON `json:"neurons"`
	Axons    []string     `json:"axons"`
}

type BrainJSON struct {
	Root         string       `json:"root"`
	Regions      []RegionJSON `json:"regions"`
	BombSource   string       `json:"bombSource"`
	FiredNeurons int          `json:"firedNeurons"`
	TotalNeurons int          `json:"totalNeurons"`
	TotalCounter int          `json:"totalCounter"`
}

type AddNeuronReq struct {
	Region string `json:"region"`
	Path   string `json:"path"`
}

type AddBombReq struct {
	Region string `json:"region"`
	Name   string `json:"name"`
}

// ─── CORS middleware ───
func withCORSDashboard(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(200)
			return
		}
		h(w, r)
	}
}

// ─── Check if a process with given image name is running ───
func isProcessRunning(imageName string) bool {
	if runtime.GOOS != "windows" {
		out, err := exec.Command("pgrep", "-f", imageName).Output()
		return err == nil && len(out) > 0
	}
	out, err := exec.Command("tasklist", "/FI", "IMAGENAME eq "+imageName, "/NH", "/FO", "CSV").Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), imageName)
}

// ─── Check if a node.js script is running ───
func isNodeScriptRunning(scriptName string) bool {
	if runtime.GOOS != "windows" {
		out, err := exec.Command("pgrep", "-f", scriptName).Output()
		return err == nil && len(out) > 0
	}
	out, err := exec.Command("wmic", "process", "where", "name='node.exe'", "get", "commandline", "/format:list").Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), scriptName)
}

// ─── Count neuron files ───
func countNeuronFiles(brainRoot string) int {
	count := 0
	filepath.Walk(brainRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".neuron") {
			count++
		}
		return nil
	})
	return count
}

// ─── Build health JSON ───
func buildHealthJSON(brainRoot string) HealthJSON {
	processes := []ProcessHealth{
		// 인프라 데몬
		{Name: "neuronfs", Role: "인지 엔진 (API 서버 + 대시보드)", Running: isProcessRunning("neuronfs.exe")},
		{Name: "agent-bridge", Role: "CDP 에이전트 통신 브릿지", Running: isNodeScriptRunning("agent-bridge")},
		{Name: "auto-accept", Role: "CDP 자동 수락 + 교정 감지", Running: isNodeScriptRunning("auto-accept")},
		{Name: "watchdog", Role: "전체 프로세스 생존 감시 + harness 주기 실행", Running: isNodeScriptRunning("watchdog") || isProcessRunning("powershell.exe")},
	}

	return HealthJSON{
		API:        true, // 이 응답이 올 시점에 API는 살아있음
		Processes:  processes,
		OS:         runtime.GOOS,
		BrainRoot:  brainRoot,
		NeuronFile: countNeuronFiles(brainRoot),
	}
}

// ─── Build brain JSON from scan ───
func buildBrainJSONResponse(brainRoot string) BrainJSON {
	brain := scanBrain(brainRoot)
	result := runSubsumption(brain)

	data := BrainJSON{
		Root:         brain.Root,
		BombSource:   result.BombSource,
		FiredNeurons: result.FiredNeurons,
		TotalNeurons: result.TotalNeurons,
		TotalCounter: result.TotalCounter,
	}

	for _, region := range brain.Regions {
		rj := RegionJSON{
			Name:     region.Name,
			Icon:     regionIcons[region.Name],
			Ko:       regionKo[region.Name],
			Priority: region.Priority,
			HasBomb:  region.HasBomb,
			Axons:    region.Axons,
		}
		for _, n := range region.Neurons {
			rj.Neurons = append(rj.Neurons, NeuronJSON{
				Name:      n.Name,
				Path:      strings.ReplaceAll(n.Path, string(filepath.Separator), "/"),
				Counter:   n.Counter,
				Contra:    n.Contra,
				Dopamine:  n.Dopamine,
				HasBomb:   n.HasBomb,
				HasMemory: n.HasMemory,
				IsDormant: n.IsDormant,
				Depth:     n.Depth,
				ModTime:   n.ModTime.UnixMilli(),
			})
		}
		data.Regions = append(data.Regions, rj)
	}
	return data
}

// ─── Dashboard Server (--dashboard mode) ───

func startDashboard(brainRoot string, port int) {
	fmt.Printf("[NeuronFS] Dashboard: http://localhost:%d\n", port)
	fmt.Printf("[NeuronFS] Brain: %s\n", brainRoot)
	fmt.Printf("[NeuronFS] Axiom: Folder=Neuron | File=Trace | Path=Sentence\n")

	mux := http.NewServeMux()

	// GET / — Dashboard HTML (exact match "/" or fallback for non-API paths)
	mux.HandleFunc("/", withCORSDashboard(func(w http.ResponseWriter, r *http.Request) {
		// Serve dashboard HTML for root and all non-API routes (SPA support)
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		// Return empty 204 for favicon, manifest etc to suppress 404 in console
		if r.URL.Path == "/favicon.ico" || r.URL.Path == "/manifest.json" {
			w.WriteHeader(204)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, dashboardHTML)
	}))

	// GET /api/health — system process health check
	mux.HandleFunc("/api/health", withCORSDashboard(func(w http.ResponseWriter, r *http.Request) {
		health := buildHealthJSON(brainRoot)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(health)
	}))

	// GET /api/brain — scan and return full state
	mux.HandleFunc("/api/brain", withCORSDashboard(func(w http.ResponseWriter, r *http.Request) {
		data := buildBrainJSONResponse(brainRoot)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	}))

	// POST /api/inject — inject rules to GEMINI.md
	mux.HandleFunc("/api/inject", withCORSDashboard(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "POST only", 405)
			return
		}
		brain := scanBrain(brainRoot)
		result := runSubsumption(brain)
		rules := emitRules(result)
		injectToGemini(brainRoot, rules)
		w.Write([]byte(fmt.Sprintf("OK — %d neurons injected, activation: %d",
			result.FiredNeurons, result.TotalCounter)))
	}))

	// POST /api/neuron — add a new neuron (create folder + 1.neuron)
	mux.HandleFunc("/api/neuron", withCORSDashboard(func(w http.ResponseWriter, r *http.Request) {
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

		neuronDir := filepath.Join(brainRoot, req.Region, path)
		if err := os.MkdirAll(neuronDir, 0750); err != nil {
			http.Error(w, "mkdir failed: "+err.Error(), 500)
			return
		}

		counterFile := filepath.Join(neuronDir, "1.neuron")
		if err := os.WriteFile(counterFile, []byte(""), 0600); err != nil {
			http.Error(w, "write failed: "+err.Error(), 500)
			return
		}

		fmt.Printf("[GROWTH] New neuron: %s/%s\n", req.Region, path)
		w.Write([]byte("OK — " + req.Region + "/" + path))
	}))

	// POST /api/bomb — create bomb.neuron in a region
	mux.HandleFunc("/api/bomb", withCORSDashboard(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "POST only", 405)
			return
		}
		var req AddBombReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad json", 400)
			return
		}

		// Brainstem is immutable — no bomb allowed
		if req.Region == "brainstem" {
			http.Error(w, "brainstem neurons are immutable — bomb denied", 403)
			return
		}

		bombDir := filepath.Join(brainRoot, req.Region, req.Name)
		os.MkdirAll(bombDir, 0750)
		bombFile := filepath.Join(bombDir, "bomb.neuron")
		os.WriteFile(bombFile, []byte(""), 0600)

		fmt.Printf("[BOMB] 💀 %s/%s\n", req.Region, req.Name)
		w.Write([]byte("BOMB placed: " + req.Region + "/" + req.Name))
	}))

	// POST /api/increment — increment a neuron's counter
	mux.HandleFunc("/api/increment", withCORSDashboard(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "POST only", 405)
			return
		}
		var req AddNeuronReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad json", 400)
			return
		}

		neuronDir := filepath.Join(brainRoot, req.Region, req.Path)
		if _, err := os.Stat(neuronDir); os.IsNotExist(err) {
			http.Error(w, "neuron not found", 404)
			return
		}

		files, _ := filepath.Glob(filepath.Join(neuronDir, "*.neuron"))
		currentCounter := 0
		var counterFilePath string
		for _, f := range files {
			fname := filepath.Base(f)
			if m := counterRegex.FindStringSubmatch(fname); m != nil {
				n := 0
				fmt.Sscanf(m[1], "%d", &n)
				if n > currentCounter {
					currentCounter = n
					counterFilePath = f
				}
			}
		}

		if counterFilePath != "" {
			os.Remove(counterFilePath)
		}
		newCounter := currentCounter + 1
		newFile := filepath.Join(neuronDir, fmt.Sprintf("%d.neuron", newCounter))
		os.WriteFile(newFile, []byte(""), 0600)

		fmt.Printf("[FIRE] %s/%s: %d → %d\n", req.Region, req.Path, currentCounter, newCounter)
		w.Write([]byte(fmt.Sprintf("%d", newCounter)))
	}))

	// POST /api/fire — fire (increment) a neuron by path
	mux.HandleFunc("/api/fire", withCORSDashboard(func(w http.ResponseWriter, r *http.Request) {
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
		fireNeuron(brainRoot, path)
		w.Write([]byte("OK — fired: " + path))
	}))

	// POST /api/dedup — deduplicate similar neurons
	mux.HandleFunc("/api/dedup", withCORSDashboard(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "POST only", 405)
			return
		}
		deduplicateNeurons(brainRoot)
		w.Write([]byte("OK — dedup complete"))
	}))

	// POST /api/signal — add signal (dopamine/bomb/memory) to a neuron
	mux.HandleFunc("/api/signal", withCORSDashboard(func(w http.ResponseWriter, r *http.Request) {
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
		if err := signalNeuron(brainRoot, path, sigType); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		w.Write([]byte("OK — " + sigType + ": " + path))
	}))

	// POST /api/rollback — rollback (decrement) a neuron's counter
	mux.HandleFunc("/api/rollback", withCORSDashboard(func(w http.ResponseWriter, r *http.Request) {
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
		if err := rollbackNeuron(brainRoot, path); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		w.Write([]byte("OK — rolled back: " + path))
	}))

	// POST /api/contra — add inhibitory signal to a neuron
	mux.HandleFunc("/api/contra", withCORSDashboard(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "403 Forbidden: API write operations (.contra) are disabled for security. Use CLI.", 403)
	}))
	// POST /api/report — stackable report queue
	mux.HandleFunc("/api/report", withCORSDashboard(func(w http.ResponseWriter, r *http.Request) {
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
			"status":   "confirmed",
			"message":  req.Message,
			"priority": req.Priority,
			"pending":  pending,
			"ack":      "새로운 보고가 확인되었습니다. 사용자의 요청 처리 후 팔로업합니다.",
		})
	}))

	// GET /api/reports — list pending reports
	mux.HandleFunc("/api/reports", withCORSDashboard(func(w http.ResponseWriter, r *http.Request) {
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

	http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", port), mux)
}
