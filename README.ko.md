<p align="center">
  <img src="https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat-square&logo=go" />
  <img src="https://img.shields.io/badge/Infra-$0-brightgreen?style=flat-square" />
  <img src="https://img.shields.io/badge/Neurons-433+-blue?style=flat-square" />
  <img src="https://img.shields.io/badge/Axons-6-purple?style=flat-square" />
  <img src="https://img.shields.io/badge/Zero_Dependencies-black?style=flat-square" />
  <img src="https://img.shields.io/badge/MIT-green?style=flat-square" />
</p>

<p align="center">
  <img src="docs/dashboard.png" alt="NeuronFS 대시보드 — 3D 뇌 시각화" width="800" />
  <br/>
  <a href="https://dashboarddeploy-six.vercel.app/"><strong>3D 대시보드 라이브 데모</strong></a>
</p>

<p align="center"><a href="README.ko.md">🇰🇷 한국어</a> · <a href="README.md">🇺🇸 English</a> · <a href="MANIFESTO.md">📜 매니페스토</a></p>

# NeuronFS
### *파일시스템 네이티브 계층형 규칙 메모리 — 무의존성 하네스 엔지니어링(Harness Engineering)*

> *"거대한 AI 모델에 더 많은 컨텍스트를 욱여넣는 것보다, 시스템의 뼈대(구조)를 완벽하게 설계하여 AI에 대한 의존도를 0으로 수렴하게 만드는 것."*
>
> AI가 "console.log 쓰지 마"라는 지시를 9번 어겼다. 10번째에 `mkdir brain/cortex/frontend/coding/禁console_log`를 만들었다. 폴더 이름이 물리적 규칙으로 시스템 프롬프트에 강제 삽입되었다. 카운터(가중치)가 17이 되었다. AI가 두 번 다시 해당 실수를 반복하지 않는다.
> 
> 이것이 NeuronFS가 추구하는 진정한 **하네스 엔지니어링(Harness Engineering)**의 본질이다.

---

## 요약 (TL;DR)

**`mkdir`이 시스템 프롬프트를 대체한다.** 폴더가 뉴런이고, 경로가 문장이며, 파일이 시냅스 가중치다.

### 기존 방식 대비 3대 우위

1. **비용 제로 (Zero Cost):** 에이전트의 기억을 관리하기 위해 Mem0나 Letta 같은 Vector DB를 사용하면 서버 배포 유지 비용이 발생하지만, NeuronFS는 당신의 로컬 OS 파일시스템을 직접 쓰므로 인프라 비용이 **₩0**이다.
2. **토큰 효율성과 관리의 용이성:** 수천 줄의 텍스트 뭉치에서 특정 규칙을 찾아내고 수정하는 것은 사람을 미치게 하지만, 트리 형태의 폴더 구조(`ls -R`)에서는 규칙의 탐색과 계층화, 물리적 삭제가 매우 직관적이다.
3. **극강의 이식성 (Portability):** 어떠한 외부 종속성(Dependencies)도 없는 단일 Go 언어 바이너리로 빌드되어, 어떤 OS 환경에서든 복사만 하면 즉시 실행할 수 있으며 곧바로 MCP(Model Context Protocol) 서버로도 동작한다.

```bash
# 규칙 생성 = 폴더 생성
mkdir -p brain/brainstem/禁fallback
touch brain/brainstem/禁fallback/1.neuron

# 컴파일 = 시스템 프롬프트 자동 생성 (Cursor, Windsurf, Claude Desktop 등)
neuronfs ./brain --emit cursor   # → .cursorrules
neuronfs ./brain --emit claude   # → CLAUDE.md
neuronfs ./brain --emit all      # → 모든 AI 포맷 동시 출력
```

---

## 설치 (The One-Liner Quickstart)

오픈소스 단일 바이너리 Go 엔진. 외부 의존성(Dependencies) 패키지 제로.

```bash
# Mac / Linux
curl -sL https://raw.githubusercontent.com/rhino-acoustic/NeuronFS/main/install.sh | bash

# Windows (PowerShell)
iwr https://raw.githubusercontent.com/rhino-acoustic/NeuronFS/main/install.ps1 -useb | iex

# 나만의 오프라인 뇌 초기화 (비어있는 7개 영역 기본 스캐폴딩 생성)
# ※ 대화형 프롬프트에서 [2]번 Master's Brain 옵션 선택 시 프리미엄 거버넌스 뼈대 복사 가능
neuronfs --init ./my_brain        

export GROQ_API_KEY="<your-groq-api-key>"      # Llama3 70B 기반 자율 폴더 통합 옵션용 (로컬 Ollama 연결 지원!)

# 컴파일 및 실행
neuronfs ./my_brain --emit all    # 시스템 프롬프트 컴파일
neuronfs ./my_brain --consolidate # Llama 3 기반 자동 뇌세포 통합/압축 엔진 (옵션)
neuronfs ./my_brain --api         # 대시보드 시각화 (localhost:9090)
```

---

## 목차

| | 섹션 | 내용 |
|---|---|---|
| 💡 | [핵심 구조](#핵심-구조) | 폴더 = 뉴런, 경로 = 문장, 카운터 = 가중치 |
| 🧬 | [뇌 영역](#뇌-영역) | 7개 영역, 우선순위, 호르몬 시스템 |
| ⚖️ | [거버넌스](#거버넌스) | 3-Tier 주입, bomb 서킷 브레이커, 하네스 |
| 🏗️ | [아키텍처](#아키텍처) | 자율 루프, CLI, MCP, 멀티에이전트 |
| 📊 | [벤치마크](#벤치마크) | 성능, 경쟁사 비교 |
| ⚠️ | [한계](#한계) | 안 되는 것에 대한 솔직한 이야기 |
| ❓ | [FAQ](#faq) | 예상 질문과 답 |
| 📜 | [Changelog](#changelog) | 버전 업데이트 이력 |

---

## 핵심 구조

> **Unix는 "Everything is a file"이라 했다. 우리는 말한다: Everything is folders.**

| 개념 | 생물학 | NeuronFS | OS 프리미티브 |
|------|--------|----------|--------------|
| 뉴런 | 세포체 | 디렉토리 | `mkdir` |
| 규칙 | 발화 패턴 | 전체 경로 | 경로 문자열 |
| 가중치 | 시냅스 강도 | 카운터 파일명 | `N.neuron` |
| 보상 | 도파민 | 보상 파일 | `dopamineN.neuron` |
| 차단 | 세포사멸 | `bomb.neuron` | `touch` |
| 수면 | 시냅스 정리 | `*.dormant` | `mv` |
| 연결 | 축삭 | `.axon` 파일 | 심링크 |
| 교차 참조 | Attention Residual | Axon Query-Key 매칭 | 선택적 집계 |

### Path = Sentence

경로가 곧 자연어 명령이 된다. 깊이가 구체성이다:

```
brain/cortex/NAS파일전송/                    → 카테고리
brain/cortex/NAS파일전송/禁Copy-Item_UNC비호환/  → 구체적 행동 강령
brain/cortex/NAS파일전송/robocopy_대용량/        → 세부 맥락
```

컴파일 결과: `cortex > NAS파일전송 > 禁Copy-Item UNC비호환`

### 한자 마이크로옵코드

`禁` (1글자) = "NEVER_DO" (8글자). 폴더명에 3~5배 더 많은 토큰 의미를 압축한다:

| 한자 | 의미 | 예시 |
|------|------|------|
| **禁** | 금지 | `禁fallback` |
| **必** | 필수 | `必KI자동참조` |
| **推** | 추천 | `推robocopy_대용량` |
| **警** | 경고 | `警DB삭제_확인필수` |

### 자율 진화망 (Auto-Evolution)

`.cursorrules`는 사람이 직접 편집을 강제받는 정적 파일이다. NeuronFS의 자율 진화 파이프라인은 다르다:

1. **auto-consolidate**: 폴더 파편화 해결. LLM(Groq 또는 로컬 모델)이 유사한 에러 폴더들을 분류하여 단일 뉴런으로 병합하고 기존 카운터를 승계.
2. **auto-neuronize**: 교정 로그(corrections)를 분석하여 반복을 방지하는 억제형(Contra) 규칙을 생성.
3. **auto-polarize**: 긍정형 "use_X" 규칙을 감지해 마이크로옵코드 기반의 강력한 억제형("禁X")으로 자동 전환 제안.

### Attention Residuals (교차 영역 지능)

[Kimi의 Attention Residuals 논문](https://arxiv.org/abs/2603.15031)에서 영감을 받아, `.axon` 연결을 통한 **선택적 교차 참조**를 구현:

- 각 영역의 TOP 뉴런에서 **쿼리 키워드** 생성
- 연결된 영역의 뉴런 경로와 **키 매칭** 수행
- 상위 3개 관련 뉴런이 `_rules.md`에 자동 노출
- 거버넌스 뉴런(禁/推)은 무조건 부스트

```
ego/_rules.md를 읽으면 자동 표시:
## 🔗 Axon 참조 (Attention Residuals)
- tools > 推: precise tool usage (c:65)    ← cortex에서
- tools > 절대 금지: ls usage (c:57)       ← cortex에서
- ops > 절대 금지: general commands (c:48)  ← cortex에서
```

### 자율 하네스 사이클 (Autonomous Harness Cycle)

AI 25회 상호작용마다, 하네스 엔진(Node.js 사이드카)이 자동으로:

1. 교정 로그의 **실패 패턴 분석**
2. Groq LLM을 통한 **禁(금지)/推(추천) 뉴런 자동 생성**
3. 관련 영역 간 **`.axon` 교차 링크 생성**
4. 해당 실수는 **구조적으로 재발 불가능** — 프롬프트가 아니라 시스템이 막는다

---

## 뇌 영역

7개 뇌 영역이 Brooks의 Subsumption Architecture로 계층화된다. **낮은 P(우선순위)가 높은 P의 명령을 항상 물리적으로 억제한다.**

```
brainstem(P0) > limbic(P1) > hippocampus(P2) > sensors(P3) > cortex(P4) > ego(P5) > prefrontal(P6)
```

| 뇌 영역 | 우선순위 | 역할 | 예시 |
|---------|---------|------|------|
| **brainstem** | P0 | 절대 불변 원칙 | `禁fallback`, `禁SSOT중복` |
| **limbic** | P1 | 감정 필터, 호르몬 | `도파민_보상`, `아드레날린_비상` |
| **hippocampus** | P2 | 기억, 세션 복원 | `에러패턴`, `KI_자동참조` |
| **sensors** | P3 | 환경 제약 | `NAS/禁Copy`, `디자인/sandstone` |
| **cortex** | P4 | 지식, 기술 (최다) | `react/hooks`, `backend/supabase` |
| **ego** | P5 | 톤, 성향 | `전문가_간결`, `한국어_의심하고검증` |
| **prefrontal** | P6 | 목표, 프로젝트 | `current_sprint`, `long_term` |

---

## 거버넌스

### 서킷 브레이커 (Circuit Breaker)

| bomb 위치 | 결과 |
|-----------|------|
| brainstem (P0) | **전체 뇌 중단**. GEMINI.md 자체가 비워지며 AI를 사실상 침묵시킴. |
| cortex (P4) | brainstem~sensors까지만 출력. 해당 기술(코딩) 영역만 완벽 격리/차단. |

bomb.neuron은 글자로 "하지 마"라고 비는(Begging) 것이 아니라, **해당 영역의 프롬프트 렌더링 자체를 멈추는 비상 정지 버튼**이다.
해제: `rm brain_v4/.../bomb.neuron` — 파일 삭제 1개.

### 하네스 (Harness)

강력한 자동 검증 스크립트가 로컬 환경에서 돌아간다:
- brainstem 불변성 확인 및 훼손 시 파기
- axon 무결성 검사
- 파괴적 명령(통합) 전 `Pre-Git Lock` 스냅샷 (데이터 복원 강제)
- 전역 무한 락 방어 캡슐 모듈 탑재 (`SafeExec` 30초 데드락 타임아웃)
- **자율 하네스 사이클**: 25회 상호작용마다 Groq 기반 禁/推 뉴런 자동 생성
- **Attention Residuals**: `.axon` 교차 링크로 영역 간 선택적 참조

---

## 아키텍처

> **⚠️ 전사 아키텍처는 NeuronFS의 규칙에 따라 '뇌(폴더) 구조'로 자체 흡수(Subsumption)되었습니다.**
> 시스템 컴포넌트(Go, 다중 에이전트 브릿지, NAS 부팅 시퀀스 등)의 통합 상관관계는 루트 마크다운에 존재하지 않으며, **`brain_v4/cortex/neuronfs/아키텍처/`** 하위의 물리적 뉴런(폴더) 구조로 완전 통합(SSOT)되었습니다.
> 이로 통해 시스템은 자신의 아키텍처 구조 자체를 스스로 인지하고 실시간 진화합니다.

### 자율 루프

```
사용자 교정 → corrections 텍스트 생성 → neuronize 모터 감지 → mkdir (오류 방지 뉴런/폴더 강제 생성)
                                                           ↓
                                                .cursorrules 자동 재컴파일
                                                           ↓
                                                다음 질문부터 AI가 해당 실수를 절대 반복하지 않음
```

### CLI 인터페이스

```bash
neuronfs <brain> --emit <target>   # 프롬프트 컴파일 (gemini/cursor/claude/all)
neuronfs <brain> --consolidate     # Llama 3 70B 병합 로직 가동
neuronfs <brain> --api             # 대시보드 (localhost:9090)
neuronfs <brain> --watch           # 파일 감시 + 실시간 반영
neuronfs <brain> --grow <path>     # 뉴런 생성
neuronfs <brain> --fire <path>     # 가중치 카운터 +1 증가
```

### 왜 Go인가?

단일 실행 파일 바이너리(Single Binary). 어떤 외부 의존성(Node_modules, Python venv)도 없다. 다운로드 받아서 아무 폴더에나 놓으면 즉시 시스템 파일 트리를 감시(fsnotify)하고 터미널에서 동작한다. 극강의 이식성과 영속성.

---

## 벤치마크

| 지표 | 수치 |
|------|------|
| 1,000개 폴더 렌더링 스캔 속도 | 271ms (1초 미만) |
| 규칙 폴더 추가 | OS 기본(`mkdir`) 사용, 0ms |
| 로컬 디스크 사용량 | 4.3MB (순수 텍스트/폴더 구조) |
| 유지보수/런타임 통신 비용 | **$0** (프롬프트 관리/저장 비용 전면 무료) |
| brainstem (절대 원칙) 준수율 | **94.9%** (353회 주입 중 위반 18회) |

### 경쟁사 비교

| | `.cursorrules` 하드코딩 | 벡터 DB (RAG) | **NeuronFS (CLI)** |
|---|---|---|---|
| 지식 구조 1000줄 초과 시 | 토큰 폭파, 유지보수 헬 | ✅ 검색 속도 빠름 | **✅ OS 폴더 트리 기반 분산** |
| 인프라 의존 비용 | 무료 | 서버 대여 ($70/월) | **무료 ($0)** |
| 에디터 툴 교체 시 | 호환 안 됨 (다시 작성) | DB 마이그레이션 필요 | **폴더만 복사하면 끝** |
| 자가 성장 통제 | 불가 | 블랙박스 (어떤 벡터인지 안 보임) | **가시적 폴더 (교정 시 mkdir 자동화)** |
| 절대 원칙(물리 차단) | 프롬프트로 구걸해야 함 | 제한적 | **✅ 서킷 브레이커 (bomb.neuron)** |

---


## 철학과 온톨로지 (Palantir AIP)

왜 폴더일까요? Palantir(팔란티어)의 AIP가 폭발적인 성과를 낸 이유는 가장 똑똑한 AI를 써서가 아니라, 기업의 데이터와 행동을 하나의 **온톨로지(Ontology, 실재의 구조화)**로 묶어냈기 때문입니다.

NeuronFS는 이 거대한 철학을 로컬 파일시스템으로 가져옵니다. AI에게 1,000줄짜리 텍스트를 던져주고 "잘 기억해"라고 구걸하는 대신, 당신의 비즈니스 로직을 물리적 폴더 경로(cortex/frontend/禁console_log)로 박제합니다. 
AI의 환각(Hallucination) 자체를 OS가 물리적으로 막을 수는 없습니다. 하지만 OS 레벨 권한 분리를 통해 프롬프트 생성 규칙이 무너지거나 훼손되는 일만큼은 확실히 하드 락(Hard Lock)을 걸어 방어합니다.

## 하이브리드 거버넌스 한계 극복 (Hybrid Memory Architecture)

**"우리는 RAG를 거부하는 것이 아니라, RAG의 환각(Hallucination)을 통제하는 L1 거버넌스 캐시입니다."**

NeuronFS는 대규모 분산 환경(MSA)이나 범용 벡터 검색 시스템과 대척점에 있지 않습니다. 오히려 완벽한 상호 보완재(Hybrid)로 작동하도록 아키텍처가 의도적으로 분리되었습니다.

*   **Tier 1 & 2 (NeuronFS 결정론적 지배):** 절대 불변 규칙(`brainstem`), 워크플로 제약(`sensors`). "데이터베이스 강제 백업", "평문 토큰 금지"와 같은 핵심 거버넌스는 확률(유사도 80%)에 의존하면 안 됩니다. 100% 동일한 경로를 갖는 디렉토리의 하드 락(Hard Lock)이 필요합니다. 지연시간(Latency) 제로.
*   **Tier 3 (Vector DB / RAG 위임):** 방대한 API 규격, 수년간 누적된 에러 로그(`hippocampus`). 이처럼 모호하고 거대한 컨텍스트는 수천 개의 폴더로 쪼개는 등 오버엔지니어링 하지 않고, LlamaIndex 등 기존 RAG 파이프라인과 프레임워크에 위임하여 유연성을 확보합니다.

즉, AI 에이전트가 무턱대고 거대한 벡터 DB를 횡단하기 전에, **NeuronFS(Tier 1,2)가 우선 개입하여 '절대 피해야 할 명령(禁)'을 가드레일로 먼저 깔아주는 것**이 완성된 엔터프라이즈 하이브리드 확장 모델입니다. OS 폴더가 L1 명령어 캐시, RAG가 L2 메인 램 역할을 수행합니다.

---

## FAQ

**Q: "아니 그래서, 결국 마지막엔 시스템 프롬프트(텍스트)로 합쳐넣는 거 아니야? 그냥 텍스트 파일이나 노션에 규칙 적어두는 거랑 뭐가 달라?"**  
**A:** 다릅니다. 1,000줄짜리 텍스트 스파게티 속에서 규칙 하나를 찾고, 우선순위를 조정하고, 삭제하는 행위. 그것은 사람을 미치게 만듭니다. 우리는 **"문자열 공간"에서 "운영체제 물리 폴더 공간"으로 지식의 차원을 격상**시킨 것입니다.
텍스트 안에서 `!!중요!!`를 수십 번 적는(프롬프트로 구걸하는) 행위 대신, NeuronFS는 **권한 분리(Cascade 계층)**와 **접근 금지선(bomb.neuron 물리 차단)**을 제공합니다. 하나가 발동하면 하위 계층 문자열 자체가 아예 렌더링 생태계에서 '차단'되는 하드 락 아키텍처입니다.

**Q: "뉴런(폴더)이 1,000개를 넘어가면 프롬프트 토큰 터지는 거 아닌가?"**  
**A:** 세 가지 방어망이 있습니다. ① 3-Tier 온디맨드 렌더링 (대화 흐름에 맞는 폴더만 동적으로 묶어줌) ② 30일 접속 없는 폴더는 휴면(Dormant) 수면 상태 전환. ③ `--consolidate` 병합. 중복된 폴더들을 Llama 3나 로컬 모델이 단일 슈퍼 뉴런으로 깔끔하게 묶습니다.

**Q: "MBTI 에이전트(ENTP, ISTJ)로 나눈 건 장난인가?"**  
**A:** "너는 QA 역할이야"라는 전형적인 룰 기반(Role-based) 프롬프트는 에이전트를 상자 안에 가둡니다. 우리는 역할을 주지 않고, "어떤 문제든 공격적으로 해체할지(ENTP) 보수적으로 접근할지(ISTJ)"라는 **성향(Personality)**을 뇌의 `ego` 폴더에 박아 넣었을 때 AI의 성능이 훨씬 더 유기적으로 변화한다는 것을 증명했습니다.

---

## Changelog

**v4.4 (2026-04-04)** — **Attention Residuals** 교차 참조 구현 (`.axon` 기반). 자율 하네스 사이클 (Groq 禁/推 자동 생성). UTF-8 BOM 파싱 버그 수정. 433뉴런, 6 axon.
**v4.3 (2026-04-02)** — 자율 엔진 Llama 3 전면 포팅 ($0 비용) 및 SafeExec 하드 락 이식.
**v4.2 (2026-03-31)** — 자율 진화(Auto-Evolution) 파이프라인 완성. Groq 교정 로그 분석 / 한자 마이크로옵코드 최적화.

전체 변경 이력 확인: [LIFECYCLE.md](LIFECYCLE.md)

---

MIT License · Copyright (c) 2026

동의하면 Star. [아니면 Issue 제기하기.](../../issues)


<!--
Easter Egg for the code divers:
Hey 666, easy - only the Word stands as absolute truth (777). 
This? It's just a well-organized folder built by someone who wanted to vibe-code without going insane.
-->
