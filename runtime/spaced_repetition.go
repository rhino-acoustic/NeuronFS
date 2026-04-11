// spaced_repetition.go — Spaced Repetition for Neuron Reinforcement
// 오래 미사용된 고활성 뉴런을 자동으로 재발화하여 기억 유지.
// FSRS(Free Spaced Repetition Scheduler) 간소화 버전.
//
// PROVIDES: spacedRepetitionFire
// DEPENDS ON: brain.go (scanBrain), neuron_crud.go (fireNeuron)
// CALLED BY: transcript.go (runIdleLoop)
//
// 설계철학: 활성 뉴런이 사용되지 않으면 decay로 잊혀짐.
// Spaced Repetition은 중요한 뉴런(counter 높음)이 잊히지 않도록 주기적 재발화.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// spacedRepetitionFire: 14일 이상 미사용 + counter >= 10인 뉴런을 자동 fire (+1)
// 이미 dormant인 뉴런은 제외. brainstem(P0)은 영구이므로 제외.
func spacedRepetitionFire(brainRoot string) {
	brain := scanBrain(brainRoot)
	cutoff := time.Now().AddDate(0, 0, -14) // 14일 미사용
	reinforced := 0

	for _, region := range brain.Regions {
		if region.Name == "brainstem" {
			continue // P0 = 영구
		}
		for _, n := range region.Neurons {
			if n.IsDormant || n.Counter < 10 {
				continue
			}
			// modtime 확인
			entries, _ := os.ReadDir(n.FullPath)
			var newest time.Time
			for _, e := range entries {
				if fi, err := e.Info(); err == nil && fi.ModTime().After(newest) {
					newest = fi.ModTime()
				}
			}
			if newest.IsZero() || newest.After(cutoff) {
				continue // 최근 사용됨 — 재발화 불필요
			}

			// 재발화 (+1)
			relPath, _ := filepath.Rel(brainRoot, n.FullPath)
			fireNeuron(brainRoot, relPath)
			fmt.Printf("[SPACED] ♻️ Reinforced: %s (counter=%d, %dd idle)\n",
				relPath, n.Counter, int(time.Since(newest).Hours()/24))
			reinforced++

			// 한 사이클에 최대 5개만 재발화 (부하 방지)
			if reinforced >= 5 {
				break
			}
		}
		if reinforced >= 5 {
			break
		}
	}

	if reinforced > 0 {
		fmt.Printf("[SPACED] ✅ %d neurons reinforced via spaced repetition\n", reinforced)
		logEpisode(brainRoot, "SPACED_REPETITION", fmt.Sprintf("%d neurons reinforced", reinforced))
	}
}
