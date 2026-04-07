# Demo Brain — NeuronFS Starter Kit

This is a **sanitized skeleton brain** for demonstration purposes.
Clone and immediately experience NeuronFS governance.

## Structure

```
demo_brain/
├── brainstem/          (P0 — Absolute principles)
│   ├── 禁/
│   │   ├── hardcoding/
│   │   ├── data_leak/
│   │   └── fallback/
│   └── 기본원칙/
├── limbic/             (P1 — Emotional filter)
├── hippocampus/        (P2 — Memory, error patterns)
│   └── 에러패턴/
├── sensors/            (P3 — Environment)
│   └── 환경/
├── cortex/             (P4 — Knowledge, skills)
│   ├── dev/
│   │   ├── 禁/console_log/
│   │   └── 推/테스트코드/
│   └── methodology/
│       ├── 검증_후_보고/
│       └── 시제_준수/
├── ego/                (P5 — Persona, tone)
│   └── 행동양식/
└── prefrontal/         (P6 — Goals, plans)
    └── 방법론/
```

## Quick Start

```bash
cd NeuronFS/runtime
go build -o neuronfs .
./neuronfs --brain ../demo_brain --emit all
```

## Note

This is a **starter skeleton** with 13 neurons.
The production brain (brain_v4/) contains 675+ neurons — available as encrypted .jloot cartridges.

## License

AGPL-3.0 — See [LICENSE](../LICENSE)
