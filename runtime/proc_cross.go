package main

// ━━━ proc_cross.go ━━━
// Cross-platform process management helpers.
// Abstracts away taskkill (Windows) vs pkill (Unix)
// to eliminate platform-specific hardcoding.

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// killProcessByName kills a process by its image name.
// Returns true if the process was found and killed.
// Windows: tasklist + taskkill /F /IM
// Unix:    pgrep + pkill -f
func killProcessByName(name string) bool {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("tasklist", "/FI", fmt.Sprintf("IMAGENAME eq %s", name), "/NH")
		out, err := cmd.CombinedOutput()
		if err != nil || !strings.Contains(string(out), name) {
			return false
		}
		exec.Command("taskkill", "/F", "/IM", name).Run()
		return true

	default: // linux, darwin
		cmd := exec.Command("pgrep", "-f", name)
		if err := cmd.Run(); err != nil {
			return false
		}
		exec.Command("pkill", "-f", name).Run()
		return true
	}
}
