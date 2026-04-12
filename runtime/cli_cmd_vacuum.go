package main

// VacuumCmd implements the Command interface for the legacy --vacuum action.
// PROVIDES: Strangler Fig CLI migration for the "vacuum" mode.
type VacuumCmd struct{}

func (c *VacuumCmd) Name() string {
	return "--vacuum"
}

func (c *VacuumCmd) Execute(brainRoot string, args []string) error {
	runVacuum(brainRoot)
	return nil
}
