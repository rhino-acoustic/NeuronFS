# 📚 NeuronFS Documentation

## Core Documents

| Document | Description |
|----------|-------------|
| [ARCHITECTURE.md](ARCHITECTURE.md) | Runtime code structure — 30 Go files, ~10,920 lines |
| [AUDIT_REPORT.md](AUDIT_REPORT.md) | Third-party architecture & market audit |
| [WARGAME_SCENARIOS.md](WARGAME_SCENARIOS.md) | Context loss recovery — 4 tested scenarios |
| [CHANGELOG.md](CHANGELOG.md) | Version history and breaking changes |

## Quick Links

- **Live Dashboard**: `http://localhost:9090/`
- **Code Map API**: `http://localhost:9090/api/codemap`
- **Brain State API**: `http://localhost:9090/api/state`
- **MCP Server**: stdio mode via `neuronfs brain_v4 --mcp`

## Architecture Overview

```
neuronfs.exe ─── single Go binary, zero dependencies
     │
     ├── --watch      → fsnotify real-time brain monitoring
     ├── --supervisor  → 3-process daemon (watch + dashboard + API)
     ├── --mcp        → MCP stdio server for AI IDE integration
     ├── --emit all   → compile brain to GEMINI.md / .cursorrules / CLAUDE.md
     └── --diag       → diagnostic output
```

## For New AI Sessions

1. Read `runtime/CODE_MAP.md` first
2. Or call `GET /api/codemap` for live file structure
3. Check the dashboard 📋 CODE panel for visual overview
