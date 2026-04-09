package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func diagPathExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

// RunHarness checks the structural integrity of the NeuronFS brain
// and logs failures into the agent inbox.
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
		if !diagPathExists(p) {
			fails = append(fails, fmt.Sprintf("영역 누락: %s", r))
		} else {
			passes++
		}
	}

	// ── Check 2: Brainstem 필수 뉴런 (必/禁 구조) ──
	bsPath := filepath.Join(brainRoot, "brainstem")
	for _, hanja := range []string{"必", "禁"} {
		hp := filepath.Join(bsPath, hanja)
		if !diagPathExists(hp) {
			fails = append(fails, fmt.Sprintf("brainstem 한자 폴더 누락: %s", hanja))
		} else {
			entries, _ := os.ReadDir(hp)
			childCount := 0
			for _, e := range entries {
				if e.IsDir() && !strings.HasPrefix(e.Name(), "_") {
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
