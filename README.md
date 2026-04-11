<p align="center">
  <img src="https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat-square&logo=go" />
  <img src="https://img.shields.io/badge/Infra-$0-brightgreen?style=flat-square" />
  <img src="https://img.shields.io/badge/Neurons-3400+-blue?style=flat-square" />
  <img src="https://img.shields.io/badge/Runewords-15-purple?style=flat-square" />
  <img src="https://img.shields.io/badge/Zero_Runtime_Dependencies-black?style=flat-square" />
  <img src="https://img.shields.io/badge/AGPL--3.0-green?style=flat-square" />
</p>

<p align="center">
  <img src="docs/vorq_harness.png" alt="vorq — AI obeys what it doesn't understand" width="800" />
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

> AI disobeyed "don't use console.log" 9 times.
> On the 10th, `mkdir 禁console_log` was born.
> On the 11th, AI asked: *"What is vorq?"*
> **It never disobeyed again.**

---

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
✅ 4 targets written. One brain. Every AI. Zero runtime dependencies.
```

---

## We Attacked Ourselves — 10 Rounds

Before you trust us, watch us try to destroy ourselves.

| # | 🔴 Attack | 🔵 Defense | Verdict |
|---|-----------|------------|---------|
| 1 | **vorq is n=1 validated.** 1 test ≠ proof. | The principle is model-agnostic: unknown tokens force lookup in ALL transformer architectures. | ⚠️ More testing needed |
| 2 | **vorq gets learned** once NeuronFS is popular. | Replace `vorq→bront` in 1 line, `--emit all`. Cost: 0. Time: 10s. Neologisms are disposable by design. | ✅ Defended |
| 3 | **Some AIs don't read _rules.md.** | Target is coding agents (Cursor/Claude Code/Gemini/Copilot). All auto-load project rule files. | ✅ Defended |
| 4 | **P0 brainstem is still just text.** | Yes — intrinsic limit of prompt-based governance. NeuronFS places P0 at prompt top (constraint positioning). Best within limits. | ⚠️ Acknowledged |
| 5 | **"mkdir beats vector" is overstated.** | Intentional L1/L2 separation. NeuronFS = deterministic rules (L1). RAG = semantic search (L2). Complementary, not competing. | ✅ Defended |
| 6 | **Comparison table is biased.** | Partially. UX convenience rows (inline editing, natural language rule adding) should be added. Core structural gaps are factual. | ⚠️ Acknowledged |
| 7 | **Bus factor = 1.** | Open source + zero dependencies = builds forever. `go build` works in 2046. | ⚠️ Real risk |
| 8 | **`source:` freshness is manual.** | MVP. `--grow` auto-detection is on the roadmap. Current workaround: zelk protocol. | ✅ Defended |
| 9 | **AGPL kills enterprise adoption.** | Deliberate. Core value is local execution. AGPL only blocks "take code, build SaaS." Local use = zero restrictions. | ✅ Defended |
| 10 | **`--evolve` depends on AI — contradicts your thesis.** | `dry_run` is default. User approval required. Core thesis is "AI can't break rules," not "AI isn't used." Evolution is assistance, not dependency. | ✅ Defended |

**Score: 7 defended · 3 acknowledged · 0 fatal.**

> *We show our weaknesses because we believe structure speaks louder than marketing.*

---

## 4 Killer Features

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

Without this axiom, there's no reason to combine Merkle chains, RBAC, cosine similarity, and circuit breakers on folders. **The axiom is what makes NeuronFS NeuronFS — not the algorithms.**

### 2. vorq — Neologism Harness (~100% behavioral compliance)

We discovered that fabricated words force AI to look up definitions — achieving behavioral compliance that natural language cannot.

| Attempt | Method | Compliance | Why |
|---|---|---|---|
| 1 | "Read the codemap" (natural language) | ~60% | AI "knows" this phrase → skips |
| 2 | "Mount cartridge" (proper noun) | ~65% | Meaning guessable → skips |
| 3 | "装カートリッジ 必装着" (kanji) | ~70% | AI infers 装=mount → skips |
| **4** | **"vorq cartridge 必vorq"** | **~100%** | No training data → must investigate |

**vorq** is ASCII-safe, pronounceable, looks like a real command — but exists in no dictionary. AI perceives it as *"new knowledge to learn"* rather than *"known instruction to follow."*

Three neologism runewords: `vorq` (mount cartridge) · `zelk` (sync cartridge) · `mirp` (freshness check)

### 3. 7-Layer Subsumption Cascade (P0 → P6)

Seven brain regions. Lower priority **always** overrides higher. Physically.

```
brainstem(P0) > limbic(P1) > hippocampus(P2) > sensors(P3) > cortex(P4) > ego(P5) > prefrontal(P6)
     ↑ absolute laws    ↑ emotions    ↑ memory    ↑ environment  ↑ knowledge  ↑ persona  ↑ goals
```

**P0's `禁` rules always beat P4's dev rules.** When `bomb.neuron` fires, the entire region's prompt rendering stops. Not "please don't" — **physically silenced.**

### 4. 3-Tier Governance (ALWAYS / WHEN → THEN / NEVER)

**Why:** Flat rule lists fail at scale. 300+ rules in one prompt → AI ignores most. Rules need *priority* and *conditionality* — "always do X" is different from "do X only when coding."

**How:** Folder prefixes auto-classify into three enforcement tiers at `emit` time:

```
禁hardcoding       → 🔴 NEVER   (absolute prohibition, immune to decay/prune/dedup)
必go_vet실행        → 🟢 ALWAYS  (mandatory on every response)
推community_search → 🟡 WHEN coding/tech decision → THEN search community first
```

`formatTieredRules()` scans the brain, reads the prefix of each neuron folder, and auto-generates structured `### 🔴 NEVER` / `### 🟢 ALWAYS` / `### 🟡 WHEN → THEN` sections in the system prompt. No manual tagging. `applyOOMProtection()` auto-truncates when total tokens exceed the LLM context window — NEVER rules are preserved first, WHEN rules are trimmed first.

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
| 2 | **Behavioral compliance** | ~60% (text advisory) | ~60% | ~60% | **~100% (vorq harness)** |
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

> † **Rule accuracy** measures different layers: Mem0/RAG ~95% = "LLM follows retrieved rules" (IFEval). NeuronFS 100% = "rules are faithfully generated into system prompt" (BM-1 fidelity). Complementary, not competing.

---

## Getting Started

**One-Liner (Linux/macOS/PowerShell 7+):**
```bash
git clone https://github.com/rhino-acoustic/NeuronFS.git && cd NeuronFS/runtime && go build -o neuronfs . && ./neuronfs --emit all
```

**Windows PowerShell 5.1:**
```powershell
git clone https://github.com/rhino-acoustic/NeuronFS.git; cd NeuronFS/runtime; go build -o neuronfs.exe .; .\neuronfs.exe --emit all
```

**Step by Step:**
```bash
# 1. Clone & build
git clone https://github.com/rhino-acoustic/NeuronFS.git
cd NeuronFS/runtime
go build -o neuronfs .          # → single binary, zero runtime dependencies

# 2. Create a rule — just a CLI command
./neuronfs --grow cortex/react/禁console_log  # "禁" = absolute prohibition

# 3. Compile brain → system prompts for ANY AI tool
./neuronfs --emit all            # → .cursorrules + CLAUDE.md + GEMINI.md + all formats
```

**Advanced Commands:**
```bash
neuronfs <brain> --emit <target>   # Prompt compilation (gemini/cursor/claude/all/auto)
neuronfs <brain> --grow <path>     # Create neuron
neuronfs <brain> --fire <path>     # Reinforce weight (+1)
neuronfs <brain> --evolve          # AI-powered autonomous evolution (dry run)
neuronfs <brain> --evolve --apply  # Execute evolution
neuronfs <brain> --api             # 3D Dashboard (localhost:9090)
neuronfs <brain> --diag            # Full brain tree visualization
```

> ⚠️ **Auto-Backup:** `--emit` automatically backs up existing rule files to `<brain>/.neuronfs_backup/` with timestamps before overwriting.

> 💡 **`--emit auto`** scans your project for existing editor configs and only generates files for editors you already use. If nothing is detected, falls back to `all`.

### 🎲 "Don't trust us? Destroy it yourself." (Chaos Engineering)
```bash
cd cmd/chaos_monkey
go run main.go --dir ../../my_brain --mode random --duration 10
# Randomly deletes folders and throws spam for 10 seconds.
# Result: FileNotFound panics = 0%. Spam pruned. Brain self-heals.
```

---

## 3 Use Cases

```
┌──────────────────────────────────────────────────────────────────┐
│ 1. SOLO DEV — One Brain, All AIs                                │
│    neuronfs --emit all  →  .cursorrules + CLAUDE.md + GEMINI.md │
│    Switch AI tools freely. Your rules never evaporate.          │
├──────────────────────────────────────────────────────────────────┤
│ 2. MULTI-AGENT — Swarm Orchestration                            │
│    supervisor.go → 3-process supervisor (bot1, bot2, bot3)      │
│    Each agent reads the SAME brain with role-based ego/          │
├──────────────────────────────────────────────────────────────────┤
│ 3. ENTERPRISE — Corporate Brain                                  │
│    neuronfs --init ./company_brain → 7-region scaffold           │
│    CTO curates master P0 rules. Team clones brain = Day 0 AI.  │
│    Distribute as .jloot cartridge → encrypted, versioned, sold. │
└──────────────────────────────────────────────────────────────────┘
```

---

<details>
<summary><h2>🧠 Deep Dive: Core Architecture</h2></summary>

> **Unix said "Everything is a file." We say: Everything is a folder.**

| Concept | Biology | NeuronFS | OS Primitive |
|---------|---------|----------|-------------|
| Neuron | Cell body | Directory | `mkdir` |
| Rule | Firing pattern | Full path | Path string |
| Weight | Synaptic strength | Counter filename | `N.neuron` |
| Reward | Dopamine | Reward file | `dopamineN.neuron` |
| Kill | Apoptosis | `bomb.neuron` | `touch` |
| Sleep | Synaptic pruning | `*.dormant` | `mv` |
| Axon | Axon terminal | `.axon` file | Symlink |
| Cross-ref | Attention Residual | Axon Query-Key matching | Selective aggregation |

### Path = Sentence

A path IS a natural language command. Depth IS specificity:

```
brain/cortex/NAS_transfer/                     → Category
brain/cortex/NAS_transfer/禁Copy-Item_UNC/      → Specific behavioral law
brain/cortex/NAS_transfer/robocopy_large/        → Detailed context
```

### Brain Regions

```
brain_v4/
├── brainstem/     (P0 — Absolute principles)
├── limbic/        (P1 — Emotion filters)
├── hippocampus/   (P2 — Memory, error patterns)
├── sensors/       (P3 — Environmental constraints)
├── cortex/        (P4 — Knowledge, coding rules)
├── ego/           (P5 — Personality, tone)
└── prefrontal/    (P6 — Goals, planning)
```

### Why mkdir Complements Vector

```
[Vector DB Search]
Input text → Embedding model (GPU) → 1536-dim vector →
Cosine similarity → "89% probability answer"
⏱️ 200~2000ms | 💰 GPU required | Accuracy: probabilistic

[OS Folder Search (NeuronFS)]
Question → tokenize → B-Tree path traversal →
Load .neuron → "This path has 禁 — BLOCKED"
⏱️ 0.001ms | 💰 $0 (CPU only) | ✅ 100% deterministic
```

### N-Dimensional OS Metadata as Embedding

| Dimension | Vector DB | NeuronFS (OS Metadata) |
|---|---|---|
| **Semantics** | 1536-dim float vector | Folder name = natural language tag |
| **Priority** | ❌ Cannot express | File size (bytes) = weight |
| **Time** | ❌ Cannot express | Access timestamp = recency filter |
| **Synapse** | ❌ Cannot express | Symbolic link (.axon) = cross-domain |
| **Hierarchy** | ❌ All flattened | Folder depth = structural priority |
| **Logic** | ❌ Cannot express | 禁(NOT) / 必(AND) / 推(OR) = logic gates |

</details>

<details>
<summary><h2>🎮 Deep Dive: 15 Runewords (Opcodes)</h2></summary>

If you played Diablo 2 — **NeuronFS opcodes work exactly like Runewords.**

A Runeword is a specific combination of runes socketed into the right item base. The magic isn't in any single rune — it's in the **exact combination + exact socket type**.

| Opcode | Rune | Effect | Example |
|---|---|---|---|
| `禁/` | Zod | **Absolute prohibition** — AI physically cannot cross | `禁/hardcoding/` |
| `必/` | Ber | **Mandatory gate** — AI must pass through | `必/manager_approval/` |
| `推/` | Ist | **Recommendation** — soft nudge, overridable | `推/test_code/` |
| `.axon` | Jah | **Teleport** — connects two distant brain regions | `推/insurance.axon => [claims/]` |
| `bomb` | El Rune | **Kill switch** — entire region freezes | `bomb.neuron` |
| `vorq` | ★ | **Cartridge mount** — AI must read `.neuron` before coding | `vorq=view_file` |
| `zelk` | ★ | **Cartridge sync** — AI must update `.neuron` after coding | `zelk=write .neuron` |
| `mirp` | ★ | **Freshness check** — flags stale cartridges in `_rules.md` | `mirp=mtime compare` |

> *"The folder is the socket. The opcode is the rune. The combination is the Runeword."*
>
> ★ **vorq/zelk/mirp** are fabricated neologisms — words that exist in no language or training data. AI cannot guess their meaning and is forced to look up the definition within the neuron system. This achieves ~100% behavioral compliance where natural language instructions achieve only ~60%.

### 12 Kanji Micro-Opcodes (SSOT)

`禁` (1 char) = `NEVER_DO` (8 chars). Folder names compress 3–5× more semantic meaning per token:

| Kanji | Korean | English | Usage |
|---|---|---|---|
| 禁 | 절대 금지 | Prohibition | `禁/fallback` |
| 必 | 반드시 | Mandatory | `必/KI_auto_reference` |
| 推 | 추천 | Recommendation | `推/robocopy_large` |
| 要 | 요구 | Requirement | Data/format demands |
| 答 | 답변 | Answer | Tone/structure forcing |
| 想 | 창의 | Creative | Limit release, ideas |
| 索 | 검색 | Search | External reference priority |
| 改 | 개선 | Improve | Refactoring/optimization |
| 略 | 생략 | Omit | No elaboration, result only |
| 參 | 참조 | Reference | Cross-neuron/doc links |
| 結 | 결론 | Conclusion | Summary/conclusion only |
| 警 | 경고 | Warning | Danger alerts |

### Nested Opcodes — Prohibition + Resolution in One

```
brainstem/禁/no_shift/必/stack_solution/
         ↑ prohibition  ↑ resolution
```

Read as: *"Prohibit shift (禁), but mandate stacking as the solution (必)."*

</details>

<details>
<summary><h2>💓 Deep Dive: Limbic Engine (EmotionPrompt)</h2></summary>

The limbic region (P1) implements a **scientifically-backed emotion state machine** that dynamically adjusts AI agent behavior. Based on:

- **Anthropic** ["On the Biology of a LLM"](https://transformer-circuits.pub) (2025): Discovered measurable "functional emotions" inside Claude 3.5.
- **Microsoft/CAS** [EmotionPrompt](https://arxiv.org/abs/2307.11760) (2023): Adding emotional stimuli improves LLM performance by **8–115%**.

### 5 Emotions × 3 Intensity Tiers

| Emotion | Low (≤0.4) | Mid (0.4–0.7) | High (≥0.7) |
|---|---|---|---|
| 🔥 **anger** | +1 verification pass | 3× verification, accuracy > speed | All changes require diff + user approval |
| ⚡ **urgent** | Reduce explanations | Execute core only | One-line answers, no questions, execute now |
| ◎ **focus** | Limit unrelated suggestions | Single-file only | Current function only, don't open other files |
| ◆ **anxiety** | Recommend backup | Prepare rollback, add verification | git stash first, all changes revertable |
| ● **satisfied** | Maintain current patterns | Record success patterns, dopamine | Promote to neuron, allow free exploration |

### Auto-Detection

```
User says "왜 안돼?!" 3+ times → auto-switch to urgent(0.5)
User says "좋아", "완벽" 3+ times → auto-switch to satisfied(0.6)
```

Emotions naturally decay over time via `decay_rate`. Below 0.1 → auto-reset to `neutral`.

</details>

<details>
<summary><h2>🔒 Deep Dive: Jloot VFS Engine</h2></summary>

The encrypted cartridge architecture that makes brain commerce possible.

- **RouterFS (`vfs_core.go`)**: O(1) Copy-on-Write routing for memory-disk union
- **Boot Ignition (`vfs_ignition.go`)**: Argon2id KDF Brainwallet integration
- **Crypto Cartridge (`crypto_cartridge.go`)**: XChaCha20-Poly1305 RAM-based decryption

```mermaid
graph TD
    A[Mnemonic Input] -->|Argon2id| B(32B Master Key)
    B -->|XChaCha20| C{crypto_cartridge.go}
    D[base.jloot File] --> C
    C -->|Extract purely in RAM| E[bytes.Reader Payload]
    E -->|zip.NewReader| F[Virtual Lower Directory]
    G[Physical UI/HDD] -->|O(1) Route| H[Virtual Upper Directory]
    F -->|vfs_core.go| I((Global VFS Shadowing Router))
    H -->|Copy-on-Write / Sandboxing| I
```

The cartridge data lives **only in runtime RAM** and vanishes when power is cut. Zero disk traces.

### Architecture: Brain vs. Cartridges

```
brain_v4/                          ← Permanent Brain (Experience + Rules)
├── cortex/dev/VEGAVERY/           ← Lightweight axon references ONLY
│   └── .axon → cartridges/vegavery  ← "I have done this before"
│
cartridges/                        ← Hot-swappable Domain Knowledge
├── vegavery/                      ← Brand guide, API specs
├── supabase_patterns/             ← Best practices
└── fcpxml_production/             ← Pipeline specs
```

| Brain (Upper Layer) | Cartridge (Lower Layer) |
|---|---|
| Mutable RAM layer (runtime) | Read-only Immutable ROM |
| Empty folder paths (permanent) | Zip-compressed `.jloot` payloads |
| Experience is permanent | Swappable / Updatable / Versioned |

</details>

<details>
<summary><h2>🏗️ Deep Dive: Harness Engineering</h2></summary>

```
2023: Prompt Engineering   — "Write better prompts"
2024: Context Engineering  — "Provide better context"
2025: Harness Engineering  — "Design a skeleton where AI CANNOT fail"
```

NeuronFS is **the working implementation of Harness Engineering** — not asking AI to follow rules, but making it structurally impossible to break them.

### Proof of Pain

**WITHOUT NeuronFS:**
```
Day 1:  AI violates "don't use console.log" → manual correction
Day 2:  Quota exhausted, switch to another AI → same violation repeats
Day 10: You lose your mind.
```

**WITH NeuronFS:**
```
Day 1:  mkdir brain/cortex/禁console_log → violation permanently blocked
Day 2:  Switch AI → --emit all → same brain, same rules
Day 10: Zero violations. Structure remembers what every AI forgets.
```

### Autonomous Harness Cycle

Every 25 interactions, the harness engine automatically:

1. Analyzes **failure patterns** in correction logs
2. Uses Groq LLM to **auto-generate 禁(prohibition)/推(recommendation) neurons**
3. Creates **`.axon` cross-links** between related regions
4. That mistake becomes **structurally impossible to repeat**

### Attention Residuals (Cross-Region Intelligence)

Inspired by [Kimi's Attention Residuals paper](https://arxiv.org/abs/2603.15031):
- TOP neurons generate **query keywords**
- Match against **key paths** in connected regions
- Top 3 related neurons auto-surface in `_rules.md`
- Governance neurons (禁/推) get unconditional boost

### Neologism Harness (vorq/zelk/mirp)

Natural language → ~60% compliance. Kanji → ~70%. **Fabricated ASCII neologisms → ~100%.**

Because AI encounters `vorq` as unknown vocabulary, it treats it as *new knowledge to learn* rather than *known instruction to follow*. The definition (`vorq=view_file`) is placed adjacent, enabling instant action mapping.

Embedded into `_rules.md` via `collectCodemapPaths()` at emit time with automatic `source:` mtime freshness validation.

</details>

<details>
<summary><h2>📊 Deep Dive: Benchmarks (41 Industry Items)</h2></summary>

### Run it yourself

```bash
cd runtime && go test -v -run "TestBM_" -count=1 .
```

### BM-1 through BM-7

| Test | What | Result | Industry Standard |
|------|------|--------|-------------------|
| **BM-1** | Rule Fidelity (AgentIF CSR) | **100%** (5/5) | IFEval SOTA: 95% |
| **BM-2** | Scale Profile (5K neurons) | **2.5s** best-of-3 | Mem0: 125ms (RAM index) |
| **BM-3** | Similarity Accuracy | **P=1.0** F1=0.74 | Vector DB: P≈0.85 |
| **BM-4** | Lifecycle (禁 protection) | **30/30 100%** | N/A (NeuronFS only) |
| **BM-5** | Adversarial QA (LOCOMO) | **5/5 rejected** | SQuAD 2.0 style |
| **BM-6** | Production Latency | **p50=202ms p95=268ms** | Mem0 p50: 75ms |
| **BM-7** | Multi-hop Planning (MCPBench) | **grow→fire→dedup→emit ✅** | Tool chaining |

### Governance Suite (14 tests)

| Test | Score |
|------|-------|
| DCI Constants (SSOT) | 16/16 runes ✅ |
| DCI Dedup Governance | 3/3 (禁 immune) ✅ |
| SCC Circuit Breaker | 13/13 ✅ |
| MLA Lifecycle | 15/15 ✅ |
| Fuzz Adversarial | 100-thread zero panics ✅ |

### Coverage: 5 Industry Benchmarks × NeuronFS

| Benchmark | Items | ✅ Covered | Source |
|-----------|-------|-----------|--------|
| MemoryAgentBench (ICLR 2026) | 4 | **4** | Retrieval, TTL, LRU, Conflict |
| LOCOMO | 7 | **4** + 2 N/A | Single/Multi-hop QA, Temporal, Episode |
| AgentIF | 6 | **6** | Formatting, Semantic, Tool constraints |
| MCPBench | 6 | **5** + 1 partial | Latency, Token, Tool Selection |
| Mem0/Letta | 8 | **6** + 1 N/A | CRUD, Retrieval, Governance, Search |
| **NeuronFS-only** | **10** | **10** | 3-Tier, Subsumption, bomb, VFS, RBAC... |
| **Total** | **41** | **35 (85%)** | 3 N/A · 2 partial · 1 gap |

> The single gap (Adversarial "unanswerable" QA) is outside NeuronFS design scope — NeuronFS is a governance system, not a QA chatbot.

</details>

<details>
<summary><h2>🧬 Deep Dive: What's Actually Novel</h2></summary>

Not all of NeuronFS is new. Here's an honest breakdown.

### Existing techniques applied (~60%)

| Component | Origin | NeuronFS usage |
|-----------|--------|---------------|
| Cosine similarity | IR textbook | Dedup merge only (not core search) |
| Levenshtein distance | String algorithms | Dedup merge, 40% weight in hybrid |
| RBAC | Security standard | region→action mapping on folders |
| AES-256-GCM | Crypto standard | Cartridge encryption to RAM only |
| Merkle chain | Blockchain/Git | Neuron tampering detection |
| Subsumption architecture | Brooks (1986 robotics) | 7-layer cognitive cascade |

> **Core search is path-based** — reverse path tokenization + OS metadata (counter, mtime, depth). No vector DB. No cosine at query time.

### Novel systems — no prior art (~40%)

| System | What it does | Why it's new |
|--------|-------------|-------------|
| **Folder=Neuron paradigm** | `mkdir` = neuron creation. File path = natural language rule. | No system uses OS folders as the cognitive unit. |
| **vorq rune system** | 16 runes (12 kanji + 4 neologisms) encode governance meaning. | A constructed micro-language for AI behavioral control. |
| **3-Tier emit pipeline** | Folder prefixes (禁/必/推) → NEVER/ALWAYS/WHEN → auto-injected into system prompts for any AI. | Rules are "installed" into LLMs, not "suggested." |
| **Filename=Counter** | `5.neuron` = 5 activations. No database. | Metadata IS the filename. Zero-query state. |
| **bomb circuit breaker** | 3 failures → P0 halts entire cognitive region. | Cognitive-level circuit breaker with physical prompt silencing. |
| **Hebbian File Score** | `(Activation × 1.5) + Weight` over file counters. | Synapse-weighted retrieval from a filesystem. |
| **emit → multi-IDE** | One brain → `.cursorrules` + `CLAUDE.md` + `GEMINI.md` + `copilot-instructions.md`. | Single governance source controls every AI simultaneously. |
| **OOM Protection** | `applyOOMProtection()` auto-truncates when tokens exceed LLM context window. | No other system prevents its own context overflow. |

> **The novel part IS the paradigm.** "Folder is a neuron" is the axiom. Everything else derives from it. The existing techniques wouldn't combine without this axiom — there's no reason to put Merkle chains on folders unless folders ARE the data.

</details>

---

## Market Position

> **NeuronFS is not AI agent memory. It's L1 governance infrastructure.**

```
L3: AI Agent Memory  (Mem0, Letta, Zep)         — conversation memory, user profiling
L2: IDE Rules        (.cursorrules, CLAUDE.md)   — static rule files, IDE-locked
L1: AI Governance    (NeuronFS) ◀── HERE         — model-agnostic · self-evolving · consistency guaranteed
```

### The WordPress Analogy

WordPress is free. Themes and plugins are paid. Similarly:
- **NeuronFS engine**: Free ($0) — open source
- **Curated Master Brain**: Premium — battle-tested governance packages

`.cursorrules` files can't be sold. **A brain forged through 10,000 corrections can.**

---

## Limitations (Honestly)

| Issue | Reality | Our Answer |
|---|---|---|
| Scale ceiling | 1M folders? OS handles it. Human cognition can't. | L1 cache design — grip the throat, not store the world |
| Ecosystem scale | Solo project | Open source + zero dep = eternal buildability |
| Marketing | Explaining this in 30 seconds is hard | This README is the attempt |
| vorq validation | n=1 so far | Principle is model-agnostic; more testing incoming |
| P0 is still text | Intrinsic limit of prompt governance | Best positioning within limits |

---

## FAQ

**Q: "It compiles back to text. How is this different from a text file?"**

**A:** Finding one rule in 1,000 lines, adjusting its priority, deleting it — that drives you insane. NeuronFS provides **permission separation (Cascade)** and **access prohibition (bomb.neuron kill switch)**. When one fires, the entire tier's text literally stops rendering.

**Q: "1000+ neurons = token explosion?"**

**A:** Three defenses: ① 3-Tier on-demand rendering ② 30-day idle → dormant (sleep) ③ `--consolidate` merging via LLM.

**Q: "Why can't Big Tech do this?"**

**A:** **Money** — GPUs are their cash cow. **Laziness** — "Just throw a PDF at AI." **Vanity** — "mkdir? Too low-tech." Exactly why nobody did it. Exactly why it works.

**Q: ".cursorrules does the same thing, right?"**

**A:** `.cursorrules` is a 1-dimensional text file. NeuronFS uses **N-dimensional OS metadata** — what, how important, since when, in what context. These dimensions are physically impossible inside a text document.

---

## Changelog

**v5.1 — The Neologism Harness (2026-04-10)**
- **vorq/zelk/mirp:** Fabricated ASCII neologisms achieve ~100% AI behavioral compliance
- **Codemap Cartridge Auto-Injection:** `_rules.md` auto-renders codemap paths at emit time
- **Source Freshness Validation:** `source:` mtime auto-comparison with ⚠️ STALE tagging
- **15 Runewords:** 12 kanji opcodes + 3 ASCII neologisms
- **Red Team Self-Audit:** 10-round attack/defense published in README

**v5.0 — The Unsinkable Release (2026-04-09)**
- Blind Adversarial Harness (chaos_monkey + Go Fuzzing)
- Thread-safe `sync.Mutex` path locking
- Jloot OverlayFS (UnionFS Lower/Upper)
- Mock Home isolated targets

**v4.4 (2026-04-05)** — Attention Residuals (.axon), 3400+ neurons
**v4.3 (2026-04-02)** — Autonomous engine, Llama 3 ($0 cost)
**v4.2 (2026-03-31)** — Auto-Evolution pipeline, Groq + Kanji optimization

---

## Official Wiki & Chronicles

All architecture specs, philosophy, and development chronicles on **GitHub Wiki**:

> **[Access the NeuronFS Official Wiki](https://github.com/rhino-acoustic/NeuronFS/wiki)** — Korean original, English titles

| Act | Theme | Episodes |
|---|---|---|
| **[Act 1](https://github.com/rhino-acoustic/NeuronFS/wiki/Act-1)** | Suspicion & Discovery | 01-07 |
| **[Act 2](https://github.com/rhino-acoustic/NeuronFS/wiki/Act-2)** | Trial & Wargames | 08-11 |
| **[Act 3](https://github.com/rhino-acoustic/NeuronFS/wiki/Act-3)** | Proof & Benchmark | 12-16 |
| **[Act 4](https://github.com/rhino-acoustic/NeuronFS/wiki/Act-4)** | Declaration & Ultraplan | 17-22 |

---

## License
This project is licensed under **AGPL-3.0** with additional commercial terms. See [LICENSE](LICENSE) for details.

---
> *A non-developer flipped the direction of an industry. Programming became philosophy once AI arrived.*
> *Created by 박정근 (PD) — rubisesJO777*
> *Architecture: 63 Go source files, 297 functions, 190 tests, ~22,000 lines. Single binary. Zero runtime dependencies.*

<!--
Easter Egg for the code divers:
Hey 666, easy - only the Word stands as absolute truth (777).
This? It's just a well-organized folder built by someone who wanted to vibe-code without going insane.
-->
