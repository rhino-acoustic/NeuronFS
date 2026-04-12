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
// Module: Safe Self-Repair Pipeline — Git-Isolated Auto-Fix (V12-C / Phase 64)
// When a repair proposal exists, this pipeline:
// 1. git stash (preserve current state)
// 1.5 Parse and run `git apply` for the patch inside proposal.
// 2. go vet + go build (verify)
// 3. Success → git commit (confirm)
// 4. Failure → git reset --hard && git clean && git stash pop (rollback)
// NEVER auto-applies without verification.
// ============================================================================

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

		// Only process pending
		if !strings.Contains(content, "status: pending") {
			continue
		}

		fmt.Printf("[Self-Repair] 제안 발견: %s\n", e.Name())

		// Step 1: Git stash (안전 격리용 베이스라인 설정)
		if !gitStash(runtimeDir) {
			updateProposalStatus(proposalPath, content, "rejected", "git stash 실패")
			continue
		}

		// Step 1.5: Patch(패치) 스크립트 발췌 및 적용
		// 만약 이 단계가 실패(applyError)하더라도, 컴파일 검증 전에 이미 거부 처리.
		patchApplied := applyPatchFromContent(runtimeDir, content)
		if !patchApplied {
			gitStashPop(runtimeDir)
			updateProposalStatus(proposalPath, content, "rejected", "파싱된 Patch 블록 적용(git apply) 실패 또는 발견되지 않음")
			fmt.Printf("[Self-Repair] ❌ 적용 실패, 롤백: %s\n", e.Name())
			RecordAudit(brainRoot, "self_repair", "rejected", proposalPath, "패치 적용 실패", false)
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
			RecordAudit(brainRoot, "self_repair", "applied", proposalPath, "수리 패치 및 검증 통과", true)
		} else {
			// Step 4: 실패 → 롤백 (Hard Reset 후 Pop)
			gitRollbackHard(runtimeDir)
			gitStashPop(runtimeDir)
			updateProposalStatus(proposalPath, content, "rejected", "go vet/build 실패")
			fmt.Printf("[Self-Repair] ❌ 수리 검증 실패, 강제 롤백: %s\n", e.Name())
			RecordAudit(brainRoot, "self_repair", "rejected", proposalPath, "빌드 검증 실패 → 롤백", false)
		}
	}
}

// ── Patch Helper ──

func applyPatchFromContent(dir, content string) bool {
	// Look for ```patch ... ``` or ```diff ... ```
	startIdx := strings.Index(content, "```patch\n")
	if startIdx == -1 {
		startIdx = strings.Index(content, "```diff\n")
		if startIdx == -1 {
			return false // No patch found
		}
		startIdx += 8
	} else {
		startIdx += 9
	}
	endIdx := strings.Index(content[startIdx:], "```")
	if endIdx == -1 {
		return false
	}
	patchStr := content[startIdx : startIdx+endIdx]

	patchFile := filepath.Join(dir, "temp_self_repair.patch")
	err := os.WriteFile(patchFile, []byte(patchStr), 0644)
	if err != nil {
		return false
	}
	defer os.Remove(patchFile)

	// Apply patch via git
	cmd := exec.Command("git", "apply", "--whitespace=nowarn", patchFile)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("[Git] apply patch 실패: %s\n", string(out))
		return false
	}
	
	fmt.Printf("[Git] patch 부분 변경 파이프라인 주입 완료.\n")
	return true
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
	return true
}

func gitRollbackHard(dir string) {
	exec.Command("git", "reset", "--hard").Run()
	exec.Command("git", "clean", "-fd").Run()
}

func gitStashPop(dir string) {
	cmd := exec.Command("git", "stash", "pop")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("[Git] stash pop (no stash to pop?): %s\n", string(out))
	}
}

func gitCommitRepair(dir, proposalName string) {
	addCmd := exec.Command("git", "add", "-A")
	addCmd.Dir = dir
	addCmd.Run()

	msg := fmt.Sprintf("AutoRepair: %s (Phase 64 자가 수리 적용 완료)", proposalName)
	commitCmd := exec.Command("git", "commit", "-m", msg)
	commitCmd.Dir = dir
	commitCmd.Run()
}

func runGoVet(dir string) bool {
	cmd := exec.Command("go", "vet", "./...")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("[GoVet] 검증 오작동 감지 롤백 트리거: %s\n", string(out))
		return false
	}
	return true
}

func runGoBuild(dir string) bool {
	cmd := exec.Command("go", "build", ".")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("[GoBuild] 컴파일 에러 감지 롤백 트리거: %s\n", string(out))
		return false
	}
	return true
}

func updateProposalStatus(path, content, newStatus, reason string) {
	updated := strings.Replace(content, "status: pending", fmt.Sprintf("status: %s\nresolution: %s\nresolved_at: %s", newStatus, reason, time.Now().Format(time.RFC3339)), 1)
	_ = os.WriteFile(path, []byte(updated), 0644)
}
