package main

import (
	"fmt"
	"os"
)

// ToolCmd implements the Command interface for the legacy --tool action.
type ToolCmd struct{}

func (c *ToolCmd) Name() string {
	return "--tool"
}

func (c *ToolCmd) Execute(brainRoot string, args []string) error {
	toolName := ""
	argsJson := ""
	for i, arg := range args {
		if arg == "--tool" && i+2 < len(args) {
			toolName = args[i+1]
			argsJson = args[i+2]
			break
		}
	}
	if toolName == "" {
		fmt.Println("[FATAL] Usage: neuronfs <brain> --tool <toolname> <args_json>")
		os.Exit(1)
	}
	runWorkerTool(brainRoot, toolName, argsJson)
	return nil
}
