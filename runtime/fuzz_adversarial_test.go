package main

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// FuzzNeuronCrud throws pure random byte strings and invalid paths at the engine
// to ensure growNeuron, fireNeuron, rollbackNeuron never panics, breaks the lock, or loops infinitely.
func FuzzNeuronCrud(f *testing.F) {
	// Add some seed inputs (both normal and edge cases)
	f.Add("cortex/test1", "dopamine")
	f.Add("ego/../../test", "bomb")
	f.Add("invalid_region/test", "memory")
	f.Add("cortex/禁금지사항/123", "dopamine")
	f.Add("prefrontal/\x00\xffnull", "invalid")

	// Create an isolated brain for fuzzing to not touch real system
	tempDir := filepath.Join(os.TempDir(), fmt.Sprintf("fuzz_brain_%d", rand.Int()))
	setupFuzzBrain(tempDir)
	defer os.RemoveAll(tempDir) // cleanup

	f.Fuzz(func(t *testing.T, randomPath string, sigType string) {
		// No matter what the fuzzer throws at us, these functions must NOT panic.
		// They can return an error, but panic = fail.
		
		// Randomly test Grow
		_ = growNeuron(tempDir, randomPath)

		// Randomly test Fire
		fireNeuron(tempDir, randomPath)

		// Randomly test Rollback
		_ = rollbackNeuron(tempDir, randomPath)

		// Randomly test Signal
		_ = signalNeuron(tempDir, randomPath, sigType)
		
		// If we reach here without panic, NeuronFS survived this iteration
	})
}

// setupFuzzBrain manually creates minimum structure
func setupFuzzBrain(dir string) {
	os.MkdirAll(dir, 0750)
	for _, r := range []string{"brainstem", "limbic", "hippocampus", "sensors", "cortex", "ego", "prefrontal"} {
		rp := filepath.Join(dir, r)
		os.MkdirAll(rp, 0750)
		os.WriteFile(filepath.Join(rp, "_rules.md"), []byte("# init"), 0600)
	}
}

// runSubsumptionFuzz throws random text files and fake Axons to ensure cycle breaker handles it
func FuzzSubsumptionGraph(f *testing.F) {
	f.Add([]byte("TARGET: cortex/does_not_exist"), []byte("gibberish"))
	f.Add([]byte("TARGET: cortex/cad1"), []byte(""))

	tempDir := filepath.Join(os.TempDir(), fmt.Sprintf("fuzz_brain_sub_%d", rand.Int()))
	setupFuzzBrain(tempDir)
	defer os.RemoveAll(tempDir) 

	f.Fuzz(func(t *testing.T, payload1 []byte, payload2 []byte) {
		// Randomly generate axon targeting loop
		// Let's write random payloads to axon files and force subsumption to process it
		f1 := filepath.Join(tempDir, "cortex", "a.axon")
		f2 := filepath.Join(tempDir, "cortex", "b.axon")
		os.WriteFile(f1, []byte(strings.ToValidUTF8(string(payload1), "")), 0600)
		os.WriteFile(f2, []byte(strings.ToValidUTF8(string(payload2), "")), 0600)

		brain := scanBrain(tempDir)
		_ = runSubsumption(brain)
		
		// If neither scanBrain nor runSubsumption panics from the random payload, we pass.
	})
}
