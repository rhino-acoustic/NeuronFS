package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEmitAgentInbox(t *testing.T) {
	dir := t.TempDir()
	brainRoot := dir

	// _agents/bot1/inbox/ ?�성
	bot1Inbox := filepath.Join(brainRoot, "_agents", "bot1", "inbox")
	os.MkdirAll(bot1Inbox, 0755)
	os.WriteFile(filepath.Join(bot1Inbox, "test_task.md"), []byte("# [?�청] 빌드 검�?n\n**발신: FORGE (ENTP)**\n"), 0644)

	// _agents/enfp/inbox/ ?�성
	enfpInbox := filepath.Join(brainRoot, "_agents", "enfp", "inbox")
	os.MkdirAll(enfpInbox, 0755)
	os.WriteFile(filepath.Join(enfpInbox, "review_req.md"), []byte("**발신: ANCHOR (bot1)**\n\n# 리뷰 ?�청\n"), 0644)
	os.WriteFile(filepath.Join(enfpInbox, "deck_req.md"), []byte("# Enterprise ?�일�???n\n**발신: FORGE (ENTP)**\n"), 0644)

	result := emitAgentInbox(brainRoot)

	t.Logf("Result:\n%s", result)

	if result == "" {
		t.Fatal("emitAgentInbox returned empty string")
	}

	if !strings.Contains(result, "?�이?�트 ?�신??(볼륨 ?�인??") {
		t.Error("missing header '?�이?�트 ?�신??(볼륨 ?�인??'")
	}

	if !strings.Contains(result, "[bot1]** 미확??메시지: 1�?) {
		t.Error("missing bot1 inbox count")
	}

	if !strings.Contains(result, "[enfp]** 미확??메시지: 2�?) {
		t.Error("missing enfp inbox count")
	}
}

