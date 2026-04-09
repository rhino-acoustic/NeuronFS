package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

// ─── Global API Usage Counters (atomic, goroutine-safe) ───

var groqCallCount int64       // 총 호출 수
var groqTokensIn int64        // 입력 토큰
var groqTokensOut int64       // 출력 토큰
var groqErrorCount int64      // 에러 수
var groqLastCall atomic.Value // string: 마지막 호출 시각

// ─── Groq API Types ───

type groqMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type groqRequest struct {
	Model          string              `json:"model"`
	Messages       []groqMessage       `json:"messages"`
	Temperature    float64             `json:"temperature"`
	MaxTokens      int                 `json:"max_tokens"`
	TopP           float64             `json:"top_p"`
	Stream         bool                `json:"stream"`
	ResponseFormat *groqResponseFormat `json:"response_format,omitempty"`
}

type groqResponseFormat struct {
	Type string `json:"type"`
}

type groqChoice struct {
	Message groqMessage `json:"message"`
}

type groqResponse struct {
	Choices []groqChoice `json:"choices"`
	Error   *groqError   `json:"error,omitempty"`
	Usage   *groqUsage   `json:"usage,omitempty"`
}

type groqError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

type groqUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ─── Evolution Action Types ───

type evoAction struct {
	Type   string `json:"type"`   // grow | fire | signal | decay | merge | prune
	Path   string `json:"path"`   // neuron path
	Signal string `json:"signal"` // for signal type: dopamine | bomb
	Reason string `json:"reason"` // why this action
}

type evoResult struct {
	Summary  string      `json:"summary"`
	Actions  []evoAction `json:"actions"`
	Insights []string    `json:"insights"`
}

// ─── callGroqRaw: returns raw response content string ───

func callGroqRaw(apiKey string, req groqRequest) (string, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", "https://api.groq.com/openai/v1/chat/completions",
		strings.NewReader(string(jsonData)))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		atomic.AddInt64(&groqErrorCount, 1)
		return "", fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	body := new(strings.Builder)
	if _, err := fmt.Fprintf(body, ""); err != nil {
		return "", err
	}
	buf := make([]byte, 32768)
	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			body.Write(buf[:n])
		}
		if readErr != nil {
			break
		}
	}

	// Track call
	atomic.AddInt64(&groqCallCount, 1)
	groqLastCall.Store(time.Now().Format("2006-01-02T15:04:05"))

	if resp.StatusCode != 200 {
		atomic.AddInt64(&groqErrorCount, 1)
		return "", fmt.Errorf("API error %d: %s", resp.StatusCode, body.String())
	}

	var groqResp groqResponse
	if err := json.Unmarshal([]byte(body.String()), &groqResp); err != nil {
		atomic.AddInt64(&groqErrorCount, 1)
		return "", fmt.Errorf("unmarshal response: %w", err)
	}

	// Track token usage
	if groqResp.Usage != nil {
		atomic.AddInt64(&groqTokensIn, int64(groqResp.Usage.PromptTokens))
		atomic.AddInt64(&groqTokensOut, int64(groqResp.Usage.CompletionTokens))
	}

	if groqResp.Error != nil {
		atomic.AddInt64(&groqErrorCount, 1)
		return "", fmt.Errorf("groq error: %s", groqResp.Error.Message)
	}

	if len(groqResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return strings.TrimSpace(groqResp.Choices[0].Message.Content), nil
}

// ─── Call Groq API (Evolve) ───

func callGroq(apiKey string, prompt string) (*evoResult, error) {
	reqBody := groqRequest{
		Model: "llama-3.3-70b-versatile",
		Messages: []groqMessage{
			{Role: "system", Content: "당신은 NeuronFS 뇌의 백혈구(자가면역 세포)입니다. 사용자의 교정 로그와 에러 내역을 분석하여, 미래의 AI 에이전트들이 같은 실수를 절대 반복하지 못하도록 강력한 억제(Contra) 규칙을 만드십시오.\n\n[Rule Writing Guidelines]\n1. 파일명 (Filename): 부정/금지형 명사로 10자 이내 작성 (예: 반복루프_금지.md, 절대경로_의존X.md)\n2. 종결어미: ~해야 합니다, ~하는 것이 좋습니다 금지. ~~마라, ~~할 것, ~~금지 등 군더더기 없는 명령조(Imperative) 사용.\n3. 서문 금지: 알겠습니다, 다음은 규칙입니다 같은 응답 생성 절대 금지. 오직 Markdown 본문만 출력할 것.\n\n또한 기존 긍정형 뉴런을 부정형으로 전환할 경우, 내부 본문의 첫 문장에 금지의 이유(Rationale)를 단 한 줄의 강력한 메타포로 서술하십시오."},
			{Role: "user", Content: prompt},
		},
		Temperature:    EvolveTemp,
		MaxTokens:      EvolveTokens,
		TopP:           EvolveTopP,
		Stream:         false,
		ResponseFormat: &groqResponseFormat{Type: "json_object"},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var groqResp groqResponse
	if err := json.Unmarshal(body, &groqResp); err != nil {
		return nil, fmt.Errorf("unmarshal groq response: %w", err)
	}

	if groqResp.Error != nil {
		return nil, fmt.Errorf("groq error: %s (%s)", groqResp.Error.Message, groqResp.Error.Type)
	}

	if len(groqResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	content := groqResp.Choices[0].Message.Content
	content = strings.TrimSpace(content)

	var evoResp evoResult
	if err := json.Unmarshal([]byte(content), &evoResp); err != nil {
		if idx := strings.Index(content, "{"); idx >= 0 {
			content = content[idx:]
			if lastIdx := strings.LastIndex(content, "}"); lastIdx >= 0 {
				content = content[:lastIdx+1]
			}
			if err2 := json.Unmarshal([]byte(content), &evoResp); err2 != nil {
				return nil, fmt.Errorf("parse evolution result: %w\nRaw: %s", err2, truncate(content, 500))
			}
		} else {
			return nil, fmt.Errorf("parse evolution result: %w\nRaw: %s", err, truncate(content, 500))
		}
	}

	var validActions []evoAction
	for _, a := range evoResp.Actions {
		a.Path = strings.ReplaceAll(a.Path, "\\\\", "/")
		a.Path = strings.TrimPrefix(a.Path, "brain/")
		a.Path = strings.TrimPrefix(a.Path, "brain_v4/")

		parts := strings.SplitN(a.Path, "/", 2)
		if len(parts) < 2 {
			continue
		}
		region := parts[0]
		if _, ok := regionPriority[region]; !ok {
			continue
		}

		if region == "brainstem" && (a.Type == "grow" || a.Type == "prune" || a.Type == "decay") {
			continue
		}

		if region == "limbic" && (a.Type == "grow" || a.Type == "prune" || a.Type == "decay") {
			continue
		}

		if region == "sensors" && strings.HasPrefix(parts[1], "brand") {
			continue
		}

		switch a.Type {
		case "grow", "fire", "signal", "prune", "decay":
			validActions = append(validActions, a)
		}
	}
	evoResp.Actions = validActions

	return &evoResp, nil
}

// ─── GetGroqUsage ───
func GetGroqUsage() map[string]interface{} {
	lastCall := ""
	if v := groqLastCall.Load(); v != nil {
		lastCall = v.(string)
	}
	return map[string]interface{}{
		"calls":      atomic.LoadInt64(&groqCallCount),
		"tokens_in":  atomic.LoadInt64(&groqTokensIn),
		"tokens_out": atomic.LoadInt64(&groqTokensOut),
		"errors":     atomic.LoadInt64(&groqErrorCount),
		"last_call":  lastCall,
		"model":      "llama-3.3-70b-versatile",
	}
}
