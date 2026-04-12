package main

import "fmt"

// DashboardV2Cmd implements the Command interface for the Dashboard V2.
type DashboardV2Cmd struct{}

func (c *DashboardV2Cmd) Name() string {
	return "--dashboard-v2"
}

func (c *DashboardV2Cmd) Execute(brainRoot string, args []string) error {
	fmt.Printf("[NeuronFS] Dashboard V2 starting on port :%d/v2\n", APIPort)
	startAPI(brainRoot, APIPort)
	return nil
}
