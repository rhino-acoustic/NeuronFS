package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGraphQL_QueryStatus(t *testing.T) {
	brainRoot := t.TempDir()

	query := `{"query": "{ status }"}`
	req, err := http.NewRequest("POST", "/graphql", bytes.NewBufferString(query))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := HandleGraphQL(brainRoot)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected data field")
	}

	statusStr, ok := data["status"].(string)
	if !ok || statusStr != "STABLE / GRAPHQL-ENABLED" {
		t.Errorf("Unexpected status: %v", statusStr)
	}
}

func TestWebhook_TriggerNoPanic(t *testing.T) {
	// Drain channel quickly
	go func() {
		for range WebhookQueue {
		}
	}()

	TriggerWebhook("TEST_EVENT", "Test Message", nil)
	// If it doesn't block or panic, it succeeds
}
