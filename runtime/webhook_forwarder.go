package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// ============================================================================
// Module: Webhook Forwarder Daemon (Phase 32)
// Sends push notifications to external B2B services (Slack/Discord) asynchronously.
// ============================================================================

type WebhookPayload struct {
	Event   string      `json:"event"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// WebhookQueue is a buffered channel to prevent blocking the main runtime
var WebhookQueue = make(chan WebhookPayload, 100)

type webhookConfig struct {
	Endpoints []string `json:"endpoints"`
}

func startWebhookDaemon(brainRoot string) {
	fmt.Println("[Webhook] Daemon initialized. Listening for outbound events...")
	
	client := &http.Client{Timeout: 5 * time.Second}

	for payload := range WebhookQueue {
		// Load endpoints every time to allow dynamic updates
		endpoints := getWebhookEndpoints(brainRoot)
		
		if len(endpoints) == 0 {
			continue // No webhooks registered, silently drop
		}

		jsonData, err := json.Marshal(payload)
		if err != nil {
			fmt.Printf("[Webhook] Failed to marshal payload: %v\n", err)
			continue
		}

		for _, endpoint := range endpoints {
			go pushWebhook(client, endpoint, jsonData)
		}
	}
}

func getWebhookEndpoints(brainRoot string) []string {
	configPath := filepath.Join(brainRoot, "_config", "webhooks.json")
	
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil
	}

	var conf webhookConfig
	if err := json.Unmarshal(data, &conf); err != nil {
		return nil
	}

	return conf.Endpoints
}

func pushWebhook(client *http.Client, endpoint string, jsonData []byte) {
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("[Webhook] Delivery failed to %s: %v\n", endpoint, err)
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode >= 400 {
		fmt.Printf("[Webhook] Endpoint %s returned status %d\n", endpoint, resp.StatusCode)
	}
}

// TriggerWebhook is a helper function to safely enqueue an event
func TriggerWebhook(event string, message string, data interface{}) {
	payload := WebhookPayload{
		Event:   event,
		Message: message,
		Data:    data,
	}
	select {
	case WebhookQueue <- payload:
		// success
	default:
		fmt.Println("[Webhook] Warning: Queue is full. Dropping event.")
	}
}
