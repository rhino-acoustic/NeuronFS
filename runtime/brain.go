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
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var regionPriority = map[string]int{
	"brainstem":   0,
	"limbic":      1,
	"hippocampus": 2,
	"sensors":     3,
	"cortex":      4,
	"ego":         5,
	"prefrontal":  6,
}

var regionIcons = map[string]string{
	"brainstem":   "🛡️",
	"limbic":      "💓",
	"hippocampus": "📝",
	"sensors":     "👁️",
	"cortex":      "🧠",
	"ego":         "🎭",
	"prefrontal":  "🎯",
}

var regionKo = map[string]string{
	"brainstem":   "양심/본능",
	"limbic":      "감정 필터",
	"hippocampus": "기록/기억",
	"sensors":     "환경 제약",
	"cortex":      "지식/기술",
	"ego":         "성향/톤",
	"prefrontal":  "목표/계획",
}

// ─── Neuron = a folder ───
type Neuron struct {
	Name      string    // folder name
	Path      string    // relative path from region root (e.g. "frontend/css/glass_blur20")
	FullPath  string    // absolute path
	Counter   int       // from N.neuron filename (correction count)
	Contra    int       // from N.contra filename (inhibition count)
	Dopamine  int       // from dopamineN.neuron filename (reward count)
	Intensity int       // Counter - Contra + Dopamine (net activation)
	Polarity  float64   // net / total (-1.0=pure inhibition, +1.0=pure excitation)
	HasBomb   bool      // bomb.neuron exists
	HasMemory bool      // memoryN.neuron exists
	HasGoal   bool      // .goal file exists (todo/objective)
	GoalText  string    // content of .goal file if present
	Geofence  string    // content of .geofence file if present
	IsDormant bool      // .dormant file exists
	Depth     int       // depth within region
	ModTime   time.Time // most recent .neuron file modification
	BirthTime time.Time // folder creation time (grow 시점)
}

// Emission thresholds
const (
	emitThreshold = 5 // min counter to appear in region listings
	spotlightDays = 7 // days a new neuron gets spotlight regardless of counter
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
	fi, err := os.Stat(folderPath)
	if err != nil {
		return time.Now()
	}
	d, ok := fi.Sys().(*syscall.Win32FileAttributeData)
	if !ok {
		return fi.ModTime() // fallback: non-Windows
	}
	return time.Unix(0, d.CreationTime.Nanoseconds())
}

// SCAN: Folder tree → Brain structure
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
func scanBrain(root string) Brain {
	brain := Brain{Root: root}

	regionsToScan := make(map[string]string)
	entries, err := os.ReadDir(root)
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
	if _, err := os.Stat(sharedPath); err == nil {
		regionsToScan["shared"] = sharedPath
		regionPriority["shared"] = 7
		regionIcons["shared"] = "🔗"
		regionKo["shared"] = "공유 지식"
	}

	for name, regionPath := range regionsToScan {
		priority := regionPriority[name]

		region := Region{
			Name:     name,
			Priority: priority,
			Path:     regionPath,
		}

		// Scan for .axon files at region root
		axonFiles, _ := filepath.Glob(filepath.Join(regionPath, "*.axon"))
		for _, af := range axonFiles {
			content, _ := os.ReadFile(af)
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
		rootNeuronFiles, _ := filepath.Glob(filepath.Join(regionPath, "*.neuron"))
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
				if fileInfo, err := os.Stat(nf); err == nil {
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
		filepath.Walk(regionPath, func(path string, info os.FileInfo, err error) error {
			if err != nil || !info.IsDir() || path == regionPath {
				return nil
			}

			baseName := filepath.Base(path)
			if strings.HasPrefix(baseName, "_") || strings.HasPrefix(baseName, ".") {
				if baseName == "_sandbox" {
					return nil
				}
				return filepath.SkipDir
			}

			relPath, _ := filepath.Rel(regionPath, path)
			depth := strings.Count(relPath, string(filepath.Separator))

			n, exists := neuronMap[relPath]
			if !exists {
				n = &Neuron{
					Name:      baseName,
					Path:      relPath,
					FullPath:  path,
					Depth:     depth,
					ModTime:   info.ModTime(),
					BirthTime: getFolderBirthTime(path),
				}
				neuronMap[relPath] = n
			} else {
				n.FullPath = path
				n.Depth = depth
				if info.ModTime().After(n.ModTime) {
					n.ModTime = info.ModTime()
				}
			}

			neuronFiles, _ := filepath.Glob(filepath.Join(path, "*.neuron"))
			contraFiles, _ := filepath.Glob(filepath.Join(path, "*.contra"))
			neuronFiles = append(neuronFiles, contraFiles...)
			goalFiles, _ := filepath.Glob(filepath.Join(path, "*.goal"))

			for _, nf := range neuronFiles {
				fname := filepath.Base(nf)
				if nfInfo, err := os.Stat(nf); err == nil {
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
				if content, err := os.ReadFile(goalFiles[0]); err == nil && len(content) > 0 {
					n.GoalText = strings.TrimSpace(string(content))
				}
			}

			geofenceFiles, _ := filepath.Glob(filepath.Join(path, "*.geofence"))
			if len(geofenceFiles) > 0 {
				if content, err := os.ReadFile(geofenceFiles[0]); err == nil && len(content) > 0 {
					n.Geofence = strings.TrimSpace(string(content))
				}
			}

			dormantFiles, _ := filepath.Glob(filepath.Join(path, "*.dormant"))
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
			region.Neurons = append(region.Neurons, *n)
		}

		brain.Regions = append(brain.Regions, region)
	}

	// Sort regions by priority
	sort.Slice(brain.Regions, func(i, j int) bool {
		return brain.Regions[i].Priority < brain.Regions[j].Priority
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
			if !n.IsDormant {
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
