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
### *A self-evolving AI brain made of folders. Zero infra. Zero dependencies.*

**Unix said "Everything is a file." We say: Everything is folders.**

Folders are neurons. Paths are sentences. Counters are synaptic weights. The filesystem is the brain.

| Before | NeuronFS |
|--------|----------|
| 1000-line prompt | **One folder** |
| "Don't forget this rule" | `mkdir 禁fallback` |
| Vector DB $70/mo | **$0** |
| AI can ignore it | Harness **detects violations** → correction loop |
| Write rules in files | **The folder name IS the instruction. 0KB.** |
| Switch models = reset | `cp -r brain/` — **1 second** |

> *"Don't beg with prompts. Design the pipeline."*

### How This Differs from .cursorrules — Self-Evolution

`.cursorrules` and `CLAUDE.md` are **static files you edit by hand.** NeuronFS is different:

```
AI makes mistake → correction → corrections.jsonl → mkdir (neuron auto-created)
AI does well     → praise     → dopamine.neuron (reward signal)
Same mistake 3×  → bomb.neuron (hard stop on that output)
30 days unused   → *.dormant (auto-sleep)
     ↓
Automatically reflected in next session's system prompt
```

**This is the feedback loop.** You don't open a file and edit it. The brain grows on its own. Existing tools don't have this. `.cursorrules` is static text. NeuronFS is a **living context management system.**

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

## The Core Idea

**Folders compose into priority-ordered command sentences — and they evolve.**

This is not a static config file. This is not a checklist. Look at what `neuronfs --emit` actually produces:

```
sensors: "반드시 nas: 禁NAS직접쓰기 로컬만, 禁corrections NAS기록, 동기화 로컬에서NAS 단방향..."
ego: "트랜지스터 게이트분해. 한국어로 사고하고 응답. 전문가 간결."
brainstem: "절대 禁영어사고 한국어로 생각하고 대답. 禁뉴런구조 임의변경. 토론말고 실행."
```

Each line is a **sentence**, not a list. Neurons are compressed by activation weight into priority-ordered commands. The AI reads one dense paragraph per brain region — not 326 individual rules.

**And it evolves.** When the PD corrects the AI tomorrow, a new neuron is created. The sentence changes. The weight shifts. The brain rewrites itself.

### Three Phases — Not Just an Idea

This is not a whitepaper. It went through three phases:

| Phase | What Happened | Proof |
|-------|---------------|-------|
| **① Imagination** | "What if folders were neurons?" A non-developer's thought experiment, born from frustration with prompts that AI kept ignoring. | [MANIFESTO.md](MANIFESTO.md) — the full philosophical arc |
| **② Implementation** | Built in Go. Single binary. 328 neurons across 7 brain regions. Axons wiring cross-domain knowledge flow. Running daily since Jan 2026. | [runtime/](runtime/) — 4,000+ lines of Go |
| **③ Verification** | 15-point automated harness. Scans for violations, proposes fixes, never modifies directly. PD approves every merge. | [scripts/harness.ps1](scripts/harness.ps1) — PASS: 15, FAIL: 0 |

> *Making this public wasn't easy. This idea is personal. But if it helps one person escape prompt hell, it's worth it.*

---

## The Problem

> **Key: Every text-based command is a "suggestion" to AI, not a law.**

We beg AI with text.

*"Please don't forget this rule."*  *"Absolutely never use fallback."*  *"I told you 36 times."*

Thousand-line system prompts. $70/month vector databases. RAG pipelines that need servers, embeddings, cosine similarity — just to say *"don't use console.log."*

The AI ignores it all when token pressure rises. Prompts are suggestions, not laws.

> *"Prompt engineering is the act of being enslaved to a black box — begging, pleading, groveling."*

---

## The Answer: Structure, Not Text

> **Key: `mkdir` replaces thousand-line prompts. The folder path IS the natural language command.**

> *"It wasn't the concept that was missing. It was using folders. That's the core."*

NeuronFS replaces text-based AI rules with **OS filesystem structure.**

```bash
mkdir -p brain/cortex/testing/no_console_log
touch brain/cortex/testing/no_console_log/1.neuron
```

That's a neuron. Path = rule. Filename = counter. Zero infrastructure.

**Path length limits force granular decomposition:**

OS has path length limits (Windows: 260 chars). As paths deepen like `brain/cortex/frontend/react/hooks/禁console_log/`, names must stay short. This **naturally forces hierarchical decomposition:**

```
✗ Flat:         brain/cortex/frontend_react_hooks_never_use_console_log_always_check/  (long, meaning blob)
✓ Hierarchical: brain/cortex/frontend/react/hooks/禁console_log/  (short names, clear meaning)
```

Longer path → shorter names → deeper hierarchy → **transistor-gate decomposition, enforced by the OS itself.**

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

> *"Forget files. Just folders."*

### Mimicking the Biological Brain — Not Metaphor, Implementation

> **Key: The 7 brain regions aren't just names borrowed from neuroscience. Hormones, emotions, memory, conscience — each is implemented as folders.**

| Region | Biological Origin | NeuronFS Implementation | Live Neuron Examples |
|--------|------------------|------------------------|---------------------|
| **brainstem** (P0) | Brainstem — survival reflexes | Immutable rules. Harness detects and rejects modification attempts | `禁영어사고` (13×), `禁SSOT중복` (2×) |
| **limbic** (P1) | Limbic system — emotions, hormones | **Emotion filter**. Detects PD's tone → behavioral bias | `dopamine` (6), `adrenaline` (5) |
| **hippocampus** (P2) | Hippocampus — memory formation | **Record/recall**. Session logs, error patterns, context restore | `KI auto-reference` (6), `error patterns` (6) |
| **sensors** (P3) | Sensory organs — environment | **Environment constraints**. OS, NAS, brand, tools | `禁NAS직접쓰기`, `OS Windows 11` |
| **cortex** (P4) | Cerebral cortex — knowledge | **212 neurons**. Frontend, backend, community knowledge | `RLS always on`, `no pm solo work` (6) |
| **ego** (P5) | Self — personality, expression | **Tone/style**. Korean-native thinking, concise execution | `expert concise`, `results first` |
| **prefrontal** (P6) | Prefrontal cortex — planning | **Goals, projects, sprints** | `GitHub public launch`, `video pipeline v17` |

**Hormone system — actually implemented:**

- **Dopamine (`dopamine.neuron`):** Created when PD praises. Positive weight boost. Biology: reward circuit.
- **Adrenaline (`adrenaline.neuron`):** PD says "urgent" → limbic detects → lower P suppresses higher P. Biology: fight-or-flight.
- **Bomb (`bomb.neuron`):** 3 repeated failures → created. Entire subtree deactivated. Biology: apoptosis (cell death).

**Subsumption Cascade (Brooks Architecture):**

Lower P **always** suppresses higher P. If brainstem has a bomb → limbic through prefrontal all ignored. Conscience (P0) always overrides goals (P6).

### Why Hanja and Korean — Token Efficiency

> **Key: 禁 (1 character) = "forbidden/never do" (17 characters). Korean and Hanja compress 3-5× more meaning per token into folder names.**

| Neuron Name | Tokens (GPT-4) | English Equivalent | Tokens |
|-------------|----------------|-------------------|--------|
| `禁fallback` | ~3 | `NEVER_USE_FALLBACK_SOLUTIONS` | ~6 |
| `禁SSOT중복` | ~4 | `NEVER_DUPLICATE_SINGLE_SOURCE_OF_TRUTH` | ~9 |
| `推robocopy_대용량` | ~5 | `RECOMMEND_ROBOCOPY_FOR_LARGE_FILES` | ~8 |
| `반드시_KI자동참조` | ~5 | `ALWAYS_AUTO_REFERENCE_KNOWLEDGE_ITEMS` | ~8 |

**Hanja prefixes (1 character = 1 rule type):**
- **禁** = Forbidden (금지) — hard ban
- **推** = Recommended (추천) — soft preference  
- **반드시** = Mandatory — absolute requirement

When the brain has 326 neurons and each name is a folder path, **every saved token matters.** Hanja isn't aesthetic — it's compression. The same brain in English would consume 2-3× more tokens in the system prompt.

### Design Philosophy Roots

> **Key: NeuronFS didn't come from theory. It implemented proven software engineering principles as folder structures.**

| Principle | Original Meaning | NeuronFS Application |
|-----------|-----------------|---------------------|
| **Strangler Fig** | Replace legacy gradually, don't rewrite | Don't delete prompts overnight. Create neurons one by one. Prompts shrink. Neurons eventually replace them completely |
| **SSOT** | One source of truth | The folder tree is the **only source**. GEMINI.md is compiled output. Edits happen in folders only |
| **Brooks Subsumption** | Lower layers suppress higher ones | brainstem (P0) always beats prefrontal (P6). Conscience > Goals |
| **Hebbian Learning** | "Neurons that fire together wire together" | Frequently corrected neurons gain higher counters, appear earlier in the compiled system prompt |
| **Apoptosis** | Cell death — remove what's not needed | `bomb.neuron` = kill problematic neurons. `dormant/` = sleep unused ones |

### 3-Tier Brain Activation — Only Wake What You Need

> **Key: Don't read 328 neurons every time. Detect the task, deeply activate only the relevant brain region.**

| Tier | What's Read | When | Tokens |
|------|------------|------|--------|
| **Tier 1: Always on** | brainstem + limbic | Every conversation start | ~200 |
| **Tier 2: Summary** | All regions compressed (GEMINI.md) | Injected into system prompt | ~800 |
| **Tier 3: Deep activation** | Specific region's full `_rules.md` | Task detected | ~2000 |

```
User: "Fix the CSS"
  → Task detected: Design/UI
  → Tier 3: read cortex/_rules.md (212 neurons)
  → Other regions: Tier 2 summary only

User: "Copy to NAS"
  → Task detected: NAS/files
  → Tier 3: read sensors/_rules.md (31 neurons — 禁NAS직접쓰기 etc.)
  → Other regions: Tier 2 summary only
```

**Biological parallel:** When you solve a math problem, your prefrontal cortex activates strongly while visual cortex runs at minimum. A brain at 100% everywhere = wasted energy. For AI, tokens are energy.

### Axons — Cross-Region Wiring

> **Key: At 3,000 neurons across 20 regions, axons become more important than individual neurons. They are the topology of organizational thought.**

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
| **brainstem = immutable** | P0 rules are protected. Harness detects and rejects modification attempts | Self-modification attacks |

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

### What Most People Don't Know: AI Doesn't Re-Read Rules

> **Key: The system prompt loads once at conversation start. If neurons change mid-conversation, AI doesn't know. MCP solves this.**

```
Problem: mkdir creates new neuron → but AI in current conversation doesn't know
         → Why? System prompt already loaded. No way to detect file changes.

Solution: NeuronFS Go runtime acts as an MCP (Model Context Protocol) server
          → AI can query brain state in real-time
          → fsnotify detects file changes → recompile → push to AI
```

Without this, changing folders mid-conversation would be useless. **MCP turns static folder rules into a live, queryable brain.**

### Integration Difficulty by AI Tool

> **Key: Our Antigravity setup is advanced. CLI and Claude Code users have it much easier.**

| AI Tool | Integration | Difficulty |
|---------|------------|-----------|
| **Gemini CLI / Claude Code** | GEMINI.md / CLAUDE.md auto-loaded | ⭐ Ready to use |
| **Cursor / Windsurf** | .cursorrules auto-loaded | ⭐ Ready to use |
| **Google Antigravity** | Custom CDP auto-accept script required | ⭐⭐⭐ Advanced |

**Antigravity Auto-Accept (custom-built):** Connects via CDP to `localhost:9000` → auto-clicks the Run button AI proposes. Core to fully autonomous operation, but this is our environment-specific setup. CLI users don't need any of this.

> *The CDP auto-accept tool will be published as a separate repository.*

---

## Multi-Agent — Personality, Not Roles

> **Key: When you split agents by "role," they freeze when work falls outside their role. When you split by "personality (MBTI)," they find work that matches their temperament.**

```
┌─────────────────────────────────────────┐
│          PD (Project Director)          │
│   Non-developer. Sets direction.        │
│   Corrector. "That's wrong" → neuron    │
└────────────────┬────────────────────────┘
                 │ corrections/directives
┌────────────────▼────────────────────────┐
│          PM (AI — Antigravity)          │
│   Converts PD's corrections to          │
│   structure. Reads brain_v4/ → execute  │
└───┬────────────┬────────────────┬───────┘
    │            │                │
┌───▼───┐  ┌────▼────┐  ┌───────▼──────┐
│ANCHOR │  │ FORGE   │  │   MUSE       │
│ ISTJ  │  │  ENTP   │  │   ENFP       │
│careful│  │  bold   │  │  creative    │
│verify │  │  build  │  │  docs/UX     │
└───────┘  └─────────┘  └──────────────┘
    │            │                │
    └────────────┴────────────────┘
           Same brain (brain_v4/)
```

**PD can't code.** But every correction starts from PD. When PD says "that's wrong," PM(AI) writes to corrections.jsonl, and the supervisor converts it to a folder. **Humans correct, structure learns.**

All agents **share the same brain (brain_v4/)**. One brain. But each agent has a **different personality** — like people.

| Agent | MBTI | Temperament | Work Style |
|-------|------|-------------|-----------|
| **ANCHOR** | ISTJ | Conservative, principled, meticulous | Harness validation, governance, violation detection |
| **FORGE** | ENTP | Aggressive, experimental, fast | Code generation, refactoring, new experiments |
| **MUSE** | ENFP | Creative, empathetic, big-picture | Documentation, UX, community, ideas |

**Why personality over roles:**

- Role-based: "You are QA" → problem outside QA scope → "Not my job" 🛑
- Personality-based: "You are ISTJ" → any problem → approach it methodically, rigorously ✅

```
Same brain, different personalities:

ANCHOR(ISTJ): reads brain_v4/ → "This neuron violates harness rule #7"
FORGE(ENTP):  reads brain_v4/ → "Split this neuron into 3 for better granularity"  
MUSE(ENFP):   reads brain_v4/ → "This neuron name isn't intuitive enough"
```

Same brain, same rules, but **personality drives what each agent notices.** Just like a real team.

---

## Competitors Comparison

> **Key: Other tools let AI decide what to remember. NeuronFS lets the user decide.**

|  | .cursorrules | Mem0 | Letta (MemGPT) | **NeuronFS** |
|--|-------------|------|----------------|-------------|
| 1000+ rules | Token overflow ❌ | ✅ (vector DB) | ✅ (tiered memory) | ✅ (folder tree) |
| Infra cost | $0 | Server/API $$$ | Server $$$ | **$0** |
| Switch AI | Copy file | Migration | Migration | **No change** (it's folders) |
| User owns rules | ✅ | ❌ (AI extracts) | ❌ (agent manages) | **✅ (mkdir)** |
| Self-growth | ❌ | ✅ | ✅ | **✅ (correction→neuron)** |
| Immutable guardrail | ❌ | ❌ | ❌ | **✅ (brainstem + harness detection)** |
| Auditability | git diff | Query | Logs | **ls -R** |

### What the Output Looks Like

**`brainstem/_rules.md`** — Auto-generated from folder structure. Harness detects and rejects modifications:

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

> **Key: If the brain is zero-infrastructure, the runtime should be too. One binary. Zero dependencies.**

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

## People Ask: "When AI Gets Better, You Can Add More Neurons, Right?"

> **We think in reverse.**

People say: "When 1M-token context comes, we can put a whole book in the system prompt." More rules, longer explanations, bigger brains.

**Is that really the right direction?**

"How much can you fit at once" is the wrong question. Small, certain gates composed into a pipeline approach 100% control.

The person who built NeuronFS can't write code. A 40-year-old Korean man who met the vibe coding era. His philosophy on AI:

| Phase | Purpose | AI Dependency |
|-------|---------|--------------|
| **① Architecture** | Use top models to design structure | High (Opus, Gemini) |
| **② Normalization** | Structure hardens → cheaper models achieve same result | Medium (Groq, LLaMA) |
| **③ Autonomous** | Structure complete → runs without AI | **Converges to 0** |

> *"High-end models build the structure. In the final chapter, the goal is to normalize away AI entirely — or use it at transistor-level minimum."*

**When someone asks "can you add more?", we ask the reverse: "As architecture matures, can't a weaker AI do the same job?"**

This is why NeuronFS exists. Don't make the model smarter. Make the pipeline stricter. As structure hardens, AI dependency converges to zero.

---

## Limitations & Honest Status

> **This is not a perfect system. But it works in production every day.**

| Item | Status | Response |
|------|--------|----------|
| AI enforcement | AI can't be 100% forced to follow GEMINI.md | Harness detects violations → correction loop. **Measured 94.9%** (18 brainstem violations / 353 total fires) |
| Semantic search | No vector embeddings — **by design** | Folder structure IS the search. Past 500 neurons: `tree` + dashboard |
| External validation | Currently validated in 1-person production only | Seeking community feedback post-launch |
| Windows-first | Currently running on Windows 11 | Go binary cross-compiles to macOS/Linux instantly |

> Admit limitations first and they become trust. Hide them and HN tears you apart in 3 minutes.

---

## FAQ — Honest Answers to Hard Questions

**Q: "Isn't this just putting text into a system prompt?"**

Yes. `neuronfs --emit` compiles the folder tree into text. AI reads text. But the point is **who manages that text and how.** Editing a 1000-line prompt directly vs. running `mkdir` to add a rule produce the same output — but the maintenance cost is different. Git manages text files. NeuronFS manages system prompts.

**Q: "The Palantir comparison is a stretch."**

Palantir's ontology and NeuronFS solve different problems at different scales. The comparison is about **principle**, not implementation — "Don't solve everything with one giant model. Break decisions into small gates and pipeline them." That principle is the same. NeuronFS isn't claiming to be Palantir — it's saying the same principle works at $0 for individuals.

**Q: "PASS: 15 is self-validation, not proof."**

Fair. It's a self-built harness testing 15 items: brainstem immutability, axon integrity, dormant cleanup, etc. No external benchmarks exist yet. But this harness **runs daily like CI** — zero regressions in 3 months of production is minimal proof. Stronger validation will come from the community.

**Q: "Is MBTI for agents scientific?"**

The debate about MBTI's scientific validity is irrelevant here. The point is **"split by temperament, not role."** "You are QA" → stops at out-of-scope work. "You are conservative and principled" → approaches any work that way. MBTI is just a framework to express temperament. Big Five or any other model works too.

**Q: "Are 326 neurons a lot?"**

For a personal project, yes. For enterprise, no. The point isn't the number — it's **manageability.** Finding line 237 in a 1000-line prompt vs. finding `brain/cortex/frontend/react/禁console_log/` — which is easier?

**Q: "Won't tokens explode when neurons reach 1000, 2000?"**

No. Three defenses: ① **3-Tier Activation** — only the relevant region gets a deep read; others get summaries. ② **Dormant auto-cleanup** — 30-day unfired neurons move to `*.dormant`, excluded from compilation. ③ **Periodic consolidation** — duplicate/similar neurons merge into parent neurons. Active token quota stays **bounded** regardless of neuron count.

**Q: "How do you express complex conditional logic (if A then B, unless C) with just folders?"**

You don't. Each region's `_rules.md` handles conditional branching as text. Folders = neurons (concepts), `_rules.md` = detailed rules for that region. The 3-Tier system loads only the relevant `_rules.md` on demand. Complex conditionals go in `_rules.md`; folder names stay as short keywords. That's the design principle.

**Q: "Isn't the CDP auto-accept brittle against browser updates?"**

Yes. CDP-based integration is brittle. **This is specific to the creator's setup.** If you use Gemini CLI, Claude Code, Cursor, or any CLI-based tool, this problem doesn't exist — GEMINI.md auto-loads. CDP is a choice for "fully autonomous operation," not a requirement of NeuronFS.

**Q: "If bomb.neuron fires, doesn't the rule disappear? Won't AI break it more?"**

bomb doesn't remove a rule. It's a **circuit breaker that stops the entire region's output.** Verified in code:

| bomb location | result |
|--------------|--------|
| brainstem (P0) | **Entire brain stops**. 0 neurons output. GEMINI.md goes empty |
| limbic (P1) | Only brainstem outputs. 6 other regions fully blocked |
| cortex (P4) | Only brainstem~sensors. Coding region itself blocked |

"bomb on console.log ban" → not "delete that rule" but "**stop all cortex output**" → AI can't code at all → PD removes bomb → normal recovery. Not abandonment — **emergency stop button.**

**Real-world activation:** bomb is not just design — it has fired in production. 2026-03-29, cortex and hippocampus regions, 1 activation each. PD removed bomb files, normal operation restored.

**How to remove:** `rm brain_v4/cortex/.../bomb.neuron` — delete the file. That's it. No CLI command. "Everything is folders" means file deletion = disarm. Intentionally simple — complex disarm procedures in emergencies hurt usability. After removal, next `--emit` auto-recovers.

**Q: "When bomb fires, does the AI just quietly stop?"**

No. It **physically stops.** When bomb.neuron is detected, `triggerPhysicalHook()` fires — a USB red siren literally starts spinning. The agent halts, PD sees it with their eyes, investigates, then `rm bomb.neuron` to restore.

This is not a metaphor — it's a **literal circuit breaker.** A hard stop to prevent token waste. Not a software notification. A physical alarm.

<p align="center">
  <img src="docs/bomb_alert.png" alt="bomb.neuron physical alert — fullscreen red flash + USB siren" width="700" />
  <br/>
  <sub>bomb.neuron detected → fullscreen red flash + USB siren + Telegram alert. A literal emergency stop.</sub>
</p>

```go
// physical_hooks.go — OS physical interrupt on bomb detection
func triggerPhysicalHook(regionName string) {
    // PowerShell beep + USB siren trigger
    cmd := exec.Command("powershell", "-NoProfile", "-Command",
        "[console]::beep(1000, 500); Write-Warning 'NEURONFS FATAL: BOMB detected'")
    _ = cmd.Run()
}
```

---

### Why Korean? A Token Advantage for Everyone

NeuronFS was built in Korean. This isn't a limitation — it's an **advantage.**

| | English | Korean (Hanja) |
|--|---------|---------------|
| "Never use English for thinking" | 6 tokens | `禁영어사고` = **1 token** |
| "Always verify before delivery" | 5 tokens | `推검증후납품` = **1 token** |
| Folder name length | Long paths | 2-4 chars = same meaning |

Hanja (漢字) characters compress 6-word rules into 2-character folder names. This means:
- **6× fewer tokens** consumed per rule
- **More rules fit** in the same context window
- **OS path limits** (260 chars) hit much later

**English speakers:** You can use NeuronFS in English (`NEVER_use_fallback/`). But consider mixing in Hanja prefixes (`禁fallback/`) — your AI reads them correctly and you save tokens.

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

### The Philosophy Behind It

I'm a 40-year-old Korean male who can't write a single line of code. I met the vibe coding era and rode it.

Everyone asks: *"If the AI gets better, can you add more neurons?"*

I think the opposite. **When the structure matures, even a cheaper AI can do the same job.**

The trajectory:
1. **High-end model builds the structure.** Use GPT-4/Gemini/Claude to architect the brain — define regions, axons, governance rules.
2. **Structure normalizes.** Once 326 neurons are in place, the brain becomes a deterministic pipeline. The AI doesn't need to "understand" your rules — it reads folders.
3. **End state: transistor-level minimization.** The goal is not more AI, but **less AI**. When every decision is pre-decomposed into yes/no gates (folders), you need barely any model intelligence to execute correctly. You could run on GPT-3.5 and get the same results.

> *"Don't wait for the next model. Build the structure so well that you don't need it."*

This is why NeuronFS exists. Not to feed more context to bigger models — but to **normalize AI usage down to zero.** The brain is the product. The AI is just the reader.

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

**A startup ships an AI product.** `brainstem/NEVER_expose_user_PII/` — the harness detects and rejects any modification to this rule. No intern, no junior dev's AI, no "creative" prompt can bypass it. The guardrail is structure, not a comment in code review.

**A law firm shares neurons across offices.** NAS shared folder. Tokyo and Seoul read the same `brain/legal/contracts/` neurons. `robocopy` syncs once per hour. Zero infrastructure.

**10 AI agents manage a factory.** Each agent forks `brain/base/` and evolves specialized neurons for their domain. Quality control agent has 500 neurons. Logistics agent has 300. They all share `brainstem/` — the constitutional rules.

326 neurons. Zero infrastructure. Zero dependencies. The filesystem itself is the framework.

> *"Don't beg with prompts. Design the pipeline."*
>
> *"Individuals beating corporations. Release it free."*

**⭐ Star if you agree. [Issue if you don't.](../../issues)**

---

<p align="center">
  <sub>MIT License · Copyright (c) 2026 박정근 (PD) rubisesJO777 — VEGAVERY RUN®</sub><br/>
  <sub><a href="MANIFESTO.md">📜 Full Manifesto</a> · <a href="https://instagram.com/rubises">@rubises</a></sub>
</p>

<p align="center"><sub><i>Hey 666, easy — only the Word stands as absolute truth (777). This? Just a well-organized folder.</i></sub></p>
