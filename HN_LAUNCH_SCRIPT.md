Title: Show HN: NeuronFS – Brain-like governance for AI agent memory (zero infra)
URL: https://github.com/rhino-acoustic/NeuronFS

Text:
Hi HN, I built NeuronFS — a governance layer for AI agent memory that runs entirely on your local file system. No database, no server, no runtime process.

File-based memory isn't entirely new, but we are approaching it from a "hard governance" angle rather than pure storage. 

Why we built this instead of using Vector DBs or existing Frameworks:
- Engram: Vector DBs have no emergency brakes. NeuronFS is the only system with `bomb.neuron`, a physical circuit breaker that halts output immediately if a threshold is crossed 3 times.
- Axe: Lightweight agents execute fast, but rule-based hallucinations still happen. Our `brainstem` (P0) structurally prevents P6 agents from bypassing core safety.
- Anamnesis & Mem0: No need for heavy DBs. A simple folder structure (`mkdir 禁fallback`) is the most perfect 0KB weight-calculation engine.
- Letta (MemFS): Their git-backed memory is great, but requires a Letta server. We use Zero Infrastructure. Just your OS.

How it works (The Medium is the Core): Your file system maps to a brain structure.

brain_v4/
├── brainstem/   # P0: Safety rules (read-only, immutable)
├── limbic/      # P1: Emotional signals (dopamine, contra)
├── hippocampus/ # P2: Session logs and recall
├── sensors/     # P3: Environment constraints (OS, tools)
├── cortex/      # P4: Learned knowledge (326+ neurons)
├── ego/         # P5: Personality and tone
└── prefrontal/  # P6: Goals and active plans

1. Subsumption Cascade: P0 always overrides P6. The AI cannot bypass its safety rules to achieve a goal — enforced by folder priority, not by prompt engineering.
2. Automatic Forgetting & Firing: Neurons have activation counters in their filenames (`skill.3.neuron` = referenced 3 times). Repeated mistakes trigger a `.contra` file or a physical circuit breaker (`bomb.neuron`).
3. OS-level Security: If an attacker gets write access, prompt injection is your least concern. Filesystem security is more transparent than black-box DB poisoning.

Honest limitations:
- Semantic search uses Jaccard similarity, not vector embeddings (by design).
- File system I/O may bottleneck beyond ~10,000 neurons (currently 343 in production).
- The system assumes a "one brain per user" model for now.

Numbers: 343+ neurons, 7 brain regions, 938+ total activations. Full brain scan: ~1ms. Disk usage: ~4.3MB. MIT license.

I'd appreciate feedback — especially around the governance model. Does subsumption-based priority make sense for agent safety? What attack vectors am I missing?
