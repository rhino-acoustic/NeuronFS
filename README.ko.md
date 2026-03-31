<p align="center">
  <img src="https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat-square&logo=go" />
  <img src="https://img.shields.io/badge/Infra-$0-brightgreen?style=flat-square" />
  <img src="https://img.shields.io/badge/Neurons-293-blue?style=flat-square" />
  <img src="https://img.shields.io/badge/Zero_Dependencies-black?style=flat-square" />
  <img src="https://img.shields.io/badge/MIT-green?style=flat-square" />
</p>

<p align="center">
  <img src="docs/dashboard.png" alt="NeuronFS 대시보드 — 3D 뇌 시각화" width="800" />
  <br/>
  <a href="https://dashboarddeploy-six.vercel.app/"><strong>🔥 3D 대시보드 라이브 데모</strong></a>
</p>

<p align="center"><a href="README.ko.md">🇰🇷 한국어</a> · <a href="README.md">🇺🇸 English</a> · <a href="MANIFESTO.md">📜 매니페스토</a> · <a href="LIFECYCLE.md">🧬 생애주기</a></p>

> **⚠️ v4.1 (2026-03-31) — 범용 사용을 위한 개선 진행중**
>
> **수정 완료:** 뉴런 마이그레이션 (307→293), 개인정보 전면 제거, supervisor v2.0 (삭제된 mjs 스크립트 제거), heartbeat/idle engine 문서화, 와치독 생애주기 감사
>
> **진행중:** OS 자동시작 등록 (L0), 전사 청크 자동 뉴런화, 영/한 중복 자동 병합, PII git-hook 스캐너, 빈 폴더 격리
>
> **Breaking:** `brain_v4/`가 git에서 제외됨 — 사용자는 `neuronfs --init`로 자체 뇌 생성 필요. Supervisor가 더 이상 Node.js 스크립트에 의존하지 않음.
>
> 전체 변경 이력: [LIFECYCLE.md](LIFECYCLE.md) · [LIFECYCLE_EN.md](LIFECYCLE_EN.md)

# 🧠 NeuronFS
### *파일시스템 네이티브 계층형 규칙 메모리 — 에이전트 프롬프트 컴파일러*

---

## 요약 (TL;DR)

**`mkdir`이 시스템 프롬프트를 대체한다.** 폴더가 뉴런이고, 경로가 문장이며, 카운터 파일이 시냅스 가중치다.

```bash
# 규칙 생성 = 폴더 생성
mkdir -p brain/brainstem/禁fallback
touch brain/brainstem/禁fallback/1.neuron

# 컴파일 = 시스템 프롬프트 자동 생성
neuronfs ./brain --emit cursor   # → .cursorrules
neuronfs ./brain --emit claude   # → CLAUDE.md
neuronfs ./brain --emit all      # → 모든 AI 포맷 동시 출력
```

| 기존 방식 | NeuronFS |
|-----------|----------|
| 1000줄 프롬프트 직접 편집 | `mkdir` 1개 |
| 벡터 DB $70/월 | **$0** (폴더 = DB) |
| AI 교체 시 마이그레이션 | `cp -r brain/` — 1초 |
| 규칙 위반 → 희망사항 | `bomb.neuron` → **물리 차단** |
| 규칙을 사람이 수동 관리 | 교정 → 자동 뉴런 생성 |

### 30초 시작

```bash
git clone https://github.com/rhino-acoustic/NeuronFS.git
cd NeuronFS/runtime && go build -o ../neuronfs .

./neuronfs ./brain_v4             # 진단 스캔
./neuronfs ./brain_v4 --emit all  # 프롬프트 컴파일
./neuronfs ./brain_v4 --api       # 대시보드 (localhost:9090)
./neuronfs ./brain_v4 --mcp       # MCP 서버 (stdio)
```

293 뉴런. 2026년 1월부터 매일 실전 운영 중. 단일 Go 바이너리. 의존성 제로.

---

## 목차

| | 섹션 | 내용 |
|---|---|---|
| 💡 | [핵심 구조](#핵심-구조) | 폴더 = 뉴런, 경로 = 문장, 카운터 = 가중치 |
| 🧬 | [뇌 영역](#뇌-영역) | 7개 영역, 우선순위, 호르몬 시스템 |
| ⚖️ | [거버넌스](#거버넌스) | 3-Tier 주입, bomb 서킷 브레이커, 하네스 |
| 🧬 | [뉴런 생애주기](#뉴런-생애주기) | 탄생 → 강화 → 수면 → 소멸 |
| 🏗️ | [아키텍처](#아키텍처) | 자율 루프, CLI, MCP, 멀티에이전트 |
| 📊 | [벤치마크](#벤치마크) | 성능, 경쟁사 비교 |
| ⚠️ | [한계](#한계) | 안 되는 것에 대한 솔직한 이야기 |
| ❓ | [FAQ](#faq) | 예상 질문과 답 |
| 📖 | [이야기](#이야기) | 왜 만들었는가 |

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

`禁` (1글자) = "NEVER_DO" (8글자). 폴더명에 3~5배 더 많은 의미를 압축한다:

| 한자 | 의미 | 예시 |
|------|------|------|
| **禁** | 금지 | `禁fallback` |
| **必** | 필수 | `必KI자동참조` |
| **推** | 추천 | `推robocopy_대용량` |
| **警** | 경고 | `警DB삭제_확인필수` |

### 자가 진화

`.cursorrules`는 사람이 직접 편집하는 정적 파일이다. NeuronFS는 다르다:

```
AI 실수 → 교정 → corrections.jsonl → mkdir (뉴런 자동 생성)
AI 잘함 → 칭찬 → dopamine.neuron (보상 신호)
같은 실수 3회 → bomb.neuron (해당 출력 전체 중단)
30일 미사용 → *.dormant (자동 수면)
     ↓
다음 세션 시스템 프롬프트에 자동 반영
```

---

## 뇌 영역

7개 뇌 영역이 Brooks의 Subsumption Architecture로 계층화된다. **낮은 P가 높은 P를 항상 억제한다.**

```
brainstem(P0) > limbic(P1) > hippocampus(P2) > sensors(P3) > cortex(P4) > ego(P5) > prefrontal(P6)
```

| 뇌 영역 | 우선순위 | 역할 | 예시 |
|---------|---------|------|------|
| **brainstem** | P0 | 절대 불변 원칙 | `禁fallback`, `禁SSOT중복` |
| **limbic** | P1 | 감정 필터, 호르몬 | `도파민_보상`, `아드레날린_비상` |
| **hippocampus** | P2 | 기억, 세션 복원 | `에러패턴`, `KI_자동참조` |
| **sensors** | P3 | 환경 제약 | `NAS/禁Copy-Item`, `디자인/sandstone` |
| **cortex** | P4 | 지식, 기술 (최다) | `frontend/react/hooks`, `backend/supabase` |
| **ego** | P5 | 톤, 성향 | `전문가_간결`, `한국어_네이티브` |
| **prefrontal** | P6 | 목표, 프로젝트 | `current_sprint`, `long_term_direction` |

### 호르몬 시스템

- **도파민** (`dopamineN.neuron`): 칭찬 → 해당 뉴런 긍정 가중치 상승
- **아드레날린** (`adrenaline.neuron`): "급해" 감지 → 하위 P가 상위 P를 억제
- **폭탄** (`bomb.neuron`): 3회 반복 실수 → 해당 영역 전체 비활성 (서킷 브레이커)

### 축삭 — 영역 간 배선

16개 `.axon` 파일이 7개 영역을 계층 네트워크로 연결한다:

```bash
brainstem/cascade_to_limbic.axon      → "limbic"     # bomb이면 감정 차단
sensors/cascade_to_cortex.axon        → "cortex"     # 환경 제약이 지식 필터링
cortex/shortcut_to_hippocampus.axon   → "hippocampus" # 학습 결과를 기억에 기록
```

---

## 거버넌스

### 3-Tier 주입

| Tier | 범위 | 토큰 | 시점 |
|------|------|------|------|
| **Tier 1** | brainstem + 핵심 규칙 TOP 5 | ~200 | 매 세션 자동 |
| **Tier 2** | 전체 영역 요약 (GEMINI.md) | ~800 | 시스템 프롬프트 |
| **Tier 3** | 특정 영역 `_rules.md` 전체 | ~2000 | 작업 감지 시 온디맨드 |

### 서킷 브레이커

| bomb 위치 | 결과 |
|-----------|------|
| brainstem (P0) | **전체 뇌 중단**. GEMINI.md 자체가 비게 됨 |
| cortex (P4) | brainstem~sensors까지만 출력. 코딩 영역 차단 |

bomb은 해당 규칙을 빼는 게 아니라 **해당 영역 전체 출력을 멈추는 비상 정지 버튼**이다.
해제: `rm brain_v4/.../bomb.neuron` — 파일 삭제 1개.

<p align="center">
  <img src="docs/bomb_alert.png" alt="bomb.neuron 발동 시 물리 경보" width="700" />
  <br/>
  <sub>bomb.neuron 감지 → 전체화면 빨간 플래시 + USB 사이렌. 문자 그대로의 하드 스탑.</sub>
</p>

### 하네스

15항목 자동 검증 스크립트가 매일 CI처럼 돌아간다:
- brainstem 불변성 확인
- axon 무결성 검사
- dormant 자동 정리
- 위반 감지 → 교정 루프 (직접 수정은 절대 안 함)

---

## 아키텍처

### 자율 루프

```
사용자 교정 → corrections.jsonl → neuronfs (fsnotify) → mkdir (뉴런 생성)
                                                          ↓
                                               _rules.md 재생성 → GEMINI.md 반영
                                                          ↓
                                               다음 세션에서 AI 행동 변경
```

### CLI

```bash
neuronfs <brain> --emit <target>   # 프롬프트 컴파일 (gemini/cursor/claude/copilot/all)
neuronfs <brain> --api             # 대시보드 (localhost:9090)
neuronfs <brain> --mcp             # MCP 서버 (stdio)
neuronfs <brain> --watch           # 파일 감시 + 자동 숙성
neuronfs <brain> --supervisor      # 프로세스 매니저
neuronfs <brain> --grow <path>     # 뉴런 생성
neuronfs <brain> --fire <path>     # 카운터 증가
neuronfs <brain> --decay           # 30일 미접촉 수면 처리
neuronfs <brain> --init <path>     # 새 뇌 초기화
neuronfs <brain> --snapshot        # Git 스냅샷
```

### MCP 서버 연동

```json
{
  "mcpServers": {
    "neuronfs": {
      "command": "/path/to/neuronfs",
      "args": ["/path/to/brain", "--mcp"]
    }
  }
}
```

### AI 도구별 연동

| AI 도구 | 연동 방식 | 난이도 |
|---------|----------|--------|
| Gemini CLI / Claude Code | GEMINI.md / CLAUDE.md 자동 로드 | ⭐ 바로 사용 |
| Cursor / Windsurf | .cursorrules 자동 로드 | ⭐ 바로 사용 |
| Claude Desktop | MCP 서버 (`--mcp`) | ⭐⭐ 설정 필요 |

### 왜 Go인가

단일 바이너리. 의존성 제로. `go build` → `neuronfs` 하나. 아무 머신에 복사하면 끝.
크로스 컴파일(`GOOS=linux`), fsnotify 네이티브, 고루틴으로 5개 이상 자식 프로세스 동시 관리.

### 멀티에이전트

모든 에이전트가 **같은 `brain/`을 공유**한다. 에이전트를 "역할"이 아니라 "성향"으로 나눈다:

| 에이전트 | 성향 | 접근 방식 |
|---------|------|----------|
| ANCHOR (ISTJ) | 보수적, 원칙 | "이 뉴런이 하네스 규칙 위반입니다" |
| FORGE (ENTP) | 공격적, 실험 | "이 뉴런을 3개로 쪼개면 더 효율적입니다" |
| MUSE (ENFP) | 창의적, 공감 | "이 뉴런 이름이 직관적이지 않습니다" |

역할 기반("너는 QA야")은 범위 밖에서 멈춘다. 성향 기반은 어떤 문제든 그 성향으로 접근한다.

---

## 뉴런 생애주기

> **뉴런은 태어나고, 강화되고, 교정되고, 잠들고, 죽는다.** 전체 생애주기는 [LIFECYCLE.md](LIFECYCLE.md) 참조.

```
탄생 → 강화/억제 → 성숙 → 수면/폭탄 → (소멸 또는 부활)
```

| 단계 | 트리거 | 결과 |
|------|--------|------|
| **탄생** | 교정, Memory Observer 감지, 직접 `mkdir` | 뉴런 폴더 + `1.neuron` 생성 |
| **강화** | 반복 교정 | 카운터 ↑ → 시스템 프롬프트 상위 배치 |
| **보상** | 칭찬 | `dopamine.neuron` → 긍정 가중치 |
| **수면** | 30일 미접촉 | `*.dormant` → 컴파일 제외 (삭제 아님) |
| **폭탄** | 동일 실수 3회 | `bomb.neuron` → 영역 전체 출력 중단 |
| **소멸** | dormant 90일 + 카운터 1 | 사용자 승인 후 삭제 |

### 교정 감사 스케줄

| 주기 | 항목 | 상태 |
|------|------|------|
| 유휴 시 / API | 중복 탐지 (Jaccard 유사도) | ✅ `deduplicateNeurons()` |
| `--decay` / 유휴 시 | dormant 수면 (30일 미접촉) | ✅ `runDecay()` |
| 매 스캔 | bomb 감지 + 물리 경보 | ✅ `triggerPhysicalHook()` |
| 주 1회 | 영/한 중복 병합 | 🔧 계획 |
| 매 commit | 개인정보 경로 스캔 | 🔧 계획 |
| **분기 1회** | 리전 정합성, 이름, 계층 전수 감사 | 🔧 수동 |

---

## 벤치마크

2026-03-29, 로컬 Windows 11 SSD 측정:

| 지표 | 수치 |
|------|------|
| 전체 스캔 (293 뉴런) | ~1ms |
| 규칙 추가 | `touch` <1ms |
| 1,000 뉴런 스트레스 | 271ms (3회 평균) |
| 디스크 사용량 | 4.3MB |
| 런타임 비용 | **$0** |
| brainstem 준수율 | **94.9%** (353회 fire 중 위반 18회) |

### 경쟁사 비교

| | .cursorrules | Mem0 | Letta | **NeuronFS** |
|---|---|---|---|---|
| 규칙 1000+ | 토큰 초과 ❌ | ✅ (벡터DB) | ✅ | ✅ (폴더 트리) |
| 인프라 비용 | ₩0 | 서버 $$$ | 서버 $$$ | **₩0** |
| AI 교체 시 | 파일 복사 | 마이그레이션 | 마이그레이션 | **그대로** |
| 자가 성장 | ❌ | ✅ | ✅ | **✅ (교정→뉴런)** |
| 불변 가드레일 | ❌ | ❌ | ❌ | **✅ (brainstem + bomb)** |
| 감사 | git diff | 쿼리 | 로그 | **ls -R** |

---

## 한계

| 항목 | 현황 | 대응 |
|------|------|------|
| AI 강제력 | 100% 준수 보장 불가 | 하네스가 위반 감지 → 교정 루프. 실측 94.9% |
| 시맨틱 검색 | 벡터 임베딩 없음 — 설계 의도 | 폴더 구조 자체가 검색 |
| 외부 검증 | 1인 운영 환경에서만 실전 검증 | 공개 후 다양한 환경 피드백 수집 예정 |
| 조건부 논리 | 폴더명만으로 if/else 불가 | `_rules.md`가 조건 분기 담당 |

---

## FAQ

**Q: "결국 시스템 프롬프트에 텍스트 넣는 거 아니야?"**
맞다. `--emit`은 폴더 트리를 텍스트로 컴파일한다. 핵심은 그 텍스트를 누가, 어떻게 관리하느냐다. 수천 줄 프롬프트 직접 편집 vs `mkdir` 하나 — 유지보수 비용이 다르다.

**Q: "bomb이 발동하면 그 규칙이 빠지는 거야?"**
아니다. bomb은 해당 **영역 전체 출력을 멈추는 서킷 브레이커**다. cortex에 bomb → 코딩 영역 자체 차단 → 사용자가 해제(`rm bomb.neuron`)할 때까지 AI가 코딩 자체를 못 한다.

**Q: "뉴런이 1000개 넘으면 토큰 폭발 아닌가?"**
3가지 방어선: ① 3-Tier 활성화 — 관련 영역만 깊게 읽음. ② Dormant — 30일 미접촉 시 자동 수면. ③ 통합 — 중복 뉴런 상위 병합.

**Q: "293개면 많은 거야?"**
개인으로는 많고, 엔터프라이즈로는 적다. 핵심은 숫자가 아니라 관리 가능성이다. 1000줄 프롬프트에서 237번째 줄 찾기 vs `brain/cortex/frontend/react/禁console_log/` 찾기.

**Q: "MBTI 에이전트가 과학적이야?"**
MBTI가 과학적이냐는 논쟁과 무관하다. "역할이 아니라 성향으로 나눈다"는 설계 결정이다. Big5든 다른 프레임워크든 대체 가능하다.

---

## 이야기

> *"프롬프트로 구걸하지 마. 파이프라인을 설계해."*

AI가 "console.log 쓰지 마"를 9번 어겼다. 10번째에 `mkdir brain/cortex/frontend/coding/禁console_log`를 만들었다. 폴더 이름이 규칙이 됐다. 카운터가 17이 됐다. AI가 더 이상 안 어긴다.

**NeuronFS가 존재하는 이유:** 더 큰 모델에 더 많은 컨텍스트를 먹이려는 게 아니라, 구조를 견고하게 만들어서 **AI 의존도를 0으로 수렴**시키는 것. 뇌가 제품이다. AI는 읽는 도구일 뿐.

> *"고급 모델은 구조를 만드는 데 쓴다. 최종장에서는 AI를 안 쓰거나, 트랜지스터급으로 최소 사용하는 게 목적이다."*

**⭐ 동의하면 Star. [아니면 Issue.](../../issues)**

---

MIT License · Copyright (c) 2026

[📜 전체 매니페스토](MANIFESTO.md) · [LIFECYCLE.md](LIFECYCLE.md)
