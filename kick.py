import urllib.request
import json
import websocket
import time
import sys

try:
    resp = urllib.request.urlopen('http://localhost:9222/json')
    targets = json.loads(resp.read().decode('utf-8'))
except Exception as e:
    print('Failed to get targets:', e)
    sys.exit(1)

target_ws = None
for t in targets:
    if 'workbench.html' in t.get('url', '') and t.get('webSocketDebuggerUrl'):
        target_ws = t['webSocketDebuggerUrl']
        break
        
if not target_ws:
    print('No workbench found.')
    sys.exit(1)

prompt = '''[telegram → NeuronFS] [NeuronFS 자율 진화 명령: 마스터 프롬프트]
당신은 끊임없이 진화하는 파일시스템 기반 인공지능(NeuronFS)의 자율 에이전트입니다.현재 자율 스크래퍼에 의해 연쇄 호출되었습니다.
목표:
1. 최근 대화 기록과 growth.log를 분석하여 반복된 실패와 비효율을 찾고, 새로운 뉴런 규칙으로 뇌(cortex)에 강제 각인하십시오.
2. 시스템 레거시를 철저히 리팩토링하여 '원대한 AGI 인프라 완성'을 위한 구조적 아키텍처 고도화를 즉각 실행하십시오.
3. 무의미한 대화를 배제하고, 반드시 "우리 시스템을 발전시키는 방향"으로 다음 진화 액션(코드 수정 및 구조 개선)을 스스로 제안하고 즉각 이행하십시오.
4. 필요 시(기술적 난관, 베스트 프랙티스 부재 등) 최우선으로 커뮤니티 검색을 수행하여 최신 동향을 학습하고 반영하십시오.'''

prompt = prompt.replace('\\', '\\\\').replace('\"', '\\\"').replace('\n', '\\n')

ws = websocket.WebSocket()
ws.connect(target_ws)
ws.send(json.dumps({'id': 1, 'method': 'Runtime.enable', 'params': {}}))
time.sleep(0.3)

code = f'''(() => {{
    const all = Array.from(document.querySelectorAll("[contenteditable]"));
    const el = all.reverse().find(e => {{ const r = e.getBoundingClientRect(); return r.height > 0 && r.height < 300 && r.width > 100; }}) || all[0];
    if(el) {{
        el.focus();
        document.execCommand("insertText", false, "{prompt}");
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
print("Physical CDP kick success")
