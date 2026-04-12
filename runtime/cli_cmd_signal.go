package main

import (
	"fmt"
	"os"
)

// SignalCmd implements the Command interface for the legacy --signal action.
// PROVIDES: Strangler Fig CLI migration for the "signal" mode.
type SignalCmd struct{}

func (c *SignalCmd) Name() string {
	return "--signal"
}

func (c *SignalCmd) Execute(brainRoot string, args []string) error {
	sigType := ""
	// Find sigType (the argument immediately after --signal)
	for i, arg := range args {
		if arg == "--signal" && i+1 < len(args) {
			sigType = args[i+1]
			break
		}
	}
	// Find neuronPath (the first non-flag arg after brainRoot and sigType)
	neuronPath := getNonFlagArg(1) 
	if sigType == "" || neuronPath == "" {
		fmt.Println("[FATAL] Usage: neuronfs <brain> --signal dopamine|bomb|memory <path>")
		os.Exit(1)
	}
	signalNeuron(brainRoot, neuronPath, sigType)
	return nil
}
