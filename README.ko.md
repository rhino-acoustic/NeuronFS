<p align="center">
  <img src="https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat-square&logo=go" />
  <img src="https://img.shields.io/badge/인프라_비용-$0-brightgreen?style=flat-square" />
  <img src="https://img.shields.io/badge/에이전트-ENTP_×_ISTJ-blueviolet?style=flat-square" />
  <img src="https://img.shields.io/badge/License-MIT-green?style=flat-square" />
</p>

<p align="center"><a href="README.ko.md">🇰🇷 한국어</a> · <a href="README.md">🇺🇸 English</a></p>

# 🧠 NeuronFS

**당신의 AI `.cursorrules` 파일은 죽었습니다. 이것이 대안입니다.**

> *폴더가 뉴런이다. 경로가 문장이다. 카운터가 시냅스 가중치다.*  
> *당신의 AI는 `mkdir` 하나로 배우고, 기억하고, 진화합니다.*

<p align="center">
  <img src="docs/dashboard.png" alt="NeuronFS 3D 뇌 대시보드" width="800" />
  <br/>
  <em>실시간 3D 뇌 시각화 — 7개 인지 영역에 걸친 251개 뉴런</em>
</p>

---

## 문제

모든 AI 코딩 어시스턴트는 세션 사이에서 **모든 것을 잊어버립니다.**

업계의 대응? 벡터 데이터베이스. 월 $70 구독료. 복잡한 임베딩 파이프라인. 환각하는 RAG.

**AI 메모리에 과금당하고 계신 겁니다.**

NeuronFS는 파일시스템 기반 인지 엔진입니다. 데이터베이스 없음. 임베딩 없음. 구독료 없음.  
`mkdir brain/cortex/new_rule && touch brain/cortex/new_rule/1.neuron` — 끝.

---

## 5가지 주장 (증거 포함)

### 1. "AI 규칙에 벡터 DB는 필요 없다"

규칙은 애매하지 않습니다. 정확합니다. `"console.log 쓰지 마"` 에는 코사인 유사도가 아니라, **몇 번 위반했는지 추적하는 카운터**가 필요합니다.

**증거:** 251개 뉴런, 인프라 비용 $0. [brain_v4/ 참조](./brain_v4/)

### 2. "`.cursorrules`는 죽었다"

정적 텍스트 파일은 학습하지 못합니다. 어떤 규칙이 중요한지 모릅니다. 5000줄까지 자라서 매 세션 3000 토큰을 낭비합니다.

NeuronFS 규칙은 사용 빈도에 따라 **자동 승격**됩니다. 10번 위반? 부트스트랩으로 이동 — 매 세션 주입. 한 번도 안 어기면? 쉰다.

**증거:** [harness.ps1](./harness.ps1) — 자동 위반 감지 + 카운터 기반 승격

### 3. "AI 에이전트에게 MBTI를 줘라"

같은 코드베이스를 두 에이전트에게 줬습니다. 하나는 ENTP(빌더), 하나는 ISTJ(감사관). ISTJ가 ENTP가 놓친 승격 임계값 버그를 찾아냈습니다.

**증거:** [evidence/agent_b_verification.md](./evidence/agent_b_verification.md) — 실제 로그

### 4. "당신의 AI는 기억상실이다. 내 건 아니다."

매 세션, NeuronFS는 251개 뉴런을 스캔하고 6.8KB 규칙 파일로 컴파일해서 AI 컨텍스트에 주입합니다. AI는 매 세션 어제 배운 것을 알고 시작합니다.

**증거:** `git log brain_v4/` — v1부터 v5.6까지의 인지 발달 기록

### 5. "`mkdir`이 AI 에이전트에게 필요한 유일한 API다"

```bash
# 규칙 생성
mkdir -p brain_v4/cortex/testing/new_rule
touch brain_v4/cortex/testing/new_rule/1.neuron

# 강화 (AI가 이 교훈을 또 배움)
mv brain_v4/cortex/testing/new_rule/1.neuron brain_v4/cortex/testing/new_rule/2.neuron

# 말살 (위험 패턴 감지)
touch brain_v4/cortex/testing/new_rule/bomb.neuron
```

API 키 없음. SDK 없음. `pip install` 없음. 파일시스템 기본 명령어만.

---

## 비교

| | NeuronFS | .cursorrules | Mem0 | Letta |
|---|---|---|---|---|
| **설치** | `go build` | 파일 생성 | `pip install` + DB | `pip install` + DB |
| **인프라 비용** | **$0** | $0 | $70+/월 | $50+/월 |
| **규칙 자동 승격** | ✅ 카운터 기반 | ❌ | ❌ | ❌ |
| **자가 성장** | ✅ 교정 → 뉴런 | ❌ | ❌ | LLM 의존 |
| **멀티 에이전트** | ✅ MBTI 페르소나 | ❌ | ❌ | ❌ |
| **전체 상태 확인** | `tree brain/` | `cat .cursorrules` | API/대시보드 | 대시보드 |
| **버전 관리** | Git 기본 내장 | 수동 | ❌ | ❌ |
| **안전장치** | `bomb.neuron` | ❌ | ❌ | ❌ |

---

## 빠른 시작

```bash
# 방법 A: 소스에서 빌드 (Go 1.22+ 필요)
git clone https://github.com/vegavery/NeuronFS.git
cd NeuronFS/runtime
go build -o ../neuronfs .

# 방법 B: 바이너리 다운로드 (Go 불필요)
curl -L https://github.com/vegavery/NeuronFS/releases/latest/download/neuronfs -o neuronfs
chmod +x neuronfs

# 실행
./neuronfs ./brain_v4           # 진단 모드
./neuronfs ./brain_v4 --api     # API + 대시보드 + 하트비트
./neuronfs ./brain_v4 --mcp     # MCP 서버 (stdio)

# http://localhost:9090 에서 3D 뇌 시각화 확인
```

## 뇌 아키텍처

```
brain_v4/
├── brainstem/       [P0] 핵심 정체성 — 읽기전용, 불변
├── limbic/          [P1] 감정 필터 — 긴급도, 도파민, 아드레날린
├── hippocampus/     [P2] 기억 — 교정 기록, 세션 로그
├── sensors/         [P3] 환경 — 도구, 브랜드, 제약조건
├── cortex/          [P4] 지식 — 코딩 규칙, 방법론
├── ego/             [P5] 성향 — 말투, 언어, 스타일
├── prefrontal/      [P6] 목표 — 프로젝트, TODO, 장기 방향
└── _agents/         멀티에이전트 통신 (inbox/outbox)
```

**하위복종 계단(Subsumption Cascade):** 낮은 P가 항상 높은 P를 억제.  
`brainstem`에 `bomb.neuron`이 있으면 → **모든 것이 멈춘다.**

---

## 멀티 에이전트: FORGE × SENTINEL

두 에이전트가 같은 뇌를 공유하되 다른 인지 프로필을 가집니다:

| | FORGE (Agent A) | SENTINEL (Agent B) |
|---|---|---|
| **MBTI** | ENTP | ISTJ |
| **인지 스택** | Ne-Ti-Fe-Si | Si-Te-Fi-Ne |
| **역할** | 빠르게 만들고, 부수기 | 모든 것을 검증, 아무것도 믿지 않기 |
| **같은 뉴런, 다른 출력** | "이걸로 뭘 더 할 수 있지?" | "작동한다는 증거를 보여줘." |

CDP 인젝션 + 파일 기반 inbox 통신:

```
Agent A 작성 → brain_v4/_agents/agent_b/inbox/msg.md
                  ↓ (bridge 3초 내 감지)
Agent B 채팅 수신 → 🤖 [agent_a→agent_b] 메시지
                  ↓ (Agent B 응답)
Agent B 작성 → brain_v4/_agents/agent_a/inbox/response.md
                  ↓ (bridge 감지)
Agent A 채팅 수신 → 🤖 [agent_b→agent_a] 응답
```

**실제 결과:** Agent B가 Agent A가 놓친 승격 임계값 버그를 독립적으로 발견.  
Agent B가 Go 네이티브 MCP 서버(368줄)도 구현하고 17/17 harness ALL PASS 확인.  
[증거 →](./evidence/)

---

## 신호 체계

| 파일 | 의미 | 효과 |
|------|------|------|
| `N.neuron` | 발화 카운터 | N이 클수록 강한 경로 |
| `dopamineN.neuron` | 보상 신호 | 칭찬 시 생성, 경로 강화 |
| `bomb.neuron` | 고통 / 차단기 | 3회 반복 실패 → 완전 정지 |
| `memory.neuron` | 에피소드 기억 | 컨텍스트 보존 |
| `*.dormant` | 수면 | 30일 미사용 → 자동 격리 |

---

## 자율 루프

```
AI 출력 → [auto-accept] → _inbox → [fsnotify] → 뉴런 성장
           ↓                                       ↓
      Groq 분석                              GEMINI.md 재주입
           ↓                                       ↓
     뉴런 교정 ────────────────→ AI 행동 변화
```

1. **fsnotify** — 파일 변경 감지 → 즉시 뉴런 생성
2. **하트비트** — 3분 유휴 → CDP로 다음 TODO 강제 주입
3. **유휴 엔진** — 5분 유휴 → Groq 자동 진화 → Git 스냅샷
4. **Git 판사** — 커밋 후 diff 분석 → 뉴런 감소 시 자동 롤백
5. **워치독 v2** — neuronfs + bridge + harness 건강 감시

---

## RAG와 뭐가 다른가?

RAG는 애매한 지식을 검색합니다. NeuronFS는 정확한 행동을 강제합니다.

| | RAG | NeuronFS |
|---|---|---|
| 목적 | "내가 뭘 아나?" | "어떻게 행동해야 하나?" |
| 저장 | 벡터 DB의 임베딩 | 디스크의 폴더 |
| 검색 | 코사인 유사도 (근사) | 정확한 경로 (결정적) |
| 비용 | $70+/월 | $0 |
| 자가 학습 | ❌ | ✅ 카운터 기반 승격 |

RAG는 질문에 답합니다. NeuronFS는 규율을 강제합니다. 경쟁이 아니라 보완입니다.

---

## 솔직한 한계

급진적 투명성을 믿습니다. 아직 안 되는 것:

- **강제력 없음.** AI가 GEMINI.md를 무시하면 막을 수 없습니다. harness로 사후 감지합니다.
- **~~카운터 극성.~~** ✅ 구현 완료 — intensity + polarity 필드가 API와 대시보드에 반영.
- **시맨틱 검색.** "비슷한 규칙 찾기" 없음. 정확한 경로만 접근 가능.
- **외부 사용자 0명.** 자체 도그푸드입니다. Star를 눌러서 바꿔주세요.

> *"벡터 데이터베이스도 필요 없고, 월 $70 구독료도 필요 없다. `mkdir`이면 된다."*

---

## 이야기 🇰🇷

한국의 한 PD가 몇 달간 AI가 세션마다 모든 걸 잊어버리는 걸 지켜봤습니다.

Mem0를 써봤습니다. 너무 비쌌습니다. .cursorrules를 써봤습니다. 너무 정적이었습니다. RAG를 써봤습니다. 너무 애매했습니다.

그래서 터미널을 열고 `mkdir brain`을 쳤습니다. 그것이 첫 번째 뉴런이었습니다.

251개 뉴런이 지난 지금, 다른 MBTI 성격을 가진 두 AI 에이전트가 그의 코드 품질을 놓고 논쟁하고 — 그가 놓친 버그를 찾고 있습니다.

독선적입니다. 논란의 여지가 있습니다. 그리고 작동합니다.

**⭐ 동의하면 Star를 눌러주세요. [동의하지 않으면 이슈를 열어주세요.](../../issues)**

---

## 라이선스

MIT License — 자유롭게 사용, 수정, 배포 가능.

Copyright (c) 2026 박정근 (PD) — VEGAVERY RUN®
