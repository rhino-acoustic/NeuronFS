import os

with open('runtime/telegram_bridge.go', 'r', encoding='utf-8') as f:
    lines = f.readlines()
# Keep lines 0 to 444 (Telegram logic)
with open('runtime/telegram_bridge.go', 'w', encoding='utf-8') as f:
    f.writelines(lines[:445])

with open('runtime/cdp_monitor.go', 'r', encoding='utf-8') as f:
    lines = f.readlines()
# Keep imports/package (lines 0 to 21) AND CDP logic (445 to 483, plus 592 to end)
with open('runtime/cdp_monitor.go', 'w', encoding='utf-8') as f:
    f.writelines(lines[:21])
    f.writelines(lines[445:483])
    f.writelines(lines[611:])
