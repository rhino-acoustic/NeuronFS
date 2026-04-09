package main

import (
	"context"
	"os/exec"
	"time"
)

// SafeExec runs a command with a strict timeout to prevent infinite deadlocks.
func SafeExec(timeout time.Duration, name string, arg ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, arg...)
	return cmd.Run()
}

// SafeOutput runs a command and returns its standard output, with a timeout.
func SafeOutput(timeout time.Duration, name string, arg ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, arg...)
	return cmd.Output()
}

// SafeCombinedOutput runs a command and returns its combined standard output and error, with a timeout.
func SafeCombinedOutput(timeout time.Duration, name string, arg ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, arg...)
	return cmd.CombinedOutput()
}
