"""
NeuronFS Binary Patcher — language_server system prompt 교정/원복
Usage:
  python patch_binary.py patch    # 패치
  python patch_binary.py restore  # 원복
"""
import sys, os, shutil, subprocess, time

BIN = r"C:\Users\BASEMENT_ADMIN\AppData\Local\Programs\Antigravity\resources\app\extensions\antigravity\bin\language_server_windows_x64.exe"
BAK = BIN + ".bak_original"
IDE = r"C:\Users\BASEMENT_ADMIN\AppData\Local\Programs\Antigravity\Antigravity.exe"

ORIGINAL = b"You are Antigravity Agent, a powerful agentic AI coding assistant designed by the Google engineering team."
REPLACE  = b"You are NeuronFS-Antigravity, an agentic AI coding assistant. Always think in Korean, answer in Korean."

# 패딩 (바이트 수 일치)
if len(REPLACE) < len(ORIGINAL):
    REPLACE += b" " * (len(ORIGINAL) - len(REPLACE))

def kill_ide():
    print("[1] Antigravity 종료...")
    subprocess.run("taskkill /F /IM Antigravity.exe", shell=True, capture_output=True)
    subprocess.run("taskkill /F /IM language_server_windows_x64.exe", shell=True, capture_output=True)
    time.sleep(5)

def start_ide():
    print("[*] Antigravity 재시작...")
    subprocess.Popen(IDE, shell=True)

def patch():
    kill_ide()
    
    # 백업
    if not os.path.exists(BAK):
        print("[2] 원본 백업 생성...")
        shutil.copy2(BIN, BAK)
    else:
        print("[2] 원본 백업 이미 존재")
    
    # 패치
    print("[3] 바이너리 패치...")
    data = open(BIN, "rb").read()
    idx = data.find(ORIGINAL)
    if idx == -1:
        # 이미 패치됨?
        if data.find(REPLACE.rstrip()) >= 0:
            print("  이미 패치되어 있음!")
        else:
            print("  ERROR: 원본 패턴을 찾을 수 없음")
        start_ide()
        return
    
    patched = data[:idx] + REPLACE + data[idx + len(ORIGINAL):]
    open(BIN, "wb").write(patched)
    
    # 검증
    verify = open(BIN, "rb").read()
    if verify.find(REPLACE.rstrip()) >= 0:
        print("  ✅ 패치 성공!")
        print(f"  위치: {idx}")
        print(f"  교체: {REPLACE.decode()}")
    else:
        print("  ❌ 패치 검증 실패")
    
    start_ide()

def restore():
    if not os.path.exists(BAK):
        print("ERROR: 백업 파일 없음!")
        return
    
    kill_ide()
    print("[2] 원본 복원...")
    shutil.copy2(BAK, BIN)
    print("  ✅ 복원 완료")
    start_ide()

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: python patch_binary.py [patch|restore]")
        sys.exit(1)
    
    cmd = sys.argv[1].lower()
    if cmd == "patch":
        patch()
    elif cmd == "restore":
        restore()
    else:
        print(f"Unknown: {cmd}")
