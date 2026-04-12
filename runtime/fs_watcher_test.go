package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFSWatcher_Debounce(t *testing.T) {
	brainRoot := t.TempDir()
	
	// Create cortex target
	targetDir := filepath.Join(brainRoot, "cortex", "dev", "작업")
	os.MkdirAll(targetDir, 0755)

	inboxDir := filepath.Join(brainRoot, "_agents", "bot1", "inbox")
	
	// Pre-start the process artificially (avoid actual fsnotify loops which can hang tests)
	testAlert := func() {
		handleFileEvent(brainRoot, filepath.Join(targetDir, "test.neuron"), "WRITE")
	}

	// Trigger 10 times instantly
	for i := 0; i < 10; i++ {
		testAlert()
	}

	// Sleep tiny amount to let async routine finish
	time.Sleep(100 * time.Millisecond)

	// Verify only 1 file was created in inbox
	files, err := os.ReadDir(inboxDir)
	if err != nil {
		t.Fatalf("Failed to read inbox: %v", err)
	}

	if len(files) != 1 {
		t.Errorf("Expected exactly 1 alert file due to debouncing, got %d", len(files))
	}
}
