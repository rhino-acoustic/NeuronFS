# NeuronFS Runtime Code Map
> Generated: 2026-04-12 | 76 files | 17,350 lines | Go 1.26

## System Boot Flow (main.go)
```
main()
 ├→ RenderFlatlineOnPanic()     # panic_ux.go: death screen
 ├→ findBrainRoot()             # brain.go: locate brain_v4/
 ├→ MountCartridge()            # vfs_mount.go: VFS Hybrid Layer
 │   ├→ initVFS(rootDir)        #   Upper = os.DirFS (disk)
 │   ├→ StartIgnition()         #   vfs_ignition.go: brainwallet key
 │   ├→ DecryptCartridgeToRAM() #   crypto_cartridge.go: XChaCha20
 │   └→ zip.NewReader()         #   Lower = ZIP in RAM (O(1))
 ├→ RunAwakening()              # awakening.go: first-run animation
 └→ switch mode:
     ├→ --supervisor  → runSupervisor()     # supervisor.go
     ├→ --api/--dashboard → startAPI()      # api_server.go
     ├→ --mcp → startMCPHTTPServer()        # mcp_server.go
     ├→ --emit → emitRules()                # emit_bootstrap.go
     ├→ --inject → writeAllTiers()          # inject.go
     ├→ --grow/--fire/--signal/--rollback   # neuron_crud.go
     ├→ --watch → runWatch()                # watch.go
     ├→ --evolve → runEvolve()              # mcp_handler_evolve.go
     └→ --harness → RunHarness()            # diag_harness.go
```

## VFS Hybrid Architecture (★ Core Design)
```
RouterFS (vfs_core.go)
 ├→ Lower: .jloot Cartridge (RAM, immutable)
 │   └→ XChaCha20 decrypt → ZIP → fs.FS
 ├→ Upper: os.DirFS (disk, mutable)
 └→ Union: Upper優先, Lower fallback
 
 vfs_ops.go: vfsReadDir, vfsReadFile, vfsStat, vfsWalkDir, vfsGlob
 vfs_mount.go: MountCartridge (ignition sequence)
 vfs_ignition.go: StartIgnition (brainwallet key derivation)
 vfs_sync.go: disk↔cartridge sync
```

## Supervisor (supervisor.go — 611 lines)
```
runSupervisor()
 ├→ startAPI() as goroutine           # REST API + 대시보드 (port 9090)
 ├→ superviseMCPGoroutine()           # MCP Streamable HTTP (port 9247)
 ├→ telegram_bridge.go                # Telegram 양방향 (515 lines)
 ├→ hijack_orchestrator.go            # CDP 자율주행 (334 lines)
 │   ├→ DOM 스크래퍼 (transcript parsing)
 │   ├→ [EVOLVE:proceed] 감지
 │   └→ 마스터 프롬프트 인젝션 (60s debounce)
 ├→ cdp_client.go / cdp_monitor.go    # Chrome DevTools Protocol
 └→ context_hijacker.go               # IDE 컨텍스트 하이재킹
```

## Brain Scanner (brain.go — 494 lines)
```
scanBrain(root)
 ├→ regionsToScan: brainstem→prefrontal + shared
 ├→ flat neuron scan (region root *.neuron)
 ├→ vfsWalkDir: 폴더=뉴런, 파일=트레이스
 │   ├→ counter: N.neuron 파일명 숫자
 │   ├→ contra: N.contra 파일명
 │   ├→ dopamine: dopamineN.neuron
 │   ├→ bomb: bomb.neuron
 │   ├→ rule.md: description, globs, author
 │   └→ .dormant: 휴면 상태
 └→ runSubsumption: P0→P6 계층 정렬 + bomb 체크
```

## Emit Pipeline (3-Tier Rule System)
```
writeAllTiers() (inject.go)
 ├→ processInbox()                    # _inbox 처리
 ├→ scanBrain() → runSubsumption()
 ├→ emitBootstrap() (emit_bootstrap.go)
 │   ├→ emit_tiers.go: ALWAYS/WHEN/NEVER 분류
 │   ├→ emit_format_rules.go: 3-Tier 렌더링
 │   │   ├→ ALWAYS: 상시 규칙 (NeuronFS 공리 등)
 │   │   ├→ WHEN→THEN: 조건부 트리거 (max 8)
 │   │   └→ NEVER: 절대 금지 (max 15, score 정렬)
 │   └→ emit_helpers.go: 페르소나, Subsumption, 기억
 ├→ injectToGemini()                  # → ~/.gemini/GEMINI.md
 └→ AGENTS.md 동기화                   # → NeuronFS/AGENTS.md
```

## API Server (api_server.go → 9090 port)
```
startAPI()
 ├→ registerCRUDRoutes     # api_handler_crud.go: grow/fire/signal/rollback
 ├→ registerConfigRoutes   # api_handler_config.go: autopilot/emotion/sandbox
 ├→ registerSystemRoutes   # api_handler_system.go: inject/health/evolution/codemap
 ├→ registerStaticRoutes   # api_static.go: dashboard HTML + /api/brain
 ├→ runInjectionLoop()     # inject.go: dirty flag 기반 자동 inject
 └→ runIdleLoop()          # idle_worker.go: 유휴 시 자동 작업
```

## MCP Server (mcp_server.go → 9247 port)
```
startMCPHTTPServer()
 ├→ Native Tools (mcp_tools_native.go):
 │   ├→ read_neuron        # 뉴런 규칙 반환
 │   ├→ write_message      # inbox/outbox 제어
 │   ├→ grow_neuron        # 뉴런 생성 + author 기록
 │   └→ get_dashboard_state
 ├→ Handler Tools:
 │   ├→ mcp_handler_crud.go    # grow/fire/rollback/signal/correct
 │   ├→ mcp_handler_sys.go     # status/read_brain/health_check/report
 │   ├→ mcp_handler_read.go    # read_neuron
 │   ├→ mcp_handler_evolve.go  # evolve (Groq LLM)
 │   └→ mcp_handler_temporal.go # search (시간축 검색)
 └→ mcp_proxy.go: SSE fallback
```

## Neuron Lifecycle
```
neuron_crud.go:
 ├→ growNeuron()     # 생성 + hybridSimilarity 병합 (similarity.go)
 ├→ fireNeuron()     # counter++ + hebbianTrack
 ├→ rollbackNeuron() # counter--
 └→ signalNeuron()   # dopamine/bomb/memory

lifecycle.go (538 lines):
 ├→ pruneWeakNeurons()    # counter < threshold → 제거
 ├→ runDecay()            # TTL 기반 weight 감소
 ├→ RunTTLDecay()         # spaced repetition 연동
 └→ logEpisode()          # hippocampus/episodes 기록

hebbian.go: 30초 내 동시 발화 뉴런 상관관계 추적
spaced_repetition.go: 에빙하우스 망각 곡선 적용
```

## Dashboard (dashboard.go — 3D Three.js)
```
buildBrainJSONResponse()
 ├→ scanBrain() → runSubsumption()
 ├→ Cartridge scan (flat ReadDir, NOT Walk)
 └→ JSON: regions/neurons/axons/cartridges

dashboard_html.go: HTML template (Three.js 3D)
ops_dashboard.go: 운영 대시보드 (/api/ops)
```

## Key File Groups

### Core (< 500 lines each)
| File | Lines | Purpose |
|------|-------|---------|
| brain.go | 494 | Neuron struct, scanBrain, Subsumption |
| neuron_crud.go | 366 | CRUD: grow/fire/rollback/signal |
| inject.go | 297 | dirty flag, injection loop, writeAllTiers |
| governance_consts.go | 211 | SSOT: ports, thresholds, rune keys |
| similarity.go | 261 | Cosine bigram + Levenshtein hybrid |

### VFS Layer
| File | Lines | Purpose |
|------|-------|---------|
| vfs_core.go | 127 | RouterFS (UnionFS) |
| vfs_ops.go | 125 | Global VFS operations |
| vfs_mount.go | 57 | Cartridge mount |
| vfs_ignition.go | 72 | Brainwallet key derivation |
| vfs_sync.go | 27 | Disk↔Cartridge sync |

### Emit Pipeline
| File | Lines | Purpose |
|------|-------|---------|
| emit_bootstrap.go | 117 | Entry: emitBootstrap() |
| emit_format_rules.go | 617 | 3-Tier rule rendering |
| emit_helpers.go | 690 | Persona, Subsumption, memory |
| emit_tiers.go | 382 | ALWAYS/WHEN/NEVER extraction |
| emit_inbox_data.go | 228 | Inbox data processing |

### Infrastructure
| File | Lines | Purpose |
|------|-------|---------|
| supervisor.go | 611 | Process orchestration |
| telegram_bridge.go | 515 | Telegram 양방향 |
| hijack_orchestrator.go | 334 | CDP 자율주행 |
| transcript.go | 558 | DOM scrape + 전사 |
| lifecycle.go | 538 | Decay/prune/TTL |

### API
| File | Lines | Purpose |
|------|-------|---------|
| api_server.go | 71 | Route registration |
| api_handler_crud.go | 216 | REST CRUD |
| api_handler_config.go | 349 | Config/emotion |
| api_handler_system.go | 439 | System ops |
| api_static.go | 120 | Static files + /api/brain |

### LLM Integration
| File | Lines | Purpose |
|------|-------|---------|
| llm_groq.go | 275 | Groq API client |
| llm_prompts.go | 360 | LLM prompt templates |
| cli_llm.go | 602 | CLI LLM commands |
