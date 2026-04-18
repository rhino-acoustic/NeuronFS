package pose_converter

import (
	"testing"
	"strings"
)

func TestConvertPoseData(t *testing.T) {
	expectedSub := `"joint":"shoulder_right"`
	actual, err := ConvertPoseData("shoulder_right", 10.5, 20.1, 5.0)

	if err != nil {
		t.Fatalf("ConvertPoseData failed: %v", err)
	}

	if !strings.Contains(actual, expectedSub) {
		t.Errorf("expected JSON to contain %q, but got %q", expectedSub, actual)
	}
}

func TestConvertPoseData_EmptyJoint(t *testing.T) {
	_, err := ConvertPoseData("", 10.5, 20.1, 5.0)
	if err == nil {
		t.Error("expected error for empty joint name, but got nil")
	}
}

func TestBatchConvertPoseData(t *testing.T) {
	inputs := []struct {
		Joint   string
		X, Y, Z float64
	}{
		{"joint1", 1.0, 2.0, 3.0},
		{"joint2", 4.0, 5.0, 6.0},
	}

	results, err := BatchConvertPoseData(inputs)
	if err != nil {
		t.Fatalf("BatchConvertPoseData failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 results, but got %d", len(results))
	}
}
