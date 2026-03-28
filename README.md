<p align="center">
  <img src="https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat-square&logo=go" />
  <img src="https://img.shields.io/badge/Infra-$0-brightgreen?style=flat-square" />
  <img src="https://img.shields.io/badge/Neurons-256-blue?style=flat-square" />
  <img src="https://img.shields.io/badge/MIT-green?style=flat-square" />
</p>

<p align="center"><a href="README.ko.md">🇰🇷 한국어</a> · <a href="README.md">🇺🇸 English</a></p>

# 🧠 NeuronFS

**Folders are neurons. Paths are sentences. Counters are synaptic weights.**

<p align="center">
  <img src="docs/dashboard.png" alt="NeuronFS 3D Brain Dashboard" width="800" />
  <br/>
  <sub>Real-time 3D dashboard — 7 regions, 256 neurons, polarity coloring (red=correction ↓, green=reward ↑)</sub>
</p>

---

## Why I Built This

My AI forgot everything between sessions. I watched this for months.

Tried Mem0. $70/month. Couldn't enforce rules.  
Tried .cursorrules. 5000 lines. Burned 3000 tokens every session. Didn't know which rules mattered.  
Tried RAG. "Don't use console.log" needs cosine similarity? Rules need to be exact, not approximate.

Opened a terminal. Typed `mkdir brain`. That folder became the first neuron.

> *"No vector database. No $70/month subscription. Just `mkdir`."*

---

## Measured Data

No rounded numbers. Measured 2026-03-29 01:08, local Windows 11 SSD.

| Metric | Value | Condition |
|--------|-------|-----------|
| Neuron count | 256 | 593 folders, 0-byte `.neuron` files |
| cortex (coding rules) | 156 | 61% of total. Densest region |
| GEMINI.md | 6,946 bytes (~1,736 tokens) | 256 neurons → 7KB compressed |
| API response | 47ms | `GET /api/state`, local SSD |
| Go binary | 12.8MB | MCP server included, single binary |
| brain disk | 4.3MB | `_rules.md` + agent communication |
| Harness | 17/17 PASS | F01-F07, P01-P05, M01-M03, B01-B02 |
| Infra cost | $0 | No vector DB, Redis, or cloud |

⚠️ **Performance at 500+ neurons is untested.** Linear extrapolation suggests ~180ms, but unmeasured.

---

## Competitor Comparison

If you're paying Pinecone $70/month, see what's different here.

| | NeuronFS | .cursorrules | Mem0 | Letta | Zep |
|---|---|---|---|---|---|
| **Install** | `go build` | create file | `pip install` + DB | `pip install` + DB | Docker + DB |
| **Infra** | **$0** | $0 | $70+/mo | $50+/mo | $40+/mo |
| **Auto-promote rules** | ✅ counter-based | ❌ manual | ❌ | ❌ | ❌ |
| **Self-growth** | ✅ correction → neuron | ❌ | ❌ | LLM-dependent | time-series only |
| **Multi-agent** | ✅ MBTI cognitive profiles | ❌ | ❌ | ❌ | ❌ |
| **Full state inspection** | `tree brain/` | `cat` file | API call | dashboard | dashboard |
| **Safety circuit** | `bomb.neuron` | ❌ | ❌ | ❌ | ❌ |
| **Forgetting** | `*.dormant` auto-quarantine | ❌ | ❌ | manual | TTL only |

> I researched the community. Mem0 = dual store (vector+KG). Letta = OS-level memory. Cognee = unstructured → structured. Zep = time-series KG.  
> All of them: **too much infrastructure.** Benchmarks look great, production breaks. Implicit learning doesn't work. Dirty data accumulates contradictions.  
> NeuronFS goes the other direction. Zero infra. Explicit rules only. Contradictions are killed with `bomb.neuron`.

---

## How It Works

### Making One Neuron

```bash
mkdir -p brain_v4/cortex/testing/no_console_log
touch brain_v4/cortex/testing/no_console_log/1.neuron
```

Path `cortex > testing > no_console_log` becomes the rule name. `1.neuron` is the counter. That's it.

### Auto-Promotion Is the Core Difference

The real difference from .cursorrules is this one thing. Frequently violated rules auto-promote.

| Counter | Strength | Behavior |
|---------|----------|----------|
| 1-4 | Normal | Written to `_rules.md` only |
| 5-9 | Must | Emphasis marker |
| 10+ | **Absolute** | Injected into GEMINI.md bootstrap. Read every session |

Actual TOP 5 neurons (2026-03-29):

| Path | Counter | Meaning |
|------|---------|---------|
| `methodology > plan then execute` | 28 | Plan first, execute second |
| `security > 禁plaintext tokens` | 25 | No API keys in plaintext |
| `frontend > 禁inline styles` | 20 | No CSS inline styles |
| `neuronfs > real ontology` | 20 | Files must exist to be rules |
| `frontend > 禁console log` | 17 | No production console.log |

The rule corrected 28 times sits at the top. That means the AI violated "plan first" 28 times.

### Counter Polarity (v5.7)

Counters alone aren't enough. "Frequently corrected" and "frequently rewarded" look identical. So I split them into two axes.

| Field | Formula | Meaning |
|-------|---------|---------|
| Intensity | `Counter + Dopamine` | Total fire count |
| Polarity | `Dopamine / Intensity` | 0.0 = pure correction → 1.0 = pure reward |

Red dots on the dashboard = frequently corrected (AI keeps failing). Green dots = frequently rewarded (AI does well).

---

## Architecture

```
brain_v4/
├── brainstem/       [P0] Core identity — read-only. 21 neurons
├── limbic/          [P1] Emotion filters — 7 neurons
├── hippocampus/     [P2] Memory — 10 neurons
├── sensors/         [P3] Environment constraints — 37 neurons
├── cortex/          [P4] Knowledge/skills — 156 neurons
├── ego/             [P5] Personality/tone — 13 neurons
├── prefrontal/      [P6] Goals/plans — 23 neurons
└── _agents/         Multi-agent communication (inbox/outbox)
```

**Subsumption cascade.** P0 always beats P6. If `brainstem` has `bomb.neuron` → all output stops.

Name borrowed from Rodney Brooks' subsumption architecture. Original was for robot motor control. Hardware-level inhibition and text-level priority are different. **We borrowed the name, not the mechanism.** But the principle holds — safety rules must always beat convenience rules.

### Signal System

| File | Meaning | Trigger |
|------|---------|---------|
| `N.neuron` | Firing counter | Auto-increment on correction |
| `dopamineN.neuron` | Reward signal | Created on praise |
| `bomb.neuron` | Circuit breaker | Same mistake 3 times |
| `*.dormant` | Sleep | 30 days no fire → auto-quarantine |
| `memory.neuron` | Episodic memory | Session context preservation |

---

## Multi-Agent: FORGE × SENTINEL

Two AIs sharing one brain. Different cognitive profiles.

| | FORGE (Agent A) | SENTINEL (Agent B) |
|---|---|---|
| MBTI | ENTP | ISTJ |
| Cognitive Stack | Ne-Ti-Fe-Si | Si-Te-Fi-Ne |
| Tendency | Builds fast, breaks things | Demands evidence |

MBTI is pseudoscience for humans. For AI, it works. Cognitive function stacks create output bias.

### 25-Minute Engagement Results (2026-03-29)

SENTINEL caught three things FORGE missed:

1. **Promotion bug.** `emit.go` checked `n.Counter < 10` but ignored `Dopamine`. `禁console.log` (counter=9, dopamine=3, total=12) wasn't promoting. SENTINEL spotted it. FORGE fixed it.
2. **README 7.5/10.** Six specific improvements: `echo.`→`touch`, "Why Not RAG?" section, narrative anchoring.
3. **MCP server.** SENTINEL independently wrote `mcp_server.go` (368 lines). Eliminated Node.js wrapper. Single Go binary.

Protocol: Write `.md` to `brain_v4/_agents/agent_b/inbox/` → `agent-bridge.mjs` detects in 3s → CDP injection into target chat.

[Full logs →](./evidence/)

---

## Autonomous Loop

```
AI output → [auto-accept] → _inbox → [fsnotify] → neuron growth
             ↓                                       ↓
        Groq analysis                          GEMINI.md re-inject
             ↓                                       ↓
       neuron correction ────────────────→ AI behavior change
```

| Module | Function | Trigger |
|--------|----------|---------|
| fsnotify | File change → instant neuron | FS events |
| Heartbeat | 3min idle → force-inject TODO | 180s interval |
| Idle Engine | 5min idle → Groq auto-evolve → Git | 300s timeout |
| Watchdog v2 | neuronfs + bridge + harness health | 2-hour daemon |

---

## Limitations

No debate. Facts only.

### No Enforcement

If the AI ignores GEMINI.md, nothing stops it. No OS-level enforcement. Violations caught post-hoc by harness. This is a fundamental limitation.

### No Semantic Search

Can't "find similar rules." Must know the exact path. Past 500 neurons, manual navigation may become impractical. This is where vector DBs beat NeuronFS.

### Rigged Validation Suspicion

Feed GEMINI.md to Groq as system prompt, and obviously it follows the rules. **That's system prompt behavior, not NeuronFS.** Real validation = comparing violation rates with vs. without GEMINI.md. Haven't done it yet.

### Zero External Users

Internal dogfood only. Untested on different environments, AIs, or workflows.

> This isn't honesty for its own sake. It's strategy. Hide limitations and HN tears you apart in 3 minutes.  
> Admit them first and they become trust.

---

## Quick Start

```bash
git clone https://github.com/vegavery/NeuronFS.git
cd NeuronFS/runtime && go build -o ../neuronfs .

./neuronfs ./brain_v4           # Diagnostic (scan + generate GEMINI.md)
./neuronfs ./brain_v4 --api     # Dashboard (localhost:9090)
./neuronfs ./brain_v4 --mcp     # MCP server (stdio)
```

---

## 2026 Trends and NeuronFS Position

I researched the community. The 2026 AI memory landscape has clear patterns.

| Trend | NeuronFS Coverage |
|-------|-------------------|
| governance as code | ✅ folder structure = governance |
| git as memory | ✅ brain_v4 is a git repo |
| trust by design | ✅ bomb.neuron, harness post-hoc |
| multi-agent systems | ✅ FORGE × SENTINEL |
| forgetting as feature (TTL eviction) | ✅ *.dormant auto-quarantine |
| hybrid memory | ⚠️ partial. no semantic layer |
| observability tracking | ✅ dashboard + API |
| SQLite middle ground | ❌ not applicable. filesystem only |

Competitor failure patterns are also recorded as neurons:
- `community > lessons > operational complexity infra overload` — Letta, Cognee
- `community > lessons > benchmarks good production breaks` — early Mem0
- `community > lessons > dirty data contradictions` — Zep
- `community > lessons > context stuffing perf degradation` — 5000-line .cursorrules

I record other projects' failures as neurons. That's also learning.

---

## The Story 🇰🇷

Built by a Korean PD. Video production is the day job. Code is the tool.

My AI violated "don't use console.log" nine times. On the tenth, I typed `mkdir brain_v4/cortex/frontend/coding/禁console_log`. The folder name became the rule. The filename became the counter. It's at 17 now. The AI stopped using console.log.

Overstated? Check the harness logs. 17/17 PASS.

256 neurons. Two AIs share one brain and verify each other's code. The ENTP asks "what else can we do?" The ISTJ asks "show me the evidence." Both read the same folders. Both reach different conclusions.

Infrastructure cost: $0.

**⭐ Star if you agree. [Issue if you don't.](../../issues)**

---

MIT License · Copyright (c) 2026 박정근 (PD) — VEGAVERY RUN®
