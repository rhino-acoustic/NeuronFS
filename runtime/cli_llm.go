package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

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

	var corrections []string

	historyPath := filepath.Join(brainRoot, "_inbox", "corrections_history.jsonl")
	if data, err := os.ReadFile(historyPath); err == nil && len(data) > 0 {
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				corrections = append(corrections, line)
			}
		}
	}

	correctionsPath := filepath.Join(brainRoot, "_inbox", "corrections.jsonl")
	if data, err := os.ReadFile(correctionsPath); err == nil && len(data) > 0 {
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				corrections = append(corrections, line)
			}
		}
	}

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

	var sb strings.Builder
	sb.WriteString("다음은 NeuronFS 시스템의 최근 교정 로그 및 에러 기록입니다.\n")
	sb.WriteString("이 로그들을 분석하여 '같은 실수를 방지하기 위한 contra(억제) 뉴런'을 생성하십시오.\n\n")
	sb.WriteString("## 교정 로그\n```\n")

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

	limit := 50
	if len(allNeurons) < limit {
		limit = len(allNeurons)
	}
	sb.WriteString(fmt.Sprintf("## 기존 뉴런 TOP %d (중복 생성 금지, 총 %d개)\n", limit, len(allNeurons)))
	for _, n := range allNeurons[:limit] {
		sb.WriteString(fmt.Sprintf("- %s (c:%d)\n", n.path, n.counter))
	}
	sb.WriteString("\n위 기존 뉴런과 겹치지 않는 새로운 contra 규칙만 생성하라. 최대 5개.\n")

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

	if dryRun {
		fmt.Printf("%s  ⚠️  DRY RUN — 위 contra 뉴런은 생성되지 않았습니다.%s\n", ansiYellow, ansiReset)
		logEpisode(brainRoot, "NEURONIZE:DRY", fmt.Sprintf("%d contras suggested", len(result.Contras)))
		return
	}

	created := 0
	for _, c := range result.Contras {
		region := c.Region
		if region == "" {
			region = "cortex"
		}
		if _, ok := regionPriority[region]; !ok {
			region = "cortex"
		}
		if region == "brainstem" || region == "limbic" {
			fmt.Printf("  🛡️ Blocked: %s (P0/P1 보호)\n", c.Name)
			continue
		}

		category := c.Category
		if category == "" {
			category = "contra"
		}
		safeName := sanitizeNeuronName(c.Name)
		if safeName == "" {
			continue
		}

		neuronPath := fmt.Sprintf("%s/%s/%s", region, category, safeName)

		fullPath := filepath.Join(brainRoot, strings.ReplaceAll(neuronPath, "/", string(filepath.Separator)))
		if _, err := os.Stat(fullPath); err == nil {
			fmt.Printf("  ⚠️  이미 존재: %s\n", neuronPath)
			fireNeuron(brainRoot, neuronPath)
			continue
		}

		if err := growNeuron(brainRoot, neuronPath); err != nil {
			fmt.Printf("  ❌ 생성 실패: %s — %v\n", neuronPath, err)
			continue
		}

		if c.Rationale != "" {
			rationalePath := filepath.Join(fullPath, "rationale.md")
			content := fmt.Sprintf("# %s\n\n%s\n\n---\n원인: %s\n생성: %s (Auto-Neuronize)\n",
				c.Name, c.Rationale, c.SourceError, time.Now().Format("2006-01-02 15:04"))
			os.WriteFile(rationalePath, []byte(content), 0600)
		}

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

	positivePatterns := regexp.MustCompile("(?i)^(use_|always_|prefer_|enable_|ensure_|must_|keep_|apply_)")
	englishName := regexp.MustCompile("^[a-zA-Z_]+$")

	type polarizeCandidate struct {
		Region  string
		Path    string
		Name    string
		Counter int
	}

	var candidates []polarizeCandidate

	for _, region := range brain.Regions {
		if region.Name == "brainstem" || region.Name == "limbic" {
			continue
		}

		for _, n := range region.Neurons {
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

		fullPath := filepath.Join(brainRoot, strings.ReplaceAll(newPath, "/", string(filepath.Separator)))
		if _, err := os.Stat(fullPath); err == nil {
			fmt.Printf("  ⚠️  이미 존재: %s\n", newPath)
			continue
		}

		if err := growNeuron(brainRoot, newPath); err != nil {
			fmt.Printf("  ❌ 생성 실패: %s — %v\n", newPath, err)
			continue
		}

		if s.Rationale != "" {
			rationalePath := filepath.Join(fullPath, "rationale.md")
			content := fmt.Sprintf("# %s\n\n%s\n\n---\n원본: %s\n생성: %s (Polarity Shift)\n",
				s.NewName, s.Rationale, s.OriginalPath, time.Now().Format("2006-01-02 15:04"))
			os.WriteFile(rationalePath, []byte(content), 0600)
		}

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

// ─── Main evolve function ───

func runEvolve(brainRoot string, dryRun bool) {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		fmt.Println("[FATAL] GROQ_API_KEY not set")
		fmt.Println("  Set: $env:GROQ_API_KEY = '<your-groq-api-key>'")
		os.Exit(1)
	}

	fmt.Println("═══ NeuronFS Evolve Engine ═══")
	fmt.Println("  🧬 Groq-powered autonomous brain evolution")
	if dryRun {
		fmt.Println("  ⚠️  DRY RUN — suggestions only, no execution")
	}
	fmt.Println()

	processInbox(brainRoot)

	episodes := collectEpisodes(brainRoot)
	fmt.Printf("  📝 Episodes collected: %d\n", len(episodes))

	brain := scanBrain(brainRoot)
	result := runSubsumption(brain)
	brainSummary := buildBrainSummary(brain, result)
	fmt.Printf("  🧠 Brain: %d neurons, activation: %d\n", result.TotalNeurons, result.TotalCounter)

	prompt := buildEvolvePrompt(episodes, brainSummary, result)

	fmt.Println("\n  🌐 Calling Groq API (llama-3.3-70b-versatile)...")
	startTime := time.Now()

	evoResp, err := callGroq(apiKey, prompt)
	if err != nil {
		fmt.Printf("[ERROR] Groq API: %v\n", err)
		os.Exit(1)
	}

	elapsed := time.Since(startTime)
	fmt.Printf("  ✅ Response received in %s\n\n", elapsed.Round(time.Millisecond))

	fmt.Println("╔══════════════════════════════════════╗")
	fmt.Println("║   🧬 EVOLUTION ANALYSIS              ║")
	fmt.Println("╚══════════════════════════════════════╝")
	fmt.Println()
	fmt.Printf("  📋 Summary: %s\n\n", evoResp.Summary)

	if len(evoResp.Insights) > 0 {
		fmt.Println("  💡 Insights:")
		for _, insight := range evoResp.Insights {
			fmt.Printf("    • %s\n", insight)
		}
		fmt.Println()
	}

	if len(evoResp.Actions) == 0 {
		fmt.Println("  ✅ No actions recommended — brain is in good shape.")
		logEpisode(brainRoot, "EVOLVE", "No actions needed. Brain healthy.")
		return
	}

	fmt.Printf("  🎯 Actions (%d):\n", len(evoResp.Actions))
	for i, action := range evoResp.Actions {
		icon := actionIcon(action.Type)
		fmt.Printf("    %d. %s [%s] %s\n", i+1, icon, action.Type, action.Path)
		if action.Signal != "" {
			fmt.Printf("       Signal: %s\n", action.Signal)
		}
		fmt.Printf("       Reason: %s\n", action.Reason)
	}
	fmt.Println()

	if dryRun {
		fmt.Println("  ⚠️  DRY RUN — no actions executed.")
		fmt.Println("  Run without --dry-run to apply these changes.")
		logEpisode(brainRoot, "EVOLVE:DRY", fmt.Sprintf("%d actions suggested", len(evoResp.Actions)))
		return
	}

	fmt.Println("  📸 [Phase 61] Snapshotting current brain state before execution...")
	gitSnapshot(brainRoot)

	fmt.Println("  ⚡ Executing actions...")
	executed := 0
	skipped := 0

	for _, action := range evoResp.Actions {
		switch action.Type {
		case "grow":
			if strings.HasPrefix(action.Path, "brainstem") || strings.HasPrefix(action.Path, "limbic") {
				fmt.Printf("    🛡️ Blocked: %s (P0/P1 보호 — grow 금지)\n", action.Path)
				skipped++
				continue
			}
			err := growNeuron(brainRoot, action.Path)
			if err != nil {
				fmt.Printf("    ❌ grow %s: %v\n", action.Path, err)
				skipped++
			} else {
				executed++
				go sendTelegramEvolve(brainRoot, action)
			}

		case "fire":
			fireNeuron(brainRoot, action.Path)
			executed++

		case "signal":
			if action.Signal == "" {
				action.Signal = "dopamine"
			}
			err := signalNeuron(brainRoot, action.Path, action.Signal)
			if err != nil {
				fmt.Printf("    ❌ signal %s %s: %v\n", action.Signal, action.Path, err)
				skipped++
			} else {
				executed++
			}

		case "prune", "decay":
			fullPath := filepath.Join(brainRoot, strings.ReplaceAll(action.Path, "/", string(filepath.Separator)))
			if _, err := os.Stat(fullPath); err == nil {
				dormantFile := filepath.Join(fullPath, "evolve.dormant")
				os.WriteFile(dormantFile, []byte(fmt.Sprintf("Evolved: %s\nReason: %s\n",
					time.Now().Format("2006-01-02"), action.Reason)), 0600)
				fmt.Printf("    💤 Pruned: %s\n", action.Path)
				executed++
			} else {
				fmt.Printf("    ❌ prune %s: not found\n", action.Path)
				skipped++
			}

		default:
			fmt.Printf("    ⚠️  Unknown action type: %s\n", action.Type)
			skipped++
		}
	}

	fmt.Printf("\n  📊 Result: %d executed, %d skipped\n", executed, skipped)

	if executed > 0 || len(evoResp.Actions) == 0 {
		signalDir := filepath.Join(brainRoot, "hippocampus", "_signals")
		if entries, err := os.ReadDir(signalDir); err == nil {
			cleared := 0
			for _, e := range entries {
				if strings.HasSuffix(e.Name(), ".json") {
					os.Remove(filepath.Join(signalDir, e.Name()))
					cleared++
				}
			}
			if cleared > 0 {
				fmt.Printf("  🧹 Cleared %d processed signals\n", cleared)
			}
		}
	}

	logEpisode(brainRoot, "EVOLVE", fmt.Sprintf("%d actions executed, %d skipped. Summary: %s",
		executed, skipped, truncate(evoResp.Summary, 200)))

	if executed > 0 {
		// Phase 61 Closed Loop: Validation Gate
		fmt.Printf("\n  🛡️ [Phase 61] Validating Evolution Integrity...\n")
		if err := VerifyBrainIntegrity(brainRoot); err != nil {
			fmt.Printf("  ❌ [VALIDATOR] Integrity check failed: %v\n", err)
			fmt.Printf("  🧨 Rolling back to previous Git snapshot...\n")
			
			if rbErr := rollbackAll(brainRoot); rbErr != nil {
				fmt.Printf("  🚨 [FATAL] Rollback failed: %v\n", rbErr)
			} else {
				fmt.Printf("  ✅ [RESTORE] Brain successfully rolled back to safe state.\n")
				// Append to corrections.jsonl to learn from this failure
				corrPath := filepath.Join(brainRoot, "_inbox", "corrections.jsonl")
				failLog := fmt.Sprintf(`{"ts":"%s","error":"Evolve Validation Failed","reason":"%v","type":"auto_rollback"}`+"\n", time.Now().Format(time.RFC3339), err)
				
				f, _ := os.OpenFile(corrPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if f != nil {
					f.WriteString(failLog)
					f.Close()
				}
			}
			return // Abort reinjection since we failed
		}

		fmt.Printf("  ✅ [Phase 61] Evolution accepted. Brain is stable.\n")
		autoReinject(brainRoot)
	}
}
