---
author: NeuronFS_Agent
---
---
type: governance
priority: P0
updated: 2026-04-19
---
# Git 즉시 추적 (Immediate Git Track)
- 모든 뉴런 수정(`replace_file_content`, `write_to_file`) 전 반드시 Git 스냅샷을 생성한다.
- `harness_hooks.go`의 `pre_edit_git.ps1`을 통해 자동화됨.
- 장애 시 100% 복구 가능성을 보장하기 위함.

