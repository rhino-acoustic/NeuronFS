/**
 * v18b — 모든 CDP 타겟(page+worker)에 Network.enable
 * 채팅 gRPC-Web 요청이 어디서 발생하는지 최종 탐색
 */
import WebSocket from 'ws';
import http from 'http';
import fs from 'fs';
import path from 'path';

const DUMP_DIR = 'C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\brain_v4\\_agents\\global_inbox\\cdp_captures';

async function getTargets() {
    return new Promise((resolve, reject) => {
        http.get('http://127.0.0.1:9000/json', (res) => {
            let data = '';
            res.on('data', c => data += c);
            res.on('end', () => resolve(JSON.parse(data)));
        }).on('error', reject);
    });
}

function monitorTarget(wsUrl, label) {
    const ws = new WebSocket(wsUrl);
    let id = 0;
    
    ws.on('open', () => {
        console.log(`[${label}] connected`);
        ws.send(JSON.stringify({ id: ++id, method: 'Network.enable', params: {} }));
    });
    
    ws.on('message', (d) => {
        const msg = JSON.parse(d.toString());
        
        if (msg.method === 'Network.requestWillBeSent') {
            const r = msg.params;
            const url = r.request ? r.request.url : '';
            
            // cascade/chat 키워드 또는 비 heartbeat gRPC만 출력
            const isInteresting = url.includes('cascade') || url.includes('Cascade') ||
                                  url.includes('chat') || url.includes('Chat') ||
                                  url.includes('SendUser') || url.includes('StartCascade') ||
                                  url.includes('StreamAgent') || url.includes('GetChat') ||
                                  (url.includes('14905') && !url.includes('Heartbeat') && !url.includes('GetUnleash'));
            
            if (isInteresting) {
                console.log(`\n🎯🎯🎯 [${label}] ${r.request.method} ${url}`);
                if (r.request.postData) {
                    console.log(`   POST(${r.request.postData.length}B): ${r.request.postData.substring(0, 300)}`);
                    
                    const ts = Date.now();
                    const fname = `chat_net_${ts}`;
                    fs.writeFileSync(path.join(DUMP_DIR, fname + '.bin'), r.request.postData);
                    fs.writeFileSync(path.join(DUMP_DIR, fname + '.bin.meta.json'), JSON.stringify({
                        ts: new Date(ts).toISOString(),
                        label: 'CHAT_NET:' + url,
                        size: r.request.postData.length,
                        source: 'v18_cdp_' + label
                    }));
                }
            } else if (url.includes('14905')) {
                // 다른 gRPC 요청 — 간결하게
                const rpcName = url.split('/').pop();
                process.stdout.write(`· [${label}] ${rpcName} `);
            }
        }
    });
    
    ws.on('error', (e) => {
        console.log(`[${label}] err: ${e.message}`);
    });
}

async function main() {
    console.log('=== v18b — ALL targets Network monitor ===\n');
    
    const targets = await getTargets();
    console.log(`${targets.length} targets\n`);
    
    for (let i = 0; i < targets.length; i++) {
        const t = targets[i];
        if (t.webSocketDebuggerUrl) {
            const label = `${i}:${t.type}:${(t.title || 'worker').substring(0, 15)}`;
            monitorTarget(t.webSocketDebuggerUrl, label);
        }
    }
    
    console.log('\n🎯 대기 중 — 채팅을 보내면 CHAT_NET으로 캡처됩니다\n');
}

main().catch(e => { console.error(e); process.exit(1); });
