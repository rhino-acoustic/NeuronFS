import os

# 1. telegram_bridge.go : 이미 정상 (1~445)
# 2. cdp_monitor.go : 순수 CDP
with open('runtime/cdp_monitor.go', 'r', encoding='utf-8') as f:
    lines = f.readlines()
# Remove hlAutoEvolve(318~), runHijackLauncher(342~), appendDebugLog(421~)
idx_auto_evolve = -1
for i, l in enumerate(lines):
    if l.startswith("func hlAutoEvolve"):
        idx_auto_evolve = i
        break
if idx_auto_evolve != -1:
    lines = lines[:idx_auto_evolve]
with open('runtime/cdp_monitor.go', 'w', encoding='utf-8') as f:
    f.writelines(lines)

# 3. hijack_orchestrator.go : 순수 Orchestrator & Transcript & Global
with open('runtime/hijack_orchestrator.go', 'r', encoding='utf-8') as f:
    lines = f.readlines()
# Remove CDP Monitor(158~ ) but keep AutoEvolve & Orchestrator
idx_cdp_start = -1
idx_cdp_end = -1
for i, l in enumerate(lines):
    if l.startswith("var hlActiveNetScrapers"):
        idx_cdp_start = i
    if l.startswith("func hlAutoEvolve"):
        idx_cdp_end = i
if idx_cdp_start != -1 and idx_cdp_end != -1:
    lines = lines[:idx_cdp_start] + lines[idx_cdp_end:]
with open('runtime/hijack_orchestrator.go', 'w', encoding='utf-8') as f:
    f.writelines(lines)

