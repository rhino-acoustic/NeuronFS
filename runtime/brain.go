package main

// ━━━ brain.go ━━━
// Module: Brain Data Structures + Scanner
//
// PROVIDES:
//   Neuron, Region, Brain, SubsumptionResult (structs)
//   regionPriority, regionIcons, regionKo (maps)
//   counterRegex, dopamineRegex (regex)
//   scanBrain, runSubsumption, findBrainRoot,
//   getFolderBirthTime, activationBar
//
// CONSUMED BY:
//   ALL files — core data structures used everywhere
//
// DEPENDS ON:
//   (stdlib only — foundational module)

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Region metadata → governance_consts.go (SSOT)
var regionPriority = RegionPriority
var regionIcons = RegionIcons
var regionKo = RegionKo

// ─── Neuron = a folder ───
type Neuron struct {
	Name        string    // folder name
	Path        string    // relative path from region root (e.g. "frontend/css/glass_blur20")
	FullPath    string    // absolute path
	Counter     int       // from N.neuron filename (correction count)
	Contra      int       // from N.contra filename (inhibition count)
	Dopamine    int       // from dopamineN.neuron filename (reward count)
	Intensity   int       // Counter - Contra + Dopamine (net activation)
	Polarity    float64   // net / total (-1.0=pure inhibition, +1.0=pure excitation)
	HasBomb     bool      // bomb.neuron exists
	HasMemory   bool      // memoryN.neuron exists
	HasGoal     bool      // .goal file exists (todo/objective)
	GoalText    string    // content of .goal file if present
	Geofence    string    // content of .geofence file if present
	Description string    // natural language rule from rule.md (first line or description: field)
	Globs       string    // file pattern scope from rule.md (e.g. "*.go")
	Author      string    // agent identity for data lineage (e.g. "gemini-2.5", "claude-3.5")
	IsDormant   bool      // .dormant file exists
	Depth       int       // depth within region
	ModTime     time.Time // most recent .neuron file modification
	BirthTime   time.Time // folder creation time (grow 시점)
}

// Emission thresholds → governance_consts.go (SSOT)
var (
	emitThreshold = EmitThreshold
	spotlightDays = SpotlightDays
)

// ─── Region ───
type Region struct {
	Name     string
	Priority int
	Path     string
	Neurons  []Neuron
	Axons    []string // .axon targets
	HasBomb  bool     // any neuron in this region has bomb
}

// ─── Brain ───
type Brain struct {
	Root    string
	Regions []Region
}

// ─── Subsumption Result ───
type SubsumptionResult struct {
	ActiveRegions  []Region
	BlockedRegions []string
	BombSource     string
	FiredNeurons   int
	TotalNeurons   int
	TotalCounter   int
}

// ─── Regex for trace files ───
var counterRegex = regexp.MustCompile(`^(\d+)\.neuron$`)
var dopamineRegex = regexp.MustCompile(`^dopamine(\d+)\.neuron$`)

// ─── Find brain root ───
func findBrainRoot() string {
	// First non-flag arg
	for _, arg := range os.Args[1:] {
		if !strings.HasPrefix(arg, "--") {
			abs, err := filepath.Abs(arg)
			if err == nil {
				if info, err := os.Stat(abs); err == nil && info.IsDir() {
					return abs
				}
			}
		}
	}
	// Fallback: look for brain/ or brain_v4/ nearby
	home := os.Getenv("USERPROFILE")
	candidates := []string{
		"brain_v4", "brain",
		filepath.Join("..", "brain_v4"), filepath.Join("..", "brain"),
	}
	if home != "" {
		candidates = append(candidates,
			filepath.Join(home, "NeuronFS", "brain_v4"),
			filepath.Join(home, "NeuronFS", "brain"),
		)
	}
	for _, c := range candidates {
		abs, _ := filepath.Abs(c)
		if info, err := os.Stat(abs); err == nil && info.IsDir() {
			return abs
		}
	}
	return ""
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// getFolderBirthTime returns the folder's creation time (Windows CreationTime).
// This is used for recency boost — new neurons get protection based on when they were born,
// not when they were last fired.
func getFolderBirthTime(folderPath string) time.Time {
	fi, err := vfsStat(folderPath)
	if err != nil {
		return time.Now()
	}
	if sys := fi.Sys(); sys != nil {
		if d, ok := sys.(*syscall.Win32FileAttributeData); ok {
			return time.Unix(0, d.CreationTime.Nanoseconds())
		}
	}
	return fi.ModTime() // fallback: non-Windows or VFS
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// BRAIN CACHE (Phase 19: High-speed In-Memory sync.Map)
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
var rootBrainCache sync.Map

type brainCacheEntry struct {
	Brain Brain
	Time  time.Time
}

func InvalidateBrainCache(root string) {
	rootBrainCache.Delete(root)
}

// SCAN: Folder tree → Brain structure
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
func scanBrain(root string) Brain {
	if os.Getenv("NEURONFS_TEST_ISOLATION") != "1" {
		if entry, ok := rootBrainCache.Load(root); ok {
			ce := entry.(*brainCacheEntry)
			if time.Since(ce.Time) < 1500*time.Millisecond {
				return ce.Brain
			}
		}
	}

	brain := Brain{Root: root}

	regionsToScan := make(map[string]string)
	entries, err := vfsReadDir(root)
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				name := entry.Name()
				if _, ok := regionPriority[name]; ok {
					regionsToScan[name] = filepath.Join(root, name)
				}
			}
		}
	}

	sharedPath := filepath.Join(root, ".neuronfs", "shared")
	if _, err := vfsStat(sharedPath); err == nil {
		regionsToScan["shared"] = sharedPath
		// regionPriority, regionIcons, regionKo는 이제 SSOT(governance_consts.go)에 영구 정의됨
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	for name, regionPath := range regionsToScan {
		wg.Add(1)
		go func(name, regionPath string) {
			defer wg.Done()
			priority := regionPriority[name]

		region := Region{
			Name:     name,
			Priority: priority,
			Path:     regionPath,
		}

		// Scan for .axon files at region root
		axonFiles, _ := vfsGlob(filepath.Join(regionPath, "*.axon"))
		for _, af := range axonFiles {
			content, _ := vfsReadFile(af)
			target := strings.TrimSpace(string(content))
			// Remove UTF-8 BOM if present
			target = strings.TrimPrefix(target, "\xEF\xBB\xBF")
			target = strings.TrimPrefix(target, "TARGET: ")
			if target != "" {
				region.Axons = append(region.Axons, target)
			}
		}

		// Scan flat neurons at region root (e.g., brainstem: 禁fallback.1.neuron)
		// Pattern: NeuronName.Counter.neuron or NeuronName.neuron
		flatNeuronRegex := regexp.MustCompile(`^(.+)\.(\d+)\.neuron$`)
		flatNeuronSimple := regexp.MustCompile(`^(.+)\.neuron$`)
		rootNeuronFiles, _ := vfsGlob(filepath.Join(regionPath, "*.neuron"))
		neuronMap := make(map[string]*Neuron) // group by neuron name (Path)
		for _, nf := range rootNeuronFiles {
			fname := filepath.Base(nf)
			var neuronName string
			var counter int

			if m := flatNeuronRegex.FindStringSubmatch(fname); m != nil {
				neuronName = m[1]
				counter, _ = strconv.Atoi(m[2])
			} else if m := flatNeuronSimple.FindStringSubmatch(fname); m != nil {
				neuronName = m[1]
				if neuronName == "bomb" || strings.HasPrefix(neuronName, "dopamine") || strings.HasPrefix(neuronName, "memory") {
					continue
				}
			} else {
				continue
			}

			if existing, ok := neuronMap[neuronName]; ok {
				if counter > existing.Counter {
					existing.Counter = counter
				}
			} else {
				n := &Neuron{
					Name:     neuronName,
					Path:     neuronName,
					FullPath: filepath.Join(regionPath, neuronName),
					Depth:    0,
					Counter:  counter,
				}
				if fileInfo, err := vfsStat(nf); err == nil {
					n.ModTime = fileInfo.ModTime()
				} else {
					n.ModTime = time.Now()
				}
				// BirthTime = 폴더 생성 시간 (grow 시점)
				n.BirthTime = getFolderBirthTime(n.FullPath)
				if fname == "bomb.neuron" {
					n.HasBomb = true
					region.HasBomb = true
				} else if strings.HasPrefix(fname, "bomb_") {
					n.HasBomb = true
				}
				neuronMap[neuronName] = n
			}
		}

		// Walk for neuron folders — Axiom: Folder=Neuron, File=Trace
		vfsWalkDir(regionPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil || !d.IsDir() || path == regionPath {
				return nil
			}

			baseName := filepath.Base(path)
			if strings.HasPrefix(baseName, "_") || strings.HasPrefix(baseName, ".") || baseName == "working_memory" {
				if baseName == "_sandbox" {
					return nil
				}
				return filepath.SkipDir
			}

			// 한자 옵코드 폴더(禁/必/推 등)는 구조 폴더 — 뉴런이 아님
			// neuronMap에 등록하지 않고 Walk만 계속 (하위 폴더는 등록)
			if isHanjaFolder(baseName) {
				return nil // skip registration, continue walking children
			}

			relPath, _ := filepath.Rel(regionPath, path)
			depth := strings.Count(relPath, string(filepath.Separator))

			n, exists := neuronMap[relPath]
			var dModTime time.Time
			if dInfo, err := d.Info(); err == nil {
				dModTime = dInfo.ModTime()
			} else {
				dModTime = time.Now()
			}

			if !exists {
				n = &Neuron{
					Name:      baseName,
					Path:      relPath,
					FullPath:  path,
					Depth:     depth,
					ModTime:   dModTime,
					BirthTime: getFolderBirthTime(path),
				}
				neuronMap[relPath] = n
			} else {
				n.FullPath = path
				n.Depth = depth
				if dModTime.After(n.ModTime) {
					n.ModTime = dModTime
				}
			}

			entries, _ := vfsReadDir(path)
			var neuronFiles []string
			var goalFiles []string
			var geofenceFiles []string
			var dormantFiles []string

			for _, e := range entries {
				if e.IsDir() {
					continue
				}
				name := e.Name()
				fullf := filepath.Join(path, name)
				
				if strings.HasSuffix(name, ".neuron") || strings.HasSuffix(name, ".contra") {
					neuronFiles = append(neuronFiles, fullf)
				} else if strings.HasSuffix(name, ".goal") {
					goalFiles = append(goalFiles, fullf)
				} else if strings.HasSuffix(name, ".geofence") {
					geofenceFiles = append(geofenceFiles, fullf)
				} else if strings.HasSuffix(name, ".dormant") {
					dormantFiles = append(dormantFiles, fullf)
				}
			}

			for _, nf := range neuronFiles {
				fname := filepath.Base(nf)
				if nfInfo, err := vfsStat(nf); err == nil {
					if nfInfo.ModTime().After(n.ModTime) {
						n.ModTime = nfInfo.ModTime()
					}
				}

				if m := counterRegex.FindStringSubmatch(fname); m != nil {
					cnt, _ := strconv.Atoi(m[1])
					if cnt > n.Counter {
						n.Counter = cnt
					}
				}

				if m := dopamineRegex.FindStringSubmatch(fname); m != nil {
					cnt, _ := strconv.Atoi(m[1])
					n.Dopamine += cnt
				}

				if strings.HasSuffix(fname, ".contra") && region.Name != "brainstem" {
					base := strings.TrimSuffix(fname, ".contra")
					if cnt, err := strconv.Atoi(base); err == nil && cnt > n.Contra {
						n.Contra = cnt
					}
				}

				if fname == "bomb.neuron" {
					n.HasBomb = true
					region.HasBomb = true
				}
				if strings.HasPrefix(fname, "memory") {
					n.HasMemory = true
				}
			}

			if len(goalFiles) > 0 {
				n.HasGoal = true
				if content, err := vfsReadFile(goalFiles[0]); err == nil && len(content) > 0 {
					n.GoalText = strings.TrimSpace(string(content))
				}
			}

			if len(geofenceFiles) > 0 {
				if content, err := vfsReadFile(geofenceFiles[0]); err == nil && len(content) > 0 {
					n.Geofence = strings.TrimSpace(string(content))
				}
			}

			if len(dormantFiles) > 0 {
				n.IsDormant = true
			}

			return nil
		})

		// Finalize compute elements and ensure deterministic order
		var paths []string
		for path := range neuronMap {
			paths = append(paths, path)
		}
		sort.Strings(paths)

		for _, path := range paths {
			n := neuronMap[path]
			n.Intensity = n.Counter - n.Contra + n.Dopamine
			totalSignals := n.Counter + n.Contra + n.Dopamine
			if totalSignals > 0 {
				n.Polarity = float64(n.Counter+n.Dopamine-n.Contra) / float64(totalSignals)
			} else {
				n.Polarity = 0.5
			}

			// ━━━ Natural Language Rule: rule.md → Description (post-Walk) ━━━
			// Priority: rule.md frontmatter > rule.md first line > .neuron filename > empty
			if n.Description == "" && n.FullPath != "" {
				ruleFile := filepath.Join(n.FullPath, "rule.md")
				if ruleContent, err := vfsReadFile(ruleFile); err == nil {
					ruleText := strings.TrimSpace(string(ruleContent))
					inFM := false
					for _, line := range strings.Split(ruleText, "\n") {
						line = strings.TrimSpace(line)
						if line == "---" {
							inFM = !inFM
							continue
						}
						if inFM {
							if strings.HasPrefix(line, "description:") {
								n.Description = strings.TrimSpace(strings.TrimPrefix(line, "description:"))
							}
							if strings.HasPrefix(line, "globs:") {
								n.Globs = strings.TrimSpace(strings.TrimPrefix(line, "globs:"))
							}
							if strings.HasPrefix(line, "author:") {
								n.Author = strings.TrimSpace(strings.TrimPrefix(line, "author:"))
							}
						}
					}
					// Fallback: first content line
					if n.Description == "" {
						for _, line := range strings.Split(ruleText, "\n") {
							line = strings.TrimSpace(line)
							if strings.HasPrefix(line, "#") {
								line = strings.TrimSpace(strings.TrimPrefix(line, "#"))
								line = strings.TrimSpace(strings.TrimPrefix(line, "#"))
							}
							if line != "" && line != "---" {
								runes := []rune(line)
								if len(runes) > 150 {
									line = string(runes[:150]) + "..."
								}
								n.Description = line
								break
							}
						}
					}
				}
				// Fallback 2: .neuron 파일 본문 첫 줄 또는 파일 이름
				if n.Description == "" && n.FullPath != "" {
					entries, _ := os.ReadDir(n.FullPath)
					for _, e := range entries {
						if !e.IsDir() && strings.HasSuffix(e.Name(), ".neuron") {
							// 1) 본문 내용 파싱 시도
							if content, err := os.ReadFile(filepath.Join(n.FullPath, e.Name())); err == nil {
								for _, line := range strings.Split(string(content), "\n") {
									line = strings.TrimSpace(line)
									// BOM 제어
									line = strings.TrimPrefix(line, "\xEF\xBB\xBF")
									// JSON 및 시스템 메타데이터 스킵
									if strings.HasPrefix(line, "{") || strings.HasPrefix(line, "counter:") || strings.HasPrefix(line, "created:") || strings.HasPrefix(line, "last_fired:") {
										continue
									}
									// Markdown 헤더 기호 제거
									if strings.HasPrefix(line, "#") {
										line = strings.TrimSpace(strings.TrimPrefix(line, "#"))
										line = strings.TrimSpace(strings.TrimPrefix(line, "#")) // ## 지원
									}
									if line != "" && line != "---" {
										runes := []rune(line)
										if len(runes) > 150 {
											line = string(runes[:150]) + "..."
										}
										n.Description = line
										break
									}
								}
							}
							
							// 2) 본문에 없으면 파일 이름 파싱
							if n.Description == "" {
								name := strings.TrimSuffix(e.Name(), ".neuron")
								isReserved := true
								for _, r := range name {
									if r < '0' || r > '9' {
										isReserved = false
										break
									}
								}
								if name == "consolidated" || name == "merged" {
									isReserved = true
								}
								if !isReserved && len(name) > 2 {
									n.Description = strings.ReplaceAll(name, "_", " ")
								}
							}
							if n.Description != "" {
								break
							}
						}
					}
				}
			}

			// Clean mojibake/encoding issues from description
			if n.Description != "" && isMojibake(n.Description) {
				n.Description = ""
			}

			region.Neurons = append(region.Neurons, *n)
		}

		mu.Lock()
		brain.Regions = append(brain.Regions, region)
		mu.Unlock()
	}(name, regionPath)
	}

	wg.Wait()

	// Sort regions by priority
	sort.Slice(brain.Regions, func(i, j int) bool {
		return brain.Regions[i].Priority < brain.Regions[j].Priority
	})

	rootBrainCache.Store(root, &brainCacheEntry{
		Brain: brain,
		Time:  time.Now(),
	})

	return brain
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// SUBSUMPTION: Priority cascade + bomb detection
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
func runSubsumption(brain Brain) SubsumptionResult {
	result := SubsumptionResult{}
	blocked := false

	for _, region := range brain.Regions {
		activeNeurons := 0
		for _, n := range region.Neurons {
			if !n.IsDormant && (n.Counter+n.Dopamine) > 0 {
				activeNeurons++
				result.TotalCounter += n.Counter
			}
		}
		result.TotalNeurons += activeNeurons

		if blocked {
			result.BlockedRegions = append(result.BlockedRegions, region.Name)
			continue
		}

		if region.HasBomb {
			result.BombSource = region.Name
			result.BlockedRegions = append(result.BlockedRegions, region.Name+" [BOMB]")
			blocked = true

			// Physical Hook Trigger (e.g., ring alarm or strict kill process)
			// Triggered when a bomb is found in the geofenced context.
			triggerPhysicalHook(region.Name)
			continue
		}

		result.ActiveRegions = append(result.ActiveRegions, region)
		result.FiredNeurons += activeNeurons
	}

	return result
}

// emitRules is a compatibility wrapper for the tiered system.
// Returns bootstrap content (Tier 1) for GEMINI.md injection.
// The full tier system (index + per-region) is handled by writeAllTiers.
