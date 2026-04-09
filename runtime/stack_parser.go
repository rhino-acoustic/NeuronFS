package main

import (
	"fmt"
	"regexp"
	"strings"
)

// appFileRegex matches non-runtime, non-stdlib source lines like:
// "\tC:/Dev/NeuronFS/main.go:52 +0x4c"
// "/home/user/NeuronFS/brain.go:120 "
var appFileRegex = regexp.MustCompile(`(?m)^\s*(.+\.go):(\d+)\s`)

func extractFaultNode(rawStack string) string {
	lines := strings.Split(rawStack, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip Go stdlib runtime frames (not app code in dirs named "runtime")
		if strings.Contains(trimmed, "/go/src/runtime/") ||
			strings.Contains(trimmed, "\\go\\src\\runtime\\") ||
			strings.Contains(trimmed, "src/runtime/") ||
			strings.Contains(trimmed, "src\\runtime\\") ||
			strings.Contains(trimmed, "testing/") ||
			strings.Contains(trimmed, "debug.Stack") ||
			strings.Contains(trimmed, "runtime.gopanic") ||
			strings.Contains(trimmed, "runtime.goexit") {
			continue
		}

		m := appFileRegex.FindStringSubmatch(trimmed)
		if m != nil {
			filePath := m[1]
			lineNum := m[2]

			// Extract just the filename (no full path noise)
			parts := strings.Split(strings.ReplaceAll(filePath, "\\", "/"), "/")
			shortFile := filePath
			if len(parts) >= 2 {
				shortFile = strings.Join(parts[len(parts)-2:], "/")
			}

			return fmt.Sprintf("%s:%s", shortFile, lineNum)
		}
	}
	return "unknown (stack unparseable)"
}
