package main

import (
	"fmt"
	"os/exec"
	"path/filepath"
)

// VerifyBrainIntegrity simulates the "Verification" stage of the Phase 61 closed-loop auto-evolve.
// It executes the internal harness check to ensure no logical rules or axioms are broken.
func VerifyBrainIntegrity(brainRoot string) error {
	fmt.Printf("[VALIDATOR] Running integrity checks (Phase 61)...\n")

	// 1. Check basic model parsing and subsumption rule overlap in memory
	brain := scanBrain(brainRoot)
	res := runSubsumption(brain)
	if res.TotalNeurons == 0 {
		return fmt.Errorf("brain structure completely wiped out or unreachable")
	}

	// 2. Invoke the harness_check tool as a subprocess to perform Deep Axiom Validation
	//    The harness_check runs strict internal rules (e.g., P0/P1 constraints, orphan checks)
	nfsExePath, err := filepath.Abs(".") 
	if err != nil {
		nfsExePath = "."
	}
	
	// Ensure we use 'go run' within the harness_check dir or just compile and run it.
	// We'll use "go run" on harness_check directly for the validation loop.
	harnessCmd := exec.Command("go", "run", "./cmd/harness_check", brainRoot)
	harnessCmd.Dir = nfsExePath
	
	// Execute the harness and capture output
	out, err := harnessCmd.CombinedOutput()
	if err != nil {
		fmt.Printf("[VALIDATOR] Harness checks failed:\n%s\n", string(out))
		return fmt.Errorf("harness check failed (exit %v)", err)
	}

	// 3. (Optional) Run `go vet ./...` to ensure the generated codebase logic is intact?
	// For now, testing the AI's JSON rules is more critical than the generic Go rules in this context.
	
	fmt.Printf("[VALIDATOR] Integrity check PASSED.\n")
	return nil
}
