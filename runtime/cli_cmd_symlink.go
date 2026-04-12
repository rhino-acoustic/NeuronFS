package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// SymlinkCmd implements the Command interface for the legacy --symlink action.
type SymlinkCmd struct{}

func (c *SymlinkCmd) Name() string {
	return "--symlink"
}

func (c *SymlinkCmd) Execute(brainRoot string, args []string) error {
	targetDir := ""
	for i, arg := range args {
		if arg == "--symlink" && i+1 < len(args) {
			targetDir = args[i+1]
			break
		}
	}
	if targetDir == "" {
		fmt.Println("[FATAL] Usage: neuronfs <brain> --symlink <global_path>")
		os.Exit(1)
	}

	sharedDir := filepath.Join(brainRoot, ".neuronfs", "shared")
	os.MkdirAll(filepath.Dir(sharedDir), 0750)

	absTarget, _ := filepath.Abs(targetDir)
	err := os.Symlink(absTarget, sharedDir)
	if err != nil {
		fmt.Printf("[FATAL] Failed to link: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("[SYNK] Linked %s -> %s\n", absTarget, sharedDir)
	return nil
}
