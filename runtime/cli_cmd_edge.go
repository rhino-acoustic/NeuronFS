package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// EdgeFixCmd implements a CLI command to read the latest bot1 inbox error
// and pipe it to the Two-Track hybrid LLM engine for an optimal fix.
type EdgeFixCmd struct{}

func (c *EdgeFixCmd) Name() string {
	return "--edge-fix"
}

func (c *EdgeFixCmd) Execute(brainRoot string, args []string) error {
	inboxDir := filepath.Join(brainRoot, "_agents", "bot1", "inbox")
	files, err := os.ReadDir(inboxDir)
	if err != nil {
		return fmt.Errorf("no bot1 inbox found: %w", err)
	}

	var latest string
	var latestTime int64
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".md") && !strings.HasPrefix(f.Name(), "_") {
			info, err := f.Info()
			if err == nil && info.ModTime().UnixNano() > latestTime {
				latestTime = info.ModTime().UnixNano()
				latest = f.Name()
			}
		}
	}

	if latest == "" {
		fmt.Println("[Edge] No errors in inbox. System stable.")
		return nil
	}

	content, err := os.ReadFile(filepath.Join(inboxDir, latest))
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", latest, err)
	}

	fmt.Printf("[Edge] Targeting inbox item: %s\n", latest)
	fmt.Println("[Hybrid] Waking up Cloud and Edge LLMs concurrently...")

	prompt := fmt.Sprintf("You are an expert Next.js developer. Please provide a brief fix for the following Turbopack error:\n\n%s", string(content))
	
	resp, track, err := InvokeTwoTrack("llama3", prompt)
	if err != nil {
		fmt.Printf("[Hybrid] System completely unreachable (%v)\n", err)
		return nil
	}

	fmt.Printf("\n=== Hybrid Inference Result (Winner: %s) ===\n", track)
	fmt.Println(resp)
	fmt.Println("=============================")

	return nil
}
