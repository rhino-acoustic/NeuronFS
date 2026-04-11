package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
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

// ─── String Utils ───

func boolStr(cond bool, trueVal, falseVal string) string {
	if cond {
		return trueVal
	}
	return falseVal
}

func truncate(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen] + "..."
	}
	return s
}

// ─── Neuron Name Utils ───

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

// ─── Context Extraction ───

// Collect hippocampus episode logs
func collectEpisodes(brainRoot string) []string {
	var result []string

	// 1. 기존 메모리 로그 수집
	logDir := filepath.Join(brainRoot, "hippocampus", "session_log")
	if entries, err := os.ReadDir(logDir); err == nil {
		memRegex := regexp.MustCompile(`^memory(\d+)\.neuron$`)
		type memFile struct {
			num     int
			content string
		}
		var mems []memFile

		for _, e := range entries {
			if m := memRegex.FindStringSubmatch(e.Name()); m != nil {
				n, _ := strconv.Atoi(m[1])
				content, err := os.ReadFile(filepath.Join(logDir, e.Name()))
				if err == nil && len(content) > 0 {
					mems = append(mems, memFile{num: n, content: strings.TrimSpace(string(content))})
				}
			}
		}
		sort.Slice(mems, func(i, j int) bool { return mems[i].num < mems[j].num })
		for _, m := range mems {
			result = append(result, "[MEMORY] "+m.content)
		}
	}

	// 2. 신규 JSON Signal 수집 (Neuro-Lifecycle)
	signalDir := filepath.Join(brainRoot, "hippocampus", "_signals")
	if sigEntries, err := os.ReadDir(signalDir); err == nil {
		for _, e := range sigEntries {
			if strings.HasSuffix(e.Name(), ".json") {
				content, err := os.ReadFile(filepath.Join(signalDir, e.Name()))
				if err == nil && len(content) > 0 {
					result = append(result, "[SIGNAL] "+strings.TrimSpace(string(content)))
				}
			}
		}
	}

	// 3. Growth trajectory (최근 10개 — 궤적 학습)
	growthFile := filepath.Join(brainRoot, "hippocampus", "session_log", "growth.log")
	if data, err := os.ReadFile(growthFile); err == nil {
		lines := strings.Split(strings.TrimSpace(string(data)), "\n")
		start := 0
		if len(lines) > 10 {
			start = len(lines) - 10
		}
		for _, line := range lines[start:] {
			line = strings.TrimSpace(line)
			if line != "" {
				result = append(result, "[TRAJECTORY] "+line)
			}
		}
	}

	return result
}

// Build brain summary for prompt
func buildBrainSummary(brain Brain, result SubsumptionResult) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Total neurons: %d | Total activation: %d | Bomb: %s\n\n",
		result.TotalNeurons, result.TotalCounter, boolStr(result.BombSource != "", result.BombSource, "none")))

	for _, region := range brain.Regions {
		icon := regionIcons[region.Name]
		sb.WriteString(fmt.Sprintf("%s %s (%d neurons):\n", icon, region.Name, len(region.Neurons)))

		// Sort by counter descending
		neurons := make([]Neuron, len(region.Neurons))
		copy(neurons, region.Neurons)
		sort.Slice(neurons, func(i, j int) bool {
			return neurons[i].Counter > neurons[j].Counter
		})

		for _, n := range neurons {
			status := ""
			if n.IsDormant {
				status = " [DORMANT]"
			}
			if n.HasBomb {
				status = " [BOMB]"
			}
			dopStr := ""
			if n.Dopamine > 0 {
				dopStr = fmt.Sprintf(" 🟢dopa:%d", n.Dopamine)
			}
			sb.WriteString(fmt.Sprintf("  - %s (counter:%d%s%s)\n", n.Path, n.Counter, dopStr, status))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// Build the evolution prompt
func buildEvolvePrompt(episodes []string, brainSummary string, _ SubsumptionResult) string {
	var sb strings.Builder

	sb.WriteString("You are the NeuronFS Evolution Engine (The REM Phase Consolidator). You analyze a cognitive AI system's short-term signals and episode logs to determine which memories should become permanent long-term rules (Neurons).\n\n")
	sb.WriteString("## NeuronFS Axioms\n")
	sb.WriteString("- Folder = Neuron (name is meaning, depth is specificity)\n")
	sb.WriteString("- File = Firing Trace (N.neuron = counter/activation strength)\n")
	sb.WriteString("- Path = Sentence (brain/cortex/frontend/css → 'cortex > frontend > css')\n")
	sb.WriteString("- Counter = Activation (higher = stronger/myelinated path)\n")
	sb.WriteString("- dopamineN.neuron = positive reinforcement\n")
	sb.WriteString("- bomb.neuron = circuit breaker (blocks entire region)\n")
	sb.WriteString("- .dormant = pruned/inactive neuron (ISOLATION, never deletion)\n\n")

	sb.WriteString("## Brain Regions (7, prioritized — Subsumption Architecture)\n")
	sb.WriteString("P0:brainstem (conscience/survival) > P1:limbic (emotion) > P2:hippocampus (memory) > P3:sensors (environment) > P4:cortex (knowledge) > P5:ego (tone/style) > P6:prefrontal (goals)\n\n")

	sb.WriteString("## 🧠 Owner Context (from brain state — DO NOT MODIFY)\n")
	sb.WriteString("The owner's identity, brand, and projects are encoded as neurons in ego/sensors/prefrontal regions.\n")
	sb.WriteString("Read the Brain State below to understand the owner's context.\n")
	sb.WriteString("NEVER modify brainstem, limbic, or sensors/brand neurons.\n\n")

	sb.WriteString("## STRICT RULES\n")
	sb.WriteString("These are read from the brainstem region neurons above. They are inviolable.\n\n")

	sb.WriteString("## Valid Regions for grow paths (분류 판단 기준)\n")
	sb.WriteString("NEVER grow into brainstem or limbic — these are READ-ONLY.\n")
	sb.WriteString("- cortex/dev/禁*: 코딩 금지 규칙 (하드코딩, 중복생성 등 범용 개발 규칙)\n")
	sb.WriteString("- cortex/dev/推*: 코딩 추천 규칙 (로컬깃활용, 프로젝트관리 등)\n")
	sb.WriteString("- cortex/methodology/*: 방법론 (코드리뷰, 테스트 전략 등)\n")
	sb.WriteString("- hippocampus/에러_패턴/*: 반복 에러 패턴 기록\n")
	sb.WriteString("- hippocampus/에피소드/*: 일회성 사건 기록 (counter=1)\n")
	sb.WriteString("- sensors/brand/*: NEVER TOUCH — 브랜드 정체성\n")
	sb.WriteString("- prefrontal/project/*: 프로젝트 목표/계획\n\n")

	sb.WriteString("## 🧠 Region 분류 사고법 (AI 판단 모델)\n")
	sb.WriteString("질문 순서대로 분류하라:\n")
	sb.WriteString("1. '이 규칙이 모든 프로젝트에 적용되는가?' → YES면 cortex (개발규칙), NO면 다음\n")
	sb.WriteString("2. '이 규칙이 특정 프로젝트/브랜드에 한정되는가?' → YES면 sensors 또는 prefrontal\n")
	sb.WriteString("3. '이것은 반복 에러 패턴인가?' → YES면 hippocampus/에러_패턴\n")
	sb.WriteString("4. '이것은 일회성 사건인가?' → YES면 hippocampus/에피소드 (또는 무시)\n")
	sb.WriteString("5. '300명 수용 가능한 장소' 같은 검색 결과 → NEVER promote (Signal로 남겨라)\n")
	sb.WriteString("6. brainstem에 넣을 만한 범용 절대규칙은 극히 드물다 — 99%는 cortex에 간다\n\n")

	sb.WriteString("## Current Brain State\n")
	sb.WriteString("```\n")
	sb.WriteString(brainSummary)
	sb.WriteString("```\n\n")

	if len(episodes) > 0 {
		sb.WriteString("## Recent Signals & Episode Log (Short-Term Memory)\n")
		sb.WriteString("```\n")
		// Limit episodes
		for i, e := range episodes {
			if i >= 15 {
				break
			}
			sb.WriteString(fmt.Sprintf("- %s\n", e))
		}
		sb.WriteString("```\n\n")
	}

	// Trajectory Dataset — growth.log 궤적을 episodes에서 추출
	sb.WriteString("## Growth Trajectory (Brain Evolution History)\n")
	sb.WriteString("Use this to understand trends: corrections↓ = evolution working, corrections↑ = regression.\n")
	sb.WriteString("```\n")
	for _, ep := range episodes {
		if strings.Contains(ep, "TRAJECTORY") || strings.Contains(ep, "META_EVOLVE_FAIL") {
			sb.WriteString(fmt.Sprintf("- %s\n", ep))
		}
	}
	sb.WriteString("```\n\n")

	sb.WriteString("## Output Format (valid JSON)\n")
	sb.WriteString("{\n")
	sb.WriteString("  \"summary\": \"Brief overall feeling / consolidation narrative\",\n")
	sb.WriteString("  \"insights\": [\"insight 1\", \"insight 2\"],\n")
	sb.WriteString("  \"actions\": [\n")
	sb.WriteString("    {\n")
	sb.WriteString("      \"type\": \"grow\",\n")
	sb.WriteString("      \"path\": \"cortex/dev/design_systems/fast_components\",\n")
	sb.WriteString("      \"reason\": \"Repeatedly successful pattern in frontend tasks\"\n")
	sb.WriteString("    },\n")
	sb.WriteString("    {\n")
	sb.WriteString("      \"type\": \"signal\",\n")
	sb.WriteString("      \"path\": \"limbic/rewards/dopamine_loop\",\n")
	sb.WriteString("      \"signal\": \"dopamine\",\n")
	sb.WriteString("      \"reason\": \"High user satisfaction noted\"\n")
	sb.WriteString("    }\n")
	sb.WriteString("  ]\n")
	sb.WriteString("}\n")

	return sb.String()
}
