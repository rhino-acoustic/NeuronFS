// RULE: ZERO EXTERNAL DEPENDENCY PRESERVED
// supervisor.go — NeuronFS 네이티브 프로세스 매니저
//
// watchdog.ps1 + 프로세스 관리를 Go 바이너리로 통합.
//
// Usage: neuronfs <brain_path> --supervisor
package main

import (
	"fmt"
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
		// 1단계: Graceful Shutdown 시도 (os.Interrupt)
		c.proc.Process.Signal(os.Interrupt)
		
		// 2단계: 최대 3초 대기
		done := make(chan error, 1)
		go func() {
			done <- c.proc.Wait()
		}()

		select {
		case <-done:
			// 정상 종료됨
		case <-time.After(3 * time.Second):
			// 3단계: 시간 초과 시 강제 종료
			c.proc.Process.Kill()
		}
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

// ── 메트릭 JSON 출력 ──
var svBootTime = time.Now()
var svCheckCount int

// ── Antigravity 자동 탐색 (어느 컴퓨터든 동작) ──

// ── 부트스트랩 (bat에서 마이그레이션) ──
func svBootstrap(nfsRoot string) {
	// 1. 좀비 프로세스 정리 (기존 node hijacker + 이전 neuronfs 인스턴스)
	// (a) 자신을 제외한 모든 고스트/중복 슈퍼바이저 프로세스 처단
	exeName := filepath.Base(os.Args[0])
	currentPID := os.Getpid()
	killGhostsCmd := fmt.Sprintf(`taskkill /F /IM "%s" /FI "PID ne %d"`, exeName, currentPID)
	exec.Command("cmd", "/c", killGhostsCmd).Run()
	svLog(fmt.Sprintf("🧹 중복 슈퍼바이저/고스트 프로세스 정리 완료 (PID: %d 생존)", currentPID))

	// (b) 레거시 노드 좀비
	zombieTargets := []string{"node.exe", "hijack-launcher"}
	for _, target := range zombieTargets {
		if killProcessByName(target) {
			svLog(fmt.Sprintf("🧹 구형 좀비 정리: %s", target))
		}
	}

	// 2. 로그 로테이션 (100MB 초과 삭제)
	logDir := filepath.Join(nfsRoot, "logs")
	entries, err := os.ReadDir(logDir)
	if err == nil {
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			info, err := e.Info()
			if err != nil {
				continue
			}
			if info.Size() > 100*1024*1024 { // 100MB
				target := filepath.Join(logDir, e.Name())
				os.Remove(target)
				svLog(fmt.Sprintf("🗑️ 로그 로테이션: %s (%.1fMB)", e.Name(), float64(info.Size())/(1024*1024)))
			}
		}
	}

	// 3. GROQ_API_KEY 체크
	if os.Getenv("GROQ_API_KEY") == "" {
		svLog("⚠️ GROQ_API_KEY 미설정 — Groq 배치 분석 비활성")
	} else {
		svLog("✅ GROQ_API_KEY 확인됨")
	}

	// 4. NODE_OPTIONS 정리 (Go 네이티브라 불필요하지만, 잔재 환경변수 경고)
	if os.Getenv("NODE_OPTIONS") != "" {
		svLog("⚠️ NODE_OPTIONS 환경변수 감지 — Go 네이티브에서 불필요")
	}

	// 5. Antigravity 바로가기에 CDP 플래그 자동 적용 (순서 무관)
	svPatchAntigravityShortcuts()

	// 6. 재시작 컨텍스트 감지 (159487 코드 재시작)
	restartCtxPath := filepath.Join(nfsRoot, ".restart_context")
	if fileExists(restartCtxPath) {
		svLog("🔄 재시작 컨텍스트 감지 — 대화 복귀 예약")
	}
}

func runSupervisor(brainRoot string) {
	nfsRoot := filepath.Dir(brainRoot)
	nfsExe, _ := os.Executable()
	logDir := filepath.Join(nfsRoot, "logs")
	os.MkdirAll(logDir, 0750)
	svLogPath = filepath.Join(logDir, "supervisor.log")
	nasBrain := os.Getenv("NEURONFS_NAS_BRAIN")

	fmt.Println("")
	fmt.Println("╔══════════════════════════════════════════════════╗")
	fmt.Println("║  NeuronFS Supervisor v2.2 — Full Migration       ║")
	fmt.Println("║  부트스트랩 + 프로세스 관리 + 자기 감시           ║")
	fmt.Println("╚══════════════════════════════════════════════════╝")
	fmt.Println("")

	// ★ 부트스트랩 (bat에서 마이그레이션 완료)
	svBootstrap(nfsRoot)

	// Antigravity 자동 탐색 (어느 컴퓨터든 동작)
	agExe := findAntigravity()
	agEnabled := agExe != ""
	if os.Getenv("NEURONFS_NO_AG") != "" {
		svLog("ℹ️ NEURONFS_NO_AG 설정 — Antigravity 실행 안 함 (인프라 전용)")
		agEnabled = false
	} else if agEnabled && isProcessRunning("Antigravity.exe") {
		svLog("ℹ️ Antigravity 이미 실행 중 — 새 인스턴스 생략")
		agEnabled = false
	}
	if agEnabled {
		svLog(fmt.Sprintf("🔍 Antigravity 발견: %s", agExe))
	} else if agExe == "" {
		svLog("⚠️ Antigravity 미설치 — CDP 기능 비활성")
	}

	// 워크스페이스: 환경변수 우선, 없으면 nfsRoot
	agWorkspace := os.Getenv("NEURONFS_AG_WORKSPACE")
	if agWorkspace == "" {
		agWorkspace = nfsRoot
	}

	children := []*ChildSpec{
		{Name: "antigravity", Cmd: agExe, Args: []string{agWorkspace, "--remote-debugging-port=9000"}, Dir: agWorkspace, Enabled: agEnabled},
		{Name: "neuronfs-api", Cmd: nfsExe, Args: []string{brainRoot, "--api"}, Dir: nfsRoot, Enabled: true},
		{Name: "neuronfs-watch", Cmd: nfsExe, Args: []string{brainRoot, "--watch"}, Dir: nfsRoot, Enabled: true},
		// 전체 Node.js 데몬 Go 네이티브 전환 완료
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

	// supervisor 자신과 자식 프로세스를 health에 보고
	markServiceRunning("supervisor", true)
	for _, c := range children {
		if c.Enabled {
			// "neuronfs-api" → "api", "neuronfs-watch" → "watch"
			svcName := strings.TrimPrefix(c.Name, "neuronfs-")
			markServiceRunning(svcName, true)
		}
	}

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
	go func() {
		markServiceRunning("context-hijacker", true)
		runContextHijacker(brainRoot)
	}()
	svLog("📡 Context Hijacker (Go native) 시작")

	// Go 네이티브 auto-accept (Node 대체)
	go func() {
		markServiceRunning("auto-accept", true)
		runMacroWorker(brainRoot)
	}()
	svLog("🖱️ Auto-Accept (Go native) 시작")

	// Go 네이티브 agent-bridge (Node 대체)
	go func() {
		markServiceRunning("agent-bridge", true)
		runAgentBridge(brainRoot)
	}()
	svLog("📨 Agent Bridge (Go native) 시작")

	// Go 네이티브 headless-executor (Node 대체)
	go func() {
		markServiceRunning("headless-executor", true)
		runHeadlessExecutor(brainRoot)
	}()
	svLog("⚡ Headless Executor (Go native) 시작")

	// Go 네이티브 hijack-launcher (Node 대체 — 마지막)
	go func() {
		markServiceRunning("hijack-launcher", true)
		runHijackLauncher(brainRoot)
	}()
	svLog("🚀 Hijack Launcher (Go native) 시작")

	// ── 159487 재시작 후 대화 복귀 ──
	go svRestoreConversation(nfsRoot, brainRoot)

	// Go 네이티브 챗 히스토리 보호 (state.vscdb 감시)
	go svChatHistoryGuard()
	svLog("🛡️ Chat History Guard 시작")

	// Go 네이티브 MCP Streamable HTTP (stdio 대체 — IDE 재시작에도 생존)
	go superviseMCPGoroutine(brainRoot, MCPStreamPort)
	svLog(fmt.Sprintf("🌐 MCP Streamable HTTP 시작 (:%d)", MCPStreamPort))

	// Go 네이티브 Backlog Classifier (inbox 자동 분류)
	go RunBacklogLoop(brainRoot, svLog)
	svLog("📋 Backlog Classifier 시작 (5분 주기)")

	svBootTime = time.Now()
	hlLoadTelegram(nfsRoot)
	logBootEpisode(brainRoot)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	harnessTk := time.NewTicker(10 * time.Minute)
	statusTk := time.NewTicker(60 * time.Second)
	lockTk := time.NewTicker(5 * time.Second)
	decayTk := time.NewTicker(1 * time.Hour)
	reportTk := time.NewTicker(30 * time.Minute)
	ltpTk := time.NewTicker(12 * time.Hour)
	defer harnessTk.Stop()
	defer statusTk.Stop()
	defer lockTk.Stop()
	defer decayTk.Stop()
	defer reportTk.Stop()
	defer ltpTk.Stop()

	// Initial decay run
	go RunTTLDecay(brainRoot, svLog)
	
	// Initial LTP run (Memory Defragmentation)
	go RunLTP(brainRoot, svLog)

	go func() {
		rebootFile := filepath.Join(nfsRoot, "dist", "release", "_reboot_request")
		for {
			time.Sleep(2 * time.Second)
			if fileExists(rebootFile) {
				svLog("🔥 _reboot_request 감지! 현 프로세스 강제 종료 (Hot Reload 주도권 이양)...")
				os.Exit(0)
			}
		}
	}()

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
		case <-ltpTk.C:
			go RunLTP(brainRoot, svLog)
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
			if svLogPath != "" {
				nfsR := filepath.Dir(filepath.Dir(svLogPath))
				if br := filepath.Join(nfsR, "brain_v4"); fileExists(br) {
					logCrashEpisode(br, c.Name, c.restartCount)
				}
			}
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

		// Antigravity: 이미 실행 중이면 시작하지 않고 대기
		if c.Name == "antigravity" && isProcessRunning("Antigravity.exe") {
			c.mu.Lock()
			c.running = false
			c.mu.Unlock()
			time.Sleep(30 * time.Second)
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
				client := http.Client{Timeout: 5 * time.Second}
				resp, err := client.Get(fmt.Sprintf("http://127.0.0.1:%d/api/ping", APIPort)) // 가벼운 liveness
				if err != nil {
					svLog("\033[33m[WARN] API liveness check failed — restarting\033[0m")
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

	// L3: MCP Streamable HTTP 헬스체크
	svCheckMCPHealth()
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

// ── 챗 히스토리 보호 (state.vscdb 감시 + 자동 복원) ──

// ── 159487 재시작 후 대화 복귀 ──

// svPatchAntigravityShortcuts — Antigravity 바로가기에 CDP 플래그 자동 주입
// 시작 메뉴 + 작업표시줄 + 데스크탑의 .lnk를 모두 찾아 --remote-debugging-port=9000 추가
func svPatchAntigravityShortcuts() {
	cdpFlag := "--remote-debugging-port=9000"
	home := os.Getenv("USERPROFILE")
	if home == "" {
		return
	}

	searchDirs := []string{
		filepath.Join(home, "Desktop"),
		filepath.Join(home, "AppData", "Roaming", "Microsoft", "Windows", "Start Menu", "Programs"),
		filepath.Join(home, "AppData", "Roaming", "Microsoft", "Internet Explorer", "Quick Launch", "User Pinned", "TaskBar"),
	}

	patched := 0
	for _, dir := range searchDirs {
		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			if !strings.Contains(strings.ToLower(info.Name()), "antigravity") || !strings.HasSuffix(strings.ToLower(info.Name()), ".lnk") {
				return nil
			}
			// .lnk 파일은 바이너리 — 단순 바이트 검사로 CDP 플래그 존재 여부만 확인
			data, err := os.ReadFile(path)
			if err != nil {
				return nil
			}
			if strings.Contains(string(data), cdpFlag) {
				return nil // 이미 패치됨
			}
			svLog(fmt.Sprintf("⚠️ .lnk CDP 미설정: %s (수동 설정 필요)", filepath.Base(path)))
			patched++
			return nil
		})
	}
	if patched == 0 {
		svLog("✅ Antigravity 바로가기 CDP 설정 확인 완료")
	}
}

func svRestoreConversation(nfsRoot, brainRoot string) {
	ctxPath := filepath.Join(nfsRoot, ".restart_context")
	data, err := os.ReadFile(ctxPath)
	if err != nil {
		return // 재시작 컨텍스트 없음 — 정상
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

	svLog(fmt.Sprintf("🔄 재시작 컨텍스트 감지: room=%s title=%s", ctx.Room, ctx.Title))

	// 10초 대기 — Antigravity 완전 로드
	time.Sleep(10 * time.Second)

	// CDP로 이전 대화 복원 시도
	room := ctx.Room
	if room == "" {
		room = "NeuronFS"
	}
	msg := fmt.Sprintf("이전 대화 '%s'에서 작업을 이어갑니다. 컨텍스트를 확인하고 계속하세요.", ctx.Title)
	hlCDPInject(room, msg)

	svLog(fmt.Sprintf("✅ 대화 복귀 주입 완료: %s", ctx.Title))

	// 컨텍스트 파일 소비 완료 — 삭제
	os.Remove(ctxPath)
}
// ── MCP Goroutine Supervisor (panic recovery + auto-restart) ──
var (
	mcpBrainRoot string
	mcpPort      int
	mcpRestarts  int
)

func superviseMCPGoroutine(brainRoot string, port int) {
	mcpBrainRoot = brainRoot
	mcpPort = port
	workerPort := port + 1 // e.g. 9248

	// 0. 소스코드 감시기 가동 (Hot-Reload 자동 빌드 데몬)
	go runGoSourceWatcher()

	// 1. 역방향 프록시 계층 구동 (클라이언트 연결 쉴드)
	go runMCPProxy(port, workerPort)

	// 2. 실제 MCP Worker 엔진 구동
	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					mcpRestarts++
					svLog(fmt.Sprintf("\033[31m[MCP-HEAL] panic recovered: %v (restart #%d)\033[0m", r, mcpRestarts))
				}
			}()
			markServiceRunning("mcp-http", true)
			startMCPHTTPServer(brainRoot, workerPort)
		}()

		// startMCPHTTPServer returned (shouldn't normally) — restart
		markServiceRunning("mcp-http", false)
		mcpRestarts++
		delay := time.Duration(mcpRestarts) * 2 * time.Second
		if delay > 30*time.Second {
			delay = 30 * time.Second
		}
		svLog(fmt.Sprintf("\033[36m[MCP-HEAL] restarting in %v (attempt #%d)\033[0m", delay, mcpRestarts))
		time.Sleep(delay)
	}
}

// svCheckMCPHealth checks MCP /mcp/health endpoint. If zombie (port open but no response),
// logs warning. The goroutine wrapper handles restart on crash.
var mcpFails int

func svCheckMCPHealth() {
	if mcpPort == 0 {
		return
	}
	client := http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(fmt.Sprintf("http://127.0.0.1:%d/mcp/health", mcpPort))
	if err != nil {
		mcpFails++
		if mcpFails >= 3 {
			svLog("\033[31m[MCP-FATAL] MCP 서버 L7 헬스체크 3연속 실패 (Session Lost / Deadlock). Auto-Healing을 위해 자폭(Restart)합니다.\033[0m")
			os.Exit(2)
		}
		// Check if port is still listening (zombie detection)
		if checkPort(mcpPort) {
			svLog(fmt.Sprintf("\033[33m[MCP-ZOMBIE] port :%d open but /mcp/health timeout (%d/3)\033[0m", mcpPort, mcpFails))
		} else {
			svLog(fmt.Sprintf("\033[31m[MCP-DOWN] port :%d not listening (%d/3)\033[0m", mcpPort, mcpFails))
		}
		return
	}
	resp.Body.Close()
	mcpFails = 0 // Session alive, reset counter
	// MCP alive — reset restart counter on sustained health
	if mcpRestarts > 0 {
		mcpRestarts = 0
	}
}
