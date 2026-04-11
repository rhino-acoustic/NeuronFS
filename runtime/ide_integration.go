// RULE: ZERO EXTERNAL DEPENDENCY PRESERVED
// supervisor.go — NeuronFS 네이티브 프로세스 매니저
//
// watchdog.ps1 + 프로세스 관리를 Go 바이너리로 통합.
//
// Usage: neuronfs <brain_path> --supervisor
package main

import (
	"encoding/json"
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

	// PATH에서도 탐색
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
	// 현재 chat index 크기 확인
	out, err := exec.Command("sqlite3", dbPath,
		"SELECT length(value) FROM ItemTable WHERE key='chat.ChatSessionStore.index';").Output()
	if err != nil {
		return // sqlite3 없거나 DB 잠금
	}

	sizeStr := strings.TrimSpace(string(out))
	size := 0
	fmt.Sscanf(sizeStr, "%d", &size)

	if size >= 100 {
		return // 정상 (100바이트 이상이면 유효한 데이터)
	}

	svLog(fmt.Sprintf("🚨 [ChatGuard] 챗 히스토리 유실 감지 (현재: %d bytes)", size))

	// recovery 백업에서 유효한 데이터 찾기
	entries, err := os.ReadDir(gsDir)
	if err != nil {
		return
	}

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
		out, err := exec.Command("sqlite3", backupPath,
			"SELECT length(value) FROM ItemTable WHERE key='chat.ChatSessionStore.index';").Output()
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

	// 가장 최신이면서 가장 큰 백업 선택
	best := candidates[0]
	for _, c := range candidates[1:] {
		if c.size > best.size || (c.size == best.size && c.time.After(best.time)) {
			best = c
		}
	}

	svLog(fmt.Sprintf("🔄 [ChatGuard] 복원 시도: %s (%d bytes)", filepath.Base(best.path), best.size))

	// Python으로 안전하게 복원 (parameterized query)
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
	if err != nil {
		svLog(fmt.Sprintf("🚨 [ChatGuard] 복원 실패: %v", err))
		return
	}

	if strings.TrimSpace(string(result)) == "OK" {
		svLog(fmt.Sprintf("✅ [ChatGuard] 챗 히스토리 복원 완료 (%d bytes)", best.size))
	} else {
		svLog("🚨 [ChatGuard] 복원 실패 — 데이터 무효")
	}
}

func svRestoreConversation(nfsRoot, brainRoot string) {
	ctxPath := filepath.Join(nfsRoot, ".restart_context")
	if !fileExists(ctxPath) {
		return
	}

	data, err := os.ReadFile(ctxPath)
	if err != nil {
		return
	}

	var ctx struct {
		Ts    string `json:"ts"`
		Title string `json:"title"`
		Room  string `json:"room"`
	}
	if json.Unmarshal(data, &ctx) != nil {
		os.Remove(ctxPath)
		return
	}

	svLog(fmt.Sprintf("🔄 대화 복귀 시작: %s", ctx.Title))

	// 텔레그램 알림
	hlLoadTelegram(nfsRoot)
	hlTgSend(hlTgChatID, fmt.Sprintf("🔄 NeuronFS 재시작 완료\n📌 대화 복귀 시도: %s", ctx.Title))

	// Antigravity CDP 연결 대기 (최대 60초)
	var connected bool
	for i := 0; i < 20; i++ {
		time.Sleep(3 * time.Second)
		targets, err := cdpListTargets(9000)
		if err != nil {
			continue
		}
		for _, t := range targets {
			if strings.Contains(t.URL, "workbench.html") && t.WebSocketDebuggerURL != "" {
				connected = true
				break
			}
		}
		if connected {
			break
		}
	}

	if !connected {
		svLog("⚠️ 대화 복귀 실패: Antigravity CDP 연결 안 됨")
		hlTgSend(hlTgChatID, "⚠️ 대화 복귀 실패: Antigravity CDP 연결 안 됨")
		os.Remove(ctxPath)
		return
	}

	// CDP 안정화 대기
	time.Sleep(5 * time.Second)

	// 대화 복귀 메시지 인젝션
	targets, _ := cdpListTargets(9000)
	injected := false
	for _, t := range targets {
		if !strings.Contains(t.URL, "workbench.html") || t.WebSocketDebuggerURL == "" {
			continue
		}
		// mounted room에 해당하는 창 찾기
		if ctx.Room != "" && !strings.Contains(strings.ToLower(t.Title), strings.ToLower(ctx.Room)) {
			continue
		}

		client, err := NewCDPClient(t.WebSocketDebuggerURL)
		if err != nil {
			continue
		}
		client.Call("Runtime.enable", map[string]interface{}{})
		time.Sleep(500 * time.Millisecond)

		// 이전 대화 복귀 메시지 주입
		escaped := strings.ReplaceAll(ctx.Title, `\`, `\\`)
		escaped = strings.ReplaceAll(escaped, `"`, `\"`)
		escaped = strings.ReplaceAll(escaped, "\n", `\n`)

		injectExpr := fmt.Sprintf(`(() => {
			const all = Array.from(document.querySelectorAll("[contenteditable]"));
			const el = all.reverse().find(e => {
				const r = e.getBoundingClientRect();
				return r.height > 0 && r.height < 300 && r.width > 100;
			}) || all[0];
			if(el) {
				el.focus();
				document.execCommand("insertText", false, "[NeuronFS 재시작 복귀] start.bat 재시작 완료. 이전 대화 '%s'에서 복귀. 중단된 작업이 있으면 이어서 진행해줘.");
				return "Injected";
			}
			return "NoTarget";
		})()`, escaped)

		result, err := client.Call("Runtime.evaluate", map[string]interface{}{
			"expression":    injectExpr,
			"returnByValue": true,
		})
		client.Close()

		if err == nil {
			var evalRes struct {
				Result struct {
					Value string `json:"value"`
				} `json:"result"`
			}
			json.Unmarshal(result, &evalRes)
			if evalRes.Result.Value == "Injected" {
				injected = true
				break
			}
		}
	}

	if injected {
		svLog("✅ 대화 복귀 인젝션 성공")
		hlTgSend(hlTgChatID, "✅ 대화 복귀 완료")
	} else {
		svLog("⚠️ 대화 복귀 인젝션 실패 — 수동 복귀 필요")
		hlTgSend(hlTgChatID, "⚠️ 대화 복귀 인젝션 실패")
	}

	os.Remove(ctxPath)
}

func svPatchAntigravityShortcuts() {
	home, _ := os.UserHomeDir()
	searchDirs := []string{
		filepath.Join(home, "AppData", "Roaming", "Microsoft", "Windows", "Start Menu", "Programs"),
		filepath.Join(home, "AppData", "Roaming", "Microsoft", "Internet Explorer", "Quick Launch", "User Pinned", "TaskBar"),
		filepath.Join(home, "Desktop"),
	}

	cdpFlag := "--remote-debugging-port=9000"
	psTemplate := `$shell = New-Object -ComObject WScript.Shell; $lnk = $shell.CreateShortcut('%s'); if ($lnk.Arguments -notlike '*remote-debugging-port*') { if ($lnk.Arguments) { $lnk.Arguments = $lnk.Arguments + ' %s' } else { $lnk.Arguments = '%s' }; $lnk.Save(); Write-Output 'PATCHED' } else { Write-Output 'OK' }`

	for _, dir := range searchDirs {
		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			if !strings.HasSuffix(strings.ToLower(info.Name()), ".lnk") {
				return nil
			}
			if !strings.Contains(strings.ToLower(info.Name()), "antigravity") {
				return nil
			}

			escaped := strings.ReplaceAll(path, "'", "''")
			ps := fmt.Sprintf(psTemplate, escaped, cdpFlag, cdpFlag)
			out, err := exec.Command("powershell", "-NoProfile", "-Command", ps).CombinedOutput()
			result := strings.TrimSpace(string(out))
			if err == nil && result == "PATCHED" {
				svLog(fmt.Sprintf("🔧 CDP 플래그 주입: %s", filepath.Base(path)))
			}
			return nil
		})
	}
}
