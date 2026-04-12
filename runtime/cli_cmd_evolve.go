package main

// EvolveCmd implements the Command interface for the legacy --evolve action.
// PROVIDES: Strangler Fig CLI migration for the "evolve" mode.
type EvolveCmd struct{}

func (c *EvolveCmd) Name() string {
	return "--evolve"
}

func (c *EvolveCmd) Execute(brainRoot string, args []string) error {
	dryRun := false
	for _, arg := range args {
		if arg == "--dry-run" {
			dryRun = true
		}
	}
	runEvolve(brainRoot, dryRun)
	return nil
}
