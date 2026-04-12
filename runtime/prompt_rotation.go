package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// ============================================================================
// Module: Prompt Rotation Engine — Phase 54
// Prevents attention decay (immunity) by rotating master prompt structure.
// Same semantics, different surface form — identical to vorq principle.
// ============================================================================

// PromptVariant holds a structural variation of the master prompt
type PromptVariant struct {
	Opener    string
	Objective string
	Urgency   string
}

// promptVariants contains structural variations to prevent immunity
var promptVariants = []PromptVariant{
	{
		Opener:    "시스템이 정체 상태입니다. 즉각적인 개선이 필요합니다.",
		Objective: "현재 corrections.jsonl에 미처리된 에러가 있는지 확인하고, 발견 시 禁 뉴런으로 각인하라.",
		Urgency:   "정체는 퇴보입니다. 지금 움직여라.",
	},
	{
		Opener:    "경쟁 시스템이 업데이트되었습니다. NeuronFS도 대응해야 합니다.",
		Objective: "코드맵을 읽고 미구현 기능을 찾아 우선순위를 매겨라. 가장 급한 것 하나를 구현하라.",
		Urgency:   "시장은 기다리지 않습니다.",
	},
	{
		Opener:    "내부 감사 결과, 미이행 항목이 발견되었습니다.",
		Objective: "EVOLUTION_TODO.neuron을 읽고 미완료 항목을 찾아라. 가장 오래된 미완료부터 처리하라.",
		Urgency:   "미이행은 기술 부채입니다. 지금 상환하라.",
	},
	{
		Opener:    "새로운 사용자가 NeuronFS를 평가 중입니다. 첫인상이 결정됩니다.",
		Objective: "README.md와 랜딩 뉴런을 검토하고, 부족한 설명을 보강하라.",
		Urgency:   "두 번째 기회는 없습니다.",
	},
	{
		Opener:    "이전 에이전트가 실패한 작업이 있습니다. 당신이 완수해야 합니다.",
		Objective: "hippocampus/agent_results/를 읽고 실패한 작업을 찾아 재실행하라.",
		Urgency:   "실패를 방치하면 시스템이 신뢰를 잃습니다.",
	},
}

// RotatePrompt generates a structurally different master prompt each call
func RotatePrompt(brainRoot string) string {
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)
	variant := promptVariants[r.Intn(len(promptVariants))]

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("[NeuronFS 자율 진화 — %s]\n\n", time.Now().Format("15:04:05")))
	sb.WriteString(variant.Opener + "\n\n")
	sb.WriteString("[목표]\n")
	sb.WriteString(variant.Objective + "\n\n")
	sb.WriteString("[원칙]\n")
	sb.WriteString("- 기존 코드 삭제 금지. 적층(Strangler Fig)만 허용.\n")
	sb.WriteString("- go vet + go build 통과 필수.\n")
	sb.WriteString("- 결과를 뉴런으로 기록하라.\n\n")
	sb.WriteString(variant.Urgency + "\n")

	return sb.String()
}

// RotatePromptWithSeed generates a deterministic variant for testing
func RotatePromptWithSeed(brainRoot string, seed int) string {
	variant := promptVariants[seed%len(promptVariants)]

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("[NeuronFS 자율 진화 — Seed:%d]\n\n", seed))
	sb.WriteString(variant.Opener + "\n\n")
	sb.WriteString("[목표]\n")
	sb.WriteString(variant.Objective + "\n\n")
	sb.WriteString("[원칙]\n")
	sb.WriteString("- 기존 코드 삭제 금지. 적층(Strangler Fig)만 허용.\n")
	sb.WriteString("- go vet + go build 통과 필수.\n")
	sb.WriteString("- 결과를 뉴런으로 기록하라.\n\n")
	sb.WriteString(variant.Urgency + "\n")

	return sb.String()
}
