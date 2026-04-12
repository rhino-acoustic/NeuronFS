package main

import (
	"fmt"
	"os"
)

// GrowCmd implements the Command interface for the legacy --grow action.
// PROVIDES: Strangler Fig CLI migration for the "grow" mode.
type GrowCmd struct{}

func (c *GrowCmd) Name() string {
	return "--grow"
}

func (c *GrowCmd) Execute(brainRoot string, args []string) error {
	neuronPath := getNonFlagArg(1) // brain_v4=0, path=1
	if neuronPath == "" {
		fmt.Println("[FATAL] Usage: neuronfs <brain> --grow <region/path/to/neuron>")
		fmt.Println("  Example: neuronfs brain_v4 --grow cortex/frontend/coding/no_console_log")
		os.Exit(1)
	}
	growNeuron(brainRoot, neuronPath)
	return nil
}
