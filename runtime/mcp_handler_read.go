package main

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func registerMCPReadTools(server *mcp.Server, brainRoot string) {
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
}
