package main

import (
	"os"
	"testing"
)

// TestMain — 모든 테스트 실행 전 NEURONFS_TEST_ISOLATION 설정.
// injectToGemini가 실제 GEMINI.md를 건드리지 않도록 근본 차단.
func TestMain(m *testing.M) {
	// Virtual File System 초기 상태 주입 (Jloot 우회)
	initVFS(".")

	os.Setenv("NEURONFS_TEST_ISOLATION", "1")
	defer os.Unsetenv("NEURONFS_TEST_ISOLATION")
	os.Exit(m.Run())
}
