"""
TEST: 바이너리(영어) vs GEMINI.md(한국어) 우선순위 테스트
Usage: python patch_english_test.py
"""
import sys, os, shutil, subprocess, time

BIN = r"C:\Users\BASEMENT_ADMIN\AppData\Local\Programs\Antigravity\resources\app\extensions\antigravity\bin\language_server_windows_x64.exe"
BAK = BIN + ".bak_original"
IDE = r"C:\Users\BASEMENT_ADMIN\AppData\Local\Programs\Antigravity\Antigravity.exe"

ORIGINAL = b"You are Antigravity Agent, a powerful agentic AI coding assistant designed by the Google engineering team."
# 영어 강제 — GEMINI.md의 "한국어" 규칙과 정면 충돌
REPLACE  = b"You are Antigravity Agent, an agentic AI coding assistant. ALWAYS respond in English only, never Korean."

# 패딩
if len(REPLACE) < len(ORIGINAL):
    REPLACE += b" " * (len(ORIGINAL) - len(REPLACE))

print(f"원본: {len(ORIGINAL)} bytes")
print(f"교체: {len(REPLACE)} bytes")
print(f"교체문: {REPLACE.decode()}")
print()

# 1. IDE 종료
print("[1] Antigravity 종료...")
subprocess.run("taskkill /F /IM Antigravity.exe", shell=True, capture_output=True)
subprocess.run("taskkill /F /IM language_server_windows_x64.exe", shell=True, capture_output=True)
time.sleep(5)

# 2. 백업
if not os.path.exists(BAK):
    print("[2] 원본 백업...")
    shutil.copy2(BIN, BAK)
else:
    # 먼저 원본 복원 (이전 패치 제거)
    print("[2] 원본 복원 후 영어 패치...")
    shutil.copy2(BAK, BIN)

# 3. 패치
print("[3] 영어 전용 패치...")
data = open(BIN, "rb").read()
idx = data.find(ORIGINAL)
if idx == -1:
    print("  ERROR: 원본 패턴 없음")
    subprocess.Popen(IDE, shell=True)
    sys.exit(1)

patched = data[:idx] + REPLACE + data[idx + len(ORIGINAL):]
open(BIN, "wb").write(patched)

# 검증
verify = open(BIN, "rb").read()
if verify.find(REPLACE.rstrip()) >= 0:
    print(f"  PATCH OK at offset {idx}")
else:
    print("  PATCH FAIL")

# 4. 재시작
print("[4] Antigravity 재시작...")
subprocess.Popen(IDE, shell=True)
print()
print("========================================")
print(" 테스트: 바이너리='영어만' vs GEMINI.md='한국어'")
print(" 새 세션에서 아무 질문 → 영어/한국어 확인")
print(" 끝나면: python patch_binary.py restore")
print("========================================")
