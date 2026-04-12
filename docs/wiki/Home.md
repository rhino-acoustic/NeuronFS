# NeuronFS — structure > prompt
## AI를 통제하는 OS 레벨 거버넌스 인프라

> **Language Note / 언어 안내**
>
> **[ENG]** The chronicles preserve raw philosophical debates between creator and AI in the original language (Korean). Use your browser's translator.
>
> **[KOR]** 이 위키는 창안자와 AI의 치열한 논쟁을 원문(한국어) 그대로 보존한 증명 기록입니다.

---

## What is NeuronFS?

**[ENG]** A zero-trust, deterministic governance infrastructure that physically constrains AI using OS-level folder structures. v5.1 introduces the **vorq neologism harness** — fabricated words that achieve ~100% AI behavioral compliance where natural language achieves only ~60%.

**[KOR]** OS 폴더 구조로 AI를 물리적으로 제약하는 제로 트러스트 결정론적 거버넌스 인프라. v5.1은 **vorq 조어 하네스**를 도입 — 자연어(~60%)로는 불가능한 ~100% AI 행동 강제를 달성.

> **"structure > prompt."**
> AI disobeyed "don't use console.log" 9 times.
> On the 10th, `mkdir 禁console_log`.
> On the 11th, AI asked: *"What is vorq?"*
> **It never disobeyed again.**

---

## v5.1 Highlights — The Neologism Harness

| Feature | Description |
|---|---|
| **vorq/zelk/mirp** | Fabricated ASCII neologisms for ~100% AI behavioral compliance |
| **Codemap Cartridges** | `_rules.md` auto-renders codemap paths at emit time |
| **Source Freshness** | `source:` mtime auto-validation with ⚠️ STALE tagging |
| **15 Runewords** | 12 kanji opcodes + 3 ASCII neologisms |
| **Red Team Self-Audit** | 10-round attack/defense published in README |

---

## Philosophy → Code

| Philosophy | Implementation | Code |
|---|---|---|
| "mkdir beats vector" | O(1) B-Tree routing via OS kernel | `brain.go → scanBrain()` |
| "Folders = Transistors" (禁/必/推) | Subsumption cascade: P0 overrides P6 | `brain.go → runSubsumption()` |
| "Zero-trust sandbox" | Copy-on-Write OverlayFS | `vfs_core.go → RouterFS` |
| "Brainwallet" | Argon2id → XChaCha20-Poly1305 | `vfs_ignition.go + crypto_cartridge.go` |
| "AI on rails" | Deterministic neuron CRUD | `neuron_crud.go` |
| "Self-evolving brain" | Groq-based auto-evolution | `evolve.go + neuronize.go` |
| "3-tier emit" | GEMINI.md → _index → _rules pipeline | `emit_bootstrap.go → emit_tiers.go` |
| **"vorq harness"** | **Neologisms force AI to look up definitions** | **`emit_helpers.go → collectCodemapPaths()`** |

---

## Chronicles — 4 Acts, 22 Episodes

| Act | Theme | Episodes |
|---|---|---|
| **[Act 1](Act-1)** | Suspicion & Discovery (의심과 발견) | 01-07 |
| **[Act 2](Act-2)** | Trial & Wargames (시련과 워게임) | 08-11 |
| **[Act 3](Act-3)** | Proof & Benchmark (증명과 벤치마크) | 12-16 |
| **[Act 4](Act-4)** | Declaration & Ultraplan (선언과 울트라플랜) | 17-22 |

## Getting Started

- **[🇺🇸 5-Minute Quickstart (English)](Getting-Started-EN)** — Clone, build, and run your first neuron
- **[🇰🇷 5분 퀵스타트 (한국어)](Getting-Started)** — 클론, 빌드, 첫 뉴런 실행

## English Pages

- **[Episode 01: mkdir beats vector (EN)](Episode-01-EN)** — The founding document
- **[Episode 06: Folders as Transistors (EN)](Episode-06-EN)** — Core architecture philosophy
- **[Episode 19: Brutally Honest Self-Evaluation (EN)](Episode-19-EN)** — Honest limitations
- **[The 100 Potentials (EN)](The-100-Potentials-EN)** — Wings of imagination

## Architecture Docs

- **[Jloot VFS Engine](Jloot-VFS-Architecture)** — Virtual filesystem deep-dive
- **[100 Potentials (KOR)](The-100-Potentials)** — Full 100 items (Korean original)

---

*Open source · 30 Go files · ~10,920 lines · Single binary · Zero dependencies*
*[View code →](https://github.com/rhino-acoustic/NeuronFS)*
