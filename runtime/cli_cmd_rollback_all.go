package main

import (
	"fmt"
	"os"
)

// RollbackAllCmd implements the Command interface for the legacy --rollback-all action.
// PROVIDES: Strangler Fig CLI migration for the "rollback-all" mode.
type RollbackAllCmd struct{}

func (c *RollbackAllCmd) Name() string {
	return "--rollback-all"
}

func (c *RollbackAllCmd) Execute(brainRoot string, args []string) error {
	fmt.Printf("%s[PRUNE] Toxic memories detected. Purging ALL corrupted synapses...%s\n", ansiMagenta, ansiReset)
	if err := rollbackAll(brainRoot); err != nil {
		fmt.Printf("%s[TRAUMA] Global rollback failed: %v%s\n", ansiRed, err, ansiReset)
		os.Exit(1)
	}
	fmt.Printf("%s[RESTORE] Brainstem architecture fully re-aligned via SSOT.%s\n", ansiWhite, ansiReset)
	fmt.Printf("%s[NEURON] Cortex online. Process stabilized.%s\n", ansiGreen, ansiReset)
	return nil
}
