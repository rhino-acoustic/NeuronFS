package main

import (
	"fmt"
	"strings"
)

// EmitCmd implements the Command interface for the core --emit action.
type EmitCmd struct{}

func (c *EmitCmd) Name() string {
	return "--emit"
}

func (c *EmitCmd) Execute(brainRoot string, args []string) error {
	emitTarget := ""
	for i, arg := range args {
		if arg == "--emit" && i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
			candidate := strings.ToLower(args[i+1])
			if candidate == "gemini" || candidate == "cursor" || candidate == "claude" || candidate == "copilot" || candidate == "generic" || candidate == "all" || candidate == "auto" {
				emitTarget = candidate
			}
		}
	}

	if emitTarget != "" {
		processInbox(brainRoot)
		writeAllTiersForTargets(brainRoot, emitTarget)
		return nil
	}

	brain := scanBrain(brainRoot)
	result := runSubsumption(brain)
	fmt.Print(emitRules(result))
	return nil
}
