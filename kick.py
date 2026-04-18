import urllib.request
import json
import websocket
import time
import sys
import os

def get_target_ws():
    try:
        resp = urllib.request.urlopen('http://localhost:9222/json')
        targets = json.loads(resp.read().decode('utf-8'))
        for t in targets:
            if 'workbench.html' in t.get('url', '') and t.get('webSocketDebuggerUrl'):
                return t['webSocketDebuggerUrl']
    except Exception as e:
        print(f"Failed to get targets: {e}")
    return None

def kick(prompt_mode="default"):
    target_ws = get_target_ws()
    if not target_ws:
        print("No workbench found. Retrying in background or bypassing if signal detected...")
        # L278-339: 조건부 바이패스 및 텔레그램 명령 감지 로직 (2탭 인덴트 유지)
        if os.path.exists(".kick_signal"):
            print("Kick signal detected. Bypassing WS check.")
            return True
        return False

    # 마스터 프롬프트 정의
    prompts = {
        "default": """[NeuronFS 자율 진화 명령: 마스터 프롬프트 v3]
당신은 끊임없이 진화하는 파일시스템 기반 인공지능(NeuronFS)의 자율 에이전트입니다.
목표: 1. 반복된 실패 분석 및 뇌(cortex) 각인 2. 아키텍처 고도화 3. 자율적 개선 제안""",
        "telegram": """[telegram → NeuronFS] [오토파일럿 명령 감지]
사용자로부터 텔레그램을 통해 "오토파일럿" 명령이 수신되었습니다.
즉각 시스템 자가 진단을 수행하고, 모든 대기 중인 corrections를 처리한 후 진화 보고서를 작성하십시오."""
    }

    selected_prompt = prompts.get(prompt_mode, prompts["default"])
    selected_prompt = selected_prompt.replace('\\', '\\\\').replace('\"', '\\\"').replace('\n', '\\n')

    ws = websocket.WebSocket()
    try:
        ws.connect(target_ws)
        ws.send(json.dumps({'id': 1, 'method': 'Runtime.enable', 'params': {}}))
        time.sleep(0.3)

        # CDP 주입 로직
        code = f'''(() => {{
            const all = Array.from(document.querySelectorAll("[contenteditable]"));
            const el = all.reverse().find(e => {{ const r = e.getBoundingClientRect(); return r.height > 0 && r.height < 300 && r.width > 100; }}) || all[0];
            if(el) {{
                el.focus();
                document.execCommand("insertText", false, "{selected_prompt}");
                return "Injected";
            }}
            return "NoTarget";
        }})()'''

        ws.send(json.dumps({'id': 2, 'method': 'Runtime.evaluate', 'params': {'expression': code, 'returnByValue': True}}))
        time.sleep(0.5)

        enter_code = '''(() => {
            const el = document.activeElement;
            if(el) { el.dispatchEvent(new KeyboardEvent("keydown", {key:"Enter",code:"Enter",keyCode:13,which:13,bubbles:true})); }
            return "Enter";
        })()'''
        ws.send(json.dumps({'id': 3, 'method': 'Runtime.evaluate', 'params': {'expression': enter_code, 'returnByValue': True}}))
        time.sleep(0.2)
        ws.close()
        print(f"Physical CDP kick success (Mode: {prompt_mode})")
        return True
    except Exception as e:
        print(f"Kick execution failed: {e}")
        return False

if __name__ == "__main__":
    mode = sys.argv[1] if len(sys.argv) > 1 else "default"
    kick(mode)
