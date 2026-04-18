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
