package main

import "fmt"

// APICmd implements the Command interface for the legacy --api action.
type APICmd struct{}

func (c *APICmd) Name() string {
	return "--api"
}

func (c *APICmd) Execute(brainRoot string, args []string) error {
	fmt.Printf("[NeuronFS] API mode starting on port :%d\n", APIPort)

	// Phase 25: Initialize Libp2p/mDNS
	go func() {
		if err := InitializeP2PNode(brainRoot); err != nil {
			fmt.Printf("[P2P] Initialization failed: %v\n", err)
		}
	}()

	startAPI(brainRoot, APIPort)
	return nil
}
