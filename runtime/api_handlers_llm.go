package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// ─── API handlers for LLM ───

func handleNeuronizeAPI(brainRoot string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "POST only", 405)
			return
		}

		var req struct {
			DryRun bool `json:"dry_run"`
		}
		json.NewDecoder(r.Body).Decode(&req)

		apiKey := os.Getenv("GROQ_API_KEY")
		if apiKey == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(500)
			json.NewEncoder(w).Encode(map[string]string{"error": "GROQ_API_KEY not set"})
			return
		}

		// Collect corrections (same as runNeuronize)
		var corrections []string
		correctionsPath := filepath.Join(brainRoot, "_inbox", "corrections.jsonl")
		if data, err := os.ReadFile(correctionsPath); err == nil && len(data) > 0 {
			for _, line := range strings.Split(string(data), "\n") {
				line = strings.TrimSpace(line)
				if line != "" {
					corrections = append(corrections, line)
				}
			}
		}
		episodes := collectEpisodes(brainRoot)
		errorKeywords := []string{"ERROR", "FAIL", "TRAUMA", "ROLLBACK"}
		for _, ep := range episodes {
			for _, kw := range errorKeywords {
				if strings.Contains(strings.ToUpper(ep), kw) {
					corrections = append(corrections, ep)
					break
				}
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":      "neuronize_queued",
			"corrections": len(corrections),
			"dry_run":     req.DryRun,
			"message":     fmt.Sprintf("%d correction sources detected. Use CLI for full execution.", len(corrections)),
		})
	}
}

func handlePolarizeAPI(brainRoot string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "POST only", 405)
			return
		}

		brain := scanBrain(brainRoot)

		positivePatterns := regexp.MustCompile(`(?i)^(use_|always_|prefer_|enable_|ensure_|must_|keep_|apply_)`)
		englishName := regexp.MustCompile(`^[a-zA-Z_]+$`)

		type candidate struct {
			Region  string `json:"region"`
			Path    string `json:"path"`
			Name    string `json:"name"`
			Counter int    `json:"counter"`
			NewName string `json:"new_name"`
		}

		var candidates []candidate
		for _, region := range brain.Regions {
			if region.Name == "brainstem" || region.Name == "limbic" {
				continue
			}
			for _, n := range region.Neurons {
				if !englishName.MatchString(n.Name) {
					continue
				}
				if positivePatterns.MatchString(n.Name) {
					candidates = append(candidates, candidate{
						Region:  region.Name,
						Path:    n.Path,
						Name:    n.Name,
						Counter: n.Counter,
						NewName: ruleBasedPolarize(n.Name),
					})
				}
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"candidates": candidates,
			"total":      len(candidates),
			"message":    "Use CLI --polarize to execute shifts.",
		})
	}
}

func handleEvolveAPI(brainRoot string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "POST only", 405)
			return
		}

		var req struct {
			DryRun bool `json:"dry_run"`
		}
		json.NewDecoder(r.Body).Decode(&req)

		apiKey := os.Getenv("GROQ_API_KEY")
		if apiKey == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(500)
			json.NewEncoder(w).Encode(map[string]string{"error": "GROQ_API_KEY not set"})
			return
		}

		// Collect data
		episodes := collectEpisodes(brainRoot)
		brain := scanBrain(brainRoot)
		result := runSubsumption(brain)
		brainSummary := buildBrainSummary(brain, result)
		prompt := buildEvolvePrompt(episodes, brainSummary, result)

		// Call Groq
		evoResp, err := callGroq(apiKey, prompt)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(500)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		// Execute if not dry run
		executed := 0
		skipped := 0
		if !req.DryRun {
			for _, action := range evoResp.Actions {
				switch action.Type {
				case "grow":
					if err := growNeuron(brainRoot, action.Path); err != nil {
						skipped++
					} else {
						executed++
						go sendTelegramEvolve(brainRoot, action)
					}
				case "fire":
					fireNeuron(brainRoot, action.Path)
					executed++
				case "signal":
					if action.Signal == "" {
						action.Signal = "dopamine"
					}
					if err := signalNeuron(brainRoot, action.Path, action.Signal); err != nil {
						skipped++
					} else {
						executed++
					}
				case "prune", "decay":
					fullPath := filepath.Join(brainRoot, strings.ReplaceAll(action.Path, "/", string(filepath.Separator)))
					if _, err := os.Stat(fullPath); err == nil {
						dormantFile := filepath.Join(fullPath, "evolve.dormant")
						os.WriteFile(dormantFile, []byte(fmt.Sprintf("Evolved: %s\nReason: %s\n",
							time.Now().Format("2006-01-02"), action.Reason)), 0600)
						executed++
					} else {
						skipped++
					}
				default:
					skipped++
				}
			}

			if executed > 0 {
				logEpisode(brainRoot, "EVOLVE:API", fmt.Sprintf("%d executed, %d skipped", executed, skipped))
				autoReinject(brainRoot)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"summary":  evoResp.Summary,
			"insights": evoResp.Insights,
			"actions":  evoResp.Actions,
			"executed": executed,
			"skipped":  skipped,
			"dry_run":  req.DryRun,
		})
	}
}
