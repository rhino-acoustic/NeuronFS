package main

import (
	"fmt"
	"os"
)

// RollbackCmd implements the Command interface for the legacy --rollback action.
// PROVIDES: Strangler Fig CLI migration for the "rollback" mode.
type RollbackCmd struct{}

func (c *RollbackCmd) Name() string {
	return "--rollback"
}

func (c *RollbackCmd) Execute(brainRoot string, args []string) error {
	neuronPath := getNonFlagArg(1)
	if neuronPath == "" {
		fmt.Println("[FATAL] Usage: neuronfs <brain> --rollback <region/path/to/neuron>")
		os.Exit(1)
	}
	if err := rollbackNeuron(brainRoot, neuronPath); err != nil {
		fmt.Printf("[ERROR] %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("\033[35m[PRUNE] Toxic memories detected. Purging corrupted synapses...\033[0m\n")
	fmt.Printf("\033[37m[RESTORE] Brainstem architecture fully re-aligned via SSOT.\033[0m\n")
	return nil
}
