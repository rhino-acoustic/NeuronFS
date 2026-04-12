package main

import "fmt"

// APICmd implements the Command interface for the legacy --api action.
type APICmd struct{}

func (c *APICmd) Name() string {
	return "--api"
}

func (c *APICmd) Execute(brainRoot string, args []string) error {
	fmt.Printf("[NeuronFS] API mode starting on port :%d\n", APIPort)
	startAPI(brainRoot, APIPort)
	return nil
}
