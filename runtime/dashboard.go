// PROVIDES: buildDashboard, renderSystemGraph
// DEPENDS ON: brain.go (scanBrain), api_server.go (health)
package main

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// portAlive — TCP 포트 연결 가능 여부 (500ms 타임아웃)
func portAlive(port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 500*time.Millisecond)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// ─── Health check models ───

type ProcessHealth struct {
	Name    string `json:"name"`
	Role    string `json:"role"`
	Running bool   `json:"running"`
	PID     string `json:"pid,omitempty"`
}

type SubsystemStatus struct {
	Status   string `json:"status"`
	Detail   string `json:"detail,omitempty"`
}

type BacklogStatus struct {
	Pending  int `json:"pending"`
	Backlog  int `json:"backlog"`
	Archive  int `json:"archive"`
}

type HealthJSON struct {
	API        bool                      `json:"api"`
	Processes  []ProcessHealth           `json:"processes"`
	Subsystems map[string]SubsystemStatus `json:"subsystems"`
	Backlog    BacklogStatus             `json:"backlog"`
	OS         string                    `json:"os"`
	BrainRoot  string                    `json:"brainRoot"`
	NeuronFile int                       `json:"neuronFiles"`
	Uptime     string                    `json:"uptime"`
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

type CartridgeJSON struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
}

type BrainJSON struct {
	Root         string          `json:"root"`
	Regions      []RegionJSON    `json:"regions"`
	Cartridges   []CartridgeJSON `json:"cartridges"`
	BombSource   string          `json:"bombSource"`
	FiredNeurons int             `json:"firedNeurons"`
	TotalNeurons int             `json:"totalNeurons"`
	TotalCounter int             `json:"totalCounter"`
}

type AddNeuronReq struct {
	Region string `json:"region"`
	Path   string `json:"path"`
}

type AddBombReq struct {
	Region string `json:"region"`
	Name   string `json:"name"`
}

// ─── Check if a process with given image name is running ───
func isProcessRunning(imageName string) bool {
	if runtime.GOOS != "windows" {
		out, err := SafeOutput(ExecTimeoutShell, "pgrep", "-f", imageName)
		return err == nil && len(out) > 0
	}
	out, err := SafeOutput(ExecTimeoutShell, "tasklist", "/FI", "IMAGENAME eq "+imageName, "/NH", "/FO", "CSV")
	if err != nil {
		return false
	}
	return strings.Contains(string(out), imageName)
}

// ─── Check if a node.js script is running ───
func isNodeScriptRunning(scriptName string) bool {
	if runtime.GOOS != "windows" {
		out, err := SafeOutput(ExecTimeoutShell, "pgrep", "-f", scriptName)
		return err == nil && len(out) > 0
	}
	out, err := SafeOutput(ExecTimeoutShell, "wmic", "process", "where", "name='node.exe'", "get", "commandline", "/format:list")
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
// Go 네이티브: supervisor goroutine 상태 추적
var goNativeServices sync.Map // name → bool (running)

func markServiceRunning(name string, running bool) {
	goNativeServices.Store(name, running)
}

func isGoServiceRunning(name string) bool {
	if v, ok := goNativeServices.Load(name); ok {
		return v.(bool)
	}
	return false
}

func buildHealthJSON(brainRoot string) HealthJSON {
	subs := make(map[string]SubsystemStatus)

	// MCP check (9247)
	mcpOK := portAlive(MCPStreamPort)
	if mcpOK {
		subs["mcp"] = SubsystemStatus{Status: "alive", Detail: fmt.Sprintf("port %d", MCPStreamPort)}
	} else {
		subs["mcp"] = SubsystemStatus{Status: "dead", Detail: fmt.Sprintf("port %d unreachable", MCPStreamPort)}
	}

	// CDP check (9000)
	cdpOK := portAlive(hlCDPPort)
	if cdpOK {
		subs["cdp"] = SubsystemStatus{Status: "alive", Detail: fmt.Sprintf("port %d", hlCDPPort)}
	} else {
		subs["cdp"] = SubsystemStatus{Status: "dead", Detail: fmt.Sprintf("port %d unreachable", hlCDPPort)}
	}

	// Telegram
	if hlTgToken != "" {
		subs["telegram"] = SubsystemStatus{Status: "alive", Detail: fmt.Sprintf("offset=%d", hlTgOffset)}
	} else {
		subs["telegram"] = SubsystemStatus{Status: "no_token"}
	}

	// Go services
	for _, svc := range []string{"watch", "api", "supervisor"} {
		if isGoServiceRunning(svc) {
			subs[svc] = SubsystemStatus{Status: "alive"}
		} else {
			subs[svc] = SubsystemStatus{Status: "unknown"}
		}
	}

	// Backlog counts
	inboxDir := filepath.Join(brainRoot, "_agents", "NeuronFS", "inbox")
	var pending, backlogCount, archiveCount int
	if entries, err := os.ReadDir(inboxDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() {
				pending++
			}
		}
	}
	if entries, err := os.ReadDir(filepath.Join(inboxDir, "_backlog")); err == nil {
		backlogCount = len(entries)
	}
	if entries, err := os.ReadDir(filepath.Join(inboxDir, "_archive")); err == nil {
		archiveCount = len(entries)
	}

	return HealthJSON{
		API:        true,
		Processes:  []ProcessHealth{},
		Subsystems: subs,
		Backlog:    BacklogStatus{Pending: pending, Backlog: backlogCount, Archive: archiveCount},
		OS:         runtime.GOOS,
		BrainRoot:  brainRoot,
		NeuronFile: countNeuronFiles(brainRoot),
		Uptime:     time.Since(startTime).Round(time.Second).String(),
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

	// Scan Jloot Cartridges (flat scan only — avoid walking 7K+ files)
	searchPaths := []string{
		filepath.Join(brainRoot, ".."),
		filepath.Join(brainRoot, "..", "tools", "jloot"),
	}
	for _, p := range searchPaths {
		entries, err := os.ReadDir(p)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".jloot") {
				info, _ := e.Info()
				if info != nil {
					data.Cartridges = append(data.Cartridges, CartridgeJSON{Name: info.Name(), Size: info.Size()})
				}
			}
		}
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

// ──────────────────────────────────────────────────────────
// T5: 시스템 플로우차트 API (적층)
// ──────────────────────────────────────────────────────────

// SystemFlowNode represents a node in the system flow diagram
type SystemFlowNode struct {
	ID       string   `json:"id"`
	Label    string   `json:"label"`
	Type     string   `json:"type"`     // "process", "data", "decision", "cron"
	Status   string   `json:"status"`   // "active", "idle", "error"
	Children []string `json:"children"` // connected node IDs
}

// buildSystemFlow returns the complete system architecture as a flow diagram
func buildSystemFlow(brainRoot string) []SystemFlowNode {
	return []SystemFlowNode{
		// Core processes
		{ID: "supervisor", Label: "Supervisor", Type: "process", Status: boolStatus(true), Children: []string{"api", "mcp", "cdp", "autopilot", "cron"}},
		{ID: "api", Label: "REST API (:9090)", Type: "process", Status: boolStatus(portAlive(9090)), Children: []string{"dashboard", "brain"}},
		{ID: "mcp", Label: "MCP Server (:9247)", Type: "process", Status: boolStatus(portAlive(9247)), Children: []string{"brain"}},
		{ID: "cdp", Label: "CDP (:9000)", Type: "process", Status: boolStatus(portAlive(9000)), Children: []string{"auto_accept", "hijack"}},

		// Background systems
		{ID: "autopilot", Label: "Autopilot (hlAutoEvolve)", Type: "process", Status: "active", Children: []string{"gemini_cli", "evolve"}},
		{ID: "gemini_cli", Label: "Gemini CLI", Type: "process", Status: "idle", Children: []string{"brain"}},
		{ID: "auto_accept", Label: "Auto Accept (A3)", Type: "process", Status: "active", Children: []string{"cdp"}},
		{ID: "hijack", Label: "Telegram→IDE (A6)", Type: "process", Status: "active", Children: []string{"cdp"}},

		// Cron cycle
		{ID: "cron", Label: "Cron (매시간)", Type: "cron", Status: "active", Children: []string{"git_snap", "archive", "prune", "categorize", "verify", "dedup"}},
		{ID: "git_snap", Label: "Step1: Git Snapshot", Type: "cron", Status: "active", Children: []string{}},
		{ID: "archive", Label: "Step2: 전사 백업", Type: "cron", Status: "active", Children: []string{"transcripts"}},
		{ID: "prune", Label: "Step2.5: 50건 정리", Type: "cron", Status: "active", Children: []string{"transcripts"}},
		{ID: "categorize", Label: "Step4: 카테고리분류", Type: "cron", Status: "active", Children: []string{"gemini_cli", "analysis"}},
		{ID: "verify", Label: "Step6: 외부검증", Type: "cron", Status: "active", Children: []string{}},
		{ID: "dedup", Label: "Step7: Dedup (6h)", Type: "cron", Status: "active", Children: []string{"brain"}},

		// Data stores
		{ID: "brain", Label: "brain_v4 (SSOT)", Type: "data", Status: "active", Children: []string{}},
		{ID: "transcripts", Label: "_transcripts", Type: "data", Status: "active", Children: []string{}},
		{ID: "analysis", Label: "hippocampus/전사분석", Type: "data", Status: "active", Children: []string{}},
		{ID: "dashboard", Label: "V3 Dashboard", Type: "process", Status: boolStatus(portAlive(9090)), Children: []string{}},

		// Emit targets
		{ID: "evolve", Label: "--emit auto", Type: "process", Status: "active", Children: []string{"gemini_md", "cursorrules", "claude_md", "ki"}},
		{ID: "gemini_md", Label: "GEMINI.md", Type: "data", Status: "active", Children: []string{}},
		{ID: "cursorrules", Label: ".cursorrules", Type: "data", Status: "active", Children: []string{}},
		{ID: "claude_md", Label: "CLAUDE.md", Type: "data", Status: "active", Children: []string{}},
		{ID: "ki", Label: "KI (Antigravity)", Type: "data", Status: "active", Children: []string{}},
	}
}

func boolStatus(alive bool) string {
	if alive {
		return "active"
	}
	return "error"
}
