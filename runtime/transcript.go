package main

// ━━━ transcript.go ━━━
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
			fmt.Printf("[GIT] ⚠️ init failed: %v\n", err)
			return
		}
		gitignore := filepath.Join(brainRoot, ".gitignore")
		os.WriteFile(gitignore, []byte("*.dormant\n"), 0600)
		fmt.Printf("[GIT] 📂 Initialized git repo in %s\n", brainRoot)
	}

	// Check for changes
	out, err := SafeOutputDir(ExecTimeoutQuery, brainRoot, "git", "status", "--porcelain")
	if err != nil || len(out) == 0 {
		fmt.Println("[GIT] No changes to snapshot")
		return
	}

	// Stage all
	if err := SafeExecDir(ExecTimeoutGit, brainRoot, "git", "add", "-A"); err != nil {
		fmt.Printf("[GIT] ⚠️ add failed: %v\n", err)
		return
	}

	// Build commit message from current brain state
	brain := scanBrain(brainRoot)
	result := runSubsumption(brain)
	changes := strings.Count(string(out), "\n")
	timestamp := time.Now().Format("01-02 15:04")
	msg := fmt.Sprintf("[%s] %d neurons, act:%d, Δ%d files",
		timestamp, result.TotalNeurons, result.TotalCounter, changes)

	if err := SafeExecDir(ExecTimeoutGit, brainRoot, "git", "commit", "-m", msg, "--no-verify"); err != nil {
		return
	}
	fmt.Printf("[GIT] 📸 %s\n", msg)

	// ── git diff 진화판정: 뉴런 순감소이면 자동 rollback ──
	diffOut, err := SafeOutputDir(ExecTimeoutGit, brainRoot, "git", "diff", "HEAD~1", "--stat")
	if err == nil {
		diffStr := string(diffOut)
		deletions := strings.Count(diffStr, "deletion")
		insertions := strings.Count(diffStr, "insertion")
		if deletions > insertions*2 && deletions > 5 {
			// 삭제가 삽입의 2배 이상이고 5건 초과이면 퇴화로 판정
			fmt.Printf("[GIT] ⚠️ 퇴화 감지 (삭제 %d > 삽입 %d×2) — 자동 rollback\n", deletions, insertions)
			if err := SafeExecDir(ExecTimeoutGit, brainRoot, "git", "revert", "HEAD", "--no-edit"); err != nil {
				fmt.Printf("[GIT] ❌ rollback 실패: %v\n", err)
			} else {
				fmt.Println("[GIT] ✅ 퇴화 commit이 revert 되었습니다")
			}
		} else {
			fmt.Printf("[GIT] ✅ 진화 판정 통과 (ins:%d, del:%d)\n", insertions, deletions)
		}
	}
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// IDLE ENGINE: Auto evolve → snapshot → NAS sync
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

const (
	idleThresholdMinutes = 5  // minutes of no API activity → trigger idle cycle
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
// When idle is detected: digest transcripts → neuronize → evolve → snapshot → NAS sync
func runIdleLoop(brainRoot string) {
	lastEvolveTime := time.Now()

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

		idleEvolveRunning = true
		fmt.Printf("\n[IDLE] 💤 %s idle detected — starting autonomous cycle...\n", idleDuration.Round(time.Second))

		// 0. Transcript Digestion — 전사 파일에서 교정 턴 추출 후 neuronize
		apiKey := os.Getenv("GROQ_API_KEY")
		if apiKey != "" {
			pendingTurns := digestTranscripts(brainRoot)
			if pendingTurns >= 10 {
				fmt.Printf("[IDLE] 🧬 전사 청크 %d턴 누적 — Auto-Neuronize 실행...\n", pendingTurns)
				runNeuronize(brainRoot, false)
			}
		}

		// 1. Evolve — 재활성화 (brainstem/limbic 하드가드 + region 분류 AI 모델 탑재 완료)
		if apiKey != "" {
			fmt.Println("[IDLE] 🧬 Evolve 실행 (region 분류 AI 판단 모델 탑재)...")
			runEvolve(brainRoot, false)
		}

		// 2. Auto-decay (mark neurons untouched for 7+ days as dormant)
		fmt.Println("[IDLE] 💤 Running auto-decay (7 days)...")
		runDecay(brainRoot, 7)

		// 3. Prune: 推-prefix neurons with activation ≤1 and inactive 3+ days → delete
		fmt.Println("[IDLE] 🪦 Running prune (推 low-value cleanup)...")
		pruneWeakNeurons(brainRoot)

		// 4. Consolidate (merge semantically similar neurons, hybrid >= 0.4, counter 합산)
		fmt.Println("[IDLE] 🔀 Running consolidate (hybrid similarity + counter merge)...")
		deduplicateNeurons(brainRoot)

		// 4. Growth tracking (뇌 성장 이력 추적)
		brain := scanBrain(brainRoot)
		result := runSubsumption(brain)
		growthLogDir := filepath.Join(brainRoot, "hippocampus", "session_log")
		os.MkdirAll(growthLogDir, 0750)
		growthLogFile := filepath.Join(growthLogDir, "growth.log")
		entry := fmt.Sprintf("%s: neurons=%d, activation=%d, regions=%d\n",
			time.Now().Format("2006-01-02_15:04"), result.TotalNeurons, result.TotalCounter, len(result.ActiveRegions))
		f, _ := os.OpenFile(growthLogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if f != nil {
			f.WriteString(entry)
			f.Close()
		}
		fmt.Printf("[GROWTH] 📈 %s", entry)

		// 5. Git snapshot
		fmt.Println("[IDLE] 📸 Git snapshot...")
		gitSnapshot(brainRoot)

		// 6. NAS sync (if Z: available)
		nasTarget := `Z:\VOL1\VGVR\BRAIN\LW\system\neurons\brain_v4`
		if _, err := os.Stat(nasTarget); err == nil {
			fmt.Println("[IDLE] 📡 NAS sync...")
			out, err := SafeCombinedOutput(ExecTimeoutSync, "robocopy", brainRoot, nasTarget, "/MIR", "/XD", ".git", "/XF", "*.dormant", "/NFL", "/NDL", "/NP", "/NJH", "/NJS")
			if err != nil {
				// robocopy exit code 1 = files copied (success), only >=8 is error
				if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() >= 8 {
					fmt.Printf("[IDLE] ❌ NAS sync error (exit %d): %s\n", exitErr.ExitCode(), string(out))
				} else {
					fmt.Printf("[IDLE] ✅ NAS synced\n")
				}
			} else {
				fmt.Println("[IDLE] ✅ NAS synced (no changes)")
			}
		} else {
			fmt.Println("[IDLE] ⚠️  NAS Z: not available — skipping sync")
		}

		// 7. Heartbeat 기록
		writeHeartbeat(brainRoot, result)

		// 8. CODE_MAP 자동 갱신 (파일 줄 수 업데이트)
		refreshCodeMap(brainRoot)

		// 9. go vet 자동 검증 (import 동기화 체크)
		runtimeDir := filepath.Join(filepath.Dir(brainRoot), "runtime")
		if _, err := os.Stat(runtimeDir); err == nil {
			if vetOut, err := SafeCombinedOutputDir(ExecTimeoutGit, runtimeDir, "go", "vet", "./..."); err != nil {
				fmt.Printf("[IDLE] ⚠️ go vet 이상 감지: %s\n", strings.TrimSpace(string(vetOut)))
				logEpisode(brainRoot, "VET_FAIL", string(vetOut))
			} else {
				fmt.Println("[IDLE] ✅ go vet 통과")
			}
		}

		lastEvolveTime = time.Now()
		idleEvolveRunning = false
		fmt.Printf("[IDLE] ✅ Autonomous cycle complete at %s\n\n", lastEvolveTime.Format("15:04:05"))
	}
}

// digestTranscripts는 _transcripts/ 파일에서 교정/에러 턴을 추출하여
// corrections_history.jsonl에 추가한다. cursor.json으로 위치 추적.
// 반환값: 이번에 추출된 교정 턴 수
func digestTranscripts(brainRoot string) int {
	transcriptsDir := filepath.Join(brainRoot, "_transcripts")
	cursorPath := filepath.Join(transcriptsDir, ".cursor.json")

	// cursor 읽기
	type cursorEntry struct {
		ByteOffset int64  `json:"byte_offset"`
		LastProc   string `json:"last_processed"`
	}
	cursors := make(map[string]cursorEntry)
	if data, err := os.ReadFile(cursorPath); err == nil {
		json.Unmarshal(data, &cursors)
	}

	// 오늘 날짜 파일
	today := time.Now().Format("2006-01-02") + ".txt"
	todayPath := filepath.Join(transcriptsDir, today)

	info, err := os.Stat(todayPath)
	if err != nil {
		return 0
	}

	cursor := cursors[today]
	if info.Size() <= cursor.ByteOffset {
		return 0 // 새 내용 없음
	}

	// 새 내용 읽기 (cursor 위치부터)
	file, err := os.Open(todayPath)
	if err != nil {
		return 0
	}
	defer file.Close()

	file.Seek(cursor.ByteOffset, 0)
	// 최대 1MB만 읽기 (메모리 절약)
	maxRead := int64(1024 * 1024)
	remaining := info.Size() - cursor.ByteOffset
	if remaining > maxRead {
		remaining = maxRead
	}
	buf := make([]byte, remaining)
	n, _ := file.Read(buf)
	newContent := string(buf[:n])

	// 교정/에러 키워드 필터링
	correctionKeywords := []string{
		"아니", "아냐", "잘못", "다시", "왜 ", "왜또", "안돼", "안됨",
		"에러", "오류", "실패", "멈춤", "404", "500",
		"금지", "하지마", "반드시", "항상", "절대",
	}

	// 감정 감지 키워드 → limbic 자동 fire
	frustrationKeywords := []string{"?", "!!", "아니", "왜", "답답", "안돼", "뭐야", "다시", "허울"}
	satisfactionKeywords := []string{"ㅋㅋ", "좋아", "완벽", "굿", "오", "진행", "맞아"}
	frustrationCount := 0
	satisfactionCount := 0

	lines := strings.Split(newContent, "\n")
	var corrections []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) < 3 {
			continue
		}

		// 감정 감지 (사용자 발화 라인만)
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
		// [HH:MM:SS PD] 패턴이면 사용자 발화
		if !strings.Contains(line, " PD]") && !strings.Contains(line, "교정") {
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

	// 감정 결과 → limbic 뉴런 자동 fire + _state.json 갱신 (EmotionPrompt 자동 전환)
	if frustrationCount >= 3 {
		fireNeuron(brainRoot, "limbic/긴급_사용자답답함감지")
		// Auto-switch emotion to urgent (intensity scales with frustration)
		intensity := 0.5
		if frustrationCount >= 5 {
			intensity = 0.7
		}
		if frustrationCount >= 8 {
			intensity = 0.9
		}
		autoSetEmotion(brainRoot, "긴급", intensity)
		fmt.Printf("[LIMBIC] 😤 답답함 %d회 감지 → 긴급 모드 (intensity: %.1f)\n", frustrationCount, intensity)
	}
	if satisfactionCount >= 3 {
		fireNeuron(brainRoot, "limbic/칭찬_사용자만족감지")
		autoSetEmotion(brainRoot, "만족", 0.6)
		fmt.Printf("[LIMBIC] 😊 만족 %d회 감지 → 만족 모드\n", satisfactionCount)
	}

	// cursor 갱신
	cursor.ByteOffset = cursor.ByteOffset + int64(n)
	cursor.LastProc = time.Now().Format(time.RFC3339)
	cursors[today] = cursor
	cursorData, _ := json.MarshalIndent(cursors, "", "  ")
	os.WriteFile(cursorPath, cursorData, 0600)

	// 교정 턴을 corrections_history.jsonl에 추가
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
		fmt.Printf("[DIGEST] 📝 전사에서 %d건 교정 턴 추출\n", len(corrections))
	}

	return len(corrections)
}

// writeHeartbeat는 idle engine 상태를 _heartbeat.json에 기록하고,
// 뉴런 폭발 감지 시 git snapshot 후 GEMINI.md에 통합 지시를 주입한다.
func writeHeartbeat(brainRoot string, result SubsumptionResult) {
	heartbeatPath := filepath.Join(brainRoot, "_heartbeat.json")

	// 이전 heartbeat 읽기
	prevNeurons := 0
	if prev, err := os.ReadFile(heartbeatPath); err == nil {
		var prevHB map[string]interface{}
		if json.Unmarshal(prev, &prevHB) == nil {
			if n, ok := prevHB["neurons"].(float64); ok {
				prevNeurons = int(n)
			}
		}
	}

	// 현재 상태 기록
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

	// 뉴런 폭발 감지: 20개 이상 증가 시 통합 지시 주입
	growth := result.TotalNeurons - prevNeurons
	if prevNeurons > 0 && growth >= 20 {
		fmt.Printf("[HEARTBEAT] 🔥 뉴런 폭발 감지: %d→%d (+%d) — 통합 지시 주입\n",
			prevNeurons, result.TotalNeurons, growth)

		// 1. git snapshot 선행 (롤백 보장)
		fmt.Println("[HEARTBEAT] 📸 통합 전 git snapshot...")
		gitSnapshot(brainRoot)

		// 2. GEMINI.md에 통합 지시 주입
		directive := fmt.Sprintf(
			"\n\n> [!IMPORTANT]\n> **[HEARTBEAT %s] 뉴런 폭발 감지: %d→%d (+%d)**\n"+
				"> 유사 뉴런 통합이 필요합니다. `neuronfs --dedup` 실행 또는 수동으로 유사 폴더를 병합하세요.\n"+
				"> git snapshot이 선행되었으므로 롤백 가능합니다.\n",
			time.Now().Format("15:04"),
			prevNeurons, result.TotalNeurons, growth)

		// brainstem에 통합 지시 뉴런 생성 (임시)
		consolidateDir := filepath.Join(brainRoot, "brainstem", "뉴런통합_필요")
		os.MkdirAll(consolidateDir, 0750)
		counterFile := filepath.Join(consolidateDir, fmt.Sprintf("%d.neuron", growth))
		os.WriteFile(counterFile, []byte(directive), 0600)

		// writeAllTiers로 GEMINI.md 즉시 갱신
		writeAllTiers(brainRoot)

		fmt.Printf("[HEARTBEAT] ✅ 통합 지시 주입 완료 — brainstem/뉴런통합_필요 생성\n")
	}
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// DEDUP: 중복 뉴런 폴더 병합 (카운터 합산)
// ━━━ Deduplication → lifecycle.go ━━━
// MOVED: deduplicateNeurons

// ━━━ REST API + Rollback → api_server.go ━━━
// MOVED: startAPI, rollbackAll
