package main

// ━━━ neuron_crud.go ━━━
// PROVIDES: growNeuron, fireNeuron, rollbackNeuron, signalNeuron
// DEPENDS ON: brain.go, similarity.go, lifecycle.go

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

var (
	neuronLocks = make(map[string]*sync.Mutex)
	nLockMutex  sync.Mutex
)

func lockNeuronPath(path string) *sync.Mutex {
	nLockMutex.Lock()
	defer nLockMutex.Unlock()
	if _, ok := neuronLocks[path]; !ok {
		neuronLocks[path] = &sync.Mutex{}
	}
	return neuronLocks[path]
}

// growNeuron creates a new neuron folder with 1.neuron
// If a similar neuron already exists (hybrid similarity >= MergeThreshold), fire that instead (consolidation)
// Uses Cosine Bigram (60%) + Levenshtein (40%) instead of legacy Jaccard
// Prefix-aware: 禁X and 推X are treated as DIFFERENT neurons (polarity matters)
// Returns error instead of os.Exit so REST API won't crash
func growNeuron(brainRoot string, neuronPath string) error {
	neuronPath = strings.ReplaceAll(neuronPath, "/", string(filepath.Separator))
	mu := lockNeuronPath(neuronPath)
	mu.Lock()
	defer mu.Unlock()

	// ── Hanja prefix normalization: 禁하드코딩 → 禁/하드코딩 ──
	neuronPath = normalizeHanjaPath(neuronPath)

	// ── Guard: block unlimited session_log growth ──
	if strings.Contains(neuronPath, "session_log") {
		slDir := filepath.Join(brainRoot, filepath.Dir(neuronPath), "session_log")
		if info, err := vfsStat(slDir); err == nil && info.IsDir() {
			existing, _ := vfsGlob(filepath.Join(slDir, "*.neuron"))
			if len(existing) >= SessionLogCap {
				fmt.Printf("[SKIP] session_log capped at 3 (has %d)\n", len(existing))
				return nil
			}
		}
	}

	fullPath := filepath.Join(brainRoot, neuronPath)

	// Validate region
	parts := strings.SplitN(neuronPath, string(filepath.Separator), 2)
	region := parts[0]
	if _, ok := regionPriority[region]; !ok {
		err := fmt.Errorf("invalid region: %s (valid: brainstem,limbic,hippocampus,sensors,cortex,ego,prefrontal)", region)
		fmt.Printf("[FATAL] %v\n", err)
		return err
	}

	// Check if neuron already exists (exact match in either physical or virtual layer)
	if info, err := vfsStat(fullPath); err == nil && info.IsDir() {
		fmt.Printf("[SKIP] Neuron already exists: %s\n", neuronPath)
		return nil
	}

	// ── Synaptic Consolidation: merge similar neurons (GLOBAL scan) ──
	// Tokenize the new neuron's leaf name
	leafName := filepath.Base(neuronPath)
	newTokens := tokenize(leafName)
	if len(newTokens) > 0 {
		// Walk ALL regions to find similar neurons (cross-region dedup)
		bestMatch := ""
		bestSimilarity := 0.0

		for _, scanRegion := range []string{"brainstem", "cortex", "ego", "prefrontal", "hippocampus", "sensors", "limbic"} {
			scanPath := filepath.Join(brainRoot, scanRegion)
			if _, err := vfsStat(scanPath); os.IsNotExist(err) {
				continue
			}
			vfsWalkDir(scanPath, func(path string, d fs.DirEntry, err error) error {
				if err != nil || !d.IsDir() || path == scanPath {
					return nil
				}
				neuronFiles, _ := vfsGlob(filepath.Join(path, "*.neuron"))
				if len(neuronFiles) == 0 {
					return nil
				}
				existingLeaf := filepath.Base(path)
				// Prefix-aware: 접두어가 다르면 별개로 취급
				newPrefix := extractPrefix(leafName)
				existPrefix := extractPrefix(existingLeaf)
				if newPrefix != "" && existPrefix != "" && newPrefix != existPrefix {
					return nil // 禁X vs 推X → 별개
				}
				existingTokens := tokenize(existingLeaf)
				sim := hybridSimilarity(newTokens, existingTokens)
				if sim > bestSimilarity {
					bestSimilarity = sim
					rel, _ := filepath.Rel(brainRoot, path)
					bestMatch = rel
				}
				return nil
			})
		}

		if bestSimilarity >= MergeThreshold && bestMatch != "" {
			fmt.Printf("[MERGE] 🔗 '%s' ≈ '%s' (%.0f%% similar) → firing existing\n",
				neuronPath, bestMatch, bestSimilarity*100)
			fireNeuron(brainRoot, bestMatch)
			logEpisode(brainRoot, "MERGE", fmt.Sprintf("%s → %s (%.0f%%)", neuronPath, bestMatch, bestSimilarity*100))
			return nil
		}
	}

	// Create folder
	if err := os.MkdirAll(fullPath, 0750); err != nil {
		fmt.Printf("[ERROR] mkdir: %v\n", err)
		return err
	}

	// Create 1.neuron
	neuronFile := filepath.Join(fullPath, "1.neuron")
	if err := os.WriteFile(neuronFile, []byte{}, 0600); err != nil {
		fmt.Printf("[ERROR] create trace: %v\n", err)
		return err
	}

	fmt.Printf("[GROW] ✅ %s → 1.neuron\n", neuronPath)

	// Log to hippocampus
	logEpisode(brainRoot, "GROW", neuronPath)

	// Mark dirty — periodic loop will handle injection
	markBrainDirty()
	return nil
}

// fireNeuron increments the counter of an existing neuron
// Usage: neuronfs brain_v4 --fire cortex/frontend/coding/no_console_log
func fireNeuron(brainRoot string, neuronPath string) {
	neuronPath = strings.ReplaceAll(neuronPath, "/", string(filepath.Separator))
	
	mu := lockNeuronPath(neuronPath)
	mu.Lock()
	defer mu.Unlock()

	fullPath := filepath.Join(brainRoot, neuronPath)

	if _, err := vfsStat(fullPath); os.IsNotExist(err) {
		fmt.Printf("[WARN] Neuron not found: %s — (Auto-grow disabled by Neuro-Lifecycle architecture)\n", neuronPath)
		return
	}

	// Find current counter (reads from virtual OS spanning Cartridge + Disk)
	currentCounter := 0
	currentFile := ""
	entries, _ := vfsReadDir(fullPath)
	for _, e := range entries {
		if m := counterRegex.FindStringSubmatch(e.Name()); m != nil {
			n, _ := strconv.Atoi(m[1])
			if n > currentCounter {
				currentCounter = n
				currentFile = filepath.Join(fullPath, e.Name())
			}
		}
	}

	newCounter := currentCounter + 1

	// Delete old counter file (Only if it exists in Physical UpperDir)
	if currentFile != "" {
		if _, physicalErr := os.Stat(currentFile); physicalErr == nil {
			if err := os.Remove(currentFile); err != nil {
				fmt.Fprintf(os.Stderr, "[WARN] fire: old counter cleanup: %v\n", err)
			}
		}
	}

	// Ensure physical folder exists before writing to UpperDir
	os.MkdirAll(fullPath, 0750)

	// Create new counter file
	newFile := filepath.Join(fullPath, fmt.Sprintf("%d.neuron", newCounter))
	if err := os.WriteFile(newFile, []byte{}, 0600); err != nil {
		fmt.Printf("[ERROR] fire: %v\n", err)
		return
	}

	fmt.Printf("[FIRE] 🔥 %s → %d → %d\n", neuronPath, currentCounter, newCounter)

	logEpisode(brainRoot, "FIRE", fmt.Sprintf("%s (%d→%d)", neuronPath, currentCounter, newCounter))
	markBrainDirty()
}

// rollbackNeuron decrements the counter of an existing neuron
// Minimum counter is 1 (won't go below). Returns error for API usage.
// Usage: neuronfs brain_v4 --rollback cortex/frontend/coding/no_console_log
func rollbackNeuron(brainRoot string, neuronPath string) error {
	neuronPath = strings.Trim(strings.ReplaceAll(neuronPath, "\\", "/"), "/")
	
	mu := lockNeuronPath(neuronPath)
	mu.Lock()
	defer mu.Unlock()

	fullPath := filepath.Join(brainRoot, neuronPath)

	if _, err := vfsStat(fullPath); os.IsNotExist(err) {
		err := fmt.Errorf("neuron not found: %s", neuronPath)
		fmt.Printf("[ROLLBACK] ❌ %v\n", err)
		return err
	}

	// Find current counter from VFS
	currentCounter := 0
	currentFile := ""
	entries, _ := vfsReadDir(fullPath)
	for _, e := range entries {
		if m := counterRegex.FindStringSubmatch(e.Name()); m != nil {
			n, _ := strconv.Atoi(m[1])
			if n > currentCounter {
				currentCounter = n
				currentFile = filepath.Join(fullPath, e.Name())
			}
		}
	}

	if currentCounter <= 1 {
		fmt.Printf("[ROLLBACK] ⚠️ %s counter already at minimum (%d)\n", neuronPath, currentCounter)
		return fmt.Errorf("counter at minimum: %d", currentCounter)
	}

	newCounter := currentCounter - 1

	// Delete old counter file from UpperDir only
	if currentFile != "" {
		if _, physicalErr := os.Stat(currentFile); physicalErr == nil {
			if err := os.Remove(currentFile); err != nil {
				fmt.Fprintf(os.Stderr, "[WARN] rollback: old counter cleanup: %v\n", err)
			}
		}
	}

	os.MkdirAll(fullPath, 0750)

	// Create new counter file
	newFile := filepath.Join(fullPath, fmt.Sprintf("%d.neuron", newCounter))
	if err := os.WriteFile(newFile, []byte{}, 0600); err != nil {
		fmt.Printf("[ERROR] rollback: %v\n", err)
		return err
	}

	fmt.Printf("[ROLLBACK] ⏪ %s → %d → %d\n", neuronPath, currentCounter, newCounter)

	logEpisode(brainRoot, "ROLLBACK", fmt.Sprintf("%s (%d→%d)", neuronPath, currentCounter, newCounter))
	markBrainDirty()
	return nil
}

// signalNeuron adds dopamine/bomb/memory signal to a neuron
// Returns error instead of os.Exit so REST API won't crash
func signalNeuron(brainRoot string, neuronPath string, sigType string) error {
	neuronPath = strings.ReplaceAll(neuronPath, "/", string(filepath.Separator))
	
	mu := lockNeuronPath(neuronPath)
	mu.Lock()
	defer mu.Unlock()

	fullPath := filepath.Join(brainRoot, neuronPath)

	if _, err := vfsStat(fullPath); os.IsNotExist(err) {
		err := fmt.Errorf("neuron not found: %s", neuronPath)
		fmt.Printf("[FATAL] %v\n", err)
		return err
	}

	os.MkdirAll(fullPath, 0750)

	switch sigType {
	case "dopamine":
		nextDopa := 1
		entries, _ := vfsReadDir(fullPath)
		for _, e := range entries {
			if m := dopamineRegex.FindStringSubmatch(e.Name()); m != nil {
				n, _ := strconv.Atoi(m[1])
				if n >= nextDopa {
					nextDopa = n + 1
				}
			}
		}
		df := filepath.Join(fullPath, fmt.Sprintf("dopamine%d.neuron", nextDopa))
		os.WriteFile(df, []byte{}, 0600)
		fmt.Printf("[SIGNAL] 🟢 dopamine%d → %s\n", nextDopa, neuronPath)

	case "bomb":
		bf := filepath.Join(fullPath, "bomb.neuron")
		os.WriteFile(bf, []byte{}, 0600)
		fmt.Printf("[SIGNAL] 💣 BOMB → %s\n", neuronPath)

	case "memory":
		nextMem := 1
		memRegex := regexp.MustCompile(`^memory(\d+)\.neuron$`)
		entries, _ := vfsReadDir(fullPath)
		for _, e := range entries {
			if m := memRegex.FindStringSubmatch(e.Name()); m != nil {
				n, _ := strconv.Atoi(m[1])
				if n >= nextMem {
					nextMem = n + 1
				}
			}
		}
		mf := filepath.Join(fullPath, fmt.Sprintf("memory%d.neuron", nextMem))
		os.WriteFile(mf, []byte{}, 0600)
		fmt.Printf("[SIGNAL] 📝 memory%d → %s\n", nextMem, neuronPath)

	default:
		err := fmt.Errorf("unknown signal type: %s (use dopamine|bomb|memory)", sigType)
		fmt.Printf("[FATAL] %v\n", err)
		return err
	}

	logEpisode(brainRoot, "SIGNAL:"+sigType, neuronPath)
	markBrainDirty()
	return nil
}

// ━━━ Lifecycle (prune/decay/episode) → lifecycle.go ━━━
// MOVED: pruneWeakNeurons, runDecay, logEpisode, MaxEpisodes

// normalizeHanjaPath splits rune-prefixed names into parent/child structure.
// e.g. "cortex/dev/禁하드코딩" → "cortex/dev/禁/하드코딩"
// Rune opcodes are single-character structural containers (Path=Sentence).
// Rune list is defined in governance_consts.go (SSOT).

func normalizeHanjaPath(p string) string {
	sep := string(filepath.Separator)
	parts := strings.Split(p, sep)
	var result []string
	for _, part := range parts {
		normalized := false
		for _, h := range RuneKeys() {
			if strings.HasPrefix(part, h) && len(part) > len(h) {
				// Split: 禁하드코딩 → 禁, 하드코딩
				result = append(result, h, part[len(h):])
				normalized = true
				break
			}
		}
		if !normalized {
			result = append(result, part)
		}
	}
	return strings.Join(result, sep)
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
