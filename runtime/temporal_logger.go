package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ============================================================================
// Module: Temporal Semantic Memory Logger (Phase 37)
// Captures time-series snapshots of .neuron files when mutated, providing a 4D
// memory timeline for the Cognitive OS.
// ============================================================================

// RecordTemporalSnapshot creates a historical delta packing when a file changes.
func RecordTemporalSnapshot(brainRoot string, targetFilePath string) {
	if brainRoot == "" || targetFilePath == "" {
		return
	}

	// Only track cortex memory (skipping agent inbox to avoid recursion loops)
	if !strings.Contains(targetFilePath, string(os.PathSeparator)+"cortex"+string(os.PathSeparator)) {
		return
	}

	ext := filepath.Ext(targetFilePath)
	if ext != ".neuron" && ext != ".md" {
		return
	}

	// Calculate hash of current state
	fileData, err := os.ReadFile(targetFilePath)
	if err != nil {
		return
	}
	hashBytes := sha256.Sum256(fileData)
	hashStr := hex.EncodeToString(hashBytes[:])[:8]

	targetFileName := filepath.Base(targetFilePath)

	timestamp := time.Now().UnixMilli()
	deltaName := fmt.Sprintf("%d_%s_%s.delta", timestamp, targetFileName, hashStr)

	temporalDir := filepath.Join(brainRoot, "hippocampus", "temporal_log")
	_ = os.MkdirAll(temporalDir, 0755)

	deltaPath := filepath.Join(temporalDir, deltaName)

	// In a real delta system we would diff, but since we cannot intercept pre-write,
	// we save the fully snapshotted text representing the "T+n" state of the timeline.
	header := fmt.Sprintf("--- TSM_Snapshot ---\nFile: %s\nTime: %d\nHash: %s\n--------------------\n", targetFilePath, timestamp, hashStr)
	
	f, err := os.Create(deltaPath)
	if err != nil {
		return
	}
	defer f.Close()

	_, _ = f.WriteString(header)
	_, _ = f.Write(fileData)

	fmt.Printf("[Temporal-4D] Snapshot packed: %s\n", deltaName)
}
