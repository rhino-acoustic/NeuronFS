package main

import (
	"fmt"
	"os"
)

// FireCmd implements the Command interface for the legacy --fire action.
// PROVIDES: Strangler Fig CLI migration for the "fire" mode.
type FireCmd struct{}

func (c *FireCmd) Name() string {
	return "--fire"
}

func (c *FireCmd) Execute(brainRoot string, args []string) error {
	neuronPath := getNonFlagArg(1)
	if neuronPath == "" {
		fmt.Println("[FATAL] Usage: neuronfs <brain> --fire <region/path/to/neuron>")
		os.Exit(1)
	}
	fireNeuron(brainRoot, neuronPath)
	return nil
}
