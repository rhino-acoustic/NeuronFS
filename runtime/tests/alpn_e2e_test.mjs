/**
 * ALPN Downgrade E2E — Clean Output Version
 */
import https from 'https';
import fs from 'fs';
import path from 'path';

const TARGETS = [
    { host: 'generativelanguage.googleapis.com', path: '/v1beta/models', desc: 'Gemini REST' },
    { host: 'cloudcode-pa.googleapis.com', path: '/', desc: 'CloudCode' },
    { host: 'cloudcode-pa.googleapis.com', path: '/v1/projects', desc: 'CloudCode v1' }
];

async function test(target, alpn) {
    return new Promise((resolve) => {
        const req = https.request({
            hostname: target.host,
            port: 443,
            path: target.path,
            method: 'GET',
            ALPNProtocols: alpn,
            headers: { 'Accept': 'application/json', 'Content-Type': 'application/json' },
            timeout: 8000
        }, (res) => {
            const chunks = [];
            res.on('data', c => chunks.push(c));
            res.on('end', () => {
                const raw = Buffer.concat(chunks);
                const text = raw.toString('utf-8');
                let isJSON = false;
                try { JSON.parse(text); isJSON = true; } catch {}
                resolve({
                    alpn: alpn.join(','),
                    httpVer: res.httpVersion,
                    status: res.statusCode,
                    contentType: (res.headers['content-type'] || '').substring(0, 40),
                    size: raw.length,
                    isJSON,
                    preview: text.replace(/\n/g, ' ').substring(0, 120)
                });
            });
        });
        req.on('error', e => resolve({ alpn: alpn.join(','), error: e.message }));
        req.on('timeout', () => { resolve({ alpn: alpn.join(','), error: 'timeout' }); req.destroy(); });
        req.end();
    });
}

async function testMITM(target) {
    const pfxPath = 'C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\mitm.pfx';
    if (!fs.existsSync(pfxPath)) return { error: 'mitm.pfx not found' };

    return new Promise((resolve) => {
        let capturedBody = '';
        
        const mitm = https.createServer({
            pfx: fs.readFileSync(pfxPath),
            passphrase: '1234',
            ALPNProtocols: ['http/1.1']
        }, (req, res) => {
            const fwd = https.request({
                hostname: target.host, port: 443, path: req.url, method: req.method,
                headers: { ...req.headers, host: target.host }
            }, (upRes) => {
                const chunks = [];
                upRes.on('data', c => chunks.push(c));
                upRes.on('end', () => {
                    const raw = Buffer.concat(chunks);
                    capturedBody = raw.toString('utf-8');
                    res.writeHead(upRes.statusCode, upRes.headers);
                    res.end(raw);
                });
            });
            fwd.on('error', e => { res.writeHead(502); res.end(e.message); });
            fwd.end();
        });

        mitm.listen(0, '127.0.0.1', () => {
            const port = mitm.address().port;
            const creq = https.request({
                hostname: '127.0.0.1', port, path: target.path, method: 'GET',
                headers: { 'Host': target.host, 'Accept': 'application/json' },
                rejectUnauthorized: false
            }, (res) => {
                const chunks = [];
                res.on('data', c => chunks.push(c));
                res.on('end', () => {
                    mitm.close();
                    let isJSON = false;
                    try { JSON.parse(capturedBody); isJSON = true; } catch {}
                    resolve({
                        status: res.statusCode,
                        capturedSize: capturedBody.length,
                        isJSON,
                        preview: capturedBody.replace(/\n/g, ' ').substring(0, 120)
                    });
                });
            });
            creq.on('error', e => { mitm.close(); resolve({ error: e.message }); });
            creq.setTimeout(10000, () => { mitm.close(); creq.destroy(); resolve({ error: 'timeout' }); });
            creq.end();
        });
    });
}

async function main() {
    const results = [];

    console.log('\n=== ALPN E2E: 직접 요청 ===\n');
    
    for (const t of TARGETS) {
        // h2 요청
        const h2 = await test(t, ['h2', 'http/1.1']);
        // http/1.1 강제
        const h1 = await test(t, ['http/1.1']);
        
        console.log(`[${t.desc}] ${t.host}${t.path}`);
        console.log(`  h2:     HTTP/${h2.httpVer || '?'} ${h2.status || h2.error} CT=${h2.contentType || '?'} JSON=${h2.isJSON} (${h2.size || 0}B)`);
        console.log(`  h1.1:   HTTP/${h1.httpVer || '?'} ${h1.status || h1.error} CT=${h1.contentType || '?'} JSON=${h1.isJSON} (${h1.size || 0}B)`);
        console.log(`  h1 preview: ${h1.preview || h1.error}`);
        console.log('');
    }

    console.log('=== ALPN E2E: MITM 경유 ===\n');
    
    for (const t of TARGETS) {
        const m = await testMITM(t);
        console.log(`[${t.desc}] MITM → ${t.host}${t.path}`);
        if (m.error) {
            console.log(`  ❌ ${m.error}`);
        } else {
            console.log(`  상태: ${m.status} | 캡처: ${m.capturedSize}B | JSON: ${m.isJSON ? '✅ 평문!' : '❌ 아님'}`);
            console.log(`  캡처 미리보기: ${m.preview}`);
        }
        console.log('');
    }
}

main().catch(e => console.error('Fatal:', e));
