# NeuronFS Runtime Architecture

> Auto-generated from CODE_MAP system. See `/api/codemap` for live data.

## File Structure (30 Go files, ~10,920 lines)

```
runtime/
├── main.go              396L  CLI dispatcher (entry point only)
├── brain.go             439L  Brain structs, scanBrain, runSubsumption
├── similarity.go        261L  Tokenize, hybridSimilarity (pure leaf)
├── lifecycle.go         378L  Prune, decay, dedup, logEpisode
├── neuron_crud.go       277L  Grow, fire, rollback, signal
├── inject.go            287L  Dirty flag, inbox, injection loop
├── transcript.go        405L  Git snapshot, idle engine, digest
├── watch.go             135L  fsnotify watcher (zero-polling)
├── diag.go              261L  Diagnostics, refreshCodeMap
├── emit.go              858L  Tier 0-2 rules generation
├── emit_helpers.go      581L  Index, region rules, tree rendering
├── api_server.go        967L  REST API + /api/codemap
├── mcp_server.go        828L  MCP stdio server
├── mcp_tools_native.go  155L  MCP tool registration
├── neuronize.go         760L  Groq-based auto-neuronize
├── evolve.go            644L  Autonomous evolution engine
├── supervisor.go        577L  Multi-process supervisor
├── dashboard.go         486L  Dashboard HTTP server
├── dashboard_html.go     11L  go:embed dashboard.html
├── dashboard.html      1170L  3D brain topology UI
├── flatline_poc.go      420L  Panic death screen
├── init.go              251L  Brain initialization
├── access_control.go    152L  RBAC
├── crypto_neuron.go      81L  AES-256 encryption
├── dek_manager.go       318L  DEK key management
├── merkle_chain.go      205L  Integrity hash chain
├── adapter.go           114L  Multi-IDE adapter
├── physical_hooks.go    157L  USB/Telegram alarms
├── awakening.go         346L  Boot sequence
├── cli_commands.go       66L  --stats, --vacuum
└── exec_safe.go          31L  Safe exec wrapper
```

## Design Principles

1. **ALL files are `package main`** — single binary, no sub-packages
2. **File name = Documentation** — self-referential architecture
3. **PROVIDES/DEPENDS headers** in every file for AI context recovery
4. **Strangler Fig pattern** — modularized from 3,538-line monolith

## Verification

```bash
cd runtime/
go vet ./...   # Must pass with 0 errors
go build .     # ~8s build time
```

## Live Dashboard

- `http://localhost:9090/` — 3D brain topology
- `http://localhost:9090/api/codemap` — runtime file tree JSON
- `http://localhost:9090/api/state` — brain state JSON
