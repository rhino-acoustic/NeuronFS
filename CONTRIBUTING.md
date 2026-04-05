# Contributing to NeuronFS

Thank you for your interest in contributing to NeuronFS.

## Quick Links

- [Issues](https://github.com/rhino-acoustic/NeuronFS/issues) — Bug reports and feature requests
- [README](README.md) — Project overview
- [Limitations](README.md#limitations) — What NeuronFS can't do (read this first)

## Ways to Contribute

### 🐛 Bug Reports

Found a bug? [Open an issue](https://github.com/rhino-acoustic/NeuronFS/issues/new) with:
1. What you did (steps to reproduce)
2. What you expected
3. What actually happened
4. Environment (OS, Go version, neuron count)

### 💡 Feature Requests

Have an idea? [Open an issue](https://github.com/rhino-acoustic/NeuronFS/issues/new) with the `enhancement` label.

Some areas where we know we need help:
- **Semantic search layer** — NeuronFS has no "find similar rules" capability. This is where vector DBs beat us.
- **Cross-platform testing** — We develop on Windows. Linux/macOS testing is minimal.
- **Multi-editor emit** — Support for `.cursorrules`, `CLAUDE.md`, `.github/copilot-instructions.md` output formats.
- **Stress testing at scale** — We've tested up to 3,400+ neurons. What happens at 5,000? 10,000?
- **A/B validation** — Comparing violation rates with vs. without GEMINI.md. We haven't done this yet.

### 🔧 Pull Requests

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/your-feature`)
3. Make your changes in `runtime/`
4. Run the harness: `powershell -File harness.ps1`
5. Ensure all tests pass
6. Submit a PR with a clear description

### 📝 Documentation

- Fix typos, improve explanations
- Translate to other languages
- Write tutorials or blog posts about NeuronFS

## Development Setup

```bash
git clone https://github.com/rhino-acoustic/NeuronFS.git
cd NeuronFS/runtime
go build -ldflags="-s -w" -trimpath -buildvcs=false -o ../neuronfs .

# Run diagnostic
../neuronfs ../brain_v4

# Run dashboard
../neuronfs ../brain_v4 --api

# Run MCP server
../neuronfs ../brain_v4 --mcp
```

## Project Structure

```
NeuronFS/
├── runtime/              # Go source (30 files, ~10,920 lines)
│   ├── main.go           # CLI dispatcher (396L)
│   ├── brain.go          # Brain scanner/structs
│   ├── emit.go           # _rules.md generator (3-tier)
│   ├── api_server.go     # REST API + /api/codemap
│   ├── mcp_server.go     # MCP stdio server
│   ├── watch.go          # fsnotify watcher
│   ├── evolve.go         # Groq autonomous evolution
│   ├── dashboard.html    # 3D brain dashboard (go:embed)
│   └── ...               # 22 more domain modules
├── brain_v4/             # The brain (3,400+ neurons)
├── docs/                 # Architecture, audit, changelog
└── evidence/             # Multi-agent verification logs
```

## Code Style

- Go standard formatting (`gofmt`)
- No `console.log` equivalent in Go (use `fmt.Fprintf(os.Stderr, ...)` for debug output)
- No inline styles in HTML templates
- Plan first, execute second (this rule was violated 36 times — it's our top neuron)

## Community

- Be respectful and constructive
- Admit limitations honestly
- Data over opinions

---

**⭐ Star if you find this useful. [Issue if you don't.](https://github.com/rhino-acoustic/NeuronFS/issues)**
