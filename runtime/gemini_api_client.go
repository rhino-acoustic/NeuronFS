package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// ============================================================================
// Module: Gemini API Direct Client — Phase 55
// Bypasses CLI overhead (Node.js boot + MCP + auth = 23s) by calling
// Gemini REST API directly. Expected latency: 3-5s per request.
// ============================================================================

// GeminiRequest holds the API request structure
type GeminiRequest struct {
	Contents []GeminiContent `json:"contents"`
}

// GeminiContent holds a single message
type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
}

// GeminiPart holds message content
type GeminiPart struct {
	Text string `json:"text"`
}

// GeminiResponse holds the API response
type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []GeminiPart `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

// GeminiClient provides direct API access without CLI overhead
type GeminiClient struct {
	APIKey  string
	Model   string
	BaseURL string
}

// NewGeminiClient creates a client, reading API key from environment or file
func NewGeminiClient() (*GeminiClient, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		// Try reading from brainstem config
		keyPath := os.Getenv("NEURONFS_BRAIN")
		if keyPath != "" {
			data, err := os.ReadFile(keyPath + "/brainstem/api_keys.neuron")
			if err == nil {
				apiKey = string(bytes.TrimSpace(data))
			}
		}
	}
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY 환경변수 또는 brainstem/api_keys.neuron 필요")
	}

	return &GeminiClient{
		APIKey:  apiKey,
		Model:   "gemini-2.0-flash",
		BaseURL: "https://generativelanguage.googleapis.com/v1beta/models",
	}, nil
}

// Query sends a prompt and returns the response text
func (gc *GeminiClient) Query(prompt string) (string, time.Duration, error) {
	start := time.Now()

	reqBody := GeminiRequest{
		Contents: []GeminiContent{
			{Parts: []GeminiPart{{Text: prompt}}},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", 0, fmt.Errorf("JSON 마샬링 실패: %w", err)
	}

	url := fmt.Sprintf("%s/%s:generateContent?key=%s", gc.BaseURL, gc.Model, gc.APIKey)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", time.Since(start), fmt.Errorf("API 호출 실패: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", time.Since(start), fmt.Errorf("응답 읽기 실패: %w", err)
	}

	if resp.StatusCode != 200 {
		return "", time.Since(start), fmt.Errorf("API 에러 %d: %s", resp.StatusCode, string(body))
	}

	var geminiResp GeminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return "", time.Since(start), fmt.Errorf("JSON 파싱 실패: %w", err)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", time.Since(start), fmt.Errorf("빈 응답")
	}

	text := geminiResp.Candidates[0].Content.Parts[0].Text
	return text, time.Since(start), nil
}

// QueryWithEmotion sends a prompt wrapped with EmotionPrompt
func (gc *GeminiClient) QueryWithEmotion(prompt string, tier EmotionLevel) (string, time.Duration, error) {
	wrappedPrompt := WrapWithEmotion(prompt, tier)
	return gc.Query(wrappedPrompt)
}
