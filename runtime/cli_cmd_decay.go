package main

import (
	"fmt"
	"strconv"
	"strings"
)

// DecayCmd implements the Command interface for the legacy --decay action.
// PROVIDES: Strangler Fig CLI migration for the "decay" mode.
type DecayCmd struct{}

func (c *DecayCmd) Name() string {
	return "--decay"
}

func (c *DecayCmd) Execute(brainRoot string, args []string) error {
	daysStr := "30" // Default decay days
	// Find days (the argument immediately after --decay)
	for i, arg := range args {
		if arg == "--decay" && i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
			daysStr = args[i+1]
			break
		}
	}
	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 {
		fmt.Printf("[WARN] Invalid decay days '%s', using default 30 days.\n", daysStr)
		days = 30
	}
	runDecay(brainRoot, days)
	return nil
}
