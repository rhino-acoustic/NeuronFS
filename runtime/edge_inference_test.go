package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEdgeInference_Mock(t *testing.T) {
	// Create mock server simulating Ollama endpoint
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Fatalf("expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/api/generate" {
			t.Fatalf("expected /api/generate path, got %s", r.URL.Path)
		}

		var req OllamaRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		resp := OllamaResponse{
			Model:    req.Model,
			Response: "Mock edge answer for: " + req.Prompt,
			Done:     true,
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	// Temporarily override endpoint
	origEndpoint := OllamaLocalEndpoint
	OllamaLocalEndpoint = ts.URL + "/api/generate"
	defer func() { OllamaLocalEndpoint = origEndpoint }()

	// For the test, we mock CheckEdgeAvailability to pass. But InvokeLocalLLM directly calls it.
	// Since we changed the endpoint but CheckEdgeAvailability pings localhost:11434, the test might fail if Ollama isn't running.
	// Let's refactor CheckEdgeAvailability to check the actual host/port of OllamaLocalEndpoint.
}
