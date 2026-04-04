/**
 * NeuronFS v14 — TCP Proxy Interceptor
 * 
 * Language Server(PID 43908)의 gRPC 포트 14912 트래픽을 캡처하기 위해
 * 나가는 연결을 가로채는 대신, PID 50980 → 14912 연결의 
 * 데이터를 tcpdump/npcap 없이 캡처하는 방법:
 * 
 * Windows raw socket으로 loopback 패킷 캡처가 불가하므로,
 * 대안: Electron DevTools Protocol (CDP)를 이용해
 * 9000 포트의 렌더러 페이지에서 채팅 패널의 메시지를 직접 추출
 */
import WebSocket from 'ws';
import fs from 'fs';
import path from 'path';
import http from 'http';

const DUMP_DIR = 'C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\brain_v4\\_agents\\global_inbox\\cdp_captures';

// 1단계: 포트 9000의 모든 DevTools 대상 확인
async function getTargets() {
    return new Promise((resolve, reject) => {
        http.get('http://127.0.0.1:9000/json', (res) => {
            let data = '';
            res.on('data', c => data += c);
            res.on('end', () => {
                try { resolve(JSON.parse(data)); } catch { reject(new Error('parse failed')); }
            });
        }).on('error', reject);
    });
}

async function evaluateInPage(wsUrl, expr) {
    return new Promise((resolve, reject) => {
        const ws = new WebSocket(wsUrl);
        let id = 0;
        ws.on('open', () => {
            id++;
            ws.send(JSON.stringify({
                id, method: 'Runtime.evaluate',
                params: { expression: expr, returnByValue: true }
            }));
        });
        ws.on('message', (d) => {
            const r = JSON.parse(d.toString());
            if (r.id === id) {
                ws.close();
                resolve(r.result);
            }
        });
        ws.on('error', (e) => { reject(e); });
        setTimeout(() => { ws.close(); reject(new Error('timeout')); }, 8000);
    });
}

async function main() {
    console.log('=== v14 — Electron 렌더러 채팅 추출 ===\n');
    
    const targets = await getTargets();
    console.log(`DevTools targets: ${targets.length}\n`);
    
    for (const t of targets) {
        console.log(`  [${t.type}] "${t.title}" → ${t.webSocketDebuggerUrl ? 'WS OK' : 'NO WS'}`);
    }
    
    // 채팅 패널이 포함된 렌더러 찾기
    const chatTargets = targets.filter(t => 
        t.type === 'page' && t.webSocketDebuggerUrl
    );
    
    console.log(`\nPage targets to probe: ${chatTargets.length}\n`);
    
    for (const t of chatTargets) {
        console.log(`--- Probing "${t.title}" ---`);
        try {
            // 1. DOM에서 채팅 관련 요소 확인
            const chatCheck = await evaluateInPage(t.webSocketDebuggerUrl, `
                (function() {
                    // VSCode API 확인
                    const results = [];
                    
                    // acquireVsCodeApi 확인
                    if (typeof acquireVsCodeApi !== 'undefined') results.push('acquireVsCodeApi exists');
                    
                    // window.__vscode 확인
                    if (window.__vscode) results.push('__vscode exists');
                    
                    // iframes 확인
                    results.push('iframes: ' + document.querySelectorAll('iframe').length);
                    
                    // webview-* 요소 확인
                    results.push('webviews: ' + document.querySelectorAll('webview, [id*=webview]').length);
                    
                    // cascade/chat 관련 요소
                    const allText = document.body ? document.body.innerText.substring(0, 500) : '';
                    if (allText.includes('cascade') || allText.includes('Cascade') || allText.includes('Chat'))
                        results.push('has_chat_text');
                    
                    // globalThis의 키 중 관련 것
                    const globals = Object.keys(globalThis).filter(k => 
                        k.toLowerCase().includes('cascade') || 
                        k.toLowerCase().includes('chat') ||
                        k.toLowerCase().includes('vscode') ||
                        k.toLowerCase().includes('grpc')
                    );
                    if (globals.length > 0) results.push('globals: ' + globals.join(','));
                    
                    return results.join(' | ');
                })()
            `);
            console.log(`  Result: ${chatCheck.value || chatCheck.description || JSON.stringify(chatCheck)}`);
        } catch(e) {
            console.log(`  Error: ${e.message}`);
        }
    }
}

main().catch(e => { console.error('Error:', e.message); process.exit(1); });
