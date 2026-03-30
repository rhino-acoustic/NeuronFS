Title: I got tired of Vector DBs for agent memory, so I built a 0KB governance engine using my local filesystem (NeuronFS)

**TL;DR:** I built an open-source tool ([NeuronFS](https://github.com/rhino-acoustic/NeuronFS)) that lets you control your AI agent's memory and rules purely through OS folders. No Vector DB, no Letta runtime server. A folder (`mkdir cortex/never_do_this`) becomes an immutable rule. It even has a physical circuit breaker (`bomb.neuron`) that halts the AI if it breaks safety thresholds 3 times. 

Context: 
File-based memory isn't entirely new. Letta recently shipped MemFS, and Engram uses vector DBs with Ebbinghaus curves. Both solve the "where to store memories" problem. Both require heavy infrastructure or specific servers.

NeuronFS solves a different problem: **Who decides which memories matter, and how do we physically stop the AI from bypassing safety rules?**

How it works: Your file system maps strictly to a brain structure.

```text
brain_v4/
├── brainstem/   # P0: Safety rules (read-only, immutable)
├── limbic/      # P1: Emotional signals (dopamine, contra)
├── hippocampus/ # P2: Session logs and recall
├── sensors/     # P3: Environment constraints (OS, tools)
├── cortex/      # P4: Learned knowledge (326+ neurons)
├── ego/         # P5: Personality and tone
└── prefrontal/  # P6: Goals and active plans
```

Why we built it (The "Governance" Edge):
1. **Vs Engram/VectorDBs:** Vector DBs have no emergency brakes. NeuronFS physically halts the process (`bomb.neuron`) if an agent makes the same mistake recursively. You don't have this level of physical safety in standard RAG/Mem0.
2. **Vs Axe/Agent Frameworks:** Lightweight agents are fast, but complex rules drift. Our `brainstem (P0)` always overrides frontend plans `prefrontal (P6)`. Folder hierarchy structurally prevents rule-based hallucinations at the root.
3. **Vs Anamnesis / Letta MemFS:** Letta's git-backed memory is great but requires their server. Anamnesis uses heavy DBs. We use Zero Infrastructure. Just your OS. A simple folder structure is the most perfect 0KB weight-calculation engine.

Limitations:
- By design, semantic search uses Jaccard similarity, not vector embeddings. 
- File I/O may bottleneck beyond ~10,000 neurons (we have 343 currently in production).
- Assumptions: A "one brain per user" model for now.

Numbers: 343+ neurons, 7 brain regions, 938+ total activations. Full brain scan: ~1ms. Disk usage: ~4.3MB. MIT license.

GitHub Repo: https://github.com/rhino-acoustic/NeuronFS

I'd love to hear feedback from this community—especially on the Subsumption Cascade model. Does physical folder priority make sense for hard agent safety? What attack vectors am I missing?
