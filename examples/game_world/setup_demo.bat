@echo off
chcp 65001 >nul
REM NeuronFS Game World Demo — Windows Setup
REM mkdir = spawn, touch = state, bomb.neuron = zone lock

set WORLD=game_world_demo

echo 🎮 NeuronFS Game World Demo
echo ═══════════════════════════

echo [1/6] Physics rules (P0 — unbreakable)...
mkdir "%WORLD%\brainstem\禁\clip_through_walls" 2>nul
echo. > "%WORLD%\brainstem\禁\clip_through_walls\1.neuron"
mkdir "%WORLD%\brainstem\禁\negative_hp" 2>nul
echo. > "%WORLD%\brainstem\禁\negative_hp\1.neuron"
mkdir "%WORLD%\brainstem\禁\attack_ally" 2>nul
echo. > "%WORLD%\brainstem\禁\attack_ally\1.neuron"

echo [2/6] Forest zone...
mkdir "%WORLD%\sensors\zone_forest\npc_goblin_scout" 2>nul
echo. > "%WORLD%\sensors\zone_forest\npc_goblin_scout\3.neuron"
mkdir "%WORLD%\sensors\zone_forest\npc_goblin_chief\loot_table\rare_sword" 2>nul
echo. > "%WORLD%\sensors\zone_forest\npc_goblin_chief\20.neuron"
echo. > "%WORLD%\sensors\zone_forest\npc_goblin_chief\loot_table\rare_sword\1.neuron"

echo [3/6] Castle zone...
mkdir "%WORLD%\sensors\zone_castle\npc_merchant\quest_escort" 2>nul
echo. > "%WORLD%\sensors\zone_castle\npc_merchant\10.neuron"
echo. > "%WORLD%\sensors\zone_castle\npc_merchant\quest_escort\1.neuron"
mkdir "%WORLD%\sensors\zone_castle\npc_guard\禁\attack_king" 2>nul
echo. > "%WORLD%\sensors\zone_castle\npc_guard\15.neuron"
echo. > "%WORLD%\sensors\zone_castle\npc_guard\禁\attack_king\1.neuron"

echo [4/6] Players...
mkdir "%WORLD%\cortex\player_01\inventory\sword" 2>nul
mkdir "%WORLD%\cortex\player_01\inventory\potion" 2>nul
mkdir "%WORLD%\cortex\player_01\inventory\gold" 2>nul
mkdir "%WORLD%\cortex\player_01\skills\fireball" 2>nul
mkdir "%WORLD%\cortex\player_01\skills\heal" 2>nul
echo. > "%WORLD%\cortex\player_01\50.neuron"
echo. > "%WORLD%\cortex\player_01\inventory\sword\3.neuron"
echo. > "%WORLD%\cortex\player_01\inventory\potion\5.neuron"
echo. > "%WORLD%\cortex\player_01\inventory\gold\500.neuron"
echo. > "%WORLD%\cortex\player_01\skills\fireball\10.neuron"
echo. > "%WORLD%\cortex\player_01\skills\heal\5.neuron"

echo [5/6] Governance layers...
mkdir "%WORLD%\limbic" 2>nul
mkdir "%WORLD%\hippocampus\session_log" 2>nul
mkdir "%WORLD%\ego" 2>nul
mkdir "%WORLD%\prefrontal" 2>nul

echo [6/6] Done!
echo.
echo 🌍 World: %WORLD%\
echo.
echo Try:
echo   neuronfs %WORLD% --diag
echo   neuronfs %WORLD% --fire sensors/zone_forest/npc_goblin_scout
echo   echo. ^> %WORLD%\sensors\zone_castle\bomb.neuron  ← castle siege!
