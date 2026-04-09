package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func pathExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func main() {
	brainRoot := os.Args[1]

	regionOrder := []string{"brainstem", "limbic", "hippocampus", "sensors", "cortex", "ego", "prefrontal"}

	total := 0
	pass := 0
	fail := 0
	var failures []string

	// ── Check 1: Essential regions exist ──
	for _, r := range regionOrder {
		total++
		p := filepath.Join(brainRoot, r)
		if !pathExists(p) {
			fail++
			failures = append(failures, fmt.Sprintf("C1: 영역 누락: %s", r))
		} else {
			pass++
		}
	}

	// ── Check 2: Brainstem 필수 뉴런 (必/禁 구조) ──
	bsPath := filepath.Join(brainRoot, "brainstem")
	for _, hanja := range []string{"必", "禁"} {
		total++
		hp := filepath.Join(bsPath, hanja)
		if !pathExists(hp) {
			fail++
			failures = append(failures, fmt.Sprintf("C2: brainstem 한자 폴더 누락: %s", hanja))
		} else {
			entries, _ := os.ReadDir(hp)
			childCount := 0
			for _, e := range entries {
				if e.IsDir() && !strings.HasPrefix(e.Name(), "_") {
					childCount++
				}
			}
			if childCount == 0 {
				fail++
				failures = append(failures, fmt.Sprintf("C2: brainstem/%s 하위 뉴런 0개", hanja))
			} else {
				pass++
			}
		}
	}

	// ── Check 3: _rules.md 존재 (최근 2시간 내 갱신) ──
	for _, r := range regionOrder {
		total++
		rulesPath := filepath.Join(brainRoot, r, "_rules.md")
		info, err := os.Stat(rulesPath)
		if err != nil {
			fail++
			failures = append(failures, fmt.Sprintf("C3: _rules.md 누락: %s", r))
		} else if time.Since(info.ModTime()) > 2*time.Hour {
			fail++
			failures = append(failures, fmt.Sprintf("C3: _rules.md 미갱신 (2h+): %s (마지막: %s)", r, info.ModTime().Format("15:04:05")))
		} else {
			pass++
		}
	}

	// ── Check 4: GEMINI.md 마커 무결성 ──
	total++
	home := os.Getenv("USERPROFILE")
	if home != "" {
		geminiPath := filepath.Join(home, ".gemini", "GEMINI.md")
		data, err := os.ReadFile(geminiPath)
		if err != nil {
			fail++
			failures = append(failures, "C4: GEMINI.md 없음")
		} else {
			content := string(data)
			if !strings.Contains(content, "<!-- NEURONFS:START -->") || !strings.Contains(content, "<!-- NEURONFS:END -->") {
				fail++
				failures = append(failures, "C4: GEMINI.md 마커 손상")
			} else {
				pass++
			}
		}
	}

	// ── Check 5: bomb 상태 확인 ──
	total++
	bombCount := 0
	for _, r := range regionOrder {
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
		fail++
		failures = append(failures, fmt.Sprintf("C5: 활성 bomb: %d개", bombCount))
	} else {
		pass++
	}

	// ── Check 6: GEMINI.md 인코딩 무결성 ──
	total++
	if home != "" {
		geminiPath := filepath.Join(home, ".gemini", "GEMINI.md")
		data, _ := os.ReadFile(geminiPath)
		content := string(data)
		if strings.Contains(content, "?쒓") || strings.Contains(content, "?딆") || strings.Contains(content, "TestEmitTarget") {
			fail++
			failures = append(failures, "C6: GEMINI.md 인코딩 깨짐 또는 테스트 오염 잔해")
		} else {
			pass++
		}
	}

	// ── Check 7: _rules.md 인코딩 무결성 (cortex) ──
	total++
	cortexRules := filepath.Join(brainRoot, "cortex", "_rules.md")
	if data, err := os.ReadFile(cortexRules); err == nil {
		content := string(data)
		if strings.Contains(content, "?딆") || strings.Contains(content, "?ㅼ") || strings.Contains(content, "梨쀫") {
			fail++
			failures = append(failures, "C7: cortex/_rules.md 인코딩 깨짐 (mojibake)")
		} else {
			pass++
		}
	}

	// ── Check 8: corrections.jsonl 존재 ──
	total++
	correctionsPath := filepath.Join(brainRoot, "_inbox", "corrections.jsonl")
	if !pathExists(correctionsPath) {
		fail++
		failures = append(failures, "C8: corrections.jsonl 없음")
	} else {
		pass++
	}

	// ── Report ──
	pct := float64(pass) / float64(total) * 100
	fmt.Println("━━━ NeuronFS Harness Report ━━━")
	fmt.Printf("통과: %d/%d (%.1f%%)\n\n", pass, total, pct)

	if len(failures) > 0 {
		fmt.Println("❌ 실패 항목:")
		for _, f := range failures {
			fmt.Printf("  - %s\n", f)
		}
	} else {
		fmt.Println("✅ 모든 검사 통과!")
	}

	if pct < 100 {
		os.Exit(1)
	}
}
