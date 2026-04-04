/**
 * cdp-core.mjs — CDP 공용 모듈
 * 
 * 13개 파일에 흩어져있던 getJson/WS접속/DOM포커스/텍스트주입 로직을
 * 단일 모듈로 통합. 모든 CDP 스크립트는 이 모듈을 import해서 사용.
 * 
 * Usage:
 *   import { getJson, connectCDP, focusChat, injectText } from './cdp-core.mjs';
 */
import http from 'http';
import WebSocket from 'ws';

const CDP_PORT = 9000;

/**
 * HTTP GET → JSON 파싱
 */
export function getJson(url) {
    return new Promise((resolve, reject) => {
        http.get(url, res => {
            let d = '';
            res.on('data', c => d += c);
            res.on('end', () => {
                try { resolve(JSON.parse(d)); }
                catch (e) { reject(e); }
            });
        }).on('error', reject);
    });
}

/**
 * CDP 타겟 목록 조회
 */
export async function listTargets(port) {
    const p = port || CDP_PORT;
    return getJson(`http://127.0.0.1:${p}/json/list`);
}

/**
 * 특정 에이전트의 workbench 타겟 찾기
 * @param {string} agentName - 'bot1', 'entp', 'enfp' 등
 * @param {number} port - CDP 포트 (기본 9000)
 */
export async function findTarget(agentName, port) {
    const list = await listTargets(port);
    return list.find(t =>
        t.url?.includes('workbench.html') &&
        t.title?.toLowerCase().startsWith(agentName.toLowerCase()) &&
        t.type === 'page'
    );
}

/**
 * WebSocket + RPC 호출 인터페이스 생성
 * @param {string} wsUrl - webSocketDebuggerUrl
 * @returns {{ ws, call, close }}
 */
export async function connectCDP(wsUrl) {
    const ws = new WebSocket(wsUrl);
    await new Promise((resolve, reject) => {
        ws.on('open', resolve);
        ws.on('error', reject);
    });

    let idCounter = 1;
    const pending = new Map();

    ws.on('message', msg => {
        try {
            const data = JSON.parse(msg);
            if (data.id !== undefined && pending.has(data.id)) {
                const { resolve, reject } = pending.get(data.id);
                pending.delete(data.id);
                if (data.error) reject(data.error);
                else resolve(data.result || data);
            }
        } catch {}
    });

    const call = (method, params) => new Promise((resolve, reject) => {
        if (ws.readyState !== WebSocket.OPEN) {
            reject(new Error('ws closed'));
            return;
        }
        const id = idCounter++;
        const t = setTimeout(() => {
            pending.delete(id);
            reject(new Error('timeout'));
        }, 10000);
        pending.set(id, {
            resolve: r => { clearTimeout(t); resolve(r); },
            reject: e => { clearTimeout(t); reject(e); }
        });
        ws.send(JSON.stringify({ id, method, params }));
    });

    await call('Runtime.enable', {});
    await new Promise(r => setTimeout(r, 300));

    return {
        ws,
        call,
        close: () => ws.close()
    };
}

/**
 * 에이전트에 직접 연결 (findTarget + connectCDP)
 */
export async function connectToAgent(agentName, port) {
    const target = await findTarget(agentName, port);
    if (!target) throw new Error(`${agentName} workbench target not found`);
    const conn = await connectCDP(target.webSocketDebuggerUrl);
    return { ...conn, target };
}

/**
 * 챗 입력창 포커스 (Shadow DOM 전체 순회)
 * @returns {boolean} 성공 여부
 */
export async function focusChat(call) {
    const focusScript = `(() => {
        function collectAll(root) {
            const found = [];
            const walk = n => {
                if (!n) return;
                if (n.shadowRoot) walk(n.shadowRoot);
                for (const c of (n.children || [])) {
                    if (c.nodeType === 1) { found.push(c); walk(c); }
                }
            };
            walk(root);
            return found;
        }
        const allEls = collectAll(document.body);
        let ci = allEls.find(el =>
            (el.getAttribute('aria-label') || '').toLowerCase() === 'message input' &&
            el.getAttribute('role') === 'textbox'
        );
        if (!ci) {
            ci = allEls.find(el =>
                el.getAttribute('contenteditable') === 'true' &&
                el.getAttribute('role') === 'textbox' &&
                el.offsetParent !== null
            );
        }
        if (!ci) return { ok: false };
        ci.textContent = '';
        ci.focus();
        const o = { view: window, bubbles: true, cancelable: true };
        try { ci.dispatchEvent(new PointerEvent('pointerdown', { ...o, pointerId: 1 })); } catch {}
        try { ci.dispatchEvent(new MouseEvent('mousedown', o)); } catch {}
        try { ci.dispatchEvent(new MouseEvent('mouseup', o)); } catch {}
        try { ci.click(); } catch {}
        return { ok: true };
    })()`;

    const res = await call('Runtime.evaluate', { expression: focusScript, returnByValue: true });
    return res?.result?.value?.ok || res?.value?.ok || false;
}

/**
 * 텍스트 입력 + Enter 전송 (포커스 포함)
 * @returns {boolean} 성공 여부
 */
export async function injectText(call, text) {
    const focused = await focusChat(call);
    if (!focused) return false;

    await new Promise(r => setTimeout(r, 200));
    await call('Input.insertText', { text });
    await new Promise(r => setTimeout(r, 200));
    await call('Input.dispatchKeyEvent', {
        type: 'keyDown', key: 'Enter', code: 'Enter',
        windowsVirtualKeyCode: 13, nativeVirtualKeyCode: 13
    });
    await call('Input.dispatchKeyEvent', {
        type: 'keyUp', key: 'Enter', code: 'Enter',
        windowsVirtualKeyCode: 13, nativeVirtualKeyCode: 13
    });
    return true;
}

/**
 * 에이전트에 메시지 주입 (원스텝: 연결 → 포커스 → 입력 → Enter → 닫기)
 */
export async function injectToAgent(agentName, message, port) {
    const { call, close } = await connectToAgent(agentName, port);
    try {
        const ok = await injectText(call, message);
        return ok;
    } finally {
        close();
    }
}

/**
 * Shadow DOM 전체 순회 후 요소 수집 스크립트 (Runtime.evaluate용)
 * 여러 곳에서 사용하는 collectAll 패턴의 표준화
 */
export const COLLECT_ALL_SCRIPT = `
function collectAll(root) {
    const found = [];
    const walk = n => {
        if (!n) return;
        if (n.shadowRoot) walk(n.shadowRoot);
        for (const c of (n.children || [])) {
            if (c.nodeType === 1) { found.push(c); walk(c); }
        }
    };
    walk(root);
    return found;
}`;

export default {
    getJson, listTargets, findTarget,
    connectCDP, connectToAgent,
    focusChat, injectText, injectToAgent,
    COLLECT_ALL_SCRIPT
};
