// backlog.go — inbox 자동 분류 파이프라인
// 텔레그램 inbox 파일을 자동 분류하여 적절한 위치로 이동
// SSOT: governance_consts.go의 DirAgents, DirInbox 참조
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"
)

// BacklogClassification 분류 결과
type BacklogClassification int

const (
	ClassArchive BacklogClassification = iota // 중복/핑/짧은 반복
	ClassBacklog                              // 실제 요구사항 → 처리 대기
	ClassCommand                              // / 명령어 → 즉시 실행
)

// classifyInboxMessage 메시지 내용 기반 자동 분류
func classifyInboxMessage(content string) BacklogClassification {
	// 헤더 제거
	body := content
	if idx := strings.Index(content, "\n\n"); idx >= 0 {
		body = content[idx+2:]
	}
	body = strings.TrimSpace(body)

	// 명령어
	if strings.HasPrefix(body, "/") {
		return ClassCommand
	}

	// 핑/짧은 반복 (5글자 미만)
	if utf8.RuneCountInString(body) < 5 {
		return ClassArchive
	}

	// 물음표만
	if strings.Trim(body, "?？") == "" {
		return ClassArchive
	}

	// 마스터프롬프트 반복
	if strings.Contains(body, "마스터 프롬프트") || strings.Contains(body, "NeuronFS 자율 진화 명령") {
		return ClassArchive
	}

	// 나머지 → backlog
	return ClassBacklog
}

// RunBacklogClassifier inbox 파일을 분류하여 이동
// 호출: supervisor에서 주기적으로 실행
func RunBacklogClassifier(brainRoot string, logger func(string)) {
	inboxDir := filepath.Join(brainRoot, DirAgents, "NeuronFS", "inbox")
	archiveDir := filepath.Join(inboxDir, "_archive")
	backlogDir := filepath.Join(inboxDir, "_backlog")

	os.MkdirAll(archiveDir, 0750)
	os.MkdirAll(backlogDir, 0750)

	entries, err := os.ReadDir(inboxDir)
	if err != nil {
		return
	}

	var archiveCount, backlogCount, commandCount int

	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".md") || strings.HasPrefix(name, "_") {
			continue
		}

		fp := filepath.Join(inboxDir, name)
		data, err := os.ReadFile(fp)
		if err != nil {
			continue
		}

		class := classifyInboxMessage(string(data))
		switch class {
		case ClassArchive:
			os.Rename(fp, filepath.Join(archiveDir, name))
			archiveCount++
		case ClassBacklog:
			os.Rename(fp, filepath.Join(backlogDir, name))
			backlogCount++
		case ClassCommand:
			// 명령어는 현재 inbox에 유지 (agent_bridge가 처리)
			commandCount++
		}
	}

	if archiveCount+backlogCount+commandCount > 0 && logger != nil {
		logger(fmt.Sprintf("📋 backlog 분류: archive=%d backlog=%d cmd=%d",
			archiveCount, backlogCount, commandCount))
	}
}

// backlogStats inbox/backlog/archive 파일 수 통계
func backlogStats(brainRoot string) (pending, backlog, archive int) {
	inboxDir := filepath.Join(brainRoot, DirAgents, "NeuronFS", "inbox")
	if entries, err := os.ReadDir(inboxDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") && !strings.HasPrefix(e.Name(), "_") {
				pending++
			}
		}
	}
	if entries, err := os.ReadDir(filepath.Join(inboxDir, "_backlog")); err == nil {
		for _, e := range entries {
			if !e.IsDir() {
				backlog++
			}
		}
	}
	if entries, err := os.ReadDir(filepath.Join(inboxDir, "_archive")); err == nil {
		for _, e := range entries {
			if !e.IsDir() {
				archive++
			}
		}
	}
	return
}

// RunBacklogLoop — supervisor에서 호출, 5분마다 분류 실행
func RunBacklogLoop(brainRoot string, logger func(string)) {
	for {
		RunBacklogClassifier(brainRoot, logger)
		time.Sleep(5 * time.Minute)
	}
}
