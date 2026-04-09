package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func registerMCPEvolveTools(server *mcp.Server, brainRoot string) {
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

}
