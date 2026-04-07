# NeuronFS Runtime — CODE_MAP
<!-- AUTO-GENERATED: Regenerated after each modularization step -->
<!-- Last updated: 2026-04-07T13:57 -->

## Architecture Overview
NeuronFS is a Folder-as-Neuron governance engine. All Go files are `package main`.
**Every function is documented.** Missing = bug.

### Neuro-Lifecycle Flow (2026-04-07 확정)
```
[대화 중] processChunk / harnessCycle → _signals/*.json (Signal 기록만)
[30분마다] 자동통합 스케줄러 → neuronfs --evolve (Groq 분석 → 승격/폐기)
[승격 시] 🧬 NEURON EVOLVED → 텔레그램 알림
```
⚠️ processChunk/harnessCycle은 더 이상 .neuron 파일을 직접 생성하지 않음 (2026-04-07 차단됨)

## Emit Flow (CRITICAL — 3 separate paths!)
```
neuronfs --emit       → case "emit"        → emitRules() → stdout ONLY (no file write!)
neuronfs --emit all   → case "emit-target" → writeAllTiersForTargets() → IDE files + _rules.md
neuronfs --inject     → case "inject"      → writeAllTiers() → GEMINI.md + _index.md + _rules.md
```
⚠️ `--emit` (no target) does NOT write _rules.md. Use `--inject` or `--emit all`.

## File Map (26 files, ~10,400 lines)

### 🏗 Core: VFS Engine (가상 파일 시스템)
| File | Lines | Funcs | Purpose |
|------|-------|-------|---------|
| **vfs_core.go** | 80 | 2 | RouterFS 구조체, UnionFS 병합(O(1) 섀도잉) 설계 |
| **vfs_ops.go** | 60 | 5 | os.* 및 filepath.* 를 대체하는 전역 래퍼 (Glob, WalkDir 등) |
| **vfs_mount.go** | 40 | 1 | .jloot 카트리지 마운т(임시 zip.Reader) 및 부팅 로직 |

---

### 🏗 Core: Data Structures + Scan
| File | Lines | Funcs | Purpose |
|------|-------|-------|---------|
| **brain.go** | 485 | 4 | Brain/Region/Neuron structs, scanBrain, runSubsumption |

| Function | Signature | Does |
|----------|-----------|------|
| `findBrainRoot` | `() string` | OS.Args + parent traversal로 brain_v4 경로 탐색 |
| `getFolderBirthTime` | `(folderPath) time.Time` | 폴더 생성 시간 (Windows CreationTime) |
| `scanBrain` | `(root) Brain` | Walk 전 영역 → neuronMap 구성 → Region/Neuron 배열 반환 |
| `runSubsumption` | `(brain) SubsumptionResult` | P0→P6 우선순위 적용, bomb 차단, dormant 필터 |

---

### 📡 Emit: Rule Generation (3파일 분리 완료)
| File | Lines | Funcs | Purpose |
|------|-------|-------|---------|
| **emit_bootstrap.go** | 564 | 4 | Tier 1 컨텐츠 생성 (GEMINI.md 문자열) |
| **emit_tiers.go** | 299 | 4 | 파일 쓰기 오케스트레이션 (3-tier write) |
| **emit_helpers.go** | 611 | 11 | Tier 2+3 컨텐츠 + tree 렌더링 |

#### Call Graph
```
main.go --inject → writeAllTiers (emit_tiers.go)
                     ├→ scanBrain → runSubsumption → applyOOMProtection
                     ├→ emitBootstrap (emit_bootstrap.go) → injectToGemini   [Tier 1]
                     ├→ emitIndex (emit_helpers.go) → _index.md              [Tier 2]
                     ├→ emitRegionRules (emit_helpers.go) × 7 → _rules.md    [Tier 3]
                     └→ generateBrainJSON (diag.go)

main.go --emit all → writeAllTiersForTargets (emit_tiers.go)
                      ├→ same as above
                      └→ writes to IDE-specific files (cursor, claude, etc.)

main.go --emit     → emitRules (main.go) → emitBootstrap → stdout ONLY!
```

#### emit_bootstrap.go
| Function | Does |
|----------|------|
| `emitBootstrap` | SubsumptionResult → GEMINI.md 문자열 (Tier 1 content) |
| `emitAgentInbox` | _inbox 폴더 스캔 → 에이전트 수신함 섹션 생성 |
| `extractInboxPreview` | 파일 내용에서 첫 줄 preview 추출 |
| `emitSessionMemory` | session_log → 세션 메모리 섹션 생성 |

#### emit_tiers.go
| Function | Does |
|----------|------|
| `writeAllTiers` | brain scan → GEMINI.md + _index.md + 7×_rules.md 일괄 작성 |
| `applyOOMProtection` | 뉴런 과다 시 low-weight 드롭 |
| `writeAllTiersForTargets` | brain scan → IDE별 파일 + _index.md + 7×_rules.md |
| `doInjectToFile` | 파일에 NEURONFS 마커 구간 교체 |

#### emit_helpers.go
| Function | Does |
|----------|------|
| `emitIndex` | brain → _index.md 문자열 (Tier 2 content) |
| `emitRegionRules` | Region → _rules.md 문자열 (Tier 3 content) |
| `renderTree` | treeNode → markdown indent 렌더링 (isHanjaFolder 감지) |
| `isHanjaFolder` | 한자 1글자 폴더 감지 (禁/必/推 등) |
| `renderTreeWithPrefix` | 한자 하위 노드에 koreanPrefix 전파 렌더링 |
| `handleReadRegion` | HTTP handler: MCP/API용 _rules.md 실시간 생성 |
| `splitNeuronPath` | 경로 → 파트 배열 (OS separator + "/" 둘 다 처리) |
| `pathToSentence` | 경로 → 자연어 문장 변환 (한자→한국어, _→공백) |
| `collectAllNeurons` | SubsumptionResult → 전체 뉴런 배열 |
| `sortedActiveNeurons` | 뉴런 배열 정렬 (counter 내림차순, limit) |
| `axonBoostNeurons` | axon 참조: 연결된 영역 뉴런 cross-reference |

---

### 🏗 Core: CLI + Orchestrator
| File | Lines | Funcs | Purpose |
|------|-------|-------|---------|
| **main.go** | 396 | 5 | CLI entry, mode dispatch, inject helpers |

| Function | Does |
|----------|------|
| `main` | CLI 파싱 → mode switch dispatch |
| `emitRules` | SubsumptionResult → emitBootstrap 래퍼 (stdout용) |
| `activationBar` | counter → ASCII bar (█████) |
| `injectToGemini` | rules → ~/.gemini/GEMINI.md 전체 덮어쓰기 |
| `doInject` | 기존 파일에 NEURONFS 마커 구간 교체 |

---

### 🧬 Neuron CRUD + Lifecycle
| File | Lines | Funcs | Purpose |
|------|-------|-------|---------|
| **neuron_crud.go** | 277 | 4 | grow/fire/rollback/signal |
| **lifecycle.go** | 378 | 4 | prune/decay/dedup/logEpisode |
| **similarity.go** | 261 | 10 | tokenize/stem/jaccard/hybrid/cosine/levenshtein |

#### neuron_crud.go
| Function | Does |
|----------|------|
| `growNeuron` | 새 뉴런 생성 (similarity ≥ 0.4 → merge) |
| `fireNeuron` | counter +1 (rename N.neuron → N+1.neuron) |
| `rollbackNeuron` | counter를 1로 리셋 |
| `signalNeuron` | dopamine/bomb/memory 시그널 전송 |

#### lifecycle.go
| Function | Does |
|----------|------|
| `pruneWeakNeurons` | counter=0 empty 뉴런 dormant 전환 |
| `runDecay` | N일 미발화 뉴런 → *.dormant |
| `logEpisode` | hippocampus 에피소드 기록 |
| `deduplicateNeurons` | hybridSimilarity로 중복 뉴런 병합 |

---

### 💉 Injection Pipeline
| File | Lines | Funcs | Purpose |
|------|-------|-------|---------|
| **inject.go** | 287 | 6 | dirty flag, inbox, injection loop |

| Function | Does |
|----------|------|
| `markBrainDirty` | dirty flag set (mutex-safe) |
| `consumeDirty` | dirty flag consume + reset |
| `computeMountHash` | brain 폴더 hash 변경 감지 |
| `autoReinject` | 변경 시 writeAllTiers 트리거 |
| `processInbox` | _inbox 폴더 처리 (corrections → neuronize) |
| `runInjectionLoop` | 30초 간격 dirty check + re-inject |

---

### 📜 Transcript + Idle
| File | Lines | Funcs | Purpose |
|------|-------|-------|---------|
| **transcript.go** | 404 | 6 | git snapshot, idle loop, digest, heartbeat |

| Function | Does |
|----------|------|
| `gitSnapshot` | brain 폴더 git commit 스냅샷 |
| `touchActivity` | 마지막 활동 시간 갱신 |
| `getLastActivity` | 마지막 활동 시간 조회 |
| `runIdleLoop` | 5분 간격 idle 작업 (dedup, heartbeat, transcript digest) |
| `digestTranscripts` | _transcripts 폴더 정리 |
| `writeHeartbeat` | _heartbeat.json 작성 (brain 상태 스냅샷) |

---

### 🌐 API + Interface
| File | Lines | Funcs | Purpose |
|------|-------|-------|---------|
| **api_server.go** | 67 | 1 | REST API entry + 라우트 등록 |
| **api_handlers.go** | 795 | 4 | CRUD/Config/System 핸들러 + rollbackAll |
| **api_static.go** | 101 | 1 | 대시보드/정적 파일 서빙 |
| **mcp_server.go** | 828 | 6 | MCP stdio (AI IDE 통합) |
| **mcp_tools_native.go** | 155 | 1 | MCP 도구 등록 |
| **dashboard.go** | 486 | 7 | Dashboard HTML 서버 |
| **dashboard_html.go** | 9 | 0 | 임베디드 HTML (go:embed) |
| **adapter.go** | 114 | 1 | 멀티 IDE 어댑터 |

#### Call Graph (API)
```
main.go --api → startAPI (api_server.go)
                  ├→ registerCRUDRoutes (api_handlers.go)
                  │    grow, fire, signal, decay, state, evolve, dedup, read, inject, rollback
                  ├→ registerConfigRoutes (api_handlers.go)
                  │    principles, emotion, sandbox
                  ├→ registerSystemRoutes (api_handlers.go)
                  │    integrity, community, report(s), evolution, retrieve, codemap
                  └→ registerStaticRoutes (api_static.go)
                       /, /3d, /v2, /cards, /brain.obj, /brain_state.json
```

---

### 🧬 Growth + Evolution
| File | Lines | Funcs | Purpose |
|------|-------|-------|---------|
| **evolve.go** | 644 | 9 | Groq 기반 자율 진화 |
| **neuronize.go** | 760 | 7 | Groq 기반 contra 생성 |
| **init.go** | 251 | 3 | initBrain 스캐폴드 |

---

### ⚙️ Infrastructure
| File | Lines | Funcs | Purpose |
|------|-------|-------|---------|
| **supervisor.go** | 577 | 13 | 프로세스 관리자 |
| **awakening.go** | 346 | 16 | 부팅 애니메이션 |
| **flatline_poc.go** | 420 | 14 | 패닉 사망 화면 |
| **watch.go** | 135 | 1 | fsnotify 감시 |
| **diag.go** | 261 | 4 | 진단 출력, brain JSON, CODE_MAP 갱신 |
| **cli_commands.go** | 66 | 2 | stats/vacuum CLI |
| **exec_safe.go** | 31 | 3 | 안전 실행 래퍼 |
| **physical_hooks.go** | 157 | 4 | USB/Telegram 알람 |

---

### 🌉 Hijack Launcher (Node.js — 텔레그램 브릿지 + CDP)
| File | Lines | Purpose |
|------|-------|---------|
| **hijackers/hijack-launcher.mjs** | 1295 | 텔레그램 브릿지, CDP 네트워크 캡처, 전사 기록, 자동통합 |

| Function/Section | Lines | Does |
|-----------|-------|------|
| `tgPoll` / `tgLog` | 60~280 | 텔레그램 수신/발신 (polling + sendMessage) |
| `_sendToTelegramInner` | 130~170 | 텔레그램 발송 (신호 로그 SKIP 필터 포함) |
| **자동통합 스케줄러** | 310~325 | 30분마다 `_signals/` 확인 → `--evolve` 실행 |
| `groqExtractNeurons` | 380~460 | Groq API로 대화에서 규칙 추출 |
| `processChunk` | 462~508 | 10메시지마다 → **유사도 분류**: 기존 뉴런 50%+ 유사 → fire(강화) / 아니면 Signal |
| `harnessCycle` | 600~765 | 50메시지마다 → **유사도 분류**: 동일 로직 |
| `discoverProcesses` | 730~780 | Antigravity 프로세스 CDP 포트 탐지 |
| `attachCDPNetwork` | 950~1100 | CDP WebSocket으로 네트워크 이벤트 캡처 |
| `attachDOMScraper` | 1100~1200 | CDP DOM 스크래핑 (프로젝트 이름 추출) |
| `domSockets` / CDP Injection | 140~165 | @멘션 라우팅 → 프로젝트별 CDP DOM 직접 입력 |

#### Call Graph (Hijack Launcher)
```
entry → discoverProcesses() → attachCDPNetwork() per tab
          ├→ USER 메시지 감지 → processChunk() → _signals/*.json
          ├→ AI 응답 감지 → 전사 기록
          └→ 50메시지마다 → harnessCycle() → _signals/*.json

tgPoll() → 텔레그램 메시지 수신
         ├→ @멘션 → domSockets[project] → CDP DOM injection
         └→ inbox/* 파일 기록

setInterval(30분) → _signals/ 확인 → exec('neuronfs --evolve')
```

---

### 🔒 Security
| File | Lines | Funcs | Purpose |
|------|-------|-------|---------|
| **access_control.go** | 152 | 4 | RBAC 정책 |
| **crypto_neuron.go** | 81 | 3 | AES-256 암호화 |
| **dek_manager.go** | 318 | 11 | DEK 키 관리 |
| **merkle_chain.go** | 205 | 4 | 무결성 해시 체인 |

## Dependency Graph
```
similarity.go (leaf — zero deps)
    ↑
brain.go (structs + scan — stdlib only)
    ↑
lifecycle.go (prune, decay, dedup)
    ↑
neuron_crud.go (grow, fire, signal, rollback)
    ↑
inject.go (dirty flag, inbox, injection loop)
    ↑
main.go (orchestrator: CLI dispatch)
    ↑              ↑
api_server.go    mcp_server.go
  ├→ api_handlers.go
  └→ api_static.go
    ↑
emit_bootstrap.go + emit_tiers.go + emit_helpers.go (rule generation)
    ↑
evolve.go, neuronize.go (Groq AI growth)
    ↑
supervisor.go (process management)
```

## ⚠️ Known Issues
1. ~~`emit.go` — writeAllTiers + writeAllTiersForTargets 중복 로직~~ → **emit_tiers.go로 분리 완료** ✅
2. ~~`api_server.go` L916 — 단일 파일에 모든 라우트~~ → **api_server + api_handlers + api_static 분리 완료** ✅
3. `mcp_server.go` L828 — tool이 많지만 구조 건전. 분리 불필요.
4. `main.go` — injectToGemini/doInject가 main.go에 존재 (향후 이동 후보)
5. `emit_bootstrap.go` L169 — DEBUG stderr 출력이 프로덕션에 남아있음

## Critical Rules
1. ALL files are `package main` — no sub-packages
2. Functions are shared across files without import
3. Import sync is the #1 failure mode when refactoring
4. Always verify: `go vet ./...` → `go build .` after ANY change
5. `--emit` ≠ `--inject`: emit은 stdout만, inject가 파일 작성
