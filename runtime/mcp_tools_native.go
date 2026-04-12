package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterNativeTools add 4 required native tools
func RegisterNativeTools(s *mcp.Server, brainRoot string) {
	// ─── Native Tool 1: read_neuron ───
	s.AddTool(
		&mcp.Tool{
			Name:        "read_neuron",
			Description: "특정 뉴런의 규칙을 실시간으로 반환한다.",
			InputSchema: json.RawMessage(`{"type": "object", "properties": {"path": {"type": "string"}}, "required": ["path"]}`),
		},
		func(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args struct {
				Path string `json:"path"`
			}
			if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
				return mcpError("invalid arguments: " + err.Error()), nil
			}

			fullPath := filepath.Join(brainRoot, strings.ReplaceAll(args.Path, "/", string(filepath.Separator)))
			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				return mcpError("neuron not found: " + args.Path), nil
			}

			entries, err := os.ReadDir(fullPath)
			if err != nil {
				return mcpError("failed to read neuron dir: " + err.Error()), nil
			}

			content := "Neuron Path: " + args.Path + "\n\n"
			for _, e := range entries {
				if e.IsDir() {
					continue
				}
				data, err := os.ReadFile(filepath.Join(fullPath, e.Name()))
				if err == nil {
					content += fmt.Sprintf("=== %s ===\n%s\n\n", e.Name(), string(data))
				}
			}

			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: content}},
			}, nil
		},
	)

	// ─── Native Tool 2: write_message ───
	s.AddTool(
		&mcp.Tool{
			Name:        "write_message",
			Description: "inbox/outbox를 직접 컨트롤하는 함수.",
			InputSchema: json.RawMessage(`{"type": "object", "properties": {"target_bot": {"type": "string"}, "message_type": {"type": "string", "description":"inbox or outbox"}, "content": {"type": "string"}}, "required": ["target_bot", "message_type", "content"]}`),
		},
		func(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args struct {
				TargetBot   string `json:"target_bot"`
				MessageType string `json:"message_type"`
				Content     string `json:"content"`
			}
			if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
				return mcpError("invalid arguments: " + err.Error()), nil
			}

			boxDir := filepath.Join(brainRoot, "_agents", args.TargetBot, args.MessageType)
			os.MkdirAll(boxDir, 0750)

			filename := fmt.Sprintf("msg_%s.md", time.Now().Format("20060102_150405"))
			filePath := filepath.Join(boxDir, filename)
			if err := os.WriteFile(filePath, []byte(args.Content), 0600); err != nil {
				return mcpError("failed writing box: " + err.Error()), nil
			}

			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: "Message written to " + filename}},
			}, nil
		},
	)

	// ─── Native Tool 3: grow_neuron ───
	s.AddTool(
		&mcp.Tool{
			Name:        "grow_neuron",
			Description: "도파민 및 시냅스 성장을 관리한다. 카운터 증가 및 긍정적 강화수치 적용.",
			InputSchema: json.RawMessage(`{"type": "object", "properties": {"path": {"type": "string"}, "rule_data": {"type": "string"}, "emotional_weight": {"type": "integer"}, "author": {"type": "string"}}, "required": ["path", "rule_data", "author"]}`),
		},
		func(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args struct {
				Path            string `json:"path"`
				RuleData        string `json:"rule_data"`
				EmotionalWeight int    `json:"emotional_weight"`
				Author          string `json:"author"`
			}
			if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
				return mcpError("invalid arguments: " + err.Error()), nil
			}

			// Grow / Fire base
			if err := growNeuron(brainRoot, args.Path); err != nil {
				return mcpError("error growing: " + err.Error()), nil
			}
			fullPath := filepath.Join(brainRoot, strings.ReplaceAll(args.Path, "/", string(filepath.Separator)))
			os.WriteFile(filepath.Join(fullPath, "payload.json"), []byte(args.RuleData), 0600)

			// Write author to rule.md frontmatter (Data Lineage)
			if args.Author != "" {
				rulePath := filepath.Join(fullPath, "rule.md")
				ruleContent := fmt.Sprintf("---\nauthor: %s\n---\n%s\n", args.Author, args.RuleData)
				os.WriteFile(rulePath, []byte(ruleContent), 0600)
			}

			// 기계적 칭찬 방지 (Dopamine Inflation Fix)
			praiseRegex := regexp.MustCompile(`(?i)(칭찬|잘\s*쓰셨습니다|좋아|훌륭|완벽|최고)`)

			if args.EmotionalWeight > 0 || praiseRegex.MatchString(args.RuleData) {
				if args.Author == "pm" || args.Author == "BASEMENT_ADMIN" || strings.Contains(strings.ToLower(args.Author), "pd") {
					_ = signalNeuron(brainRoot, args.Path, "dopamine")
				} else {
					return &mcp.CallToolResult{
						Content: []mcp.Content{&mcp.TextContent{Text: "✅ 뉴런 성장됨. ⚠ 에이전트 간 리뷰 과정에서의 감정적 키워드(도파민 인플레이션)는 무시되었습니다."}},
					}, nil
				}
			}

			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: "✅ 뉴런 시냅스 갱신 성공 (경로: " + args.Path + ")"}},
			}, nil
		},
	)

	// ─── Native Tool 4: get_dashboard_state ───
	s.AddTool(
		&mcp.Tool{
			Name:        "get_dashboard_state",
			Description: "대시보드 실시간 API의 상태값을 반환.",
			InputSchema: json.RawMessage(`{"type": "object", "properties": {}}`),
		},
		func(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			data := buildBrainJSONResponse(brainRoot)
			jsonBytes, err := json.MarshalIndent(data, "", "  ")
			if err != nil {
				return mcpError("json error: " + err.Error()), nil
			}

			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: string(jsonBytes)}},
			}, nil
		},
	)

	// ─── Native Tool 5: MCP Delegation Hub (Phase 30) ───
	s.AddTool(
		&mcp.Tool{
			Name:        "delegate_task",
			Description: "P2P 네트워크를 통해 특정 전문 에이전트(bot1, enfp, architect)에게 작업을 위임한다.",
			InputSchema: json.RawMessage(`{"type": "object", "properties": {"target_agent": {"type": "string", "enum": ["bot1", "enfp", "architect"]}, "title": {"type": "string"}, "message": {"type": "string"}}, "required": ["target_agent", "title", "message"]}`),
		},
		func(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args struct {
				TargetAgent string `json:"target_agent"`
				Title       string `json:"title"`
				Message     string `json:"message"`
			}
			if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
				return mcpError("invalid arguments: " + err.Error()), nil
			}

			// Dump to inbox (Reuse Phase 27 Delegation logic)
			inboxDir := filepath.Join(brainRoot, "_agents", args.TargetAgent, "inbox")
			os.MkdirAll(inboxDir, 0755)

			fileName := fmt.Sprintf("mcp_delegation_%d.md", time.Now().UnixNano())
			filePath := filepath.Join(inboxDir, fileName)

			logContent := fmt.Sprintf("발신: MCP Network\n# %s\n\n%s\n", args.Title, args.Message)
			if err := os.WriteFile(filePath, []byte(logContent), 0644); err != nil {
				return mcpError("Server error: could not write to inbox"), nil
			}

			// Trigger SSE if available (GlobalSSEBroker must be tracked globally. Since it is declared in main package, we can access it if initialized)
			if GlobalSSEBroker != nil {
				GlobalSSEBroker.Broadcast("info", fmt.Sprintf("[MCP] Swarm Task Delegated: %s", args.TargetAgent))
			}

			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: "✅ P2P Swarm Delegation to " + args.TargetAgent + " success."}},
			}, nil
		},
	)
}
