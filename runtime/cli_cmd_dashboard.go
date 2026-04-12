package main

// DashboardCmd implements the Command interface for the legacy --dashboard action.
// PROVIDES: Strangler Fig CLI migration for the "dashboard" mode.
type DashboardCmd struct{}

func (c *DashboardCmd) Name() string {
	return "--dashboard"
}

func (c *DashboardCmd) Execute(brainRoot string, args []string) error {
	startAPI(brainRoot, APIPort)
	return nil
}
