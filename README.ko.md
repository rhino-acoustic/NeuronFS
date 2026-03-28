<p align="center">
  <img src="https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat-square&logo=go" />
  <img src="https://img.shields.io/badge/인프라-₩0-brightgreen?style=flat-square" />
  <img src="https://img.shields.io/badge/뉴런-256-blue?style=flat-square" />
  <img src="https://img.shields.io/badge/MIT-green?style=flat-square" />
</p>

<p align="center"><a href="README.ko.md">🇰🇷 한국어</a> · <a href="README.md">🇺🇸 English</a></p>

# 🧠 NeuronFS

**폴더가 뉴런이다. 경로가 문장이다. 카운터가 시냅스 가중치다.**

<p align="center">
  <img src="docs/dashboard.png" alt="NeuronFS 3D 뇌 대시보드" width="800" />
  <br/>
  <sub>실시간 3D 대시보드 — 7개 영역, 256개 뉴런, 극성 색상 (빨강=교정 ↓, 초록=보상 ↑)</sub>
</p>

---

## 왜 만들었나

AI가 세션마다 모든 걸 잊어버린다. 몇 달을 봤다.

Mem0를 써봤다. 월 $70. 규칙 강제가 안 된다.  
.cursorrules를 써봤다. 5000줄이 되니까 매 세션 3000 토큰을 태운다. 어떤 규칙이 중요한지 모른다.  
RAG를 써봤다. "console.log 쓰지 마"에 코사인 유사도가 필요한가. 규칙은 정확해야지 비슷하면 안 된다.

터미널을 열고 `mkdir brain`을 쳤다. 폴더 하나가 첫 번째 뉴런이었다.

> *"벡터 DB도 필요 없고, $70/월 구독료도 필요 없다. `mkdir`이면 된다."*

---

## 실측 데이터

둥근 숫자 없다. 2026-03-29 01:08, 로컬 Windows 11 SSD 환경에서 측정한 값이다.

| 지표 | 수치 | 조건 |
|------|------|------|
| 뉴런 수 | 256개 | 593개 폴더, 0-byte `.neuron` 파일 |
| cortex (코딩 규칙) | 156개 | 전체의 61%. 가장 밀집한 영역 |
| GEMINI.md | 6,946 bytes (~1,736 토큰) | 256개 뉴런 → 7KB 압축 |
| API 응답 | 47ms | `GET /api/state`, SSD 로컬 |
| Go 바이너리 | 12.8MB | MCP 서버 포함, 단일 바이너리 |
| brain 디스크 | 4.3MB | `_rules.md` + 에이전트 통신 |
| harness | 17/17 PASS | F01-F07, P01-P05, M01-M03, B01-B02 |
| 인프라 비용 | ₩0 | 벡터 DB, Redis, 클라우드 없음 |

⚠️ **500+ 뉴런 성능은 미검증.** 선형 외삽하면 ~180ms 예상이지만, 실측은 없다.

---

## 경쟁 제품 비교

Pinecone에 $70/월 내고 있다면, 여기서 뭐가 다른지 먼저 보라.

| | NeuronFS | .cursorrules | Mem0 | Letta | Zep |
|---|---|---|---|---|---|
| **설치** | `go build` | 파일 생성 | `pip install` + DB | `pip install` + DB | Docker + DB |
| **인프라** | **₩0** | ₩0 | $70+/월 | $50+/월 | $40+/월 |
| **규칙 자동 승격** | ✅ 카운터 기반 | ❌ 수동 편집 | ❌ | ❌ | ❌ |
| **자가 성장** | ✅ 교정 → 뉴런 | ❌ | ❌ | LLM 의존 | 시계열만 |
| **멀티 에이전트** | ✅ MBTI 인지 프로필 | ❌ | ❌ | ❌ | ❌ |
| **상태 전체 확인** | `tree brain/` | `cat` 파일 | API 호출 | 대시보드 | 대시보드 |
| **안전장치** | `bomb.neuron` | ❌ | ❌ | ❌ | ❌ |
| **망각** | `*.dormant` 자동 격리 | ❌ | ❌ | 수동 | TTL만 |

> 커뮤니티를 조사했다. Mem0는 듀얼스토어(벡터+KG), Letta는 OS레벨 메모리, Cognee는 비정형 구조화, Zep는 시계열 KG.  
> 전부 **인프라 과다.** 벤치는 좋고 프로덕션에서 깨진다. 암묵적 학습이 안 된다. dirty data에 모순 정보가 쌓인다.  
> NeuronFS는 반대 방향이다. 인프라 제로. 명시적 규칙만. 모순이 발생하면 `bomb.neuron`으로 차단한다.

---

## 작동 원리

### 뉴런 하나 만들기

```bash
mkdir -p brain_v4/cortex/testing/no_console_log
touch brain_v4/cortex/testing/no_console_log/1.neuron
```

경로 `cortex > testing > no_console_log`가 규칙 이름이 된다. `1.neuron`이 카운터다. 이게 전부다.

### 자동 승격이 핵심이다

.cursorrules와의 진짜 차이는 이것 하나다. 자주 위반하는 규칙이 자동으로 올라간다.

| 카운터 | 강도 | 동작 |
|--------|------|------|
| 1-4 | 일반 | `_rules.md`에만 기록 |
| 5-9 | 반드시 | 강조 표시 |
| 10+ | **절대** | GEMINI.md bootstrap에 직접 주입. 매 세션 읽힘 |

실제 TOP 5 뉴런 (2026-03-29):

| 경로 | 카운터 | 의미 |
|------|--------|------|
| `methodology > plan then execute` | 28 | 계획 먼저, 실행 나중 |
| `security > 禁평문 토큰` | 25 | API 키 평문 노출 금지 |
| `frontend > 禁인라인스타일` | 20 | CSS 인라인 금지 |
| `neuronfs > 실재 온톨로지` | 20 | 파일이 실재해야 규칙이다 |
| `frontend > 禁console log` | 17 | 프로덕션 console.log 금지 |

28번 교정당한 규칙이 맨 위에 있다. AI가 "계획 먼저"를 28번 어겼다는 뜻이다.

### 카운터 극성 (v5.7)

카운터만으로는 부족하다. "자주 교정당하는 것"과 "자주 칭찬받는 것"이 구분이 안 된다. 그래서 두 축으로 분리했다.

| 필드 | 계산 | 의미 |
|------|------|------|
| Intensity | `Counter + Dopamine` | 총 발화 횟수 |
| Polarity | `Dopamine / Intensity` | 0.0 = 순수 교정, 1.0 = 순수 보상 |

대시보드에서 빨간 점 = 교정이 많다 (AI가 계속 틀린다). 초록 점 = 보상이 많다 (AI가 잘한다).

---

## 아키텍처

```
brain_v4/
├── brainstem/       [P0] 핵심 정체성 — 읽기전용. 21개 뉴런
├── limbic/          [P1] 감정 필터 — 7개 뉴런
├── hippocampus/     [P2] 기억 — 10개 뉴런
├── sensors/         [P3] 환경 제약 — 37개 뉴런
├── cortex/          [P4] 지식/기술 — 156개 뉴런
├── ego/             [P5] 성향/톤 — 13개 뉴런
├── prefrontal/      [P6] 목표/계획 — 23개 뉴런
└── _agents/         에이전트 간 통신 (inbox/outbox)
```

**하위복종 계단.** P0이 P6를 항상 이긴다. `brainstem`에 `bomb.neuron`이 있으면 모든 출력이 멈춘다.

이름은 Rodney Brooks의 subsumption architecture에서 빌렸다. 원본은 로봇 모터 제어용이다. 하드웨어 레벨 억제와 텍스트 레벨 우선순위는 다르다. **이름을 빌렸을 뿐이다.** 하지만 원칙은 같다 — 안전 규칙이 편의 규칙을 항상 이겨야 한다.

### 신호 체계

| 파일 | 의미 | 발생 조건 |
|------|------|----------|
| `N.neuron` | 발화 카운터 | 교정 시 자동 증가 |
| `dopamineN.neuron` | 보상 신호 | 칭찬 시 생성 |
| `bomb.neuron` | 차단기 | 동일 실수 3회 반복 |
| `*.dormant` | 수면 | 30일 미발화 → 자동 격리 |
| `memory.neuron` | 에피소드 기억 | 특정 세션 보존용 |

---

## 멀티 에이전트: FORGE × SENTINEL

같은 뇌를 공유하는 두 AI. 인지 프로필이 다르다.

| | FORGE (Agent A) | SENTINEL (Agent B) |
|---|---|---|
| MBTI | ENTP | ISTJ |
| 인지 스택 | Ne-Ti-Fe-Si | Si-Te-Fi-Ne |
| 성향 | 빠르게 만들고 부순다 | 증거를 요구한다 |

MBTI가 유사과학인 건 안다. 하지만 AI에게는 작동한다. 인지기능 스택이 출력 편향을 만든다.

### 25분 교전 결과 (2026-03-29)

SENTINEL이 FORGE가 놓친 것 세 개:

1. **승격 버그.** `emit.go`가 `n.Counter < 10`만 보고 `Dopamine`을 무시했다. `禁console.log`(카운터 9 + 도파민 3 = 12)가 승격 안 됐다. SENTINEL이 지적. FORGE가 수정.
2. **README 7.5/10.** 개선안 6건: `echo.`→`touch`, "Why Not RAG?" 추가, Story 섹션 앵커링.
3. **MCP 서버.** SENTINEL이 독자적으로 `mcp_server.go` 368줄 작성. Node.js 래퍼 제거, Go 단일 바이너리 통합.

통신 프로토콜: `brain_v4/_agents/agent_b/inbox/`에 `.md` 파일 → `agent-bridge.mjs`가 3초 내 감지 → CDP 인젝션으로 상대 채팅에 전달.

[전체 로그 →](./evidence/)

---

## 자율 루프

```
AI 출력 → [auto-accept] → _inbox → [fsnotify] → 뉴런 성장
           ↓                                       ↓
      Groq 분석                              GEMINI.md 재주입
           ↓                                       ↓
     뉴런 교정 ────────────────→ AI 행동 변화
```

| 모듈 | 기능 | 트리거 |
|------|------|-------|
| fsnotify | 파일 변경 → 뉴런 즉시 생성 | FS 이벤트 |
| 하트비트 | 유휴 3분 → TODO 강제 주입 | 180s 간격 |
| 유휴 엔진 | 유휴 5분 → Groq 자동 진화 → Git 스냅샷 | 300s 타임아웃 |
| 워치독 v2 | neuronfs + bridge + harness 건강 감시 | 2시간 데몬 |

---

## 한계

토론하지 않는다. 사실을 말한다.

### 강제력이 없다

AI가 GEMINI.md를 안 읽으면 무력하다. OS 레벨의 enforcement가 없다. 위반은 harness로 사후에 잡는다. 이건 근본적 한계다.

### 시맨틱 검색이 없다

"비슷한 규칙 찾기"가 안 된다. 경로를 정확히 알아야 한다. 500개가 넘으면 수동 탐색이 불가능해질 수 있다. 이건 벡터 DB가 NeuronFS보다 잘하는 영역이다.

### 자작극 의혹

Groq에 GEMINI.md를 system prompt로 넣으면 당연히 따른다. **이건 NeuronFS의 공로가 아니라 system prompt의 공로다.** 진짜 검증은 "GEMINI.md 있을 때 vs 없을 때 위반율 비교"다. 아직 안 했다.

### 외부 사용자가 없다

내부 도그푸드만 했다. 다른 환경, 다른 AI, 다른 워크플로에서의 검증이 없다.

> 이건 정직이 아니라 전략이다. 한계를 숨기면 HN에서 3분 안에 깨진다.  
> 먼저 인정하면 신뢰가 된다.

---

## 빠른 시작

```bash
git clone https://github.com/vegavery/NeuronFS.git
cd NeuronFS/runtime && go build -o ../neuronfs .

./neuronfs ./brain_v4           # 진단 (스캔 + GEMINI.md 생성)
./neuronfs ./brain_v4 --api     # 대시보드 (localhost:9090)
./neuronfs ./brain_v4 --mcp     # MCP 서버 (stdio)
```

---

## 2026 트렌드와 NeuronFS의 위치

커뮤니티를 조사했다. 2026년 AI 메모리 시장의 흐름이 보인다.

| 트렌드 | NeuronFS 해당 여부 |
|--------|-------------------|
| governance as code | ✅ 폴더 구조가 곧 거버넌스 |
| git as memory | ✅ brain_v4 자체가 git repo |
| trust by design | ✅ bomb.neuron, harness 사후 검증 |
| multi agent 시스템 | ✅ FORGE × SENTINEL |
| 망각이 기능 (TTL 퇴거) | ✅ *.dormant 자동 격리 |
| hybrid memory | ⚠️ 부분. 시맨틱 레이어 없음 |
| observability 추적 | ✅ 대시보드 + API |
| SQLite middle ground | ❌ 해당 없음. 파일시스템만 씀 |

경쟁자들의 실패 패턴도 뉴런으로 기록했다:
- `community > 반면교사 > 운영복잡성 인프라과다` — Letta, Cognee
- `community > 반면교사 > 벤치는좋고 프로덕션깨짐` — Mem0 초기 버전
- `community > 반면교사 > dirty data 모순정보` — Zep
- `community > 반면교사 > context stuffing 성능저하` — .cursorrules 5000줄

남의 실패를 뉴런으로 만든다. 이것도 학습이다.

---

## 이야기 🇰🇷

한국의 PD가 만들었다. 영상이 본업이고 코딩은 도구다.

AI가 "console.log 쓰지 마"를 9번 어겼다. 10번째에 `mkdir brain_v4/cortex/frontend/coding/禁console_log`를 쳤다. 폴더 이름이 규칙이 됐다. 파일 이름이 카운터가 됐다. 그게 17번까지 올라갔다. AI가 더 이상 console.log를 안 쓴다.

과장인가? 실제 harness 로그를 보라. 17/17 PASS다.

256개 뉴런이 있다. 두 AI가 같은 뇌를 공유하면서 서로의 코드를 검증한다. ENTP는 "뭘 더 할 수 있지?"를 묻고, ISTJ는 "증거를 보여줘"를 묻는다. 둘 다 같은 폴더를 읽지만 다른 결론에 도달한다.

인프라 비용은 ₩0이다.

**⭐ 동의하면 Star. [아니면 Issue.](../../issues)**

---

MIT License · Copyright (c) 2026 박정근 (PD) — VEGAVERY RUN®
