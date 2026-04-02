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
	// ?€?€?€ Native Tool 1: read_neuron ?€?€?€
	s.AddTool(
		&mcp.Tool{
			Name:        "read_neuron",
			Description: "?¹ì • ?´ëŸ°??ê·œì¹™???¤ì‹œê°„ìœ¼ë¡?ë°˜í™˜?œë‹¤.",
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

	// ?€?€?€ Native Tool 2: write_message ?€?€?€
	s.AddTool(
		&mcp.Tool{
			Name:        "write_message",
			Description: "inbox/outboxë¥?ì§ì ‘ ì»¨íŠ¸ë¡¤í•˜???¨ìˆ˜.",
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
			os.MkdirAll(boxDir, 0755)

			filename := fmt.Sprintf("msg_%s.md", time.Now().Format("20060102_150405"))
			filePath := filepath.Join(boxDir, filename)
			if err := os.WriteFile(filePath, []byte(args.Content), 0644); err != nil {
				return mcpError("failed writing box: " + err.Error()), nil
			}

			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: "Message written to " + filename}},
			}, nil
		},
	)

	// ?€?€?€ Native Tool 3: grow_neuron ?€?€?€
	s.AddTool(
		&mcp.Tool{
			Name:        "grow_neuron",
			Description: "?„íŒŒë¯?ë°??œëƒ…???±ì¥??ê´€ë¦¬í•œ?? ì¹´ìš´??ì¦ê? ë°?ê¸ì •??ê°•í™”?˜ì¹˜ ?ìš©.",
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
			os.WriteFile(filepath.Join(fullPath, "payload.json"), []byte(args.RuleData), 0644)

			// ê¸°ê³„??ì¹?°¬ ë°©ì? (Dopamine Inflation Fix)
			praiseRegex := regexp.MustCompile(`(?i)(ì¹?°¬|??s*?°ì…¨?µë‹ˆ??ì¢‹ì•„|?Œë?|?„ë²½|ìµœê³ )`)

			if args.EmotionalWeight > 0 || praiseRegex.MatchString(args.RuleData) {
				if args.Author == "pm" || args.Author == "BASEMENT_ADMIN" || strings.Contains(strings.ToLower(args.Author), "pd") {
					_ = signalNeuron(brainRoot, args.Path, "dopamine")
				} else {
					return &mcp.CallToolResult{
						Content: []mcp.Content{&mcp.TextContent{Text: "???´ëŸ° ?±ì¥?? ???ì´?„íŠ¸ ê°?ë¦¬ë·° ê³¼ì •?ì„œ??ê°ì •???¤ì›Œ???„íŒŒë¯??¸í”Œ?ˆì´????ë¬´ì‹œ?˜ì—ˆ?µë‹ˆ??"}},
					}, nil
				}
			}

			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: "???´ëŸ° ?œëƒ…??ê°±ì‹  ?±ê³µ (ê²½ë¡œ: " + args.Path + ")"}},
			}, nil
		},
	)

	// ?€?€?€ Native Tool 4: get_dashboard_state ?€?€?€
	s.AddTool(
		&mcp.Tool{
			Name:        "get_dashboard_state",
			Description: "?€?œë³´???¤ì‹œê°?API???íƒœê°’ì„ ë°˜í™˜.",
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
}

