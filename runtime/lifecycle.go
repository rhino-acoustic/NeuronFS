package main

// ━━━ lifecycle.go ━━━
// Module: Neuron Lifecycle Management (prune, decay, dedup, episode logging)
//
// PROVIDES:
//   pruneWeakNeurons, runDecay, logEpisode, deduplicateNeurons
//
// CONSUMED BY:
//   main.go         → runIdleLoop() calls all lifecycle functions
//   dashboard.go    → health_check triggers deduplicateNeurons()
//
// DEPENDS ON:
//   similarity.go   → hybridSimilarity(), tokenize(), extractPrefix()
//   main.go         → scanBrain(), writeAllTiers(), markBrainDirty()
//   main.go         → counterRegex, regionPriority (package vars)

import (
"fmt"
"os"
"path/filepath"
"regexp"
"sort"
"strconv"
"strings"
"time"
)

// pruneWeakNeurons marks 推-prefix neurons with activation ≤1 and 3+ days inactive as dormant.
// 추천 뉴런은 우선순위가 낮다 — fire되지 않으면 자연 도태.
func pruneWeakNeurons(brainRoot string) {
	cutoff := time.Now().AddDate(0, 0, -3) // 3일
	pruned := 0

	for _, regionName := range []string{"cortex", "ego", "prefrontal", "limbic", "hippocampus", "sensors"} {
		regionPath := filepath.Join(brainRoot, regionName)
		if _, err := os.Stat(regionPath); os.IsNotExist(err) {
			continue
		}

		filepath.Walk(regionPath, func(path string, info os.FileInfo, err error) error {
			if err != nil || !info.IsDir() || path == regionPath {
				return nil
			}

			leafName := filepath.Base(path)
			runes := []rune(leafName)
			if len(runes) == 0 || runes[0] != '推' {
				return nil // 推 접두어만 대상
			}

			neuronFiles, _ := filepath.Glob(filepath.Join(path, "*.neuron"))
			if len(neuronFiles) == 0 {
				return nil
			}

			// 이미 dormant면 스킵
			dormantFiles, _ := filepath.Glob(filepath.Join(path, "*.dormant"))
			if len(dormantFiles) > 0 {
				return nil
			}

			// 카운터 확인
			maxCounter := 0
			var newestMod time.Time
			for _, nf := range neuronFiles {
				base := filepath.Base(nf)
				if m := counterRegex.FindStringSubmatch(base); m != nil {
					n := 0
					fmt.Sscanf(m[1], "%d", &n)
					if n > maxCounter {
						maxCounter = n
					}
				}
				if fi, err := os.Stat(nf); err == nil && fi.ModTime().After(newestMod) {
					newestMod = fi.ModTime()
				}
			}

			// 활성 ≤1 AND 3일+ 미사용 → dormant
			if maxCounter <= 1 && !newestMod.IsZero() && newestMod.Before(cutoff) {
				df := filepath.Join(path, "prune.dormant")
				os.WriteFile(df, []byte(fmt.Sprintf("Pruned: %s\nReason: 推 prefix, activation=%d, inactive %d days\n",
					time.Now().Format("2006-01-02"),
					maxCounter,
					int(time.Since(newestMod).Hours()/24))), 0644)

				relPath, _ := filepath.Rel(brainRoot, path)
				fmt.Printf("[PRUNE] 🪦 %s (activation=%d, %d days idle)\n", relPath, maxCounter,
					int(time.Since(newestMod).Hours()/24))
				pruned++
			}
			return nil
		})
	}

	if pruned > 0 {
		fmt.Printf("[PRUNE] ✅ %d 건 推 뉴런 dormant 처리\n", pruned)
		logEpisode(brainRoot, "PRUNE", fmt.Sprintf("%d weak 推 neurons dormant", pruned))
	} else {
		fmt.Println("[PRUNE] ✓ 도태 대상 없음")
	}
}

// runDecay moves neurons untouched for N days to dormant state
// Usage: neuronfs brain_v4 --decay 30
func runDecay(brainRoot string, days int) {
	cutoff := time.Now().AddDate(0, 0, -days)
	decayed := 0
	total := 0

	// brainstem 제외: P0 거버넌스 규칙(禁/必)은 영구 — decay 대상이 아님
	for _, regionName := range []string{"limbic", "hippocampus", "sensors", "cortex", "ego", "prefrontal"} {
		regionPath := filepath.Join(brainRoot, regionName)
		if _, err := os.Stat(regionPath); os.IsNotExist(err) {
			continue
		}

		filepath.Walk(regionPath, func(path string, info os.FileInfo, err error) error {
			if err != nil || !info.IsDir() || path == regionPath {
				return nil
			}

			// 禁/必 접두어 뉴런은 거버넌스 → decay 면제
			leafName := filepath.Base(path)
			for _, h := range []string{"禁", "必"} {
				if strings.HasPrefix(leafName, h) {
					return nil
				}
			}

			// Check if this is a neuron folder (has .neuron files)
			neuronFiles, _ := filepath.Glob(filepath.Join(path, "*.neuron"))
			if len(neuronFiles) == 0 {
				return nil
			}
			total++

			// Skip if already dormant
			dormantFiles, _ := filepath.Glob(filepath.Join(path, "*.dormant"))
			if len(dormantFiles) > 0 {
				return nil
			}

			// Find the most recent .neuron file modification time
			var newestMod time.Time
			for _, nf := range neuronFiles {
				fi, err := os.Stat(nf)
				if err == nil && fi.ModTime().After(newestMod) {
					newestMod = fi.ModTime()
				}
			}

			if !newestMod.IsZero() && newestMod.Before(cutoff) {
				// Mark as dormant
				df := filepath.Join(path, "decay.dormant")
				os.WriteFile(df, []byte(fmt.Sprintf("Decayed: %s\nLast active: %s\nThreshold: %d days\n",
					time.Now().Format("2006-01-02"),
					newestMod.Format("2006-01-02"),
					days)), 0644)

				relPath, _ := filepath.Rel(brainRoot, path)
				ageDays := int(time.Since(newestMod).Hours() / 24)
				fmt.Printf("[DECAY] 💤 %s (inactive %d days)\n", relPath, ageDays)
				decayed++
			}

			return nil
		})
	}

	fmt.Printf("[DECAY] Scanned %d neurons, decayed %d (threshold: %d days)\n", total, decayed, days)

	if decayed > 0 {
		logEpisode(brainRoot, "DECAY", fmt.Sprintf("%d neurons dormant (>%d days)", decayed, days))
		markBrainDirty()
	}
}

// logEpisode records an event in hippocampus/session_log
// Circular buffer: keeps only the most recent 10 episodes
const maxEpisodes = 10

// logEpisode writes an event log to the hippocampus memory store.
func logEpisode(brainRoot string, event string, detail string) {
	logDir := filepath.Join(brainRoot, "hippocampus", "session_log")
	os.MkdirAll(logDir, 0755)

	memRegex := regexp.MustCompile(`^memory(\d+)\.neuron$`)
	entries, _ := os.ReadDir(logDir)

	// Collect all memory files with their numbers
	type memEntry struct {
		num  int
		name string
	}
	var mems []memEntry
	for _, e := range entries {
		if m := memRegex.FindStringSubmatch(e.Name()); m != nil {
			n, _ := strconv.Atoi(m[1])
			mems = append(mems, memEntry{num: n, name: e.Name()})
		}
	}

	// Sort by number ascending
	sort.Slice(mems, func(i, j int) bool { return mems[i].num < mems[j].num })

	// Evict oldest if at limit
	if len(mems) >= maxEpisodes {
		evictCount := len(mems) - maxEpisodes + 1
		for i := 0; i < evictCount; i++ {
			os.Remove(filepath.Join(logDir, mems[i].name))
		}
		fmt.Printf("[MEMORY] 🗑️ Evicted %d old episodes (circular buffer %d)\n", evictCount, maxEpisodes)
	}

	// Find next number
	nextN := 1
	if len(mems) > 0 {
		nextN = mems[len(mems)-1].num + 1
	}

	content := fmt.Sprintf("%s | %s | %s\n", time.Now().Format("2006-01-02T15:04:05"), event, detail)
	memFile := filepath.Join(logDir, fmt.Sprintf("memory%d.neuron", nextN))
	os.WriteFile(memFile, []byte(content), 0644)
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// deduplicateNeurons scans brain for semantically similar neuron folders
// and merges them: cross-region comparison enabled, P-priority survivor selection
// Uses hybrid similarity (Cosine Bigram 60% + Levenshtein 40%, >= 0.6) on folder names
// Generates axon files at victim location to maintain neural pathways
func deduplicateNeurons(brainRoot string) {
	brain := scanBrain(brainRoot)

	type neuronRef struct {
		fullPath string
		counter  int
		region   string
		relPath  string
		tokens   []string
		depth    int
		priority int // regionPriority (lower = stronger)
	}

	// Collect all active leaf neurons (brainstem 포함!)
	var allRefs []neuronRef
	for _, region := range brain.Regions {
		rp := regionPriority[region.Name]
		for _, n := range region.Neurons {
			if n.IsDormant {
				continue
			}
			leafName := filepath.Base(n.FullPath)
			tokens := tokenize(leafName)
			if len(tokens) == 0 {
				continue
			}
			allRefs = append(allRefs, neuronRef{
				fullPath: n.FullPath,
				counter:  n.Counter,
				region:   region.Name,
				relPath:  n.Path,
				tokens:   tokens,
				depth:    n.Depth,
				priority: rp,
			})
		}
	}

	// Compare all pairs (O(n²))
	merged := make(map[int]bool)
	mergeCount := 0

	for i := 0; i < len(allRefs); i++ {
		if merged[i] {
			continue
		}
		for j := i + 1; j < len(allRefs); j++ {
			if merged[j] {
				continue
			}

			sim := hybridSimilarity(allRefs[i].tokens, allRefs[j].tokens)
			if sim < 0.6 {
				continue
			}

			// 접두어 극성 보호: 禁X와 推X는 절대 병합 안 함
			prefixI := extractPrefix(filepath.Base(allRefs[i].fullPath))
			prefixJ := extractPrefix(filepath.Base(allRefs[j].fullPath))
			if prefixI != "" && prefixJ != "" && prefixI != prefixJ {
				continue // 극성 충돌 → 별개
			}

			// 유사도 0.6 이상 — 병합 대상
			// 생존자 선택 로직:
			//   1. 교차 영역: P가 낮은(=강한) 쪽이 생존
			//   2. 같은 영역: depth 깊거나 counter 높은 쪽
			survivorIdx := i
			victimIdx := j
			isCrossRegion := allRefs[i].region != allRefs[j].region

			if isCrossRegion {
				// P가 낮은 쪽 생존 (brainstem P0 > ego P5)
				if allRefs[j].priority < allRefs[i].priority {
					survivorIdx, victimIdx = j, i
				}
			} else {
				// 같은 영역: depth/counter 비교
				if allRefs[j].depth > allRefs[i].depth ||
					(allRefs[j].depth == allRefs[i].depth && allRefs[j].counter > allRefs[i].counter) {
					survivorIdx, victimIdx = j, i
				}
			}
			merged[victimIdx] = true
			survivor := &allRefs[survivorIdx]
			victim := &allRefs[victimIdx]

			// 카운터 합산 + 보너스 (+1)
			totalCounter := survivor.counter + victim.counter + 1

			mergeTag := "SAME"
			if isCrossRegion {
				mergeTag = "CROSS"
			}
			fmt.Printf("[DEDUP] 🔀 [%s] 병합 (sim=%.2f): %s/%s (%d) ← %s/%s (%d) → %d\n",
				mergeTag, sim,
				survivor.region, filepath.Base(survivor.fullPath), survivor.counter,
				victim.region, filepath.Base(victim.fullPath), victim.counter,
				totalCounter)

			// 생존자 카운터 갱신
			surviveFiles, _ := filepath.Glob(filepath.Join(survivor.fullPath, "*.neuron"))
			for _, f := range surviveFiles {
				base := filepath.Base(f)
				if counterRegex.MatchString(base) {
					os.Remove(f)
				}
			}
			newCounterFile := filepath.Join(survivor.fullPath, fmt.Sprintf("%d.neuron", totalCounter))
			os.WriteFile(newCounterFile, []byte(""), 0644)

			// victim의 dopamine/memory 시그널을 생존자로 이동
			victimFiles, _ := filepath.Glob(filepath.Join(victim.fullPath, "*.neuron"))
			for _, f := range victimFiles {
				base := filepath.Base(f)
				if strings.HasPrefix(base, "dopamine") || strings.HasPrefix(base, "memory") {
					destFile := filepath.Join(survivor.fullPath, base)
					if _, err := os.Stat(destFile); os.IsNotExist(err) {
						os.Rename(f, destFile)
					}
				}
			}

			// 교차 영역: victim 부모에 axon 파일 생성 (연결 보존)
			if isCrossRegion {
				victimParent := filepath.Dir(victim.fullPath)
				axonFileName := fmt.Sprintf("connect_%s.axon", survivor.region)
				axonPath := filepath.Join(victimParent, axonFileName)
				if _, err := os.Stat(axonPath); os.IsNotExist(err) {
					axonContent := fmt.Sprintf("target: %s\nmerged_from: %s\nsimilarity: %.2f\n",
						survivor.relPath, victim.relPath, sim)
					os.WriteFile(axonPath, []byte(axonContent), 0644)
					fmt.Printf("[AXON] 🔗 %s → %s\n", victim.region, survivor.region)
				}
			}

			// victim 폴더 삭제
			os.RemoveAll(victim.fullPath)
			survivor.counter = totalCounter
			mergeCount++

			logEpisode(brainRoot, "DEDUP", fmt.Sprintf("[%s] %s/%s → %s/%s (sim=%.2f)",
				mergeTag, victim.region, filepath.Base(victim.fullPath),
				survivor.region, filepath.Base(survivor.fullPath), sim))
		}
	}

	if mergeCount > 0 {
		fmt.Printf("[DEDUP] ✅ %d 건 중복 뉴런 병합 완료 (카운터 합산+보너스)\n", mergeCount)
		writeAllTiers(brainRoot)
	} else {
		fmt.Println("[DEDUP] ✓ 중복 뉴런 없음")
	}
}
