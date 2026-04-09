package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func registerMCPCRUDTools(server *mcp.Server, brainRoot string) {
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

			signalPath := filepath.Join(brainRoot, "hippocampus", "_signals")
			os.MkdirAll(signalPath, 0755)
			ts := fmt.Sprintf("%d", time.Now().UnixMilli())
			sigFile := filepath.Join(signalPath, fmt.Sprintf("signal_%s.json", ts))

			payload := map[string]string{
				"type": "GROW_INTENT",
				"path": args.Path,
				"ts":   time.Now().Format("2006-01-02T15:04:05"),
			}
			data, _ := json.Marshal(payload)
			os.WriteFile(sigFile, data, 0644)

			return mcpWithRules(brainRoot, fmt.Sprintf("🌱 신호 기록됨 (수면(REM) 통합 대기): %s", args.Path)), nil
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
			return mcpWithRules(brainRoot, fmt.Sprintf("🔥 fired: %s", args.Path)), nil
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
			Description: "PD 교정을 기록한다. 뉴런 생성/발화 + corrections 로그 동시 기록. 하네스 사이클이 이 로그를 분석하여 개인화 뉴런을 자동 생성한다.",
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

			// ── 1. 교정 로그 기록 (하네스 개인화 데이터 소스) ──
			inboxDir := filepath.Join(brainRoot, "_inbox")
			os.MkdirAll(inboxDir, 0755)

			corrEntry := fmt.Sprintf(`{"path":"%s","text":"%s","ts":"%s"}`,
				strings.ReplaceAll(args.Path, `"`, `\"`),
				strings.ReplaceAll(args.Text, `"`, `\"`),
				time.Now().Format("2006-01-02T15:04:05"))

			// corrections.jsonl — 하네스 사이클 입력 (processInbox가 소비 후 비움)
			corrPath := filepath.Join(inboxDir, "corrections.jsonl")
			f, _ := os.OpenFile(corrPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if f != nil {
				f.WriteString(corrEntry + "\n")
				f.Close()
			}

			// corrections_history.jsonl — 영구 이력 (status 도구에서 조회)
			histPath := filepath.Join(inboxDir, "corrections_history.jsonl")
			h, _ := os.OpenFile(histPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if h != nil {
				h.WriteString(corrEntry + "\n")
				h.Close()
			}

			// ── 2. 뉴런 신호 기록 (직접 생성 안함) ──
			// (단, 기존에 존재하는 뉴런이라면 fire는 허용)
			neuronPath := strings.ReplaceAll(args.Path, "/", string(filepath.Separator))
			fullPath := filepath.Join(brainRoot, neuronPath)

			action := "signal recorded"
			if info, err := os.Stat(fullPath); err == nil && info.IsDir() {
				fireNeuron(brainRoot, args.Path)
				action = "fired (카운터 +1)"
			} else {
				signalPath := filepath.Join(brainRoot, "hippocampus", "_signals")
				os.MkdirAll(signalPath, 0755)
				ts := fmt.Sprintf("%d", time.Now().UnixMilli())
				sigFile := filepath.Join(signalPath, fmt.Sprintf("signal_%s.json", ts))

				payload := map[string]string{
					"type": "CORRECT_INTENT",
					"path": args.Path,
					"text": args.Text,
					"ts":   time.Now().Format("2006-01-02T15:04:05"),
				}
				data, _ := json.Marshal(payload)
				os.WriteFile(sigFile, data, 0644)
			}

			return mcpWithRules(brainRoot, fmt.Sprintf("📝 교정 반영 (Signal): %s\n사유: %s\n결과: %s\n⚡ 하네스 로그 및 REM 수면 큐에 기록됨", args.Path, args.Text, action)), nil
		},
	)
}
