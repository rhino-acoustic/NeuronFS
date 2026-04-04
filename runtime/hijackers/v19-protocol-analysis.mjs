/**
 * v19 — 통신 구조 분석 + Fetch 인젝션 가능성 테스트
 * 
 * CDP Fetch.enable으로 gRPC-Web 요청을 가로채서:
 * 1. 요청 구조 완전 분석 (헤더, 바디, 타이밍)
 * 2. 수정 후 전달 가능성 검증
 */
import WebSocket from 'ws';
import http from 'http';
import fs from 'fs';

const CDP_PORT = 9000;

async function getTargets() {
    return new Promise((resolve, reject) => {
        http.get(`http://127.0.0.1:${CDP_PORT}/json`, (res) => {
            let data = '';
            res.on('data', c => data += c);
            res.on('end', () => resolve(JSON.parse(data)));
        }).on('error', reject);
    });
}

async function analyzeProtocol() {
    const targets = await getTargets();
    const mainPage = targets.find(t => t.type === 'page' && t.title.includes('BASEMENT_ADMIN'));
    if (!mainPage) { console.log('No main page'); process.exit(1); }
    
    console.log('Target:', mainPage.title);
    
    const ws = new WebSocket(mainPage.webSocketDebuggerUrl);
    let id = 0;
    let requestDetails = {};
    
    ws.on('open', () => {
        console.log('Connected\n');
        
        // Network.enable — 요청/응답 전체 관찰
        ws.send(JSON.stringify({ id: ++id, method: 'Network.enable', params: { maxTotalBufferSize: 50000000 } }));
        
        // Fetch.enable — 요청 가로채기 (SendUserCascadeMessage만)
        // handleAuthRequests: true로 하면 인증 요청도 잡힘
        ws.send(JSON.stringify({ id: ++id, method: 'Fetch.enable', params: {
            patterns: [
                { urlPattern: '*LanguageServerService*', requestStage: 'Request' }
            ],
            handleAuthRequests: false
        }}));
    });
    
    ws.on('message', (d) => {
        const msg = JSON.parse(d.toString());
        
        // ===== Network 이벤트: 요청 전체 라이프사이클 =====
        if (msg.method === 'Network.requestWillBeSent') {
            const r = msg.params;
            const url = r.request?.url || '';
            if (!url.includes('LanguageServerService')) return;
            if (url.includes('Heartbeat') || url.includes('GetUnleash') || url.includes('GetStatus')) return;
            
            const rpc = url.split('/').pop();
            const reqId = r.requestId;
            
            requestDetails[reqId] = {
                rpc,
                ts: Date.now(),
                method: r.request.method,
                headers: r.request.headers,
                postDataLen: r.request.postData?.length || 0,
                url
            };
            
            console.log(`\n${'='.repeat(70)}`);
            console.log(`📤 REQUEST: ${rpc} (${reqId})`);
            console.log(`   Method: ${r.request.method}`);
            console.log(`   URL: ${url}`);
            console.log(`   Headers:`);
            for (const [k, v] of Object.entries(r.request.headers || {})) {
                console.log(`     ${k}: ${v.substring(0, 80)}`);
            }
            if (r.request.postData) {
                console.log(`   Body (${r.request.postData.length}B):`);
                console.log(`     ${r.request.postData.substring(0, 300)}`);
            }
            console.log(`   Initiator: ${r.initiator?.type || '?'} ${r.initiator?.url?.substring(0, 80) || ''}`);
        }
        
        if (msg.method === 'Network.responseReceived') {
            const r = msg.params;
            const url = r.response?.url || '';
            if (!url.includes('LanguageServerService')) return;
            if (url.includes('Heartbeat') || url.includes('GetUnleash') || url.includes('GetStatus')) return;
            
            const rpc = url.split('/').pop();
            const reqId = r.requestId;
            const detail = requestDetails[reqId];
            const elapsed = detail ? Date.now() - detail.ts : '?';
            
            console.log(`\n📥 RESPONSE: ${rpc} (${reqId}) — ${elapsed}ms`);
            console.log(`   Status: ${r.response.status}`);
            console.log(`   Headers:`);
            for (const [k, v] of Object.entries(r.response.headers || {})) {
                console.log(`     ${k}: ${v.substring(0, 80)}`);
            }
            
            // 응답 본문 가져오기
            ws.send(JSON.stringify({
                id: ++id,
                method: 'Network.getResponseBody',
                params: { requestId: reqId }
            }));
            
            // 응답 ID 매핑
            requestDetails[`resp_${id}`] = { rpc, reqId };
        }
        
        if (msg.method === 'Network.loadingFinished') {
            const reqId = msg.params.requestId;
            const detail = requestDetails[reqId];
            if (detail) {
                const elapsed = Date.now() - detail.ts;
                console.log(`\n✅ COMPLETE: ${detail.rpc} — total ${elapsed}ms, ${msg.params.encodedDataLength || '?'}B transferred`);
            }
        }
        
        // Network.getResponseBody 응답
        if (msg.id && requestDetails[`resp_${msg.id}`]) {
            const detail = requestDetails[`resp_${msg.id}`];
            const body = msg.result?.body || '';
            const base64 = msg.result?.base64Encoded;
            console.log(`\n📦 RESPONSE BODY (${detail.rpc}):`);
            console.log(`   base64: ${base64}`);
            console.log(`   length: ${body.length}`);
            if (!base64 && body.length < 500) {
                console.log(`   content: ${body}`);
            } else if (!base64) {
                console.log(`   preview: ${body.substring(0, 300)}...`);
            } else {
                const buf = Buffer.from(body, 'base64');
                console.log(`   decoded (${buf.length}B): ${buf.toString('utf8').substring(0, 300)}`);
            }
        }
        
        // ===== Fetch 이벤트: 요청 가로채기 =====
        if (msg.method === 'Fetch.requestPaused') {
            const r = msg.params;
            const url = r.request?.url || '';
            const rpc = url.split('/').pop();
            
            console.log(`\n🔒 FETCH PAUSED: ${rpc}`);
            console.log(`   requestId: ${r.requestId}`);
            console.log(`   networkId: ${r.networkId}`);
            console.log(`   resourceType: ${r.resourceType}`);
            
            if (r.request.postData) {
                console.log(`   postData (${r.request.postData.length}B): ${r.request.postData.substring(0, 200)}`);
                
                // ★ 여기서 postData를 수정하여 continueRequest에 넣으면 인젝션 성공!
                // 예: JSON에 추가 컨텍스트를 삽입
                console.log(`   ⚡ INJECTION POSSIBLE: Fetch.continueRequest with modified postData`);
            }
            
            // 수정 없이 통과 (테스트용)
            ws.send(JSON.stringify({
                id: ++id,
                method: 'Fetch.continueRequest',
                params: { requestId: r.requestId }
            }));
        }
    });
    
    ws.on('error', (e) => console.log('Error:', e.message));
    
    console.log('🎯 분석 대기 중 — 채팅을 보내면 전체 프로토콜이 출력됩니다\n');
}

analyzeProtocol().catch(e => { console.error(e); process.exit(1); });
