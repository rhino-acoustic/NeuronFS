package main

import "fmt"

// HarnessCmd executes the diagnostics test harness.
type HarnessCmd struct{}

func (c *HarnessCmd) Name() string {
	return "--harness"
}

func (c *HarnessCmd) Execute(brainRoot string, args []string) error {
	RunHarness(brainRoot, func(msg string) { fmt.Println(msg) })
	return nil
}
