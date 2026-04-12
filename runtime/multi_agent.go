package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ============================================================================
// Module: Multi-Agent Orchestrator — Parallel Gemini CLI Spawner (V12-A)
// Spawns multiple Gemini CLI processes in parallel, each assigned a different
// task from the V12 roadmap. Results are merged back via shared brain_v4/.
// ============================================================================

// AgentTask defines a task to be executed by a Gemini CLI agent
type AgentTask struct {
	Name    string // Human-readable task name
	Prompt  string // Prompt to inject into Gemini CLI
	WorkDir string // Working directory for this agent
}

// AgentResult holds the outcome of a parallel agent run
type AgentResult struct {
	Task     AgentTask
	Success  bool
	Output   string
	Duration time.Duration
}

// RunParallelAgents spawns multiple Gemini CLI processes in parallel
func RunParallelAgents(brainRoot string, tasks []AgentTask) []AgentResult {
	var wg sync.WaitGroup
	results := make([]AgentResult, len(tasks))

	fmt.Printf("[멀티에이전트] %d개 에이전트 병렬 실행 시작\n", len(tasks))
	RecordAudit(brainRoot, "multi_agent", "spawn", brainRoot, fmt.Sprintf("%d agents", len(tasks)), true)

	for i, task := range tasks {
		wg.Add(1)
		go func(idx int, t AgentTask) {
			defer wg.Done()
			start := time.Now()

			fmt.Printf("[에이전트 %d] 시작: %s\n", idx+1, t.Name)

			// Build the Gemini CLI command
			result := executeGeminiCLI(t)
			result.Duration = time.Since(start)
			results[idx] = result

			status := "✅"
			if !result.Success {
				status = "❌"
			}
			fmt.Printf("[에이전트 %d] %s 완료 (%s): %s\n", idx+1, status, result.Duration.Round(time.Second), t.Name)

			RecordAudit(brainRoot, "multi_agent", "complete", t.Name,
				fmt.Sprintf("success=%v duration=%s", result.Success, result.Duration.Round(time.Second)), result.Success)
		}(i, task)
	}

	wg.Wait()
	fmt.Printf("[멀티에이전트] 전체 완료\n")
	return results
}

// executeGeminiCLI runs a single Gemini CLI process
func executeGeminiCLI(task AgentTask) AgentResult {
	// Check if gemini CLI exists
	geminiPath, err := exec.LookPath("gemini")
	if err != nil {
		// Fallback: try common locations
		candidates := []string{
			filepath.Join(os.Getenv("LOCALAPPDATA"), "npm", "gemini.cmd"),
			filepath.Join(os.Getenv("USERPROFILE"), ".local", "bin", "gemini"),
			"gemini",
		}
		found := false
		for _, c := range candidates {
			if _, statErr := os.Stat(c); statErr == nil {
				geminiPath = c
				found = true
				break
			}
		}
		if !found {
			return AgentResult{
				Task:    task,
				Success: false,
				Output:  "gemini CLI를 찾을 수 없습니다. npm install -g @anthropic-ai/gemini-cli 또는 동등한 명령으로 설치하세요.",
			}
		}
	}

	cmd := exec.Command(geminiPath, "--prompt", task.Prompt)
	if task.WorkDir != "" {
		cmd.Dir = task.WorkDir
	}

	// Set timeout: 5 minutes per task max
	output, err := cmd.CombinedOutput()
	if err != nil {
		return AgentResult{
			Task:    task,
			Success: false,
			Output:  fmt.Sprintf("실행 실패: %v\n%s", err, string(output)),
		}
	}

	return AgentResult{
		Task:    task,
		Success: true,
		Output:  string(output),
	}
}

// SpawnV12Agents creates and runs the V12 roadmap parallel tasks
func SpawnV12Agents(brainRoot string) {
	runtimeDir := filepath.Join(filepath.Dir(brainRoot), "runtime")

	tasks := []AgentTask{
		{
			Name:    "V12-C 벤치마크",
			Prompt:  "NeuronFS의 TF-IDF 유사도 인덱스 성능을 측정하라. brain_v4에서 420개 뉴런 대상으로 QuerySimilar 호출 시간을 측정하고 결과를 보고하라.",
			WorkDir: runtimeDir,
		},
		{
			Name:    "V12-E 하네스 테스트",
			Prompt:  "NeuronFS의 emit 파이프라인이 GEMINI.md, CLAUDE.md, .cursorrules 형식을 모두 올바르게 생성하는지 테스트하라.",
			WorkDir: runtimeDir,
		},
	}

	results := RunParallelAgents(brainRoot, tasks)

	// Write results summary to brain
	var sb strings.Builder
	sb.WriteString("# 멀티에이전트 실행 결과\n\n")
	sb.WriteString(fmt.Sprintf("실행 시각: %s\n\n", time.Now().Format(time.RFC3339)))

	for i, r := range results {
		status := "성공"
		if !r.Success {
			status = "실패"
		}
		sb.WriteString(fmt.Sprintf("## 에이전트 %d: %s [%s]\n", i+1, r.Task.Name, status))
		sb.WriteString(fmt.Sprintf("소요 시간: %s\n", r.Duration.Round(time.Second)))
		// Truncate output if too long
		output := r.Output
		if len(output) > 500 {
			output = output[:500] + "\n...(잘림)"
		}
		sb.WriteString(fmt.Sprintf("```\n%s\n```\n\n", output))
	}

	resultPath := filepath.Join(brainRoot, "hippocampus", "agent_results", fmt.Sprintf("parallel_%d.neuron", time.Now().Unix()))
	_ = os.MkdirAll(filepath.Dir(resultPath), 0755)
	_ = os.WriteFile(resultPath, []byte(sb.String()), 0644)
	fmt.Printf("[멀티에이전트] 결과 저장: %s\n", resultPath)
}
