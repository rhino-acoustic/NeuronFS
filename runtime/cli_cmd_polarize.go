package main

// PolarizeCmd implements the Command interface for the legacy --polarize action.
type PolarizeCmd struct{}

func (c *PolarizeCmd) Name() string {
	return "--polarize"
}

func (c *PolarizeCmd) Execute(brainRoot string, args []string) error {
	dryRun := false
	for _, arg := range args {
		if arg == "--dry-run" {
			dryRun = true
			break
		}
	}
	runPolarize(brainRoot, dryRun)
	return nil
}
