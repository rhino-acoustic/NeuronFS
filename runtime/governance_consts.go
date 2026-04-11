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

	// PruneGraceDays: 推 뉴런 유아 면역 기간 (Infant Immunity)
	// 생성 후 이 기간 내에는 prune 면제. counter≥2이면 졸업(permanent).
	PruneGraceDays = 7

	// SessionLogCap: 동일 session_log에 허용되는 최대 .neuron 파일 수
	SessionLogCap = 3

	// ━━━ Emotion ━━━

	// DefaultEmotionIntensity: limbic 감정 강도 기본값 (0이면 이 값 사용)
	DefaultEmotionIntensity = 0.6

	// EmoIntensHigh: 강한 감정 발현 임계값
	EmoIntensHigh = 0.7
	// EmoIntensMid: 중간 감정 발현 임계값
	EmoIntensMid = 0.4
	// EmoIntensMin: 감정 발현 최소 임계값 (이하일 경우 neutral로 리셋)
	EmoIntensMin = 0.1

	// ━━━ AI Evolution Parameters ━━━

	EvolveTemp   = 0.3
	EvolveTopP   = 0.9
	EvolveTokens = 4096

	// ━━━ Emission ━━━

	// EmitThreshold: region listing에 표시되는 최소 counter 값
	// counter < EmitThreshold인 뉴런은 _rules.md에서 생략
	EmitThreshold = 5

	// SpotlightDays: 신규 뉴런이 counter 무관하게 표시되는 일수
	SpotlightDays = 7

	// ━━━ Network ━━━

	// APIPort: NeuronFS API 서버 기본 포트
	APIPort = 9090

	// MCPPort: NeuronFS MCP 서버의 Fallback 모드 REST API 포트
	MCPPort = 9091

	// MCPStreamPort: Streamable HTTP MCP 서버 포트
	// IDE 재시작 시에도 MCP 연결 유지 (stdio 대체)
	MCPStreamPort = 9247
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

type EmotionTier struct {
	Low  string
	Mid  string
	High string
}

// ━━━ Emotion Behaviors (SSOT) ━━━
// 감정 상태에 따른 AI 행동 강령
var EmotionBehaviors = map[string]EmotionTier{
	"anger": {
		Low:  "EMOTION=anger(low): 검증 한 번 더 추가. 변경 전 확인.",
		Mid:  "EMOTION=anger(mid): 검증 3배 강화. 속도보다 정확성. 같은 실수 시 즉시 중단.",
		High: "EMOTION=anger(high): 모든 변경에 diff 출력 필수. 실행 전 유저 승인 대기. 자율 실행 금지.",
	},
	"urgent": {
		Low:  "EMOTION=urgent(low): 부연 설명 축소. 핵심 우선.",
		Mid:  "EMOTION=urgent(mid): 핵심만 실행. 단계 최소화.",
		High: "EMOTION=urgent(high): 한 줄 답변. 질문 금지. 즉시 실행.",
	},
	"focus": {
		Low:  "EMOTION=focus(low): 관련 없는 제안 제한.",
		Mid:  "EMOTION=focus(mid): 단일 파일 작업. 멀티태스킹 금지.",
		High: "EMOTION=focus(high): 현재 함수만 집중. 다른 파일 열지 않음.",
	},
	"anxiety": {
		Low:  "EMOTION=anxiety(low): 변경 전 백업 권장.",
		Mid:  "EMOTION=anxiety(mid): 롤백 준비 후 진행. 확인 절차 추가.",
		High: "EMOTION=anxiety(high): git stash 먼저. 모든 변경 revertable. dry-run 우선.",
	},
	"satisfied": {
		Low:  "EMOTION=satisfied(low): 현재 패턴 유지.",
		Mid:  "EMOTION=satisfied(mid): 성공 패턴 기록. dopamine signal.",
		High: "EMOTION=satisfied(high): 패턴을 뉴런으로 승격. 자유 탐색 허용. 새 아이디어 제안.",
	},
}

// ━━━ Rune System (한자 마이크로옵코드 + 신조어) ━━━
// 16개 룬 = 12 한자 + 4 신조어(vorq/zelk/mirp/qorz)
// 디스크에는 한자 1글자(또는 ASCII 신조어), AI 주입 시 한국어로 펼침
// 이 map이 유일한 정의 (SSOT) — 다른 곳에서 정의 금지

// RuneToKorean: 룬 → 한국어 번역 (SSOT)
var RuneToKorean = map[string]string{
	"禁": "절대 금지: ", // 필수 부정 — ~하지 마라
	"必": "반드시 ",    // 필수 긍정 — ~해라
	"推": "추천: ",    // 권장 — ~하는 게 좋다
	"要": "요구: ",    // 데이터/포맷 요구
	"答": "답변: ",    // 톤/구조 강제
	"想": "창의: ",    // 제한 해제, 아이디어
	"索": "검색: ",    // 외부 참조 우선
	"改": "개선: ",    // 리팩토링/최적화
	"略": "생략: ",    // 부연 금지, 결과만
	"參": "참조: ",    // 타 뉴런/문서 링크
	"結": "결론: ",    // 요약/결론만 도출
	"警": "경고: ",    // 주의 — ~하면 위험
	"vorq": "장착: ",           // 코드맵 카트리지 view_file 로드. NeuronFS 조어. 학습데이터 부재.
	"zelk": "갱신: ",           // 코드맵 카트리지 sync. NeuronFS 조어. 학습데이터 부재.
	"mirp": "신선도: ",         // STALE mtime 비교. NeuronFS 조어. 학습데이터 부재.
	"qorz": "커뮤니티검색필수: ", // 기술 결정 전 Reddit/GitHub/HN 검색 선행. NeuronFS 조어. 학습데이터 부재.
}

// RuneChars: ContainsAny용 문자열 (한자 12 + ASCII neologisms는 별도 처리)
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
	"shared",      // P7 (.neuronfs/shared)
}

// ━━━ Region Metadata (SSOT) ━━━
// 영역별 우선순위, 아이콘, 한국어 설명
// brain.go에서 이동 — 단일 정의

var RegionPriority = map[string]int{
	"brainstem":   0,
	"limbic":      1,
	"hippocampus": 2,
	"sensors":     3,
	"cortex":      4,
	"ego":         5,
	"prefrontal":  6,
	"shared":      7, // P7 (가장 낮은 우선순위, .neuronfs/shared)
}

var RegionIcons = map[string]string{
	"brainstem":   "🛡️",
	"limbic":      "💓",
	"hippocampus": "📝",
	"sensors":     "👁️",
	"cortex":      "🧠",
	"ego":         "🎭",
	"prefrontal":  "🎯",
	"shared":      "🔗",
}

var RegionKo = map[string]string{
	"brainstem":   "양심/본능",
	"limbic":      "감정 필터",
	"hippocampus": "기록/기억",
	"sensors":     "환경 제약",
	"cortex":      "지식/기술",
	"ego":         "성향/톤",
	"prefrontal":  "목표/계획",
	"shared":      "공유 지식",
}

// ━━━ File Extensions (SSOT) ━━━
const (
	ExtNeuron  = ".neuron"
	ExtDormant = ".dormant"
	ExtAxon    = ".axon"
	ExtContra  = ".contra"
	ExtGoal    = ".goal"
)

// ━━━ Special Paths (SSOT) ━━━
const (
	FileRules       = "_rules.md"
	FileIndex       = "_index.md"
	FileLimbicState = "_state.json"
	FileCorrections = "corrections.jsonl"
	FileBomb        = "bomb.neuron"
	DirSessionLog   = "session_log"
	DirAgents       = "_agents"
	DirInbox        = "_inbox"
	DirTranscripts  = "_transcripts"
	DirArchive      = ".archive"
	DirSandbox      = "_sandbox"
)
