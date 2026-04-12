package main

// NeuronizeCmd implements the Command interface for the legacy --neuronize action.
type NeuronizeCmd struct{}

func (c *NeuronizeCmd) Name() string {
	return "--neuronize"
}

func (c *NeuronizeCmd) Execute(brainRoot string, args []string) error {
	dryRun := false
	for _, arg := range args {
		if arg == "--dry-run" {
			dryRun = true
			break
		}
	}
	runNeuronize(brainRoot, dryRun)
	return nil
}
