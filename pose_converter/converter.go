package pose_converter

import (
	"encoding/json"
	"fmt"
	"strings"
)

// PoseData represents 3D joint coordinates.
type PoseData struct {
	Joint string  `json:"joint"`
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	Z     float64 `json:"z"`
}

// ConvertPoseData converts 3D joint data to JSON.
func ConvertPoseData(joint string, x, y, z float64) (string, error) {
	if strings.TrimSpace(joint) == "" {
		return "", fmt.Errorf("joint name cannot be empty")
	}

	pose := PoseData{
		Joint: joint,
		X:     x,
		Y:     y,
		Z:     z,
	}

	jsonData, err := json.Marshal(pose)
	if err != nil {
		return "", fmt.Errorf("failed to marshal pose data: %w", err)
	}
	return string(jsonData), nil
}
