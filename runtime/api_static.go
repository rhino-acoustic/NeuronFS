// api_static.go — 대시보드/정적 파일 서빙
//
// PROVIDES: registerStaticRoutes
// 하드코딩 경로 제거: neuronfsRoot로 통일

package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func registerStaticRoutes(mux *http.ServeMux, brainRoot string, withCORS func(http.HandlerFunc) http.HandlerFunc) {
	neuronfsRoot := filepath.Dir(brainRoot)

	// GET /3d — Dashboard HTML
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

	// GET /brain.obj
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

	// GET /brain_state.json
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

	// GET /v2 — Dashboard V2
	mux.HandleFunc("/v2", withCORS(func(w http.ResponseWriter, r *http.Request) {
		v2Path := filepath.Join(neuronfsRoot, "dashboard_v2.html")
		data, err := os.ReadFile(v2Path)
		if err != nil {
			http.Error(w, "dashboard_v2.html not found", 404)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(data)
	}))

	// GET /api/dashboard.svg — Generate dashboard vector snapshot
	mux.HandleFunc("/api/dashboard.svg", withCORS(func(w http.ResponseWriter, r *http.Request) {
		brain := scanBrain(brainRoot)
		res := runSubsumption(brain)
		svgContent := GenerateDashboardSVG(brain, res.TotalNeurons, res.TotalCounter)
		w.Header().Set("Content-Type", "image/svg+xml")
		w.Write([]byte(svgContent))
	}))

	// GET / — Main Dashboard
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
			http.Error(w, "brain_dashboard.html not found. Please ensure the UI is placed in the NeuronFS root.", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(data)
	}))

	// GET /bible — Bible demo (NeuronFS showcase with real filesystem data)
	mux.HandleFunc("/bible", withCORS(func(w http.ResponseWriter, r *http.Request) {
		// Look for index.html in examples/game_world relative to the runtime binary,
		// or relative to the neuronfs dist root
		candidates := []string{
			filepath.Join(neuronfsRoot, "..", "examples", "game_world", "index.html"),
			filepath.Join(neuronfsRoot, "..", "..", "examples", "game_world", "index.html"),
		}
		for _, p := range candidates {
			data, err := os.ReadFile(p)
			if err == nil {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.Write(data)
				return
			}
		}
		http.Error(w, "examples/game_world/index.html not found", 404)
	}))

	fmt.Printf("  POST /api/report  {message,priority} — Stackable report queue\n")
	fmt.Printf("  GET  /api/reports                — List pending reports\n")
	fmt.Printf("  GET  /api/evolution              — Git-based neural evolution timeline\n")
	fmt.Printf("  GET  /api/retrieve               — Hebbian Retrieval & LLM Router (Phase 8)\n")
	fmt.Printf("  GET  /api/codemap                — Runtime file tree snapshot\n")
}
