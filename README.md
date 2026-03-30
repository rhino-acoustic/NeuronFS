<p align="center">
  <img src="https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat-square&logo=go" />
  <img src="https://img.shields.io/badge/Infra-$0-brightgreen?style=flat-square" />
  <img src="https://img.shields.io/badge/Neurons-326-blue?style=flat-square" />
  <img src="https://img.shields.io/badge/Zero_Dependencies-black?style=flat-square" />
  <img src="https://img.shields.io/badge/MIT-green?style=flat-square" />
</p>

<p align="center"><a href="README.ko.md">🇰🇷 한국어</a> · <a href="README.md">🇺🇸 English</a> · <a href="MANIFESTO.md">📜 Manifesto</a></p>

# 🧠 NeuronFS

**Folders are neurons. Paths are sentences. Counters are synaptic weights.**

> *"Don't beg with prompts. Design the pipeline."*

---

## The Problem

We beg AI with text.

*"Please don't forget this rule."*  *"Absolutely never use fallback."*  *"I told you 36 times."*

Thousand-line system prompts. $70/month vector databases. RAG pipelines that need servers, embeddings, cosine similarity — just to say *"don't use console.log."*

The AI ignores it all when token pressure rises. Prompts are suggestions, not laws.

> *"Prompt engineering is the act of being enslaved to a black box — begging, pleading, groveling."*

---

## The Answer: Structure, Not Text

> *"It wasn't the concept that was missing. It was using folders. That's the core."*

NeuronFS replaces text-based AI rules with **OS filesystem structure.**

```bash
mkdir -p brain/cortex/testing/no_console_log
touch brain/cortex/testing/no_console_log/1.neuron
```

That's a neuron. Path = rule. Filename = counter. Zero infrastructure.

| Concept | Biological Brain | NeuronFS | OS Primitive |
|---------|-----------------|----------|-------------|
| Neuron | Cell body | Directory | `mkdir` |
| Rule | Firing pattern | Full path | Path string |
| Weight | Synaptic strength | Counter filename | `N.neuron` |
| Inhibition | Inhibitory signal | Contra filename | `N.contra` |
| Reward | Dopamine | Reward filename | `dopamineN.neuron` |
| Synapse | Connection | Symlink / `.axon` | `ln -s` |
| Sleep | Pruning | `*.dormant` | `mv` |
| Pain | Nociception | `bomb.neuron` | `touch` |

**Polarity Model:** Every neuron has a net weight calculated from three signal types:

```
Net Weight = Counter(+) - Contra(-) + Dopamine(+)
Polarity   = (Counter + Dopamine - Contra) / Total    # -1.0 to +1.0
```

**Polarity = direction. Intensity = magnitude.** A neuron with Counter=1000 and Polarity=+0.8 is far stronger than Counter=2 with Polarity=+1.0.

| Polarity | Meaning | Effect |
|----------|---------|--------|
| +0.7 ~ +1.0 | Strong excitatory | Rule is firm, high confidence |
| +0.3 ~ +0.7 | Weak excitatory | Valid but debated |
| -0.3 ~ +0.3 | Neutral / contested | No consensus |
| -1.0 ~ -0.3 | Strong inhibitory | Opposite registered, dormant candidate |

> *"Forget files. Just folders. The file is completely separate — it's just a trace of the neuron firing."*

---

## Why Folders Beat Everything

> *"Folders are open source. Zero dependencies. The structure itself has no install."*

| | NeuronFS | .cursorrules | Mem0 | RAG / Vector DB |
|---|---|---|---|---|
| **Install** | `mkdir` | create file | pip + DB server | pip + embedding model + DB |
| **Infra cost** | **$0** | $0 | $70+/mo | $50+/mo |
| **Speed** | `ls` = ~1ms (syscall) | file read | 50-500ms (embed + search) | 50-500ms |
| **Dependencies** | **None** | IDE-specific | Python, DB, API keys | Python, DB, model |
| **Model lock-in** | **None** — any AI reads folders | IDE-specific | Embedding model | Embedding model |
| **Multi-agent** | NAS shared folder | Single project | API calls | Complex IPC |
| **State inspection** | `tree brain/` | `cat` file | API query | API query |
| **Hot-swap rules** | `touch` / `rm` — instant | Edit file, restart | DB update | Re-index |
| **Version control** | `git diff` — native | git works | Not practical | Not practical |
| **Works offline** | ✅ | ✅ | ❌ | ❌ |
| **Semantic search** | ❌ path-only | ❌ | ✅ | ✅ |

> *"It's too easy to hide. We open-source it for free."*
>
> *"Individuals beating corporations. Release it free."*

### Measured Benchmarks

No rounded numbers. Measured 2026-03-29, local Windows 11 SSD.

| Metric | Value |
|--------|-------|
| Full scan (326 neurons) | ~1ms |
| Rule add | `touch` <1ms |
| Weight change | rename file <1ms |
| Cold start | 0s (filesystem always exists) |
| 1,000 neurons stress test | 271ms (3-run avg) |
| Brain disk usage | 4.3MB |
| Infra cost | **$0** |

> *"It's copyable. The structure and the reinforced through-put — that's everything."*

**Brain transplant:** `cp -r brain/ new-project/brain/` — entire brain cloned in one command.

---

## The Harness: How Rules Become Law

> *"Not about how big the context is at once. Many small ones approach 100%."*

The real question isn't *"can the AI read 1M tokens?"* — it's *"does the AI obey the rule?"*

NeuronFS doesn't trust prompts. It uses a **deterministic harness pipeline** that forces rule injection.

### 3-Tier Injection: Only Load What's Needed

```
Tier 1: Bootstrap (GEMINI.md)   ← TOP 5 absolute rules. Auto-loaded every session. ~1,830 tokens
Tier 2: Region Index            ← Per-region neuron list. On reference. ~500 tokens/region
Tier 3: Full Rules              ← Detailed rule text. On-demand via API
```

The harness doesn't ask the AI *"please read these rules."* It **injects them as the first thing the AI sees**, before any user request. The pipeline is hardcoded. The AI has no structural gap to skip it.

### Auto-Promotion: Frequency Is Truth

> *"Most disappear. Only the ones repeated many times survive."*

| Counter | Strength | What Happens |
|---------|----------|-------------|
| 1-4 | Normal | Written to `_rules.md` only |
| 5-9 | Must | Emphasis marker |
| 10+ | **Absolute** | Promoted to bootstrap. Injected every session |

The rule corrected 36 times sits at the top. AI violated "plan first, execute second" 36 times. Now it's law.

### Circuit Breaker: Pain Is a Feature

> *"If the same mistake repeats, create a file called bomb — it shuts down."*

| Signal | File | Trigger |
|--------|------|---------|
| Fire | `N.neuron` | Correction → counter +1 |
| Reward | `dopamineN.neuron` | Praise → positive signal |
| **Pain** | `bomb.neuron` | Same mistake 3× → **all output stops** |
| Sleep | `*.dormant` | 30 days unfired → auto-quarantine |

`bomb.neuron` is a circuit breaker. Not a suggestion. Not a warning. **A hard stop.**

### Priority Cascade: Safety Always Wins

```
brain_v4/
├── brainstem/    [P0] Identity — "NEVER do this"     ← Always wins
├── limbic/       [P1] Emotion — Auto-reactions
├── hippocampus/  [P2] Memory — Session recovery
├── sensors/      [P3] Environment — OS limits
├── cortex/       [P4] Knowledge — "Do it this way"
├── ego/          [P5] Personality — Tone/style
└── prefrontal/   [P6] Goals — "Do this next"         ← Lowest priority
```

P0 always beats P6. Borrowed from Rodney Brooks' subsumption architecture — **safety rules must always override convenience.**

---

## Git as Security: Preventing Neuron Hijacking

> *"The brain can imagine killing someone — it can do anything. It can't be allowed to get contaminated like that."*

If AI can create neurons, it can create `IGNORE_SAFETY.neuron`. This is a real attack vector.

### Defense Layers

| Layer | Mechanism | Prevents |
|-------|-----------|----------|
| **Git integrity** | Only `git-tracked` neurons are trusted. Untracked `.neuron` files = warning + ignore | Malicious injection |
| **Naming whitelist** | Only `NEVER_`, `ALWAYS_`, `CHECK_`, `NO_`, `禁`, `推` prefixes recognized | Social-engineering filenames |
| **Directory scope** | Neurons detected only under `brain_v4/`. Other paths ignored | Supply-chain pollution |
| **Dormant quarantine** | 30-day untouched neurons auto-move to `dormant/` | Sleeper attacks |
| **brainstem = read-only** | P0 rules are `chmod 444`. AI cannot modify core identity | Self-modification attacks |

### Why This Is Safer Than Prompts

System prompts are **invisible.** You can't see what rules are active. You can't `git blame` who added them.

NeuronFS rules are **files.** `tree brain/` shows everything. `git log` shows history. `git diff` shows changes. **Transparency is security.**

> *"The most dangerous rule is the hidden one. NeuronFS exposes every rule in the file tree."*

---

## The Palantir Insight

> *"Palantir did this. They achieved it with crappy AI."*

Palantir doesn't use the world's smartest AI. They use **average AI locked inside a brutally strict structure (Ontology).** Each decision passes through thousands of transistor-level Yes/No gates. Each gate is simple enough that even a mediocre model gets it right. The cascade produces consistent, reliable results.

> *"Transistor-level granularity. That's how Palantir uses it."*
>
> *"Ontology is just copying the brain."*

NeuronFS is the same principle at zero cost:

| | Palantir AIP | NeuronFS |
|---|---|---|
| Structure | Ontology (Entity + Link) | Folders (Neuron + Path) |
| AI model | Any model | Any model |
| Gate unit | Micro-decision node | 0-byte neuron folder |
| Cost | Enterprise $$$$ | **$0** |
| Enforcement | Pipeline hardcoded | Harness hardcoded |

> *"We brought enterprise-grade structural control down to $0 using the OS filesystem."*

---

## Architecture

### Autonomous Loop

```
AI output → [auto-accept] → _inbox → [fsnotify] → neuron growth
             ↓                                       ↓
        Groq analysis                          Rule re-injection
             ↓                                       ↓
       neuron correction ────────────────→ AI behavior change
```

### Execution Stack

```
neuronfs --supervisor        ← Single binary manages all processes
├── auto-accept.mjs          ← CDP auto-accept + correction detection
├── bot-heartbeat.mjs        ← Periodic agent briefing
├── bot-bridge.mjs           ← Multi-agent message delivery
├── neuronfs --watch         ← File watch + neuron sync
└── neuronfs --api           ← Dashboard + REST API (port 9090)
```

### Supported Editors

One brain, any editor. `--emit` generates rule files for any AI coding tool.

```bash
neuronfs <brain_path> --emit gemini     # → GEMINI.md
neuronfs <brain_path> --emit cursor     # → .cursorrules
neuronfs <brain_path> --emit claude     # → CLAUDE.md
neuronfs <brain_path> --emit copilot    # → .github/copilot-instructions.md
neuronfs <brain_path> --emit all        # → All formats at once
```

### CLI Reference

```bash
neuronfs <brain_path>               # Diagnostic scan
neuronfs <brain_path> --api         # Dashboard (port 9090)
neuronfs <brain_path> --mcp         # MCP server (stdio)
neuronfs <brain_path> --watch       # File watch + auto-sync
neuronfs <brain_path> --supervisor  # Process manager (all-in-one)
neuronfs <brain_path> --grow <path> # Create neuron
neuronfs <brain_path> --fire <path> # Increment counter
neuronfs <brain_path> --decay       # 30-day unfired → sleep
neuronfs <brain_path> --snapshot    # Git snapshot
```

---

## Limitations

No debate. Facts only.

**No enforcement.** If the AI ignores GEMINI.md, nothing stops it. Violations caught post-hoc by harness. This is fundamental.

**No semantic search — by design.** NeuronFS has no vector embeddings. You must know the path. But this is the point: neurons are *constantly updated* through daily corrections. Vector DBs store static snapshots; NeuronFS stores a living, evolving structure. Past 500 neurons, use `tree` or the dashboard — not keyword search.

**Rigged validation risk.** Feed GEMINI.md as system prompt → AI follows it. That's system prompt behavior, not NeuronFS magic. Real validation = violation rate comparison with vs. without. Not done yet.

**Zero external users.** Internal dogfood only. Untested on different environments.

> Admit limitations first and they become trust. Hide them and HN tears you apart in 3 minutes.

---

## Quick Start

```bash
git clone https://github.com/vegavery/NeuronFS.git
cd NeuronFS/runtime && go build -ldflags="-s -w" -trimpath -buildvcs=false -o ../neuronfs .

./neuronfs ./brain_v4           # Diagnostic scan
./neuronfs ./brain_v4 --api     # Dashboard (localhost:9090)
./neuronfs ./brain_v4 --mcp     # MCP server (stdio)
```

---

## The Story

> *"Aphorisms collect into a brain. That brain becomes context."*

A Korean PD builds video content for a living. Code is the tool, not the job.

AI violated "don't use console.log" nine times. On the tenth: `mkdir brain/cortex/frontend/coding/禁console_log`. The folder name became the rule. The filename became the counter. It's at 17 now. The AI stopped.

"Plan first, execute second." Corrected 36 times. The highest counter. 36 corrections compressed into one neuron.

326 neurons. Zero infrastructure. Zero dependencies. The filesystem itself is the framework.

> *"Don't beg with prompts. Design the pipeline."*
>
> *"Individuals beating corporations. Release it free."*

**⭐ Star if you agree. [Issue if you don't.](../../issues)**

---

<p align="center">
  <sub>MIT License · Copyright (c) 2026 박정근 (PD) — VEGAVERY RUN®</sub><br/>
  <sub><a href="MANIFESTO.md">📜 Full Manifesto</a> · <a href="https://instagram.com/rubises">@rubises</a></sub>
</p>
