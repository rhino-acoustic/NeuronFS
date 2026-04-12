# Episode 16. P0 Brainstem Always Overrides P4 Cortex
## "7계층 뇌" — 인간의 뇌를 OS 디렉토리 계층으로 복제하다

> **Language Note / 언어 안내**
>
> **[ENG]** This document details the Subsumption Cascade — NeuronFS's 7-layer brain architecture where lower priority layers (P0 brainstem) always physically override higher ones (P4 cortex), inspired by Rodney Brooks' 1986 robotics paradigm. It also explains the self-evolution loop and 3-tier emit pipeline.
>
> **[KOR]** 로드니 브룩스의 1986년 포섭 아키텍처에서 영감을 받은 7계층 뇌 구조를 설명합니다. 하위 계층(P0 뇌간)이 상위 계층(P4 대뇌피질)을 항상 물리적으로 덮어쓰는 Subsumption Cascade와 자율 진화 루프를 다룹니다.

---

### 킥

> **"brainstem이 cortex를 항상 이겨야 해. 뇌간이 대뇌피질을 이기는 거야."**

사람이 아무리 논리적으로 생각해도(대뇌피질), 뜨거운 냄비에 손이 닿으면 **반사적으로 뗀다(뇌간)**. 이유를 묻지 않는다. 뇌간은 피질보다 빠르고, 피질의 판단을 무시한다.

NeuronFS는 이 생물학적 원리를 **폴더 계층의 우선순위**로 구현한다.

---

### 7계층 Subsumption Cascade

```
brainstem (P0) ←→ limbic (P1) ←→ hippocampus (P2) ←→ sensors (P3)
     ↕                                                      ↕
cortex (P4) ←→ ego (P5) ←→ prefrontal (P6)

규칙: 낮은 P가 높은 P를 항상 물리적으로 덮어쓴다.
bomb = 전체 즉시 정지.
```

| 계층 | 인간의 뇌 | NeuronFS | 예시 | 발화 조건 |
|---|---|---|---|---|
| P0 | 뇌간 (생존 반사) | brainstem/ | 禁/보안위반 | 3회 반복 실패 → bomb |
| P1 | 변연계 (감정) | limbic/ | 분노 감지 → 톤 조절 | 감정 키워드 탐지 |
| P2 | 해마 (기억) | hippocampus/ | 에러 패턴 기록 | corrections.jsonl 기록 |
| P3 | 감각기관 | sensors/ | NAS 경로, 브랜드 | 환경 변수 변경 시 |
| P4 | 대뇌피질 (지식) | cortex/ | React 규칙, DB 패턴 | 코드 작성 중 |
| P5 | 자아 | ego/ | 말투, 포맷 | 모든 응답 |
| P6 | 전두엽 (목표) | prefrontal/ | 스프린트 계획 | 장기 프로젝트 |

### 충돌 시뮬레이션: P4와 P0가 부딪히면?

```
상황: AI가 코드를 작성하다가 API 키를 하드코딩하려 함

[P4 cortex/dev/edge_functions/]  → "이 함수에 API 키 넣어야겠다"
          ↓ 충돌
[P0 brainstem/]                  → "禁/하드코딩.neuron 발견"
          ↓ 결과
P0 WIN. P4 명령 무효화. AI: "환경변수로 분리합니다."
```

**AI가 아무리 똑똑한 코드를 짜려 해도(P4), 보안 위반이면 뇌간(P0)이 전기를 끊는다.** 논리보다 생존이 우선. 이것이 포섭(Subsumption)이다.

---

### 3-Tier Emit Pipeline (3단계 방출)

뇌의 규칙이 IDE에 도达하는 3가지 경로:

```
[단계 1: stdout] neuronfs --emit
→ 터미널에 규칙 미리보기만 출력 (파일 변경 없음)

[단계 2: IDE 주입] neuronfs --emit all
→ _rules.md 자동 생성 + IDE가 감지

[단계 3: 전체 주입] neuronfs --inject  
→ GEMINI.md + _index.md + _rules.md 전체 주입
→ AI 에이전트가 다음 응답부터 즉시 새 규칙 적용
```

### 자율 진화 (evolve.go) — 뇌가 스스로 자란다

```
[실시간] 대화 중 processChunk()
 → 유의미한 패턴 발견 → _signals/*.json에 기록만

[30분마다] 자동 스케줄러
 → neuronfs --evolve 실행
 → Groq가 축적된 시그널을 분석
 → "이 패턴은 뉴런으로 승격할 가치가 있다" 판단

[승격 시]
 → 새 .neuron 파일 자동 생성
 → 텔레그램 알림: "NEURON EVOLVED: cortex/dev/새_패턴"
 → 다음 emit에서 자동으로 IDE에 반영
```

> *"AI가 대화하면서 발견한 패턴이 뉴런으로 승격되고, 그 뉴런이 다음 대화를 통제한다. 뇌가 경험으로부터 자란다. 이것이 corrections.jsonl → evolve → neuron의 자가 학습 루프다."*

---

### 로드니 브룩스(1986)와의 비교

| | 브룩스의 포섭 구조 | NeuronFS |
|---|---|---|
| 대상 | 물리적 로봇 | AI 에이전트 |
| P0 | 장애물 피하기 (하드웨어) | 禁/보안위반 (폴더) |
| P1 | 배회하기 | 감정 필터 |
| P4 | 목적지 가기 | 코드 지식 |
| 매체 | 전자 회로 | OS 파일 시스템 |
| 핵심 | 멍청한 반사신경의 계층적 조합 = 지능 | 0바이트 폴더의 계층적 조합 = AI 거버넌스 |

> *"브룩스는 1986년에 증명했다: '똑똑한 중앙 두뇌' 없이도 계층적 반사신경만으로 로봇이 지능적으로 행동할 수 있다. NeuronFS는 2026년에 같은 원리를 증명한다: '똑똑한 프롬프트' 없이도 계층적 폴더 구조만으로 AI가 안전하게 행동할 수 있다."*

---

[Back to Act 3](Act-3) | [Ep.15](Episode-15-Brainwallet-and-Zero-Trace-Encryption) | [Ep.17](Episode-17-First-Markets-Law-and-Medicine)
