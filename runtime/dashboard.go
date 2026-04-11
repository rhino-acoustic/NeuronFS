package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
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
	processes := []ProcessHealth{
		// Go 네이티브 서비스 (supervisor goroutine 기반)
		{Name: "neuronfs-api", Role: "인지 엔진 (REST API + 대시보드)", Running: isProcessRunning("neuronfs.exe")},
		{Name: "auto-accept", Role: "CDP 자동 수락 + NEURON 감지 + Groq 배치", Running: isGoServiceRunning("auto-accept")},
		{Name: "agent-bridge", Role: "에이전트 라우팅 브릿지", Running: isGoServiceRunning("agent-bridge")},
		{Name: "context-hijacker", Role: "MITM 컨텍스트 캡처", Running: isGoServiceRunning("context-hijacker")},
		{Name: "headless-executor", Role: "Inbox 명령 샌드박스 실행", Running: isGoServiceRunning("headless-executor")},
		{Name: "hijack-launcher", Role: "TG 양방향 + CDP 전사 + 자동통합", Running: isGoServiceRunning("hijack-launcher")},
		{Name: "supervisor", Role: "전체 프로세스 관리 + 자기 감시", Running: isProcessRunning("neuronfs.exe")},
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

	// Scan Jloot Cartridges
	searchPaths := []string{
		filepath.Join(brainRoot, "..", "runtime"),
		filepath.Join(brainRoot, "..", "tools", "jloot"),
	}
	for _, p := range searchPaths {
		filepath.Walk(p, func(path string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() && strings.HasSuffix(info.Name(), ".jloot") {
				data.Cartridges = append(data.Cartridges, CartridgeJSON{Name: info.Name(), Size: info.Size()})
			}
			return nil
		})
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
