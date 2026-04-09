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

	// _agents/bot1/inbox/ 생성
	bot1Inbox := filepath.Join(brainRoot, "_agents", "bot1", "inbox")
	os.MkdirAll(bot1Inbox, 0750)
	os.WriteFile(filepath.Join(bot1Inbox, "test_task.md"), []byte("# [요청] 빌드 검증\n\n**발신: FORGE (ENTP)**\n"), 0600)

	// _agents/enfp/inbox/ 생성
	enfpInbox := filepath.Join(brainRoot, "_agents", "enfp", "inbox")
	os.MkdirAll(enfpInbox, 0750)
	os.WriteFile(filepath.Join(enfpInbox, "review_req.md"), []byte("**발신: ANCHOR (bot1)**\n\n# 리뷰 요청\n"), 0600)
	os.WriteFile(filepath.Join(enfpInbox, "deck_req.md"), []byte("# Enterprise 세일즈 덱\n\n**발신: FORGE (ENTP)**\n"), 0600)

	result := emitAgentInbox(brainRoot)

	t.Logf("Result:\n%s", result)

	if result == "" {
		t.Fatal("emitAgentInbox returned empty string")
	}

	if !strings.Contains(result, "에이전트 수신함") {
		t.Error("missing header '에이전트 수신함'")
	}

	if !strings.Contains(result, "[bot1] inbox (1건)") {
		t.Error("missing bot1 inbox count")
	}

	if !strings.Contains(result, "[enfp] inbox (2건)") {
		t.Error("missing enfp inbox count")
	}

	if !strings.Contains(result, "빌드 검증") {
		t.Error("missing bot1 message preview")
	}
}

func TestExtractInboxPreview(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		filename string
		wantSub  string
	}{
		{
			name:     "with sender and title",
			content:  "# [요청] 빌드 검증\n\n**발신: FORGE (ENTP)**\n",
			filename: "test.md",
			wantSub:  "빌드 검증",
		},
		{
			name:     "sender before title",
			content:  "**발신: ANCHOR (bot1)**\n\n# 리뷰 요청\n",
			filename: "test.md",
			wantSub:  "ANCHOR",
		},
		{
			name:     "no title no sender",
			content:  "그냥 텍스트입니다\n",
			filename: "plain.md",
			wantSub:  "그냥 텍스트",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractInboxPreview(tt.content, tt.filename)
			if !strings.Contains(result, tt.wantSub) {
				t.Errorf("expected to contain %q, got %q", tt.wantSub, result)
			}
		})
	}
}
