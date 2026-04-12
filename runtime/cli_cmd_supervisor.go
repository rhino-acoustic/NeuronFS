package main

// SupervisorCmd implements the Command interface for the legacy --supervisor action.
// PROVIDES: Strangler Fig CLI migration for the "supervisor" mode.
type SupervisorCmd struct{}

func (c *SupervisorCmd) Name() string {
	return "--supervisor"
}

func (c *SupervisorCmd) Execute(brainRoot string, args []string) error {
	runSupervisor(brainRoot)
	return nil
}
