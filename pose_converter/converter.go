package pose_converter

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
)

// PoseData represents 3D joint coordinates.
type PoseData struct {
	Joint string  `json:"joint"`
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	Z     float64 `json:"z"`
}

// roundTo6 rounds a float64 to 6 decimal places.
func roundTo6(v float64) float64 {
	return math.Round(v*1e6) / 1e6
}

// ConvertPoseData converts 3D joint data to JSON with precision reduction (6 decimal places).
func ConvertPoseData(joint string, x, y, z float64) (string, error) {
	if strings.TrimSpace(joint) == "" {
		return "", fmt.Errorf("joint name cannot be empty")
	}

	pose := PoseData{
		Joint: joint,
		X:     roundTo6(x),
		Y:     roundTo6(y),
		Z:     roundTo6(z),
	}

	jsonData, err := json.Marshal(pose)
	if err != nil {
		return "", fmt.Errorf("failed to marshal pose data: %w", err)
	}
	return string(jsonData), nil
}

// BatchConvertPoseData converts multiple joint data in a single call.
func BatchConvertPoseData(inputs []struct {
	Joint string
	X, Y, Z float64
}) ([]string, error) {
	results := make([]string, 0, len(inputs))
	for _, input := range inputs {
		res, err := ConvertPoseData(input.Joint, input.X, input.Y, input.Z)
		if err != nil {
			return nil, err
		}
		results = append(results, res)
	}
	return results, nil
}
