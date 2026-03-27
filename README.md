# NeuronFS

> **AI를 제어하는 건 프롬프트가 아니라 폴더 구조다.**

![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)

**[한국어 README →](README.ko.md)**

---

## The Problem

Every AI coding agent (Cursor, Windsurf, Copilot, Gemini Code Assist) reads a rules file on startup — `.cursorrules`, `GEMINI.md`, `AGENTS.md`, etc.

These files grow into **massive, unstructured text walls** that:
- Have no priority system (rule #47 has the same weight as rule #1)
- Can't grow or shrink based on what actually works
- Become unmaintainable beyond ~50 rules
- Offer no visibility into which rules the AI actually follows

## The Insight

**What if the rules file wasn't a file at all — but a directory tree?**

```
brain/
├── brainstem/                    # Absolute rules — always enforced
│   ├── never_delete_production/
│   │   └── 99.neuron            # 99 = activation strength
│   └── verify_before_deploy/
│       └── 50.neuron
├── cortex/                       # Knowledge — applied contextually
│   ├── frontend/
│   │   └── react/
│   │       └── hooks_pattern/
│   │           └── 15.neuron
│   └── backend/
│       └── supabase/
│           └── rls_always_on/
│               └── 20.neuron
└── sensors/                      # Environmental constraints
    └── nas_write_cmd_only/
        └── 30.neuron
```

A scanner reads this tree and compiles it into the rules file — sorted by priority, weighted by activation counter, structured by hierarchy.

**Folder = concept. File = signal strength. Path = full sentence.**

`cortex/frontend/react/hooks_pattern/15.neuron` says: *"In the cortex (knowledge), for frontend, specifically React, apply hooks patterns — with activation weight 15."*

## Core Axioms

| Axiom | Meaning |
|---|---|
| **Folder = Neuron** | Each directory is a concept. Its name is its meaning. |
| **Path = Sentence** | The full path reads as a natural language rule. |
| **Counter = Strength** | `N.neuron` — higher N = stronger enforcement. |
| **Depth = Specificity** | `cortex/frontend/` is broad. `cortex/frontend/react/hooks/useCallback/` is precise. |

## Why Folders Beat Files

| | Single Rules File | NeuronFS (Folder Tree) |
|---|---|---|
| **Priority** | None (flat text) | Structural hierarchy |
| **Add a rule** | Edit text, hope order is right | `mkdir` + `touch` — done |
| **Remove a rule** | Find and delete text | `rm -rf` the folder |
| **Track changes** | Diff a monolith | `git log` per folder |
| **Scale** | Breaks at ~100 rules | Tested at 150+ neurons, path to 1000+ |
| **Visibility** | Read the whole file | `tree brain/` — instant audit |
| **Cost** | ₩0 | ₩0 |

## Self-Growth

The AI can modify its own rule tree during operation:

```bash
# User corrects the AI → AI creates a new rule
mkdir -p brain/cortex/frontend/no_console_log
touch brain/cortex/frontend/no_console_log/1.neuron

# Rule works well → increment counter (reinforcement)
mv 1.neuron 2.neuron

# Same mistake 3 times → circuit breaker
touch brain/cortex/frontend/no_console_log/bomb.neuron
```

Every `git commit` becomes a cognitive growth log.

## Quick Start

```bash
# 1. Create a brain
mkdir -p brain/brainstem/verify_before_deliver
touch brain/brainstem/verify_before_deliver/1.neuron

# 2. Add domain knowledge
mkdir -p brain/cortex/frontend/react/hooks_pattern
touch brain/cortex/frontend/react/hooks_pattern/1.neuron

# 3. Scan — a simple script reads the tree and generates your rules file
# (Scanner implementation is environment-specific)
```

The scanner compiles the folder tree into whatever format your AI agent needs: `.cursorrules`, `GEMINI.md`, `AGENTS.md`, or plain text.

## Status

This is a **working concept** in daily production use with one AI agent (Gemini/Antigravity).

What's proven:
- ✅ Folder-based rules are faster to manage than text files
- ✅ Counter-based activation weight works for priority
- ✅ AI can `mkdir` to create its own rules in real-time
- ✅ `git log` provides a cognitive development history

What's in progress:
- 🔧 Go runtime for automated scanning and injection
- 🔧 Multi-agent compatibility testing
- 🔧 Performance benchmarks at scale (500+ neurons)

---

> 📜 **[Read the Manifesto →](MANIFESTO_EN.md)** — the full philosophical framework behind NeuronFS.

---

**Created by:** [@rhino-acoustic](https://github.com/rhino-acoustic)
