// Simple ALPN E2E test — no Korean characters to avoid encoding issues
const https = require('https');

function testEndpoint(host, urlPath, alpn) {
    return new Promise((resolve) => {
        const req = https.request({
            hostname: host,
            port: 443,
            path: urlPath,
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
                    httpVer: res.httpVersion,
                    status: res.statusCode,
                    ct: (res.headers['content-type'] || '?').substring(0, 50),
                    size: raw.length,
                    isJSON,
                    hasBinary: raw.some(b => b === 0x00),
                    preview: text.replace(/[\n\r]/g, ' ').substring(0, 200)
                });
            });
        });
        req.on('error', e => resolve({ error: e.message }));
        req.on('timeout', () => { resolve({ error: 'timeout' }); req.destroy(); });
        req.end();
    });
}

async function main() {
    const tests = [
        { host: 'generativelanguage.googleapis.com', path: '/v1beta/models', label: 'Gemini' },
        { host: 'cloudcode-pa.googleapis.com', path: '/', label: 'CloudCode root' },
        { host: 'cloudcode-pa.googleapis.com', path: '/v1', label: 'CloudCode v1' }
    ];

    console.log('');
    console.log('=== ALPN Downgrade E2E Results ===');
    console.log('');

    for (const t of tests) {
        console.log(`--- ${t.label}: ${t.host}${t.path} ---`);

        const h1 = await testEndpoint(t.host, t.path, ['http/1.1']);
        if (h1.error) {
            console.log(`  [http/1.1] ERROR: ${h1.error}`);
        } else {
            console.log(`  [http/1.1] HTTP/${h1.httpVer} ${h1.status} | CT: ${h1.ct} | Size: ${h1.size}B`);
            console.log(`  [http/1.1] JSON: ${h1.isJSON} | Binary: ${h1.hasBinary}`);
            console.log(`  [http/1.1] Preview: ${h1.preview.substring(0, 150)}`);
        }
        console.log('');
    }

    // MITM test
    console.log('--- MITM Proxy Test ---');
    const fs = require('fs');
    const pfxPath = 'C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\mitm.pfx';
    if (!fs.existsSync(pfxPath)) {
        console.log('  mitm.pfx not found, skipping');
        return;
    }

    const target = { host: 'generativelanguage.googleapis.com', path: '/v1beta/models' };
    
    await new Promise((resolve) => {
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
                    const text = raw.toString('utf-8');
                    let isJSON = false;
                    try { JSON.parse(text); isJSON = true; } catch {}
                    console.log(`  MITM captured: HTTP/${upRes.httpVersion} ${upRes.statusCode}`);
                    console.log(`  MITM CT: ${(upRes.headers['content-type'] || '?').substring(0, 50)}`);
                    console.log(`  MITM JSON: ${isJSON} | Size: ${raw.length}B`);
                    console.log(`  MITM Preview: ${text.replace(/[\n\r]/g, ' ').substring(0, 150)}`);
                    res.writeHead(upRes.statusCode, upRes.headers);
                    res.end(raw);
                });
            });
            fwd.on('error', e => { console.log(`  MITM fwd error: ${e.message}`); res.writeHead(502); res.end(); });
            fwd.end();
        });

        mitm.listen(0, '127.0.0.1', () => {
            const port = mitm.address().port;
            console.log(`  MITM listening on port ${port}`);
            
            const creq = https.request({
                hostname: '127.0.0.1', port, path: target.path, method: 'GET',
                headers: { 'Host': target.host, 'Accept': 'application/json' },
                rejectUnauthorized: false
            }, (res) => {
                res.resume();
                res.on('end', () => { mitm.close(); resolve(); });
            });
            creq.on('error', e => { console.log(`  Client error: ${e.message}`); mitm.close(); resolve(); });
            creq.end();
        });
    });

    console.log('');
    console.log('=== Done ===');
}

main().catch(e => console.error('Fatal:', e));
