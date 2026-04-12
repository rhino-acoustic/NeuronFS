package main

// StatsCmd implements the Command interface for the legacy --stats action.
// PROVIDES: Strangler Fig CLI migration for the "stats" mode.
type StatsCmd struct{}

func (c *StatsCmd) Name() string {
	return "--stats"
}

func (c *StatsCmd) Execute(brainRoot string, args []string) error {
	runStats(brainRoot)
	return nil
}
