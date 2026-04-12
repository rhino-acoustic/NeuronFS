package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ============================================================================
// Module: Locale Configuration — V13 Phase 49
// Reads language setting from brainstem/config.neuron
// Controls system message language for Telegram, logs, and dashboard.
// ============================================================================

// Locale holds the current language configuration
type Locale struct {
	Lang     string            // "ko", "en", "ja"
	Messages map[string]string // key -> localized message
}

var localeMessages = map[string]map[string]string{
	"ko": {
		"scan_start":     "뇌 스캔 시작",
		"scan_complete":  "스캔 완료: %d개 뉴런",
		"emit_start":     "규칙 생성 시작",
		"emit_complete":  "규칙 생성 완료",
		"error_detected": "에러 감지: %s",
		"error_escalated": "에러 에스컬레이션: %s → P0",
		"bomb_triggered": "BOMB 발동! 시스템 정지. 인간 리뷰 필요.",
		"repair_success": "자가 수정 성공: %s",
		"repair_failed":  "자가 수정 실패, 롤백: %s",
		"agent_spawn":    "%d개 에이전트 병렬 실행 시작",
		"agent_complete": "에이전트 %d 완료: %s",
		"cartridge_export": "카트리지 내보내기: %s (%d 바이트)",
		"cartridge_import": "카트리지 가져오기 완료: %s",
		"pii_masked":     "PII 마스킹: %s",
		"benchmark_start": "벤치마크 시작",
		"git_stash":      "git stash 성공",
		"git_rollback":   "git stash pop 완료 (롤백)",
		"git_commit":     "git commit 완료: %s",
	},
	"en": {
		"scan_start":     "Brain scan started",
		"scan_complete":  "Scan complete: %d neurons",
		"emit_start":     "Rule generation started",
		"emit_complete":  "Rule generation complete",
		"error_detected": "Error detected: %s",
		"error_escalated": "Error escalated: %s → P0",
		"bomb_triggered": "BOMB triggered! System halted. Human review required.",
		"repair_success": "Self-repair success: %s",
		"repair_failed":  "Self-repair failed, rollback: %s",
		"agent_spawn":    "%d agents spawned in parallel",
		"agent_complete": "Agent %d complete: %s",
		"cartridge_export": "Cartridge exported: %s (%d bytes)",
		"cartridge_import": "Cartridge imported: %s",
		"pii_masked":     "PII masked: %s",
		"benchmark_start": "Benchmark started",
		"git_stash":      "git stash success",
		"git_rollback":   "git stash pop complete (rollback)",
		"git_commit":     "git commit complete: %s",
	},
}

// LoadLocale reads language from brainstem/config.neuron
func LoadLocale(brainRoot string) Locale {
	lang := "ko" // default

	configPaths := []string{
		filepath.Join(brainRoot, "brainstem", "config.neuron"),
		filepath.Join(brainRoot, "brainstem", "lang.neuron"),
	}

	for _, path := range configPaths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		content := strings.TrimSpace(string(data))
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "lang:") {
				lang = strings.TrimSpace(strings.TrimPrefix(line, "lang:"))
				break
			}
		}
	}

	msgs, ok := localeMessages[lang]
	if !ok {
		msgs = localeMessages["ko"]
		lang = "ko"
	}

	return Locale{Lang: lang, Messages: msgs}
}

// Msg returns a localized message by key
func (l Locale) Msg(key string) string {
	if msg, ok := l.Messages[key]; ok {
		return msg
	}
	return key
}

// Msgf returns a formatted localized message
func (l Locale) Msgf(key string, args ...interface{}) string {
	if msg, ok := l.Messages[key]; ok {
		return fmt.Sprintf(msg, args...)
	}
	return fmt.Sprintf(key, args...)
}
