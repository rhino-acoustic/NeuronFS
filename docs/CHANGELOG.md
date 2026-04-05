# Changelog

All notable changes to NeuronFS are documented here.

## [v4.4] - 2026-04-05

### Added
- **📋 Dashboard CODE Panel**: Runtime file tree visualization with `/api/codemap` API
- **go:embed**: Dashboard HTML externalized from Go string literal (1,156 → 11 lines)
- **Wargame scenarios**: 4 context-loss recovery scenarios tested and documented

### Changed
- **Extreme modularization**: `main.go` reduced from 3,538 → 396 lines (-88.8%)
- **emit.go split**: emit.go (825L) + emit_helpers.go (581L) from original 1,432L
- **New modules**: `watch.go` (135L), `diag.go` (261L), `neuron_crud.go` (277L), `inject.go` (287L), `transcript.go` (405L)
- **Total**: 30 Go source files, ~10,920 lines. 3,400+ neurons

### Fixed
- 18 broken `.axon` files recovered (UTF-8 encoding fix)
- Removed unused imports across all modules
- IDLE loop now auto-refreshes CODE_MAP.md and runs `go vet`

## [v4.3] - 2026-04-04

### Added
- **3D Brain Dashboard**: Three.js topology visualization
- **MCP Server**: Native stdio integration for AI IDEs
- **Merkle Chain**: Integrity verification for neuron state

### Changed
- Brain structure migrated from v3 to v4 format
- Subsumption cascade refined (P0-P6 priority)

## [v4.0] - 2026-03-30

### Added
- **Folder-as-Neuron paradigm**: `mkdir = create rule`
- **Subsumption Architecture**: 7 brain regions (brainstem → prefrontal)
- **Auto-neuronize**: Groq LLaMA-3.3-70B conversation analysis
- **Multi-IDE emit**: GEMINI.md, .cursorrules, CLAUDE.md simultaneous generation

## [v3.0] - 2026-03-28

### Added
- Initial brain_v3 structure
- Basic fsnotify watching

## [v1.0] - 2026-03-26

### Added
- Project inception
- First neuron filesystem prototype
