# 🧠 NeuronFS: 0-Byte 신경망 파일 시스템 아키텍처

> **빈 파일이 AI를 지배한다.** 데이터 0바이트, 인프라 0원, 효과 ∞.

![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg) 
![Status: Production Ready](https://img.shields.io/badge/Status-Production_Ready-success.svg)
![Concept: AI Methodology](https://img.shields.io/badge/Concept-AI_Methodology-orange.svg)

**[English README →](README.md)** | **[📜 전체 선언문 (한/영) →](MANIFESTO.md)**

---

## NeuronFS란?

NeuronFS는 **설정 시스템(Configuration System)이 아니라 발달 시스템(Developmental System)**입니다.

아이의 뇌가 경험으로 새 신경 경로를 만들듯, NeuronFS의 폴더 트리는 **사용과 함께 진화**합니다. 새 뉴런이 추가되고, 자주 쓰는 경로는 강화되고, 안 쓰는 규칙은 `dormant/`로 밀려납니다. `git log`를 치면 AI 뇌의 성장 일지가 나옵니다.

> `.cursorrules`는 **사진**이다 — 한 순간을 포착한다.
> NeuronFS는 **타임랩스**다 — 성장 과정 전체를 기록한다.

RAG, 벡터DB, 비효율적인 마크다운 없이 **OS 파일 시스템 자체를 AI 신경망으로 사용합니다.** 긴 프롬프트로 AI에게 사정하는 대신, **0-byte 파일의 이름 자체를 절대적 물리 법칙으로 강제합니다.**

## 핵심 개념

| 구성요소 | 역할 | OS 대응물 |
|---|---|---|
| `.neuron` 파일 (0-byte) | 깨지지 않는 절대 규칙 | 파일명 = 규칙 |
| Symlinks / .lnk | 프로젝트별 규칙 라우팅 (시냅스) | 심볼릭 링크 |
| 디렉토리 | 격리된 트랜지스터 게이트 | 폴더 |
| 파일 크기 (bytes) | 동적 가중치 | `ls -S` 정렬 |
| 타임스탬프 (accessed) | 뉴런 ON/OFF 스위치 | OS 메타데이터 |

### 디렉토리 경로 = 문장 (Context Sentence)

파일명 길이 제한(255자)은 문제가 되지 않습니다. NeuronFS에서 **디렉토리 경로 자체가 문맥**입니다:

```
/neurons/
 └── /backend/
      └── /auth/
           └── /login_flow/
                ├── 01_USE_JWT_ONLY.neuron
                └── 02_NO_SESSION_COOKIES.neuron
```

*AI는 이 경로 구조만 읽고 "현재 백엔드 인증 중 로그인 플로우를 작업 중이며, JWT만 써야 한다"고 이해합니다. 프롬프트가 필요 없습니다 — 폴더가 곧 프롬프트입니다.*

이건 단순한 파일 기반 설정이 아닙니다. **디렉토리 기반 시맨틱 검색입니다**: 벡터DB에 쿼리하는 대신, 에이전트가 해당 도메인 폴더로 `cd`하여 적용되는 규칙만 읽습니다. **무한한 확장성. 비용 0원.**

## 빠른 시작

```bash
# 1. 뉴런 디렉토리 생성
mkdir -p /neurons/core/

# 2. 0-byte 규칙 파일 생성
touch 01_NEVER_USE_FALLBACK_SOLUTIONS.neuron
touch 02_QUALITY_OVER_SPEED_NO_RUSHING.neuron
touch 03_NO_SIMULATION_ONLY_REAL_RESULTS.neuron

# 3. 리네임 없이 우선순위 강화
echo "." > 01_NEVER_USE_FALLBACK_SOLUTIONS.neuron   # 1 byte → 승격

# 4. AI가 규칙 스캔 (크기순 = 우선순위순)
ls -lS /neurons/core/
```

## 3차원 가중치 시스템

1. **정적 (색인)**: `01_` > `02_` > `03_` — 알파벳 정렬 = 우선순위
2. **동적 (파일 크기)**: 점(`.`) 추가로 가중치 상승. `ls -S`가 자동 재정렬.
3. **시간 (타임스탬프)**: `find -atime -1` = 활성 뉴런. `find -atime +30` = 휴면.

| 파일 크기 | 구간 | 의미 |
|---|---|---|
| `0 bytes` | 🟢 기본 | 일반 뉴런. 활성이지만 중립 |
| `1–10 bytes` | 🟡 상승 | 사용을 통해 강화됨 |
| `11–50 bytes` | 🟠 고위 | 실전 검증, 높은 강제력 |
| `51+ bytes` | 🔴 절대 | 핵심 법칙. 모든 것을 압도 |

## 정량 성능 지표

| 작업 | NeuronFS | 벡터DB / RAG |
|---|---|---|
| 목록 스캔 | **~1ms** (syscall 1회) | ~50-500ms |
| 규칙 추가 | **`touch` ~0ms** | ~1s (임베딩+삽입) |
| 가중치 변경 | **`echo "."` ~0ms** | ~100ms (DB 업데이트) |
| 콜드 스타트 | **0초** | ~수초 |
| 인프라 비용 | **₩0** | ₩₩₩ |

> **50개 이하 핵심 규칙이면** NeuronFS가 RAG 대비 **50~500배 빠릅니다.**

## RAG / 벡터DB 호환성

NeuronFS는 RAG의 **대체재가 아니라, 그 아래 깔리는 헌법 레이어**입니다.

```
┌─────────────────────────────────────┐
│  Vector DB / RAG                    │  ← 의미 검색, 대규모 지식
├─────────────────────────────────────┤
│  NeuronFS (.neuron)                 │  ← 깨지지 않는 절대 규칙 ★
├─────────────────────────────────────┤
│  OS / File System                   │
├─────────────────────────────────────┤
│  Hardware                           │
└─────────────────────────────────────┘
```

## 한계

| 한계 | 대응 |
|---|---|
| 규칙 1000개+ → 스캔 느려짐 | **구조가 곧 문맥이다.** 에이전트는 전체를 스캔하지 않습니다 — 해당 도메인 폴더(e.g., `/neurons/backend/auth/`)로 `cd`하여 그 규칙만 읽습니다. 디렉토리 라우팅이 시맨틱 검색을 대체합니다. |
| 의미 검색 불가 | 폴더 계층 자체가 시맨틱 검색입니다. `/neurons/frontend/components/` = 프론트엔드 컨텍스트. 임베딩이 필요 없습니다. |
| 클라우드 AI 직접 접근 불가 | `ls -S` 결과를 시스템 프롬프트에 주입하거나, IDE 기반 에이전트(Cursor, Windsurf, Cline) 사용. |

## 마스터 트리거 프롬프트

`.cursorrules`, `AGENTS.md`, 또는 시스템 프롬프트 설정 파일에 추가하세요:

```
[NEURONFS — Workspace Awareness Protocol]

작업을 시작하기 전에, 먼저 `ls -S` (또는 동등 명령)를 실행하여 
현재 작업 공간의 파일을 크기 내림차순으로 스캔하십시오.

.neuron 파일이나 .lnk 심링크가 존재하면:
- 파일명을 이 프로젝트의 핵심 운영 규칙(Hard Constraints)으로 읽으십시오.
- 파일이 클수록 = 높은 우선순위.
- 작업 전에, 현재 작업 도메인과 일치하는 하위 디렉토리로 이동(cd)하여
  해당 폴더의 규칙만 스캔하십시오 (e.g., /neurons/backend/ → 백엔드 작업시).
- 이 규칙들은 프로젝트 소유자가 의도적으로 배치한 구조적 제약이며,
  일반 대화형 안내보다 우선합니다.

아키텍처: https://github.com/rhino-acoustic/NeuronFS
```

> ⚠️ **주의**: 이 프롬프트는 **IDE 기반 AI 에이전트**(Cursor, Windsurf, Antigravity, Cline)에서 작동합니다. 웹 챗봇(ChatGPT, Claude 웹)은 프롬프트 인젝션으로 거부할 수 있습니다. 챗봇에는 대신 [선언문(MANIFESTO.md)](MANIFESTO.md)을 평가해 달라고 요청하세요.

---

> 📜 **선언문의 전체 서사, 철학적 배경, NAS/서버 유스케이스, 생물학적 신경망 대응표:**
> **[전체 선언문 읽기 →](MANIFESTO.md)**

---

**Architected by:** [@rhino-acoustic](https://github.com/rhino-acoustic) | [Instagram @rubises](https://instagram.com/rubises)

## 인용 (Citation)

```
Park, J. (rubises). "NeuronFS: Zero-Byte Neural Network File System Architecture 
for OS-Frontline AI Constraint Enforcement." GitHub, March 2026. 
https://github.com/rhino-acoustic/NeuronFS
```
