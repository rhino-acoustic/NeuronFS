# NeuronFS for Games — Folder-Based World State

> **메모리 DB 없이 게임 월드를 관리한다. `mkdir`이 곧 스폰이고, `rm`이 곧 처치다.**

## 왜?

| 전통 방식 | NeuronFS |
|---|---|
| Redis/SQLite에 NPC 상태 저장 | **폴더 = NPC 상태** |
| 서버 크래시 → 데이터 유실 | 폴더는 남는다 |
| 디버깅: SQL 쿼리 | `ls npc_goblin_01/` |
| 의존성: Redis + DB 드라이버 | **의존성 0** |

## 구조

```
game_world/
├── brainstem/          ← P0: 절대 규칙 (물리 법칙)
│   └── 禁/
│       ├── clip_through_walls/1.neuron
│       └── negative_hp/1.neuron
│
├── zone_forest/        ← 숲 영역
│   ├── npc_goblin_scout/
│   │   ├── 3.neuron                ← HP = 3
│   │   ├── aggro.axon              ← TARGET: player_01
│   │   └── 禁/healing/1.neuron     ← 치유 불가 (독 디버프)
│   │
│   ├── npc_goblin_chief/
│   │   ├── 20.neuron               ← HP = 20 (보스)
│   │   ├── loot_table/
│   │   │   ├── rare_sword/1.neuron
│   │   │   └── gold_pouch/3.neuron
│   │   └── 推/retreat_hp_low/1.neuron  ← HP 낮으면 후퇴 (추천)
│   │
│   └── item_chest_01/
│       ├── 1.neuron                ← 열린 횟수
│       └── potion/5.neuron         ← 물약 5개
│
├── zone_castle/        ← 성 영역
│   ├── npc_merchant/
│   │   ├── 10.neuron               ← 신뢰도 10
│   │   └── quest_escort/
│   │       └── 1.neuron            ← 활성 퀘스트
│   ├── npc_guard/
│   │   ├── 15.neuron               ← HP = 15
│   │   └── 禁/attack_king/1.neuron ← 왕 공격 절대 금지
│   └── bomb.neuron                 ← 성 봉쇄 (공성전 이벤트)
│
├── player_01/          ← 플레이어 상태
│   ├── 50.neuron                   ← 레벨 50
│   ├── inventory/
│   │   ├── sword/3.neuron          ← 검 3개
│   │   ├── potion/5.neuron         ← 물약 5개
│   │   └── gold/500.neuron         ← 골드 500
│   ├── 禁/pvp/1.neuron             ← PVP 금지 (안전 지역)
│   └── skills/
│       ├── fireball/10.neuron      ← 파이어볼 레벨 10
│       └── heal/5.neuron           ← 힐 레벨 5
│
└── player_02/
    ├── 30.neuron
    └── inventory/
        └── shield/1.neuron
```

## 조작

```bash
# 월드 초기화
neuronfs game_world --init

# 고블린 스폰
neuronfs game_world --grow zone_forest/npc_goblin_scout

# 고블린에게 데미지 (HP +1, 역전 카운터로 사용 시)
neuronfs game_world --fire zone_forest/npc_goblin_scout

# 고블린 처치 (bomb = 비활성화)
neuronfs game_world --signal zone_forest/npc_goblin_scout bomb

# 성 봉쇄 (bomb.neuron → zone_castle 전체 잠금)
neuronfs game_world --signal zone_castle bomb

# 봉쇄 해제
rm game_world/zone_castle/bomb.neuron

# 월드 상태 확인
neuronfs game_world --diag
```

## NPC AI 거버넌스 = Subsumption Cascade

NPC의 행동 우선순위가 NeuronFS의 7계층과 동일:

```
P0 brainstem: 절대 규칙 (벽 뚫기 금지, 아군 공격 금지)
P1 limbic:    감정 (공포 → 도주, 분노 → 공격)
P4 cortex:    전투 로직 (순찰, 추격)
P6 prefrontal: 목표 (퀘스트 완수, 거점 방어)
```

**P0의 `禁attack_ally`는 P4의 전투 로직을 항상 이긴다.** — 이건 게임 AI의 가장 기본적인 요구사항이다.

## Subsumption이 해결하는 게임 문제

| 게임 문제 | NeuronFS 해법 |
|---|---|
| NPC가 아군을 공격함 | `brainstem/禁/attack_ally/` — P0가 모든 전투 로직을 억제 |
| 보스가 벽 밖으로 나감 | `brainstem/禁/clip_through_walls/` |
| 성 봉쇄 이벤트 | `zone_castle/bomb.neuron` — 영역 전체 즉시 비활성화 |
| HP가 음수 | `brainstem/禁/negative_hp/` |
| NPC 행동 디버깅 | `ls -la npc_goblin_scout/` — DB 쿼리 불필요 |

## 핵심 가치

```
mkdir = 스폰
rm = 처치  
mv = 이동
cat = 상태 확인
bomb.neuron = 영역 잠금

메모리 DB 의존성: 0
사용법 학습 시간: 0 (ls/mkdir을 아는 사람이면 끝)
```
