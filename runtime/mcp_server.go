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

	fmt.Fprintf(os.Stderr, "[MCP] 🧠 NeuronFS MCP server starting (stdio)...\n")

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
			Description: "적층형 보고 큐. 사용자 보고/요청을 큐에 쌓는다. 현재 작업 완료 후 heartbeat가 자동 팔로업. priority: urgent(즉시)/normal(적층)/low(유휴시)",
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

	// ─── Tool 10: heartbeat_ack ───
	server.AddTool(
		&mcp.Tool{
			Name:        "heartbeat_ack",
			Description: "Heartbeat 주입 수신 확인. bot1이 heartbeat 프롬프트를 받으면 반드시 이 도구를 호출하여 수신 완료를 기록한다. result: 작업 결과 요약.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"result": {
						"type": "string",
						"description": "처리 결과 요약. 예: '로그 스캔 완료: 뉴런 2개 생성' 또는 '추출할 뉴런 없음'"
					}
				},
				"required": ["result"]
			}`),
		},
		func(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args struct {
				Result string `json:"result"`
			}
			if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
				return mcpError("invalid arguments"), nil
			}

			ackFile := filepath.Join(brainRoot, "_inbox", "heartbeat_ack.json")
			ackData := map[string]interface{}{
				"acked_at": time.Now().Format("2006-01-02 15:04:05"),
				"result":   args.Result,
			}
			data, _ := json.MarshalIndent(ackData, "", "  ")
			os.WriteFile(ackFile, data, 0644)

			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{
					Text: fmt.Sprintf("✅ Heartbeat 수신 확인 완료.\n결과: %s\n다음 heartbeat는 쿨다운 후 전송됩니다.", args.Result),
				}},
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
