# NeuronFS

> **AI를 제어하는 건 프롬프트가 아니라 폴더 구조다.**

![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)

**[English README →](README.md)** | **[선언문 →](MANIFESTO.md)**

---

## 문제

AI 코딩 에이전트(Cursor, Windsurf, Copilot, Gemini Code Assist)는 시작할 때 규칙 파일을 읽는다 — `.cursorrules`, `GEMINI.md`, `AGENTS.md` 등.

이 파일들은 **거대한 비정형 텍스트 덩어리**로 성장한다:
- 우선순위가 없다 (47번째 규칙과 1번째 규칙의 무게가 같다)
- 실제로 잘 먹히는 규칙과 안 먹히는 규칙을 구분할 방법이 없다
- 50개 넘으면 관리 불가능
- AI가 어떤 규칙을 따르는지 가시성이 전무하다

## 통찰

**규칙 파일이 파일이 아니라 디렉토리 트리라면?**

```
brain/
├── brainstem/                    # 절대 규칙 — 항상 강제
│   ├── never_delete_production/
│   │   └── 99.neuron            # 99 = 활성화 강도
│   └── verify_before_deploy/
│       └── 50.neuron
├── cortex/                       # 지식 — 맥락에 따라 적용
│   ├── frontend/
│   │   └── react/
│   │       └── hooks_pattern/
│   │           └── 15.neuron
│   └── backend/
│       └── supabase/
│           └── rls_always_on/
│               └── 20.neuron
└── sensors/                      # 환경 제약
    └── nas_write_cmd_only/
        └── 30.neuron
```

스캐너가 이 트리를 읽고 규칙 파일로 컴파일한다 — 우선순위순 정렬, 활성화 카운터 기반 가중치, 계층 구조 반영.

**폴더 = 개념. 파일 = 신호 강도. 경로 = 문장.**

`cortex/frontend/react/hooks_pattern/15.neuron`의 의미:
*"cortex(지식)에서, frontend, 그 중 React의 hooks 패턴을 적용 — 활성화 가중치 15."*

## 핵심 공리

| 공리 | 의미 |
|---|---|
| **폴더 = 뉴런** | 디렉토리 하나가 개념 하나. 이름이 곧 의미. |
| **경로 = 문장** | 전체 경로가 자연어 규칙으로 읽힌다. |
| **카운터 = 강도** | `N.neuron` — N이 클수록 강하게 강제. |
| **깊이 = 구체성** | `cortex/frontend/`은 넓은 범위. `cortex/frontend/react/hooks/useCallback/`은 정밀. |

## 왜 폴더가 파일을 이기는가

| | 단일 규칙 파일 | NeuronFS (폴더 트리) |
|---|---|---|
| **우선순위** | 없음 (평문 텍스트) | 구조적 계층 |
| **규칙 추가** | 텍스트 편집, 순서 고려 | `mkdir` + `touch` — 끝 |
| **규칙 삭제** | 찾아서 지우기 | `rm -rf` 폴더 |
| **변경 추적** | 모놀리스 diff | 폴더별 `git log` |
| **확장성** | ~100개에서 붕괴 | 150+ 뉴런 테스트 완료, 1000+까지 설계 |
| **가시성** | 전체 파일 읽어야 함 | `tree brain/` — 즉시 감사 |
| **비용** | ₩0 | ₩0 |

## 자가 성장

AI가 작동 중에 자기 규칙 트리를 수정할 수 있다:

```bash
# 사용자가 AI를 교정 → AI가 새 규칙 생성
mkdir -p brain/cortex/frontend/no_console_log
touch brain/cortex/frontend/no_console_log/1.neuron

# 규칙이 잘 작동 → 카운터 증가 (강화 학습)
mv 1.neuron 2.neuron

# 같은 실수 3번 반복 → 서킷 브레이커
touch brain/cortex/frontend/no_console_log/bomb.neuron
```

`git commit` 하나하나가 인지 성장 로그가 된다.

## 빠른 시작

```bash
# 1. 뇌 만들기
mkdir -p brain/brainstem/verify_before_deliver
touch brain/brainstem/verify_before_deliver/1.neuron

# 2. 도메인 지식 추가
mkdir -p brain/cortex/frontend/react/hooks_pattern
touch brain/cortex/frontend/react/hooks_pattern/1.neuron

# 3. 스캔 — 스크립트가 트리를 읽고 규칙 파일 생성
# (스캐너 구현은 환경에 따라 다름)
```

스캐너가 폴더 트리를 AI 에이전트가 필요한 형식으로 컴파일한다: `.cursorrules`, `GEMINI.md`, `AGENTS.md`, 또는 평문 텍스트.

## 현재 상태

**일일 실서비스 운영 중** — AI 에이전트 1개(Gemini/Antigravity)와 함께 사용 중.

검증된 것:
- ✅ 폴더 기반 규칙이 텍스트 파일보다 관리가 빠름
- ✅ 카운터 기반 활성화 가중치가 우선순위 제어에 작동함
- ✅ AI가 런타임에 `mkdir`로 자기 규칙을 생성할 수 있음
- ✅ `git log`가 인지 발달 이력 제공

진행 중:
- 🔧 Go 런타임 (자동 스캔/주입)
- 🔧 멀티 에이전트 호환성 테스트
- 🔧 대규모(500+ 뉴런) 성능 벤치마크

---

> 📜 **[선언문 읽기 →](MANIFESTO.md)** — NeuronFS의 철학적 배경 전체.

---

**Created by:** [@rhino-acoustic](https://github.com/rhino-acoustic)
