package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
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
		// grow는 즉시 생성이 아니라 hippocampus signal을 생성하여 REM 수면 시 백그라운드 병합되게 함.
		signalPath := filepath.Join(brainRoot, "hippocampus", "_signals")
		os.MkdirAll(signalPath, 0750)
		ts := fmt.Sprintf("%d", time.Now().UnixMilli())
		sigFile := filepath.Join(signalPath, fmt.Sprintf("signal_%s.json", ts))

		payload := map[string]string{
			"type": "GROW_INTENT",
			"path": args.Path,
			"ts":   time.Now().Format("2006-01-02T15:04:05"),
		}
		data, _ := json.Marshal(payload)
		os.WriteFile(sigFile, data, 0600)
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

	case "inject_tick":
		// Called statelessly by the supervisor when fsnotify triggers for _inbox
		processInbox(brainRoot)
		if consumeDirty() {
			autoReinject(brainRoot)
		}

	default:
		fmt.Printf("error: unknown tool action %s\n", toolName)
		os.Exit(1)
	}
}
