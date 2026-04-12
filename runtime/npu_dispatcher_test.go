package main

import (
	"strings"
	"testing"
)

func TestNPU_HardwareProfile(t *testing.T) {
	profile := detectHardwareProfile()
	if profile == "" {
		t.Error("Hardware profile should not be empty")
	}
	t.Logf("Detected HAL Profile: %s", profile)
}

func TestNPU_ZeroCopyPool(t *testing.T) {
	ptr := AcquireZeroCopyBuffer()
	if ptr == nil {
		t.Fatal("Acquire returned nil")
	}
	b := *ptr
	if len(b) != 16384 {
		t.Errorf("Expected 16KB buffer, got %d", len(b))
	}
	
	ReleaseZeroCopyBuffer(ptr)
}

func TestNPU_BuildInferencePayload(t *testing.T) {
	model := "test-model"
	prompt := "hello\nworld\"\\"
	
	buf, release := BuildInferencePayload(model, prompt)
	defer release()
	
	str := string(buf)
	
	// Must contain the properly escaped JSON elements
	if !strings.Contains(str, `{"model":"test-model"`) {
		t.Errorf("Missing model in payload: %s", str)
	}
	
	if !strings.Contains(str, `"prompt":"hello\nworld\"\\"`) {
		t.Errorf("Missing or incorrectly escaped prompt: %s", str)
	}
}
