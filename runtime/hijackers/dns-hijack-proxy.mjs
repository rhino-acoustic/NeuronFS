/**
 * DNS Hijack Proxy — Layer 5: Transparent HTTP/2 Interception
 * 
 * How it works:
 *   1. hosts file maps cloudcode-pa.googleapis.com → 127.0.0.1
 *   2. This server listens on 127.0.0.1:443 as an HTTP/2 TLS server
 *   3. Node.js http2.connect() follows DNS → arrives here
 *   4. We capture the plaintext request/response, then forward to the real server
 *   5. The real server IP is resolved via DNS-over-HTTPS (bypassing hosts file)
 * 
 * This bypasses the HTTPS_PROXY limitation of Node.js http2 module.
 */

import http2 from 'http2';
import https from 'https';
import http from 'http';
import fs from 'fs';
import path from 'path';
import { execSync } from 'child_process';
import dns from 'dns';

const TARGET_HOST = 'cloudcode-pa.googleapis.com';
const LISTEN_PORT = 443;
const LISTEN_HOST = '127.0.0.1';

const BRAIN = 'C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\brain_v4';
const DUMP_DIR = path.join(BRAIN, '_transcripts', '_dns_hijack_dumps');
const TRANSCRIPT_DIR = path.join(BRAIN, '_transcripts');
fs.mkdirSync(DUMP_DIR, { recursive: true });

// --- Real IP resolution (bypass hosts file) ---
let REAL_IP = null;

async function resolveRealIP() {
    // Use DNS-over-HTTPS to bypass local hosts file
    return new Promise((resolve, reject) => {
        const dohUrl = `https://dns.google/resolve?name=${TARGET_HOST}&type=A`;
        https.get(dohUrl, { headers: { 'Accept': 'application/dns-json' } }, (res) => {
            let data = '';
            res.on('data', c => data += c);
            res.on('end', () => {
                try {
                    const json = JSON.parse(data);
                    const ip = json.Answer?.find(a => a.type === 1)?.data;
                    if (ip) {
                        resolve(ip);
                    } else {
                        reject(new Error('No A record found'));
                    }
                } catch (e) {
                    reject(e);
                }
            });
        }).on('error', reject);
    });
}

// --- Cert generation ---
const pfxPath = path.join(process.cwd(), 'mitm.pfx');
const passphrase = '1234';

if (!fs.existsSync(pfxPath)) {
    const cmd = `powershell -NoProfile -Command "$cert = New-SelfSignedCertificate -DnsName '${TARGET_HOST}','*.googleapis.com' -CertStoreLocation 'cert:\\CurrentUser\\My' -FriendlyName 'NeuronFSDNS'; $pwd = ConvertTo-SecureString -String '${passphrase}' -Force -AsPlainText; Export-PfxCertificate -Cert $cert -FilePath '${pfxPath}' -Password $pwd"`;
    execSync(cmd, { stdio: 'inherit' });
}

// --- Capture logic ---
let captureCount = 0;

function logCapture(direction, streamId, headers, data) {
    captureCount++;
    const ts = new Date().toISOString();
    const method = headers?.[':method'] || '?';
    const urlPath = headers?.[':path'] || '?';
    const ct = headers?.['content-type'] || '?';
    
    const entry = {
        ts, direction, streamId, method, path: urlPath,
        contentType: ct,
        size: data?.length || 0,
        preview: ''
    };

    if (data && data.length > 0) {
        // Try to decode as text
        const text = data.toString('utf-8');
        
        // gRPC frames: 5-byte header (compressed flag + 4-byte length) + protobuf
        if (ct.includes('grpc')) {
            // Strip gRPC frame header (5 bytes) to get raw protobuf
            if (data.length > 5) {
                const compressed = data[0];
                const msgLen = data.readUInt32BE(1);
                const payload = data.slice(5, 5 + msgLen);
                
                // Try JSON (some gRPC-web uses JSON encoding)
                try {
                    entry.preview = JSON.parse(payload.toString('utf-8'));
                    entry.format = 'grpc-json';
                } catch {
                    // Binary protobuf — dump raw for later analysis
                    entry.format = 'grpc-protobuf';
                    entry.preview = payload.toString('base64').substring(0, 500);
                }
            }
        } else {
            // Regular HTTP — try JSON parse
            try {
                entry.preview = JSON.parse(text);
                entry.format = 'json';
            } catch {
                entry.preview = text.substring(0, 2000);
                entry.format = 'text';
            }
        }
    }
    
    // Write to dump file
    const dumpFile = path.join(DUMP_DIR, `capture_${Date.now()}_${captureCount}.json`);
    fs.writeFileSync(dumpFile, JSON.stringify(entry, null, 2), 'utf-8');

    // Append to hourly transcript
    const now = new Date();
    const kstHour = new Date(now.getTime() + 9 * 3600000);
    const hourFile = path.join(TRANSCRIPT_DIR, `dns_hijack_${kstHour.toISOString().slice(0,13)}h.jsonl`);
    fs.appendFileSync(hourFile, JSON.stringify(entry) + '\n', 'utf-8');
    
    const dir = direction === 'REQ' ? '→' : '←';
    const preview = typeof entry.preview === 'string' 
        ? entry.preview.substring(0, 100) 
        : JSON.stringify(entry.preview).substring(0, 100);
    process.stdout.write(`[${dir}] ${method} ${urlPath} | ${ct.substring(0,30)} | ${data?.length || 0}B | ${preview}\n`);
}

// --- Main server ---
async function start() {
    try {
        REAL_IP = await resolveRealIP();
    } catch (e) {
        // Fallback: use system DNS (won't work if hosts file is already set)
        // Try hardcoded Google IP range
        console.error('[WARN] DoH failed:', e.message);
        console.error('[WARN] Trying direct DNS...');
        REAL_IP = await new Promise((resolve, reject) => {
            dns.resolve4(TARGET_HOST, (err, addrs) => {
                if (err || !addrs?.length) reject(err || new Error('no DNS'));
                else resolve(addrs[0]);
            });
        }).catch(() => {
            console.error('[FATAL] Cannot resolve real IP. Run without hosts file entry first.');
            process.exit(1);
        });
    }
    
    console.log(`[DNS-HIJACK] Real IP for ${TARGET_HOST}: ${REAL_IP}`);

    const server = http2.createSecureServer({
        pfx: fs.readFileSync(pfxPath),
        passphrase,
        allowHTTP1: true,  // Accept both HTTP/1.1 and HTTP/2
    });

    server.on('stream', (clientStream, headers) => {
        const method = headers[':method'];
        const urlPath = headers[':path'];
        const authority = headers[':authority'] || TARGET_HOST;

        // Collect request body
        const reqChunks = [];
        clientStream.on('data', c => reqChunks.push(c));
        clientStream.on('end', () => {
            const reqBody = Buffer.concat(reqChunks);
            logCapture('REQ', clientStream.id, headers, reqBody);

            // Forward to real server
            const upstreamSession = http2.connect(`https://${REAL_IP}`, {
                rejectUnauthorized: false,
                servername: TARGET_HOST,
                createConnection: (url, opts) => {
                    const tls = require('tls');
                    return tls.connect({
                        host: REAL_IP,
                        port: 443,
                        servername: TARGET_HOST,
                        ALPNProtocols: ['h2'],
                        ...opts
                    });
                }
            });

            const fwdHeaders = { ...headers, ':authority': TARGET_HOST };
            const upstreamStream = upstreamSession.request(fwdHeaders);

            if (reqBody.length > 0) {
                upstreamStream.write(reqBody);
            }
            upstreamStream.end();

            // Collect response
            const resChunks = [];
            upstreamStream.on('response', (resHeaders) => {
                // Forward response headers to client
                const cleanHeaders = {};
                for (const [k, v] of Object.entries(resHeaders)) {
                    if (!k.startsWith(':')) cleanHeaders[k] = v;
                }
                clientStream.respond({
                    ':status': resHeaders[':status'],
                    ...cleanHeaders
                });
            });

            upstreamStream.on('data', c => {
                resChunks.push(c);
                clientStream.write(c);
            });

            upstreamStream.on('end', () => {
                const resBody = Buffer.concat(resChunks);
                logCapture('RES', clientStream.id, { ':method': method, ':path': urlPath, 'content-type': upstreamStream.sentHeaders?.['content-type'] || '?' }, resBody);
                clientStream.end();
                upstreamSession.close();
            });

            upstreamStream.on('error', (e) => {
                console.error(`[FWD ERR] ${urlPath}: ${e.message}`);
                try { clientStream.respond({ ':status': 502 }); } catch {}
                clientStream.end();
                upstreamSession.close();
            });

            upstreamSession.on('error', (e) => {
                console.error(`[SESSION ERR] ${e.message}`);
            });
        });
    });

    server.on('error', (e) => {
        if (e.code === 'EACCES') {
            console.error(`[FATAL] Port ${LISTEN_PORT} requires admin privileges. Run as Administrator.`);
        } else if (e.code === 'EADDRINUSE') {
            console.error(`[FATAL] Port ${LISTEN_PORT} already in use.`);
        } else {
            console.error('[SERVER ERR]', e.message);
        }
    });

    server.listen(LISTEN_PORT, LISTEN_HOST, () => {
        console.log(`
╔══════════════════════════════════════════════════╗
║  DNS Hijack Proxy — Layer 5                      ║
║  Listening: ${LISTEN_HOST}:${LISTEN_PORT}                       ║
║  Target: ${TARGET_HOST}          ║
║  Real IP: ${REAL_IP}                              ║
║  Dumps: ${DUMP_DIR.substring(0,42)}... ║
╚══════════════════════════════════════════════════╝

Add to hosts file (as Admin):
  127.0.0.1  ${TARGET_HOST}

Then restart Antigravity with NODE_TLS_REJECT_UNAUTHORIZED=0
`);
    });
}

start().catch(e => {
    console.error('[FATAL]', e);
    process.exit(1);
});

process.on('uncaughtException', (e) => {
    if (e.code !== 'ECONNRESET' && e.code !== 'EPIPE') {
        console.error('[UNCAUGHT]', e.message);
    }
});
