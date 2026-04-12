package main

import (
	"fmt"
	"os"
)

// McpCmd implements the Command interface for the legacy --mcp action.
// PROVIDES: Strangler Fig CLI migration for the "mcp" mode.
type McpCmd struct{}

func (c *McpCmd) Name() string {
	return "--mcp"
}

func (c *McpCmd) Execute(brainRoot string, args []string) error {
	// MCP Streamable HTTP server + background loops
	// HTTP transport: IDE 재시작에도 연결 유지
	go func() {
		mcpAPIPort := MCPPort
		fmt.Fprintf(os.Stderr, "[MCP] REST API on :%d (fallback)\n", mcpAPIPort)
		startAPI(brainRoot, mcpAPIPort)
	}()
	go runInjectionLoop(brainRoot)
	go runIdleLoop(brainRoot)
	startMCPHTTPServer(brainRoot, MCPStreamPort) // blocking: HTTP server
	return nil
}
