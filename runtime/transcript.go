package main

// ?????? transcript.go ??????
// PROVIDES: gitSnapshot, touchActivity, getLastActivity, runIdleLoop, digestTranscripts, writeHeartbeat
// DEPENDS ON: brain.go, lifecycle.go, emit.go, inject.go

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// autoSetEmotion writes emotion state to limbic/_state.json and triggers re-injection.
// Called automatically by transcript digestion when user frustration/satisfaction is detected.
func autoSetEmotion(brainRoot string, emotion string, intensity float64) {
	stateFile := filepath.Join(brainRoot, "limbic", "_state.json")
	os.MkdirAll(filepath.Dir(stateFile), 0750)
	state := map[string]interface{}{
		"emotion":   emotion,
		"intensity": intensity,
		"since":     time.Now().Format("2006-01-02T15:04:05"),
		"trigger":   "auto-transcript",
	}
	data, _ := json.MarshalIndent(state, "", "  ")
	os.WriteFile(stateFile, data, 0600)
	// Trigger re-injection so GEMINI.md picks up the new emotion behavior
	markBrainDirty()
}
func gitSnapshot(brainRoot string) {
	// Check if git is available
	if _, err := exec.LookPath("git"); err != nil {
		fmt.Println("[GIT] git not found, skipping snapshot")
		return
	}

	// Auto-init if not a git repo
	gitDir := filepath.Join(brainRoot, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		if err := SafeExecDir(ExecTimeoutGit, brainRoot, "git", "init"); err != nil {
			fmt.Printf("[GIT] ?? init failed: %v\n", err)
			return
		}
		gitignore := filepath.Join(brainRoot, ".gitignore")
		os.WriteFile(gitignore, []byte("*.dormant\n"), 0600)
		fmt.Printf("[GIT] ?? Initialized git repo in %s\n", brainRoot)
	}

	// Check for changes
	out, err := SafeOutputDir(ExecTimeoutQuery, brainRoot, "git", "status", "--porcelain")
	if err != nil || len(out) == 0 {
		fmt.Println("[GIT] No changes to snapshot")
		return
	}

	// Stage all
	if err := SafeExecDir(ExecTimeoutGit, brainRoot, "git", "add", "-A"); err != nil {
		fmt.Printf("[GIT] ?? add failed: %v\n", err)
		return
	}

	// Build commit message from current brain state
	brain := scanBrain(brainRoot)
	result := runSubsumption(brain)
	changes := strings.Count(string(out), "\n")
	timestamp := time.Now().Format("01-02 15:04")
	msg := fmt.Sprintf("[%s] %d neurons, act:%d, ??%d files",
		timestamp, result.TotalNeurons, result.TotalCounter, changes)

	if err := SafeExecDir(ExecTimeoutGit, brainRoot, "git", "commit", "-m", msg, "--no-verify"); err != nil {
		return
	}
	fmt.Printf("[GIT] ?? %s\n", msg)

	// ???? git diff ???????: ???? ????????? ??? rollback ????
	diffOut, err := SafeOutputDir(ExecTimeoutGit, brainRoot, "git", "diff", "HEAD~1", "--stat")
	if err == nil {
		diffStr := string(diffOut)
		deletions := strings.Count(diffStr, "deletion")
		insertions := strings.Count(diffStr, "insertion")
		if deletions > insertions*2 && deletions > 5 {
			// ?????? ?????? 2?? ?????? 5?? ?????? ????? ????
			fmt.Printf("[GIT] ?? ??? ???? (???? %d > ???? %d??2) ? ??? rollback\n", deletions, insertions)
			if err := SafeExecDir(ExecTimeoutGit, brainRoot, "git", "revert", "HEAD", "--no-edit"); err != nil {
				fmt.Printf("[GIT] ? rollback ????: %v\n", err)
			} else {
				fmt.Println("[GIT] ? ??? commit?? revert ????????")
			}
		} else {
			fmt.Printf("[GIT] ? ??? ???? ??? (ins:%d, del:%d)\n", insertions, deletions)
		}
	}
}

// ??????????????????????????????????????????????????????????????????????????????????????????????????????
// IDLE ENGINE: Auto evolve ?? snapshot ?? NAS sync
// ??????????????????????????????????????????????????????????????????????????????????????????????????????

const (
	idleThresholdMinutes = 5  // minutes of no API activity ?? trigger idle cycle
	idleCheckInterval    = 30 // seconds between idle checks
)

var (
	lastAPIActivity   time.Time
	lastAPIActivityMu sync.Mutex
	idleEvolveRunning bool
)

// touchActivity updates the system's last recorded activity timestamp.
func touchActivity() {
	lastAPIActivityMu.Lock()
	lastAPIActivity = time.Now()
	lastAPIActivityMu.Unlock()
}

// getLastActivity returns the latest timestamp among multiple tracked activity records.
func getLastActivity() time.Time {
	lastAPIActivityMu.Lock()
	defer lastAPIActivityMu.Unlock()
	return lastAPIActivity
}

// runIdleLoop runs in a goroutine, checking for idle state periodically.
// When idle is detected: digest transcripts ?? neuronize ?? evolve ?? snapshot ?? NAS sync
func runIdleLoop(brainRoot string) {
	lastEvolveTime := time.Now().Add(-1 * time.Hour) // ???? ??? u idle ????? ???

	for {
		time.Sleep(time.Duration(idleCheckInterval) * time.Second)

		lastAct := getLastActivity()
		idleDuration := time.Since(lastAct)

		// Need at least idleThresholdMinutes of idle AND at least 30 min since last evolve
		if idleDuration < time.Duration(idleThresholdMinutes)*time.Minute {
			continue
		}
		if time.Since(lastEvolveTime) < 30*time.Minute {
			continue
		}
		if idleEvolveRunning {
			continue
		}

		// ?????? ????????? ?? ??? (? ????? ????)
		if fileExists(filepath.Join(filepath.Dir(brainRoot), "telegram-bridge", ".auto_evolve_disabled")) {
			// ??? ???????? ??? ????? ??????????? ??? ???
			continue
		}

		idleEvolveRunning = true
		fmt.Printf("\n[IDLE] ?? %s idle detected ? calling stateless core worker...\n", idleDuration.Round(time.Second))

		// 전사 자동 백업 (24시간 경과분)
		archiveOldTranscripts(brainRoot)

		nfsExe := filepath.Join(filepath.Dir(brainRoot), "neuronfs.exe")
		out, err := exec.Command(nfsExe, brainRoot, "--tool", "idle_core", "{}").CombinedOutput()
		if err != nil {
			fmt.Printf("[IDLE-WORKER] Error: %v\nOutput: %s\n", err, string(out))
			lastEvolveTime = time.Now()
			idleEvolveRunning = false
			continue
		}

		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		summary := ""
		for _, line := range lines {
			if strings.HasPrefix(line, "[SUMMARY]") {
				summary = strings.TrimPrefix(line, "[SUMMARY] ")
			} else {
				fmt.Println(line)
			}
		}
		if summary == "" {
			summary = "[NeuronFS IDLE] ??? ???? ??? (???? ??? ?? ??? o?? ???)"
		}

		lastEvolveTime = time.Now()
		idleEvolveRunning = false

		// CDP inject ? ????? ???? ???? AI ???? ?????
		injectIdleResult(summary)
	}
}

// injectIdleResult injects heartbeat summary into AI input via CDP.
func injectIdleResult(summary string) {
	// Idle Injection also uses CDPQueue to prevent concurrent CDP access and hangs
	select {
	case CDPQueue <- CDPJob{TargetRoom: "IDLE_INJECT", Payload: summary}:
	default:
		GlobalSSEBroker.Broadcastf("warn", "?? CDP Queue full! Dropping idle result.")
	}
}

// injectIdleResultSync performs the actual self-healing and injection
func injectIdleResultSync(summary string) {
	escaped := strings.ReplaceAll(summary, `"`, `\"`)
	escaped = strings.ReplaceAll(escaped, "\n", "\\n")

	script := fmt.Sprintf(`(() => {
		const all = Array.from(document.querySelectorAll("[contenteditable]"));
		const el = all.reverse().find(e => {
			const r = e.getBoundingClientRect();
			return r.height > 0 && r.height < 300 && r.width > 100;
		}) || all[0];
		if (el) {
			el.focus();
			document.execCommand("insertText", false, "%s");
			el.dispatchEvent(new KeyboardEvent("keydown", {key:"Enter",code:"Enter",keyCode:13,which:13,bubbles:true}));
			return "Injected";
		}
		return "NoTarget";
	})()`, escaped)

	// 1??: aaAgents (auto-accept ????)
	injected := false
	aaAgents.Range(func(k, v interface{}) bool {
		a := v.(*aaAgent)
		result, err := a.client.Call("Runtime.evaluate", map[string]interface{}{
			"expression":    script,
			"returnByValue": true,
		})
		if err != nil {
			// ???????: ???? ???? ????
			fmt.Printf("[IDLE] ?? aaAgent %s ???? ???? ? ????\n", a.name)
			aaAgents.Delete(k)
			return true
		}
		var evalRes struct {
			Result struct {
				Value string `json:"value"`
			} `json:"result"`
		}
		json.Unmarshal(result, &evalRes)
		if evalRes.Result.Value == "Injected" {
			fmt.Printf("[IDLE] ?? CDP ?????? ???? ?? [%s]\n", a.name)
			injected = true
			return false
		}
		return true
	})
	if injected {
		return
	}

	// 2??: ??????? ? aaAgents ??? + ???? CDP (3? ??o?)
	var diagLog []string
	for retry := 0; retry < 3; retry++ {
		if retry > 0 {
			diagLog = append(diagLog, fmt.Sprintf("??o? %d/3 (5?? ???)", retry+1))
			time.Sleep(5 * time.Second)
		}

		// ???????: aaScanTargets ???? ???????? ?? ?? ????
		aaScanTargets()

		// ???? CDP ????
		targets, err := cdpListTargets(CDPPort)
		if err != nil {
			diagLog = append(diagLog, fmt.Sprintf("CDP:9000 ??? ? %v", err))
			continue
		}

		workbenchFound := false
		for _, t := range targets {
			if !strings.Contains(t.URL, "workbench.html") {
				continue
			}
			workbenchFound = true
			if t.WebSocketDebuggerURL == "" {
				diagLog = append(diagLog, fmt.Sprintf("wsURL ???? (title=%s)", t.Title))
				continue
			}

			client, cErr := NewCDPClient(t.WebSocketDebuggerURL)
			if cErr != nil {
				diagLog = append(diagLog, fmt.Sprintf("WS ???? ? %v", cErr))
				continue
			}
			client.Call("Runtime.enable", map[string]interface{}{})
			time.Sleep(300 * time.Millisecond)
			r, rErr := client.Call("Runtime.evaluate", map[string]interface{}{
				"expression":    script,
				"returnByValue": true,
			})
			client.Close()

			if rErr != nil {
				diagLog = append(diagLog, fmt.Sprintf("evaluate ???? ? %v", rErr))
				continue
			}
			if r != nil {
				var er struct {
					Result struct {
						Value string `json:"value"`
					} `json:"result"`
				}
				json.Unmarshal(r, &er)
				if er.Result.Value == "Injected" {
					fmt.Printf("[IDLE] ? ??????? ???? (?o? %d)\n", retry+1)
					return
				}
				if er.Result.Value == "NoTarget" {
					diagLog = append(diagLog, "contenteditable ???? ? ???a ?????")
				}
			}
		}
		if !workbenchFound {
			diagLog = append(diagLog, "workbench.html ???? ? IDE ???")
		}
	}

	// 3? ????: ???? ????? ?????????? ???? + IDE?? ?????? ?o?
	diagStr := strings.Join(diagLog, " | ")
	fmt.Printf("[IDLE] ? CDP 3? ????: %s\n", diagStr)
	GlobalSSEBroker.Broadcastf("error", "[IDLE] CDP 3? ????: %s", diagStr)

	// ?????? ???
	if hlTgToken != "" && hlTgChatID != "" {
		hlTgSend(hlTgChatID, fmt.Sprintf("?? [??????? ????] %s\n?????: %s", summary, diagStr))
	}

	// ???? ????: hlCDPInject?? ????? ??? ?????? (??? ?????)
	debugMsg := fmt.Sprintf("[NeuronFS ??????? ????] %s | ????: %s ? ???? ???? ?? ???? ????", summary, diagStr)
	go hlCDPInject(hlTgMountedRoom, debugMsg)
}

// detectFailedEvolutions scans for neurons inactive 30+ days.
// Only ?? (recommendation) neurons are checked ? ??/?? are immune.
// Returns list of paths considered "failed evolution" attempts.
func detectFailedEvolutions(brainRoot string) []string {
	var failed []string
	cutoff := time.Now().AddDate(0, 0, -30)

	regions := []string{"cortex", "sensors", "ego", "prefrontal"}
	for _, region := range regions {
		regionDir := filepath.Join(brainRoot, region)
		filepath.Walk(regionDir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info == nil || info.IsDir() {
				return nil
			}
			if !strings.HasSuffix(info.Name(), ".neuron") {
				return nil
			}
			// ??/?? ??
			if strings.Contains(path, "??") || strings.Contains(path, "??") {
				return nil
			}
			// ?? ?????? + 30?? ?????
			if strings.Contains(path, "??") && info.ModTime().Before(cutoff) {
				rel, _ := filepath.Rel(brainRoot, path)
				failed = append(failed, rel)
			}
			return nil
		})
	}
	return failed
}

// digestTranscripts?? _transcripts/ ??????? ????/???? ???? ???????
// corrections_history.jsonl?? ??????. cursor.json???? ??? ????.
// ?????: ????? ????? ???? ?? ??
func digestTranscripts(brainRoot string) int {
	transcriptsDir := filepath.Join(brainRoot, "_transcripts")
	cursorPath := filepath.Join(transcriptsDir, ".cursor.json")

	// cursor ????
	type cursorEntry struct {
		ByteOffset int64  `json:"byte_offset"`
		LastProc   string `json:"last_processed"`
	}
	cursors := make(map[string]cursorEntry)
	if data, err := os.ReadFile(cursorPath); err == nil {
		json.Unmarshal(data, &cursors)
	}

	// ???? ???? ????
	today := time.Now().Format("2006-01-02") + ".txt"
	todayPath := filepath.Join(transcriptsDir, today)

	info, err := os.Stat(todayPath)
	if err != nil {
		return 0
	}

	cursor := cursors[today]
	if info.Size() < cursor.ByteOffset {
		cursor.ByteOffset = 0
	}
	if info.Size() <= cursor.ByteOffset {
		return 0
	}

	// ?? ???? ???? (cursor ???????)
	file, err := os.Open(todayPath)
	if err != nil {
		return 0
	}
	defer file.Close()

	file.Seek(cursor.ByteOffset, 0)
	// ??? 1MB?? ???? (??? ????)
	maxRead := int64(1024 * 1024)
	remaining := info.Size() - cursor.ByteOffset
	if remaining > maxRead {
		remaining = maxRead
	}
	buf := make([]byte, remaining)
	n, _ := file.Read(buf)
	newContent := string(buf[:n])

	// ????/???? ????? ?????
	correctionKeywords := []string{
		"???", "???", "???", "???", "?? ", "???", "???", "???",
		"????", "????", "????", "????", "404", "500",
		"????", "??????", "????", "???", "????",
	}

	// ???? ???? ????? ?? limbic ??? fire
	frustrationKeywords := []string{"?", "!!", "???", "??", "???", "???", "????", "???", "???"}
	satisfactionKeywords := []string{"????", "????", "???", "??", "??", "????", "????"}
	frustrationCount := 0
	satisfactionCount := 0

	lines := strings.Split(newContent, "\n")
	var corrections []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) < 3 {
			continue
		}

		// ???? ???? (????? ??? ??????)
		if strings.Contains(line, " PD]") {
			for _, kw := range frustrationKeywords {
				if strings.Contains(line, kw) {
					frustrationCount++
					break
				}
			}
			for _, kw := range satisfactionKeywords {
				if strings.Contains(line, kw) {
					satisfactionCount++
					break
				}
			}
		}

		if len(line) < 10 {
			continue
		}
		// [HH:MM:SS PD] ??????? ????? ???
		if !strings.Contains(line, " PD]") && !strings.Contains(line, "????") {
			continue
		}
		for _, kw := range correctionKeywords {
			if strings.Contains(line, kw) {
				if len(line) > 300 {
					line = line[:300]
				}
				corrections = append(corrections, line)
				break
			}
		}
	}

	// ???? ??? ?? limbic ???? ??? fire + _state.json ???? (EmotionPrompt ??? ???)
	if frustrationCount >= 3 {
		fireNeuron(brainRoot, "limbic/???_????????????")
		// Auto-switch emotion to urgent (intensity scales with frustration)
		intensity := 0.5
		if frustrationCount >= 5 {
			intensity = 0.7
		}
		if frustrationCount >= 8 {
			intensity = 0.9
		}
		autoSetEmotion(brainRoot, "???", intensity)
		fmt.Printf("[LIMBIC] ?? ????? %d? ???? ?? ??? ??? (intensity: %.1f)\n", frustrationCount, intensity)
		GlobalSSEBroker.Broadcastf("warn", "[LIMBIC] ????? %d? ???? ?? ??? ??? (intensity: %.1f)", frustrationCount, intensity)
	}
	if satisfactionCount >= 3 {
		fireNeuron(brainRoot, "limbic/???_????????????")
		autoSetEmotion(brainRoot, "????", 0.6)
		fmt.Printf("[LIMBIC] ?? ???? %d? ???? ?? ???? ???\n", satisfactionCount)
		GlobalSSEBroker.Broadcastf("success", "[LIMBIC] ???? %d? ???? ?? ???? ???", satisfactionCount)
	}

	// cursor ????
	cursor.ByteOffset = cursor.ByteOffset + int64(n)
	cursor.LastProc = time.Now().Format(time.RFC3339)
	cursors[today] = cursor
	cursorData, _ := json.MarshalIndent(cursors, "", "  ")
	os.WriteFile(cursorPath, cursorData, 0600)

	// ???? ???? corrections_history.jsonl?? ???
	if len(corrections) > 0 {
		historyPath := filepath.Join(brainRoot, "_inbox", "corrections_history.jsonl")
		f, err := os.OpenFile(historyPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err == nil {
			for _, c := range corrections {
				entry := fmt.Sprintf(`{"type":"transcript_correction","text":"%s","time":"%s"}`,
					strings.ReplaceAll(c, `"`, `\"`),
					time.Now().Format(time.RFC3339))
				f.WriteString(entry + "\n")
			}
			f.Close()
		}
		fmt.Printf("[DIGEST] ?? ??????? %d?? ???? ?? ????\n", len(corrections))
	}

	return len(corrections)
}

// writeHeartbeat?? idle engine ?????? _heartbeat.json?? ??????,
// ???? ???? ???? ?? git snapshot ?? GEMINI.md?? ???? ?????? ???????.
func writeHeartbeat(brainRoot string, result SubsumptionResult) {
	heartbeatPath := filepath.Join(brainRoot, "_heartbeat.json")

	// ???? heartbeat ????
	prevNeurons := 0
	if prev, err := os.ReadFile(heartbeatPath); err == nil {
		var prevHB map[string]interface{}
		if json.Unmarshal(prev, &prevHB) == nil {
			if n, ok := prevHB["neurons"].(float64); ok {
				prevNeurons = int(n)
			}
		}
	}

	// ???? ???? ???
	hb := map[string]interface{}{
		"last_cycle":   time.Now().Format(time.RFC3339),
		"neurons":      result.TotalNeurons,
		"activation":   result.TotalCounter,
		"regions":      len(result.ActiveRegions),
		"prev_neurons": prevNeurons,
		"growth_delta": result.TotalNeurons - prevNeurons,
	}
	data, _ := json.MarshalIndent(hb, "", "  ")
	os.WriteFile(heartbeatPath, data, 0600)

	// ???? ???? ????: 20?? ??? ???? ?? ???? ???? ????
	growth := result.TotalNeurons - prevNeurons
	if prevNeurons > 0 && growth >= 20 {
		fmt.Printf("[HEARTBEAT] ?? ???? ???? ????: %d??%d (+%d) ? ???? ???? ????\n",
			prevNeurons, result.TotalNeurons, growth)
		GlobalSSEBroker.Broadcastf("info", "[HEARTBEAT] ???? ???? ????: %d??%d (+%d) ? ???? ???? ????", prevNeurons, result.TotalNeurons, growth)

		// 1. git snapshot ???? (??? ????)
		fmt.Println("[HEARTBEAT] ?? ???? ?? git snapshot...")
		gitSnapshot(brainRoot)

		// 2. GEMINI.md?? ???? ???? ????
		directive := fmt.Sprintf(
			"\n\n> [!IMPORTANT]\n> **[HEARTBEAT %s] ???? ???? ????: %d??%d (+%d)**\n"+
				"> ???? ???? ?????? ???????. `neuronfs --dedup` ???? ??? ???????? ???? ?????? ?????????.\n"+
				"> git snapshot?? ??????????? ??? ????????.\n",
			time.Now().Format("15:04"),
			prevNeurons, result.TotalNeurons, growth)

		// brainstem?? ???? ???? ???? ???? (???)
		consolidateDir := filepath.Join(brainRoot, "brainstem", "????????_???")
		os.MkdirAll(consolidateDir, 0750)
		counterFile := filepath.Join(consolidateDir, fmt.Sprintf("%d.neuron", growth))
		os.WriteFile(counterFile, []byte(directive), 0600)

		// writeAllTiers?? GEMINI.md ??? ????
		writeAllTiers(brainRoot)

		fmt.Printf("[HEARTBEAT] ? ???? ???? ???? ??? ? brainstem/????????_??? ????\n")
	}
}

// ??????????????????????????????????????????????????????????????????????????????????????????????????????
// DEDUP: ??? ???? ???? ???? (????? ???)
// ?????? Deduplication ?? lifecycle.go ??????
// MOVED: deduplicateNeurons

// ?????? REST API + Rollback ?? api_server.go ??????
// MOVED: startAPI, rollbackAll


// ──────────────────────────────────────────────────────────
// TRANSCRIPT ARCHIVE: 24시간 경과한 라이브 전사를 자동 백업 (적층)
// ──────────────────────────────────────────────────────────

// archiveOldTranscripts moves transcript files older than 24h
// from _transcripts/ to _transcripts/_backup_YYYYMMDD/.
// This ensures no live transcript is lost due to rotation.
// Called from runIdleLoop. Existing transcripts are NEVER deleted.
func archiveOldTranscripts(brainRoot string) int {
	transcriptsDir := filepath.Join(brainRoot, "_transcripts")
	cutoff := time.Now().Add(-24 * time.Hour)
	moved := 0

	entries, err := os.ReadDir(transcriptsDir)
	if err != nil {
		return 0
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".txt") {
			continue
		}

		info, err := e.Info()
		if err != nil || info.ModTime().After(cutoff) {
			continue
		}

		dateStr := ""
		parts := strings.Split(name, "_")
		for _, p := range parts {
			if len(p) == 10 && strings.Count(p, "-") == 2 {
				dateStr = strings.ReplaceAll(p, "-", "")
				break
			}
		}
		if dateStr == "" {
			dateStr = info.ModTime().Format("20060102")
		}

		backupDir := filepath.Join(transcriptsDir, "_backup_"+dateStr)
		os.MkdirAll(backupDir, 0750)

		src := filepath.Join(transcriptsDir, name)
		dst := filepath.Join(backupDir, name)

		if fileExists(dst) {
			continue
		}

		data, err := os.ReadFile(src)
		if err != nil {
			continue
		}
		if err := os.WriteFile(dst, data, 0600); err != nil {
			continue
		}
		dstInfo, err := os.Stat(dst)
		if err != nil || dstInfo.Size() != info.Size() {
			os.Remove(dst)
			continue
		}
		os.Remove(src)
		moved++
		fmt.Printf("[ARCHIVE] 📦 %s → %s\n", name, "_backup_"+dateStr)
	}

	if moved > 0 {
		fmt.Printf("[ARCHIVE] ✅ %d건 전사 백업 완료\n", moved)
	}
	return moved
}
