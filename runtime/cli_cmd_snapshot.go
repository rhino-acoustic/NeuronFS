package main

// SnapshotCmd implements the Command interface for the legacy --snapshot action.
// PROVIDES: Strangler Fig CLI migration for the "snapshot" mode.
type SnapshotCmd struct{}

func (c *SnapshotCmd) Name() string {
	return "--snapshot"
}

func (c *SnapshotCmd) Execute(brainRoot string, args []string) error {
	gitSnapshot(brainRoot)
	return nil
}
