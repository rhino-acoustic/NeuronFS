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

// runTranscriptCategorizer runs every hour as a cron cycle:
// 1. git snapshot (선행)
// 2. 전사 백업 (24시간 경과분 이관)
// 3. 정황 수집 (세션, git push, 빌드 상태)
// 4. Gemini CLI 위임 카테고리 분류
// via Gemini CLI. Stacked on top of existing transcript system.
func runTranscriptCategorizer(brainRoot string) {
	// 초기 대기: 서버 안정화 후 시작
	time.Sleep(5 * time.Minute)

	lastNeuronCleanup := time.Now().Add(-7 * time.Hour) // 첫 사이클에서 즉시 실행

	for {
		nfsRoot := filepath.Dir(brainRoot)

		// 자율주행 비활성 체크
		if fileExists(filepath.Join(nfsRoot, "telegram-bridge", ".auto_evolve_disabled")) {
			time.Sleep(1 * time.Hour)
			continue
		}

		svLog("[CRON] ⏰ 전사 크론 사이클 시작")

		// ── Step 1: git 선행 ──
		cronGitSnapshot(nfsRoot)

		// ── Step 2: 전사 백업 (24시간 경과분) ──
		archiveOldTranscripts(brainRoot)

		// ── Step 2.5: 라이브 전사 50건 초과 시 자동 정리 ──
		pruneExcessTranscripts(brainRoot, 50)

		// ── Step 3: 정황 수집 ──
		ctx := collectCronContext(brainRoot, nfsRoot)
		svLog(fmt.Sprintf("[CRON] 📋 정황: %s", ctx))

		// ── Step 4: 카테고리 분류 (전사 있을 때만 Gemini CLI 위임) ──
		recentCount := countRecentTranscripts(brainRoot, 1)
		if recentCount > 0 {
			categorizeRecentTranscripts(brainRoot, 1)
		} else {
			svLog("[CRON] 💤 최근 전사 없음 — CLI 스킵")
		}

		// ── Step 5: 정황 기록 ──
		writeCronLog(brainRoot, ctx)

		// ── Step 6: 뇌 외부 구축 + 프로세스 뉴런 검증 ──
		verifyBrainExternals(brainRoot, nfsRoot)

		// ── Step 7: 뉴런 발화 정리 (6시간 간격) ──
		if time.Since(lastNeuronCleanup) >= 6*time.Hour {
			svLog("[CRON] 🧹 뉴런 정리 시작")
			runDecay(brainRoot, 7)
			deduplicateNeurons(brainRoot)
			lastNeuronCleanup = time.Now()
		}

		// ── Step 8: 자동 헬스 체크 (Verification-on-Resume 패턴) ──
		runHealthReport(brainRoot, nfsRoot)

		// ── Step 9: 자가 수리 파이프라인 (repair_proposals 스캔) ──
		AttemptSelfRepair(brainRoot)

		svLog("[CRON] ✅ 전사 크론 사이클 완료")
		time.Sleep(1 * time.Hour)
	}
}

// cronGitSnapshot performs a quick git add+commit before transcript processing
func cronGitSnapshot(nfsRoot string) {
	if err := SafeExecDir(ExecTimeoutGit, nfsRoot, "git", "add", "-A"); err != nil {
		return
	}
	msg := fmt.Sprintf("[cron] hourly snapshot %s", time.Now().Format("01-02 15:04"))
	SafeExecDir(ExecTimeoutGit, nfsRoot, "git", "commit", "-m", msg, "--no-verify", "--allow-empty-message")
	svLog("[CRON] 📸 git snapshot 완료")
}

// collectCronContext gathers system state for transcript enrichment
func collectCronContext(brainRoot, nfsRoot string) string {
	var parts []string

	// 1. 활성 세션 수 (CDP 타겟)
	targets, err := cdpListTargets(CDPPort)
	if err == nil {
		pages := 0
		for _, t := range targets {
			if strings.Contains(t.URL, "workbench") {
				pages++
			}
		}
		parts = append(parts, fmt.Sprintf("세션:%d", pages))
	}

	// 2. git push 여부
	out, err := SafeOutputDir(ExecTimeoutQuery, nfsRoot, "git", "log", "--oneline", "origin/main..HEAD")
	if err == nil {
		unpushed := strings.Count(strings.TrimSpace(string(out)), "\n") + 1
		if len(strings.TrimSpace(string(out))) == 0 {
			unpushed = 0
		}
		parts = append(parts, fmt.Sprintf("미push:%d", unpushed))
	}

	// 3. 빌드 상태 (neuronfs.exe 최종 빌드 시각)
	exe := filepath.Join(nfsRoot, "neuronfs.exe")
	if info, err := os.Stat(exe); err == nil {
		age := time.Since(info.ModTime())
		parts = append(parts, fmt.Sprintf("빌드:%s전", age.Round(time.Minute)))
	}

	// 4. 전사 파일 수
	transcriptsDir := filepath.Join(brainRoot, "_transcripts")
	entries, _ := os.ReadDir(transcriptsDir)
	txtCount := 0
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".txt") {
			txtCount++
		}
	}
	parts = append(parts, fmt.Sprintf("전사:%d건", txtCount))

	// 5. corrections 상태
	corrPath := filepath.Join(brainRoot, "_inbox", "corrections.jsonl")
	if info, err := os.Stat(corrPath); err == nil {
		parts = append(parts, fmt.Sprintf("교정:%dKB", info.Size()/1024))
	}

	return strings.Join(parts, " | ")
}

// writeCronLog appends a line to the cron log for audit trail
func writeCronLog(brainRoot, context string) {
	logPath := filepath.Join(brainRoot, "_inbox", "cron.log")
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return
	}
	defer f.Close()
	f.WriteString(fmt.Sprintf("[%s] %s\n", time.Now().Format("2006-01-02 15:04:05"), context))
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

		// GEMINI.md 규칙 간섭 방지: NeuronFS 외부에서 CLI 실행
		cleanDir := filepath.Join(os.TempDir(), "neuronfs_categorize")
		os.MkdirAll(cleanDir, 0750)
		result := executeGeminiCLI(AgentTask{
			Name:    "transcript_categorize",
			Prompt:  prompt,
			WorkDir: cleanDir,
		})

		if !result.Success || len(result.Output) < 10 {
			svLog(fmt.Sprintf("[CATEGORIZE] ⚠️ %s 분류 실패", fname))
			continue
		}

		// 4. JSON 파싱
		catResult := parseCategoryResult(result.Output)
		if catResult.Category == "" {
			svLog(fmt.Sprintf("[CATEGORIZE] ⚠️ %s JSON 파싱 실패", fname))
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
		svLog(fmt.Sprintf("[CATEGORIZE] ✅ %s → %s/%s", fname, catResult.Category, catResult.Title))
	}

		svLog(fmt.Sprintf("[CATEGORIZE] 📊 %d건 전사 분류 완료", categorized))
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
	// 1차: 정상 JSON 시도
	if json.Unmarshal([]byte(text), &result) == nil && result.Category != "" {
		return result
	}
	// 2차: relaxed — 따옴표 없는 키/값 보정 (LLM이 {key:value} 형태로 출력 시)
	fixed := fixRelaxedJSON(text)
	json.Unmarshal([]byte(fixed), &result)
	return result
}

// fixRelaxedJSON은 {key:"value"} 또는 {key:value} 를 {"key":"value"}로 보정
func fixRelaxedJSON(s string) string {
	// 키에 따옴표 추가: category: → "category":
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", "")
	// bare key 패턴: {key: or ,key:
	var b strings.Builder
	inStr := false
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if ch == '"' && (i == 0 || s[i-1] != '\\') {
			inStr = !inStr
		}
		b.WriteByte(ch)
	}
	raw := b.String()

	// 간단한 regex 대체: `키:` → `"키":`  (JSON 키 보정)
	import_free := raw
	for _, key := range []string{"category", "title", "summary"} {
		// {category: → {"category":
		import_free = strings.ReplaceAll(import_free, key+":", "\""+key+"\":")
		import_free = strings.ReplaceAll(import_free, key+" :", "\""+key+"\" :")
	}
	// 값이 따옴표 없는 경우: :"value" 는 OK, :value 는 "value"로
	// 이미 키를 보정했으니 JSON이 될 확률이 높아짐
	return import_free
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

// ──────────────────────────────────────────────────────────
// T1: 전사 50건 초과 시 자동 정리 (적층)
// ──────────────────────────────────────────────────────────

// pruneExcessTranscripts archives oldest transcripts when count exceeds maxLive
func pruneExcessTranscripts(brainRoot string, maxLive int) int {
	transcriptsDir := filepath.Join(brainRoot, "_transcripts")
	entries, err := os.ReadDir(transcriptsDir)
	if err != nil {
		return 0
	}

	// txt 파일만 수집
	var txtFiles []os.DirEntry
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".txt") {
			txtFiles = append(txtFiles, e)
		}
	}

	if len(txtFiles) <= maxLive {
		return 0
	}

	// 오래된 순으로 정렬 (ModTime 기준)
	type fileAge struct {
		entry os.DirEntry
		mtime time.Time
	}
	var sorted []fileAge
	for _, e := range txtFiles {
		info, err := e.Info()
		if err != nil {
			continue
		}
		sorted = append(sorted, fileAge{e, info.ModTime()})
	}
	// 직접 정렬 (sort 패키지 불사용)
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i].mtime.After(sorted[j].mtime) {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	// 초과분을 백업으로 이동
	excess := len(sorted) - maxLive
	moved := 0
	for i := 0; i < excess; i++ {
		name := sorted[i].entry.Name()
		info, _ := sorted[i].entry.Info()

		dateStr := info.ModTime().Format("20060102")
		parts := strings.Split(name, "_")
		for _, p := range parts {
			if len(p) == 10 && strings.Count(p, "-") == 2 {
				dateStr = strings.ReplaceAll(p, "-", "")
				break
			}
		}

		backupDir := filepath.Join(transcriptsDir, "_backup_"+dateStr)
		os.MkdirAll(backupDir, 0750)

		src := filepath.Join(transcriptsDir, name)
		dst := filepath.Join(backupDir, name)

		if fileExists(dst) {
			os.Remove(src) // 이미 백업에 있으면 원본만 제거
			moved++
			continue
		}

		data, err := os.ReadFile(src)
		if err != nil {
			continue
		}
		if err := os.WriteFile(dst, data, 0600); err != nil {
			continue
		}
		os.Remove(src)
		moved++
	}

		svLog(fmt.Sprintf("[PRUNE] 🗑️ 라이브 전사 %d건 → 백업 이관 (잔여 %d건)", moved, len(sorted)-moved))
	return moved
}

// ──────────────────────────────────────────────────────────
// T8: 뇌 외부 구축 + 프로세스 뉴런 검증 (적층)
// ──────────────────────────────────────────────────────────

// verifyBrainExternals checks that essential files outside brain_v4 exist
// and records the state as a neuron for audit trail.
func verifyBrainExternals(brainRoot, nfsRoot string) {
	type externalCheck struct {
		Name   string
		Path   string
		Exists bool
	}

	checks := []externalCheck{
		{"neuronfs.exe", filepath.Join(nfsRoot, "neuronfs.exe"), false},
		{"start.bat", filepath.Join(nfsRoot, "start.bat"), false},
		{"CLAUDE.md", filepath.Join(nfsRoot, "CLAUDE.md"), false},
		{"GEMINI.md", filepath.Join(os.Getenv("USERPROFILE"), ".gemini", "GEMINI.md"), false},
		{".cursorrules", filepath.Join(nfsRoot, ".cursorrules"), false},
		{"telegram-bridge", filepath.Join(nfsRoot, "telegram-bridge", "index.js"), false},
		{"corrections.jsonl", filepath.Join(brainRoot, "_inbox", "corrections.jsonl"), false},
	}

	missing := 0
	for i, c := range checks {
		checks[i].Exists = fileExists(c.Path)
		if !checks[i].Exists {
			missing++
		}
	}

	// 프로세스 뉴런에 기록
	progressDir := filepath.Join(brainRoot, "hippocampus", "전사분석", "_진행상황")
	os.MkdirAll(progressDir, 0750)

	var lines []string
	lines = append(lines, fmt.Sprintf("## 뇌 외부 구축 검증 [%s]", time.Now().Format("2006-01-02 15:04")))
	for _, c := range checks {
		status := "✅"
		if !c.Exists {
			status = "🔴"
		}
		lines = append(lines, fmt.Sprintf("- %s %s", status, c.Name))
	}
	if missing > 0 {
		lines = append(lines, fmt.Sprintf("\n⚠️ %d건 누락", missing))
	}

	os.WriteFile(filepath.Join(progressDir, "externals.md"), []byte(strings.Join(lines, "\n")), 0600)

	if missing > 0 {
		svLog(fmt.Sprintf("[VERIFY] ⚠️ 뇌 외부 %d건 누락", missing))
	} else {
		svLog("[VERIFY] ✅ 뇌 외부 구축 완전")
	}
}

// countRecentTranscripts counts transcript files modified within the last N hours
func countRecentTranscripts(brainRoot string, hoursBack int) int {
	transcriptsDir := filepath.Join(brainRoot, "_transcripts")
	cutoff := time.Now().Add(-time.Duration(hoursBack) * time.Hour)
	count := 0

	entries, err := os.ReadDir(transcriptsDir)
	if err != nil {
		return 0
	}

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".txt") {
			continue
		}
		info, err := e.Info()
		if err != nil || info.ModTime().Before(cutoff) || info.Size() < 100 {
			continue
		}
		count++
	}
	return count
}
