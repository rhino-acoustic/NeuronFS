# 🏗️ NeuronFS Runtime Architecture

> *"파일명만 나열하면 시스템 전체가 읽힌다. 이것이 Path = Sentence의 자기 증명이다."*

---

## The Journey: 모놀리스에서 뇌 구조로

```
Day 1 (2026-03-26)  │  main.go 1,200줄 — 하나의 파일에 모든 것
Day 5 (2026-03-30)  │  main.go 2,800줄 — 성장통. 함수를 찾으려면 Ctrl+F
Day 8 (2026-04-02)  │  main.go 3,538줄 — 한계. AI가 맥락을 잃기 시작
Day 10 (2026-04-05) │  30개 파일, 10,920줄 — 파일명 = 문서. AI가 즉시 복귀
```

### 왜 분리했나?

3,538줄짜리 `main.go`를 AI에게 보여주면:
- **"이 코드베이스 뭐야?"** → 답을 얻으려면 3분 이상 스크롤
- 새 세션마다 문맥 복구에 **10분+** 소요
- 하나를 고치면 다른 하나가 깨지는 **공포**

분리 후:
- **"이 코드베이스 뭐야?"** → 파일명 30개를 읽으면 **10초**에 파악
- `/api/codemap` 한 번 호출 → JSON으로 전체 구조 즉시 복구
- 파일명 자체가 문서 — **NeuronFS가 자기 철학을 코드에 증명**

---

## The Beautiful Structure: 자기참조 아키텍처

NeuronFS의 핵심 철학은 **"Path = Sentence"** — 경로 자체가 의미를 담는다.

이 원리가 코드베이스 자체에도 적용되어 있다:

```
brain.go         → 뇌를 스캔한다
emit.go          → 규칙을 방출한다
inject.go        → 프롬프트에 주입한다
evolve.go        → 스스로 진화한다
watch.go         → 변화를 감시한다
lifecycle.go     → 생명주기를 관리한다
supervisor.go    → 자식을 감독한다
similarity.go    → 유사도를 계산한다
neuron_crud.go   → 뉴런을 생성/삭제한다
transcript.go    → 대화를 기록한다
```

**이것은 단순한 "좋은 네이밍"이 아니다.** NeuronFS가 뇌의 폴더에게 요구하는 것과 동일한 규칙을 **자기 코드에도 적용한 재귀적 자기 증명**이다.

---

## 30 Files — The Complete Map

> 실시간 데이터: `GET http://localhost:9090/api/codemap`

### 🫀 Core (뇌의 심장)

| 파일 | 줄 | 무엇을 하는가 |
|------|-----|-------------|
| `main.go` | 396 | CLI 디스패처. 오직 진입점만. |
| `brain.go` | 439 | 뇌 스캔, 구조체, Subsumption 실행 |
| `init.go` | 251 | 뇌 초기화 (`--init`) |
| `awakening.go` | 346 | 부팅 시퀀스 |

### 🧠 Intelligence (지능)

| 파일 | 줄 | 무엇을 하는가 |
|------|-----|-------------|
| `emit.go` | 858 | _rules.md 생성 — Tier 0/1/2 규칙 컴파일 |
| `emit_helpers.go` | 581 | 인덱스, 트리 렌더링, 경로 유틸 |
| `emit_tiers.go` | 273 | `--emit auto/all` + 자동 백업 + 에디터 감지 |
| `neuronize.go` | 760 | Groq LLaMA 기반 자율 뉴런 생성 |
| `evolve.go` | 644 | 자율 진화 엔진 |
| `similarity.go` | 261 | 토큰화, 유사도 계산 (순수 리프) |
| `lifecycle.go` | 378 | 프루닝, 디케이, 중복 제거 |

### 🔌 Interface (외부 연결)

| 파일 | 줄 | 무엇을 하는가 |
|------|-----|-------------|
| `api_server.go` | 967 | REST API + `/api/codemap` |
| `mcp_server.go` | 828 | MCP stdio 서버 (AI IDE 통합) |
| `mcp_tools_native.go` | 155 | MCP 도구 등록 |
| `dashboard.go` | 486 | 대시보드 HTTP 서버 |
| `dashboard_html.go` | 11 | `go:embed dashboard.html` (1,156줄 → 11줄) |
| `adapter.go` | 114 | 멀티-IDE 어댑터 |

### ⚡ Operations (운영)

| 파일 | 줄 | 무엇을 하는가 |
|------|-----|-------------|
| `neuron_crud.go` | 277 | grow, fire, rollback, signal |
| `inject.go` | 287 | 더티 플래그, 인박스, 인젝션 루프 |
| `watch.go` | 135 | fsnotify 제로폴링 감시 |
| `supervisor.go` | 577 | 3-프로세스 감독자 |
| `transcript.go` | 405 | Git 스냅샷, IDLE 엔진, 전사 |
| `diag.go` | 261 | 진단, CODE_MAP 자동 갱신 |
| `cli_commands.go` | 66 | --stats, --vacuum |
| `exec_safe.go` | 31 | 30초 타임아웃 안전 exec |

### 🔒 Security (보안)

| 파일 | 줄 | 무엇을 하는가 |
|------|-----|-------------|
| `access_control.go` | 152 | RBAC 접근 제어 |
| `crypto_neuron.go` | 81 | AES-256 암호화 |
| `dek_manager.go` | 318 | DEK 키 관리 |
| `merkle_chain.go` | 205 | 무결성 머클 해시체인 |

### 🛡️ Resilience (회복력)

| 파일 | 줄 | 무엇을 하는가 |
|------|-----|-------------|
| `flatline_poc.go` | 420 | 패닉 데스 스크린 |
| `physical_hooks.go` | 157 | USB/텔레그램 물리 알람 |

---

## 모듈화의 변천사

```
v4.0  main.go  ████████████████████████████████████████  3,538줄
v4.3  main.go  ████████████████                          710줄   (-80%)
v4.4  main.go  ████████                                  396줄   (-89%)

v4.0  emit.go  ██████████████████████████████            1,432줄
v4.4  emit.go  ████████████████████                      858줄
      helpers  ████████████                              581줄   (분리)

v4.0  dashboard_html.go  ████████████████████████        1,156줄
v4.4  dashboard_html.go  █                               11줄    (-99%, go:embed)
      dashboard.html     ████████████████████████        1,170줄 (외부화)
```

### 원칙: Strangler Fig Pattern

모놀리스를 한 번에 쪼개지 않았다. **교살무화과**처럼 하나씩 감싸서 분리:

1. `similarity.go` — 순수 함수, 의존성 제로 → 먼저 분리
2. `lifecycle.go` — 프루닝/디케이, brain.go만 의존 → 안전하게 분리
3. `emit.go` + `emit_helpers.go` — 가장 큰 덩어리를 두 개로
4. `watch.go`, `diag.go` — main.go에서 마지막으로 추출
5. `dashboard_html.go` — go:embed로 HTML 외부화 (1,156줄 → 11줄)

**매 단계마다 `go vet` + `go build` 3중 검증.** 한 번도 빌드가 깨지지 않았다.

---

## Defense in Depth: 문맥 복구 8계층

| 계층 | 방어 | 자동화 |
|------|------|--------|
| L0 | GEMINI.md에 CODE_MAP 경로 주입 | ✅ emit.go |
| L1 | `/api/codemap` JSON API | ✅ 실시간 |
| L2 | 대시보드 📋 CODE 패널 | ✅ 30초 새로고침 |
| L3 | CODE_MAP.md 파일 (IDLE 갱신) | ✅ 5분마다 |
| L4 | git snapshot (자동 커밋) | ✅ IDLE 루프 |
| L5 | NAS Z: 물리 백업 | ✅ robocopy |
| L6 | go vet 자동 실행 | ✅ IDLE 루프 |
| L7 | VET_FAIL 에피소드 기록 | ✅ hippocampus |

---

## 검증

```bash
cd runtime/
go vet ./...   # 0 errors — 항상
go build .     # ~8.3s — 단일 바이너리
```

## 라이브 대시보드

| 엔드포인트 | 용도 |
|-----------|------|
| `http://localhost:9090/` | 3D 뇌 토폴로지 |
| `http://localhost:9090/api/brain` | 전체 뇌 데이터 |
| `http://localhost:9090/api/usage` | Groq API 사용량 + 감정 상태 + uptime |
| `http://localhost:9090/api/emotion` | GET: 현재 감정 / POST: 감정 설정 |
| `http://localhost:9090/api/codemap` | 런타임 파일 트리 JSON |

---

## Limbic EmotionPrompt 엔진

`emit_bootstrap.go`의 감정 상태 머신은 두 개의 피어리뷰 연구에 기반합니다:

1. **Anthropic** "On the Biology of a LLM" (2025) — 기능적 감정(functional emotions) 발견
2. **Microsoft/CAS** EmotionPrompt (arXiv:2307.11760) — BIG-Bench +115%, Human Eval +10.9%

### 데이터 흐름

```
[대시보드 버튼] ──→ POST /api/emotion ──→ limbic/_state.json
[전사 분석] ──→ autoSetEmotion() ──┘         │
                                              ↓
                               emitBootstrap() → GEMINI.md
                                              ↓
                               emotionBehaviors[emo][tier] 주입
                               (5감정 × 3단계 = 15개 행동지시)
```

### 자동 감정 전환 (transcript.go)

`digestTranscripts()`가 사용자 발화를 분석:
- 답답함 키워드 3회↑ → `autoSetEmotion("긴급", 0.5~0.9)`
- 만족 키워드 3회↑ → `autoSetEmotion("만족", 0.6)`
- 시간 경과 시 `decay_rate`에 의해 자동 감쇠 → neutral 리셋

---

## 3대 유즈케이스: Solo → Multi-Agent → Enterprise

### 1. 솔로 개발자 — One Brain, All AIs

```
neuronfs --emit all → .cursorrules + CLAUDE.md + GEMINI.md 동시 생성
neuronfs --emit auto → 프로젝트 내 존재하는 에디터 설정만 자동 감지
```

AI 도구를 자유롭게 전환해도 규칙이 증발하지 않는다.

### 2. 멀티에이전트 — Swarm Orchestration

```
                    ┌─ bot1 (ego/ENTP) ─── 공격적 해체
                    │
supervisor.go ──────┼─ bot2 (ego/ISTJ) ─── 보수적 검증
                    │
                    └─ bot3 (ego/QA)  ─── 검증 전용
                          │
                    inject.go ← 크로스 브레인 인박스
```

- 모든 에이전트가 **같은 brain_v4/** 를 읽되, `ego/` 폴더만 다름
- `inject.go`의 인박스 시스템으로 에이전트 간 비동기 메시징
- `supervisor.go`가 3개 프로세스를 감독 (crash 시 자동 재시작)

### 3. 엔터프라이즈 — Corporate Brain (사내 브레인)

```
CTO                    Team
 ├─ 禁/보안위반          ├─ git clone company_brain
 ├─ 禁/컴플라이언스       ├─ neuronfs --emit all
 ├─ 必/코드리뷰          └─ Day 0 = 이미 시니어 수준 AI 장착
 └─ .jloot 카트리지 배포
    (암호화 · 버저닝 · 유료 판매)
```

- CTO가 P0(brainstem) 절대 규칙을 큐레이션 → 팀 전체에 배포
- `.jloot` 카트리지: XChaCha20 암호화된 뇌 패키지 → 팀/외부 판매 가능
- 신입사원 온보딩: 뇌를 clone하는 순간 10,000번의 교정이 즉시 적용

---

*이 문서는 NeuronFS의 IDLE 루프에 의해 자동 갱신됩니다.*
