// RULE: ZERO EXTERNAL DEPENDENCY PRESERVED
// ide_integration.go — NeuronFS 네이티브 프로세스 매니저 및 IDE 통합
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func findAntigravity() string {
	home, _ := os.UserHomeDir()
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		localAppData = filepath.Join(home, "AppData", "Local")
	}

	candidates := []string{
		filepath.Join(localAppData, "Programs", "Antigravity", "Antigravity.exe"),
		filepath.Join(localAppData, "Programs", "cursor", "Cursor.exe"),
		filepath.Join(localAppData, "Programs", "Microsoft VS Code", "Code.exe"),
	}

	for _, name := range []string{"Antigravity", "cursor", "code"} {
		if p, err := exec.LookPath(name); err == nil {
			candidates = append([]string{p}, candidates...)
		}
	}

	for _, c := range candidates {
		if fileExists(c) {
			return c
		}
	}
	return ""
}

// ide_integration.go의 main/fileExists/svLog는 main.go에서 제공
// 아래 독립 진입점 제거 (main.go에 통합됨)

func svChatHistoryGuard() {
	appdata := os.Getenv("APPDATA")
	if appdata == "" {
		svLog("⚠️ [ChatGuard] APPDATA 환경변수 없음 — 비활성")
		return
	}
	gsDir := filepath.Join(appdata, "Antigravity", "User", "globalStorage")
	dbPath := filepath.Join(gsDir, "state.vscdb")

	if !fileExists(dbPath) {
		svLog("⚠️ [ChatGuard] state.vscdb 없음 — 비활성")
		return
	}

	svLog("🛡️ [ChatGuard] 감시 시작: " + dbPath)

	for {
		time.Sleep(5 * time.Minute)
		svChatHistoryCheck(dbPath, gsDir)
	}
}

func svChatHistoryCheck(dbPath, gsDir string) {
	out, err := exec.Command("sqlite3", dbPath, "SELECT length(value) FROM ItemTable WHERE key='chat.ChatSessionStore.index';").Output()
	if err != nil {
		return
	}

	size := 0
	fmt.Sscanf(strings.TrimSpace(string(out)), "%d", &size)

	if size >= 100 {
		return
	}

	svLog(fmt.Sprintf("🚨 [ChatGuard] 챗 히스토리 유실 감지 (현재: %d bytes)", size))

	entries, _ := os.ReadDir(gsDir)
	type candidate struct {
		path string
		size int
		time time.Time
	}
	var candidates []candidate

	for _, e := range entries {
		if !strings.HasPrefix(e.Name(), "state.vscdb.agmercium_recovery_") {
			continue
		}
		backupPath := filepath.Join(gsDir, e.Name())
		out, err := exec.Command("sqlite3", backupPath, "SELECT length(value) FROM ItemTable WHERE key='chat.ChatSessionStore.index';").Output()
		if err != nil {
			continue
		}
		bSize := 0
		fmt.Sscanf(strings.TrimSpace(string(out)), "%d", &bSize)
		if bSize >= 100 {
			info, _ := e.Info()
			if info != nil {
				candidates = append(candidates, candidate{backupPath, bSize, info.ModTime()})
			}
		}
	}

	if len(candidates) == 0 {
		svLog("🚨 [ChatGuard] 유효한 백업 없음 — 복원 불가")
		return
	}

	best := candidates[0]
	for _, c := range candidates[1:] {
		if c.size > best.size || (c.size == best.size && c.time.After(best.time)) {
			best = c
		}
	}

	svLog(fmt.Sprintf("🔄 [ChatGuard] 복원 시도: %s (%d bytes)", filepath.Base(best.path), best.size))

	pyScript := fmt.Sprintf(`import sqlite3
src = sqlite3.connect(r'%s')
data = src.execute("SELECT value FROM ItemTable WHERE key='chat.ChatSessionStore.index'").fetchone()
src.close()
if data and len(data[0]) > 100:
    dst = sqlite3.connect(r'%s')
    dst.execute("UPDATE ItemTable SET value=? WHERE key='chat.ChatSessionStore.index'", (data[0],))
    dst.commit()
    dst.close()
    print("OK")
else:
    print("FAIL")
`, best.path, dbPath)

	cmd := exec.Command("python", "-c", pyScript)
	result, err := cmd.Output()
	if err == nil && strings.TrimSpace(string(result)) == "OK" {
		svLog(fmt.Sprintf("✅ [ChatGuard] 챗 히스토리 복원 완료 (%d bytes)", best.size))
	} else {
		svLog("🚨 [ChatGuard] 복원 실패")
	}
}
