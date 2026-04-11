// hebbian.go — Hebbian Co-activation Learning + Sleep-Time Consolidation
// "함께 발화하는 뉴런은 함께 강화된다" (Neurons that fire together, wire together)
//
// PROVIDES: hebbianTrack, sleepConsolidate
// DEPENDS ON: neuron_crud.go (fireNeuron calls this), transcript.go (idle loop)
//
// 설계철학: Folder=Neuron, File=Trace 공리를 준수.
// 새 파일 타입 없이 hippocampus/session_log/co-activation.jsonl에 JSONL 기록.
// evolve 프롬프트가 이 데이터를 참조하여 뉴런 간 연관성을 학습한다.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

// sleepConsolidate — Sleep-Time Memory Consolidation
// Idle 사이클에서 호출. co-activation.jsonl을 분석하여
// 3회 이상 동시 발화된 뉴런 쌍에 대해 axon 파일을 자동 생성.
// 분석 후 co-activation.jsonl을 rotate (최근 100줄만 보존).
func sleepConsolidate(brainRoot string) {
	logPath := filepath.Join(brainRoot, "hippocampus", "session_log", hebbianLogFile)

	f, err := os.Open(logPath)
	if err != nil {
		return // co-activation 데이터 없음
	}
	defer f.Close()

	// 쌍별 카운트 집계
	type pair struct{ a, b string }
	counts := make(map[pair]int)

	scanner := bufio.NewScanner(f)
	lineCount := 0
	for scanner.Scan() {
		lineCount++
		var ca coActivation
		if json.Unmarshal(scanner.Bytes(), &ca) == nil && ca.A != "" && ca.B != "" {
			// 정규화: 알파벳 순으로 정렬하여 (A,B) == (B,A) 동일 처리
			a, b := ca.A, ca.B
			if a > b {
				a, b = b, a
			}
			counts[pair{a, b}]++
		}
	}

	if lineCount == 0 {
		return
	}

	// 3회 이상 동시 발화 → axon 생성
	created := 0
	for p, count := range counts {
		if count < 3 {
			continue
		}

		// A 뉴런 폴더에 axon 생성 (B를 타겟으로)
		aDir := filepath.Join(brainRoot, strings.ReplaceAll(p.a, "/", string(filepath.Separator)))
		bDir := filepath.Join(brainRoot, strings.ReplaceAll(p.b, "/", string(filepath.Separator)))

		// 양 폴더 모두 존재해야 함
		if _, err := os.Stat(aDir); os.IsNotExist(err) {
			continue
		}
		if _, err := os.Stat(bDir); os.IsNotExist(err) {
			continue
		}

		// axon 파일명: hebbian_{leaf}.axon
		bLeaf := filepath.Base(bDir)
		axonFile := filepath.Join(aDir, fmt.Sprintf("hebbian_%s.axon", bLeaf))
		if _, err := os.Stat(axonFile); err == nil {
			continue // 이미 존재
		}

		content := fmt.Sprintf("target: %s\nsource: hebbian_coactivation\ncount: %d\ncreated: %s\n",
			p.b, count, time.Now().Format("2006-01-02"))
		os.WriteFile(axonFile, []byte(content), 0600)
		fmt.Printf("[SLEEP] 🔗 Hebbian axon: %s → %s (co-fired %dx)\n", p.a, p.b, count)
		created++
	}

	// Rotate: 최근 100줄만 보존
	if lineCount > 100 {
		if data, err := os.ReadFile(logPath); err == nil {
			lines := strings.Split(strings.TrimSpace(string(data)), "\n")
			keep := lines[len(lines)-100:]
			os.WriteFile(logPath, []byte(strings.Join(keep, "\n")+"\n"), 0600)
			fmt.Printf("[SLEEP] 🔄 co-activation rotated: %d → 100 lines\n", lineCount)
		}
	}

	if created > 0 {
		logEpisode(brainRoot, "SLEEP_CONSOLIDATION", fmt.Sprintf("%d hebbian axons created", created))
		markBrainDirty()
	}
}
