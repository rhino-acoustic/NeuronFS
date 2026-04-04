// RULE: ZERO EXTERNAL DEPENDENCY PRESERVED
// supervisor.go — NeuronFS 네이티브 프로세스 매니저
//
// watchdog.ps1 + 프로세스 관리를 Go 바이너리로 통합.
//
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
	"strconv"
	"net/http"
	"runtime"
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
	fmt.Println("║  NeuronFS Supervisor v2.1 — Self-Monitoring      ║")
	fmt.Println("║  프로세스 자동재시작 + 자기 감시                  ║")
	fmt.Println("╚══════════════════════════════════════════════════╝")
	fmt.Println("")

	hijackDir := filepath.Join(userHome, "_architecture_hijack_v4")

	children := []*ChildSpec{
		{Name: "neuronfs-api", Cmd: nfsExe, Args: []string{brainRoot, "--api"}, Dir: nfsRoot, Enabled: true},
		{Name: "neuronfs-watch", Cmd: nfsExe, Args: []string{brainRoot, "--watch"}, Dir: nfsRoot, Enabled: true},
		{Name: "auto-accept", Cmd: "node", Args: []string{filepath.Join(aaDir, "auto-accept.mjs")}, Dir: aaDir, Enabled: svPathExists(filepath.Join(aaDir, "auto-accept.mjs"))},
		{Name: "agent-bridge", Cmd: "node", Args: []string{filepath.Join(nfsRoot, "runtime", "agent-bridge.mjs")}, Dir: nfsRoot, Enabled: true},
		{Name: "headless-executor", Cmd: "node", Args: []string{filepath.Join(hijackDir, "headless-executor.mjs")}, Dir: hijackDir, Enabled: svPathExists(filepath.Join(hijackDir, "headless-executor.mjs"))},
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

	if nasBrain != "" && svPathExists(nasBrain) {
		go svNasSync(brainRoot, nasBrain, stopCh)
		svLog("🔄 NAS 동기화 활성 (5초)")
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	harnessTk := time.NewTicker(10 * time.Minute)
	statusTk := time.NewTicker(60 * time.Second)
	lockTk := time.NewTicker(5 * time.Second)
	decayTk := time.NewTicker(1 * time.Hour)
	defer harnessTk.Stop()
	defer statusTk.Stop()
	defer lockTk.Stop()
	defer decayTk.Stop()

	// Initial decay run
	go svTTLDecay(brainRoot)

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
		case <-decayTk.C:
			go svTTLDecay(brainRoot)
		}
	}
}

func svParseFrontmatter(content string) (int, time.Time, int) {
	lines := strings.Split(content, "\n")
	if len(lines) < 3 || strings.TrimSpace(lines[0]) != "---" {
		return -1, time.Time{}, -1
	}
	weight := -1
	var lastAct time.Time
	endIdx := len(lines[0]) + 1
	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "---" {
			endIdx += len(lines[i]) + 1
			return weight, lastAct, endIdx
		}
		if strings.HasPrefix(line, "weight:") {
			if w, err := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(line, "weight:"))); err == nil {
				weight = w
			}
		} else if strings.HasPrefix(line, "last_activated:") {
			if t, err := time.Parse(time.RFC3339, strings.TrimSpace(strings.TrimPrefix(line, "last_activated:"))); err == nil {
				lastAct = t
			}
		}
		endIdx += len(lines[i]) + 1
	}
	return -1, time.Time{}, -1
}

func svUpdateWeightFrontmatter(content string, newWeight int) string {
	lines := strings.Split(content, "\n")
	for i, l := range lines {
		if strings.HasPrefix(strings.TrimSpace(l), "weight:") {
			lines[i] = fmt.Sprintf("weight: %d", newWeight)
			break
		}
	}
	return strings.Join(lines, "\n")
}

func svTTLDecay(brainRoot string) {
	// Let's implement TTL decay
	// Note: regionPriority is exported from main.go
	for regionName := range regionPriority {
		regionPath := filepath.Join(brainRoot, regionName)
		if !svPathExists(regionPath) {
			continue
		}
		filepath.Walk(regionPath, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			if !strings.HasSuffix(path, ".neuron") {
				return nil
			}
			contentBytes, err := os.ReadFile(path)
			if err != nil {
				return nil
			}
			content := string(contentBytes)
			weight, lastAct, endIdx := svParseFrontmatter(content)
			
			if weight == -1 || lastAct.IsZero() {
				return nil
			}
			
			if time.Since(lastAct) > 24*time.Hour {
				newWeight := weight - 1
				if newWeight <= 0 {
					archiveDir := filepath.Join(brainRoot, ".archive")
					os.MkdirAll(archiveDir, 0755)
					dest := filepath.Join(archiveDir, filepath.Base(path))
					os.Rename(path, dest)
					svLog(fmt.Sprintf("\033[90m[PRUNE] Synaptic decay complete: %s moved to archive (weight 0)\033[0m", filepath.Base(path)))
					return nil
				}
				
				newFrontmatter := svUpdateWeightFrontmatter(content[:endIdx], newWeight)
				newContent := newFrontmatter + content[endIdx:]
				os.WriteFile(path, []byte(newContent), 0644)
				svLog(fmt.Sprintf("\033[33m[DECAY] Synaptic weight degraded for %s (new weight: %d)\033[0m", filepath.Base(path), newWeight))
			}
			return nil
		})
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

	// Check deadlocks and OOM for the HTTP API (NeuronFS API Server usually binds port 9090)
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
			// Memory Check (tasklist)
			out, err := exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %d", pid), "/FO", "CSV", "/NH").Output()
			if err == nil {
				parts := strings.Split(string(out), "\",\"")
				if len(parts) >= 5 {
					memStr := strings.ReplaceAll(parts[4], "\"", "")
					memStr = strings.ReplaceAll(memStr, " K", "")
					memStr = strings.ReplaceAll(memStr, ",", "")
					memStr = strings.TrimSpace(memStr)
					var memKB int64
					fmt.Sscanf(memStr, "%d", &memKB)
					if memKB > 1024*50 { // 50MB Limit
						svLog("\033[31m[TRAUMA] Synaptic overload detected (Amyloid Plaque > 50MB). Triggering in-memory profile...\033[0m")
						
						// In-Memory Parsing (Zero External Dependency)
						var records []runtime.MemProfileRecord
						n, ok := runtime.MemProfile(nil, true)
						for {
							records = make([]runtime.MemProfileRecord, n+50)
							n, ok = runtime.MemProfile(records, true)
							if ok {
								records = records[:n]
								break
							}
						}
						
						// Sort manually
						for i := 0; i < len(records); i++ {
							for j := i + 1; j < len(records); j++ {
								if records[i].InUseBytes() < records[j].InUseBytes() {
									records[i], records[j] = records[j], records[i]
								}
							}
						}
						
						outbox := filepath.Join(filepath.Dir(svLogPath), "..", "brain_v4", "_agents", "bot1", "outbox")
						if !svPathExists(outbox) {
							outbox = filepath.Join(filepath.Dir(svLogPath), "..", "brain", "_agents", "bot1", "outbox")
						}
						os.MkdirAll(outbox, 0755)
						dumpPath := filepath.Join(outbox, "pprof_heap_dump.txt")
						
						dumpOut := "=== Top 5 Memory Leaks (In-Memory Parsed) ===\n"
						limit := 5
						if len(records) < 5 { limit = len(records) }
						for i := 0; i < limit; i++ {
							r := records[i]
							caller := "unknown"
							if len(r.Stack0) > 0 {
								fn := runtime.FuncForPC(r.Stack0[0])
								if fn != nil { caller = fn.Name() }
							}
							dumpOut += fmt.Sprintf("InUse: %d KB | Objects: %d | Func: %s\n", r.InUseBytes()/1024, r.InUseObjects(), caller)
						}
						
						if err := os.WriteFile(dumpPath, []byte(dumpOut), 0644); err == nil {
							svLog(fmt.Sprintf("\033[35m[DIAG] Saved top 5 heap allocs to %s\033[0m", dumpPath))
						} else {
						    svLog(fmt.Sprintf("\033[33m[WARN] profile write failed: %v\033[0m", err))
						}
						
						// Flatline death screen — OOM visual feedback
						RenderFlatlineOnOOM(c.Name, memKB, dumpOut)

						c.stop()
					}
				}
			}
			
			// Deadlock Check: API Server ping
			if c.Name == "neuronfs-api" {
				client := http.Client{Timeout: 3 * time.Second}
				resp, err := client.Get("http://127.0.0.1:9090/api/health")
				if err != nil {
					svLog("\033[31m[TRAUMA] Synaptic overload detected. Memory integrity compromised.\033[0m")
					c.stop()
				} else if resp != nil {
					resp.Body.Close()
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
