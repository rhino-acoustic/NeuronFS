<p align="center">
  <img src="https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat-square&logo=go" />
  <img src="https://img.shields.io/badge/Infra-$0-brightgreen?style=flat-square" />
  <img src="https://img.shields.io/badge/7_Regions-Self_Evolving-blue?style=flat-square" />
  <img src="https://img.shields.io/badge/Go_Source-~10K_lines-informational?style=flat-square" />
  <img src="https://img.shields.io/badge/Zero_Dependencies-black?style=flat-square" />
  <img src="https://img.shields.io/badge/MIT-green?style=flat-square" />
</p>

<p align="center">
  <img src="docs/dashboard.png" alt="NeuronFS Dashboard — 3D Brain Visualization" width="800" />
  <br/>
  <a href="https://dashboarddeploy-six.vercel.app/"><strong>3D Dashboard Live Demo</strong></a>
</p>

<p align="center"><a href="README.ko.md">🇰🇷 한국어</a> · <a href="README.md">🇺🇸 English</a> · <a href="docs/ARCHITECTURE.md">🏗️ Architecture</a> · <a href="docs/CHANGELOG.md">📋 Changelog</a></p>

# NeuronFS
### *Filesystem-Native Hierarchical Ruleset Memory — A Zero-Dependency Harness Engineering Platform*

> *"Instead of cramming more context into a massive AI model, design the skeleton (structure) perfectly so that your reliance on the AI converges to zero."*
>
> An AI violated the "no console.log" rule 9 times. On the 10th time, we executed `mkdir brain/cortex/frontend/coding/禁console_log`. The folder itself became a physical rule injected into the system prompt. The counter (weight) hit 17. The AI never made that mistake again.
> 
> This is the true essence of **Harness Engineering** that NeuronFS pursues.

---

## TL;DR

**`mkdir` replaces the system prompt.** Folders are Neurons, paths are sentences, and files are synaptic weights.

### 4 Core Advantages over Legacy Systems

1. **Zero Cost:** While vector DBs like Mem0 or Letta incur server hosting fees to manage an agent's memory, NeuronFS uses your local OS filesystem natively, reducing your infrastructure cost to **$0**.
2. **Token Efficiency & Ultimate Manageability:** Finding and editing a specific rule inside a thousand-line text blob drives humans insane. In a hierarchical folder tree (`ls -R`), discovering, layering, and physically deleting rules is visually intuitive and instantly effective.
3. **Extreme Portability:** Built as a single Go binary with absolutely zero external dependencies. Drop it into any OS environment, and it runs immediately. It also seamlessly operates as an MCP (Model Context Protocol) server.
4. **Model-Agnostic Governance:** Quota limits force everyone to switch between AI models daily. `.cursorrules` only works in Cursor. `CLAUDE.md` only works in Claude Code. **NeuronFS compiles one brain into ALL formats simultaneously** — switch AI models freely without losing a single rule.

```bash
# Create a rule = Create a folder
mkdir -p brain/brainstem/禁fallback
touch brain/brainstem/禁fallback/1.neuron

# Compile = Auto-generate System Prompts (Cursor, Windsurf, Claude Desktop, etc.)
neuronfs ./brain --emit cursor   # → .cursorrules
neuronfs ./brain --emit claude   # → CLAUDE.md
neuronfs ./brain --emit all      # → ALL formats simultaneously — switch AI freely
```

---

## Installation (The One-Liner Quickstart)

Open-source standalone Go engine. Zero external dependencies.

```bash
# Mac / Linux
curl -sL https://raw.githubusercontent.com/rhino-acoustic/NeuronFS/main/install.sh | bash

# Windows (PowerShell)
iwr https://raw.githubusercontent.com/rhino-acoustic/NeuronFS/main/install.ps1 -useb | iex

# Initialize your offline brain (Creates the baseline scaffolding of 7 regions)
# ※ Select option [2] Master's Brain to clone standard premium engineering governance!
neuronfs --init ./my_brain        

export GROQ_API_KEY="<your-groq-api-key>"      # For Llama3 70B consolidation (Local Ollama fallback supported!)

# Compile and Run
neuronfs ./my_brain --emit all    # Compile system prompts
neuronfs ./my_brain --consolidate # Auto-merge neuron fragmentation via Llama 3 (Optional)
neuronfs ./my_brain --api         # Serve Dashboard at localhost:9090
```

---

## Table of Contents

| Section | Detail |
|---|---|
| [Core Architecture](#core-architecture) | Folders = Neurons, Paths = Rules, Counters = Weights |
| [Market Position](#market-position) | L1 Governance vs L2 IDE Rules vs L3 Agent Memory |
| [Brain Regions](#brain-regions) | 7 Regions, Subsumption Hierarchy, Hormone System |
| [Governance](#governance) | 3-Tier Injection, Bomb Circuit Breakers, Harness |
| [CLI & Architecture](#cli--architecture) | Auto-Loop, CLI, MCP, 30-file modular runtime |
| [Benchmarks](#benchmarks) | Performance vs RAG |
| [FAQ](#faq) | Expected objections |
| [Changelog](#changelog) | Recent updates |

---

## Core Architecture

> **Unix said "Everything is a file". We say: Everything is folders.**

| Concept | Biology | NeuronFS | OS Primitive |
|------|--------|----------|--------------|
| Neuron | Soma | Directory | `mkdir` |
| Rule | Firing Pattern | Full Path | Path String |
| Weight | Synaptic Strength | Counter Filename | `N.neuron` |
| Reward | Dopamine | Reward File | `dopamineN.neuron` |
| Block | Apoptosis | `bomb.neuron` | `touch` |
| Sleep | Synaptic Pruning | `*.dormant` | `mv` |
| Connection | Axon | `.axon` File | Symlink |
| Cross-Ref | Attention Residual | Axon Query-Key Match | Selective Aggregation |

### Path = Sentence

The path itself forms the natural language command. Depth equals specificity:

```
brain/cortex/NAS_File_Transfer/                    → Broad category
brain/cortex/NAS_File_Transfer/禁Copy-Item_UNC_no/  → Specific restriction
brain/cortex/NAS_File_Transfer/robocopy_large/     → Micro-context
```

### Kanji Micro-Opcodes

`禁` (1 character) = "NEVER_DO_THIS" (13 characters). Compressing semantics by 3-5x:

| Kanji | Meaning | Example |
|------|------|------|
| **禁** | FORBIDDEN | `禁fallback` |
| **必** | REQUIRED | `必Reference_KI` |
| **推** | RECOMMENDED | `推robocopy_for_large_files` |

### Auto-Evolution Pipeline

`.cursorrules` is a static file you are forced to edit manually. NeuronFS evolves autonomously:

1. **auto-consolidate**: Mitigates folder fragmentation. LLM (Groq or local) detects redundant error folders and merges them into a single Super Neuron.
2. **auto-neuronize**: Analyzes correction logs to auto-generate inhibitory (Contra) rules.
3. **auto-polarize**: Converts weak positive-form `use_X` rules into mathematically strong inhibitory (`禁X`) micro-opcode formats.

### Attention Residuals (Cross-Region Intelligence)

Inspired by [Kimi's Attention Residuals paper](https://arxiv.org/abs/2603.15031), NeuronFS implements **selective cross-referencing** via `.axon` connections. Instead of each brain region being an isolated silo:

- Each region's top neurons generate **query keywords**
- Target regions' neurons are scored against these keywords (**key matching**)
- Top 3 most relevant cross-region neurons auto-surface in `_rules.md`
- Governance neurons (禁/推) receive unconditional boost

```
reading ego/_rules.md now shows:
## 🔗 Axon 참조 (Attention Residuals)
- tools > 推: precise tool usage (c:65)    ← from cortex
- tools > 절대 금지: ls usage (c:57)       ← from cortex  
- ops > 절대 금지: general commands (c:48)  ← from cortex
```

### Autonomous Harness Cycle

Every 25 AI interactions, the harness engine (Node.js sidecar) automatically:

1. **Analyzes** correction logs for failure patterns
2. **Generates** 禁 (prohibition) and 推 (recommendation) neurons via Groq LLM
3. **Creates** `.axon` cross-links between related regions
4. The error that triggered the cycle **can never structurally recur**

---

## Market Position

> **NeuronFS is not AI Agent Memory. It is L1 Governance Infrastructure.**

```
L3: AI Agent Memory  (Mem0, Letta, Zep)         — Conversation memory, user profiling
L2: IDE Rules        (.cursorrules, CLAUDE.md)   — Static rule files, IDE-locked
L1: AI Governance    (NeuronFS) ◀── HERE         — Model-agnostic · Self-evolving · Consistent
```

### The Multi-AI Consistency Problem

In 2026, **every developer juggles multiple AI models** due to quota limits:

```
Morning: Claude (Opus quota burned) → Afternoon: Gemini → Evening: GPT
Claude learned "禁console.log" → Gemini doesn't know → Violation again → Pain
```

`.cursorrules` is Cursor-only. `CLAUDE.md` is Claude-only. **Switch models, lose your rules.**

NeuronFS solves this: **One brain → All formats simultaneously.**

| | .cursor/rules/ | NeuronFS |
|---|---|---|
| Multi-AI Support | ❌ IDE-locked | ✅ `--emit all` |
| Self-Evolution | ❌ Manual edit | ✅ auto-neuronize |
| Circuit Breaker | ❌ None | ✅ bomb.neuron |
| Sellable | ❌ Copy/paste files | ✅ **Curated brains as packages** |

### The WordPress Analogy

WordPress is free. Themes and plugins are paid. Similarly:
- **NeuronFS engine**: Free ($0) — MIT licensed
- **Curated Master Brains**: Premium — battle-tested governance packages for React, Next.js, Supabase, etc.

You can't sell a `.cursorrules` file. **You can sell a brain that evolved through 10,000 corrections.**

---

## Brain Regions

7 brain regions are layered via Brooks' Subsumption Architecture. **Lower P (Priority) layers always physically inhibit higher P layers.**

```
brainstem(P0) > limbic(P1) > hippocampus(P2) > sensors(P3) > cortex(P4) > ego(P5) > prefrontal(P6)
```

| Region | Priority | Role | Example |
|---------|---------|------|------|
| **brainstem** | P0 | Absolute principles | `禁fallback`, `禁Duplicate_SSOT` |
| **limbic** | P1 | Emotion filters | `dopamine_reward`, `adrenaline_rush` |
| **hippocampus** | P2 | Memory, session | `error_patterns`, `auto_search` |
| **sensors** | P3 | Environment constraints | `NAS/禁Copy`, `Design/sandstone` |
| **cortex** | P4 | Knowledge base (Max) | `react/hooks`, `backend/supabase` |
| **ego** | P5 | Tone & Persona | `expert_dry`, `korean_native` |
| **prefrontal** | P6 | Goals, Sprints | `current_sprint`, `long_term` |

---

## Governance

### Circuit Breaker (Bomb Neuron)

| Bomb Location | Result |
|-----------|------|
| brainstem (P0) | **Total Brain Halt**. GEMINI.md goes blank, effectively silencing the AI. |
| cortex (P4) | Renders brainstem~sensors only. Perfectly quarantines the specific tech (coding) region. |

A bomb.neuron does not 'beg in text' to stop doing something. It is a **hard emergency stop button that halts the rendering of the parent prompt entirely**. 
Unlocking it requires physics: `rm brain_v4/.../bomb.neuron`.

### Harness Protection

Robust local verification scripts:
- Pre-Git Lock Snapshot enforced before any destructive neuron consolidation.
- System-wide `SafeExec` (30-sec timeout) deadlock encapsulation.
- **Autonomous harness cycle**: Groq-powered 禁/推 neuron generation every 25 interactions.
- **Axon integrity**: Cross-region `.axon` links validated during brain scan.

---

## CLI & Architecture

### CLI Interface

```bash
neuronfs <brain> --emit <target>   # Compile prompts (gemini/cursor/claude/all)
neuronfs <brain> --consolidate     # Run Llama 3 70B consolidation
neuronfs <brain> --api             # Serve HTTP Dashboard (localhost:9090)
neuronfs <brain> --watch           # Watch for fsnotify changes
neuronfs <brain> --grow <path>     # Sprout a neuron
neuronfs <brain> --fire <path>     # Increase weight 
```

### Why Go?

A Single Binary. Zero `node_modules` or Python `venv` hell. Drop it anywhere, watch folders natively (`fsnotify`), and run it. The ultimate portability.

### Self-Referential Architecture

NeuronFS's core principle — **"Path = Sentence"** — is applied to its own codebase. File names alone tell you the entire system:

```
brain.go → Brain scan       inject.go → Injection       emit.go → Rule generation
lifecycle.go → Lifecycle     evolve.go → Evolution       similarity.go → Similarity
neuron_crud.go → CRUD        watch.go → Monitoring       supervisor.go → Management
```

This isn't just "good naming." It's a recursive, self-referential architecture that **proves its own philosophy through its code structure**. ~50 Go files, ~10K lines — yet any AI can rebuild full context in under 30 seconds by reading file names.

### Proof of Pain: Why You Need This

**Without NeuronFS:**
```
Day 1: AI violates "no console.log" → you correct it manually
Day 2: Switch to different AI (quota) → same violation again
Day 3: Repeat. Day 4: Repeat. Day 10: You lose your mind.
```

**With NeuronFS:**
```
Day 1: mkdir brain/cortex/禁console_log → violation permanently blocked
Day 2: Switch AI → --emit all → same brain, same rules
Day 10: Zero violations. The structure remembers what every AI forgets.
```

### Harness Engineering: The Next Paradigm

```
2023: Prompt Engineering   — "Write better prompts"
2024: Context Engineering  — "Feed better context"
2025: Harness Engineering  — "Design the skeleton so AI can't fail"
```

NeuronFS is a working implementation of **Harness Engineering** — not asking the AI to follow rules, but making it structurally impossible to break them. `bomb.neuron` doesn't beg; it halts. `禁` doesn't suggest; it structurally prevents.

---

## Benchmarks

| Metric | Result |
|------|------|
| Live neurons | **Varies** (7 regions, self-evolving · e.g. 500~3,000+) |
| Scan speed (1,000+ folders) | < 1 second |
| Rule addition latency | OS Native (`mkdir`), 0ms |
| Go source | **~50 files, ~10K lines** (modular) |
| Build time | **~8s** (single binary) |
| Local Disk Footprint | < 5MB (Pure text/folders) |
| Maintenance / API Cost | **$0** (No vector DB server required) |
| .jloot cartridge | Bible 66 books = 1.4MB, TF-IDF search 0ms |
| brainstem compliance rate | **94.9%** (18 violations per 353 injections) |

### Competitor Comparison

| | Hardcoded `.cursorrules` | Vector DB (RAG) | **NeuronFS (CLI)** |
|---|---|---|---|
| > 1000 rules | Token explosion / Maintenance Hell | ✅ Fast chunk retrieval | **✅ OS Folder Tree scattering** |
| Multi-AI | ❌ IDE-locked | ✅ API-based | **✅ `--emit all` (All formats)** |
| Infra Cost | Free | Server Cost ($70/mo) | **Free ($0)** |
| Auto-Growth | Impossible | Blackbox | **Visible folders (`mkdir`)** |
| Absolute Override | Must beg the AI | Limited | **✅ Circuit Breaker (bomb.neuron)** |
| Sellable | ❌ File copy | ❌ DB dump | **✅ Curated Brain packages** |

---


## Philosophy & Palantir Ontology

Why folders? Palantir's AIP (Artificial Intelligence Platform) success isn't just about using the smartest LLM; it's about connecting actions to an **Ontology** (a structured representation of reality).

NeuronFS shares a similar philosophy but scales it down for local filesystems. Instead of relying on an LLM to magically remember your 1000-line prompt, NeuronFS binds your business logic and restrictions into physical paths (cortex/frontend/no_console_log). 
We do not guarantee that the AI will follow the rules 100% (hallucinations exist). However, we lock the **prompt generation process** at the OS level so that human or AI errors cannot easily corrupt the core principles.

## Hybrid Memory Architecture (Overcoming Limitations)

**"We are not hostile to RAG; we are the L1 Governance Cache that controls RAG hallucinations."**

NeuronFS is not designed to compete with large-scale MSA (Microservices Architecture) environments or generalized Vector DBs. Instead, the architecture is intentionally separated to act as a **perfect hybrid complement**.

*   **Tier 1 & 2 (NeuronFS Deterministic Domination):** Absolute immutable rules (`brainstem`) and workflow constraints (`sensors`). Critical governance like "Force DB backups" or "Never use plain-text tokens" should never rely on 80% similarity probabilities. They require the **Hard Lock** of a 100% path-matched directory tree. Zero latency.
*   **Tier 3 (Vector DB / RAG Delegation):** Massive API documentation or years of accumulated error logs (`hippocampus`). Splitting ambiguous, enormous context into thousands of folders is over-engineering. We delegate this to existing RAG pipelines (like LlamaIndex) for flexibility.

In enterprise integration, before an AI agent recklessly scours a massive Vector DB, **NeuronFS (Tier 1 & 2) intervenes first to lay out the 'absolute constraints' as a guardrail.** OS folders serve as the L1 Instruction Cache; RAG serves as the L2 Main Memory.

### .jloot Cartridge System (Replace RAG Entirely)

For teams that don't need enterprise RAG at all, NeuronFS offers **.jloot cartridges** — hot-swappable knowledge packs loaded directly into memory.

```
Bible (66 books, ~31,000 verses) = 1.4MB .jloot → TF-IDF search: 0ms
Swap in law.jloot, medical.jloot, or your own domain — instant hot-reload.
```

| | Traditional RAG | .jloot Cartridge |
|---|---|---|
| Infrastructure | Vector DB + embeddings server | **Single file, memory-resident** |
| Search | Cosine similarity (~200ms) | **TF-IDF (0ms)** |
| Cost | $70+/mo hosting | **$0** |
| Swap domain | Re-index entire DB | **Mount different .jloot** |
| Hallucination defense | None built-in | **3-layer harness** (prompt + JSON constraint + source verification) |

No embeddings. No vector database. No server. A `.jloot` file is a compressed, indexed knowledge pack that your local Go binary loads into RAM.

### Telegram Bridge (Remote AI Monitoring)

Monitor and control your AI agents from anywhere via Telegram:

- **Outbound**: Transcript file watcher → Telegram (💬AI / 👤USER / 🧠THINK / ⚡CMD)
- **Inbound**: Telegram message → agent inbox → CDP injection into IDE
- `/rooms` — Live CDP window list, `/mount` — switch target IDE window
- Progressive message editing, echo-loop prevention, HTML formatting
- Managed by Go Supervisor (auto-restart on crash)

### vs Karpathy's Obsidian Knowledge Base (2026-04)

Karpathy recently shared an Obsidian-based knowledge system — also folder-based, also no vector DB. The difference is purpose:

| | Karpathy Obsidian | NeuronFS |
|---|---|---|
| Purpose | Knowledge retrieval (RAG replacement) | **AI behavior control (Governance)** |
| Weight | Obsidian app + Claude Code | **OS folders only (ultra-lightweight)** |
| Learning | None | Firing counters + dopamine/bomb |
| Priority | All documents equal | **P0 always overrides P6** |
| Search | Claude navigates index.md | **TF-IDF 0ms (.jloot)** |
| AI support | Claude Code only | **Multi-IDE simultaneous emit** |
| Evolution | Human curates wikis | **Groq autonomous growth** |

Karpathy's system is "how to make AI read your documents." NeuronFS is **"how to make AI obey your laws."**

---

## FAQ

**Q: "Isn't this just putting text in a system prompt? How is this different from throwing rules in a Notion document?"**  
**A:** Yes, it ends up as text. But hunting down and editing priority lines inside a 1000-line prompt spaghetti drives humans insane. We elevated the dimension of knowledge from "string space" into **"OS physical folder space"**.
Instead of appending `!!CRITICAL!!` 50 times in a prompt, NeuronFS offers **Structural Privilege Cascade** and **Physical Quarantine (bomb.neuron)**. One bomb disables the rendering of entire sub-hierarchies. It's a hard lock architecture, not a textual suggestion.

**Q: "If neurons exceed 1000, won't the context window explode?"**  
**A:** We have three defense layers: ① 3-Tier On-Demand Rendering (injects what is needed). ② 30-day Dormant pruning. ③ `--consolidate`, where Llama 3 (or a local model) melts overlapping garbage folders into a single optimized Super Neuron.

**Q: "MBTI Agents (ENTP, ISTJ)? Are you serious?"**  
**A:** Prompting "You are a QA engineer" traps an AI in a narrow box. By injecting broad **Personality (Tone/Aggression)** via MBTI into the `ego` folder, we've statistically proven the AI achieves greater dynamic problem-solving over rigid role-play constraints.

---

## Changelog

**v4.5 (2026-04-06)** — **Telegram Bridge**: Bidirectional AI monitoring via Telegram. .jloot cartridge hot-swap (Bible 66 books = 1.4MB, TF-IDF 0ms). Supervisor integration (7 processes, auto-restart). API/Emit module split. 10-model LLM hallucination benchmark. Karpathy Obsidian comparison.
**v4.4 (2026-04-05)** — **Extreme Modularization**: main.go 3,538→396 lines (-89%). 30 Go files, ~10,920 lines total. `go:embed` dashboard externalization. `/api/codemap` dashboard CODE panel. **Attention Residuals** cross-referencing via `.axon` connections. Autonomous harness cycle (Groq-powered 禁/推 auto-generation). 3,400+ neurons, 23 axons.
**v4.3 (2026-04-02)** — Ported Zero-Cost Consolidation (Llama 3 locally supported) & SafeExec hard lock integration.
**v4.2 (2026-03-31)** — Auto-Evolution pipeline complete. Groq correction parsing / Kanji opcode optimization.

Full history: [docs/CHANGELOG.md](docs/CHANGELOG.md)

---

MIT License · Copyright (c) 2026

Star if you agree. [Issue if you don't.](../../issues)


<!--
Easter Egg for the code divers:
Hey 666, easy - only the Word stands as absolute truth (777). 
This? It's just a well-organized folder built by someone who wanted to vibe-code without going insane.
-->
