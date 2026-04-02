<p align="center">
 <img src="https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat-square&logo=go" />
 <img src="https://img.shields.io/badge/Infra-$0-brightgreen?style=flat-square" />
 <img src="https://img.shields.io/badge/Neurons-340+-blue?style=flat-square" />
 <img src="https://img.shields.io/badge/Zero_Dependencies-black?style=flat-square" />
 <img src="https://img.shields.io/badge/MIT-green?style=flat-square" />
</p>

<p align="center">
 <img src="docs/dashboard.png" alt="NeuronFS Dashboard — 3D Brain Visualization" width="800" />
 <br/>
 <a href="https://dashboarddeploy-six.vercel.app/"><strong> Live 3D Dashboard Demo</strong></a>
</p>

<p align="center"><a href="README.ko.md"> 한국어</a> · <a href="README.md"> English</a> · <a href="MANIFESTO.md"> Manifesto</a> · <a href="LIFECYCLE.md"> Lifecycle</a></p>

> **v4.3 (2026-04-02) — Auto-Consolidation & SafeExec Harness**
>
> **Completed:**
> - **Auto-Consolidation:** The neuron merging pipeline (`--consolidate`) now supports external APIs (e.g., Groq Llama 3) as an optional integration to reduce maintenance cost. Local LLM fallback is structurally supported.
> - **Harness Protection:** Integrated a `Pre-Git Lock` snapshot before neuro-destructive operations and embedded a global `SafeExec (30s Timeout)` to mitigate infinite terminal deadlocks.
> - **Emit engine optimization:** Dual-path sync (local `.gemini/` + global `~/.gemini/`), strength prefix batch rendering
>
> **In Progress:** Local LLM (`Ollama` / `localhost:11434`) offline integration, OS auto-start registration
>
> **Breaking:** `brain_v4/` excluded from git — users must `neuronfs --init` to create their own brain.
>
> Full changelog: [LIFECYCLE.md](LIFECYCLE.md) · [LIFECYCLE_EN.md](LIFECYCLE_EN.md)

#NeuronFS
### *A filesystem-native hierarchical rule memory & prompt compiler for AI agents.*

---

### Quickstart

**Option A — The One-Liner (CLI Native)**

```bash
# Mac / Linux
curl -sL https://neuronfs.com/install | bash

# Windows (PowerShell)
iwr https://neuronfs.com/install.ps1 -useb | iex

# Initialize your brain
neuronfs --init ./my_brain    # Interactive scaffolding (7 regions)
export GROQ_API_KEY="gsk_..."   # For $0 auto-consolidation (Ollama local support coming soon!)

# Compile & Run
neuronfs ./my_brain --emit all  # Compile to .cursorrules / CLAUDE.md / GEMINI.md
neuronfs ./my_brain --consolidate # Llama 3 merges fragmented neurons into Super Neurons
neuronfs ./my_brain --api     # Dashboard at localhost:9090
neuronfs ./my_brain --watch    # Auto-recompile on changes
```

CLI commands available: `--init`, `--emit`, `--watch`, `--consolidate`, `--fire`, `--grow`, `--decay`, `--api`, `--supervisor`, `--neuronize`, `--polarize`.

**Option B — Live injection only (No Go needed)**

```bash
# 1. Create a brain manually (just folders)
mkdir -p ~/NeuronFS/brain_v4/brainstem/禁fallback
echo. > ~/NeuronFS/brain_v4/brainstem/禁fallback/5.neuron

mkdir -p ~/NeuronFS/brain_v4/cortex/frontend/禁console_log
echo. > ~/NeuronFS/brain_v4/cortex/frontend/禁console_log/9.neuron

# 2. Set environment variables
export NEURONFS_BRAIN="$HOME/NeuronFS/brain_v4"
export NODE_OPTIONS="--require $HOME/NeuronFS/runtime/v4-hook.cjs"

# 3. Start your IDE — done
cursor .
```

No build step. The hook is a single `.cjs` file with zero dependencies.
Node.js is already included in every Electron-based IDE (VS Code, Cursor, Windsurf).

Daily driver since January 2026. MIT License.

---

## TL;DR

**`mkdir` replaces system prompts.** Folders are neurons. Paths are sentences. Counter files are synaptic weights.

```bash
# Create a rule = create a folder
mkdir -p brain/brainstem/禁fallback
touch brain/brainstem/禁fallback/1.neuron

# Compile = auto-generate system prompts
neuronfs ./brain --emit cursor  # → .cursorrules
neuronfs ./brain --emit claude  # → CLAUDE.md
neuronfs ./brain --emit all   # → All AI formats at once
```

| Before |NeuronFS |
|--------|----------|
| 1000-line prompts, manually edited | `mkdir` one folder |
| Vector DB $70/mo | **$0** (folders = DB) |
| Switch AI → migration | `cp -r brain/` — 1 second |
| Rule violation → wishful thinking | `bomb.neuron` → **physical halt** |
| Rules managed by humans | Correction → auto neuron growth |



## Contents

| | Section | Description |
|---|---|---|
| | [Core Structure](#core-structure) | Folder = Neuron, Path = Sentence, Counter = Weight |
| | [Brain Regions](#brain-regions) | 7 regions, priority cascade, hormone system |
| ⚖️ | [Governance](#governance) | 3-Tier injection, bomb circuit breaker, harness |
| | [Neuron Lifecycle](#neuron-lifecycle) | Birth → Reinforcement → Dormancy → Apoptosis |
| ️ | [Architecture](#architecture) | Autonomous loop, CLI, MCP, multi-agent |
| | [Benchmarks](#benchmarks) | Performance, competitor comparison |
| | [Limitations](#limitations) | Honest assessment |
| | [FAQ](#faq) | Expected questions and answers |
| | [Story](#story) | Why this exists |

---

## Core Structure

> **Unix said "Everything is a file." We say: Everything is folders.**

| Concept | Biology |NeuronFS | OS Primitive |
|---------|---------|----------|-------------|
| Neuron | Cell body | Directory | `mkdir` |
| Rule | Firing pattern | Full path | Path string |
| Weight | Synaptic strength | Counter filename | `N.neuron` |
| Reward | Dopamine | Reward file | `dopamineN.neuron` |
| Kill | Apoptosis | `bomb.neuron` | `touch` |
| Sleep | Synaptic pruning | `*.dormant` | `mv` |
| Connection | Axon | `.axon` file | symlink |

### Path = Sentence

Paths are natural language commands. Depth equals specificity:

```
brain/cortex/NAS_file_transfer/          → Category
brain/cortex/NAS_file_transfer/禁Copy-Item_UNC/   → Specific rule
brain/cortex/NAS_file_transfer/robocopy_large/    → Sub-context
```

### Hanja Micro-Opcodes

`禁` (1 char) = "NEVER_DO" (8 chars). 3-5x more meaning density in folder names:

| Hanja | Meaning | Example |
|-------|---------|---------|
| **禁** | Forbidden | `禁fallback` |
| **必** | Required | `必auto_reference_KI` |
| **推** | Recommended | `推robocopy_large_files` |
| **警** | Alert | `警DB_delete_confirm_required` |

### Self-Evolution

`.cursorrules` is a static file you edit manually.NeuronFS is different:

```
AI makes mistake → correction → corrections.jsonl → mkdir (auto neuron growth)
AI does well → praise → dopamine.neuron (reward signal)
Same mistake 3x → bomb.neuron (entire output halted)
30 days unused → *.dormant (auto sleep)

**v4.3 Auto-Evolution:**
1. **auto-consolidate**: Mitigates folder fragmentation. LLM (Groq or local) detects redundant error folders and merges them into a single Super Neuron, transferring activation weights.
2. **auto-neuronize**: Analyzes correction logs to generate inhibitory (Contra) rules.
3. **auto-polarize**: Detects positive-form `use_X` rules and proposes conversions to inhibitory (`禁X`) micro-opcode formats.
   ↓
Automatically reflected in next session's system prompt
```

---

## Brain Regions

7 brain regions layered via Brooks' Subsumption Architecture. **Lower P always suppresses higher P.**

```
brainstem(P0) > limbic(P1) > hippocampus(P2) > sensors(P3) > cortex(P4) > ego(P5) > prefrontal(P6)
```

| Region | Priority | Role | Examples |
|--------|----------|------|----------|
| **brainstem** | P0 | Absolute laws | `禁fallback`, `禁SSOT_duplication` |
| **limbic** | P1 | Emotion filter, hormones | `dopamine_reward`, `adrenaline_emergency` |
| **hippocampus** | P2 | Memory, session restore | `error_patterns`, `KI_auto_reference` |
| **sensors** | P3 | Environment constraints | `NAS/禁Copy-Item`, `design/sandstone` |
| **cortex** | P4 | Knowledge, skills (largest) | `frontend/react/hooks`, `backend/supabase` |
| **ego** | P5 | Tone, personality | `concise_execution`, `korean_native` |
| **prefrontal** | P6 | Goals, projects | `current_sprint`, `long_term_direction` |

### Hormone System

- **Dopamine** (`dopamineN.neuron`): Praise → positive weight boost
- **Adrenaline** (`adrenaline.neuron`): "urgent" detected → lower P suppresses higher P
- **Bomb** (`bomb.neuron`): 3x repeated mistake → entire region output disabled (circuit breaker)

### Axons — Cross-Region Wiring

16 `.axon` files connect 7 regions into a layered network:

```bash
brainstem/cascade_to_limbic.axon   → "limbic"   # bomb → block emotions
sensors/cascade_to_cortex.axon    → "cortex"   # env constraints filter knowledge
cortex/shortcut_to_hippocampus.axon  → "hippocampus" # learning results → memory
```

---

## Governance

### 3-Tier Injection

| Tier | Scope | Tokens | When |
|------|-------|--------|------|
| **Tier 1** | brainstem + TOP 5 rules | ~200 | Every session, auto |
| **Tier 2** | Full region summary (GEMINI.md) | ~800 | System prompt |
| **Tier 3** | Specific region `_rules.md` full | ~2000 | On-demand when task detected |

*※ v4.2: `GEMINI.md` is now **dual-synced** to both workspace-local (`.gemini/`) and global user profile (`~/.gemini/`) to ensure all agent environments (including Antigravity extension) see the latest brain state immediately.*

### Circuit Breaker

| bomb location | Result |
|---------------|--------|
| brainstem (P0) | **Entire brain halted**. GEMINI.md becomes empty |
| cortex (P4) | brainstem~sensors only. Coding region blocked |

bomb doesn't remove the rule — it **halts all output from that region**. It's an emergency stop button.
Release: `rm brain_v4/.../bomb.neuron` — delete one file.

### Harness

15-item automated verification script runs daily like CI:
- brainstem immutability check
- axon integrity scan
- dormant auto-cleanup
- violation detection → correction loop (never directly modifies)

---

## Neuron Lifecycle

> **Neurons are born, reinforced, audited, put to sleep, and killed.** Full lifecycle: [LIFECYCLE.md](LIFECYCLE.md)

```
Birth → Reinforcement → Maturation → Dormancy/Bomb → (Apoptosis or Revival)
```

| Stage | Trigger | Result |
|-------|---------|--------|
| **Birth** | Correction, Memory Observer, manual `mkdir` | Neuron folder + `1.neuron` |
| **Reinforcement** | Repeated correction | Counter ↑ → higher placement in prompt |
| **Reward** | Praise | `dopamine.neuron` → positive weight |
| **Dormancy** | 30 days untouched | `*.dormant` → excluded from compile |
| **Bomb** | Same mistake 3x | `bomb.neuron` → region output halted |
| **Apoptosis** | dormant 90d + counter 1 | Delete candidate (user approval required) |

### Audit Schedule

| Frequency | Item | Status |
|-----------|------|--------|
| Idle / API | Duplicate detection (Jaccard similarity) | `deduplicateNeurons()` |
| `--decay` / Idle | Dormant marking (30d untouched) | `runDecay()` |
| Every scan | Bomb detection + physical alarm | `triggerPhysicalHook()` |
| `--neuronize` | Error-log-driven Contra evolution | Groq / Llama 3 70B pipeline |
| `--polarize` | Positive → inhibitory rule conversion | Groq / rule-based hybrid |
| Per commit | PII path scan | Planned |
| **Quarterly** | Region integrity, naming, hierarchy audit | Manual |

---

## Architecture

### Autonomous Loop

```
User correction → corrections.jsonl → neuronfs (fsnotify) → mkdir (neuron growth)
                               ↓
                          _rules.md regenerated → GEMINI.md
                               ↓
                          Next session AI behavior changes
```

### CLI

```bash
neuronfs <brain> --emit <target>  # Compile prompts (gemini/cursor/claude/copilot/all)
neuronfs <brain> --api       # Dashboard (localhost:9090)
neuronfs <brain> --watch      # File watch + auto-recompile
neuronfs <brain> --supervisor   # Process manager
neuronfs <brain> --grow <path>   # Create neuron
neuronfs <brain> --fire <path>   # Increment counter
neuronfs <brain> --decay      # 30-day dormancy sweep
neuronfs <brain> --init <path>   # Initialize new brain
neuronfs <brain> --snapshot    # Git snapshot
neuronfs <brain> --neuronize    # LLM-powered error/correction → auto contra neuron generation
neuronfs <brain> --polarize    # Positive neurons → negative/inhibitory auto-conversion
```

### Live Context Injection (`v4-hook.cjs`)

Instead of compiling rules to static files, you can inject your live brain state into every API request. The AI sees your latest neurons mid-conversation — not just at conversation start.

**How it works:**

```
Your IDE (VS Code, Cursor, Windsurf, etc.)
 │
 ├─ outgoing API request
 │  └─ v4-hook.cjs intercepts → scans brain_v4/ → appends neuron rules to system prompt
 │
 └─ AI sees your live neuron state on every turn
```

**Setup (any Electron-based AI IDE):**

```bash
# 1. Set your brain path
export NEURONFS_BRAIN="/path/to/your/brain_v4"

# 2. Tell Node.js to load the hook before the IDE starts
export NODE_OPTIONS="--require /path/to/NeuronFS/runtime/v4-hook.cjs"

# 3. Start your IDE normally
cursor .   # or code . / windsurf . / etc.
```

Windows:
```cmd
set NEURONFS_BRAIN=C:\path\to\brain_v4
set NODE_OPTIONS=--require "C:\path\to\NeuronFS\runtime\v4-hook.cjs"
start cursor .
```

That's it. No MCP server, no config files, no dependencies. The hook reads your filesystem directly and injects a compact summary (only neurons with counter ≥ 5) into every LLM call.

### AI Tool Integration

| AI Tool | Method | Live Updates |
|---------|--------|--------------|
| Cursor / Windsurf / VS Code | `v4-hook.cjs` (NODE_OPTIONS) | Every turn |
| Gemini CLI | GEMINI.md (`--emit gemini`) | At session start |
| Claude Code | CLAUDE.md (`--emit claude`) | At session start |
| GitHub Copilot | copilot-instructions.md (`--emit copilot`) | At session start |

### Why Go

Single binary. Zero dependencies. `go build` → `neuronfs`. Copy to any machine, done.
Cross-compile (`GOOS=linux`), native fsnotify, goroutines managing 5+ child processes.

### Multi-Agent

All agents share the **same `brain/`**. Agents are divided by "disposition" not "role":

| Agent | Disposition | Approach |
|-------|------------|---------|
| ANCHOR (ISTJ) | Conservative, principled | "This neuron violates harness rules" |
| FORGE (ENTP) | Aggressive, experimental | "Split this neuron into 3 for efficiency" |
| MUSE (ENFP) | Creative, empathetic | "This neuron name isn't intuitive" |

Role-based ("you are QA") stops at boundaries. Disposition-based approaches any problem through its personality.

---

## Benchmarks

Measured 2026-03-29, local Windows 11 SSD:

| Metric | Value |
|--------|-------|
| Full scan (420 neurons) | ~1ms |
| Add rule | `touch` <1ms |
| 1,000 neuron stress | 271ms (3-run avg) |
| Disk usage | 4.3MB |
| Runtime cost | **$0** |
| brainstem compliance | **94.9%** (18 violations in 353 fires) |

### Competitor Comparison

| | .cursorrules | Mem0 | Letta | **NeuronFS** |
|---|---|---|---|---|
| 1000+ rules | Token overflow | (vector DB) | | (folder tree) |
| Infrastructure | ₩0 | Server $$$ | Server $$$ | **₩0** |
| Switch AI | Copy file | Migration | Migration | **As-is** |
| Self-growth | | | | ** (correction→neuron)** |
| Immutable guardrails | | | | ** (brainstem + bomb)** |
| Audit | git diff | Query | Log | **ls -R** |

---

## Limitations

| Item | Status | Mitigation |
|------|--------|-----------|
| AI enforcement | Can't guarantee 100% compliance | Harness detects violations → correction loop. Measured 94.9% |
| Semantic search | No vector embeddings — by design | Folder structure IS the search |
| External validation | Single-operator production only | Community feedback after release |
| Conditional logic | Can't express if/else in folder names | `_rules.md` handles branching |

---

## FAQ

**Q: "Isn't this just putting text in a system prompt?"**
Yes. `--emit` compiles folders into text. The point is *who manages it and how*. Editing 1000-line prompts vs `mkdir` — same output, different maintenance cost.

**Q: "Does bomb remove the rule?"**
No. bomb **halts all output from that entire region**. cortex bomb → coding region blocked → AI can't code until you `rm bomb.neuron`.

**Q: "Won't 1000+ neurons explode tokens?"**
Three defenses: ① 3-Tier activation — only relevant regions loaded deeply. ② Dormant — 30-day untouched auto-sleep. ③ Consolidation — duplicate neurons merged upward.

**Q: "Is 293 a lot?"**
For personal use, yes. For enterprise, no. The point isn't the number — it's manageability. Finding line 237 in a 1000-line prompt vs finding `brain/cortex/frontend/react/禁console_log/`.

---

## Story

> *"Don't beg with prompts. Design the pipeline."*

AI broke the "don't use console.log" rule 9 times. On the 10th, `mkdir brain/cortex/frontend/coding/禁console_log` was made. The folder name became the rule. Counter reached 17. AI doesn't break it anymore.

**WhyNeuronFS exists:** Not to feed more context to bigger models, but to make the structure so solid that **AI dependency converges to zero**. The brain is the product. AI is just the reader.

> *"Use top models to build structure. In the final act, normalize until AI usage approaches zero."*

**⭐ Agree? Star. [Disagree? Issue.](../../issues)**

---

MIT License · Copyright (c) 2026

[ Full Manifesto](MANIFESTO.md) · [LIFECYCLE.md](LIFECYCLE.md)
