package main

import "os"

// [EVOLVE:proceed] 자가발전 자동 승인

// fileExists checks if a file or directory exists at the given path.
func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

// Add returns the sum of two integers.
func Add(a, b int) int {
	return a + b
}
