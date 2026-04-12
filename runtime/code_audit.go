package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
)

// ============================================================================
// Module: Code Audit — Phase 52 & Phase 62 (Autonomous Evolution)
// Computes metrics: Go files, lines, funcs, test files using Goroutines + sync/atomic
// ============================================================================

// AuditResult holds basic code audit metrics
type AuditResult struct {
	GoFiles    int
	TestFiles  int
	TotalLines int
	FuncCount  int
}

// RunCodeAudit scans rootDir concurrently resolving the I/O bottleneck
func RunCodeAudit(rootDir string) AuditResult {
	var goFiles, testFiles, totalLines, funcCount int64
	var wg sync.WaitGroup

	// Semaphore to prevent "too many open files" panic on massive repositories
	sem := make(chan struct{}, 100)

	_ = filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(d.Name(), ".go") {
			return nil
		}

		atomic.AddInt64(&goFiles, 1)
		if strings.HasSuffix(d.Name(), "_test.go") {
			atomic.AddInt64(&testFiles, 1)
		}

		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			sem <- struct{}{}        // Acquire token
			defer func() { <-sem }() // Release token

			file, err := os.Open(p)
			if err != nil {
				return
			}
			defer file.Close()

			var localLines, localFuncs int64
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := scanner.Text()
				localLines++
				trimmed := strings.TrimSpace(line)
				if strings.HasPrefix(trimmed, "func ") {
					localFuncs++
				}
			}

			atomic.AddInt64(&totalLines, localLines)
			atomic.AddInt64(&funcCount, localFuncs)
		}(path)

		return nil
	})

	wg.Wait() // Wait for all goroutines to finish

	// Terminal reporting (matches original structure)
	fmt.Printf("[코드감사] %s\n", rootDir)
	fmt.Printf("  Go 파일: %d개\n", goFiles)
	fmt.Printf("  테스트 파일: %d개\n", testFiles)
	fmt.Printf("  총 라인: %d줄\n", totalLines)
	fmt.Printf("  함수 수: %d개\n", funcCount)

	return AuditResult{
		GoFiles:    int(goFiles),
		TestFiles:  int(testFiles),
		TotalLines: int(totalLines),
		FuncCount:  int(funcCount),
	}
}
