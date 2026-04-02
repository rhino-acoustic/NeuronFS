package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// HebbianNeuron represents a retrieved neuron with biological metrics
type HebbianNeuron struct {
	Path          string  `json:"path"`
	Activation    int     `json:"activation"`
	Weight        int     `json:"weight"`
	CognitiveTier string  `json:"cognitive_tier"`
	HebbianScore  float64 `json:"hebbian_score"`
	Content       string  `json:"content"`
}

// RetrieveResponse is the JSON response for the LLM router
type RetrieveResponse struct {
	Router       string          `json:"router"`
	ActiveTokens int             `json:"active_tokens_est"`
	Neurons      []HebbianNeuron `json:"neurons"`
}

func handleRetrieve(brainRoot string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		brain := scanBrain(brainRoot)
		result := runSubsumption(brain)

		var hebbianNeurons []HebbianNeuron
		requiresHighTier := false
		totalTokens := 0

		for _, region := range result.ActiveRegions {
			for _, n := range region.Neurons {
				if n.IsDormant {
					continue
				}

				// Find the trace file to extract metadata
				traceFiles, _ := filepath.Glob(filepath.Join(n.FullPath, "*.neuron"))
				weight := n.Counter + n.Dopamine - n.Contra
				cogTier := "LOW"
				content := ""

				if len(traceFiles) > 0 {
					traceContent, err := os.ReadFile(traceFiles[0])
					if err == nil {
						content = string(traceContent)
						// svParseFrontmatter parses weight. 
						// To parse cognitive_tier, we do a simple string search.
						lines := strings.Split(content, "\n")
						for _, l := range lines {
							if strings.HasPrefix(strings.TrimSpace(l), "cognitive_tier:") {
								tier := strings.ToUpper(strings.TrimSpace(strings.TrimPrefix(l, "cognitive_tier:")))
								tier = strings.ReplaceAll(tier, "\"", "") // remove quotes
								tier = strings.ReplaceAll(tier, "'", "")
								if tier == "HIGH" || tier == "LOW" {
									cogTier = tier
								}
								break
							}
						}
						// Estimate tokens as bytes / 4
						totalTokens += len(content) / 4
					}
				}

				if cogTier == "HIGH" {
					requiresHighTier = true
				}

				// Hebbian Score formula: (Activation * 1.5) + Weight.
				score := float64(n.Counter)*1.5 + float64(weight)

				hebbianNeurons = append(hebbianNeurons, HebbianNeuron{
					Path:          n.Path,
					Activation:    n.Counter,
					Weight:        weight,
					CognitiveTier: cogTier,
					HebbianScore:  score,
					Content:       content,
				})
			}
		}

		// Sort by HebbianScore descending
		sort.Slice(hebbianNeurons, func(i, j int) bool {
			return hebbianNeurons[i].HebbianScore > hebbianNeurons[j].HebbianScore
		})

		router := "Groq Llama 3 8B"
		if requiresHighTier {
			router = "Claude 3.5 Sonnet / VibeVoice-ASR (8.3B)"
			fmt.Printf("\033[35m[AURA] High cognitive load detected. Routing to %s.\033[0m\n", router)
		} else {
			fmt.Printf("\033[36m[PULSE] Low cognitive load. Routing to %s.\033[0m\n", router)
		}

		resp := RetrieveResponse{
			Router:       router,
			ActiveTokens: totalTokens,
			Neurons:      hebbianNeurons,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

