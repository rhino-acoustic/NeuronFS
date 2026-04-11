// hebbian.go — Hebbian Co-activation Learning
// "함께 발화하는 뉴런은 함께 강화된다" (Neurons that fire together, wire together)
//
// PROVIDES: hebbianTrack
// DEPENDS ON: neuron_crud.go (fireNeuron calls this)
//
// 설계철학: Folder=Neuron, File=Trace 공리를 준수.
// 새 파일 타입 없이 hippocampus/session_log/co-activation.jsonl에 JSONL 기록.
// evolve 프롬프트가 이 데이터를 참조하여 뉴런 간 연관성을 학습한다.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ── Co-activation Ring Buffer (최근 30초 내 발화된 뉴런 추적) ──
var (
	hebbianMu      sync.Mutex
	hebbianRecent  []hebbianEntry
	hebbianMaxAge  = 30 * time.Second
	hebbianLogFile = "co-activation.jsonl"
)

type hebbianEntry struct {
	path string
	ts   time.Time
}

type coActivation struct {
	Ts    string `json:"ts"`
	A     string `json:"a"`
	B     string `json:"b"`
	Delta string `json:"delta_ms"`
}

// hebbianTrack records co-activation pairs for neurons fired within 30 seconds.
// Called after each fire operation.
func hebbianTrack(brainRoot, neuronPath string) {
	hebbianMu.Lock()
	defer hebbianMu.Unlock()

	now := time.Now()

	// 30초 초과 엔트리 제거
	cutoff := now.Add(-hebbianMaxAge)
	valid := hebbianRecent[:0]
	for _, e := range hebbianRecent {
		if e.ts.After(cutoff) {
			valid = append(valid, e)
		}
	}
	hebbianRecent = valid

	// 30초 내 발화된 뉴런들과 co-activation 기록
	logPath := filepath.Join(brainRoot, "hippocampus", "session_log", hebbianLogFile)
	os.MkdirAll(filepath.Dir(logPath), 0750)

	for _, prev := range hebbianRecent {
		if prev.path == neuronPath {
			continue // 자기 자신 제외
		}
		delta := now.Sub(prev.ts)
		record := coActivation{
			Ts:    now.Format("2006-01-02_15:04:05"),
			A:     prev.path,
			B:     neuronPath,
			Delta: fmt.Sprintf("%dms", delta.Milliseconds()),
		}
		data, _ := json.Marshal(record)

		f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err == nil {
			fmt.Fprintln(f, string(data))
			f.Close()
		}
	}

	// 현재 뉴런을 ring buffer에 추가
	hebbianRecent = append(hebbianRecent, hebbianEntry{path: neuronPath, ts: now})

	// Ring buffer 최대 20개
	if len(hebbianRecent) > 20 {
		hebbianRecent = hebbianRecent[len(hebbianRecent)-20:]
	}
}
