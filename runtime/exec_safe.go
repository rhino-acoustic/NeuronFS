package main

import (
	"context"
	"os/exec"
	"time"
)

// ━━━ Exec Timeout Policy (SSOT) ━━━
// 모든 외부 프로세스 호출은 반드시 이 래퍼를 통해야 한다.
// exec.Command 직접 호출 금지 — supervisor.go의 장기 실행 자식 프로세스만 예외.

const (
	// ExecTimeoutGit: git 명령 (commit, add, log, diff 등)
	ExecTimeoutGit = 30 * time.Second

	// ExecTimeoutSync: robocopy/rsync 동기화 (대용량 파일)
	ExecTimeoutSync = 5 * time.Minute

	// ExecTimeoutShell: powershell, tasklist 등 시스템 명령
	ExecTimeoutShell = 10 * time.Second

	// ExecTimeoutQuery: 상태 조회 (git status --porcelain 등 빠른 응답 기대)
	ExecTimeoutQuery = 5 * time.Second
)

// SafeExec runs a command with a strict timeout to prevent infinite deadlocks.
func SafeExec(timeout time.Duration, name string, arg ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, arg...)
	return cmd.Run()
}

// SafeOutput runs a command and returns its standard output, with a timeout.
func SafeOutput(timeout time.Duration, name string, arg ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, arg...)
	return cmd.Output()
}

// SafeCombinedOutput runs a command and returns its combined standard output and error, with a timeout.
func SafeCombinedOutput(timeout time.Duration, name string, arg ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, arg...)
	return cmd.CombinedOutput()
}

// SafeExecDir runs a command in a specific directory with a timeout.
func SafeExecDir(timeout time.Duration, dir string, name string, arg ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, arg...)
	cmd.Dir = dir
	return cmd.Run()
}

// SafeOutputDir runs a command in a specific directory and returns stdout, with a timeout.
func SafeOutputDir(timeout time.Duration, dir string, name string, arg ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, arg...)
	cmd.Dir = dir
	return cmd.Output()
}

// SafeCombinedOutputDir runs a command in a specific directory and returns combined output, with a timeout.
func SafeCombinedOutputDir(timeout time.Duration, dir string, name string, arg ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, arg...)
	cmd.Dir = dir
	return cmd.CombinedOutput()
}
