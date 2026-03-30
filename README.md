<p align="center">
  <img src="https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat-square&logo=go" />
  <img src="https://img.shields.io/badge/Infra-$0-brightgreen?style=flat-square" />
  <img src="https://img.shields.io/badge/Neurons-326-blue?style=flat-square" />
  <img src="https://img.shields.io/badge/Zero_Dependencies-black?style=flat-square" />
  <img src="https://img.shields.io/badge/MIT-green?style=flat-square" />
</p>

<p align="center">
  <img src="docs/dashboard.png" alt="NeuronFS Dashboard — 3D Brain Visualization" width="800" />
  <br/>
  <sub>Live dashboard: 329 neurons across 7 regions. Real-time activation monitoring.</sub>
</p>

<p align="center"><a href="README.ko.md">🇰🇷 한국어</a> · <a href="README.md">🇺🇸 English</a> · <a href="MANIFESTO.md">📜 Manifesto</a></p>

# 🧠 NeuronFS

**Folders are neurons. Paths are sentences. Counters are synaptic weights.**

> *"Don't beg with prompts. Design the pipeline."*

### Contents

| | Section | What You'll Learn |
|---|---|---|
| 🔴 | [The Problem](#the-problem) | Why text-based AI rules fail at scale |
| 💡 | [Structure, Not Text](#the-answer-structure-not-text) | How `mkdir` replaces thousand-line prompts |
| 🏆 | [Why Folders Beat Everything](#why-folders-beat-everything) | $0 infra vs $70/mo vector DBs — benchmarks included |
| ⚖️ | [The Harness](#the-harness-how-rules-become-law) | 3-tier injection, circuit breakers, brainstem protection |
| 🔮 | [The Palantir Insight](#the-palantir-insight) | How a $100B company proved "dumb AI + strict structure" works |
| 🏗️ | [Architecture](#architecture) | Autonomous loop, execution stack, CLI reference |
| ⚠️ | [Limitations](#limitations) | Honest about what doesn't work |
| 📖 | [The Story](#the-story) | Why aphorisms make the brain smarter than rules |

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

<p align="center">
  <img src="docs/neuronfs_tree.png" alt="NeuronFS Folder Tree — Brain Architecture" width="720" />
</p>

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

### Axons — Cross-Region Wiring

Neurons live inside brain regions. **Axons** connect regions to each other — just like biological axons wire brain areas into a layered network.

```
brainstem ←→ limbic ←→ hippocampus ←→ sensors ←→ cortex ←→ ego ←→ prefrontal
  (P0)         (P1)       (P2)          (P3)       (P4)     (P5)      (P6)
```

Implementation: each `.axon` file in a region folder declares a target:

```bash
brain_v4/brainstem/cascade_to_limbic.axon      → "limbic"
brain_v4/limbic/cascade_from_brainstem.axon     → "brainstem"
brain_v4/cortex/shortcut_to_hippocampus.axon    → "hippocampus"
```

**What axons actually do:**
- **Cascade suppression:** Lower P suppresses higher P. If `brainstem` has a bomb, the axon to `limbic` signals "stop all higher processing."
- **Context routing:** `sensors` axon to `cortex` means "check environment constraints before applying knowledge."
- **Skill linking:** `cortex/skills/supanova/ref.axon` → links to external `SKILL.md` files, bridging folder structure to external tools.

Currently **16 axons** wire the 7 brain regions into a layered cascade — including 3 shortcut paths for emergency routing (e.g., `limbic → cortex` for adrenaline-level urgency).

**Why axons matter at scale:** At 326 neurons, you can eyeball the tree. At 3,000 neurons across 20 regions, you can't. Without axons, regions become silos — the security team's neurons never inform the deployment team's neurons. Axons are the explicit contracts that say "before you apply cortex/deployment rules, check sensors/compliance first." As the brain grows, **axons become more important than individual neurons** — they are the topology of how your organization thinks.

> *"Neurons are what you know. Axons are how knowledge flows between domains."*

---

## Why Folders Beat Everything

> *"Folders are open source. Zero dependencies. The structure itself has no install."*

| | NeuronFS | .cursorrules | Mem0 / RAG |
|---|---|---|---|
| **Install** | `mkdir` | create file | pip + DB + model |
| **Cost** | **$0** | $0 | $50-70+/mo |
| **Speed** | ~1ms (syscall) | file read | 50-500ms |
| **Dependencies** | **None** | IDE-locked | Python, DB, API |
| **Multi-agent** | NAS shared folder | ❌ | Complex IPC |
| **Version control** | `git diff` native | git works | ❌ |

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

**This was proven years ago.** Palantir built a $100B+ company not by waiting for AGI, but by perfecting the *structure around* ordinary AI. NeuronFS applies the same lesson: don't make the model smarter — make the pipeline stricter.

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

**Current production environment:** Windows 11, Google Antigravity (DeepMind), 326 neurons, daily operation since 2026-01.

### How It Works with Antigravity

NeuronFS is designed to work with **Google Antigravity** (DeepMind's agentic AI coding assistant) on Windows. The integration runs through Chrome DevTools Protocol (CDP):

1. **CDP Auto-Accept** — A Node.js script connects to Antigravity via `localhost:9000`, monitors for AI-generated actions, and auto-accepts them
2. **Real-time Transcript Scraping** — PD/AI conversations are captured and buffered
3. **Idle-triggered Groq Analysis** — When the user goes idle for 5 minutes, Groq LLaMA analyzes the transcript buffer for corrections, violations, and insights
4. **Automatic Neuron Growth** — Corrections are written to `_inbox/corrections.jsonl`, which `neuronfs --supervisor` picks up via `fsnotify` and converts to new neurons

> *The auto-accept CDP integration will be published as a separate repository. Stay tuned.*

---

## Competitors Comparison

|  | .cursorrules | Mem0 | Letta (MemGPT) | **NeuronFS** |
|--|-------------|------|----------------|-------------|
| 1000+ rules | Token overflow ❌ | ✅ (vector DB) | ✅ (tiered memory) | ✅ (folder tree) |
| Infra cost | $0 | Server/API $$$ | Server $$$ | **$0** |
| Switch AI | Copy file | Migration | Migration | **No change** (it's folders) |
| User owns rules | ✅ | ❌ (AI extracts) | ❌ (agent manages) | **✅ (mkdir)** |
| Self-growth | ❌ | ✅ | ✅ | **✅ (correction→neuron)** |
| Immutable guardrail | ❌ | ❌ | ❌ | **✅ (brainstem chmod 444)** |
| Auditability | git diff | Query | Logs | **ls -R** |

### What the Output Looks Like

**`brainstem/_rules.md`** — Auto-generated from folder structure. chmod 444. Untouchable:

```markdown
# 🛡️ BRAINSTEM — Conscience/Instinct
Active: 24 | Dormant: 0 | Activation: 138

- 절대 **禁영어사고 한국어로 생각하고 대답** (13)
- **禁뉴런구조 임의변경** (3)
- **토론말고 실행** (2)
- **禁SSOT 중복** (2)
- **PD단답이 시스템의 근원** (1)
```

**`cortex/_rules.md`** — Knowledge/skills. Grows daily:

```markdown
# 🧠 CORTEX — Knowledge/Skills
Active: 212 | Dormant: 0 | Activation: 225

- **agent ops** (0)
  - 반드시 **no pm solo work** (6)
  - **share source context** (3)
- **backend** (0)
  - **supabase** → **RLS 항상켜기** (1)
```

**`neuronfs --emit gemini` → `GEMINI.md`** — All neurons compiled into one system prompt:

```markdown
## 🛡️ brainstem (P0 — 절대 불변)
- 절대 **禁영어사고**: 한국어로 생각하고 대답 [13회 교정]
- **禁뉴런구조 임의변경** [3회 교정]
- **토론말고 실행** [2회 교정]

## 🧠 cortex (P4 — 지식/기술)
- agent ops: 반드시 **no pm solo work** [6회]
- backend/supabase: **RLS 항상켜기** [1회]
...
(326 neurons → ~8KB system prompt)
```

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

### Why Go

NeuronFS runtime is a single Go binary. No Python. No Node.js runtime dependency. No Docker.

| Why | Benefit |
|-----|---------|
| **Single binary** | `go build` → one `neuronfs.exe`. Copy to any machine. Done. |
| **Cross-compile** | `GOOS=linux go build` from Windows. ARM, x86, Mac — all from one machine |
| **fsnotify native** | Real-time file watch without polling. Neuron changes detected in <10ms |
| **goroutines** | Supervisor manages 5+ child processes concurrently. Zero thread overhead |
| **Zero GC pressure** | Folder scanning 10,000 neurons takes <50ms. No JVM warmup |
| **Static linking** | No `.dll`, no `node_modules/`, no `pip install`. The binary IS the runtime |

The philosophy: **if the brain is zero-infrastructure, the runtime should be too.** One binary, one folder, one brain.

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
git clone https://github.com/rhino-acoustic/NeuronFS.git
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

### Why It Was Built

NeuronFS was born from a real need: **building a company knowledge base.** VEGAVERY RUN® operates across CRM, video production, brand design, and e-commerce — all managed by one PD with AI. The knowledge was scattered across conversations, documents, and people's heads.

The goal was simple: when any AI — current or future — starts working on VEGAVERY, it should instantly know every rule, every preference, every mistake that was already made. Not from a 50-page onboarding doc. From the folder structure itself.

### Why Aphorisms Matter

Every correction is a **short answer**. Users don't write essays — they snap:
- *"Don't use fallback"* → `mkdir 禁fallback` → 1 neuron
- *"Don't beg with prompts. Design the pipeline."* → 1 neuron

**One short answer = one neuron.** Accumulate them and you get a folder tree. The folder tree is the brain.

But here's the key: if that short answer is an **aphorism** — a principle, a philosophy — the brain gets qualitatively smarter.

- *"Don't use console.log"* → technical fix. Good.
- *"Don't beg with prompts. Design the pipeline."* → philosophy. **The brain's direction changes.**
- *"Aphorisms collect into a brain. That brain becomes context."* → meta-cognition. The brain understands itself.

Same 326 neurons. But a brain filled with aphorisms is **fundamentally different** from a brain filled with rules. The folder names carry the user's thinking.

### Imagine

**A new developer joins your team.** Instead of reading 200 pages of docs, they clone `brain/` — and their AI already knows "don't use console.log", "plan before execute", "never touch the production DB directly." 326 corrections compressed into a folder tree that any AI reads in 1ms.

**Your company switches from Claude to Gemini.** Nothing changes. The brain is folders. Any AI reads it. Zero migration cost.

**A hospital deploys NeuronFS for medical AI.** `brainstem/NEVER_hallucinate_dosage/` — chmod 444. No AI, no matter how "creative," can override the rule. The guardrail is a folder, not a prayer in a system prompt.

**A law firm shares neurons across offices.** NAS shared folder. Tokyo and Seoul read the same `brain/legal/contracts/` neurons. `robocopy` syncs once per hour. Zero infrastructure.

**10 AI agents manage a factory.** Each agent forks `brain/base/` and evolves specialized neurons for their domain. Quality control agent has 500 neurons. Logistics agent has 300. They all share `brainstem/` — the constitutional rules.

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
