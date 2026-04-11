package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// runWorkerTool parses the MCP tool name and JSON arguments, calls the appropriate internal function,
// and outputs the result to stdout for the MCP Supervisor proxy to consume.
// This allows true stateless Hot-Swap execution of logic without dropping the supervisor.
func runWorkerTool(brainRoot string, toolName string, argsJson string) {
	switch toolName {
	case "grow":
		var args struct {
			Path string `json:"path"`
		}
		if err := json.Unmarshal([]byte(argsJson), &args); err != nil {
			fmt.Println(`{"error":"invalid grow arguments"}`)
			os.Exit(1)
		}
		// In actual Hot Swap, we don't output pure text. We output a proper mcp.CallToolResult struct
		// or standard text that the supervisor will wrap into a tool result.
		// For simplicity, we just output the message. The supervisor will catch it.
		// NOTE: Currently grow just writes signal or fires.
		growNeuron(brainRoot, args.Path) // We modify growNeuron to be callable here, but wait, grow is inside mcp_handler_crud normally.
		fmt.Printf("🌱 신호 기록됨 (수면(REM) 통합 대기): %s\n", args.Path)

	case "fire":
		var args struct {
			Path string `json:"path"`
		}
		if err := json.Unmarshal([]byte(argsJson), &args); err != nil {
			fmt.Println(`{"error":"invalid fire arguments"}`)
			os.Exit(1)
		}
		fireNeuron(brainRoot, args.Path)
		fmt.Printf("🔥 fired: %s\n", args.Path)

	case "signal":
		var args struct {
			Path string `json:"path"`
			Type string `json:"type"`
		}
		if err := json.Unmarshal([]byte(argsJson), &args); err != nil {
			fmt.Println(`{"error":"invalid signal arguments"}`)
			os.Exit(1)
		}
		if err := signalNeuron(brainRoot, args.Path, args.Type); err != nil {
			fmt.Printf("error: %v\n", err)
			os.Exit(1)
		}
		icons := map[string]string{"dopamine": "🟢", "bomb": "💣", "memory": "📝"}
		icon := icons[args.Type]
		fmt.Printf("%s %s → %s\n", icon, args.Type, args.Path)

	case "rollback":
		var args struct {
			Path string `json:"path"`
		}
		json.Unmarshal([]byte(argsJson), &args)
		if err := rollbackNeuron(brainRoot, args.Path); err != nil {
			fmt.Printf("error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("⏪ rolled back: %s\n", args.Path)

	default:
		fmt.Printf("error: unknown tool action %s\n", toolName)
		os.Exit(1)
	}
}
