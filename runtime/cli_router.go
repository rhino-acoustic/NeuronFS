package main

// Command defines the interface for all modular NeuronFS CLI commands.
// This is the core of the Strangler Fig strategy, allowing progressive migration.
type Command interface {
	// Name returns the primary CLI argument for this command (e.g. "--harness")
	Name() string
	
	// Execute runs the command.
	// brainRoot is parsed natively before routing.
	// args includes all raw arguments, allowing commands to parse their own flags.
	Execute(brainRoot string, args []string) error
}

// Router manages the registration and execution of commands.
type Router struct {
	commands map[string]Command
}

// NewRouter creates a new CLI router.
func NewRouter() *Router {
	return &Router{
		commands: make(map[string]Command),
	}
}

// Register adds a command to the router's dictionary.
func (r *Router) Register(cmd Command) {
	r.commands[cmd.Name()] = cmd
}

// TryRoute attempts to route the given CLI argument to a registered command.
// Returning false means the command hasn't been migrated yet (legacy fallback).
func (r *Router) TryRoute(target string, brainRoot string, args []string) (bool, error) {
	if cmd, exists := r.commands[target]; exists {
		err := cmd.Execute(brainRoot, args)
		return true, err
	}
	return false, nil
}
