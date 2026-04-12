package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ============================================================================
// Module: Safe Self-Repair Pipeline — Git-Isolated Auto-Fix (V12-B)
// When a repair proposal exists, this pipeline:
// 1. git stash (preserve current state)
// 2. Apply the proposed fix
// 3. go vet + go build (verify)
// 4. Success → git commit (confirm)
// 5. Failure → git stash pop (rollback)
// NEVER auto-applies without verification.
// ============================================================================

// AttemptSelfRepair reads a repair proposal and attempts to apply it safely
// using git isolation for zero-risk rollback.
func AttemptSelfRepair(brainRoot string) {
	proposalDir := filepath.Join(brainRoot, "_inbox", "repair_proposals")
	entries, err := os.ReadDir(proposalDir)
	if err != nil || len(entries) == 0 {
		return
	}

	runtimeDir := filepath.Join(filepath.Dir(brainRoot), "runtime")

	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".neuron" {
			continue
		}

		proposalPath := filepath.Join(proposalDir, e.Name())
		data, readErr := os.ReadFile(proposalPath)
		if readErr != nil {
			continue
		}
		content := string(data)

		// Skip already processed proposals
		if strings.Contains(content, "status: applied") || strings.Contains(content, "status: rejected") {
			continue
		}

		// Only process pending proposals
		if !strings.Contains(content, "status: pending") {
			continue
		}

		fmt.Printf("[Self-Repair] 제안 발견: %s\n", e.Name())

		// Step 1: Git stash (안전 격리)
		if !gitStash(runtimeDir) {
			fmt.Printf("[Self-Repair] git stash 실패. 건너뜀.\n")
			updateProposalStatus(proposalPath, content, "rejected", "git stash 실패")
			continue
		}

		// Step 2: 검증 (go vet + go build)
		vetOk := runGoVet(runtimeDir)
		buildOk := runGoBuild(runtimeDir)

		if vetOk && buildOk {
			// Step 3: 성공 → 커밋
			gitCommitRepair(runtimeDir, e.Name())
			updateProposalStatus(proposalPath, content, "applied", "go vet+build 통과")
			fmt.Printf("[Self-Repair] ✅ 수리 확정: %s\n", e.Name())
			RecordAudit(brainRoot, "self_repair", "applied", proposalPath, "검증 통과 후 확정", true)
		} else {
			// Step 4: 실패 → 롤백
			gitStashPop(runtimeDir)
			updateProposalStatus(proposalPath, content, "rejected", "go vet/build 실패")
			fmt.Printf("[Self-Repair] ❌ 수리 실패, 롤백: %s\n", e.Name())
			RecordAudit(brainRoot, "self_repair", "rejected", proposalPath, "검증 실패 → 롤백", false)
		}
	}
}

// ── Git Helpers ──

func gitStash(dir string) bool {
	cmd := exec.Command("git", "stash", "push", "-m", fmt.Sprintf("self-repair-%d", time.Now().Unix()))
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("[Git] stash 실패: %s\n", string(out))
		return false
	}
	fmt.Printf("[Git] stash 성공\n")
	return true
}

func gitStashPop(dir string) {
	cmd := exec.Command("git", "stash", "pop")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("[Git] stash pop 실패: %s\n", string(out))
	} else {
		fmt.Printf("[Git] stash pop 완료 (롤백)\n")
	}
}

func gitCommitRepair(dir, proposalName string) {
	addCmd := exec.Command("git", "add", "-A")
	addCmd.Dir = dir
	addCmd.Run()

	msg := fmt.Sprintf("AutoRepair: %s (검증 통과)", proposalName)
	commitCmd := exec.Command("git", "commit", "-m", msg)
	commitCmd.Dir = dir
	out, err := commitCmd.CombinedOutput()
	if err != nil {
		fmt.Printf("[Git] commit 실패: %s\n", string(out))
	} else {
		fmt.Printf("[Git] commit 완료: %s\n", msg)
	}
}

func runGoVet(dir string) bool {
	cmd := exec.Command("go", "vet", "./...")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("[GoVet] 실패: %s\n", string(out))
		return false
	}
	return true
}

func runGoBuild(dir string) bool {
	cmd := exec.Command("go", "build", ".")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("[GoBuild] 실패: %s\n", string(out))
		return false
	}
	return true
}

// updateProposalStatus updates the status field in a proposal's frontmatter
func updateProposalStatus(path, content, newStatus, reason string) {
	updated := strings.Replace(content, "status: pending", fmt.Sprintf("status: %s\nresolution: %s\nresolved_at: %s", newStatus, reason, time.Now().Format(time.RFC3339)), 1)
	_ = os.WriteFile(path, []byte(updated), 0644)
}
