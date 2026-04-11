package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func runIdleCoreWorker(brainRoot string) {
	SetupLogger()
	// Community Best Practice: Panic Recovery in Stateless Workers
	defer func() {
		if r := recover(); r != nil {
			slog.Error("워커 패닉 발생 (Recovered)", "error", r)
			fmt.Printf("[SUMMARY] [NeuronFS IDLE] ⚠️ 시스템 에러 감지. 유휴 워커 패닉 발생: %v\n", r)
		}
	}()

	slog.Info("stateless idle core execution started", "component", "idle_worker")

	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey != "" {
		pendingTurns := digestTranscripts(brainRoot)
		if pendingTurns >= 10 {
			slog.Info("전사 청크 누적 확인", "pending_turns", pendingTurns, "action", "Auto-Neuronize 실행")
			runNeuronize(brainRoot, false)
		}
	}

	cfg := LoadConfig(brainRoot)

	failedEvolutions := detectFailedEvolutions(brainRoot)
	if len(failedEvolutions) > 0 {
		growthLogFile := cfg.GrowthLogPath()
		f, _ := os.OpenFile(growthLogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if f != nil {
			for _, fe := range failedEvolutions {
				f.WriteString(fmt.Sprintf("%s: META_EVOLVE_FAIL path=%s (30d+ inactive)\n",
					time.Now().Format("2006-01-02_15:04"), fe))
			}
			f.Close()
		}
		slog.Warn("메타진화 실패 항목 감지됨", "failed_count", len(failedEvolutions))
	}

	if apiKey != "" {
		slog.Info("Evolve 실행", "model", "region 분류 AI 판단 모델 탑재 완료")
		runEvolve(brainRoot, false)
	}

	slog.Info("Running auto-decay", "days", 7)
	runDecay(brainRoot, 7)

	slog.Info("Running prune", "target", "推 low-value cleanup")
	pruneWeakNeurons(brainRoot)

	slog.Info("Spaced repetition", "desc", "reinforce high-value neurons")
	spacedRepetitionFire(brainRoot)

	slog.Info("Running consolidate", "desc", "hybrid similarity + counter merge")
	deduplicateNeurons(brainRoot)

	slog.Info("Sleep-time consolidation", "desc", "Hebbian → axon")
	sleepConsolidate(brainRoot)

	brain := scanBrain(brainRoot)
	result := runSubsumption(brain)
	growthLogDir := filepath.Join(cfg.HippocampusPath(), "session_log")
	os.MkdirAll(growthLogDir, 0750)
	growthLogFile := filepath.Join(growthLogDir, "growth.log")

	correctionsToday := 0
	historyPath := cfg.CorrectionsHistoryPath()
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
		slog.Warn("교정 빈도 높음", "corrections_today", correctionsToday, "recommend", "neuronize 우선 권장")
	}

	slog.Info("Git snapshot 시작")
	gitSnapshot(brainRoot)

	if cfg.NASSyncTarget != "" {
		if _, err := os.Stat(cfg.NASSyncTarget); err == nil {
			slog.Info("NAS sync 시작", "target", cfg.NASSyncTarget)
			out, err := SafeCombinedOutput(ExecTimeoutSync, "robocopy", brainRoot, cfg.NASSyncTarget, "/MIR", "/XD", ".git", "/XF", "*.dormant", "/NFL", "/NDL", "/NP", "/NJH", "/NJS")
			if err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() >= 8 {
					slog.Error("NAS sync error", "exit_code", exitErr.ExitCode(), "output", string(out))
				} else {
					slog.Info("NAS synced successfully")
				}
			} else {
				slog.Info("NAS synced", "status", "no changes")
			}
		} else {
			slog.Warn("NAS Z: not available", "status", "skipping sync")
		}
	} else {
		slog.Info("NAS Sync disabled in config", "status", "skipping sync")
	}

	writeHeartbeat(brainRoot, result)
	refreshCodeMap(brainRoot)

	runtimeDir := filepath.Join(filepath.Dir(brainRoot), "runtime")
	if _, err := os.Stat(runtimeDir); err == nil {
		if vetOut, err := SafeCombinedOutputDir(ExecTimeoutGit, runtimeDir, "go", "vet", "./..."); err != nil {
			slog.Warn("go vet 이상 감지", "output", strings.TrimSpace(string(vetOut)))
			logEpisode(brainRoot, "VET_FAIL", string(vetOut))
		} else {
			slog.Info("go vet 통과")
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
