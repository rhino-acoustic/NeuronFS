package main

// ━━━ governance_consts.go ━━━
// SSOT: NeuronFS 거버넌스 상수 (Single Source of Truth)
//
// 모든 거버넌스 관련 매직넘버를 여기에 선언.
// 코드에서 직접 숫자를 쓰지 않고 이 const를 참조.
// DCI 테스트가 이 값들을 자동 검증.
//
// ⚠️ 값을 변경하면 DCI 테스트도 반드시 업데이트하라.

const (
	// ━━━ Similarity ━━━

	// MergeThreshold: grow/dedup에서 유사 뉴런을 병합하는 최소 유사도
	// grow: 새 뉴런 생성 시 기존과 비교 → 이 값 이상이면 기존 fire
	// dedup: 전역 중복 검사 시 이 값 이상이면 병합
	MergeThreshold = 0.6

	// ━━━ Lifecycle ━━━

	// MaxEpisodes: hippocampus/session_log의 최대 memory 파일 수 (circular buffer)
	MaxEpisodes = 10

	// PruneDays: 推 뉴런이 비활성 상태로 유지되면 dormant 처리되는 일수
	PruneDays = 3

	// SessionLogCap: 동일 session_log에 허용되는 최대 .neuron 파일 수
	SessionLogCap = 3
)
