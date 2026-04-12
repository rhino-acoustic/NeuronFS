package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ============================================================================
// Module: Skill Accumulator — Phase 58
// Closed Learning Loop: success patterns → skill files → reusable knowledge
// Inspired by Hermes Agent, adapted to NeuronFS filesystem paradigm.
// Key difference: Hermes uses FTS5 DB, we use folders (zero dependency).
// ============================================================================

// SkillEntry represents a learned success pattern
type SkillEntry struct {
	Name      string    `json:"name"`
	Category  string    `json:"category"`
	Pattern   string    `json:"pattern"`
	Source    string    `json:"source"`
	LearnedAt time.Time `json:"learned_at"`
	UsedCount int       `json:"used_count"`
}

// LearnSkill records a successful pattern as a reusable skill neuron
func LearnSkill(brainRoot, category, name, pattern, source string) error {
	skillDir := filepath.Join(brainRoot, "hippocampus", "skills", category)
	if err := os.MkdirAll(skillDir, 0750); err != nil {
		return fmt.Errorf("스킬 디렉토리 생성 실패: %w", err)
	}

	safeName := strings.ReplaceAll(name, " ", "_")
	safeName = strings.ReplaceAll(safeName, "/", "_")
	skillPath := filepath.Join(skillDir, safeName+".skill")

	entry := SkillEntry{
		Name:      name,
		Category:  category,
		Pattern:   pattern,
		Source:    source,
		LearnedAt: time.Now(),
		UsedCount: 0,
	}

	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return fmt.Errorf("스킬 직렬화 실패: %w", err)
	}

	// Strangler Fig: append if exists, don't overwrite
	if _, err := os.Stat(skillPath); err == nil {
		// Skill already exists — increment usage count
		existing, readErr := os.ReadFile(skillPath)
		if readErr == nil {
			var existingEntry SkillEntry
			if json.Unmarshal(existing, &existingEntry) == nil {
				existingEntry.UsedCount++
				data, _ = json.MarshalIndent(existingEntry, "", "  ")
			}
		}
	}

	if err := os.WriteFile(skillPath, data, 0640); err != nil {
		return fmt.Errorf("스킬 저장 실패: %w", err)
	}

	fmt.Printf("[SKILL] ✅ Learned: %s/%s (source: %s)\n", category, name, source)
	return nil
}

// RecallSkills retrieves all skills for a given category
func RecallSkills(brainRoot, category string) ([]SkillEntry, error) {
	skillDir := filepath.Join(brainRoot, "hippocampus", "skills", category)
	entries, err := os.ReadDir(skillDir)
	if err != nil {
		return nil, nil // No skills yet — not an error
	}

	var skills []SkillEntry
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".skill") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(skillDir, e.Name()))
		if err != nil {
			continue
		}
		var skill SkillEntry
		if json.Unmarshal(data, &skill) == nil {
			skills = append(skills, skill)
		}
	}
	return skills, nil
}

// RecallAllSkills retrieves all skills across all categories
func RecallAllSkills(brainRoot string) ([]SkillEntry, error) {
	skillsRoot := filepath.Join(brainRoot, "hippocampus", "skills")
	var allSkills []SkillEntry

	_ = filepath.WalkDir(skillsRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(d.Name(), ".skill") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		var skill SkillEntry
		if json.Unmarshal(data, &skill) == nil {
			allSkills = append(allSkills, skill)
		}
		return nil
	})

	return allSkills, nil
}

// InjectSkillsToPrompt generates a skill context block for agent prompts
func InjectSkillsToPrompt(brainRoot, category string) string {
	skills, err := RecallSkills(brainRoot, category)
	if err != nil || len(skills) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("[Learned Skills]\n")
	for _, s := range skills {
		sb.WriteString(fmt.Sprintf("- %s: %s (used %d times)\n", s.Name, s.Pattern, s.UsedCount))
	}
	return sb.String()
}
