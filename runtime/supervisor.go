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
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type ChildSpec struct {
	Name     string
	Cmd      string
	Args     []string
	Dir      string
	Enabled  bool
	Lockable bool
	LockPath string

	mu           sync.Mutex
	proc         *exec.Cmd
	running      bool
	restartCount int
	lastCrash    time.Time
}

func (c *ChildSpec) isLocked() bool {
	if !c.Lockable || c.LockPath == "" {
		return false
	}
	_, err := os.Stat(c.LockPath)
	return err == nil
}

func (c *ChildSpec) stop() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.proc != nil && c.proc.Process != nil {
		c.proc.Process.Kill()
	}
	c.running = false
}

var svLogPath string

func svLog(msg string) {
	ts := time.Now().Format("15:04:05")
	line := fmt.Sprintf("[%s] %s", ts, msg)
	fmt.Println(line)
	if svLogPath != "" {
		f, err := os.OpenFile(svLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err == nil {
			fmt.Fprintln(f, line)
			f.Close()
		}
	}
}

// ── L2: 포트 체크 (Go 네이티브) ──
func checkPort(port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 2*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// ── 텔레그램 알림 (Go 네이티브, 외부 의존 0) ──
var svTgToken, svTgChatID string

func svLoadTelegram(nfsRoot string) {
	bridgeDir := filepath.Join(nfsRoot, "telegram-bridge")
	if data, err := os.ReadFile(filepath.Join(bridgeDir, ".token")); err == nil {
		svTgToken = strings.TrimSpace(string(data))
	}
	if data, err := os.ReadFile(filepath.Join(bridgeDir, ".chat_id")); err == nil {
		svTgChatID = strings.TrimSpace(string(data))
	}
}

func svTgAlert(msg string) {
	if svTgToken == "" || svTgChatID == "" {
		return
	}
	body := fmt.Sprintf(`{"chat_id":"%s","text":"%s"}`, svTgChatID, strings.ReplaceAll(msg, `"`, `\"`))
	resp, err := http.Post(
		fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", svTgToken),
		"application/json",
		strings.NewReader(body),
	)
	if err == nil {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

// ── 메트릭 JSON 출력 ──
var svBootTime = time.Now()
var svCheckCount int

func svWriteMetrics(children []*ChildSpec, nfsRoot string) {
	type svcEntry struct {
		Name     string `json:"name"`
		Status   string `json:"status"`
		Restarts int    `json:"restarts"`
	}
	svcs := make([]svcEntry, 0)
	for _, c := range children {
		c.mu.Lock()
		st := "DOWN"
		if c.running {
			st = "UP"
		}
		svcs = append(svcs, svcEntry{Name: c.Name, Status: st, Restarts: c.restartCount})
		c.mu.Unlock()
	}
	data := map[string]interface{}{
		"ts":         time.Now().Format(time.RFC3339),
		"bootTime":   svBootTime.Format(time.RFC3339),
		"uptimeMs":   time.Since(svBootTime).Milliseconds(),
		"checkCount": svCheckCount,
		"services":   svcs,
		"engine":     "go-supervisor",
	}
	jsonData, _ := json.MarshalIndent(data, "", "  ")
	os.WriteFile(filepath.Join(nfsRoot, "logs", "watchdog_metrics.json"), jsonData, 0600)
}

func svStatusReport(children []*ChildSpec, nfsRoot string) {
	uptime := time.Since(svBootTime).Minutes()
	report := fmt.Sprintf("📊 [NeuronFS Supervisor]\n⏱️ 가동: %.0f분 | 체크: %d회\n", uptime, svCheckCount)
	for _, c := range children {
		c.mu.Lock()
		st := "🟢 UP"
		if !c.running {
			st = "🔴 DOWN"
		}
		report += fmt.Sprintf("%s %s restarts=%d\n", st, c.Name, c.restartCount)
		c.mu.Unlock()
	}
	svLog(report)
	svTgAlert(report)
	svWriteMetrics(children, nfsRoot)
}

func runSupervisor(brainRoot string) {
	nfsRoot := filepath.Dir(brainRoot)
	nfsExe, _ := os.Executable()
	logDir := filepath.Join(nfsRoot, "logs")
	os.MkdirAll(logDir, 0750)
	svLogPath = filepath.Join(logDir, "supervisor.log")
	userHome := filepath.Dir(nfsRoot)
	aaDir := filepath.Join(userHome, "auto-accept")
	nasBrain := os.Getenv("NEURONFS_NAS_BRAIN")

	fmt.Println("")
	fmt.Println("╔══════════════════════════════════════════════════╗")
	fmt.Println("║  NeuronFS Supervisor v2.1 — Self-Monitoring      ║")
	fmt.Println("║  프로세스 자동재시작 + 자기 감시                  ║")
	fmt.Println("╚══════════════════════════════════════════════════╝")
	fmt.Println("")

	hijackDir := filepath.Join(userHome, "_architecture_hijack_v4")

	children := []*ChildSpec{
		{Name: "neuronfs-api", Cmd: nfsExe, Args: []string{brainRoot, "--api"}, Dir: nfsRoot, Enabled: true},
		{Name: "neuronfs-watch", Cmd: nfsExe, Args: []string{brainRoot, "--watch"}, Dir: nfsRoot, Enabled: true},
		{Name: "auto-accept", Cmd: "node", Args: []string{filepath.Join(aaDir, "auto-accept.mjs")}, Dir: aaDir, Enabled: fileExists(filepath.Join(aaDir, "auto-accept.mjs"))},
		{Name: "agent-bridge", Cmd: "node", Args: []string{filepath.Join(nfsRoot, "runtime", "core_agents", "agent-bridge.mjs")}, Dir: nfsRoot, Enabled: true},
		{Name: "hijack-launcher", Cmd: "node", Args: []string{filepath.Join(nfsRoot, "runtime", "hijackers", "hijack-launcher.mjs")}, Dir: nfsRoot, Enabled: true},
		{Name: "headless-executor", Cmd: "node", Args: []string{filepath.Join(hijackDir, "headless-executor.mjs")}, Dir: hijackDir, Enabled: fileExists(filepath.Join(hijackDir, "headless-executor.mjs"))},
		// context-hijacker: Go 네이티브 (Node 은퇴)
	}

	svLog("\033[35m[AURA] Awakening cognitive architecture... Supervisor online.\033[0m")
	svLog(fmt.Sprintf("   brain: %s", brainRoot))
	for _, c := range children {
		s := "활성"
		if !c.Enabled {
			s = "비활성"
		}
		extra := ""
		if c.Lockable {
			if c.isLocked() {
				extra = " [🔒 LOCKED]"
			} else {
				extra = " [PM 제어]"
			}
		}
		svLog(fmt.Sprintf("   %-18s %s%s", c.Name, s, extra))
	}
	svLog("")

	stopCh := make(chan struct{})
	var wg sync.WaitGroup
	for _, child := range children {
		if !child.Enabled {
			continue
		}
		wg.Add(1)
		go func(c *ChildSpec) {
			defer wg.Done()
			svSupervise(c, stopCh)
		}(child)
	}

	if nasBrain != "" && fileExists(nasBrain) {
		go SyncToNAS(brainRoot, nasBrain, stopCh)
		svLog("🔄 NAS 동기화 활성 (5초)")
	}

	// Go 네이티브 context hijacker (Node 대체)
	go runContextHijacker(brainRoot)
	svLog("📡 Context Hijacker (Go native) 시작")

	svBootTime = time.Now()
	svLoadTelegram(nfsRoot)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	harnessTk := time.NewTicker(10 * time.Minute)
	statusTk := time.NewTicker(60 * time.Second)
	lockTk := time.NewTicker(5 * time.Second)
	decayTk := time.NewTicker(1 * time.Hour)
	reportTk := time.NewTicker(30 * time.Minute)
	defer harnessTk.Stop()
	defer statusTk.Stop()
	defer lockTk.Stop()
	defer decayTk.Stop()
	defer reportTk.Stop()

	// Initial decay run
	go RunTTLDecay(brainRoot, svLog)

	svLog("━━━ 감시 루프 진입 ━━━")
	for {
		select {
		case <-sigCh:
			svLog("\033[90m[SLUMBER] Initiating graceful shutdown sequence...\033[0m")
			close(stopCh)
			for _, c := range children {
				c.stop()
			}
			wg.Wait()
			svLog("\033[90m[SLUMBER] Cognitive architecture offline.\033[0m")
			return
		case <-harnessTk.C:
			go RunHarness(brainRoot, svLog)
		case <-statusTk.C:
			svCheckCount++
			svStatus(children)
			svWriteMetrics(children, nfsRoot)
		case <-lockTk.C:
			for _, c := range children {
				if !c.Lockable {
					continue
				}
				c.mu.Lock()
				locked := c.isLocked()
				wasRunning := c.running
				c.mu.Unlock()
				if locked && wasRunning {
					svLog(fmt.Sprintf("🔒 %s — PM lock. 중지.", c.Name))
					c.stop()
				}
			}
		case <-decayTk.C:
			go RunTTLDecay(brainRoot, svLog)
		case <-reportTk.C:
			svStatusReport(children, nfsRoot)
		}
	}
}

func svSupervise(c *ChildSpec, stopCh <-chan struct{}) {
	const base = 1 * time.Second
	const maxD = 5 * time.Minute
	const maxCrashBeforeCircuitBreak = 10

	for {
		select {
		case <-stopCh:
			return
		default:
		}
		if c.isLocked() {
			time.Sleep(3 * time.Second)
			continue
		}

		// Circuit breaker: too many rapid restarts → suspend + alert
		if c.restartCount >= maxCrashBeforeCircuitBreak {
			svLog(fmt.Sprintf("\033[31m[TRAUMA] Circuit breaker triggered for %s. Vital signs critical (%d failures).\033[0m", c.Name, c.restartCount))
			svCrashAlert(c)
			// Wait until crash window resets (60s) or stopCh
			select {
			case <-stopCh:
				return
			case <-time.After(60 * time.Second):
				c.restartCount = 0
				svLog(fmt.Sprintf("\033[32m[HEAL] Trauma stabilized for %s. Re-engaging neurogenesis.\033[0m", c.Name))
			}
			continue
		}

		cmd := exec.Command(c.Cmd, c.Args...)
		cmd.Dir = c.Dir
		lp := filepath.Join(filepath.Dir(svLogPath), c.Name+".log")
		lf, err := os.OpenFile(lp, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err == nil {
			cmd.Stdout = lf
			cmd.Stderr = lf
		}

		c.mu.Lock()
		c.proc = cmd
		c.mu.Unlock()

		svLog(fmt.Sprintf("▶ %s 시작 (#%d)", c.Name, c.restartCount))
		if err := cmd.Start(); err != nil {
			svLog(fmt.Sprintf("\033[33m[FRACTURE] %s neurogenesis failed: %v\033[0m", c.Name, err))
			if lf != nil {
				lf.Close()
			}
			time.Sleep(base)
			continue
		}
		svLog("\033[32m[NEURON] Cortex online. Process stabilized.\033[0m")

		c.mu.Lock()
		c.running = true
		c.mu.Unlock()

		_ = cmd.Wait()
		if lf != nil {
			lf.Close()
		}

		c.mu.Lock()
		c.running = false
		c.mu.Unlock()

		select {
		case <-stopCh:
			return
		default:
		}

		if c.isLocked() {
			svLog(fmt.Sprintf("🔒 %s 종료 — lock 대기", c.Name))
			continue
		}

		now := time.Now()
		if now.Sub(c.lastCrash) > 60*time.Second {
			c.restartCount = 0
		} else {
			c.restartCount++
		}
		c.lastCrash = now

		shift := c.restartCount
		if shift > 8 {
			shift = 8
		}
		delay := base * time.Duration(1<<uint(shift))
		if delay > maxD {
			delay = maxD
		}

		svLog(fmt.Sprintf("\033[36m[HEAL] Initiating rapid neurogenesis (Rebirth in %dms)...\033[0m", delay.Milliseconds()))

		select {
		case <-stopCh:
			return
		case <-time.After(delay):
		}
	}
}

func svStatus(children []*ChildSpec) {
	var p []string
	for _, c := range children {
		c.mu.Lock()
		s := "❌"
		if c.running {
			s = "✅"
		}
		if c.isLocked() {
			s = "🔒"
		}
		if !c.Enabled {
			s = "⬛"
		}
		p = append(p, fmt.Sprintf("%s:%s", c.Name, s))
		c.mu.Unlock()
	}
	svLog("💓 " + strings.Join(p, " | "))

	for _, c := range children {
		if !c.Enabled {
			continue
		}
		c.mu.Lock()
		running := c.running
		pid := 0
		if c.proc != nil && c.proc.Process != nil {
			pid = c.proc.Process.Pid
		}
		c.mu.Unlock()

		if running && pid > 0 {
			if checkProcessMemoryOverload(c.Name, pid) {
				c.stop()
			}

			// L3: HTTP 헬스체크 (API 서버)
			if c.Name == "neuronfs-api" {
				client := http.Client{Timeout: 3 * time.Second}
				resp, err := client.Get(fmt.Sprintf("http://127.0.0.1:%d/api/health", APIPort))
				if err != nil {
					svLog("\033[31m[TRAUMA] Synaptic overload detected. Memory integrity compromised.\033[0m")
					c.stop()
				} else if resp != nil {
					resp.Body.Close()
				}

				// L2: 포트 체크
				if !checkPort(APIPort) {
					svLog(fmt.Sprintf("\033[33m[ZOMBIE] API port %d not listening\033[0m", APIPort))
				}
			}
		}
	}
}

func svCrashAlert(c *ChildSpec) {
	// Write crash alert to _agents/bot1/inbox for automatic pickup
	brainRoot := ""
	if svLogPath != "" {
		brainRoot = filepath.Dir(filepath.Dir(svLogPath))
		// svLogPath = .../logs/supervisor.log → parent of logs = NeuronFS root
		// brain is brainRoot's child brain_v4/
		candidates := []string{
			filepath.Join(brainRoot, "brain_v4"),
			filepath.Join(brainRoot, "brain"),
		}
		for _, c := range candidates {
			if fileExists(c) {
				brainRoot = c
				break
			}
		}
	}
	if brainRoot == "" {
		return
	}

	inboxDir := filepath.Join(brainRoot, "_agents", "bot1", "inbox")
	os.MkdirAll(inboxDir, 0750)
	fname := fmt.Sprintf("%s_sv_crash_alert_%s.md", time.Now().Format("20060102_150405"), c.Name)
	content := fmt.Sprintf("# from: supervisor\n# priority: urgent\n\n## 🚨 프로세스 서킷 브레이커\n\n"+
		"**프로세스:** %s\n"+
		"**연속 크래시:** %d회\n"+
		"**시각:** %s\n\n"+
		"60초 쿨다운 후 재시작을 시도합니다. 반복 발생 시 수동 개입이 필요합니다.\n",
		c.Name, c.restartCount, time.Now().Format("2006-01-02 15:04:05"))
	os.WriteFile(filepath.Join(inboxDir, fname), []byte(content), 0600)
	svLog(fmt.Sprintf("📨 크래시 알림 → %s", fname))
}
