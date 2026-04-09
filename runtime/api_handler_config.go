package main

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

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

		os.MkdirAll(filepath.Dir(principlesFile), 0750)
		os.WriteFile(principlesFile, []byte(strings.Join(clean, "\n")), 0600)

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

		os.MkdirAll(filepath.Dir(stateFile), 0750)
		stateBytes, _ := json.MarshalIndent(state, "", "  ")
		os.WriteFile(stateFile, stateBytes, 0600)

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

		os.MkdirAll(filepath.Dir(sandboxFile), 0750)
		os.WriteFile(sandboxFile, []byte(text), 0600)

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
			os.MkdirAll(neuronDir, 0750)
			os.WriteFile(filepath.Join(neuronDir, "1.neuron"), []byte{}, 0600)
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
