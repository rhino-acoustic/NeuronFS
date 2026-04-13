// codemap_gen.go — 소스 코드에서 코드맵 뉴런 자동 생성
//
// PROVIDES: generateCodemap
// DEPENDS ON: brain.go (scanBrain), inject.go (markBrainDirty)
//
// Go 소스 파일의 PROVIDES/DEPENDS 헤더를 파싱하여
// brain_v4/cortex/dev/_codemap/{파일명}/ 뉴런을 자동 생성/갱신한다.
// 소스 파일이 변경되면 STALE이 감지되어 자동 갱신 트리거.

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// generateCodemap scans Go source files for PROVIDES/DEPENDS headers
// and creates/updates codemap neurons under cortex/dev/_codemap/
// Returns number of created/updated codemap entries.
func generateCodemap(brainRoot string) int {
	projectRoot := filepath.Dir(brainRoot)
	runtimeDir := filepath.Join(projectRoot, "runtime")

	if _, err := os.Stat(runtimeDir); os.IsNotExist(err) {
		return 0
	}

	codemapRoot := filepath.Join(brainRoot, "cortex", "dev", "_codemap")
	os.MkdirAll(codemapRoot, 0750)

	updated := 0
	entries, _ := os.ReadDir(runtimeDir)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}
		// Skip test files
		if strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}

		srcPath := filepath.Join(runtimeDir, entry.Name())
		srcInfo, err := os.Stat(srcPath)
		if err != nil {
			continue
		}

		// Parse PROVIDES/DEPENDS from file header
		content, err := os.ReadFile(srcPath)
		if err != nil {
			continue
		}

		provides, depends := parseCodeHeaders(string(content))
		if provides == "" && depends == "" {
			continue // No structured header → skip
		}

		// Create codemap neuron directory
		baseName := strings.TrimSuffix(entry.Name(), ".go")
		neuronDir := filepath.Join(codemapRoot, baseName)
		os.MkdirAll(neuronDir, 0750)

		// Check if neuron exists and is fresh
		neuronFile := filepath.Join(neuronDir, "1.neuron")
		needsUpdate := true
		if neuronInfo, err := os.Stat(neuronFile); err == nil {
			// Source older than neuron → already fresh
			if srcInfo.ModTime().Before(neuronInfo.ModTime()) {
				needsUpdate = false
			}
		}

		if !needsUpdate {
			continue
		}

		// Generate neuron content
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("source: %s\n", srcPath))
		sb.WriteString(fmt.Sprintf("updated: %s\n", time.Now().Format("2006-01-02T15:04:05")))
		sb.WriteString(fmt.Sprintf("# %s\n\n", entry.Name()))
		if provides != "" {
			sb.WriteString(fmt.Sprintf("## PROVIDES\n%s\n\n", provides))
		}
		if depends != "" {
			sb.WriteString(fmt.Sprintf("## DEPENDS ON\n%s\n\n", depends))
		}

		os.WriteFile(neuronFile, []byte(sb.String()), 0600)
		updated++
	}

	if updated > 0 {
		fmt.Printf("[CODEMAP] 🗺️  %d codemap neurons updated\n", updated)
	}
	return updated
}

// parseCodeHeaders extracts PROVIDES and DEPENDS lines from Go file header comments
func parseCodeHeaders(content string) (provides, depends string) {
	lines := strings.Split(content, "\n")
	var provLines, depLines []string
	inProvides := false
	inDepends := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Stop at package declaration
		if strings.HasPrefix(line, "package ") {
			break
		}
		// Strip comment prefix
		if strings.HasPrefix(line, "//") {
			line = strings.TrimSpace(line[2:])
		} else {
			continue
		}

		if strings.HasPrefix(line, "PROVIDES:") {
			inProvides = true
			inDepends = false
			rest := strings.TrimSpace(strings.TrimPrefix(line, "PROVIDES:"))
			if rest != "" {
				provLines = append(provLines, rest)
			}
			continue
		}
		if strings.HasPrefix(line, "DEPENDS ON:") || strings.HasPrefix(line, "DEPENDS:") {
			inDepends = true
			inProvides = false
			rest := strings.TrimSpace(strings.TrimPrefix(line, "DEPENDS ON:"))
			rest = strings.TrimSpace(strings.TrimPrefix(rest, "DEPENDS:"))
			if rest != "" {
				depLines = append(depLines, rest)
			}
			continue
		}

		// Continuation lines (indented under PROVIDES/DEPENDS)
		if inProvides && line != "" && !strings.HasPrefix(line, "CALL") && !strings.HasPrefix(line, "CONSUMED") {
			provLines = append(provLines, line)
		} else if inDepends && line != "" && !strings.HasPrefix(line, "CALL") && !strings.HasPrefix(line, "CONSUMED") {
			depLines = append(depLines, line)
		} else {
			inProvides = false
			inDepends = false
		}
	}

	return strings.Join(provLines, "\n"), strings.Join(depLines, "\n")
}
