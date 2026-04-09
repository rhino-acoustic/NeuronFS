package main

import "os"

// fileExists checks if a file or directory exists at the given path.
func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}
