<p align="center">
  <img src="https://img.shields.io/badge/Go-1.26+-00ADD8?style=flat-square&logo=go" />
  <img src="https://img.shields.io/badge/Infra-$0-brightgreen?style=flat-square" />
  <img src="https://img.shields.io/badge/Neurons-405-blue?style=flat-square" />
  <img src="https://img.shields.io/badge/Runewords-16-purple?style=flat-square" />
  <img src="https://img.shields.io/badge/Zero_Runtime_Dependencies-black?style=flat-square" />
  <img src="https://img.shields.io/badge/AGPL--3.0-green?style=flat-square" />
</p>

<p align="center">
  <img src="docs/vorq_harness.png" alt="vorq — AI must look up unknown tokens" width="800" />
</p>

<p align="center">
  <img src="docs/neuronfs_hero.jpg" alt="Folders ARE the Context — mkdir beats vector" width="600" />
</p>

<p align="center">
  <a href="https://dashboarddeploy-six.vercel.app/"><strong>Live 3D Dashboard Demo</strong></a>
</p>

<p align="center"><a href="README.ko.md">🇰🇷 한국어</a> · <a href="README.md">🇺🇸 English</a></p>

# NeuronFS
### *axiom > algorithm*
### Folder **is** a neuron. Everything else derives.

> AI violated "don't use console.log" 9 times.
> On the 10th, `mkdir 禁console_log` was born.
> On the 11th, AI asked: *"What does 'vorq' mean?"*
> **It never disobeyed again.**

---

**Quick Navigation:** [Problem](#the-problem-nobody-talks-about) · [30s Proof](#30-second-proof) · [5 Features](#5-killer-features) · [Comparison](#the-comparison) · [Getting Started](#getting-started) · [Benchmarks](#-benchmarks-41-industry-items) · [Novelty](#-novelty) · [Limitations](#limitations-honestly)

## The Problem Nobody Talks About

**2026 reality: quota limits force every developer to mix multiple AIs.**

```
Morning: Claude (Opus quota burnt) → Afternoon: switch to Gemini → Evening: switch to GPT
Claude's learned "禁console.log" rule → Gemini doesn't know → violation again → pain
```

`.cursorrules` is Cursor-only. `CLAUDE.md` is Claude-only. **Switch AI = rules evaporate.**

And the deeper problem — even within ONE session:

```
You: "Please read the codemap before editing code."
AI:  "Sure!" (skips it, starts coding immediately)
```

Text instructions are followed **~60% of the time**. That's not governance. That's hope.

---

## 30-Second Proof

```bash
git clone https://github.com/rhino-acoustic/NeuronFS.git && cd NeuronFS/runtime && go build -o neuronfs . && ./neuronfs --emit all
```

**Result:**
```
[EMIT] ✅ Cursor → .cursorrules
[EMIT] ✅ Claude → CLAUDE.md
[EMIT] ✅ Gemini → ~/.gemini/GEMINI.md
[EMIT] ✅ Copilot → .github/copilot-instructions.md
✅ 4 targets written. One brain. Every AI. Zero runtime dependencies. 18MB binary.
```

---

## We Attacked Ourselves — 10 Rounds

Before you trust us, watch us try to destroy ourselves.

| # | 🔴 Attack | 🔵 Defense | Verdict |
|---|-----------|------------|---------|
| 1 | **vorq is n=1 validated.** | Model-agnostic principle: unknown tokens force lookup in ALL transformers. | ⚠️ More testing needed |
| 2 | **vorq gets learned** once popular. | Replace tokens in 1 line, `--emit all`. Cost: 0. Time: 10s. Neologisms are disposable. | ✅ Defended |
| 3 | **Some AIs don't read _rules.md.** | Target is coding agents (Cursor/Claude Code/Gemini/Copilot). All auto-load rule files. | ✅ Defended |
| 4 | **P0 brainstem is still just text.** | Inherent limit of prompt governance. P0 at top = best positioning. | ⚠️ Acknowledged |
| 5 | **"mkdir vs vector" is overstated.** | L1/L2 separation. NeuronFS = deterministic rules (L1). RAG = semantic search (L2). | ✅ Defended |
| 6 | **Comparison table is biased.** | Partially. UX convenience rows (inline editing) should be added. Structural gaps are factual. | ⚠️ Acknowledged |
| 7 | **Bus factor = 1.** | Open source + zero dependencies = `go build` works in 2046. | ⚠️ Real risk |
| 8 | **`source:` freshness is manual.** | MVP. `--grow` auto-detection is on the roadmap. | ✅ Defended |
| 9 | **AGPL kills enterprise adoption.** | Deliberate. Core value is local execution. AGPL blocks SaaS cloning. | ✅ Defended |
| 10 | **`evolve` depends on AI — contradiction.** | `dry_run` is default. Evolution is assistance, not dependency. | ✅ Defended |

**Score: 7 defended · 3 acknowledged · 0 fatal.**

---

## 5 Killer Features

### 1. The Axiom — `Folder = Neuron`

One design decision generates the entire system:

```
Axiom: "A folder IS a neuron."
  → File path IS a natural language rule
  → Filename IS activation count (5.neuron = fired 5×)
  → Folder prefix IS governance type (禁=NEVER, 必=ALWAYS, 推=WHEN)
  → Depth IS specificity
  → OS metadata IS the embedding
  → mkdir IS learning
  → rm IS forgetting
```

### 2. vorq — Neologism Harness

Structured action directives with unknown tokens achieve behavioral compliance that natural language cannot.

| Attempt | Method | Compliance | Why |
|---|---|---|---|
| 1 | "Read the codemap" (natural language) | ~60% | AI "knows" this phrase → skips |
| 2 | "Mount cartridge" (proper noun) | ~65% | Meaning guessable → skips |
| 3 | "装カートリッジ 必装着" (kanji) | ~70% | AI infers 装=mount → skips |
| **4** | **"vorq cartridge 必vorq"** | **~95%+** | Unseen token → MUST look up (n=1 observed) |

4 Runewords: `vorq` (mount) · `zelk` (sync) · `mirp` (freshness check) · `qorz` (community search before tech decisions)

### 3. 7-Layer Subsumption Cascade (P0 → P6)

Lower priority **always** overrides higher. Physically.

```
brainstem(P0) > limbic(P1) > hippocampus(P2) > sensors(P3) > cortex(P4) > ego(P5) > prefrontal(P6)
     ↑ absolute laws    ↑ emotions    ↑ memory    ↑ environment  ↑ knowledge  ↑ persona  ↑ goals
```

### 4. 3-Tier Governance (ALWAYS / WHEN → THEN / NEVER)

Folder prefixes auto-classify into three enforcement tiers at `emit` time:

```
禁hardcoding       → 🔴 NEVER   (absolute prohibition, immune to decay/prune/dedup)
必go_vet실행        → 🟢 ALWAYS  (mandatory on every response)
推community_search → 🟡 WHEN coding/tech decision → THEN search community first
```

### 5. One Brain, Every AI

```bash
neuronfs --emit all
→ .cursorrules + CLAUDE.md + GEMINI.md + copilot-instructions.md
```

Switch AI tools freely. Your rules never evaporate. One brain governs all.

---

## The Comparison

| # | | `.cursorrules` | Mem0 / Letta | RAG (Vector DB) | **NeuronFS** |
|---|---|---|---|---|---|
| 1 | **Rule accuracy** | Text = easily ignored | Probabilistic | ~95% | **100% deterministic** † |
| 2 | **Behavioral compliance** | ~60% (text advisory) | ~60% | ~60% | **~95%+ (vorq harness, n=1 observed)** ‡ |
| 3 | **Multi-AI support** | ❌ Cursor-only | API-dependent | ✅ | **✅ `--emit all` → every IDE** |
| 4 | **Priority system** | ❌ Flat text | ❌ | ❌ | **✅ 7-layer Subsumption (P0→P6)** |
| 5 | **Self-evolution** | Manual edit | Black box | Black box | **🧬 Autonomous (Groq LLM)** |
| 6 | **Kill switch** | ❌ | ❌ | ❌ | **✅ `bomb.neuron` halts region** |
| 7 | **Cartridge freshness** | ❌ Manual | ❌ | ❌ | **✅ `source:` mtime auto-check** |
| 8 | **Encrypted distribution** | ❌ | Cloud-dependent | Cloud-dependent | **✅ Jloot VFS cartridges** |
| 9 | **Infrastructure cost** | Free | $50+/mo | $70+/mo GPU | **$0 (local OS)** |
| 10 | **Dependencies** | IDE-locked | Python+Redis+DB | Python+GPU+API | **Zero runtime (single binary)** |
| 11 | **3-Tier governance** | ❌ | ❌ | ❌ | **✅ ALWAYS/WHEN/NEVER auto-classify** |
| 12 | **OOM protection** | ❌ | ❌ | ❌ | **✅ Auto-truncate on context overflow** |
| 13 | **Industry benchmark coverage** | 0/41 | ~8/41 | ~6/41 | **35/41 (85%)** |

> † **Rule accuracy** measures different layers: Mem0/RAG ~95% = "LLM follows retrieved rules" (IFEval). NeuronFS 100% = "rules are faithfully generated into system prompt" (BM-1).
>
> ‡ **Behavioral compliance** ~95%+ is based on developer observation (n=1). Principle is model-agnostic (unknown tokens force lookup).

---

## Getting Started

**One-Liner (Linux/macOS/PowerShell 7+):**
```bash
git clone https://github.com/rhino-acoustic/NeuronFS.git && cd NeuronFS/runtime && go build -o neuronfs . && ./neuronfs --emit all
```

**Step by Step:**
```bash
# 1. Clone & build
git clone https://github.com/rhino-acoustic/NeuronFS.git
cd NeuronFS/runtime
go build -o neuronfs .          # → single binary, zero runtime dependencies

# 2. Create a rule
./neuronfs --grow cortex/react/禁console_log

# 3. Compile brain
./neuronfs --emit all
```

---

## 📊 Benchmarks (41 Industry Items)

```bash
cd runtime && go test -v -run "TestBM_" -count=1 .
```

| Test | What | Result | Industry Standard |
|------|------|--------|-------------------|
| **BM-1** | Rule Fidelity (AgentIF CSR) | **100%** (5/5) | IFEval SOTA: 95% |
| **BM-2** | Scale Profile (5K neurons) | **2.5s** | Mem0: 125ms (RAM) |
| **BM-3** | Similarity Accuracy | **P=1.0** F1=0.74 | Vector DB: P≈0.85 |
| **BM-4** | Lifecycle (禁 protection) | **30/30 100%** | Industry Unique |
| **BM-5** | Adversarial QA (LOCOMO) | **5/5 rejected** | SQuAD 2.0 style |
| **BM-6** | Production Latency | **p50=202ms p95=268ms** | Mem0 p50: 75ms |
| **BM-7** | Multi-hop Planning (MCPBench) | **grow→fire→dedup→emit ✅** | Tool chaining |

| Benchmark | Items | ✅ Covered |
|-----------|-------|-----------|
| MemoryAgentBench (ICLR 2026) | 4 | **4** |
| LOCOMO | 7 | **4** + 2 N/A |
| AgentIF | 6 | **6** |
| MCPBench | 6 | **5** + 1 partial |
| Mem0/Letta | 8 | **6** + 1 N/A |
| **NeuronFS-only** | **10** | **10** |
| **Total** | **41** | **35 (85%)** |

---

## 🧬 Novelty

| System | Description | Why it's new |
|--------|-------------|-------------|
| **Folder=Neuron** | `mkdir` = neuron creation. Path = rule. | First to use OS folders as the cognitive unit. |
| **vorq Opcodes** | 16 runes (12 kanji + 4 ASCII neologisms). | Constructed micro-language for behavioral control. |
| **3-Tier emit** | 禁/必/推 → NEVER/ALWAYS/WHEN auto-injection. | Rules are "installed" into LLMs, not "suggested." |
| **Filename=Counter** | `5.neuron` = 5 activations. | Metadata IS the filename. Zero-query state. |
| **bomb circuit breaker** | 3 failures → P0 halts cognitive region. | Cognitive-level circuit breaker with physical silencing. |
| **OOM Protection** | `applyOOMProtection()` auto-truncates. | No other system prevents its own context overflow. |

---

## Official Wiki

> **[NeuronFS Official Wiki](https://github.com/rhino-acoustic/NeuronFS/wiki)** — Korean original, English titles

---

## License
AGPL-3.0 License · Copyright (c) 2026

> *Created by PD (Park Jung-geun) — rubisesJO777*
> *173 Go source files, 660+ neurons, Phase 56. Single binary. Zero runtime dependencies.*
