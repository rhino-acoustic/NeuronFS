#!/bin/bash
# NeuronFS Game World Demo — Setup Script
# Creates a sample game world with zones, NPCs, and players.
# No code changes needed. Pure NeuronFS primitives.

WORLD="game_world_demo"

echo "🎮 NeuronFS Game World Demo"
echo "═══════════════════════════"

# Initialize brain
echo "[1/6] Initializing world..."
cd "$(dirname "$0")/../.." 
# Use neuronfs or manual mkdir

# P0: Physics rules (unbreakable)
mkdir -p "$WORLD/brainstem/禁/clip_through_walls"
echo "" > "$WORLD/brainstem/禁/clip_through_walls/1.neuron"
mkdir -p "$WORLD/brainstem/禁/negative_hp"
echo "" > "$WORLD/brainstem/禁/negative_hp/1.neuron"
mkdir -p "$WORLD/brainstem/禁/attack_ally"
echo "" > "$WORLD/brainstem/禁/attack_ally/1.neuron"

# Zone: Forest
echo "[2/6] Spawning forest zone..."
mkdir -p "$WORLD/sensors/zone_forest/npc_goblin_scout"
echo "" > "$WORLD/sensors/zone_forest/npc_goblin_scout/3.neuron"
echo "TARGET: sensors/zone_forest/player_01" > "$WORLD/sensors/zone_forest/npc_goblin_scout/aggro.axon"

mkdir -p "$WORLD/sensors/zone_forest/npc_goblin_chief/loot_table/rare_sword"
echo "" > "$WORLD/sensors/zone_forest/npc_goblin_chief/20.neuron"
echo "" > "$WORLD/sensors/zone_forest/npc_goblin_chief/loot_table/rare_sword/1.neuron"

mkdir -p "$WORLD/sensors/zone_forest/item_chest_01"
echo "" > "$WORLD/sensors/zone_forest/item_chest_01/1.neuron"

# Zone: Castle
echo "[3/6] Building castle zone..."
mkdir -p "$WORLD/sensors/zone_castle/npc_merchant/quest_escort"
echo "" > "$WORLD/sensors/zone_castle/npc_merchant/10.neuron"
echo "" > "$WORLD/sensors/zone_castle/npc_merchant/quest_escort/1.neuron"

mkdir -p "$WORLD/sensors/zone_castle/npc_guard/禁/attack_king"
echo "" > "$WORLD/sensors/zone_castle/npc_guard/15.neuron"
echo "" > "$WORLD/sensors/zone_castle/npc_guard/禁/attack_king/1.neuron"

# Players
echo "[4/6] Creating players..."
mkdir -p "$WORLD/cortex/player_01/inventory/sword"
mkdir -p "$WORLD/cortex/player_01/inventory/potion"
mkdir -p "$WORLD/cortex/player_01/inventory/gold"
mkdir -p "$WORLD/cortex/player_01/skills/fireball"
mkdir -p "$WORLD/cortex/player_01/skills/heal"
echo "" > "$WORLD/cortex/player_01/50.neuron"
echo "" > "$WORLD/cortex/player_01/inventory/sword/3.neuron"
echo "" > "$WORLD/cortex/player_01/inventory/potion/5.neuron"
echo "" > "$WORLD/cortex/player_01/inventory/gold/500.neuron"
echo "" > "$WORLD/cortex/player_01/skills/fireball/10.neuron"
echo "" > "$WORLD/cortex/player_01/skills/heal/5.neuron"

mkdir -p "$WORLD/cortex/player_02/inventory/shield"
echo "" > "$WORLD/cortex/player_02/30.neuron"
echo "" > "$WORLD/cortex/player_02/inventory/shield/1.neuron"

# Remaining regions (empty but required)
echo "[5/6] Setting up governance layers..."
mkdir -p "$WORLD/limbic"
mkdir -p "$WORLD/hippocampus/session_log"
mkdir -p "$WORLD/ego"
mkdir -p "$WORLD/prefrontal"

echo "[6/6] Done!"
echo ""
echo "🌍 World created at: $WORLD/"
echo ""
echo "Try these commands:"
echo "  neuronfs $WORLD --diag           # View world tree"
echo "  neuronfs $WORLD --emit gemini    # Generate NPC rules"
echo "  neuronfs $WORLD --fire sensors/zone_forest/npc_goblin_scout  # Damage goblin"
echo ""
echo "  # Castle siege event:"
echo "  touch $WORLD/sensors/zone_castle/bomb.neuron   # Lock castle"
echo "  rm $WORLD/sensors/zone_castle/bomb.neuron       # Unlock castle"
