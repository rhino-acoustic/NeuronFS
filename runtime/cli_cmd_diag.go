package main

// DiagCmd implements the Command interface for the legacy --diag diagnostic action.
type DiagCmd struct{}

func (c *DiagCmd) Name() string {
	return "--diag"
}

func (c *DiagCmd) Execute(brainRoot string, args []string) error {
	brain := scanBrain(brainRoot)
	result := runSubsumption(brain)
	printDiag(brain, result)
	return nil
}
