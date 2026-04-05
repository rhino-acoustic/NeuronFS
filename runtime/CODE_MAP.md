# NeuronFS Runtime — CODE_MAP
<!-- AUTO-GENERATED: Regenerated after each modularization step -->
<!-- Last updated: 2026-04-05T10:19 -->

## Architecture Overview
NeuronFS is a Folder-as-Neuron cognitive engine. Each Go file in `runtime/`
is a module within `package main` — the compiler treats them as one unit.
Files communicate via shared functions and package-level variables.

## File Tree (27 files, ~10,800 lines)

### 🏗 Core (Orchestrator + Data)
| File | Lines | Functions | Role |
|------|-------|-----------|------|
| main.go | 710 | 9 | CLI entry, watch, inject dispatcher |
| brain.go | 439 | 4 | Neuron/Region/Brain structs, scanBrain, runSubsumption |
| similarity.go | 261 | 10 | tokenize, stem, hybridSimilarity (leaf — zero deps) |
| lifecycle.go | 378 | 4 | pruneWeakNeurons, runDecay, deduplicateNeurons, logEpisode |
| neuron_crud.go | 277 | 4 | growNeuron, fireNeuron, rollbackNeuron, signalNeuron |
| inject.go | 287 | 6 | markBrainDirty, processInbox, runInjectionLoop |
| transcript.go | 389 | 6 | gitSnapshot, runIdleLoop, digestTranscripts, writeHeartbeat |

### 🌐 API + Interface
| File | Lines | Functions | Role |
|------|-------|-----------|------|
| api_server.go | 916 | 2 | REST API (startAPI), rollbackAll |
| mcp_server.go | 828 | 6 | MCP stdio server (for AI IDE integration) |
| mcp_tools_native.go | 155 | 1 | Native MCP tool registration |
| dashboard.go | 486 | 7 | Dashboard server, health/brain JSON builders |
| dashboard_html.go | 1155 | 0 | Embedded HTML (card-based dashboard) |

### 📡 Emit + Rules
| File | Lines | Functions | Role |
|------|-------|-----------|------|
| emit.go | 1432 | 19 | _rules.md generation, GEMINI.md injection, tier system |
| adapter.go | 114 | 1 | Multi-IDE retrieval adapter |

### 🧬 Growth + Evolution
| File | Lines | Functions | Role |
|------|-------|-----------|------|
| evolve.go | 644 | 9 | Groq-based autonomous evolution |
| neuronize.go | 760 | 7 | Groq-based contra neuron generation |
| init.go | 251 | 1 | initBrain (first-run scaffold) |

### ⚙️ Infrastructure
| File | Lines | Functions | Role |
|------|-------|-----------|------|
| supervisor.go | 577 | 11 | Process supervisor (child process mgmt) |
| awakening.go | 346 | 10 | Boot animation + first-run detection |
| flatline_poc.go | 420 | 9 | Panic death screen (flatline EEG) |
| cli_commands.go | 66 | 2 | --stats, --vacuum CLI commands |
| exec_safe.go | 31 | 3 | Safe process execution wrappers |
| physical_hooks.go | 157 | 4 | Physical alarm triggers (USB, Telegram) |

### 🔒 Security
| File | Lines | Functions | Role |
|------|-------|-----------|------|
| access_control.go | 152 | 4 | RBAC policy enforcement |
| crypto_neuron.go | 81 | 3 | AES-256 neuron encryption |
| dek_manager.go | 318 | 1 | Data Encryption Key lifecycle |
| merkle_chain.go | 205 | 4 | Integrity verification (hash chain) |

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
main.go (orchestrator: CLI dispatch, watch)
    ↑          ↑
api_server.go  mcp_server.go
    ↑
emit.go (rule generation) → dashboard.go
    ↑
evolve.go, neuronize.go (Groq AI growth)
    ↑
supervisor.go (process management)
```

## Key Package Variables
### brain.go
- `regionPriority` — subsumption cascade (0=brainstem → 6=prefrontal)
- `counterRegex` / `dopamineRegex` — trace file patterns

### inject.go
- `brainDirty` / `brainDirtyMu` — dirty flag for batch injection
- `triggerChan` — injection trigger channel
- `lastMountHash` — mount hash for change detection

## Critical Rules
1. ALL files are `package main` — no sub-packages
2. Functions are shared across files without import
3. Import sync is the #1 failure mode when refactoring
4. Always verify: `go vet ./...` → `go build .` after ANY change
5. main.go went from 3,538 → 710 lines (-80%) via modularization
