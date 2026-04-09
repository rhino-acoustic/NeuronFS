// api_handlers.go — REST API 핸들러 (CRUD, Config, System)
//
// PROVIDES: registerCRUDRoutes, registerConfigRoutes, registerSystemRoutes, rollbackAll
// DEPENDS:  neuron_crud.go (growNeuron, fireNeuron, signalNeuron, rollbackNeuron)
//           lifecycle.go (runDecay, deduplicateNeurons)
//           brain.go (scanBrain, runSubsumption)
//           inject.go (autoReinject)
//           security/ (BuildChain, VerifyChain, loadOrCreateHMACKey)

package main

import (
	"fmt"
	"os/exec"
	"time"
)

var startTime = time.Now()

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// CRUD Routes: grow, fire, signal, decay, state
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

func rollbackAll(brainRoot string) error {
	cmd := exec.Command("git", "reset", "--hard", "HEAD~1")
	cmd.Dir = brainRoot
	_, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("\033[33m[WARNING] Hard reset failed! Initiating Quarantine Protocol...\033[0m\n")
		qBranch := fmt.Sprintf("quarantine-%s", time.Now().Format("20060102-150405"))

		cmdBranch := exec.Command("git", "checkout", "-b", qBranch)
		cmdBranch.Dir = brainRoot
		cmdBranch.Run()

		cmdAdd := exec.Command("git", "add", ".")
		cmdAdd.Dir = brainRoot
		cmdAdd.Run()

		cmdCommit := exec.Command("git", "commit", "-m", "Auto-quarantine corrupted state")
		cmdCommit.Dir = brainRoot
		cmdCommit.Run()

		fmt.Printf("\033[35m[QUARANTINE] Corrupted state isolated to branch: %s\033[0m\n", qBranch)

		cmdCheckout := exec.Command("git", "checkout", "main")
		cmdCheckout.Dir = brainRoot
		cmdCheckout.Run()

		return fmt.Errorf("git reset failed, isolated to %s. err: %v", qBranch, err)
	}

	cmdClean := exec.Command("git", "clean", "-fd")
	cmdClean.Dir = brainRoot
	cmdClean.Run()

	return nil
}
