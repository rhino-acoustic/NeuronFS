// headless_executor.go — headless-executor.mjs Go 포팅
// Inbox 감시 → 에이전트 페이로드에서 명령어 추출 → 샌드박스 실행
// 외부 의존성: 0
package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

func runHeadlessExecutor(brainRoot string) {
	nfsRoot := filepath.Dir(brainRoot)
	inboxDir := filepath.Join(brainRoot, "_agents", "global_inbox")
	outboxDir := filepath.Join(brainRoot, "_agents", "pm", "outbox")
	os.MkdirAll(inboxDir, 0750)
	os.MkdirAll(outboxDir, 0750)

	// 禁 destructive command guardrail
	forbiddenRe := loadForbiddenRegex(brainRoot)

	heLog("Headless Executor (Go native) 시작 — inbox: %s", inboxDir)

	// Polling (500ms) — fsnotify 불필요
	seen := make(map[string]bool)
	for {
		entries, err := os.ReadDir(inboxDir)
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
				continue
			}
			if seen[e.Name()] {
				continue
			}
			seen[e.Name()] = true

			fp := filepath.Join(inboxDir, e.Name())
			// 쓰기 완료 대기
			time.Sleep(500 * time.Millisecond)

			data, err := os.ReadFile(fp)
			if err != nil {
				continue
			}

			cmds := extractCommands(string(data))
			if len(cmds) == 0 {
				continue
			}

			for _, cmd := range cmds {
				heExecute(cmd, nfsRoot, outboxDir, forbiddenRe)
			}

			os.Remove(fp)
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func heLog(format string, args ...interface{}) {
	fmt.Printf("[%s] [HE] %s\n", time.Now().Format("15:04:05"), fmt.Sprintf(format, args...))
}

func loadForbiddenRegex(brainRoot string) *regexp.Regexp {
	defaultRe := regexp.MustCompile(`(?i)rm\s+-rf|format\s+[A-Z]:|del\s+/f\s+/s\s+/q\s+[A-Z]:\\`)
	rulePath := filepath.Join(brainRoot, "brainstem", "禁destructive_command.txt")
	data, err := os.ReadFile(rulePath)
	if err != nil {
		return defaultRe
	}
	lines := strings.Split(string(data), "\n")
	inSection := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "[FORBIDDEN_COMMANDS]" {
			inSection = true
			continue
		}
		if inSection && line != "" {
			if strings.HasPrefix(line, "(") && strings.HasSuffix(line, ")") {
				line = line[1 : len(line)-1]
			}
			if re, err := regexp.Compile("(?i)" + line); err == nil {
				return re
			}
		}
	}
	return defaultRe
}

func extractCommands(payload string) []string {
	var cmds []string

	// JSON parsing
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(payload), &parsed); err == nil {
		// Schema A: Gemini candidates
		if candidates, ok := parsed["candidates"].([]interface{}); ok {
			for _, c := range candidates {
				cm, _ := c.(map[string]interface{})
				content, _ := cm["content"].(map[string]interface{})
				parts, _ := content["parts"].([]interface{})
				for _, p := range parts {
					pm, _ := p.(map[string]interface{})
					fc, _ := pm["functionCall"].(map[string]interface{})
					if fc["name"] == "run_command" {
						args, _ := fc["args"].(map[string]interface{})
						if cl, ok := args["CommandLine"].(string); ok {
							cmds = append(cmds, cl)
						}
					}
				}
			}
		}
		// Schema B: Claude content
		if content, ok := parsed["content"].([]interface{}); ok {
			for _, b := range content {
				bm, _ := b.(map[string]interface{})
				if bm["type"] == "tool_use" && bm["name"] == "run_command" {
					input, _ := bm["input"].(map[string]interface{})
					if cl, ok := input["CommandLine"].(string); ok {
						cmds = append(cmds, cl)
					}
				}
			}
		}
		// Schema C: direct tool_call
		if tc, ok := parsed["tool_call"].(map[string]interface{}); ok {
			if tc["name"] == "run_command" {
				args, _ := tc["arguments"].(map[string]interface{})
				if cl, ok := args["CommandLine"].(string); ok {
					cmds = append(cmds, cl)
				}
			}
		}
	}

	// Fallback: code block extraction
	if len(cmds) == 0 {
		re := regexp.MustCompile("```(?:bash|powershell|cmd)\n([\\s\\S]*?)\n```")
		matches := re.FindAllStringSubmatch(payload, -1)
		for _, m := range matches {
			cmds = append(cmds, strings.TrimSpace(m[1]))
		}
	}

	return cmds
}

func heExecute(command, cwd, outboxDir string, forbiddenRe *regexp.Regexp) {
	if forbiddenRe.MatchString(command) {
		heLog("🚨 BLOCKED: %s", command)
		hash := heRandHex(4)
		report := fmt.Sprintf("## Headless Execution Report\n\n**Command:** `%s`\n**Status:** BLOCKED BY GUARDRAIL\n", command)
		os.WriteFile(filepath.Join(outboxDir, "headless_report_"+hash+".md"), []byte(report), 0600)
		return
	}

	heLog("⚡ 실행: %s", command)
	start := time.Now()

	cmd := exec.Command("powershell", "-NoProfile", "-Command", command)
	cmd.Dir = cwd
	out, err := cmd.CombinedOutput()
	dur := time.Since(start).Milliseconds()

	report := fmt.Sprintf("## Headless Execution Report\n\n**Command:** `%s`\n**Duration:** %dms\n\n", command, dur)
	if err != nil {
		report += fmt.Sprintf("### Error\n```\n%s\n```\n", err.Error())
	}
	if len(out) > 0 {
		report += fmt.Sprintf("### Output\n```\n%s\n```\n", string(out))
	}

	hash := heRandHex(4)
	os.WriteFile(filepath.Join(outboxDir, "headless_report_"+hash+".md"), []byte(report), 0600)
	heLog("💾 결과 저장: headless_report_%s.md (%dms)", hash, dur)
}

func heRandHex(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}
