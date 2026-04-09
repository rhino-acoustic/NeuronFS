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
	"time"
)

var startTime = time.Now()

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// CRUD Routes: grow, fire, signal, decay, state
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

func rollbackAll(brainRoot string) error {
	_, err := SafeCombinedOutputDir(ExecTimeoutGit, brainRoot, "git", "reset", "--hard", "HEAD~1")
	if err != nil {
		fmt.Printf("\033[33m[WARNING] Hard reset failed! Initiating Quarantine Protocol...\033[0m\n")
		qBranch := fmt.Sprintf("quarantine-%s", time.Now().Format("20060102-150405"))

		SafeExecDir(ExecTimeoutGit, brainRoot, "git", "checkout", "-b", qBranch)
		SafeExecDir(ExecTimeoutGit, brainRoot, "git", "add", ".")
		SafeExecDir(ExecTimeoutGit, brainRoot, "git", "commit", "-m", "Auto-quarantine corrupted state")

		fmt.Printf("\033[35m[QUARANTINE] Corrupted state isolated to branch: %s\033[0m\n", qBranch)

		SafeExecDir(ExecTimeoutGit, brainRoot, "git", "checkout", "main")

		return fmt.Errorf("git reset failed, isolated to %s. err: %v", qBranch, err)
	}

	SafeExecDir(ExecTimeoutGit, brainRoot, "git", "clean", "-fd")

	return nil
}

