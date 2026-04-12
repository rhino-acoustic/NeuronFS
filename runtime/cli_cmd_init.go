package main

import (
	"path/filepath"
	"strings"
)

// InitCmd handles initializing a new brain directory.
type InitCmd struct{}

func (c *InitCmd) Name() string {
	return "--init"
}

func (c *InitCmd) Execute(brainRoot string, args []string) error {
	initPath := ""
	for _, arg := range args {
		// skip elements that are the CLI executable, the root, or the flag itself
		// just look for the first raw argument assuming it's the path.
		// Wait, how do args look like? args = os.Args
		if !strings.HasPrefix(arg, "--") && arg != args[0] && arg != brainRoot {
			initPath = arg
			break
		}
	}
	// Better way: skip args[0] and flags
	if initPath == "" {
		for _, arg := range args[1:] {
			if !strings.HasPrefix(arg, "--") && arg != brainRoot {
				initPath = arg
				break
			}
		}
	}

	if initPath == "" {
		initPath = filepath.Join(".", "brain_v4")
	}
	abs, _ := filepath.Abs(initPath)
	initBrain(abs)
	return nil
}
