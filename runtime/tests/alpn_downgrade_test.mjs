/**
 * ALPN Downgrade Verification Script
 * 
 * 실제로 Google AI 백엔드가 HTTP/1.1 폴백을 지원하는지 테스트한다.
 * 
 * 테스트 시나리오:
 * 1. cloudcode-pa.googleapis.com에 TLS 연결 (ALPN: h2) → 정상 응답 확인
 * 2. cloudcode-pa.googleapis.com에 TLS 연결 (ALPN: http/1.1만) → 폴백 여부 확인
 * 3. generativelanguage.googleapis.com 동일 테스트
 */

import https from 'https';
import tls from 'tls';
import http2 from 'http2';

const TARGETS = [
    'cloudcode-pa.googleapis.com',
    'generativelanguage.googleapis.com',
    'gemini.googleapis.com'
];

async function testALPN(host, protocols) {
    return new Promise((resolve) => {
        const socket = tls.connect({
            host,
            port: 443,
            ALPNProtocols: protocols,
            servername: host,
            timeout: 5000
        });

        socket.on('secureConnect', () => {
            const negotiated = socket.alpnProtocol;
            const cipher = socket.getCipher();
            resolve({
                host,
                requested: protocols,
                negotiated,
                cipher: cipher?.name || 'unknown',
                success: true
            });
            socket.destroy();
        });

        socket.on('error', (err) => {
            resolve({
                host,
                requested: protocols,
                negotiated: null,
                error: err.message,
                success: false
            });
        });

        socket.on('timeout', () => {
            resolve({
                host,
                requested: protocols,
                negotiated: null,
                error: 'timeout',
                success: false
            });
            socket.destroy();
        });
    });
}

async function testHTTP11Fallback(host) {
    return new Promise((resolve) => {
        const req = https.request({
            hostname: host,
            port: 443,
            path: '/',
            method: 'GET',
            ALPNProtocols: ['http/1.1'],
            timeout: 5000,
            headers: {
                'User-Agent': 'NeuronFS-ALPN-Test/1.0'
            }
        }, (res) => {
            resolve({
                host,
                statusCode: res.statusCode,
                httpVersion: res.httpVersion,
                headers: Object.fromEntries(
                    Object.entries(res.headers).filter(([k]) => 
                        ['content-type', 'server', 'alt-svc', 'x-frame-options'].includes(k)
                    )
                ),
                success: true
            });
            res.resume();
        });

        req.on('error', (err) => {
            resolve({
                host,
                error: err.message,
                success: false
            });
        });

        req.on('timeout', () => {
            resolve({ host, error: 'timeout', success: false });
            req.destroy();
        });

        req.end();
    });
}

async function main() {
    console.log('═══════════════════════════════════════════════');
    console.log('  ALPN Downgrade Verification');
    console.log('═══════════════════════════════════════════════');
    console.log('');

    // Test 1: ALPN 협상 — h2 vs http/1.1
    console.log('── Test 1: ALPN 협상 테스트 ──');
    console.log('');
    
    for (const host of TARGETS) {
        // h2 요청
        const h2Result = await testALPN(host, ['h2', 'http/1.1']);
        console.log(`[${host}]`);
        console.log(`  h2 요청   → 협상 결과: ${h2Result.negotiated || h2Result.error}`);
        
        // http/1.1만 요청
        const h1Result = await testALPN(host, ['http/1.1']);
        console.log(`  h1.1 요청 → 협상 결과: ${h1Result.negotiated || h1Result.error}`);
        
        // http/1.1 강제 시 연결 거부 여부
        if (h1Result.success && h1Result.negotiated === 'http/1.1') {
            console.log(`  ✅ HTTP/1.1 폴백 지원 확인`);
        } else if (h1Result.success && h1Result.negotiated === false) {
            console.log(`  ⚠️ ALPN 협상 실패 (서버가 http/1.1 거부)`);
        } else if (!h1Result.success) {
            console.log(`  ❌ 연결 실패: ${h1Result.error}`);
        }
        console.log('');
    }

    // Test 2: HTTP/1.1 강제 시 실제 HTTP 응답
    console.log('── Test 2: HTTP/1.1 실제 요청 테스트 ──');
    console.log('');
    
    for (const host of TARGETS) {
        const result = await testHTTP11Fallback(host);
        console.log(`[${host}]`);
        if (result.success) {
            console.log(`  HTTP 버전: ${result.httpVersion}`);
            console.log(`  상태 코드: ${result.statusCode}`);
            console.log(`  헤더: ${JSON.stringify(result.headers, null, 2)}`);
        } else {
            console.log(`  ❌ 실패: ${result.error}`);
        }
        console.log('');
    }

    console.log('═══════════════════════════════════════════════');
    console.log('  테스트 완료');
    console.log('═══════════════════════════════════════════════');
}

main().catch(e => console.error('Fatal:', e));
