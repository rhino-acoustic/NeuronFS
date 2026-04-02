<p align="center">
  <img src="https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat-square&logo=go" />
  <img src="https://img.shields.io/badge/Infra-$0-brightgreen?style=flat-square" />
  <img src="https://img.shields.io/badge/Neurons-340+-blue?style=flat-square" />
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
### *파일시스템 네이티브 계층형 규칙 메모리 — 무의존성 프롬프트 컴파일러*

> *"프롬프트로 구걸하지 마. 파이프라인을 설계해."*
>
> AI가 "console.log 쓰지 마"라는 지시를 9번 어겼다. 10번째에 `mkdir brain/cortex/frontend/coding/禁console_log`를 만들었다. 폴더 이름이 물리적 규칙으로 시스템 프롬프트에 강제 삽입되었다. 카운터(가중치)가 17이 되었다. AI가 두 번 다시 해당 실수를 반복하지 않는다.
> 
> 고급 모델은 구조를 만드는 데 쓴다. 최종장에서는 AI에 대한 의존을 '트랜지스터'급으로 최소화하여 통제권을 되찾는 것이 NeuronFS의 목적이다.

---

## 요약 (TL;DR)

**`mkdir`이 시스템 프롬프트를 대체한다.** 폴더가 뉴런이고, 경로가 문장이며, 파일이 시냅스 가중치다.

```bash
# 규칙 생성 = 폴더 생성
mkdir -p brain/brainstem/禁fallback
touch brain/brainstem/禁fallback/1.neuron

# 컴파일 = 시스템 프롬프트 자동 생성 (Cursor, Windsurf, Claude Desktop 등)
neuronfs ./brain --emit cursor   # → .cursorrules
neuronfs ./brain --emit claude   # → CLAUDE.md
neuronfs ./brain --emit all      # → 모든 AI 포맷 동시 출력
```

| 기존 방식 | NeuronFS |
|-----------|----------|
| 1000줄 프롬프트 스파게티 편집 | `mkdir` 로 물리적 단절 관리 |
| 벡터 DB $70/월 | **$0** (로컬 폴더 = DB) |
| AI 도구 교체 시 마이그레이션 | `cp -r brain/` — 1초 복사 끝 |
| 규칙 위반 → 인간이 스트레스 받음 | `bomb.neuron` → 위반 시 물리 차단 |
| 규칙을 사람이 수동 관리 | 교정 시 자동 뉴런 폴더 생성 |

---

## 설치 (The One-Liner Quickstart)

오픈소스 단일 바이너리 Go 엔진. 외부 의존성(Dependencies) 패키지 제로.

```bash
# Mac / Linux
curl -sL https://neuronfs.com/install | bash

# Windows (PowerShell)
iwr https://neuronfs.com/install.ps1 -useb | iex

# 나만의 오프라인 뇌 초기화 (비어있는 7개 영역 기본 스캐폴딩 생성)
# ※ 대화형 프롬프트에서 [2]번 Master's Brain 옵션 선택 시 프리미엄 거버넌스 뼈대 복사 가능
neuronfs --init ./my_brain        

export GROQ_API_KEY="gsk_..."      # Llama3 70B 기반 자율 폴더 통합 옵션용 (로컬 Ollama 연결 지원!)

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

---

## 아키텍처

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

## 한계 (Limitations)

- **AI 제어 100% 보장 불가:** `brainstem`의 무결성은 OS 폴더 레벨에서 차단되지만, 생성형 AI 자체가 할루시네이션(환각)을 일으켜 규칙을 이탈하는 것 자체를 완전히 막을 수는 없습니다.
- **시맨틱 벡터 검색 미지원:** 폴더명(경로) 매칭에 극대화되어 있어, 애매모호한 자연어 문장 기반의 벡터 라우팅(RAG) 검색은 의도적으로 제외되었습니다.

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

**v4.3 (2026-04-02)** — 자율 엔진 Llama 3 전면 포팅 ($0 비용) 및 SafeExec 하드 락 이식.
**v4.2 (2026-03-31)** — 자율 진화(Auto-Evolution) 파이프라인 완성. Groq 교정 로그 분석 / 한자 마이크로옵코드 최적화.

전체 변경 이력 확인: [LIFECYCLE.md](LIFECYCLE.md)

---

MIT License · Copyright (c) 2026

동의하면 Star. [아니면 Issue 제기하기.](../../issues)
