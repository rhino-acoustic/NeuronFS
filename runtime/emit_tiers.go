// emit_tiers.go — Tier 오케스트레이션 (파일 쓰기)
//
// PROVIDES: writeAllTiers, writeAllTiersForTargets, applyOOMProtection, doInjectToFile
// DEPENDS:  brain.go (scanBrain, runSubsumption)
//           emit_bootstrap.go (emitBootstrap)
//           emit_helpers.go (emitIndex, emitRegionRules)
//           main.go (injectToGemini)
//           diag.go (generateBrainJSON)
//
// CALL GRAPH:
//   writeAllTiers(brainRoot)
//     ├→ scanBrain → runSubsumption → applyOOMProtection
//     ├→ emitBootstrap → injectToGemini          (Tier 1)
//     ├→ emitIndex → _index.md 작성               (Tier 2)
//     ├→ emitRegionRules × 7 → _rules.md 작성     (Tier 3)
//     └→ generateBrainJSON
//
//   writeAllTiersForTargets(brainRoot, target)
//     ├→ scanBrain → runSubsumption → applyOOMProtection
//     ├→ emitBootstrap → IDE 파일 직접 작성        (Tier 1)
//     ├→ emitIndex → _index.md 작성               (Tier 2)
//     ├→ emitRegionRules × 7 → _rules.md 작성     (Tier 3)
//     └→ generateBrainJSON

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// writeAllTiers: 3-tier 일괄 작성 (inject 모드)
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

func writeAllTiers(brainRoot string) {
	// [Phase 16 - LLM Agnostic]
	// Instead of hardcoding injectToGemini & AGENTS.md,
	// we now delegate to the unified dynamic dispatcher targeting auto-detected IDEs.
	writeAllTiersForTargets(brainRoot, "auto")
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// applyOOMProtection: 뉴런 과다 시 low-weight 드롭
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

func applyOOMProtection(brainRoot string, result *SubsumptionResult) int {
	type nInfo struct {
		rIdx   int
		nIdx   int
		weight float64 // effective weight (prefix × recency × counter)
		size   int
	}
	var flat []*nInfo

	now := time.Now()
	totalBytes := 0
	for i := range result.ActiveRegions {
		region := &result.ActiveRegions[i]
		for j := range region.Neurons {
			n := &region.Neurons[j]
			if n.IsDormant {
				continue
			}
			size := 0
			files, _ := filepath.Glob(filepath.Join(n.FullPath, "*.neuron"))
			for _, f := range files {
				if info, err := os.Stat(f); err == nil {
					size += int(info.Size())
				}
			}
			if size == 0 {
				size = 50
			}
			totalBytes += size

			// === Effective Weight ===
			baseWeight := float64(n.Counter + n.Dopamine - n.Contra)
			if baseWeight < 1 {
				baseWeight = 1
			}

			// 1) 접두어 가중치
			leafName := filepath.Base(n.FullPath)
			runes := []rune(leafName)
			prefixWeight := 1.0
			if len(runes) > 0 {
				switch runes[0] {
				case '必', '禁':
					prefixWeight = 2.0 // 필수/금지 = 최고 우선
				case '核':
					prefixWeight = 1.5 // 핵심 = 높음
				case '推':
					prefixWeight = 0.5 // 추천 = 낮음
				case '絶':
					prefixWeight = 2.0 // 절대 = 최고
				}
			}

			// 2) Recency boost (새 뉴런 보호기간 — 폴더 생성 시간 기준)
			age := now.Sub(n.BirthTime)
			recencyBoost := 1.0
			if age < 48*time.Hour {
				recencyBoost = 3.0 // 48시간 내 생성 = 3배 보호
			} else if age < 7*24*time.Hour {
				recencyBoost = 1.5 // 7일 내 생성 = 1.5배
			}

			effectiveWeight := baseWeight * prefixWeight * recencyBoost
			flat = append(flat, &nInfo{rIdx: i, nIdx: j, weight: effectiveWeight, size: size})
		}
	}

	if totalBytes <= 50000 {
		return 0
	}

	sort.Slice(flat, func(i, j int) bool {
		return flat[i].weight < flat[j].weight // 낮은 weight 먼저 탈락
	})

	dropped := 0
	for _, info := range flat {
		if totalBytes <= 50000 {
			break
		}
		result.ActiveRegions[info.rIdx].Neurons[info.nIdx].IsDormant = true
		totalBytes -= info.size
		dropped++
	}
	return dropped
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// EMIT TARGETS — Multi-editor support
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// EmitTarget defines a target editor configuration file
type EmitTarget struct {
	Name        string   `json:"name"`
	FileName    string   `json:"fileName"`
	SubDir      string   `json:"subDir"`
	IsGlobal    bool     `json:"isGlobal"`
	DetectPaths []string `json:"detectPaths"`
}

// loadLLMProviders dynamically builds emit targets from cortex/llm/*.json
func loadLLMProviders(brainRoot string) map[string]EmitTarget {
	targets := make(map[string]EmitTarget)
	llmDir := filepath.Join(brainRoot, "cortex", "llm")
	
	files, err := filepath.Glob(filepath.Join(llmDir, "*.json"))
	if err != nil {
		return targets
	}
	
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		var target EmitTarget
		if err := json.Unmarshal(data, &target); err == nil {
			base := filepath.Base(f)
			key := strings.TrimSuffix(base, ".json")
			targets[key] = target
		}
	}
	
	// Fallback if no providers found
	if len(targets) == 0 {
		targets["gemini"] = EmitTarget{
			Name:        "Gemini",
			FileName:    "GEMINI.md",
			SubDir:      ".gemini",
			IsGlobal:    true,
			DetectPaths: []string{".gemini"},
		}
	}
	return targets
}

// backupExistingRule backs up an existing rule file before overwriting.
// Returns the backup path if backed up, empty string if file didn't exist.
func backupExistingRule(targetPath string, brainRoot string) string {
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		return "" // No existing file to back up
	}

	// Create backup directory: <brainRoot>/.neuronfs_backup/
	backupDir := filepath.Join(brainRoot, ".neuronfs_backup")
	os.MkdirAll(backupDir, 0750)

	// Generate backup filename with timestamp
	baseName := filepath.Base(targetPath)
	ts := time.Now().Format("20060102_150405")
	backupName := fmt.Sprintf("%s.%s.bak", baseName, ts)
	backupPath := filepath.Join(backupDir, backupName)

	// Copy existing file to backup
	data, err := os.ReadFile(targetPath)
	if err != nil {
		return ""
	}
	if err := os.WriteFile(backupPath, data, 0600); err != nil {
		return ""
	}
	return backupPath
}

// homeDir returns the user's home directory, or empty string on error.
func homeDir() string {
	h, _ := os.UserHomeDir()
	return h
}

// writeAllTiersForTargets writes brain rules to specific editor target(s)
// target can be a single key (e.g. "cursor") or "all" for all targets
func writeAllTiersForTargets(brainRoot string, target string) {
	brain := scanBrain(brainRoot)
	result := runSubsumption(brain)

	dropped := applyOOMProtection(brainRoot, &result)
	if dropped > 0 {
		fmt.Printf("\033[33m[WARNING] OOM Limit. Dropped %d low-weight neurons.\033[0m\n", dropped)
	}

	// Generate bootstrap content (same for all targets)
	bootstrap := emitBootstrap(result, brainRoot)

	// Find project root (parent of brain)
	projectRoot := filepath.Dir(brainRoot)

	// Load dynamic targets from JSON
	targetMap := loadLLMProviders(brainRoot)

	// Determine which targets to write
	var targets []string
	if target == "all" {
		for k := range targetMap {
			targets = append(targets, k)
		}
		sort.Strings(targets)
	} else if target == "auto" {
		// Auto-detect: only emit to editors whose config files already exist
		autoDetectMap := make(map[string][]string)
		for k, t := range targetMap {
			var resolvedPaths []string
			for _, dp := range t.DetectPaths {
				if t.IsGlobal {
					resolvedPaths = append(resolvedPaths, filepath.Join(homeDir(), dp))
				} else {
					resolvedPaths = append(resolvedPaths, filepath.Join(projectRoot, dp))
				}
			}
			autoDetectMap[k] = resolvedPaths
		}

		for key, paths := range autoDetectMap {
			for _, p := range paths {
				if _, err := os.Stat(p); err == nil {
					targets = append(targets, key)
					break
				}
			}
		}
		sort.Strings(targets)
		if len(targets) == 0 {
			// Nothing detected — fall back to all
			fmt.Printf("[AUTO] No existing editor configs detected. Emitting all targets.\n")
			for k := range targetMap {
				targets = append(targets, k)
			}
			sort.Strings(targets)
		} else {
			fmt.Printf("[AUTO] 🔍 Detected %d editor(s): %s\n", len(targets), strings.Join(targets, ", "))
		}
	} else {
		targets = []string{target}
	}

	// Track backups for summary
	var backedUp []string

	// Write to each target
	for _, t := range targets {
		et, ok := targetMap[t]
		if !ok {
			// fallback check
			fmt.Printf("[WARN] Unknown emit target: %s\n", t)
			continue
		}

		var targetPath string
		var baseDir string
		
		if et.IsGlobal {
			homeDir, _ := os.UserHomeDir()
			if mock := os.Getenv("NEURONFS_MOCK_HOME"); mock != "" {
				homeDir = mock
			}
			baseDir = homeDir
		} else {
			baseDir = projectRoot
		}

		if et.SubDir != "" {
			subDir := filepath.Join(baseDir, et.SubDir)
			os.MkdirAll(subDir, 0750)
			targetPath = filepath.Join(subDir, et.FileName)
		} else {
			targetPath = filepath.Join(baseDir, et.FileName)
		}

		// ── Auto-backup existing file before overwrite ──
		if bkPath := backupExistingRule(targetPath, brainRoot); bkPath != "" {
			backedUp = append(backedUp, bkPath)
			fmt.Printf("\033[33m[BACKUP] 💾 %s → %s\033[0m\n", filepath.Base(targetPath), bkPath)
		}

		// 전체 덮어쓰기
		if err := os.WriteFile(targetPath, []byte(bootstrap), 0600); err != nil {
			fmt.Printf("[ERROR] Cannot write %s: %v\n", targetPath, err)
			continue
		}

		fmt.Printf("[EMIT] ✅ %s → %s\n", et.Name, targetPath)
	}

	// Also write Tier 2 + 3 (these are editor-independent)
	indexContent := emitIndex(brain, result)
	indexPath := filepath.Join(brainRoot, "_index.md")
	if err := os.WriteFile(indexPath, []byte(indexContent), 0600); err != nil {
		fmt.Printf("[WARN] Cannot write %s: %v\n", indexPath, err)
	}

	tier4Count := 0
	for _, region := range brain.Regions {
		content := emitRegionRules(region, brain)
		rulesPath := filepath.Join(region.Path, "_rules.md")
		if err := os.WriteFile(rulesPath, []byte(content), 0600); err != nil {
			fmt.Printf("[WARN] Cannot write %s: %v\n", rulesPath, err)
		}

		// ━━━ Tier 4: Subcategory _rules.md for large categories ━━━
		// 10+ 뉴런을 가진 하위 카테고리에 독립 _rules.md 생성
		subMap := make(map[string][]Neuron) // subcategory name → neurons
		for _, n := range region.Neurons {
			if n.IsDormant {
				continue
			}
			// n.Path = "dev/禁/하드코딩" → subcat = "dev"
			relPath := n.Path
			parts := strings.Split(relPath, string(filepath.Separator))
			if len(parts) < 2 {
				// Also handle forward slash
				parts = strings.Split(relPath, "/")
			}
			if len(parts) >= 2 {
				subMap[parts[0]] = append(subMap[parts[0]], n)
			}
		}
		for subName, neurons := range subMap {
			if len(neurons) < 10 {
				continue // 소규모 카테고리는 영역 _rules.md만으로 충분
			}
			// 하위 카테고리에 대한 미니 Region을 생성하여 Tier 4 _rules.md 작성
			subRegion := Region{
				Name:    region.Name + "/" + subName,
				Path:    filepath.Join(region.Path, subName),
				Neurons: neurons,
			}
			subContent := emitRegionRules(subRegion)
			subRulesPath := filepath.Join(region.Path, subName, "_rules.md")
			if err := os.WriteFile(subRulesPath, []byte(subContent), 0600); err == nil {
				tier4Count++
			}
		}
	}

	generateBrainJSON(brainRoot, brain, result)

	if len(backedUp) > 0 {
		fmt.Printf("\033[33m[WARNING] ⚠️  %d existing rule file(s) were backed up to: %s\033[0m\n",
			len(backedUp), filepath.Join(brainRoot, ".neuronfs_backup"))
		fmt.Printf("\033[33m[WARNING] To restore: copy .bak files back to their original locations.\033[0m\n")
	}
	fmt.Printf("[SYNC] ♻️  emit complete: %d target(s) + _index.md + 7x _rules.md + %dx tier4 (%d neurons, activation: %d)\n",
		len(targets), tier4Count, result.FiredNeurons, result.TotalCounter)
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// doInjectToFile: 기존 파일에 NEURONFS 마커 구간 교체
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

func doInjectToFile(filePath string, rules string) {
	existing, err := os.ReadFile(filePath)
	if err != nil {
		// File doesn't exist — create with just the rules
		os.MkdirAll(filepath.Dir(filePath), 0750)
		os.WriteFile(filePath, []byte(rules), 0600)
		return
	}

	content := string(existing)
	startMarker := "<!-- NEURONFS:START -->"
	endMarker := "<!-- NEURONFS:END -->"

	startIdx := strings.Index(content, startMarker)
	endIdx := strings.Index(content, endMarker)

	if startIdx >= 0 && endIdx >= 0 && endIdx > startIdx {
		// START 앞의 기존 preamble + END 뒤 푸터 보존
		content = content[:startIdx] + rules + content[endIdx+len(endMarker):]
	} else {
		content = rules + "\n\n" + content
	}

	tmpPath := filePath + ".tmp"
	if err := os.WriteFile(tmpPath, []byte(content), 0600); err == nil {
		os.Rename(tmpPath, filePath) // Atomic replace
	} else {
		os.WriteFile(filePath, []byte(content), 0600)
	}
}
