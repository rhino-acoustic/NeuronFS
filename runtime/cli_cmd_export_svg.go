package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// ExportSvgCmd implements the Command interface for the new --export-svg action.
type ExportSvgCmd struct{}

func (c *ExportSvgCmd) Name() string {
	return "--export-svg"
}

func (c *ExportSvgCmd) Execute(brainRoot string, args []string) error {
	brain := scanBrain(brainRoot)
	res := runSubsumption(brain)
	
	svgContent := GenerateDashboardSVG(brain, res.TotalNeurons, res.TotalCounter)
	
	outPath := filepath.Join(filepath.Dir(brainRoot), "dashboard.svg")
	if err := os.WriteFile(outPath, []byte(svgContent), 0644); err != nil {
		fmt.Printf("[FATAL] Failed to export SVG: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("[NeuronFS] Successfully exported Dashboard V2 SVG to %s\n", outPath)
	return nil
}
