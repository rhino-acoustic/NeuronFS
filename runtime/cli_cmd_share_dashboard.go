package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ShareDashboardCmd implements the Command interface for the core --share-dashboard action.
type ShareDashboardCmd struct{}

func (c *ShareDashboardCmd) Name() string {
	return "--share-dashboard"
}

func (c *ShareDashboardCmd) Execute(brainRoot string, args []string) error {
	fmt.Println("🚀 [ShareDashboard] Generating Visualization SVG...")
	
	// 1. Generate SVG
	brain := scanBrain(brainRoot)
	res := runSubsumption(brain)
	svgContent := GenerateDashboardSVG(brain, res.TotalNeurons, res.TotalCounter)
	
	// 2. Save dashboard_snapshot.svg to the brain directory or runtime
	nfsRoot := filepath.Dir(brainRoot)
	snapshotPath := filepath.Join(nfsRoot, "dashboard_snapshot.svg")
	
	err := os.WriteFile(snapshotPath, []byte(svgContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to save svg: %w", err)
	}
	fmt.Printf("💾 Stored locally at: %s\n", snapshotPath)

	// 3. Git Add and Commit
	fmt.Println("📦 Committing to Git...")
	_, err = SafeOutputDir(30*time.Second, nfsRoot, "git", "add", "dashboard_snapshot.svg")
	if err != nil {
		fmt.Printf("⚠️ Git add failed (make sure it's a git repo): %v\n", err)
	} else {
		_, err = SafeOutputDir(30*time.Second, nfsRoot, "git", "commit", "-m", "[EVOLVE:telemetry] Auto-capture Dashboard SVG Snapshot")
		if err != nil {
			fmt.Printf("⚠️ Git commit failed or nothing to commit: %v\n", err)
		}
	}

	// 4. Send to Telegram
	fmt.Println("📡 Broadcasting to Telegram...")
	hlLoadTelegram(nfsRoot)
	
	caption := fmt.Sprintf("⚡ NeuronFS Telemetry Snapshot\nTotal Neurons: %d\nTotal Activation: %d\n\nLive from Autopilot V2", res.TotalNeurons, res.TotalCounter)
	err = hlTgSendDocument(snapshotPath, caption)
	
	if err != nil {
		return fmt.Errorf("failed to share to telegram: %w", err)
	}
	
	fmt.Println("✅ Broadcast successful")
	return nil
}
