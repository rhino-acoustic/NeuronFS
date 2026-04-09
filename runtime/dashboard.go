package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
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


