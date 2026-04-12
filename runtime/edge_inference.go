package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"
)

// ============================================================================
// Module: Edge Compute (On-Device LLM Pipeline)
// Fallback local execution when cloud is unavailable or latency is critical.
// ============================================================================

var OllamaLocalEndpoint = "http://localhost:11434/api/generate"

type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type OllamaResponse struct {
	Model     string `json:"model"`
	Response  string `json:"response"`
	Done      bool   `json:"done"`
}

// CheckEdgeAvailability pings the local 11434 port to see if Ollama is running
func CheckEdgeAvailability() bool {
	parsed, err := url.Parse(OllamaLocalEndpoint)
	if err != nil {
		return false
	}
	timeout := 1 * time.Second
	conn, err := net.DialTimeout("tcp", parsed.Host, timeout)
	if err != nil {
		return false
	}
	if conn != nil {
		conn.Close()
		return true
	}
	return false
}

// InvokeLocalLLM sends a prompt to the local LLM instance
func InvokeLocalLLM(model, prompt string) (string, error) {
	if !CheckEdgeAvailability() {
		err := fmt.Errorf("edge compute (Ollama) is not running on localhost:11434")
		EngraveRuntimeError(findBrainRoot(), "Edge_Inference", err.Error())
		return "", err
	}

	// Phase 33: NPU HAL Fast-Path Payload Generation
	// Bypass json.Marshal reflection overhead using Zero-Copy Pool
	payloadBuf, releaseFn := BuildInferencePayload(model, prompt)
	defer releaseFn()

	resp, err := http.Post(OllamaLocalEndpoint, "application/json", bytes.NewBuffer(payloadBuf))
	if err != nil {
		errObj := fmt.Errorf("edge inference failed: %w", err)
		EngraveRuntimeError(findBrainRoot(), "Edge_Inference", errObj.Error())
		return "", errObj
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("edge reading response failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		errObj := fmt.Errorf("edge HTTP error: status %d", resp.StatusCode)
		EngraveRuntimeError(findBrainRoot(), "Edge_Inference", errObj.Error())
		return "", fmt.Errorf("%v, body: %s", errObj, string(bodyBytes))
	}

	var ollamaResp OllamaResponse
	if err := json.Unmarshal(bodyBytes, &ollamaResp); err != nil {
		return "", fmt.Errorf("edge response parsing failed: %w", err)
	}

	return ollamaResp.Response, nil
}
