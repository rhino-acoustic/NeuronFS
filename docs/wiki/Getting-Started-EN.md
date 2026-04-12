# Getting Started (English)
## Experience NeuronFS in 5 Minutes

> 🇰🇷 **[한국어 원문 (Korean Original)](Getting-Started)**

---

## Prerequisites

- **Go 1.21+** (https://go.dev/dl/)
- **Git**
- Windows / macOS / Linux all supported

---

## 1. Clone and Build

```bash
# Clone the repository
git clone https://github.com/rhino-acoustic/NeuronFS.git
cd NeuronFS/runtime

# Build (zero dependencies — single binary)
go build -o neuronfs .

# Verify
./neuronfs --diag
```

Build result: **~4MB single binary.** Zero external dependencies.

## 2. Inspect the Brain

```bash
# Output current brain state (read-only, no file changes)
./neuronfs --emit

# Visualize full brain tree
./neuronfs --diag
```

`--emit` outputs the currently active neuron rules to stdout. This alone gives you an instant understanding of how NeuronFS works.

## 3. Explore the Brain Structure

```
brain_v4/
├── brainstem/     (P0 — Absolute principles, brainstem)
│   ├── 禁/hardcoding/
│   └── 禁/security_violation/
├── limbic/        (P1 — Emotion filters)
├── hippocampus/   (P2 — Memory, error patterns)
├── sensors/       (P3 — Environmental constraints)
├── cortex/        (P4 — Knowledge, coding rules)
│   ├── dev/
│   ├── methodology/
│   └── frontend/
├── ego/           (P5 — Personality, tone)
└── prefrontal/    (P6 — Goals, planning)
```

**Core rule:** Lower P always physically overrides higher P.
brainstem(P0) 禁 rules always beat cortex(P4) dev rules.

### Opcodes are Runewords

If you played Diablo 2 — **NeuronFS opcodes work exactly like Runewords.**

A Runeword is a specific combination of runes socketed into the right item base. The magic isn't in any single rune — it's in the **exact combination + exact socket type**.

| Opcode | Rune | Effect | Example |
|---|---|---|---|
| `禁/` | Zod | **Absolute prohibition** — AI physically cannot cross | `禁/hardcoding/` |
| `必/` | Ber | **Mandatory gate** — AI must pass through | `必/manager_approval/` |
| `推/` | Ist | **Recommendation** — soft nudge, overridable | `推/test_code/` |
| `.axon` | Jah | **Teleport** — connects two distant brain regions | `推/insurance.axon => [claims/]` |
| `bomb` | El Rune | **Kill switch** — entire region freezes | `bomb.neuron` |

> *"Socket a 禁 rune into a brainstem folder = indestructible wall. Socket a 推 rune into cortex = soft suggestion. The folder is the socket. The opcode is the rune. The combination is the Runeword."*

### Nested Opcodes — Prohibition + Resolution in One

**NeuronFS's killer pattern:** nest opcodes to chain prohibition and solution hierarchically.

```
brainstem/禁/no_shift/必/stack_solution/
         ↑ prohibition  ↑ resolution
```

Read as: *"Prohibit shift (禁), but mandate stacking as the solution (必)."*

> *"A folder name is a philosophical declaration. Nesting creates hierarchy. Hierarchy creates governance."*

## 4. Create Your First Neuron

```bash
# Create a new neuron
./neuronfs --grow cortex/dev/my_first_rule

# Fire the neuron (activation +1)
./neuronfs --fire cortex/dev/my_first_rule

# Send dopamine signal (reinforce)
./neuronfs --signal dopamine cortex/dev/my_first_rule

# Compile brain → system prompts for your IDE
./neuronfs --emit all
```

## 5. Autonomous Evolution

```bash
# Set Groq API key (free tier available)
export GROQ_API_KEY=your_key_here

# Start brain evolution (dry run = proposals only)
./neuronfs --evolve

# Execute evolution
./neuronfs --evolve --apply
```

The AI analyzes accumulated signals and auto-generates new neurons.

---

## Demo Brain: 672 Neurons Starter Pack

The `brain_v4/` directory in this repo is a **fully functioning 672-neuron brain.** It's usable the instant you clone.

| Region | Neurons | Description |
|---|---|---|
| brainstem (P0) | ~10 | Absolute prohibition rules |
| hippocampus (P2) | ~103 | Error patterns, episodic memory |
| cortex (P4) | ~348 | Dev knowledge, methodology |
| sensors (P3) | ~4 | Environment variables |
| ego (P5) | ~1 | Behavioral style |
| prefrontal (P6) | ~2 | Long-term goals |

> *"These 672 neurons were forged from thousands of hours of real-world AI agent operation and corrections. Clone and instantly equip a senior developer's brain."*

---

## Next Steps

- [Episode 01: mkdir beats vector (EN)](Episode-01-EN) — Core philosophy
- [Episode 06: Folders as Transistors (EN)](Episode-06-EN) — Architecture deep-dive
- [Episode 19: Brutally Honest Self-Evaluation (EN)](Episode-19-EN) — Honest limitations
- [The 100 Potentials (EN)](The-100-Potentials-EN) — Wings of imagination

---

[Home](Home)
