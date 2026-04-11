package main

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
)

// FolderMap holds domain-specific folder names
type FolderMap struct {
	Cortex      string `json:"cortex"`
	Hippocampus string `json:"hippocampus"`
	Brainstem   string `json:"brainstem"`
	Limbic      string `json:"limbic"`
	Inbox       string `json:"inbox"`
	Sensors     string `json:"sensors"`
}

// NeuronConfig holds environment-agnostic configuration values
type NeuronConfig struct {
	SystemName    string    `json:"system_name"`
	NASSyncTarget string    `json:"nas_sync_target"`
	Folders       FolderMap `json:"folders"`
	BrainRoot     string    `json:"-"` // runtime inject
}

// LoadConfig reads the configuration from the brain_v4 directory, falling back to defaults if parsing fails.
func LoadConfig(brainRoot string) *NeuronConfig {
	cfg := &NeuronConfig{
		SystemName:    "NeuronFS_Standalone",
		NASSyncTarget: "", // disabled by default
		Folders: FolderMap{
			Cortex:      "cortex",
			Hippocampus: "hippocampus",
			Brainstem:   "brainstem",
			Limbic:      "limbic",
			Inbox:       "_inbox",
			Sensors:     "sensors",
		},
		BrainRoot: brainRoot,
	}

	configPath := filepath.Join(brainRoot, "_config.json")
	if data, err := os.ReadFile(configPath); err == nil {
		json.Unmarshal(data, cfg)
	}
	return cfg
}

// Helpers for paths
func (c *NeuronConfig) HippocampusPath() string {
	return filepath.Join(c.BrainRoot, c.Folders.Hippocampus)
}

func (c *NeuronConfig) InboxPath() string {
	return filepath.Join(c.BrainRoot, c.Folders.Inbox)
}

func (c *NeuronConfig) GrowthLogPath() string {
	return filepath.Join(c.HippocampusPath(), "session_log", "growth.log")
}

func (c *NeuronConfig) CorrectionsHistoryPath() string {
	return filepath.Join(c.InboxPath(), "corrections_history.jsonl")
}

// SetupLogger initializes the standard slog environment tailored for NeuronFS operations.
func SetupLogger() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	handler := slog.NewTextHandler(os.Stdout, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)
}
