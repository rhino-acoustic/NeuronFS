// episode_auto.go — supervisor 이벤트의 자동 에피소드 기록
// lifecycle.go의 logEpisode를 호출하여 hippocampus에 시스템 이벤트를 기록
// SSOT: logEpisode는 lifecycle.go에 단일 정의
package main

import (
	"fmt"
	"os"
)

// logBootEpisode — 시스템 부팅 시 호출
func logBootEpisode(brainRoot string) {
	logEpisode(brainRoot, "BOOT", fmt.Sprintf("NeuronFS Supervisor 부팅 (PID %d)", os.Getpid()))
}

// logCrashEpisode — 프로세스 크래시 시 호출
func logCrashEpisode(brainRoot, processName string, restartCount int) {
	logEpisode(brainRoot, "CRASH", fmt.Sprintf("%s 서킷브레이커 (%d회 연속)", processName, restartCount))
}

// logHarnessEpisode — harness 결과 기록
func logHarnessEpisode(brainRoot string, pass bool, total int, fails []string) {
	if pass {
		logEpisode(brainRoot, "HARNESS", fmt.Sprintf("PASS (%d checks)", total))
	} else {
		summary := fmt.Sprintf("FAIL %d/%d", len(fails), total)
		if len(fails) > 0 {
			summary += ": " + fails[0]
		}
		logEpisode(brainRoot, "HARNESS", summary)
	}
}
