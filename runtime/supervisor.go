// supervisor.go — NeuronFS 네이티브 프로세스 매니저
//
// watchdog.ps1 + 프로세스 관리를 Go 바이너리로 통합.
// heartbeat 비활성화: brain_v4/_agents/pm/heartbeat.lock 생성 시 자동 중지.
//
// Usage: neuronfs <brain_path> --supervisor
package main

import (
	"fmt"
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
		f, err := os.OpenFile(svLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err == nil {
			fmt.Fprintln(f, line)
			f.Close()
		}
	}
}

func runSupervisor(brainRoot string) {
	nfsRoot := filepath.Dir(brainRoot)
	nfsExe, _ := os.Executable()
	logDir := filepath.Join(nfsRoot, "logs")
	os.MkdirAll(logDir, 0755)
	svLogPath = filepath.Join(logDir, "supervisor.log")
	harnessScript := filepath.Join(nfsRoot, "harness.ps1")
	userHome := filepath.Dir(nfsRoot)
	aaDir := filepath.Join(userHome, "auto-accept")
	nasBrain := os.Getenv("NEURONFS_NAS_BRAIN")

	fmt.Println("")
	fmt.Println("╔══════════════════════════════════════════════════╗")
	fmt.Println("║  NeuronFS Supervisor v2.0 — Self-Monitoring      ║")
	fmt.Println("║  프로세스 자동재시작 + heartbeat + 자기 감시      ║")
	fmt.Println("╚══════════════════════════════════════════════════╝")
	fmt.Println("")

	children := []*ChildSpec{
		{Name: "neuronfs-api", Cmd: nfsExe, Args: []string{brainRoot, "--api"}, Dir: nfsRoot, Enabled: true},
		{Name: "neuronfs-watch", Cmd: nfsExe, Args: []string{brainRoot, "--watch"}, Dir: nfsRoot, Enabled: true},
		{Name: "auto-accept", Cmd: "node", Args: []string{filepath.Join(aaDir, "auto-accept.mjs")}, Dir: aaDir, Enabled: svPathExists(filepath.Join(aaDir, "auto-accept.mjs"))},
	}

	svLog("🚀 Supervisor 시작")
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

	if nasBrain != "" && svPathExists(nasBrain) {
		go svNasSync(brainRoot, nasBrain, stopCh)
		svLog("🔄 NAS 동기화 활성 (5초)")
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	harnessTk := time.NewTicker(10 * time.Minute)
	statusTk := time.NewTicker(60 * time.Second)
	lockTk := time.NewTicker(5 * time.Second)
	defer harnessTk.Stop()
	defer statusTk.Stop()
	defer lockTk.Stop()

	svLog("━━━ 감시 루프 진입 ━━━")
	for {
		select {
		case <-sigCh:
			svLog("🛑 종료 신호")
			close(stopCh)
			for _, c := range children {
				c.stop()
			}
			wg.Wait()
			svLog("🛑 완료")
			return
		case <-harnessTk.C:
			go svHarness(harnessScript, brainRoot)
		case <-statusTk.C:
			svStatus(children)
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
		}
	}
}

func svSupervise(c *ChildSpec, stopCh <-chan struct{}) {
	const base = 2 * time.Second
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
			svLog(fmt.Sprintf("🚨 %s 서킷 브레이커 발동 — %d회 연속 크래시. 재시작 중단.", c.Name, c.restartCount))
			svCrashAlert(c)
			// Wait until crash window resets (60s) or stopCh
			select {
			case <-stopCh:
				return
			case <-time.After(60 * time.Second):
				c.restartCount = 0
				svLog(fmt.Sprintf("♻️ %s 쿨다운 완료 — 재시작 재개", c.Name))
			}
			continue
		}

		cmd := exec.Command(c.Cmd, c.Args...)
		cmd.Dir = c.Dir
		lp := filepath.Join(filepath.Dir(svLogPath), c.Name+".log")
		lf, err := os.OpenFile(lp, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err == nil {
			cmd.Stdout = lf
			cmd.Stderr = lf
		}

		c.mu.Lock()
		c.proc = cmd
		c.mu.Unlock()

		svLog(fmt.Sprintf("▶ %s 시작 (#%d)", c.Name, c.restartCount))
		if err := cmd.Start(); err != nil {
			svLog(fmt.Sprintf("❌ %s 실패: %v", c.Name, err))
			if lf != nil {
				lf.Close()
			}
			time.Sleep(base)
			continue
		}

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

		ec := -1
		if cmd.ProcessState != nil {
			ec = cmd.ProcessState.ExitCode()
		}
		svLog(fmt.Sprintf("⚠️ %s exit(%d) — 재시작 %d/%d → %v 후 재시작", c.Name, ec, c.restartCount, maxCrashBeforeCircuitBreak, delay.Round(time.Second)))

		select {
		case <-stopCh:
			return
		case <-time.After(delay):
		}
	}
}

func svNasSync(brain, nas string, stopCh <-chan struct{}) {
	for {
		select {
		case <-stopCh:
			return
		default:
		}
		cmd := exec.Command("robocopy", brain, nas, "/MIR", "/FFT", "/XO", "/MT:4", "/NDL", "/NJH", "/NJS", "/NC", "/NS", "/NP")
		cmd.Run()
		select {
		case <-stopCh:
			return
		case <-time.After(5 * time.Second):
		}
	}
}

func svHarness(script, brainRoot string) {
	if !svPathExists(script) {
		return
	}
	svLog("🔍 harness 실행")
	cmd := exec.Command("powershell", "-ExecutionPolicy", "Bypass", "-File", script)
	cmd.Dir = filepath.Dir(script)
	out, err := cmd.CombinedOutput()
	if err != nil {
		svLog(fmt.Sprintf("⚠️ harness 에러: %v", err))
		return
	}
	r := string(out)
	if strings.Contains(r, "FAIL: 0") || strings.Contains(r, "FAIL:  0") {
		svLog("✅ harness PASS")
	} else {
		svLog("⚠️ harness 위반")
		d := filepath.Join(brainRoot, "_agents", "bot1", "inbox")
		os.MkdirAll(d, 0755)
		f := filepath.Join(d, time.Now().Format("20060102_150405")+"_sv_harness.md")
		os.WriteFile(f, []byte(fmt.Sprintf("# from: supervisor\n# priority: urgent\n\nharness 위반.\n\n%s\n", r)), 0644)
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
			if svPathExists(c) {
				brainRoot = c
				break
			}
		}
	}
	if brainRoot == "" {
		return
	}

	inboxDir := filepath.Join(brainRoot, "_agents", "bot1", "inbox")
	os.MkdirAll(inboxDir, 0755)
	fname := fmt.Sprintf("%s_sv_crash_alert_%s.md", time.Now().Format("20060102_150405"), c.Name)
	content := fmt.Sprintf("# from: supervisor\n# priority: urgent\n\n## 🚨 프로세스 서킷 브레이커\n\n"+
		"**프로세스:** %s\n"+
		"**연속 크래시:** %d회\n"+
		"**시각:** %s\n\n"+
		"60초 쿨다운 후 재시작을 시도합니다. 반복 발생 시 수동 개입이 필요합니다.\n",
		c.Name, c.restartCount, time.Now().Format("2006-01-02 15:04:05"))
	os.WriteFile(filepath.Join(inboxDir, fname), []byte(content), 0644)
	svLog(fmt.Sprintf("📨 크래시 알림 → %s", fname))
}

func svPathExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}
