package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHarness(t *testing.T) {
	// corrections.jsonl 파서 검증 및 processInbox 테스트 기반
	tmpDir := t.TempDir()
	inboxDir := filepath.Join(tmpDir, "_inbox")
	os.MkdirAll(inboxDir, 0755)

	content := `{"ts":"1000","type":"correction","text":"PD칭찬","path":"cortex/test/dopamine","counter_add":1,"author":"pm"}
{"ts":"1001","type":"correction","text":"완벽합니다","path":"cortex/test/fake","counter_add":1,"author":"enfp"}
{"ts":"1002","type":"correction","text":"normal rule","path":"cortex/test/normal","counter_add":1,"author":"bot1"}`

	os.WriteFile(filepath.Join(inboxDir, "corrections.jsonl"), []byte(content), 0644)

	processInbox(tmpDir)

	// 1. PM 칭찬은 도파민 마커 생성 성공해야 함
	if _, err := os.Stat(filepath.Join(tmpDir, "cortex", "test", "dopamine", "dopamine1.neuron")); os.IsNotExist(err) {
		t.Errorf("TestHarness FAILED: PM praise did not create dopamine marker")
	}

	// 2. 봇 간 기계적 칭찬은 도파민 마커를 생성하지 않아야 함
	if _, err := os.Stat(filepath.Join(tmpDir, "cortex", "test", "fake", "dopamine1.neuron")); err == nil {
		t.Errorf("TestHarness FAILED: Fake praise inflated dopamine")
	}

	// 3. 정상 규칙은 grow 자동 실행 정상 동작해야 함
	if _, err := os.Stat(filepath.Join(tmpDir, "cortex", "test", "normal", "1.neuron")); os.IsNotExist(err) {
		t.Errorf("TestHarness FAILED: Normal neuron failed to grow")
	}
}
