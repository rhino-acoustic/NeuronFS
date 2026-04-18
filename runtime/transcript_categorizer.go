package main

// 전사 카테고리 자동 분류 시스템 (적층)
// PROVIDES: runTranscriptCategorizer, categorizeRecentTranscripts
// DEPENDS ON: multi_agent.go (executeGeminiCLI), hijack_orchestrator.go (hlBuildContextualPrompt)

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// 기존 카테고리 목록 (SSOT: hippocampus/전사분석/ 디렉토리 기반)
var transcriptCategories = []string{
	"01_Autopoietic_자가진화",
	"02_Architecture_아키텍처",
	"03_CDP_하이브리드전사",
	"04_Dashboard_대시보드",
	"05_Governance_거버넌스",
	"06_Troubleshooting_디버그",
	"06_UI_UX_디자인",
	"07_Storage_저장소",
	"08_MultiAgent_멀티에이전트",
	"09_Verification_검증",
	"10_Benchmark_벤치마크",
	"11_Agent_에이전트",
	"11_Documentation_문서화",
	"12_Business_비즈니스",
}

// runTranscriptCategorizer runs every hour, categorizing recent transcripts
// via Gemini CLI. Stacked on top of existing transcript system.
func runTranscriptCategorizer(brainRoot string) {
	// 초기 대기: 서버 안정화 후 시작
	time.Sleep(5 * time.Minute)

	for {
		nfsRoot := filepath.Dir(brainRoot)

		// 자율주행 비활성 체크
		if fileExists(filepath.Join(nfsRoot, "telegram-bridge", ".auto_evolve_disabled")) {
			time.Sleep(1 * time.Hour)
			continue
		}

		categorizeRecentTranscripts(brainRoot, 1) // 최근 1시간
		time.Sleep(1 * time.Hour)
	}
}

// categorizeRecentTranscripts categorizes transcript files from the last N hours
// using Gemini CLI for content analysis and classification.
func categorizeRecentTranscripts(brainRoot string, hoursBack int) int {
	transcriptsDir := filepath.Join(brainRoot, "_transcripts")
	analysisDir := filepath.Join(brainRoot, "hippocampus", "전사분석")
	cutoff := time.Now().Add(-time.Duration(hoursBack) * time.Hour)

	// 1. 최근 N시간 전사 파일 수집
	entries, err := os.ReadDir(transcriptsDir)
	if err != nil {
		return 0
	}

	var recentFiles []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".txt") {
			continue
		}
		info, err := e.Info()
		if err != nil || info.ModTime().Before(cutoff) || info.Size() < 100 {
			continue
		}
		recentFiles = append(recentFiles, e.Name())
	}

	if len(recentFiles) == 0 {
		return 0
	}

	// 2. 각 전사 파일에서 첫 2000자 추출 + 카테고리 분류 요청
	categorized := 0
	for _, fname := range recentFiles {
		fpath := filepath.Join(transcriptsDir, fname)
		data, err := os.ReadFile(fpath)
		if err != nil {
			continue
		}

		// 이미 분류된 파일인지 체크 (.categorized.json 마커)
		markerPath := filepath.Join(transcriptsDir, ".categorized.json")
		markers := loadCategoryMarkers(markerPath)
		if _, done := markers[fname]; done {
			continue
		}

		// 첫 3000자만 추출 (토큰 절약)
		content := string(data)
		if len([]rune(content)) > 3000 {
			content = string([]rune(content)[:3000])
		}

		// 3. 카테고리 목록과 함께 Gemini CLI에 분류 요청
		catList := strings.Join(transcriptCategories, "\n")
		prompt := fmt.Sprintf(`다음 전사 파일의 내용을 분석하고 카테고리를 분류해라.

[기존 카테고리]
%s

[전사 파일: %s]
%s

[지시]
1. 위 카테고리 중 가장 적합한 것을 선택해라. 없으면 새 카테고리를 제안 (형식: NN_영문_한글)
2. 이 대화의 핵심 주제를 한 줄로 요약해라 (한글 40자 이내)
3. 반드시 JSON으로 응답: {"category":"카테고리명","title":"핵심주제","summary":"3줄요약"}
4. JSON만 출력. 설명 불필요.`, catList, fname, content)

		nfsRoot := filepath.Dir(brainRoot)
		result := executeGeminiCLI(AgentTask{
			Name:    "transcript_categorize",
			Prompt:  prompt,
			WorkDir: nfsRoot,
		})

		if !result.Success || len(result.Output) < 10 {
			fmt.Printf("[CATEGORIZE] ⚠️ %s 분류 실패\n", fname)
			continue
		}

		// 4. JSON 파싱
		catResult := parseCategoryResult(result.Output)
		if catResult.Category == "" {
			fmt.Printf("[CATEGORIZE] ⚠️ %s JSON 파싱 실패\n", fname)
			continue
		}

		// 5. 전사분석 뉴런 생성
		// 날짜+제목으로 디렉토리 생성
		dateStr := extractDateFromFilename(fname)
		neuronDir := filepath.Join(analysisDir, catResult.Category, dateStr+"_"+sanitizeDirName(catResult.Title))
		os.MkdirAll(neuronDir, 0750)

		ruleContent := fmt.Sprintf("---\ntype: %s\ndate: %s\nsource: %s\n---\n%s\n",
			strings.SplitN(catResult.Category, "_", 3)[1],
			dateStr,
			fname,
			catResult.Summary)
		os.WriteFile(filepath.Join(neuronDir, "rule.md"), []byte(ruleContent), 0600)

		// 마커 기록
		markers[fname] = catResult.Category
		saveCategoryMarkers(markerPath, markers)

		categorized++
		fmt.Printf("[CATEGORIZE] ✅ %s → %s/%s\n", fname, catResult.Category, catResult.Title)
	}

	if categorized > 0 {
		fmt.Printf("[CATEGORIZE] 📊 %d건 전사 분류 완료\n", categorized)
	}
	return categorized
}

type categoryResult struct {
	Category string `json:"category"`
	Title    string `json:"title"`
	Summary  string `json:"summary"`
}

func parseCategoryResult(output string) categoryResult {
	var result categoryResult
	// JSON 블록 추출 (```json ... ``` 또는 { ... })
	text := output
	if idx := strings.Index(text, "{"); idx >= 0 {
		text = text[idx:]
		if end := strings.LastIndex(text, "}"); end >= 0 {
			text = text[:end+1]
		}
	}
	json.Unmarshal([]byte(text), &result)
	return result
}

func extractDateFromFilename(fname string) string {
	// NeuronFS_2026-04-18_08h.txt → 4월18일
	parts := strings.Split(fname, "_")
	for _, p := range parts {
		if len(p) == 10 && strings.Count(p, "-") == 2 {
			dateParts := strings.Split(p, "-")
			if len(dateParts) == 3 {
				month := strings.TrimLeft(dateParts[1], "0")
				day := strings.TrimLeft(dateParts[2], "0")
				return month + "월" + day + "일"
			}
		}
	}
	return time.Now().Format("1월2일")
}

func sanitizeDirName(name string) string {
	// 디렉토리명으로 사용 불가한 문자 제거
	replacer := strings.NewReplacer(
		"/", "_", "\\", "_", ":", "_", "*", "_",
		"?", "_", "\"", "_", "<", "_", ">", "_", "|", "_",
		"\n", "_", "\r", "",
	)
	result := replacer.Replace(name)
	if len([]rune(result)) > 40 {
		result = string([]rune(result)[:40])
	}
	return result
}

func loadCategoryMarkers(path string) map[string]string {
	markers := make(map[string]string)
	data, err := os.ReadFile(path)
	if err != nil {
		return markers
	}
	json.Unmarshal(data, &markers)
	return markers
}

func saveCategoryMarkers(path string, markers map[string]string) {
	data, _ := json.MarshalIndent(markers, "", "  ")
	os.WriteFile(path, data, 0600)
}
