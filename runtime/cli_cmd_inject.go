package main

// InjectCmd implements the Command interface for the legacy --inject action.
// PROVIDES: Strangler Fig CLI migration for the "inject" mode.
type InjectCmd struct{}

func (c *InjectCmd) Name() string {
	return "--inject"
}

func (c *InjectCmd) Execute(brainRoot string, args []string) error {
	processInbox(brainRoot)
	writeAllTiers(brainRoot)
	return nil
}
