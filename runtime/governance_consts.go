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

	// ━━━ Emotion ━━━

	// DefaultEmotionIntensity: limbic 감정 강도 기본값 (0이면 이 값 사용)
	DefaultEmotionIntensity = 0.6

	// ━━━ Emission ━━━

	// EmitThreshold: region listing에 표시되는 최소 counter 값
	// counter < EmitThreshold인 뉴런은 _rules.md에서 생략
	EmitThreshold = 5

	// SpotlightDays: 신규 뉴런이 counter 무관하게 표시되는 일수
	SpotlightDays = 7

	// ━━━ Network ━━━

	// APIPort: NeuronFS API 서버 기본 포트
	APIPort = 9090
)

// ━━━ Emotion Mapping (SSOT) ━━━
// limbic _state.json의 한국어 감정 → 영어 변환
// emit_bootstrap에서 GEMINI.md 표시에 사용

var KoToEn = map[string]string{
	"분노": "anger",
	"긴급": "urgent",
	"만족": "satisfied",
	"불안": "anxiety",
	"집중": "focus",
}

// ━━━ Rune System (한자 마이크로옵코드) ━━━
// 12개 룬 = NeuronFS의 의미 압축 체계
// 디스크에는 한자 1글자, AI 주입 시 한국어로 펼침
// 이 map이 유일한 정의 (SSOT) — 다른 곳에서 정의 금지

// RuneToKorean: 룬 → 한국어 번역 (SSOT)
var RuneToKorean = map[string]string{
	"禁": "절대 금지: ",  // 필수 부정 — ~하지 마라
	"必": "반드시 ",      // 필수 긍정 — ~해라
	"推": "추천: ",       // 권장 — ~하는 게 좋다
	"要": "요구: ",       // 데이터/포맷 요구
	"答": "답변: ",       // 톤/구조 강제
	"想": "창의: ",       // 제한 해제, 아이디어
	"索": "검색: ",       // 외부 참조 우선
	"改": "개선: ",       // 리팩토링/최적화
	"略": "생략: ",       // 부연 금지, 결과만
	"參": "참조: ",       // 타 뉴런/문서 링크
	"結": "결론: ",       // 요약/결론만 도출
	"警": "경고: ",       // 주의 — ~하면 위험
}

// RuneChars: ContainsAny용 12룬 문자열
const RuneChars = "禁必推要答想索改略參結警"

// RuneKeys: 정규화용 룬 목록 (normalizeHanjaPath에서 사용)
func RuneKeys() []string {
	keys := make([]string, 0, len(RuneToKorean))
	for k := range RuneToKorean {
		keys = append(keys, k)
	}
	return keys
}

// ━━━ Region Names (SSOT) ━━━
// 7개 영역의 정렬된 이름 목록
// P0(brainstem) → P6(prefrontal) 순서 보장
// 코드에서 []string{"brainstem", "limbic", ...} 대신 이 함수를 사용하라

// RegionOrder: priority 순서대로 정렬된 영역 이름 (SSOT)
var RegionOrder = []string{
	"brainstem",   // P0
	"limbic",      // P1
	"hippocampus", // P2
	"sensors",     // P3
	"cortex",      // P4
	"ego",         // P5
	"prefrontal",  // P6
}

