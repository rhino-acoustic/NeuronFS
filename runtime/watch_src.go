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

// codemapCategory maps Go file basenames to _codemap categories.
// Falls back to "도구" for unknown files.
var codemapCategoryMap = map[string]string{
	"main": "부팅", "init": "부팅", "config": "부팅", "awakening": "부팅",
	"brain": "엔진", "neuron_crud": "엔진", "inject": "엔진", "similarity": "엔진",
	"emit_bootstrap": "엔진", "emit_format_rules": "엔진", "emit_helpers": "엔진",
	"emit_tiers": "엔진", "emit_inbox_data": "엔진", "governance_consts": "엔진",
	"lifecycle": "뉴런수명", "hebbian": "뉴런수명", "spaced_repetition": "뉴런수명",
	"vfs_core": "엔진", "vfs_ops": "엔진", "vfs_mount": "엔진", "vfs_ignition": "엔진",
	"supervisor": "인프라", "api_server": "인프라", "dashboard": "인프라",
	"telegram_bridge": "브릿지", "hijack_orchestrator": "브릿지", "transcript": "브릿지",
	"cdp_client": "브릿지", "cdp_monitor": "브릿지", "mcp_server": "엔진",
}

// syncCodemap scans runtime/*.go and ensures each has a _codemap neuron.
// New .go files get neurons auto-created. This is the real-time 1:1 mapping engine.
func syncCodemap(runtimeDir, brainRoot string) {
	codemapRoot := filepath.Join(brainRoot, "cortex", "dev", "_codemap")

	// Collect existing codemap neuron names (prefix before →)
	existing := map[string]bool{}
	filepath.Walk(codemapRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if strings.HasSuffix(info.Name(), ".neuron") {
			dirName := filepath.Base(filepath.Dir(path))
			prefix := strings.SplitN(dirName, "→", 2)[0]
			existing[prefix] = true
		}
		return nil
	})

	// Scan runtime/*.go
	entries, err := os.ReadDir(runtimeDir)
	if err != nil {
		return
	}

	created := 0
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		baseName := strings.TrimSuffix(e.Name(), ".go")
		if existing[baseName] {
			continue
		}

		// Determine category
		cat := "도구" // default
		if c, ok := codemapCategoryMap[baseName]; ok {
			cat = c
		}

		// Auto-grow neuron
		neuronPath := fmt.Sprintf("cortex/dev/_codemap/%s/%s→자동매핑", cat, baseName)
		growNeuron(brainRoot, neuronPath)
		created++
		fmt.Printf("\033[36m[CODEMAP] 자동 매핑: %s.go → %s/%s\033[0m\n", baseName, cat, baseName)
	}

	if created > 0 {
		fmt.Printf("\033[32m[CODEMAP] %d개 신규 Go 파일 → 뉴런 자동 생성 완료\033[0m\n", created)
	}
}

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

	// Determine brainRoot from cwd (runtime/ → ../brain_v4)
	brainRoot := filepath.Join(cwd, "..", "brain_v4")

	// Initial codemap sync on startup
	syncCodemap(cwd, brainRoot)

	var debounceTimer *time.Timer
	var debounceMu sync.Mutex
	debounceMs := 1500 * time.Millisecond // wait a bit for sequential file saves

	triggerBuild := func() {
		fmt.Printf("\033[33m[BUILD] Source mutations detected. Initiating auto re-compilation...\033[0m\n")
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

		// After successful build, sync codemap
		syncCodemap(cwd, brainRoot)
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
