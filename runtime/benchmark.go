package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ============================================================================
// Module: Performance Benchmark — TF-IDF Similarity vs Folder Scan (V12-C)
// Measures real latency of NeuronFS operations for comparison with vector DBs.
// ============================================================================

// BenchmarkResult holds performance metrics
type BenchmarkResult struct {
	Operation   string
	NeuronCount int
	DurationMs  float64
	PerItemMs   float64
}

// RunBenchmarkSuite executes all benchmarks and returns results
func RunBenchmarkSuite(brainRoot string) []BenchmarkResult {
	var results []BenchmarkResult

	fmt.Println("[벤치마크] NeuronFS 성능 측정 시작")

	// 1. Full brain scan
	start := time.Now()
	brain := scanBrain(brainRoot)
	scanDuration := time.Since(start)
	totalNeurons := 0
	for _, r := range brain.Regions {
		totalNeurons += len(r.Neurons)
	}
	results = append(results, BenchmarkResult{
		Operation:   "전체 뇌 스캔 (scanBrain)",
		NeuronCount: totalNeurons,
		DurationMs:  float64(scanDuration.Microseconds()) / 1000.0,
		PerItemMs:   float64(scanDuration.Microseconds()) / 1000.0 / float64(max(totalNeurons, 1)),
	})

	// 2. Similarity index build
	start = time.Now()
	BuildSimilarityIndex(brainRoot)
	indexDuration := time.Since(start)
	results = append(results, BenchmarkResult{
		Operation:   "TF-IDF 인덱스 빌드",
		NeuronCount: totalNeurons,
		DurationMs:  float64(indexDuration.Microseconds()) / 1000.0,
		PerItemMs:   float64(indexDuration.Microseconds()) / 1000.0 / float64(max(totalNeurons, 1)),
	})

	// 3. Similarity query
	start = time.Now()
	queryResults := QuerySimilar("에러 패턴 반복 수정", 5)
	queryDuration := time.Since(start)
	results = append(results, BenchmarkResult{
		Operation:   fmt.Sprintf("유사도 쿼리 (top 5, 결과 %d건)", len(queryResults)),
		NeuronCount: totalNeurons,
		DurationMs:  float64(queryDuration.Microseconds()) / 1000.0,
		PerItemMs:   float64(queryDuration.Microseconds()) / 1000.0 / float64(max(totalNeurons, 1)),
	})

	// 4. File read latency (single neuron)
	var singleReadMs float64
	_ = filepath.Walk(brainRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || filepath.Ext(path) != ".neuron" {
			return nil
		}
		start := time.Now()
		_, _ = os.ReadFile(path)
		singleReadMs = float64(time.Since(start).Microseconds()) / 1000.0
		return filepath.SkipAll
	})
	results = append(results, BenchmarkResult{
		Operation:   "단일 뉴런 읽기",
		NeuronCount: 1,
		DurationMs:  singleReadMs,
		PerItemMs:   singleReadMs,
	})

	// 5. Emit pipeline
	start = time.Now()
	_ = emitRules(runSubsumption(brain))
	emitDuration := time.Since(start)
	results = append(results, BenchmarkResult{
		Operation:   "Emit 규칙 생성",
		NeuronCount: totalNeurons,
		DurationMs:  float64(emitDuration.Microseconds()) / 1000.0,
		PerItemMs:   float64(emitDuration.Microseconds()) / 1000.0 / float64(max(totalNeurons, 1)),
	})

	// Print results
	fmt.Println("\n=== NeuronFS 벤치마크 결과 ===")
	fmt.Printf("%-35s %8s %10s %10s\n", "작업", "뉴런수", "총시간(ms)", "건당(ms)")
	fmt.Println(strings.Repeat("-", 70))
	for _, r := range results {
		fmt.Printf("%-35s %8d %10.2f %10.4f\n", r.Operation, r.NeuronCount, r.DurationMs, r.PerItemMs)
	}

	// Save to brain
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# NeuronFS 벤치마크 결과 (%s)\n\n", time.Now().Format("2006-01-02 15:04")))
	sb.WriteString("| 작업 | 뉴런수 | 총시간(ms) | 건당(ms) |\n")
	sb.WriteString("|------|--------|-----------|--------|\n")
	for _, r := range results {
		sb.WriteString(fmt.Sprintf("| %s | %d | %.2f | %.4f |\n", r.Operation, r.NeuronCount, r.DurationMs, r.PerItemMs))
	}
	sb.WriteString("\n## 비교 (참고치)\n")
	sb.WriteString("- Pinecone 쿼리: 50-200ms (서버 왕복)\n")
	sb.WriteString("- ChromaDB 쿼리: 20-100ms (로컬)\n")
	sb.WriteString("- NeuronFS: 서버 없음, 비용 $0\n")

	resultPath := filepath.Join(brainRoot, "hippocampus", "benchmark_results.neuron")
	_ = os.MkdirAll(filepath.Dir(resultPath), 0755)
	_ = os.WriteFile(resultPath, []byte(sb.String()), 0644)

	RecordAudit(brainRoot, "benchmark", "complete", resultPath, fmt.Sprintf("%d neurons benchmarked", totalNeurons), true)
	return results
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
