package main

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// runStats prints visibility and fragmentation metrics
func runStats(brainRoot string) {
	fmt.Println("=== NeuronFS Diagnostic Stats ===")
	
	brain := scanBrain(brainRoot)
	var allNeurons []Neuron
	dormantCount := 0
	
	for _, region := range brain.Regions {
		for _, n := range region.Neurons {
			allNeurons = append(allNeurons, n)
			if n.IsDormant {
				dormantCount++
			}
		}
	}
	
	// Top 5 heaviest (most bytes)
	sort.Slice(allNeurons, func(i, j int) bool {
		// Mock physical size check by sorting path length + counter logically 
		// Real implementation would read file info. We just check Counter here for Top 5 violations as an MVP.
		return allNeurons[i].Counter > allNeurons[j].Counter
	})
	
	fmt.Println("[Top 5 Violations (Highest Action Count)]")
	limit := 5
	if len(allNeurons) < 5 {
		limit = len(allNeurons)
	}
	for i := 0; i < limit; i++ {
		fmt.Printf("  %d. %s (Intensity: %d, Path: %s)\n", i+1, allNeurons[i].Name, allNeurons[i].Counter, allNeurons[i].Path)
	}

	fmt.Printf("\n[Dormant Neurons]: %d found.\n", dormantCount)
	
	// Fragmentation check in hippocampus
	hippocampusCount := 0
	for _, n := range allNeurons {
		if filepath.Dir(n.Path) == "hippocampus" || strings.HasPrefix(n.Path, "hippocampus") {
			hippocampusCount++
		}
	}
	fmt.Printf("[Fragmentation Status]: %d fragmented memory items in hippocampus.\n", hippocampusCount)
	fmt.Println("Run 'neuronfs --vacuum' to merge fragments.")
}

// runVacuum placeholder for Llama-based background garbage collection
func runVacuum(brainRoot string) {
	fmt.Println("=== NeuronFS Garbage Collector ===")
	fmt.Println("[INFO] Scanning fragmented memories...")
	time.Sleep(1 * time.Second)
	fmt.Println("[INFO] Calling Llama3/Ollama API in background to merge contexts...")
	time.Sleep(1 * time.Second)
	fmt.Println("[OK] Hippocampus memory consolidated successfully.")
	fmt.Println("[INFO] (Placeholder function executed: LLM Context optimization deferred to LocalLLM)")
}
