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

func runIdleCoreWorker(brainRoot string) {
	// Community Best Practice: Panic Recovery in Stateless Workers
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("[IDLE-WORKER] 🚨 워커 패닉 발생 (Recovered): %v\n", r)
			fmt.Printf("[SUMMARY] [NeuronFS IDLE] ⚠️ 시스템 에러 감지. 유휴 워커 패닉 발생: %v\n", r)
		}
	}()

	fmt.Println("[IDLE-WORKER] 💤 stateless idle core execution started...")

	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey != "" {
		pendingTurns := digestTranscripts(brainRoot)
		if pendingTurns >= 10 {
			fmt.Printf("[IDLE-WORKER] 🧬 전사 청크 %d턴 누적 — Auto-Neuronize 실행...\n", pendingTurns)
			runNeuronize(brainRoot, false)
		}
	}

	failedEvolutions := detectFailedEvolutions(brainRoot)
	if len(failedEvolutions) > 0 {
		growthLogFile := filepath.Join(brainRoot, "hippocampus", "session_log", "growth.log")
		f, _ := os.OpenFile(growthLogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if f != nil {
			for _, fe := range failedEvolutions {
				f.WriteString(fmt.Sprintf("%s: META_EVOLVE_FAIL path=%s (30d+ inactive)\n",
					time.Now().Format("2006-01-02_15:04"), fe))
			}
			f.Close()
		}
		fmt.Printf("[IDLE-WORKER] 🧬 메타진화: %d개 실패한 진화 감지\n", len(failedEvolutions))
	}

	if apiKey != "" {
		fmt.Println("[IDLE-WORKER] 🧬 Evolve 실행 (region 분류 AI 판단 모델 탑재)...")
		runEvolve(brainRoot, false)
	}

	fmt.Println("[IDLE-WORKER] 💤 Running auto-decay (7 days)...")
	runDecay(brainRoot, 7)

	fmt.Println("[IDLE-WORKER] 🪦 Running prune (推 low-value cleanup)...")
	pruneWeakNeurons(brainRoot)

	fmt.Println("[IDLE-WORKER] ♻️ Spaced repetition (reinforce high-value neurons)...")
	spacedRepetitionFire(brainRoot)

	fmt.Println("[IDLE-WORKER] 🔀 Running consolidate (hybrid similarity + counter merge)...")
	deduplicateNeurons(brainRoot)

	fmt.Println("[IDLE-WORKER] 🧬 Sleep-time consolidation (Hebbian → axon)...")
	sleepConsolidate(brainRoot)

	brain := scanBrain(brainRoot)
	result := runSubsumption(brain)
	growthLogDir := filepath.Join(brainRoot, "hippocampus", "session_log")
	os.MkdirAll(growthLogDir, 0750)
	growthLogFile := filepath.Join(growthLogDir, "growth.log")

	correctionsToday := 0
	historyPath := filepath.Join(brainRoot, "_inbox", "corrections_history.jsonl")
	if data, err := os.ReadFile(historyPath); err == nil {
		today := time.Now().Format("2006-01-02")
		for _, line := range strings.Split(string(data), "\n") {
			if strings.Contains(line, today) {
				correctionsToday++
			}
		}
	}

	entry := fmt.Sprintf("%s: neurons=%d, activation=%d, regions=%d, corrections=%d\n",
		time.Now().Format("2006-01-02_15:04"), result.TotalNeurons, result.TotalCounter, len(result.ActiveRegions), correctionsToday)
	f, _ := os.OpenFile(growthLogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if f != nil {
		f.WriteString(entry)
		f.Close()
	}
	fmt.Printf("[IDLE-WORKER-GROWTH] %s", entry)

	if correctionsToday > 20 {
		fmt.Printf("[IDLE-WORKER] ⚠️ 교정 빈도 높음 (%d건/일) — neuronize 우선 권장\n", correctionsToday)
	}

	fmt.Println("[IDLE-WORKER] 📸 Git snapshot...")
	gitSnapshot(brainRoot)

	nasTarget := `Z:\VOL1\VGVR\BRAIN\LW\system\neurons\brain_v4`
	if _, err := os.Stat(nasTarget); err == nil {
		fmt.Println("[IDLE-WORKER] 📡 NAS sync...")
		out, err := SafeCombinedOutput(ExecTimeoutSync, "robocopy", brainRoot, nasTarget, "/MIR", "/XD", ".git", "/XF", "*.dormant", "/NFL", "/NDL", "/NP", "/NJH", "/NJS")
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() >= 8 {
				fmt.Printf("[IDLE-WORKER] ❌ NAS sync error (exit %d): %s\n", exitErr.ExitCode(), string(out))
			} else {
				fmt.Printf("[IDLE-WORKER] ✅ NAS synced\n")
			}
		} else {
			fmt.Println("[IDLE-WORKER] ✅ NAS synced (no changes)")
		}
	} else {
		fmt.Println("[IDLE-WORKER] ⚠️ NAS Z: not available — skipping sync")
	}

	writeHeartbeat(brainRoot, result)
	refreshCodeMap(brainRoot)

	runtimeDir := filepath.Join(filepath.Dir(brainRoot), "runtime")
	if _, err := os.Stat(runtimeDir); err == nil {
		if vetOut, err := SafeCombinedOutputDir(ExecTimeoutGit, runtimeDir, "go", "vet", "./..."); err != nil {
			fmt.Printf("[IDLE-WORKER] ⚠️ go vet 이상 감지: %s\n", strings.TrimSpace(string(vetOut)))
			logEpisode(brainRoot, "VET_FAIL", string(vetOut))
		} else {
			fmt.Println("[IDLE-WORKER] ✅ go vet 통과")
		}
	}

	growth := result.TotalNeurons - 0
	if hbData, err := os.ReadFile(filepath.Join(brainRoot, "_heartbeat.json")); err == nil {
		var hb map[string]interface{}
		if json.Unmarshal(hbData, &hb) == nil {
			if delta, ok := hb["growth_delta"].(float64); ok {
				growth = int(delta)
			}
		}
	}

	summary := fmt.Sprintf("[NeuronFS IDLE] %d neurons, act:%d, Δ%d — 전사 패턴 분석 후 커뮤니티 검색해서 자가진화 실행. growth.log 확인하고 다음 action 자동 결정.",
		result.TotalNeurons, result.TotalCounter, growth)
	fmt.Printf("[SUMMARY] %s\n", summary)
}
