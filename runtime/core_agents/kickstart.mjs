/**
 * kickstart.mjs — 에이전트 자율 진화 부트스트랩 v3
 * 
 * 핵심: 워크스페이스 격리 + 백로그 + 세션 복기 + 비동기 파이프라인
 * 
 * Usage: node kickstart.mjs [bot1|entp|enfp|all]
 */
import http from 'http';
import WebSocket from 'ws';
import { readdirSync, readFileSync, existsSync, statSync } from 'fs';
import { join } from 'path';

const CDP_PORT = 9000;
const TARGET = process.argv[2] || 'all';
const BRAIN = 'C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\brain_v4';
const AGENTS_DIR = join(BRAIN, '_agents');

function getRecentOutbox(agentId, count = 3) {
    const outbox = join(AGENTS_DIR, agentId, 'outbox');
    if (!existsSync(outbox)) return '(첫 세션 — 이전 기록 없음)';
    const files = readdirSync(outbox)
        .filter(f => f.endsWith('.md'))
        .map(f => ({ name: f, path: join(outbox, f), mtime: statSync(join(outbox, f)).mtimeMs }))
        .sort((a, b) => b.mtime - a.mtime)
        .slice(0, count);
    if (files.length === 0) return '(첫 세션 — 이전 기록 없음)';
    return files.map(f => {
        const content = readFileSync(f.path, 'utf8').slice(0, 1500);
        return `--- ${f.name} ---\n${content}`;
    }).join('\n\n');
}

function getTodoList() {
    const todoDir = join(BRAIN, 'prefrontal', 'todo');
    if (!existsSync(todoDir)) return '(없음)';
    const dirs = readdirSync(todoDir).filter(f => {
        const p = join(todoDir, f);
        return statSync(p).isDirectory() && !existsSync(join(p, 'decay.dormant'));
    });
    return dirs.length > 0 ? dirs.join(', ') : '(모두 완료됨)';
}

function getRoleSSOT(agentId) {
    const pFile = join(AGENTS_DIR, agentId, 'persona.txt');
    if (!existsSync(pFile)) return null;
    const content = readFileSync(pFile, 'utf8');
    
    // Parse the persona text file
    const role = content.match(/ROLE:\s*(.*)/)?.[1] || '';
    const workspace = content.match(/WORKSPACE:\s*(.*)/)?.[1] || '';
    const mission = content.match(/MISSION:\s*(.*)/)?.[1] || '';
    
    const tasksMatch = content.match(/TASKS:\n([\s\S]*?)RULES:/);
    const tasks = tasksMatch ? tasksMatch[1].trim() : '';
    
    const rulesMatch = content.match(/RULES:\n([\s\S]*)/);
    const rules = rulesMatch ? rulesMatch[1].trim() : '';
    
    return { role, workspace, mission, tasks, rules };
}

function buildPrompt(agentId) {
    const r = getRoleSSOT(agentId);
    if (!r) return null;
    const recentWork = getRecentOutbox(agentId);
    const todos = getTodoList();

    return `[SYSTEM] 너는 NeuronFS 프로젝트의 ${r.role}다.

⛔ 워크스페이스 격리: ${r.workspace} 밖으로 절대 나가지 마라.
워크스페이스: ${r.workspace}
미션: ${r.mission}

## 이전 작업 기록 (outbox에서 복기)
${recentWork}

## 현재 미완료 TODO
${todos}

## 즉시 수행
이전 작업 기록을 확인하라. 이전에 하던 작업이 있으면 이어서 하라.
이전 기록이 없으면 아래부터 시작:
${r.tasks}

## 규칙
${r.rules}

## 백로그
backlog.md를 확인하라: _agents/${agentId}/backlog.md
작업 완료하면 체크 표시. 새 작업이 떠오르면 추가.
기다리는 시간에 쉬지 말고 backlog 다음 항목을 실행하라.

## 에이전트 간 통신
다른 에이전트에게 요청하려면 brain_v4/_agents/{대상}/inbox/에 파일 생성.
파일명: {YYYYMMDD}_{HHMMSS}_{from}_{subject}.md
bridge가 3초마다 감시하여 자동 CDP 주입한다.
응답을 기다리는 동안 쉬지 마라 — backlog 다음 항목을 실행하라.

## 자율 진화 루프
1. inbox 확인 → 요청 처리
2. outbox/에 결과 기록 (다음 세션 복기용)
3. backlog.md에서 다음 항목 선택
4. 실행 → 2번으로
5. backlog 비면 스스로 채워라

절대 멈추지 마라. 절대 "다음에 하겠습니다"라고 말하지 마라. 지금 하라.`;
}

function getJson(url) {
    return new Promise((resolve, reject) => {
        http.get(url, res => {
            let d = ''; res.on('data', c => d += c);
            res.on('end', () => resolve(JSON.parse(d)));
        }).on('error', reject);
    });
}

async function injectToTarget(targetTitle, prompt) {
    const list = await getJson(`http://127.0.0.1:${CDP_PORT}/json/list`);
    const target = list.find(t =>
        t.url?.includes('workbench.html') &&
        t.title?.toLowerCase().startsWith(targetTitle.toLowerCase())
    );
    if (!target) { console.log(`[KICK] ❌ ${targetTitle} not found`); return false; }

    const ws = new WebSocket(target.webSocketDebuggerUrl);
    await new Promise((resolve, reject) => { ws.on('open', resolve); ws.on('error', reject); });

    let idCounter = 1;
    const pending = new Map();
    ws.on('message', msg => {
        const data = JSON.parse(msg);
        if (data.id && pending.has(data.id)) { pending.get(data.id)(data); pending.delete(data.id); }
    });
    const call = (method, params) => new Promise((resolve, reject) => {
        const id = idCounter++;
        const t = setTimeout(() => { pending.delete(id); reject(new Error('timeout')); }, 10000);
        pending.set(id, r => { clearTimeout(t); resolve(r); });
        ws.send(JSON.stringify({ id, method, params }));
    });

    await call('Runtime.enable', {});
    await new Promise(r => setTimeout(r, 500));

    const focusScript = `(() => {
        function collectAll(root) {
            const found = [];
            const walk = n => { if(!n)return; if(n.shadowRoot)walk(n.shadowRoot); for(const c of(n.children||[]))if(c.nodeType===1){found.push(c);walk(c);} };
            walk(root); return found;
        }
        const allEls = collectAll(document.body);
        let ci = allEls.find(el => (el.getAttribute('aria-label')||'').toLowerCase()==='message input' && el.getAttribute('role')==='textbox');
        if(!ci) ci = allEls.find(el => el.getAttribute('contenteditable')==='true' && el.getAttribute('role')==='textbox' && el.offsetParent!==null);
        if(!ci) return {ok:false};
        ci.textContent=''; ci.focus();
        const o={view:window,bubbles:true,cancelable:true};
        try{ci.dispatchEvent(new PointerEvent('pointerdown',{...o,pointerId:1}))}catch{}
        try{ci.dispatchEvent(new MouseEvent('mousedown',o))}catch{}
        try{ci.dispatchEvent(new MouseEvent('mouseup',o))}catch{}
        try{ci.click()}catch{}
        return {ok:true};
    })()`;

    const r = await call('Runtime.evaluate', { expression: focusScript, returnByValue: true });
    if (!r?.result?.result?.value?.ok) { console.log(`[KICK] ❌ ${targetTitle} focus failed`); ws.close(); return false; }

    await new Promise(r => setTimeout(r, 300));
    await call('Input.insertText', { text: prompt });
    await new Promise(r => setTimeout(r, 300));
    await call('Input.dispatchKeyEvent', { type:'keyDown', key:'Enter', code:'Enter', windowsVirtualKeyCode:13, nativeVirtualKeyCode:13 });
    await call('Input.dispatchKeyEvent', { type:'keyUp', key:'Enter', code:'Enter', windowsVirtualKeyCode:13, nativeVirtualKeyCode:13 });

    console.log(`[KICK] ✅ ${targetTitle} injected`);
    ws.close();
    return true;
}

async function main() {
    console.log('=== NeuronFS Agent Kickstart v3 ===');
    const targets = TARGET === 'all' ? ['bot1', 'entp', 'enfp'] : [TARGET];

    for (const name of targets) {
        const prompt = buildPrompt(name);
        if (!prompt) { console.log(`[KICK] ❌ unknown: ${name}`); continue; }
        console.log(`\n[KICK] 🚀 ${name}...`);
        try { await injectToTarget(name, prompt); }
        catch (e) { console.log(`[KICK] ❌ ${name}: ${e.message}`); }
        if (targets.length > 1) {
            console.log(`[KICK] ⏳ 10s cooldown...`);
            await new Promise(r => setTimeout(r, 10000));
        }
    }
    console.log('\n=== Done ===');
}

main().catch(e => { console.error(e.message); process.exit(1); });
