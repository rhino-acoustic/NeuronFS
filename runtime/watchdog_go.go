// watchdog_go.go — watchdog.mjs의 Go 포팅 (스트랭글러 피그 단계 1)
//
// supervisor.go와 병행 실행 가능. 기존 watchdog.mjs를 대체.
// 텔레그램 알림 + 5방향 헬스체크 + 메트릭 JSON 출력 + SLA 추적
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ── 헬스체크 설정 ──

type HealthCheck struct {
	Type    string `json:"type"`    // "port", "http", "log"
	Port    int    `json:"port,omitempty"`
	URL     string `json:"url,omitempty"`
	LogFile string `json:"logFile,omitempty"`
	MaxAge  int    `json:"maxAge,omitempty"` // seconds
	Desc    string `json:"desc"`
}

type WatchTarget struct {
	Name         string        `json:"name"`
	ProcessMatch string        `json:"processMatch"` // "node.*hijack-launcher" or "neuronfs.exe"
	IsExe        bool          `json:"isExe"`
	Health       []HealthCheck `json:"health"`

	mu           sync.Mutex
	zombieCount  int
	restartCount int
	lastSeen     time.Time
	downSince    *time.Time
}

type WatchdogMetrics struct {
	mu          sync.Mutex
	bootTime    time.Time
	checkCount  int
	totalMs     int64
	services    []*WatchTarget
	tgToken     string
	tgChatID    string
	metricsFile string
	logFile     string
}

func NewWatchdogMetrics(nfsRoot string) *WatchdogMetrics {
	logsDir := filepath.Join(nfsRoot, "logs")
	os.MkdirAll(logsDir, 0750)

	// 텔레그램 토큰 로드
	tgToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	tgChatID := ""
	bridgeDir := filepath.Join(nfsRoot, "telegram-bridge")
	if tgToken == "" {
		if data, err := os.ReadFile(filepath.Join(bridgeDir, ".token")); err == nil {
			tgToken = strings.TrimSpace(string(data))
		}
	}
	if data, err := os.ReadFile(filepath.Join(bridgeDir, ".chat_id")); err == nil {
		tgChatID = strings.TrimSpace(string(data))
	}

	return &WatchdogMetrics{
		bootTime:    time.Now(),
		metricsFile: filepath.Join(logsDir, "watchdog_metrics.json"),
		logFile:     filepath.Join(logsDir, "watchdog_go.log"),
		tgToken:     tgToken,
		tgChatID:    tgChatID,
	}
}

func (m *WatchdogMetrics) log(msg string) {
	line := fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), msg)
	log.Println(line)
	if m.logFile != "" {
		f, err := os.OpenFile(m.logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err == nil {
			fmt.Fprintln(f, line)
			f.Close()
		}
	}
}

func (m *WatchdogMetrics) tgAlert(msg string) {
	if m.tgToken == "" || m.tgChatID == "" {
		return
	}
	body := fmt.Sprintf(`{"chat_id":"%s","text":"🚨 [NeuronFS GoWD]\n%s"}`, m.tgChatID, msg)
	resp, err := http.Post(
		fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", m.tgToken),
		"application/json",
		strings.NewReader(body),
	)
	if err == nil {
		resp.Body.Close()
	}
}

// ── L2: 포트 체크 ──
func checkPort(port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 2*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// ── L3: HTTP 헬스체크 ──
func checkHTTP(url string) (int, error) {
	client := http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return 0, err
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	return resp.StatusCode, nil
}

// ── L4: 로그 신선도 ──
func checkLogFreshness(logFile string, maxAge int) (bool, int) {
	info, err := os.Stat(logFile)
	if err != nil {
		return false, -1
	}
	ageSec := int(time.Since(info.ModTime()).Seconds())
	return ageSec <= maxAge, ageSec
}

// ── deepHealthCheck ──
func (m *WatchdogMetrics) deepHealthCheck(t *WatchTarget) {
	var failed []string

	for _, h := range t.Health {
		switch h.Type {
		case "port":
			if !checkPort(h.Port) {
				failed = append(failed, h.Desc)
			}
		case "http":
			code, err := checkHTTP(h.URL)
			if err != nil || code != 200 {
				t.mu.Lock()
				t.zombieCount++
				t.mu.Unlock()
				detail := "연결 실패"
				if err == nil {
					detail = fmt.Sprintf("HTTP %d", code)
				}
				m.log(fmt.Sprintf("⚠️ [ZOMBIE] %s %s: %s", t.Name, h.Desc, detail))
			} else {
				t.mu.Lock()
				if t.zombieCount > 0 {
					t.zombieCount--
				}
				t.mu.Unlock()
			}
		case "log":
			ok, age := checkLogFreshness(h.LogFile, h.MaxAge)
			if !ok {
				failed = append(failed, fmt.Sprintf("%s (%ds stale)", h.Desc, age))
			}
		}
	}

	if len(failed) > 0 {
		t.mu.Lock()
		t.zombieCount++
		t.mu.Unlock()
		m.log(fmt.Sprintf("⚠️ [ZOMBIE] %s 동기체크 실패: %s (%d/3)", t.Name, strings.Join(failed, ", "), t.zombieCount))
	}

	t.mu.Lock()
	zc := t.zombieCount
	t.mu.Unlock()

	if zc >= 3 {
		m.log(fmt.Sprintf("🔴 [ZOMBIE_KILL] %s 3회 연속 좀비 → 알림", t.Name))
		m.tgAlert(fmt.Sprintf("🔴 [좀비 감지] %s 응답 없음 3회", t.Name))
		t.mu.Lock()
		t.zombieCount = 0
		t.mu.Unlock()
	}
}

// ── 메트릭 JSON 출력 ──
func (m *WatchdogMetrics) writeMetrics() {
	type svcMetric struct {
		Name         string   `json:"name"`
		Status       string   `json:"status"`
		SLA          float64  `json:"sla"`
		Restarts     int      `json:"restarts"`
		ZombieCount  int      `json:"zombieCount"`
		LastSeen     string   `json:"lastSeen"`
		HealthChecks []string `json:"healthChecks"`
	}

	svcs := make([]svcMetric, 0, len(m.services))
	for _, t := range m.services {
		t.mu.Lock()
		status := "UP"
		if t.downSince != nil {
			status = "DOWN"
		}
		hc := make([]string, 0, len(t.Health))
		for _, h := range t.Health {
			hc = append(hc, h.Desc)
		}
		svcs = append(svcs, svcMetric{
			Name:         t.Name,
			Status:       status,
			SLA:          100.0, // TODO: 정밀 계산
			Restarts:     t.restartCount,
			ZombieCount:  t.zombieCount,
			LastSeen:     t.lastSeen.Format(time.RFC3339),
			HealthChecks: hc,
		})
		t.mu.Unlock()
	}

	m.mu.Lock()
	avgMs := int64(0)
	if m.checkCount > 0 {
		avgMs = m.totalMs / int64(m.checkCount)
	}
	data := map[string]interface{}{
		"ts":         time.Now().Format(time.RFC3339),
		"bootTime":   m.bootTime.Format(time.RFC3339),
		"uptimeMs":   time.Since(m.bootTime).Milliseconds(),
		"checkCount": m.checkCount,
		"avgCheckMs": avgMs,
		"services":   svcs,
		"engine":     "go", // Go 엔진 식별자
	}
	m.mu.Unlock()

	jsonData, _ := json.MarshalIndent(data, "", "  ")
	os.WriteFile(m.metricsFile, jsonData, 0600)
}

// ── 텔레그램 상태 보고 ──
func (m *WatchdogMetrics) statusReport() {
	m.mu.Lock()
	uptime := time.Since(m.bootTime).Minutes()
	avgMs := int64(0)
	if m.checkCount > 0 {
		avgMs = m.totalMs / int64(m.checkCount)
	}
	m.mu.Unlock()

	report := fmt.Sprintf("📊 [NeuronFS GoWD 상태]\n⏱️ 가동: %.0f분 | 체크: %d회 | 평균: %dms\n", uptime, m.checkCount, avgMs)
	for _, t := range m.services {
		t.mu.Lock()
		status := "🟢 UP"
		if t.downSince != nil {
			status = "🔴 DOWN"
		}
		zombie := ""
		if t.zombieCount > 0 {
			zombie = fmt.Sprintf(" zombie=%d", t.zombieCount)
		}
		report += fmt.Sprintf("%s %s: restarts=%d%s\n", status, t.Name, t.restartCount, zombie)
		t.mu.Unlock()
	}
	m.log(report)
	m.tgAlert(report)
	m.writeMetrics()
}
