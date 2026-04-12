package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// ============================================================================
// Module: Two-Track Inference & Speculative Fallback (Phase 29)
// Concept: Launch Local LLM (Draft) and Cloud LLM (Target) concurrently.
// If Cloud times out or fails (Air-gapped), the Local Edge takes over.
// ============================================================================

type twoTrackResult struct {
	Track    string // "edge" or "cloud"
	Response string
	Err      error
}

// invokeCloudLLM is a lightweight mock/wrapper for a remote cloud inference API.
func invokeCloudLLM(ctx context.Context, prompt string) (string, error) {
	apiKey := os.Getenv("NEURONFS_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("NEURONFS_API_KEY missing - cloud track disabled")
	}

	// Example Cloud POST payload (OpenAI-ish)
	reqPayload := map[string]interface{}{
		"model": "frontier-target-model-2026",
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}
	jsonData, _ := json.Marshal(reqPayload)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.neuronfs.cloud/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	// Hard limit of 8s for cloud; beyond this, edge wins.
	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("cloud network offline or timeout: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("cloud API error %d", resp.StatusCode)
	}

	bodyBytes, _ := io.ReadAll(resp.Body)
	// Simplified parsing
	var cloudResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(bodyBytes, &cloudResp); err == nil && len(cloudResp.Choices) > 0 {
		return cloudResp.Choices[0].Message.Content, nil
	}

	return string(bodyBytes), nil
}

// InvokeTwoTrack starts a race between Edge (Local) and Cloud.
// If Cloud is successful before 5 seconds, it returns Cloud.
// If Cloud fails or network is disconnected, and Edge finishes, Edge wins. 
// "Speculative Survival Mode" guarantees uptime.
func InvokeTwoTrack(model, prompt string) (finalResponse string, winningTrack string, err error) {
	resultChan := make(chan twoTrackResult, 2)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Track 1: Edge (Draft Model)
	go func() {
		resp, err := InvokeLocalLLM(model, prompt)
		resultChan <- twoTrackResult{Track: "edge", Response: resp, Err: err}
	}()

	// Track 2: Cloud (Target Model)
	go func() {
		resp, err := invokeCloudLLM(ctx, prompt)
		resultChan <- twoTrackResult{Track: "cloud", Response: resp, Err: err}
	}()

	edgeRes := twoTrackResult{Track: "edge", Err: fmt.Errorf("pending")}
	cloudRes := twoTrackResult{Track: "cloud", Err: fmt.Errorf("pending")}

	// Wait loop
	timeout := time.After(30 * time.Second) // Global hard fail
	completed := 0

	for completed < 2 {
		select {
		case res := <-resultChan:
			completed++
			if res.Track == "cloud" {
				cloudRes = res
				if res.Err == nil {
					// Cloud succeeds! It overrides Edge unconditionally if fast enough.
					return res.Response, "cloud", nil
				}
				// If cloud fails, check if Edge is already done and succeeded
				if edgeRes.Err == nil && edgeRes.Response != "" {
					return edgeRes.Response, "edge (survival fallback)", nil
				}
			} else if res.Track == "edge" {
				edgeRes = res
				// If Cloud already failed or missing API Key, Edge wins instantly
				if cloudRes.Err != nil && cloudRes.Err.Error() != "pending" {
					if res.Err == nil {
						return res.Response, "edge", nil
					}
				}
			}
		case <-timeout:
			return "", "none", fmt.Errorf("both tracks timed out globally")
		}
	}

	// Both finished. If we reach here, Cloud failed, try returning Edge.
	if edgeRes.Err == nil {
		return edgeRes.Response, "edge", nil
	}

	return "", "none", fmt.Errorf("total system failure (cloud: %v, edge: %v)", cloudRes.Err, edgeRes.Err)
}
