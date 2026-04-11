package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// runGoSourceWatcher monitors the runtime directory for .go file changes and triggers an auto-build
// so that the next Hot-Swap CLI execution uses the newly built worker binary.
func runGoSourceWatcher() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Printf("\033[31m[WATCHER] Failed to initialize source watcher: %v\033[0m\n", err)
		return
	}
	defer watcher.Close()

	// Watch the current runtime directory
	cwd, _ := os.Getwd()
	if err := watcher.Add(cwd); err != nil {
		fmt.Printf("\033[31m[WATCHER] Failed to watch %s: %v\033[0m\n", cwd, err)
		return
	}

	fmt.Printf("\033[35m[HOT-RELOAD] Active: Source mutations under %s will trigger instant re-compilation.\033[0m\n", cwd)

	var debounceTimer *time.Timer
	var debounceMu sync.Mutex
	debounceMs := 1500 * time.Millisecond // wait a bit for sequential file saves

	triggerBuild := func() {
		fmt.Printf("\033[33m[BUILD] Source mutations detected. Initiating auto re-compilation...\033[0m\n")
		// Assume we are in runtime/ and want to build into ../dist/neuronfs/neuronfs.exe
		// Execute go build
		start := time.Now()
		cmd := exec.Command("go", "build", "-o", filepath.Join("..", "dist", "neuronfs", "neuronfs.exe"), ".")
		cmd.Dir = cwd
		out, err := cmd.CombinedOutput()
		elapsed := time.Since(start)

		if err != nil {
			fmt.Printf("\033[31m[BUILD-FAIL] Auto-compilation failed: %v\nOutput: %s\033[0m\n", err, string(out))
		} else {
			fmt.Printf("\033[32m[BUILD-OK] Worker Core Regenerated in %v. Next tool call will use new binary.\033[0m\n", elapsed)
		}
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			// Only care about .go files
			if !strings.HasSuffix(event.Name, ".go") {
				continue
			}

			// Ignore chmod events only
			if event.Op == fsnotify.Chmod {
				continue
			}

			// Debounce building
			debounceMu.Lock()
			if debounceTimer != nil {
				debounceTimer.Stop()
			}
			debounceTimer = time.AfterFunc(debounceMs, triggerBuild)
			debounceMu.Unlock()

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			fmt.Printf("\033[31m[WATCHER-ERR] %v\033[0m\n", err)
		}
	}
}
