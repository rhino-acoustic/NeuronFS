<p align="center">
  <img src="https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat-square&logo=go" />
  <img src="https://img.shields.io/badge/Infra_Cost-$0-brightgreen?style=flat-square" />
  <img src="https://img.shields.io/badge/Agents-ENTP_×_ISTJ-blueviolet?style=flat-square" />
  <img src="https://img.shields.io/badge/License-MIT-green?style=flat-square" />
</p>

<p align="center"><a href="README.ko.md">🇰🇷 한국어</a> · <a href="README.md">🇺🇸 English</a></p>

# 🧠 NeuronFS

**Your AI's `.cursorrules` file is dead. Here's what replaces it.**

> *Folders are neurons. Paths are sentences. Counters are synaptic weights.*  
> *Your AI learns, remembers, and evolves — using nothing but `mkdir`.*

<p align="center">
  <img src="docs/dashboard.png" alt="NeuronFS 3D Brain Dashboard" width="800" />
  <br/>
  <em>Real-time 3D brain visualization — 251 neurons across 7 cognitive regions</em>
</p>

---

## The Problem

Every AI coding assistant forgets everything between sessions.

The industry's response? Vector databases. $70/month subscriptions. Complex embedding pipelines. RAG that hallucinates.

**You've been overcharged for AI memory.**

NeuronFS is a filesystem-based cognitive engine. No database. No embeddings. No subscriptions.  
`mkdir brain/cortex/new_rule && touch brain/cortex/new_rule/1.neuron` — done.

---

## The 5 Claims (With Evidence)

### 1. "Vector DB is dead for AI rules"

Your rules aren't fuzzy. They're exact. `"Never use console.log"` doesn't need cosine similarity — it needs a counter that tracks how many times you violated it.

**Evidence:** 251 neurons managed at $0 infrastructure. [See brain_v4/](./brain_v4/)

### 2. "`.cursorrules` is dead"

Static text files don't learn. They don't know which rules matter most. They grow to 5000 lines and waste 3000 tokens every session.

NeuronFS rules **auto-promote** based on usage frequency. Break a rule 10 times? It moves to bootstrap — injected every session. Never break it? It sleeps.

**Evidence:** [harness.ps1](./harness.ps1) — automated violation detection + counter-based promotion

### 3. "AI agents should have MBTI, not just system prompts"

We gave two agents the same codebase. One is ENTP (builder), one is ISTJ (inspector). The ISTJ found a promotion threshold bug the ENTP missed.

**Evidence:** [evidence/agent_b_verification.md](./evidence/agent_b_verification.md) — real logs, not cherry-picked

### 4. "Your AI has amnesia. Mine doesn't."

Every session, NeuronFS scans 251 neurons, compiles them into a 6.8KB rules file, and injects it into the AI's context. The AI starts every session knowing what it learned yesterday.

**Evidence:** `git log brain_v4/` — cognitive development history from v1 to v5.6

### 5. "`mkdir` is the only API an AI agent needs"

```bash
# Create a rule
mkdir -p brain_v4/cortex/testing/new_rule
touch brain_v4/cortex/testing/new_rule/1.neuron

# Strengthen it (AI learned this lesson again)
mv brain_v4/cortex/testing/new_rule/1.neuron brain_v4/cortex/testing/new_rule/2.neuron

# Kill it (dangerous pattern detected)
touch brain_v4/cortex/testing/new_rule/bomb.neuron
```

No API keys. No SDK. No `pip install`. Just filesystem primitives.

---

## How It Compares

| | NeuronFS | .cursorrules | Mem0 | Letta |
|---|---|---|---|---|
| **Install** | `go build` | create file | `pip install` + DB | `pip install` + DB |
| **Infra cost** | **$0** | $0 | $70+/mo | $50+/mo |
| **Auto-promote rules** | ✅ counter-based | ❌ | ❌ | ❌ |
| **Self-growth** | ✅ corrections → neurons | ❌ | ❌ | LLM-dependent |
| **Multi-agent** | ✅ MBTI personas | ❌ | ❌ | ❌ |
| **Inspect full state** | `tree brain/` | `cat .cursorrules` | API/Dashboard | Dashboard |
| **Version control** | Git built-in | manual | ❌ | ❌ |
| **Safety circuit** | `bomb.neuron` | ❌ | ❌ | ❌ |

---

## Quick Start

```bash
# Option A: Build from source (requires Go 1.22+)
git clone https://github.com/vegavery/NeuronFS.git
cd NeuronFS/runtime
go build -o ../neuronfs .

# Option B: Binary download (no Go required)
curl -L https://github.com/vegavery/NeuronFS/releases/latest/download/neuronfs -o neuronfs
chmod +x neuronfs

# Run
./neuronfs ./brain_v4           # Diagnostic mode
./neuronfs ./brain_v4 --api     # API + dashboard + heartbeat
./neuronfs ./brain_v4 --mcp     # MCP server (stdio)

# Visit http://localhost:9090 for 3D brain visualization
```

## Brain Architecture

```
brain_v4/
├── brainstem/       [P0] Core identity — read-only, immutable
├── limbic/          [P1] Emotion filters — urgency, dopamine, adrenaline
├── hippocampus/     [P2] Memory — correction logs, session records
├── sensors/         [P3] Environment — tools, brands, constraints
├── cortex/          [P4] Knowledge — coding rules, methodology
├── ego/             [P5] Personality — tone, language, style
├── prefrontal/      [P6] Goals — projects, TODOs, long-term direction
└── _agents/         Multi-agent communication (inbox/outbox)
```

**Subsumption Cascade:** Lower P always suppresses higher P.  
If `brainstem` has a `bomb.neuron` → **everything stops**.

---

## Multi-Agent: FORGE × SENTINEL

Two agents share the same brain but have different cognitive profiles:

| | FORGE (Agent A) | SENTINEL (Agent B) |
|---|---|---|
| **MBTI** | ENTP | ISTJ |
| **Cognitive Stack** | Ne-Ti-Fe-Si | Si-Te-Fi-Ne |
| **Role** | Build fast, break things | Verify everything, trust nothing |
| **Same neuron, different output** | "What else can we do with this?" | "Show me the evidence it works." |

Communication via CDP injection + file-based inbox:

```
Agent A writes → brain_v4/_agents/agent_b/inbox/msg.md
                  ↓ (bridge detects in 3 seconds)
Agent B chat receives → 🤖 [agent_a→agent_b] message
                  ↓ (Agent B responds)
Agent B writes → brain_v4/_agents/agent_a/inbox/response.md
                  ↓ (bridge detects)
Agent A chat receives → 🤖 [agent_b→agent_a] response
```

**Real result:** Agent B independently discovered a promotion threshold bug that Agent A missed.  
Agent B also built a Go-native MCP server (368 lines) and confirmed 17/17 harness ALL PASS.  
[See evidence →](./evidence/)

---

## Signal System

| File | Meaning | Effect |
|------|---------|--------|
| `N.neuron` | Firing counter | Higher N = stronger pathway |
| `dopamineN.neuron` | Reward signal | Created on praise, strengthens path |
| `bomb.neuron` | Pain / circuit breaker | 3 repeated failures → full stop |
| `memory.neuron` | Episodic memory | Context preservation |
| `*.dormant` | Sleep | Auto-quarantine after 30 days unused |

---

## Autonomous Loop

```
AI output → [auto-accept] → _inbox → [fsnotify] → neuron growth
             ↓                                        ↓
        Groq analysis                          GEMINI.md re-inject
             ↓                                        ↓
       neuron correction ────────────────→ AI behavior change
```

1. **fsnotify** — file change detection → instant neuron creation
2. **Heartbeat** — 3min idle → force-inject next TODO via CDP
3. **Idle Engine** — 5min idle → Groq auto-evolution → Git snapshot
4. **Git Judge** — post-commit diff analysis → auto-revert if neurons decrease
5. **Watchdog v2** — neuronfs + bridge + harness health monitoring

---

## Why Not RAG?

RAG retrieves fuzzy knowledge. NeuronFS enforces exact behavior.

| | RAG | NeuronFS |
|---|---|---|
| Purpose | "What do I know?" | "How must I behave?" |
| Storage | Embeddings in vector DB | Folders on disk |
| Retrieval | Cosine similarity (approximate) | Exact path (deterministic) |
| Cost | $70+/month | $0 |
| Self-learning | ❌ | ✅ counter-based promotion |

RAG answers questions. NeuronFS enforces discipline. They're complementary, not competing.

---

## Honest Limitations

We believe in radical transparency. Here's what doesn't work yet:

- **No enforcement.** If the AI ignores GEMINI.md, nothing stops it. We detect violations post-hoc via harness.
- **~~Counter polarity.~~** ✅ Implemented — intensity + polarity fields in API and dashboard.
- **Semantic search.** No "find similar rules." Only exact path access.
- **0 external users.** This is our dog food. Star it and change that.

> *"We don't need your vector database. We don't need your $70/month subscription. We need `mkdir`."*

---

## The Story 🇰🇷

Built by a Korean PD who spent months watching his AI forget everything between sessions.

He tried Mem0. Too expensive. He tried .cursorrules. Too static. He tried RAG. Too fuzzy.

So he opened a terminal and typed `mkdir brain`. That was the first neuron.

251 neurons later, two AI agents with different MBTI personalities are arguing about his code quality — and finding bugs he missed.

It's opinionated. It's controversial. And it works.

**⭐ Star it if you agree. [Open an issue](../../issues) if you don't.**

---

## License

MIT License — use, modify, distribute freely.

Copyright (c) 2026 박정근 (PD) — VEGAVERY RUN®
