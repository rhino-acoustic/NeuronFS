package main

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// RunHarness checks the structural integrity of the NeuronFS brain
// and logs failures into the agent inbox.
// Uses fileExists from utils.go (single source of truth for path checking).
func RunHarness(brainRoot string, logger func(string)) {
	if logger != nil {
		logger("🔍 harness 실행 (Go native)")
	}

	var fails []string
	var passes int

	// ── Check 1: Essential regions exist ──
	essentialRegions := RegionOrder
	for _, r := range essentialRegions {
		p := filepath.Join(brainRoot, r)
		if !fileExists(p) {
			fails = append(fails, fmt.Sprintf("영역 누락: %s", r))
		} else {
			passes++
		}
	}

	// ── Check 2: Brainstem 필수 뉴런 (必/禁 구조) ──
	bsPath := filepath.Join(brainRoot, "brainstem")
	for _, hanja := range []string{"必", "禁"} {
		hp := filepath.Join(bsPath, hanja)
		if !fileExists(hp) {
			fails = append(fails, fmt.Sprintf("brainstem 한자 폴더 누락: %s", hanja))
		} else {
			entries, _ := os.ReadDir(hp)
			childCount := 0
			for _, e := range entries {
				if !e.IsDir() && strings.HasSuffix(e.Name(), ".neuron") {
					childCount++
				}
			}
			if childCount == 0 {
				fails = append(fails, fmt.Sprintf("brainstem/%s 하위 뉴런 0개", hanja))
			} else {
				passes++
			}
		}
	}

	// ── Check 3: _rules.md 존재 (최근 1시간 내 갱신) ──
	for _, r := range essentialRegions {
		rulesPath := filepath.Join(brainRoot, r, "_rules.md")
		info, err := os.Stat(rulesPath)
		if err != nil {
			fails = append(fails, fmt.Sprintf("_rules.md 누락: %s", r))
		} else if time.Since(info.ModTime()) > 2*time.Hour {
			fails = append(fails, fmt.Sprintf("_rules.md 미갱신 (2h+): %s", r))
		} else {
			passes++
		}
	}

	// ── Check 4: GEMINI.md 마커 무결성 ──
	home := os.Getenv("USERPROFILE")
	if home != "" {
		geminiPath := filepath.Join(home, ".gemini", "GEMINI.md")
		data, err := os.ReadFile(geminiPath)
		if err != nil {
			fails = append(fails, "GEMINI.md 없음")
		} else {
			content := string(data)
			if !strings.Contains(content, "<!-- NEURONFS:START -->") || !strings.Contains(content, "<!-- NEURONFS:END -->") {
				fails = append(fails, "GEMINI.md 마커 손상")
			} else {
				passes++
			}
		}
	}

	// ── Check 5: bomb 상태 확인 ──
	bombCount := 0
	for _, r := range essentialRegions {
		filepath.Walk(filepath.Join(brainRoot, r), func(path string, info os.FileInfo, err error) error {
			if err != nil || info == nil {
				return nil
			}
			if info.Name() == "bomb.neuron" {
				bombCount++
			}
			return nil
		})
	}
	if bombCount > 0 {
		fails = append(fails, fmt.Sprintf("활성 bomb: %d개", bombCount))
	} else {
		passes++
	}

	// ── Check 6: 뉴런 인코딩 무결성 (mojibake 감지) ──
	mojibakeCount := 0
	for _, r := range essentialRegions {
		filepath.Walk(filepath.Join(brainRoot, r), func(path string, info os.FileInfo, err error) error {
			if err != nil || info == nil {
				return nil
			}
			if info.IsDir() {
				name := info.Name()
				if strings.HasPrefix(name, ".archive") || strings.HasPrefix(name, "_") || strings.Contains(name, "?") {
					return filepath.SkipDir
				}
				// 한자 옵코드 폴더(必/禁 등)는 통과, 깨진 1~3바이트 비ASCII 폴더명은 스킵
				if len(name) <= 3 && !strings.ContainsAny(name, RuneChars+"가-힣") {
					hasWeird := false
					for _, r := range name {
						if r > 127 && r < 0x3000 {
							hasWeird = true
							break
						}
					}
					if hasWeird {
						return filepath.SkipDir
					}
				}
				return nil
			}
			if !strings.HasSuffix(info.Name(), ".neuron") {
				return nil
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return nil
			}
			if isMojibake(string(data)) {
				mojibakeCount++
				if logger != nil {
					logger(fmt.Sprintf("    → mojibake: %s", path))
				}
			}
			return nil
		})
	}
	if mojibakeCount > 0 {
		fails = append(fails, fmt.Sprintf("인코딩 깨짐(mojibake): %d개 뉴런", mojibakeCount))
	} else {
		passes++
	}

	// ── Check 7: MCP 서버 health ──
	mcpHealthOK := false
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", MCPStreamPort), 2*time.Second)
	if err == nil {
		conn.Close()
		mcpHealthOK = true
	}
	if !mcpHealthOK {
		fails = append(fails, fmt.Sprintf("MCP 서버(%d) 응답 없음 — 좀비 또는 비활성", MCPStreamPort))
	} else {
		passes++
	}

	// ── Check 8: 코드맵 커버리지 (runtime .go vs cortex/dev/_codemap) ──
	codemapDir := filepath.Join(brainRoot, "cortex", "dev", "_codemap")
	var codemapNeurons int
	filepath.Walk(codemapDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && strings.HasSuffix(info.Name(), ".neuron") {
			codemapNeurons++
		}
		return nil
	})
	runtimeDir := filepath.Join(filepath.Dir(brainRoot), "runtime")
	var goFileCount int
	if entries, err := os.ReadDir(runtimeDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".go") && !strings.HasSuffix(e.Name(), "_test.go") {
				goFileCount++
			}
		}
	}
	if goFileCount > 0 {
		coverage := float64(codemapNeurons) / float64(goFileCount) * 100
		if coverage < 10 {
			fails = append(fails, fmt.Sprintf("코드맵 커버리지 부족: %.0f%% (%d/%d)", coverage, codemapNeurons, goFileCount))
		} else {
			passes++
		}
		if logger != nil {
			logger(fmt.Sprintf("  📊 코드맵: %d뉴런/%d파일 (%.0f%%)", codemapNeurons, goFileCount, coverage))
		}
	}

	// ── Result ──
	if len(fails) == 0 {
		if logger != nil {
			logger(fmt.Sprintf("✅ harness PASS (%d checks)", passes))
		}
	} else {
		if logger != nil {
			logger(fmt.Sprintf("⚠️ harness FAIL: %d/%d", len(fails), passes+len(fails)))
		}
		var report strings.Builder
		report.WriteString("# from: harness_engine\n# priority: urgent\n\n## 🔍 Harness 위반\n\n")
		for _, f := range fails {
			report.WriteString(fmt.Sprintf("- ❌ %s\n", f))
			if logger != nil {
				logger(fmt.Sprintf("  ❌ %s", f))
			}
		}
		d := filepath.Join(brainRoot, "_agents", "bot1", "inbox")
		os.MkdirAll(d, 0750)
		fname := filepath.Join(d, time.Now().Format("20060102_150405")+"_diag_harness.md")
		os.WriteFile(fname, []byte(report.String()), 0600)
	}
}
