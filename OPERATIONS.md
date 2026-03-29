# 🧠 NeuronFS 마스터 운영 문서

> **SSOT** — 이 문서가 NeuronFS 시스템의 유일한 진실.  
> **위치** — 코드와 함께 Git 추적. 세션 간 유실 없음.  
> **최종 갱신** — 2026-03-29 22:04 KST  
> **현재 상태** — 314 neurons | 453 activation | NOMINAL

---

## 1. Axiom (절대 원칙)

```
Folder = Neuron       ← 디렉토리 자체가 뉴런. 디렉토리명이 규칙명.
File   = Trace        ← 파일은 가중치/카운터/신호의 흔적.
Path   = Sentence     ← 경로가 규칙 문장을 형성.
```

### ⚠️ 혼동 금지

| 맞음 ✅ | 틀림 ❌ |
|---------|---------|
| 디렉토리가 뉴런 | .neuron 파일이 뉴런 |
| .neuron은 가중치 트레이스 | .neuron이 없으면 뉴런 아님 |
| 디렉토리명 = 규칙명 | 파일명 = 규칙명 |

---

## 2. 7개 리전 Subsumption 계층

```
P0 brainstem     양심/본능     불변(평생)    읽기전용(인간만)
P1 limbic        감정 필터     초 단위       시스템 자동
P2 hippocampus   기록/기억     이벤트 기반   자동 축적
P3 sensors       환경 제약     실시간        읽기전용(환경)
P4 cortex        지식/기술     분~일 단위    학습 가능
P5 ego           성향/톤       사용자 임의   사용자 설정
P6 prefrontal    목표/계획     주~월 단위    인간 설정
```

**상위 P가 하위 P를 삼킨다.** bomb(P0) → 전체 정지.

### 두 가지 뉴런 형태

| 형태 | 설명 | 리전 |
|------|------|------|
| **폴더형** | 하위 디렉토리가 뉴런 | cortex, sensors, hippocampus, prefrontal |
| **플랫형** | 리전 루트에 `이름.카운터.neuron` 직접 존재 | brainstem, limbic, ego |

---

## 3. 파일 타입

| 확장자 | 역할 | 예시 |
|--------|------|------|
| `N.neuron` | 발화 카운터 (가중치) | `16.neuron` = 16번 발화 |
| `dopamineN.neuron` | 보상 신호 | `dopamine3.neuron` |
| `bomb.neuron` | 서킷 브레이커 (3회 반복 실수) | 해당 경로 자동 차단 |
| `*.dormant` | 휴면 (30일 미발화) | 자동 격리 |
| `*.axon` | 리전 간 연결 | `cortex→hippocampus` |
| `*.memory` | 에피소드 기록 | 성공/실패 |
| `*.goal` | 목표 정의 | prefrontal 영역 |
| `*.geofence` | 컨텍스트 마스킹 | 특정 디렉토리에서만 적용 |
| `_rules.md` | 영역별 규칙 요약 (자동생성) | emit.go |

---

## 4. 실행 스택

### 진입점: `run-auto-accept.bat`

```
run-auto-accept.bat
├── Antigravity (CDP port 9000)
├── auto-accept.mjs              ← CDP 자동 수락 + Groq 분석
├── neuronfs --api               ← Dashboard + REST API (:9090)
├── neuronfs --watch             ← brain_v4 감시 + processInbox + emit
├── agent-bridge.mjs             ← 에이전트 inbox 폴링 (3초)
├── bot-heartbeat.mjs            ← 봇 idle 감지 + CDP kickstart (30초)
├── watchdog.ps1                 ← 전체 프로세스 생존 감시 (30초)
├── robocopy /MIR                ← 로컬 → NAS 동기화 (5초)
└── kickstart.mjs                ← 부팅 시 1회 에이전트 초기화
```

### 프로세스 역할 상세

| 프로세스 | 언어 | 역할 | 주기 |
|----------|------|------|------|
| **neuronfs --watch** | Go | fsnotify 감시 → processInbox → emit | 이벤트 |
| **neuronfs --api** | Go | 대시보드 + REST API | HTTP 요청 시 |
| **auto-accept.mjs** | Node | CDP로 봇 버튼 자동 클릭 | 이벤트 |
| **agent-bridge.mjs** | Node | `_agents/{bot}/inbox/` 폴링 → API 전달 | 3초 |
| **bot-heartbeat.mjs** | Node | 봇 idle 감지 → CDP kickstart | 30초 |
| **watchdog.ps1** | PS | 프로세스 사망 감시 + harness 10분 주기 | 30초 루프 |

### 통신 구조

**인프라 0원 원칙: 파일시스템이 유일한 IPC**

```
PD 교정 → AI가 corrections.jsonl 기록 (로컬 _inbox/)
  → fsnotify 감지 → processInbox()
  → 뉴런 디렉토리 생성 or fire
  → writeAllTiers() → GEMINI.md + _rules.md 갱신

watchdog 30초 루프:
  1. neuronfs/bridge/heartbeat 사망? → 자동 재시작
  2. harness 10분마다 실행 → 위반 시 bot1/inbox에 알림
  3. heartbeat 1분 로그 (전체 상태 원라인)

에이전트 간 통신:
  brain_v4/_agents/{name}/inbox/   ← 수신
  brain_v4/_agents/{name}/outbox/  ← 송신
  watchdog → bot1/inbox/ (harness 위반 알림)
```

### API 엔드포인트

| Method | Path | 소스 |
|--------|------|------|
| GET | `/api/state` | **brain_state.json 파일 읽기** (실시간 아님) |
| GET | `/api/brain` | `scanBrain()` 실시간 |
| POST | `/api/grow` | 뉴런 디렉토리 생성 |
| POST | `/api/fire` | 카운터 증가 |
| POST | `/api/signal` | 도파민/bomb/memory |
| POST | `/api/sandbox` | 대시보드 샌드박스 |

> `/api/state`는 `--watch` 모드가 갱신하는 brain_state.json을 그대로 반환.  
> 새 빌드 후 watch 프로세스도 재시작해야 API 값이 갱신됨.

### ⚠️ start_brain.bat 주의

**빌드 후 반드시 runtime/neuronfs.exe에도 복사해야 함.**

```powershell
go build -o "NeuronFS\neuronfs.exe" .
Copy-Item "NeuronFS\neuronfs.exe" "NeuronFS\runtime\neuronfs.exe"
```

---

## 5. 데이터 보호

### Git 추적 (근본 보호)

```
.gitignore에 brain_v4/ 없음 → Git이 디렉토리 추적
빈 디렉토리 → .gitkeep 파일로 Git 추적 보장
```

### NAS 동기화

```
방향: 로컬 → NAS (단방향)
robocopy "%BRAIN_PATH%" "%NAS_BRAIN%" /MIR /FFT /XO /MT:4
주기: 5초마다 반복 (run-auto-accept.bat L121)
```

### ⚠️ 쓰기 규칙 (절대)

| 규칙 | 이유 |
|------|------|
| **모든 쓰기는 로컬(`c:\...\brain_v4`)에** | `--watch`가 로컬 감시 |
| **NAS(`Z:\...`)에 직접 쓰기 금지** | /MIR로 다음 동기화 시 삭제됨 |
| **corrections.jsonl → 로컬 `_inbox/`에** | processInbox가 로컬만 읽음 |
| **`/api/grow`, `/api/fire` 사용** | API가 로컬에 생성 |

```powershell
# ✅ 올바른 corrections.jsonl 기록
$path = "c:\Users\BASEMENT_ADMIN\NeuronFS\brain_v4\_inbox\corrections.jsonl"
[IO.File]::AppendAllText($path, '{"type":"correction",...}' + "`n")

# ❌ 잘못된 기록 (NAS 직접)
$path = "Z:\VOL1\VGVR\BRAIN\...\corrections.jsonl"  # 이러면 안 됨
```

### 백업 계층

| 계층 | 위치 | 역할 | 방향 |
|------|------|------|------|
| **로컬 (작업)** | `c:\...\NeuronFS\brain_v4` | 원본. 모든 쓰기 여기에 | — |
| **Git (이력)** | `.git/` | 디렉토리 구조 + 코드 보존 | 수동 커밋 |
| **NAS (백업)** | `Z:\VOL1\VGVR\BRAIN\...\brain_v4` | 실시간 미러 | 로컬→NAS |
| **brain_state.json** | Git 42e071c | 경로 목록 스냅샷 | 비상 복구용 |

---

## 6. 검증 체크리스트

**모든 감사 시 이 순서로 확인:**

### A. 데이터 영속성 (최우선)

- [ ] `.gitignore`에 `brain_v4/` 없음
- [ ] `git status brain_v4/` — untracked 디렉토리 없음
- [ ] NAS 경로 접근 가능 (`Z:\VOL1\VGVR\BRAIN\...`)
- [ ] robocopy 프로세스 alive

### B. 뉴런 카운트

- [ ] `neuronfs brain_v4` diag 실행 → 305+ neurons
- [ ] `/api/state` totalNeurons 확인
- [ ] `/api/brain` 실시간 스캔 결과 = state와 일치

### C. 프로세스

- [ ] neuronfs --api (port 9090) alive
- [ ] neuronfs --watch alive
- [ ] agent-bridge alive
- [ ] auto-accept alive
- [ ] watchdog alive (선택)

### D. 기능

- [ ] `/api/grow` → 디렉토리 생성 확인
- [ ] `/api/fire` → 카운터 증가 확인
- [ ] `/api/signal` → dopamine 파일 생성 확인
- [ ] bomb.neuron 생성 → CIRCUIT BREAKER 발동 확인
- [ ] 대시보드 sandbox 입력 → 반영 확인

### E. Emit

- [ ] GEMINI.md 존재 + 크기 > 3KB
- [ ] 7개 리전 _rules.md 모두 > 0 bytes
- [ ] sensors/_rules.md 빈 파일 아닌지

### F. 에이전트

- [ ] bot1, entp, enfp, pm — inbox/outbox 디렉토리 존재
- [ ] pm outbox pulse 파일 100개 미만

---

## 7. 장애 이력

### 2026-03-29: 뉴런 251→40→89→305 복구

**증상:** API totalNeurons=40으로 감소  
**1차 오진:** main.go의 `.neuron` 필수 조건 때문이라 판단  
**2차 오진:** `.neuron` 마커 파일 소실이라 판단  
**진짜 원인:**
1. `.gitignore`에 `brain_v4/` 포함 → Git 미추적 → 디렉토리 소실
2. main.go에서 `.neuron` 없는 폴더를 스킵 → Axiom 위반

**해결:**
1. main.go 수정: 디렉토리 자체를 뉴런으로 인식
2. `/api/grow`로 260개 경로 복원
3. `.gitignore`에서 `brain_v4/` 제거

**교훈:** 감사 시 `.gitignore` 반드시 확인. 디렉토리가 뉴런.

### 2026-03-29: corrections.jsonl NAS 기록 사고

**증상:** corrections.jsonl에 교정 기록했는데 processInbox 미작동  
**원인:** corrections.jsonl을 NAS(`Z:\...`)에 직접 기록.  
`--watch`는 로컬 brain_v4를 감시하므로 NAS의 변경을 감지하지 못함.  
**해결:** 모든 쓰기를 로컬 경로로 통일. sensors/nas에 `禁NAS직접쓰기_로컬만` 뉴런 등록.

### 2026-03-29: `--watch` NAS 경로 연결 (run-auto-accept.bat L106)

**증상:** NAS 존재 시 `--watch`가 NAS 경로로 실행됨  
**영향:** AI가 로컬에 쓴 corrections.jsonl이 즉시 처리되지 않음 (5초 robocopy 지연)  
**수정 필요:** L106을 항상 로컬 brain_v4 경로로 변경

---

## 8. 로그 관리

### 통합 로그 디렉토리

```
NeuronFS/logs/
├── heartbeat.log      ← bot-heartbeat.mjs
├── bridge.log         ← agent-bridge.mjs
├── watchdog.log       ← watchdog.ps1
├── rotate.ps1         ← 일별 로테이션 (5MB 초과 시)
└── archive/           ← 로테이션된 로그 + 일회성 디버그
```

### 로테이션 규칙

- **5MB 초과** → `archive/`로 이동 (날짜 접미어)
- **7일 경과** → archive에서 자동 삭제
- watchdog이 harness 실행 시 `rotate.ps1` 호출

### 디버그 레벨

| 접두어 | 의미 |
|--------|------|
| `✅` | 정상 완료 |
| `⚡` | 킥/주입 |
| `⚠️` | 경고 (프로세스 사망 등) |
| `❌` | 실패 |
| `🚨` | 긴급 (bomb, harness 위반) |
| `💓` | heartbeat (1분 원라인 상태) |

---

## 9. 상호 감시 (Mutual Watchdog)

```
watchdog.ps1 → auto-accept/bridge/heartbeat/neuronfs 감시
auto-accept.mjs → watchdog.ps1 생존 확인 (30초)
```

**어느 쪽이 죽어도 상대방이 재시작.** 단일 장애점 제거.

### 뉴런 자동 생성 경로

```
AI 교정 감지 → corrections.jsonl (로컬 _inbox/)
  → fsnotify (neuronfs --watch)
  → processInbox()
  → mkdir (뉴런 디렉토리 생성)
  → counter fire
  → writeAllTiers() → GEMINI.md 갱신
```

**조건:** `neuronfs --watch` 프로세스가 살아있어야 함.

---

## 10. 금기사항

1. ❌ brain_v4 뉴런 폴더명 영어 번역/변환
2. ❌ 한자 접두어(禁, 推) 제거/변경
3. ❌ 뉴런 디렉토리 대량 삭제/재생성
4. ❌ 카운터 값 인위적 일괄 변경
5. ❌ brainstem (P0) 임의 변경
6. ❌ runtime 코드 PD 승인 없이 수정
7. ❌ `.gitignore`에 `brain_v4/` 추가

---

## 9. 멀티에이전트

| 코드명 | MBTI | 역할 |
|--------|------|------|
| ANCHOR (bot1) | ISTJ ♂ | 체계적 빌드 |
| FORGE (entp) | ENTP ♂ | 경계 파괴 |
| MUSE (enfp) | ENFP ♀ | 스토리텔링 |
| PM (pm) | — | 백로그 관제 |

통신: `_agents/{name}/inbox/outbox/` 파일시스템 비동기 메시징

---

## 10. 복구 절차

### 뉴런 디렉토리 소실 시

```powershell
# 1. brain_state.json에서 경로 추출 + /api/grow로 복원
go run restore_from_git.go  # C:\tmp\restore_from_git.go

# 2. 또는 수동
curl -X POST http://localhost:9090/api/grow -d '{"path":"cortex/frontend/css"}'
```

### neuronfs.exe 빌드 후

```powershell
# 반드시 두 곳에 복사
Push-Location NeuronFS\runtime
go build -o ..\neuronfs.exe .
Copy-Item ..\neuronfs.exe .\neuronfs.exe
Pop-Location

# 프로세스 재시작
taskkill /F /IM neuronfs.exe
Start-Process neuronfs.exe -ArgumentList "brain_v4","--api"
Start-Process neuronfs.exe -ArgumentList "brain_v4","--watch"
```

### 대시보드 sandbox 안 보일 때

main.go의 `scanBrain` Walk에서 `_sandbox` 폴더가 SkipDir 되지 않는지 확인.  
`_sandbox`는 `return nil`로 진입 허용 → 하위 폴더가 뉴런으로 인식되어야 함.

---

> **이 문서는 NeuronFS 레포에 Git 추적됩니다.**  
> **시스템 변경 시 이 문서도 함께 갱신합니다.**  
> **감사 시 섹션 6 체크리스트를 순서대로 실행합니다.**
