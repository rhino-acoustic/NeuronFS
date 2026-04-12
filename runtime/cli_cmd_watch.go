package main

import "fmt"

// WatchCmd implements the Command interface for the legacy --watch action.
// PROVIDES: Strangler Fig CLI migration for the "watch" mode.
type WatchCmd struct{}

func (c *WatchCmd) Name() string {
	return "--watch"
}

func (c *WatchCmd) Execute(brainRoot string, args []string) error {
	fmt.Println("[NeuronFS] Watch mode — monitoring brain/ for changes...")
	runWatch(brainRoot)
	return nil
}
