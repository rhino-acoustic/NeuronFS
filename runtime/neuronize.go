// neuronize.go — Groq 기반 자동 부정형 뉴런 생성 + 극성 전환 (Polarity Shift)
//
// 두 가지 파이프라인:
//   1. --neuronize: corrections 로그/에피소드 → Groq → contra 뉴런 자동 생성
//   2. --polarize:  기존 긍정형 뉴런 스캔 → 부정형 전환 제안/실행
//
// Usage:
//   neuronfs <brain_path> --neuronize           — Groq 기반 contra 뉴런 자동 생성
//   neuronfs <brain_path> --neuronize --dry-run — 제안만 (실행 안 함)
//   neuronfs <brain_path> --polarize            — 긍정형→부정형 전환 실행
//   neuronfs <brain_path> --polarize --dry-run  — 전환 대상만 리스트

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// ─── Neuronize System Prompt (ENFP 프롬프트 엔지니어링 가이드 적용) ───

const neuronizeSystemPrompt = `당신은 NeuronFS 뇌의 '백혈구(자가면역 세포)'입니다. 사용자의 교정 로그와 에러 내역을 분석하여, 미래의 AI 에이전트들이 **같은 실수를 절대 반복하지 못하도록** 강력한 억제(Contra) 규칙을 만드십시오.

**[Rule Writing Guidelines]**
1. **파일명 (Filename):** 순수 한국어로 10자 이내 금지형 명사 작성 (예: 반복루프_금지, 절대경로_의존금지, 시뮬레이션_금지). 한자(禁/必/推) 사용 절대 금지.
2. **종결어미:** "~해야 합니다", "~하는 것이 좋습니다" 금지. "~~마라", "~~할 것", "~~금지" 등 군더더기 없는 명령조(Imperative) 사용.
3. **서문 금지:** "알겠습니다", "다음은 규칙입니다" 같은 응답 생성 절대 금지. 오직 JSON만 출력할 것.
4. **이유(Rationale):** 각 규칙의 첫 문장에 금지의 이유를 단 한 줄의 강력한 메타포로 서술할 것.

**[Output Format — JSON]**
{
  "contras": [
    {
      "name": "시뮬레이션_금지",
      "region": "cortex",
      "category": "quality",
      "rationale": "시뮬레이션은 뇌의 기억을 오염시키는 환각이다. 실제 실행 결과만 기억할 것.",
      "source_error": "빌드 결과를 시뮬레이션으로 통과 처리됨"
    }
  ]
}

오직 JSON만 출력하라. Markdown 금지. 서문 금지. 해설 금지.`

// ─── Polarize System Prompt ───

const polarizeSystemPrompt = `당신은 NeuronFS 뇌의 극성 전환(Polarity Shift) 엔진입니다. 긍정형 뉴런 목록을 받아, 각각을 부정/억제형(Contra)으로 전환하는 규칙을 생성합니다.

**[전환 원칙]**
- "use_X" → "禁X_의존" 또는 "X_남용금지"
- "always_Y" → "Y만_고집금지" 
- 영어 긍정형 → 한국어 부정형 (네이티브 한국어 사용)
- 전환 시 원래 뉴런의 의도를 왜곡하지 마라. 과잉 적용 방지용 억제 규칙을 만들어라.

**[Output Format — JSON]**
{
  "shifts": [
    {
      "original_path": "cortex/frontend/use_fast_routing",
      "new_name": "禁클라측_라우팅의존",
      "new_region": "cortex",
      "new_category": "frontend",
      "rationale": "클라이언트 사이드 라우팅은 뇌의 시냅스 응답성을 떨어트린다. 오직 서버 사이드/정적 라우팅만 허용한다."
    }
  ]
}

오직 JSON만 출력하라.`

// ─── runNeuronize: corrections/episodes → Groq → contra neurons ───

func runNeuronize(brainRoot string, dryRun bool) {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		fmt.Printf("%s[TRAUMA] GROQ_API_KEY not set — 면역 체계 비활성%s\n", ansiRed, ansiReset)
		fmt.Println("  Set: $env:GROQ_API_KEY = '<your-groq-api-key>'")
		return
	}

	fmt.Printf("%s[NEURON] Auto-Neuronize Engine Initializing...%s\n", ansiCyan, ansiReset)
	if dryRun {
		fmt.Printf("%s  ⚠️  DRY RUN — 제안만, 실행 안 함%s\n", ansiYellow, ansiReset)
	}

	// 1. Collect correction sources
	var corrections []string

	// Source A: corrections_history.jsonl (persistent, not cleared by --watch)
	historyPath := filepath.Join(brainRoot, "_inbox", "corrections_history.jsonl")
	if data, err := os.ReadFile(historyPath); err == nil && len(data) > 0 {
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				corrections = append(corrections, line)
			}
		}
	}
	// Fallback: corrections.jsonl (may be empty if --watch already processed)
	correctionsPath := filepath.Join(brainRoot, "_inbox", "corrections.jsonl")
	if data, err := os.ReadFile(correctionsPath); err == nil && len(data) > 0 {
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				corrections = append(corrections, line)
			}
		}
	}

	// Source B: hippocampus episode logs (errors/corrections only)
	episodes := collectEpisodes(brainRoot)
	errorKeywords := []string{"ERROR", "FAIL", "TRAUMA", "ROLLBACK", "BLOCKED", "SECURITY", "차단"}
	for _, ep := range episodes {
		for _, kw := range errorKeywords {
			if strings.Contains(strings.ToUpper(ep), kw) {
				corrections = append(corrections, ep)
				break
			}
		}
	}

	// Source C: agent inbox crash alerts
	inboxDir := filepath.Join(brainRoot, "_agents", "bot1", "inbox")
	if entries, err := os.ReadDir(inboxDir); err == nil {
		for _, e := range entries {
			if strings.Contains(e.Name(), "crash") || strings.Contains(e.Name(), "harness") {
				data, _ := os.ReadFile(filepath.Join(inboxDir, e.Name()))
				if len(data) > 0 {
					corrections = append(corrections, string(data))
				}
			}
		}
	}

	if len(corrections) == 0 {
		fmt.Printf("%s[NEURON] 교정 로그 없음 — 면역 반응 불필요%s\n", ansiGreen, ansiReset)
		return
	}

	fmt.Printf("  📝 교정 소스 수집: %d건\n", len(corrections))

	// 2. Build prompt for Groq
	var sb strings.Builder
	sb.WriteString("다음은 NeuronFS 시스템의 최근 교정 로그 및 에러 기록입니다.\n")
	sb.WriteString("이 로그들을 분석하여 '같은 실수를 방지하기 위한 contra(억제) 뉴런'을 생성하십시오.\n\n")
	sb.WriteString("## 교정 로그\n```\n")

	// Limit to last 30 entries to fit context
	start := 0
	if len(corrections) > 30 {
		start = len(corrections) - 30
	}
	for _, c := range corrections[start:] {
		if len(c) > 300 {
			c = c[:300] + "..."
		}
		sb.WriteString(c + "\n")
	}
	sb.WriteString("```\n\n")

	// Add current brain state for context (avoid duplicates)
	// 토큰 절약: 전체 뉴런 대신 카운터 상위 50개만 전송
	brain := scanBrain(brainRoot)
	type scoredNeuron struct {
		path    string
		counter int
	}
	var allNeurons []scoredNeuron
	for _, region := range brain.Regions {
		for _, n := range region.Neurons {
			allNeurons = append(allNeurons, scoredNeuron{
				path:    region.Name + "/" + n.Path,
				counter: n.Counter,
			})
		}
	}
	// 카운터 내림차순 정렬
	sort.Slice(allNeurons, func(i, j int) bool {
		return allNeurons[i].counter > allNeurons[j].counter
	})
	// 상위 50개만 전송
	limit := 50
	if len(allNeurons) < limit {
		limit = len(allNeurons)
	}
	sb.WriteString(fmt.Sprintf("## 기존 뉴런 TOP %d (중복 생성 금지, 총 %d개)\n", limit, len(allNeurons)))
	for _, n := range allNeurons[:limit] {
		sb.WriteString(fmt.Sprintf("- %s (c:%d)\n", n.path, n.counter))
	}
	sb.WriteString("\n위 기존 뉴런과 겹치지 않는 새로운 contra 규칙만 생성하라. 최대 5개.\n")

	// 3. Call Groq
	fmt.Printf("  🌐 Groq API 호출 중 (llama-3.3-70b-versatile)...\n")
	startTime := time.Now()

	groqReq := groqRequest{
		Model: "llama-3.3-70b-versatile",
		Messages: []groqMessage{
			{Role: "system", Content: neuronizeSystemPrompt},
			{Role: "user", Content: sb.String()},
		},
		Temperature:    0.3,
		MaxTokens:      2048,
		TopP:           0.9,
		Stream:         false,
		ResponseFormat: &groqResponseFormat{Type: "json_object"},
	}

	respBody, err := callGroqRaw(apiKey, groqReq)
	if err != nil {
		fmt.Printf("%s[TRAUMA] Groq API 실패: %v%s\n", ansiRed, err, ansiReset)
		return
	}

	elapsed := time.Since(startTime)
	fmt.Printf("  ✅ 응답 수신 (%s)\n\n", elapsed.Round(time.Millisecond))

	// 4. Parse response
	type contraEntry struct {
		Name        string `json:"name"`
		Region      string `json:"region"`
		Category    string `json:"category"`
		Rationale   string `json:"rationale"`
		SourceError string `json:"source_error"`
	}
	type neuronizeResult struct {
		Contras []contraEntry `json:"contras"`
	}

	var result neuronizeResult
	if err := json.Unmarshal([]byte(respBody), &result); err != nil {
		// Try to extract JSON
		if idx := strings.Index(respBody, "{"); idx >= 0 {
			cleaned := respBody[idx:]
			if lastIdx := strings.LastIndex(cleaned, "}"); lastIdx >= 0 {
				cleaned = cleaned[:lastIdx+1]
			}
			json.Unmarshal([]byte(cleaned), &result)
		}
	}

	if len(result.Contras) == 0 {
		fmt.Printf("%s[NEURON] Groq가 contra 뉴런을 제안하지 않음 — 면역 상태 양호%s\n", ansiGreen, ansiReset)
		return
	}

	// 5. Display
	fmt.Printf("╔══════════════════════════════════════╗\n")
	fmt.Printf("║   🧬 AUTO-NEURONIZE RESULTS         ║\n")
	fmt.Printf("╚══════════════════════════════════════╝\n\n")

	for i, c := range result.Contras {
		fmt.Printf("  %d. %s禁 %s%s\n", i+1, ansiMagenta, c.Name, ansiReset)
		fmt.Printf("     영역: %s/%s\n", c.Region, c.Category)
		fmt.Printf("     이유: %s\n", c.Rationale)
		if c.SourceError != "" {
			fmt.Printf("     원인: %s\n", c.SourceError)
		}
		fmt.Println()
	}

	// 6. Execute (if not dry run)
	if dryRun {
		fmt.Printf("%s  ⚠️  DRY RUN — 위 contra 뉴런은 생성되지 않았습니다.%s\n", ansiYellow, ansiReset)
		logEpisode(brainRoot, "NEURONIZE:DRY", fmt.Sprintf("%d contras suggested", len(result.Contras)))
		return
	}

	created := 0
	for _, c := range result.Contras {
		// Validate region
		region := c.Region
		if region == "" {
			region = "cortex"
		}
		if _, ok := regionPriority[region]; !ok {
			region = "cortex"
		}
		// Block brainstem/limbic
		if region == "brainstem" || region == "limbic" {
			fmt.Printf("  🛡️ Blocked: %s (P0/P1 보호)\n", c.Name)
			continue
		}

		// Build path
		category := c.Category
		if category == "" {
			category = "contra"
		}
		// Sanitize name for filesystem
		safeName := sanitizeNeuronName(c.Name)
		if safeName == "" {
			continue
		}

		neuronPath := fmt.Sprintf("%s/%s/%s", region, category, safeName)

		// Check if already exists
		fullPath := filepath.Join(brainRoot, strings.ReplaceAll(neuronPath, "/", string(filepath.Separator)))
		if _, err := os.Stat(fullPath); err == nil {
			fmt.Printf("  ⚠️  이미 존재: %s\n", neuronPath)
			fireNeuron(brainRoot, neuronPath)
			continue
		}

		// Grow the contra neuron
		if err := growNeuron(brainRoot, neuronPath); err != nil {
			fmt.Printf("  ❌ 생성 실패: %s — %v\n", neuronPath, err)
			continue
		}

		// Write rationale to the neuron folder
		if c.Rationale != "" {
			rationalePath := filepath.Join(fullPath, "rationale.md")
			content := fmt.Sprintf("# %s\n\n%s\n\n---\n원인: %s\n생성: %s (Auto-Neuronize)\n",
				c.Name, c.Rationale, c.SourceError, time.Now().Format("2006-01-02 15:04"))
			os.WriteFile(rationalePath, []byte(content), 0600)
		}

		// Add .contra file (inhibitory signal)
		contraFile := filepath.Join(fullPath, "1.contra")
		os.WriteFile(contraFile, []byte{}, 0600)

		fmt.Printf("  %s🌱 생성: %s%s\n", ansiGreen, neuronPath, ansiReset)
		created++
	}

	fmt.Printf("\n  📊 결과: %d contra 뉴런 생성\n", created)
	logEpisode(brainRoot, "NEURONIZE", fmt.Sprintf("%d contras created from %d corrections", created, len(corrections)))

	if created > 0 {
		autoReinject(brainRoot)
	}
}

// ─── runPolarize: 긍정형 뉴런 → 부정형 전환 ───

func runPolarize(brainRoot string, dryRun bool) {
	fmt.Printf("%s[NEURON] Polarity Shift Engine Initializing...%s\n", ansiCyan, ansiReset)
	if dryRun {
		fmt.Printf("%s  ⚠️  DRY RUN — 전환 목록만 표시%s\n", ansiYellow, ansiReset)
	}

	brain := scanBrain(brainRoot)

	// Scan for positive-pattern neurons (candidates for polarity shift)
	positivePatterns := regexp.MustCompile(`(?i)^(use_|always_|prefer_|enable_|ensure_|must_|keep_|apply_)`)
	englishName := regexp.MustCompile(`^[a-zA-Z_]+$`)

	type polarizeCandidate struct {
		Region  string
		Path    string
		Name    string
		Counter int
	}

	var candidates []polarizeCandidate

	for _, region := range brain.Regions {
		// Skip protected regions
		if region.Name == "brainstem" || region.Name == "limbic" {
			continue
		}

		for _, n := range region.Neurons {
			// Only target English-named positive neurons
			if !englishName.MatchString(n.Name) {
				continue
			}
			if positivePatterns.MatchString(n.Name) {
				candidates = append(candidates, polarizeCandidate{
					Region:  region.Name,
					Path:    n.Path,
					Name:    n.Name,
					Counter: n.Counter,
				})
			}
		}
	}

	if len(candidates) == 0 {
		fmt.Printf("%s[NEURON] 전환 대상 없음 — 긍정형 영어 뉴런이 발견되지 않음%s\n", ansiGreen, ansiReset)
		return
	}

	fmt.Printf("  🔍 전환 대상 발견: %d개\n\n", len(candidates))

	// If we have GROQ_API_KEY, use Groq for smart conversion; otherwise use rules
	apiKey := os.Getenv("GROQ_API_KEY")

	type shiftEntry struct {
		OriginalPath string `json:"original_path"`
		NewName      string `json:"new_name"`
		NewRegion    string `json:"new_region"`
		NewCategory  string `json:"new_category"`
		Rationale    string `json:"rationale"`
	}
	type polarizeResult struct {
		Shifts []shiftEntry `json:"shifts"`
	}

	var shifts []shiftEntry

	if apiKey != "" && len(candidates) > 0 {
		// Groq-powered smart conversion
		var sb strings.Builder
		sb.WriteString("다음 긍정형 뉴런들을 부정/억제형(Contra)으로 전환하십시오.\n\n")
		sb.WriteString("## 전환 대상\n")
		for _, c := range candidates {
			sb.WriteString(fmt.Sprintf("- %s/%s (counter:%d)\n", c.Region, c.Path, c.Counter))
		}
		sb.WriteString("\n각 뉴런의 과잉 적용을 방지하는 억제 규칙으로 전환하라.\n")

		fmt.Printf("  🌐 Groq API 호출 중...\n")
		groqReq := groqRequest{
			Model: "llama-3.3-70b-versatile",
			Messages: []groqMessage{
				{Role: "system", Content: polarizeSystemPrompt},
				{Role: "user", Content: sb.String()},
			},
			Temperature:    0.3,
			MaxTokens:      2048,
			TopP:           0.9,
			Stream:         false,
			ResponseFormat: &groqResponseFormat{Type: "json_object"},
		}

		respBody, err := callGroqRaw(apiKey, groqReq)
		if err != nil {
			fmt.Printf("%s[TRAUMA] Groq API 실패: %v — 규칙 기반으로 전환%s\n", ansiRed, err, ansiReset)
		} else {
			var result polarizeResult
			if err := json.Unmarshal([]byte(respBody), &result); err == nil {
				shifts = result.Shifts
			}
		}
	}

	// Fallback: rule-based conversion if Groq didn't work
	if len(shifts) == 0 {
		for _, c := range candidates {
			newName := ruleBasedPolarize(c.Name)
			parts := strings.SplitN(c.Path, "/", 2)
			category := ""
			if len(parts) > 1 {
				pathParts := strings.Split(parts[0], "/")
				if len(pathParts) > 0 {
					category = pathParts[0]
				}
			}
			shifts = append(shifts, shiftEntry{
				OriginalPath: c.Region + "/" + c.Path,
				NewName:      newName,
				NewRegion:    c.Region,
				NewCategory:  category,
				Rationale:    fmt.Sprintf("%s의 과잉 적용을 억제한다.", c.Name),
			})
		}
	}

	// Display
	fmt.Printf("╔══════════════════════════════════════╗\n")
	fmt.Printf("║   🔄 POLARITY SHIFT RESULTS          ║\n")
	fmt.Printf("╚══════════════════════════════════════╝\n\n")

	for i, s := range shifts {
		fmt.Printf("  %d. %s%s%s → %s%s%s\n", i+1,
			ansiDimGray, s.OriginalPath, ansiReset,
			ansiMagenta, s.NewName, ansiReset)
		if s.Rationale != "" {
			fmt.Printf("     이유: %s\n", s.Rationale)
		}
		fmt.Println()
	}

	if dryRun {
		fmt.Printf("%s  ⚠️  DRY RUN — 전환이 실행되지 않았습니다.%s\n", ansiYellow, ansiReset)
		logEpisode(brainRoot, "POLARIZE:DRY", fmt.Sprintf("%d shifts suggested", len(shifts)))
		return
	}

	// Execute shifts
	executed := 0
	for _, s := range shifts {
		safeName := sanitizeNeuronName(s.NewName)
		if safeName == "" {
			continue
		}

		region := s.NewRegion
		if region == "" {
			region = "cortex"
		}
		category := s.NewCategory
		if category == "" {
			category = "contra"
		}

		newPath := fmt.Sprintf("%s/%s/%s", region, category, safeName)

		// Don't overwrite existing
		fullPath := filepath.Join(brainRoot, strings.ReplaceAll(newPath, "/", string(filepath.Separator)))
		if _, err := os.Stat(fullPath); err == nil {
			fmt.Printf("  ⚠️  이미 존재: %s\n", newPath)
			continue
		}

		// Create the contra neuron
		if err := growNeuron(brainRoot, newPath); err != nil {
			fmt.Printf("  ❌ 생성 실패: %s — %v\n", newPath, err)
			continue
		}

		// Write rationale
		if s.Rationale != "" {
			rationalePath := filepath.Join(fullPath, "rationale.md")
			content := fmt.Sprintf("# %s\n\n%s\n\n---\n원본: %s\n생성: %s (Polarity Shift)\n",
				s.NewName, s.Rationale, s.OriginalPath, time.Now().Format("2006-01-02 15:04"))
			os.WriteFile(rationalePath, []byte(content), 0600)
		}

		// Add .contra file
		contraFile := filepath.Join(fullPath, "1.contra")
		os.WriteFile(contraFile, []byte{}, 0600)

		fmt.Printf("  %s🔄 전환: %s → %s%s\n", ansiGreen, s.OriginalPath, newPath, ansiReset)
		executed++
	}

	fmt.Printf("\n  📊 결과: %d/%d 극성 전환 완료\n", executed, len(shifts))
	logEpisode(brainRoot, "POLARIZE", fmt.Sprintf("%d/%d polarity shifts executed", executed, len(shifts)))

	if executed > 0 {
		autoReinject(brainRoot)
	}
}

// ─── callGroqRaw: returns raw response content string ───

func ruleBasedPolarize(name string) string {
	name = strings.ToLower(name)

	replacements := map[string]string{
		"use_":    "禁",
		"always_": "禁무조건_",
		"prefer_": "禁",
		"enable_": "禁",
		"ensure_": "禁강제_",
		"must_":   "禁필수_",
		"keep_":   "禁유지강제_",
		"apply_":  "禁적용강제_",
	}

	for prefix, replacement := range replacements {
		if strings.HasPrefix(name, prefix) {
			rest := strings.TrimPrefix(name, prefix)
			return replacement + rest + "_의존"
		}
	}

	return "禁" + name
}

// ─── sanitizeNeuronName: filesystem-safe neuron name ───

func sanitizeNeuronName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') ||
			r == '_' || r == '-' ||
			(r >= 0xAC00 && r <= 0xD7AF) || // 한글 음절
			(r >= 0x3131 && r <= 0x318E) || // 한글 자모
			(r >= 0x4E00 && r <= 0x9FFF) { // 한자 CJK (禁 등)
			return r
		}
		return '_'
	}, name)

	// Remove consecutive underscores
	for strings.Contains(name, "__") {
		name = strings.ReplaceAll(name, "__", "_")
	}
	name = strings.Trim(name, "_")

	// Rune-based truncation to prevent UTF-8 mid-character split
	runes := []rune(name)
	if len(runes) > 40 {
		name = string(runes[:40])
	}
	return name
}
