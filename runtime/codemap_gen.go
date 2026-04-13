// codemap_gen.go — 소스 코드에서 코드맵 뉴런 자동 생성
//
// PROVIDES: generateCodemap, parseCodeHeaders, collectStaleCodemaps
// DEPENDS ON: emit_helpers.go (collectCodemapPaths — freshness 소비자)
//
// 워크스페이스 루트의 모든 Go 소스 파일에서 PROVIDES/DEPENDS 헤더를 파싱하여
// brain_v4/cortex/dev/_codemap/{파일명}/ 뉴런을 자동 생성/갱신한다.
// source: 필드로 소스 파일 mtime 추적 → 변경 시 STALE 감지 → inject 트리거.
// AI 불필요, 순수 Go 파싱.

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// generateCodemap scans all Go source files under the workspace root
// for PROVIDES/DEPENDS headers and creates/updates codemap neurons.
// Scans any directory containing .go files (not vendor, not .git, not test).
// Returns number of created/updated codemap entries.
func generateCodemap(brainRoot string) int {
	projectRoot := filepath.Dir(brainRoot)

	codemapRoot := filepath.Join(brainRoot, "cortex", "dev", "_codemap")
	os.MkdirAll(codemapRoot, 0750)

	updated := 0

	// Walk project root for all .go files (skip hidden dirs, vendor, node_modules, dist)
	filepath.WalkDir(projectRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			// Skip excluded directories
			if name == ".git" || name == "vendor" || name == "node_modules" ||
				name == "dist" || name == "brain_v4" || name == ".neuronfs_backup" ||
				strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			return nil
		}

		// Only .go files, skip tests
		if !strings.HasSuffix(d.Name(), ".go") || strings.HasSuffix(d.Name(), "_test.go") {
			return nil
		}

		srcInfo, statErr := os.Stat(path)
		if statErr != nil {
			return nil
		}

		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}

		provides, depends := parseCodeHeaders(string(content))
		if provides == "" && depends == "" {
			return nil // 헤더 없으면 스킵
		}

		// 뉴런 디렉토리명: 상위폴더_파일명 (충돌 방지)
		relPath, _ := filepath.Rel(projectRoot, path)
		dirPart := filepath.Dir(relPath)
		baseName := strings.TrimSuffix(d.Name(), ".go")

		// runtime/mcp_server.go → "mcp_server"
		// cmd/harness_check/main.go → "cmd_harness_check_main"
		neuronName := baseName
		if dirPart != "." && dirPart != "runtime" {
			// 비-runtime 디렉토리는 경로를 접두어로
			neuronName = strings.ReplaceAll(dirPart, string(filepath.Separator), "_") + "_" + baseName
		}

		neuronDir := filepath.Join(codemapRoot, neuronName)
		os.MkdirAll(neuronDir, 0750)

		// Freshness check: source mtime vs neuron mtime
		neuronFile := filepath.Join(neuronDir, "1.neuron")
		if neuronInfo, err := os.Stat(neuronFile); err == nil {
			if srcInfo.ModTime().Before(neuronInfo.ModTime()) {
				return nil // 소스가 뉴런보다 오래됨 → 이미 최신
			}
		}

		// Generate neuron content
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("source: %s\n", path))
		sb.WriteString(fmt.Sprintf("updated: %s\n", time.Now().Format("2006-01-02T15:04:05")))
		sb.WriteString(fmt.Sprintf("# %s\n", d.Name()))
		if dirPart != "." {
			sb.WriteString(fmt.Sprintf("dir: %s\n\n", dirPart))
		} else {
			sb.WriteString("\n")
		}
		if provides != "" {
			sb.WriteString(fmt.Sprintf("## PROVIDES\n%s\n\n", provides))
		}
		if depends != "" {
			sb.WriteString(fmt.Sprintf("## DEPENDS ON\n%s\n\n", depends))
		}

		os.WriteFile(neuronFile, []byte(sb.String()), 0600)
		updated++
		return nil
	})

	if updated > 0 {
		fmt.Printf("[CODEMAP] 🗺️  %d codemap neurons updated\n", updated)
	}
	return updated
}

// parseCodeHeaders extracts PROVIDES and DEPENDS lines from Go file header comments.
// Pure Go parsing — no AI required.
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

// collectStaleCodemaps scans all codemap neurons and returns
// entries where source file is newer than the neuron (STALE).
// Used by emitBootstrap to inject STALE warnings into GEMINI.md.
func collectStaleCodemaps(brainRoot string) []string {
	codemapRoot := filepath.Join(brainRoot, "cortex", "dev", "_codemap")
	if _, err := os.Stat(codemapRoot); os.IsNotExist(err) {
		return nil
	}

	var stale []string
	entries, _ := os.ReadDir(codemapRoot)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		neuronFile := filepath.Join(codemapRoot, entry.Name(), "1.neuron")
		neuronInfo, err := os.Stat(neuronFile)
		if err != nil {
			continue
		}

		content, err := os.ReadFile(neuronFile)
		if err != nil {
			continue
		}

		// Extract source: field
		for _, line := range strings.Split(string(content), "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "source:") {
				srcPath := strings.TrimSpace(strings.TrimPrefix(line, "source:"))
				srcInfo, err := os.Stat(srcPath)
				if err != nil {
					continue
				}
				if srcInfo.ModTime().After(neuronInfo.ModTime()) {
					stale = append(stale, fmt.Sprintf("`%s` → 뉴런: `%s`", filepath.Base(srcPath), entry.Name()))
				}
				break
			}
		}
	}
	return stale
}
