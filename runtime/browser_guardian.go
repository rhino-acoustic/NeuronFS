package main

import (
	"fmt"
	"net"
	"os/exec"
	"time"
)

// EnsureBrowserAlive 9222 디버그 포트 핑을 날려 브라우저가 죽었으면 강제 스폰한다
func EnsureBrowserAlive() {
	timeout := time.Second
	conn, err := net.DialTimeout("tcp", net.JoinHostPort("127.0.0.1", "9222"), timeout)
	if err != nil {
		fmt.Println("[GUARDIAN] 🚨 Chrome 9222 포트 침묵 감지. 즉시 워크벤치 스폰(Spawn) 실시!")

		chromePath := `C:\Program Files\Google\Chrome\Application\chrome.exe`
		// 기존 세션 복구 및 기본 타겟(AI Studio) 지정
		cmd := exec.Command(chromePath, "--remote-debugging-port=9222", "--restore-last-session", "https://aistudio.google.com/app/prompts", "https://aistudio.google.com/app/prompts")

		// Daemon 스폰을 위해 Start() 처리
		err = cmd.Start()
		if err != nil {
			fmt.Println("[GUARDIAN] ❌ Chrome 스폰 실패:", err)
		} else {
			fmt.Println("[GUARDIAN] ✅ Chrome 브라우저 스폰 가동 완료. (PID:", cmd.Process.Pid, ")")
		}
		// 스폰 딜레이 여유 시간 부여
		time.Sleep(4 * time.Second)
	} else {
		if conn != nil {
			conn.Close()
		}
	}
}
