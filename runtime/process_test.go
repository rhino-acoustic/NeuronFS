package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHarness(t *testing.T) {
	// corrections.jsonl ?Œì„œ ê²€ì¦?ë°?processInbox ?ŒìŠ¤??ê¸°ë°˜
	tmpDir := t.TempDir()
	inboxDir := filepath.Join(tmpDir, "_inbox")
	os.MkdirAll(inboxDir, 0755)

	content := `{"ts":"1000","type":"correction","text":"PDì¹?°¬","path":"cortex/test/dopamine","counter_add":1,"author":"pm"}
{"ts":"1001","type":"correction","text":"?„ë²½?©ë‹ˆ??,"path":"cortex/test/fake","counter_add":1,"author":"enfp"}
{"ts":"1002","type":"correction","text":"normal rule","path":"cortex/test/normal","counter_add":1,"author":"bot1"}`

	os.WriteFile(filepath.Join(inboxDir, "corrections.jsonl"), []byte(content), 0644)

	processInbox(tmpDir)

	// 1. PM ì¹?°¬?€ ?„íŒŒë¯?ë§ˆì»¤ ?ì„± ?±ê³µ?´ì•¼ ??	if _, err := os.Stat(filepath.Join(tmpDir, "cortex", "test", "dopamine", "dopamine1.neuron")); os.IsNotExist(err) {
		t.Errorf("TestHarness FAILED: PM praise did not create dopamine marker")
	}

	// 2. ë´?ê°?ê¸°ê³„??ì¹?°¬?€ ?„íŒŒë¯?ë§ˆì»¤ë¥??ì„±?˜ì? ?Šì•„????	if _, err := os.Stat(filepath.Join(tmpDir, "cortex", "test", "fake", "dopamine1.neuron")); err == nil {
		t.Errorf("TestHarness FAILED: Fake praise inflated dopamine")
	}

	// 3. ?•ìƒ ê·œì¹™?€ grow ?ë™ ?¤í–‰ ?•ìƒ ?™ì‘?´ì•¼ ??	if _, err := os.Stat(filepath.Join(tmpDir, "cortex", "test", "normal", "1.neuron")); os.IsNotExist(err) {
		t.Errorf("TestHarness FAILED: Normal neuron failed to grow")
	}
}

