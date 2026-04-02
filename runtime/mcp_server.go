// NeuronFS MCP Server — Go Native
//
// Node.js 래퍼(mcp/index.js)를 대체하여 Go에서 직접 MCP 프로토콜을 서빙한다.
// AI(Gemini/Claude) ↔ Go(stdio, JSON-RPC 2.0) ↔ Filesystem
//
// 도구 7개:
//   read_region  — 영역의 최신 _rules.md 반환 (읽기 = 발화)
//   read_brain   — 전체 뇌 상태 JSON
//   grow         — 뉴런 생성
//   fire         — 뉴런 발화 (카운터 증가)
//   signal       — 도파민/bomb/memory 신호
//   correct      — PD 교정 기록
//   evolve       — Groq 기반 자율 뇌 진화

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// startMCPServer bootstraps the MCP stdio server using os.Stdin/os.Stdout.
// WARNING: Only use this when os.Stdout is clean (not redirected).
func startMCPServer(brainRoot string) {
	startMCPServerWithStdout(brainRoot, os.Stdout)
}

// startMCPServerWithStdout bootstraps the MCP stdio server with a specific stdout writer.
// This is used in --mcp mode where os.Stdout is redirected to stderr to prevent
// fmt.Print* from polluting the JSON-RPC channel.
func startMCPServerWithStdout(brainRoot string, stdout *os.File) {
	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    "neuronfs",
			Version: "1.0.0",
		},
		nil,
	)

	registerMCPTools(server, brainRoot)

	fmt.Fprintf(os.Stderr, "\033[36m[NEURON] Core Initialization Complete.\033[0m\n")
	fmt.Fprintf(os.Stderr, "\033[35m[SYNAPSE] Listening on stdio via Native MCP. Zero dependencies.\033[0m\n")
	fmt.Fprintf(os.Stderr, "\033[37m  - Waiting for context pulses...\033[0m\n")
	ctx := context.Background()
	// Use IOTransport instead of StdioTransport to avoid using os.Stdout
	// (which has been redirected to stderr in --mcp mode)
	transport := &mcp.IOTransport{
		Reader: os.Stdin,
		Writer: stdout,
	}
	if _, err := server.Connect(ctx, transport, nil); err != nil {
		log.Fatalf("[MCP] FATAL: %v\n", err)
	}

	// Block forever — stdio connection runs until process dies
	select {}
}

// registerMCPTools registers all NeuronFS commands as tools in the MCP server.
func registerMCPTools(server *mcp.Server, brainRoot string) {

	// ─── Tool 1: read_region ───
	server.AddTool(
		&mcp.Tool{
			Name:        "read_region",
			Description: "영역의 최신 _rules.md를 실시간 생성하여 반환. 작업 전환 시 해당 영역을 읽으면 최신 뉴런 상태를 얻는다. 읽기 = 발화: 상위 3개 뉴런이 자동으로 활성화된다.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"region": {
						"type": "string",
						"enum": ["brainstem","limbic","hippocampus","sensors","cortex","ego","prefrontal"],
						"description": "읽을 영역. CSS/디자인→cortex, NAS→sensors, 브랜드→sensors, 프로젝트→prefrontal"
					}
				},
				"required": ["region"]
			}`),
		},
		func(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args struct {
				Region string `json:"region"`
			}
			if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
				return mcpError("invalid arguments: " + err.Error()), nil
			}

			regionName := args.Region
			if _, ok := regionPriority[regionName]; !ok {
				return mcpError("invalid region: " + regionName), nil
			}

			brain := scanBrain(brainRoot)
			var content string
			for _, region := range brain.Regions {
				if region.Name == regionName {
					content = emitRegionRules(region)
					// Write to file for view_file access
					rulesPath := filepath.Join(brainRoot, regionName, "_rules.md")
					os.WriteFile(rulesPath, []byte(content), 0644)

					// FIRE: reading = activation (top 3)
					topN := sortedActiveNeurons(region.Neurons, 3)
					for _, n := range topN {
						relPath, _ := filepath.Rel(brainRoot, n.FullPath)
						fireNeuron(brainRoot, relPath)
					}
					break
				}
			}

			if content == "" {
				return mcpError("region not found"), nil
			}

			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: content}},
			}, nil
		},
	)

	// ─── Tool 2: read_brain ───
	server.AddTool(
		&mcp.Tool{
			Name:        "read_brain",
			Description: "전체 뇌 상태를 JSON으로 반환. 영역별 뉴런 수, 활성도, axon 연결, bomb 상태 등을 포함한다.",
			InputSchema: json.RawMessage(`{"type": "object", "properties": {}}`),
		},
		func(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			data := buildBrainJSONResponse(brainRoot)
			jsonBytes, err := json.MarshalIndent(data, "", "  ")
			if err != nil {
				return mcpError("json marshal error: " + err.Error()), nil
			}
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: string(jsonBytes)}},
			}, nil
		},
	)

	// ─── Tool 3: grow ───
	server.AddTool(
		&mcp.Tool{
			Name:        "grow",
			Description: "새 뉴런을 생성한다. 경로는 region/카테고리/이름 형식. 이미 존재하면 스킵, 60% 이상 유사한 뉴런이 있으면 기존 뉴런을 발화한다.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"path": {
						"type": "string",
						"description": "뉴런 경로. 예: cortex/frontend/css/새_규칙, brainstem/禁새_금지사항"
					}
				},
				"required": ["path"]
			}`),
		},
		func(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args struct {
				Path string `json:"path"`
			}
			if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
				return mcpError("invalid arguments"), nil
			}
			if args.Path == "" {
				return mcpError("path required"), nil
			}

			if err := growNeuron(brainRoot, args.Path); err != nil {
				return mcpError(err.Error()), nil
			}
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("🌱 grown: %s", args.Path)}},
			}, nil
		},
	)

	// ─── Tool 4: fire ───
	server.AddTool(
		&mcp.Tool{
			Name:        "fire",
			Description: "기존 뉴런의 카운터를 1 증가시킨다. 뉴런이 없으면 자동으로 생성 후 발화한다.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"path": {
						"type": "string",
						"description": "뉴런 경로. 예: cortex/testing/검증_E2E"
					}
				},
				"required": ["path"]
			}`),
		},
		func(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args struct {
				Path string `json:"path"`
			}
			if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
				return mcpError("invalid arguments"), nil
			}
			if args.Path == "" {
				return mcpError("path required"), nil
			}

			fireNeuron(brainRoot, args.Path)
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("🔥 fired: %s", args.Path)}},
			}, nil
		},
	)

	// ─── Tool 5: signal ───
	server.AddTool(
		&mcp.Tool{
			Name:        "signal",
			Description: "뉴런에 신호를 보낸다. dopamine=PD 칭찬/강화, bomb=3회 반복실수 차단, memory=기억 기록.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"path": {
						"type": "string",
						"description": "뉴런 경로"
					},
					"type": {
						"type": "string",
						"enum": ["dopamine", "bomb", "memory"],
						"description": "신호 유형"
					}
				},
				"required": ["path", "type"]
			}`),
		},
		func(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args struct {
				Path string `json:"path"`
				Type string `json:"type"`
			}
			if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
				return mcpError("invalid arguments"), nil
			}
			if args.Path == "" || args.Type == "" {
				return mcpError("path and type required"), nil
			}

			if err := signalNeuron(brainRoot, args.Path, args.Type); err != nil {
				return mcpError(err.Error()), nil
			}

			icons := map[string]string{"dopamine": "🟢", "bomb": "💣", "memory": "📝"}
			icon := icons[args.Type]
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{
					Text: fmt.Sprintf("%s %s → %s", icon, args.Type, args.Path),
				}},
			}, nil
		},
	)

	// ─── Tool 6: correct ───
	server.AddTool(
		&mcp.Tool{
			Name:        "correct",
			Description: "PD 교정을 기록한다. corrections.jsonl에 쓰는 대신 직접 뉴런을 생성/발화한다. 교정은 즉시 뉴런으로 변환된다.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"path": {
						"type": "string",
						"description": "뉴런 경로. 예: cortex/methodology/새_방법론"
					},
					"text": {
						"type": "string",
						"description": "교정 사유"
					}
				},
				"required": ["path", "text"]
			}`),
		},
		func(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args struct {
				Path string `json:"path"`
				Text string `json:"text"`
			}
			if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
				return mcpError("invalid arguments"), nil
			}
			if args.Path == "" || args.Text == "" {
				return mcpError("path and text required"), nil
			}

			// Normalize path
			neuronPath := strings.ReplaceAll(args.Path, "/", string(filepath.Separator))
			fullPath := filepath.Join(brainRoot, neuronPath)

			// Check if exists → fire, else → grow + fire
			if info, err := os.Stat(fullPath); err == nil && info.IsDir() {
				fireNeuron(brainRoot, args.Path)
				return &mcp.CallToolResult{
					Content: []mcp.Content{&mcp.TextContent{
						Text: fmt.Sprintf("📝 교정 반영: %s\n사유: %s\n결과: fired (카운터 +1)", args.Path, args.Text),
					}},
				}, nil
			}

			if err := growNeuron(brainRoot, args.Path); err != nil {
				return mcpError("grow failed: " + err.Error()), nil
			}
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{
					Text: fmt.Sprintf("📝 교정 반영 (신규): %s\n사유: %s\n결과: grown", args.Path, args.Text),
				}},
			}, nil
		},
	)

	// ─── Tool 7: evolve ───
	server.AddTool(
		&mcp.Tool{
			Name:        "evolve",
			Description: "Groq 기반 자율 뇌 진화를 실행한다. dry_run=true면 제안만, false면 실행.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"dry_run": {
						"type": "boolean",
						"default": true,
						"description": "true=제안만, false=실제 실행"
					}
				}
			}`),
		},
		func(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args struct {
				DryRun *bool `json:"dry_run"`
			}
			if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
				// Default to dry run
				args.DryRun = boolPtr(true)
			}
			dryRun := true
			if args.DryRun != nil {
				dryRun = *args.DryRun
			}

			runEvolve(brainRoot, dryRun)

			mode := "DRY RUN"
			if !dryRun {
				mode = "EXECUTED"
			}
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{
					Text: fmt.Sprintf("🧬 Evolve (%s) completed", mode),
				}},
			}, nil
		},
	)

	// ─── Tool 8: report ───
	server.AddTool(
		&mcp.Tool{
			Name:        "report",
			Description: "적층형 보고 큐. 사용자 보고/요청을 큐에 쌓는다. 요청 처리 후 자동 팔로업. priority: urgent(즉시)/normal(적층)/low(유휴시)",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"message": {
						"type": "string",
						"description": "보고 내용"
					},
					"priority": {
						"type": "string",
						"enum": ["urgent", "normal", "low"],
						"default": "normal",
						"description": "urgent=즉시처리, normal=현재작업후, low=유휴시"
					}
				},
				"required": ["message"]
			}`),
		},
		func(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args struct {
				Message  string `json:"message"`
				Priority string `json:"priority"`
			}
			if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
				return mcpError("invalid arguments"), nil
			}
			if args.Message == "" {
				return mcpError("message required"), nil
			}
			if args.Priority == "" {
				args.Priority = "normal"
			}

			// Write to _inbox/reports/ as timestamped file
			reportsDir := filepath.Join(brainRoot, "_inbox", "reports")
			os.MkdirAll(reportsDir, 0755)

			ts := fmt.Sprintf("%d", time.Now().UnixMilli())
			filename := fmt.Sprintf("%s_%s.report", ts, args.Priority)
			reportPath := filepath.Join(reportsDir, filename)

			content := fmt.Sprintf("priority: %s\ntimestamp: %s\n\n%s\n",
				args.Priority,
				time.Now().Format("2006-01-02 15:04:05"),
				args.Message)
			os.WriteFile(reportPath, []byte(content), 0644)

			// Count pending reports
			entries, _ := os.ReadDir(reportsDir)
			pending := 0
			for _, e := range entries {
				if !e.IsDir() && strings.HasSuffix(e.Name(), ".report") {
					pending++
				}
			}

			priorityIcons := map[string]string{"urgent": "🔴", "normal": "🟡", "low": "🔵"}
			icon := priorityIcons[args.Priority]
			if icon == "" {
				icon = "🟡"
			}

			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{
					Text: fmt.Sprintf("%s 새로운 보고가 확인되었습니다.\n\n요약: %s\n우선순위: %s\n대기열: %d건\n\n사용자의 요청 처리 후 팔로업합니다.", icon, args.Message, args.Priority, pending),
				}},
			}, nil
		},
	)

	// ─── Tool 9: pending_reports ───
	server.AddTool(
		&mcp.Tool{
			Name:        "pending_reports",
			Description: "대기 중인 보고 목록을 반환. done=true로 호출하면 가장 오래된 보고를 처리 완료 표시.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"done": {
						"type": "boolean",
						"default": false,
						"description": "true면 가장 오래된 보고를 처리 완료로 표시"
					}
				}
			}`),
		},
		func(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args struct {
				Done *bool `json:"done"`
			}
			json.Unmarshal(req.Params.Arguments, &args)

			reportsDir := filepath.Join(brainRoot, "_inbox", "reports")
			doneDir := filepath.Join(brainRoot, "_inbox", "reports_done")

			entries, _ := os.ReadDir(reportsDir)

			// If done=true, move oldest report to done
			if args.Done != nil && *args.Done && len(entries) > 0 {
				os.MkdirAll(doneDir, 0755)
				for _, e := range entries {
					if strings.HasSuffix(e.Name(), ".report") {
						src := filepath.Join(reportsDir, e.Name())
						dst := filepath.Join(doneDir, e.Name())
						os.Rename(src, dst)
						break // move only oldest
					}
				}
				entries, _ = os.ReadDir(reportsDir) // refresh
			}

			if len(entries) == 0 {
				return &mcp.CallToolResult{
					Content: []mcp.Content{&mcp.TextContent{Text: "✅ 대기 중인 보고 없음"}},
				}, nil
			}

			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("📋 대기 중인 보고: %d건\n\n", len(entries)))
			for i, e := range entries {
				if !strings.HasSuffix(e.Name(), ".report") {
					continue
				}
				data, _ := os.ReadFile(filepath.Join(reportsDir, e.Name()))
				sb.WriteString(fmt.Sprintf("─── [%d] %s ───\n%s\n", i+1, e.Name(), string(data)))
			}

			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: sb.String()}},
			}, nil
		},
	)



	// ─── Tool 11: rollback ───
	server.AddTool(
		&mcp.Tool{
			Name:        "rollback",
			Description: "기존 뉴런의 카운터를 1 감소시킨다(최소 0). 잘못된 방향으로 진화한 뉴런을 교정할 때 사용한다.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"path": {
						"type": "string",
						"description": "뉴런 경로. 예: cortex/testing/검증_E2E"
					}
				},
				"required": ["path"]
			}`),
		},
		func(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args struct {
				Path string `json:"path"`
			}
			if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
				return mcpError("invalid arguments"), nil
			}
			if args.Path == "" {
				return mcpError("path required"), nil
			}

			if err := rollbackNeuron(brainRoot, args.Path); err != nil {
				return mcpError(err.Error()), nil
			}
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("⏪ rolled back: %s", args.Path)}},
			}, nil
		},
	)

	// ─── Tool 12: status ───
	server.AddTool(
		&mcp.Tool{
			Name:        "status",
			Description: "뇌 상태 요약 + 최근 뉴런 발화 이력을 반환. 세션 시작 시, 작업 전환 시, 교정 후 확인 시 호출.",
			InputSchema: json.RawMessage(`{"type": "object", "properties": {}}`),
		},
		func(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			brain := scanBrain(brainRoot)

			// Count totals
			totalNeurons := 0
			totalFires := 0
			var topNeurons []string
			for _, region := range brain.Regions {
				for _, n := range region.Neurons {
					totalNeurons++
					totalFires += n.Counter
					if n.Counter >= 5 {
						topNeurons = append(topNeurons, fmt.Sprintf("  %s/%s (c:%d)", region.Name, n.Path, n.Counter))
					}
				}
			}

			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("🧠 NeuronFS 상태 [%s]\n", time.Now().Format("15:04:05")))
			sb.WriteString(fmt.Sprintf("  뉴런: %d개 | 총 발화: %d회\n", totalNeurons, totalFires))
			sb.WriteString(fmt.Sprintf("  영역: %d개 활성\n\n", len(brain.Regions)))

			// Top neurons
			if len(topNeurons) > 0 {
				sb.WriteString("🔥 활성 뉴런 (c≥5):\n")
				limit := 15
				if len(topNeurons) < limit {
					limit = len(topNeurons)
				}
				for _, n := range topNeurons[:limit] {
					sb.WriteString(n + "\n")
				}
				if len(topNeurons) > 15 {
					sb.WriteString(fmt.Sprintf("  ... 외 %d개\n", len(topNeurons)-15))
				}
			}

			// Recent corrections history
			histPath := filepath.Join(brainRoot, "_inbox", "corrections_history.jsonl")
			if data, err := os.ReadFile(histPath); err == nil && len(data) > 0 {
				lines := strings.Split(strings.TrimSpace(string(data)), "\n")
				sb.WriteString("\n📝 최근 교정 이력:\n")
				start := 0
				if len(lines) > 10 {
					start = len(lines) - 10
				}
				for _, line := range lines[start:] {
					line = strings.TrimSpace(line)
					if line == "" {
						continue
					}
					var entry struct {
						Path string `json:"path"`
						Text string `json:"text"`
					}
					if json.Unmarshal([]byte(line), &entry) == nil {
						sb.WriteString(fmt.Sprintf("  - %s: %s\n", entry.Path, entry.Text))
					}
				}
			} else {
				sb.WriteString("\n📝 교정 이력 없음\n")
			}

			// Pending reports
			reportsDir := filepath.Join(brainRoot, "_inbox", "reports")
			if entries, err := os.ReadDir(reportsDir); err == nil {
				pending := 0
				for _, e := range entries {
					if strings.HasSuffix(e.Name(), ".report") {
						pending++
					}
				}
				if pending > 0 {
					sb.WriteString(fmt.Sprintf("\n📋 대기 보고: %d건\n", pending))
				}
			}

			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: sb.String()}},
			}, nil
		},
	)

	// ─── Tool 13: health_check ───
	server.AddTool(
		&mcp.Tool{
			Name:        "health_check",
			Description: "뇌 건강 검진. 중복 뉴런 탐지, 빈 폴더 감지, bomb 상태 확인, 병합 제안을 반환. 수동 호출 또는 자동 호출.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"auto_fix": {
						"type": "boolean",
						"default": false,
						"description": "true면 자동 수정 (빈 폴더 삭제, 중복 병합). false면 보고만."
					}
				}
			}`),
		},
		func(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args struct {
				AutoFix *bool `json:"auto_fix"`
			}
			json.Unmarshal(req.Params.Arguments, &args)
			autoFix := false
			if args.AutoFix != nil {
				autoFix = *args.AutoFix
			}

			brain := scanBrain(brainRoot)
			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("🏥 뇌 건강 검진 [%s]\n\n", time.Now().Format("15:04:05")))

			// 1. Detect duplicates (Jaccard similarity)
			type neuronInfo struct {
				region string
				path   string
				full   string
			}
			var allNeurons []neuronInfo
			for _, region := range brain.Regions {
				for _, n := range region.Neurons {
					allNeurons = append(allNeurons, neuronInfo{
						region: region.Name,
						path:   n.Path,
						full:   region.Name + "/" + n.Path,
					})
				}
			}

			duplicates := 0
			var dupPairs []string
			for i := 0; i < len(allNeurons); i++ {
				for j := i + 1; j < len(allNeurons); j++ {
					tokensI := strings.FieldsFunc(allNeurons[i].path, func(r rune) bool { return r == '/' || r == '_' || r == '\\' })
					tokensJ := strings.FieldsFunc(allNeurons[j].path, func(r rune) bool { return r == '/' || r == '_' || r == '\\' })
					sim := jaccardSimilarity(tokensI, tokensJ)
					if sim >= 0.7 {
						duplicates++
						dupPairs = append(dupPairs, fmt.Sprintf("  ⚠️ %.0f%% 유사: %s ↔ %s",
							sim*100, allNeurons[i].full, allNeurons[j].full))
					}
				}
			}

			if duplicates > 0 {
				sb.WriteString(fmt.Sprintf("🔄 중복 의심: %d쌍\n", duplicates))
				limit := 10
				if len(dupPairs) < limit {
					limit = len(dupPairs)
				}
				for _, p := range dupPairs[:limit] {
					sb.WriteString(p + "\n")
				}
			} else {
				sb.WriteString("🔄 중복: 없음 ✅\n")
			}

			// 2. Empty folders
			emptyCount := 0
			var emptyDirs []string
			for _, region := range brain.Regions {
				for _, n := range region.Neurons {
					if n.Counter == 0 {
						emptyCount++
						emptyDirs = append(emptyDirs, region.Name+"/"+n.Path)
					}
				}
			}
			if emptyCount > 0 {
				sb.WriteString(fmt.Sprintf("\n📂 빈 뉴런 (c:0): %d개\n", emptyCount))
				if autoFix {
					// Remove empty neuron folders
					removed := 0
					for _, d := range emptyDirs {
						fullPath := filepath.Join(brainRoot, strings.ReplaceAll(d, "/", string(filepath.Separator)))
						if err := os.RemoveAll(fullPath); err == nil {
							removed++
						}
					}
					sb.WriteString(fmt.Sprintf("  🗑️ %d개 자동 삭제\n", removed))
				}
			} else {
				sb.WriteString("\n📂 빈 뉴런: 없음 ✅\n")
			}

			// 3. Bomb status
			bombCount := 0
			for _, region := range brain.Regions {
				for _, n := range region.Neurons {
					bombPath := filepath.Join(n.FullPath, "bomb.neuron")
					if _, err := os.Stat(bombPath); err == nil {
						bombCount++
						sb.WriteString(fmt.Sprintf("\n💣 BOMB: %s/%s\n", region.Name, n.Path))
					}
				}
			}
			if bombCount == 0 {
				sb.WriteString("\n💣 Bomb: 없음 ✅\n")
			}

			// Summary
			sb.WriteString(fmt.Sprintf("\n── 총평: 뉴런 %d개 | 중복 %d쌍 | 빈 폴더 %d | bomb %d\n",
				len(allNeurons), duplicates, emptyCount, bombCount))

			if duplicates == 0 && emptyCount == 0 && bombCount == 0 {
				sb.WriteString("🟢 건강 상태: 양호\n")
			} else {
				sb.WriteString("🟡 건강 상태: 점검 필요\n")
			}

			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: sb.String()}},
			}, nil
		},
	)

	// ─── Register new feature suite ───
	RegisterNativeTools(server, brainRoot)
}

// ─── Helpers ───

func mcpError(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: "❌ " + msg}},
		IsError: true,
	}
}

// boolPtr returns a pointer to a boolean value.
func boolPtr(b bool) *bool {
	return &b
}

// logWriter returns stderr for MCP mode (stdout is reserved for JSON-RPC)
func logWriter() *os.File {
	return os.Stderr
}
