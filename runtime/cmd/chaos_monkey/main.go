package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

// Chaos Monkey: Run alongside NeuronFS to randomly disrupt the VFS layer.
func main() {
	var targetDir string
	var seed int64
	var mode string
	var duration int

	flag.StringVar(&targetDir, "dir", "", "Target brain directory to attack")
	flag.Int64Var(&seed, "seed", 0, "RNG seed (0 for random)")
	flag.StringVar(&mode, "mode", "random", "Attack mode: delete, spam, random")
	flag.IntVar(&duration, "duration", 10, "Attack duration in seconds")
	flag.Parse()

	if targetDir == "" {
		fmt.Println("Usage: chaos_monkey --dir <path> [--seed 123] [--mode random] [--duration 10]")
		os.Exit(1)
	}

	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	rand.Seed(seed)

	fmt.Printf("[CHAOS MONKEY] Starting attack on %s\n", targetDir)
	fmt.Printf("[CHAOS MONKEY] Seed: %d | Mode: %s | Duration: %ds\n", seed, mode, duration)

	startTime := time.Now()
	attackCount := 0

	for time.Since(startTime).Seconds() < float64(duration) {
		time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond) // Random jitter

		m := mode
		if mode == "random" {
			switch rand.Intn(2) {
			case 0:
				m = "delete"
			case 1:
				m = "spam"
			}
		}

		switch m {
		case "delete":
			// Randomly delete files
			regions := []string{"brainstem", "limbic", "hippocampus", "sensors", "cortex", "ego", "prefrontal"}
			r := regions[rand.Intn(len(regions))]
			rp := filepath.Join(targetDir, r)

			entries, err := os.ReadDir(rp)
			if err == nil && len(entries) > 0 {
				e := entries[rand.Intn(len(entries))]
				if e.IsDir() {
					// Delete arbitrary neuron file inside!
					neuronPath := filepath.Join(rp, e.Name())
					nfiles, _ := os.ReadDir(neuronPath)
					if len(nfiles) > 0 {
						target := filepath.Join(neuronPath, nfiles[rand.Intn(len(nfiles))].Name())
						os.Remove(target)
						fmt.Printf("💥 [ATTACK: Delete] %s\n", target)
						attackCount++
					}
				}
			}

		case "spam":
			// Spam garbage files
			r := []string{"brainstem", "limbic", "hippocampus", "sensors", "cortex", "ego", "prefrontal"}[rand.Intn(7)]
			garbagePath := filepath.Join(targetDir, r, fmt.Sprintf("garbage_%d", rand.Intn(9999)))
			os.MkdirAll(garbagePath, 0750)
			os.WriteFile(filepath.Join(garbagePath, "999.neuron"), []byte("garbage"), 0600)
			fmt.Printf("🗑️ [ATTACK: Spam] %s/999.neuron\n", garbagePath)
			attackCount++
		}
	}

	fmt.Printf("\n[CHAOS MONKEY] Attack complete! Total disruptive events: %d\n", attackCount)
	fmt.Println("Now Run NeuronFS and verify it didn't crash or hang!")
}
