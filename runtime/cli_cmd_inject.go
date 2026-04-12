package main

// InjectCmd implements the Command interface for the legacy --inject action.
// PROVIDES: Strangler Fig CLI migration for the "inject" mode.
type InjectCmd struct{}

func (c *InjectCmd) Name() string {
	return "--inject"
}

func (c *InjectCmd) Execute(brainRoot string, args []string) error {
	GlobalSSEBroker.Broadcastf("info", "[CLI] --inject Triggered: Processing Inbox & Rebuilding GEMINI.md")
	processInbox(brainRoot)
	writeAllTiers(brainRoot)
	GlobalSSEBroker.Broadcastf("success", "[CLI] --inject Compiling Tiers Complete")
	return nil
}
