package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// RunCodeAuditV2 scans the root directory for .go files (excluding tests)
// and prints all function declarations found.
func RunCodeAuditV2(rootDir string) {
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-Go files
		if info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Skip test files
		if strings.HasSuffix(path, "_test.go") {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			fmt.Printf("Error opening file %s: %v\n", path, err)
			return nil
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			trimmed := strings.TrimSpace(line)

			// Check if the line starts with "func "
			if strings.HasPrefix(trimmed, "func ") {
				fmt.Printf("File: %s | Function: %s\n", path, trimmed)
			}
		}

		if err := scanner.Err(); err != nil {
			fmt.Printf("Error scanning file %s: %v\n", path, err)
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error walking the path %s: %v\n", rootDir, err)
	}
}
