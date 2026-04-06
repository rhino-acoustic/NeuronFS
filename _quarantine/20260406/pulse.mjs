/**
 * pulse.mjs — CDP 기반 IDE 챗창 텍스트 주입기
 * 
 * Antigravity 챗 입력창 = contenteditable div (aria-label="Message input", role="textbox")
 * DOM value 조작이 아닌 CDP Input.insertText / Input.dispatchKeyEvent 사용
 */
import http from 'http';
import WebSocket from 'ws';

const TARGET_BOT = process.argv[3] || "bot1";
const PROMPT = process.argv[2] || "기본 하트비트 주입(Todo 확인 바랍니다)";
const CDP_PORT = 9000;

function getJson(url) {
    return new Promise((resolve, reject) => {
        http.get(url, res => { let d = ''; res.on('data', c => d+=c); res.on('end', ()=>resolve(JSON.parse(d))); }).on('error', reject);
    });
}

async function injectPulse() {
    console.log(`[PULSE] 💉 대상: ${TARGET_BOT} | 주입: ${PROMPT.slice(0,80)}`);
    try {
        const list = await getJson(`http://127.0.0.1:${CDP_PORT}/json/list`);
        // bot1 탭: title에 "bot1" 포함된 workbench만 타겟
        const target = list.find(t => t.url?.includes('workbench') && t.title?.toLowerCase().includes(TARGET_BOT));
        if (!target) { console.error(`[PULSE] ❌ ${TARGET_BOT} workbench 타겟 없음. 발견된 탭: ${list.filter(t => t.url?.includes('workbench')).map(t => t.title).join(', ')}`); return; }

        const ws = new WebSocket(target.webSocketDebuggerUrl);
        await new Promise((resolve, reject) => { ws.on('open', resolve); ws.on('error', reject); });

        let idCounter = 1;
        const pending = new Map();
        ws.on('message', msg => {
            const data = JSON.parse(msg);
            if (data.id && pending.has(data.id)) { pending.get(data.id)(data); pending.delete(data.id); }
        });

        const call = (method, params) => new Promise(resolve => {
            const id = idCounter++;
            pending.set(id, resolve);
            ws.send(JSON.stringify({ id, method, params }));
        });

        await call('Runtime.enable', {});
        await new Promise(r => setTimeout(r, 300));

        // 1단계: auto-accept 동일 패턴으로 챗 입력창(contenteditable div) 포커스
        const focusScript = `(() => {
            function collectAll(root) {
                const found = [];
                const walk = node => {
                    if (!node) return;
                    if (node.shadowRoot) walk(node.shadowRoot);
                    const children = node.children || [];
                    for (let i = 0; i < children.length; i++) {
                        if (children[i].nodeType === 1) { found.push(children[i]); walk(children[i]); }
                    }
                };
                walk(root);
                return found;
            }
            const allEls = collectAll(document.body);
            
            // aria-label="Message input" + role="textbox" + contenteditable
            let chatInput = allEls.find(el => 
                (el.getAttribute('aria-label')||'').toLowerCase() === 'message input' && 
                el.getAttribute('role') === 'textbox'
            );
            
            if (!chatInput) {
                // 폴백: contenteditable div with textbox role
                chatInput = allEls.find(el => 
                    el.getAttribute('contenteditable') === 'true' && 
                    el.getAttribute('role') === 'textbox' &&
                    el.offsetParent !== null
                );
            }
            
            if (!chatInput) return { ok: false };
            
            // 기존 텍스트 클리어
            chatInput.textContent = '';
            chatInput.focus();
            
            // forceClick (auto-accept 패턴 동일)
            const opts = { view: window, bubbles: true, cancelable: true };
            try { chatInput.dispatchEvent(new PointerEvent('pointerdown', { ...opts, pointerId: 1 })); } catch {}
            try { chatInput.dispatchEvent(new MouseEvent('mousedown', opts)); } catch {}
            try { chatInput.dispatchEvent(new MouseEvent('mouseup', opts)); } catch {}
            try { chatInput.click(); } catch {}
            try { chatInput.dispatchEvent(new PointerEvent('pointerup', { ...opts, pointerId: 1 })); } catch {}
            
            return { ok: true, tag: chatInput.tagName, aria: chatInput.getAttribute('aria-label') };
        })()`;

        const focusRes = await call('Runtime.evaluate', { expression: focusScript, returnByValue: true });
        const focusVal = focusRes?.result?.result?.value;
        console.log(`[PULSE] 포커스 결과: ${JSON.stringify(focusVal)}`);
        
        if (!focusVal?.ok) {
            console.error('[PULSE] ❌ 챗 입력창 포커스 실패');
            ws.close();
            return;
        }

        await new Promise(r => setTimeout(r, 300));

        // 2단계: CDP Input.insertText로 텍스트 삽입 (OS 레벨 — React state 확실 갱신)
        await call('Input.insertText', { text: PROMPT });
        console.log(`[PULSE] ✅ Input.insertText 완료`);

        await new Promise(r => setTimeout(r, 300));

        // 3단계: Enter 키 전송 (CDP Input.dispatchKeyEvent)
        await call('Input.dispatchKeyEvent', {
            type: 'keyDown',
            key: 'Enter',
            code: 'Enter',
            windowsVirtualKeyCode: 13,
            nativeVirtualKeyCode: 13
        });
        await call('Input.dispatchKeyEvent', {
            type: 'keyUp',
            key: 'Enter',
            code: 'Enter',
            windowsVirtualKeyCode: 13,
            nativeVirtualKeyCode: 13
        });

        console.log(`[PULSE] ✅ Enter 키 전송 완료 — 주입 성공`);
        ws.close();
    } catch (e) {
        console.error(`[PULSE] ❌ 에러: ${e.message}`);
    }
}
injectPulse();
