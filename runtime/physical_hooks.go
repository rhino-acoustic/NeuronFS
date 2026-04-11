package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// Physical Hooks — bomb.neuron 감지 시 물리 경보 발동
//
// 1. 모니터: 전체화면 빨간색 플래시 + BOMB 메시지
// 2. 스피커: 1000Hz 비프음 반복
// 3. 텔레그램: BotFather 봇으로 PD에게 알림 전송
// 4. USB 사이렌: COM 포트 시리얼 명령 (Adafruit Tower Light 등)
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// 텔레그램 봇 설정은 os.Getenv("TELEGRAM_BOT_TOKEN"), os.Getenv("TELEGRAM_CHAT_ID") 사용

// USB 사이렌 COM 포트 설정
// 추천 제품: Adafruit Tri-Color USB Tower Light w/ Buzzer (~$30)
// AliExpress "USB programmable signal tower light" (~$15)
const (
	usbSirenCOMPort = "" // "COM3" 등. 빈 문자열이면 스킵
)

// triggerPhysicalHook — bomb 감지 시 모든 물리 경보 발동
func triggerPhysicalHook(regionName string) {
	fmt.Fprintf(os.Stderr, "[CRITICAL] BOMB in '%s'. Physical alert triggered.\n", regionName)

	go triggerRedFlash(regionName)
	go triggerTelegram(regionName)
	go triggerUSBSiren(regionName)
}

// triggerRedFlash — 모니터 전체화면 빨간색 플래시 + 비프음
func triggerRedFlash(regionName string) {
	escapedRegion := strings.ReplaceAll(regionName, "'", "''")

	ps := fmt.Sprintf(`
# Mutex: skip if another bomb alert is already showing
$mtx = New-Object System.Threading.Mutex($false, 'Global\NeuronFS_BombAlert')
if (-not $mtx.WaitOne(0)) { exit 0 }

try {
Add-Type -AssemblyName PresentationFramework
Add-Type -AssemblyName PresentationCore

$window = New-Object System.Windows.Window
$window.WindowStyle = 'None'
$window.Width = 350
$window.Height = 100
$window.Left = 20
$window.Top = 20
$window.Topmost = $true
$window.ShowInTaskbar = $false

$grid = New-Object System.Windows.Controls.Grid
$grid.Background = [System.Windows.Media.Brushes]::Red

$text = New-Object System.Windows.Controls.TextBlock
$text.Text = [char]::ConvertFromUtf32(0x1F4A3) + " NEURONFS BOMB: %s"
$text.FontSize = 18
$text.FontWeight = 'Bold'
$text.Foreground = [System.Windows.Media.Brushes]::White
$text.HorizontalAlignment = 'Center'
$text.VerticalAlignment = 'Center'
$text.TextAlignment = 'Center'
$grid.Children.Add($text) | Out-Null
$window.Content = $grid

Start-Job -ScriptBlock { for ($i = 0; $i -lt 5; $i++) { [console]::beep(1000,300); Start-Sleep -m 200; [console]::beep(800,300); Start-Sleep -m 200 } } | Out-Null

$flash = $true
$timer = New-Object System.Windows.Threading.DispatcherTimer
$timer.Interval = [TimeSpan]::FromMilliseconds(500)
$timer.Add_Tick({ param($s,$e); if ($script:flash) { $grid.Background = [System.Windows.Media.Brushes]::Red } else { $grid.Background = [System.Windows.Media.Brushes]::DarkRed }; $script:flash = -not $script:flash })
$timer.Start()

$window.Add_MouseDown({ $window.Close() })
$window.ShowDialog() | Out-Null
} finally { $mtx.ReleaseMutex(); $mtx.Dispose() }
`, escapedRegion)

	if err := SafeExec(ExecTimeoutShell, "powershell", "-NoProfile", "-STA", "-Command", ps); err != nil {
		fmt.Fprintf(os.Stderr, "[HOOK] Red flash error: %v\n", err)
	}
}

// triggerTelegram — BotFather 봇으로 PD에게 긴급 알림
func triggerTelegram(regionName string) {
	tgToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	tgChatID := os.Getenv("TELEGRAM_CHAT_ID")

	if tgToken == "" || tgChatID == "" {
		fmt.Fprintf(os.Stderr, "[HOOK] Telegram not configured (set TELEGRAM_BOT_TOKEN and TELEGRAM_CHAT_ID env variables)\n")
		return
	}

	msg := fmt.Sprintf("NEURONFS BOMB\n\nRegion: %s\nAction: rm bomb.neuron\nStatus: Agent HALTED", regionName)
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", tgToken)

	resp, err := http.PostForm(apiURL, url.Values{
		"chat_id": {tgChatID},
		"text":    {msg},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "[HOOK] Telegram failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		fmt.Fprintf(os.Stderr, "[HOOK] Telegram alert sent.\n")
	} else {
		fmt.Fprintf(os.Stderr, "[HOOK] Telegram status %d\n", resp.StatusCode)
	}
}

// triggerUSBSiren — USB 시리얼 타워라이트 제어
// 추천 하드웨어:
//   - Adafruit Tri-Color USB Tower Light w/ Buzzer (Product ID 5125) ~$30
//     https://www.adafruit.com/product/5125
//   - AliExpress "USB signal tower light" CH340 기반 ~$15
//   - BlinkStick (blink1-tool CLI) ~$20
//
// 프로토콜: CH340 시리얼 → 0x11=Red ON, 0x21=Red OFF, 0x14=Buzzer ON
func triggerUSBSiren(regionName string) {
	if usbSirenCOMPort == "" {
		return // 미설정 시 스킵
	}

	ps := fmt.Sprintf(`
$port = New-Object System.IO.Ports.SerialPort "%s", 9600, None, 8, One
try {
    $port.Open()
    # Red ON + Buzzer ON
    $port.Write([byte[]]@(0x11, 0x14), 0, 2)
    Start-Sleep -Seconds 10
    # All OFF
    $port.Write([byte[]]@(0x21, 0x24), 0, 2)
    $port.Close()
} catch {
    Write-Error "USB siren error: $_"
}
`, usbSirenCOMPort)

	if err := SafeExec(ExecTimeoutShell, "powershell", "-NoProfile", "-Command", ps); err != nil {
		fmt.Fprintf(os.Stderr, "[HOOK] USB siren error: %v\n", err)
	}
}

// ─── EVOLVE TELEGRAM ALERTS ───

func actionIcon(actionType string) string {
	switch actionType {
	case "grow":
		return "🌱"
	case "fire":
		return "🔥"
	case "signal":
		return "📡"
	case "prune", "decay":
		return "💤"
	case "merge":
		return "🔗"
	default:
		return "❓"
	}
}

// sendTelegramSafe sends a text message via Telegram with automatic 4000-char splitting.
// This is the SSOT for all Telegram sends — no other function should call sendMessage directly.
func sendTelegramSafe(token, chatID, text string) {
	if token == "" || chatID == "" {
		return
	}
	runes := []rune(text)
	const chunkSize = 4000

	for i := 0; i < len(runes); i += chunkSize {
		end := i + chunkSize
		if end > len(runes) {
			end = len(runes)
		}
		chunk := string(runes[i:end])
		if len(runes) > chunkSize && i > 0 {
			chunk = fmt.Sprintf("(%d/%d) %s", (i/chunkSize)+1, (len(runes)+chunkSize-1)/chunkSize, chunk)
		}
		payload := map[string]string{
			"chat_id": chatID,
			"text":    chunk,
		}
		data, _ := json.Marshal(payload)
		resp, err := http.Post(
			"https://api.telegram.org/bot"+token+"/sendMessage",
			"application/json",
			bytes.NewReader(data),
		)
		if err == nil {
			resp.Body.Close()
		}
	}
}

// loadTelegramCreds reads token and chat_id from telegram-bridge directory.
func loadTelegramCreds(brainRoot string) (token, chatID string) {
	bridgeDir := filepath.Join(filepath.Dir(brainRoot), "telegram-bridge")
	if b, err := os.ReadFile(filepath.Join(bridgeDir, ".token")); err == nil {
		token = strings.TrimSpace(string(b))
	}
	if token == "" {
		token = os.Getenv("TELEGRAM_BOT_TOKEN")
	}
	if b, err := os.ReadFile(filepath.Join(bridgeDir, ".chat_id")); err == nil {
		chatID = strings.TrimSpace(string(b))
	}
	if chatID == "" {
		chatID = os.Getenv("TELEGRAM_CHAT_ID")
	}
	return
}

// sendTelegramEvolve sends a push notification about brain evolution
func sendTelegramEvolve(brainRoot string, action evoAction) {
	token, chatID := loadTelegramCreds(brainRoot)
	if token == "" || chatID == "" {
		return
	}

	icon := actionIcon(action.Type)
	msg := fmt.Sprintf("🧬 [NEURON EVOLVED] %s\n\n액션: %s %s\n사유: %s",
		action.Path, icon, strings.ToUpper(action.Type), action.Reason)
	sendTelegramSafe(token, chatID, msg)
}
