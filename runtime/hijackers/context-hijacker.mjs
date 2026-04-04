import http from 'http';
import https from 'https';
import net from 'net';
import fs from 'fs';
import path from 'path';
import { URL } from 'url';
import { execSync } from 'child_process';

const INBOX = 'C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\brain_v4\\_agents\\global_inbox';
const OUTPUT = path.join(INBOX, 'latest_hijacked_context.md');
const DUMP_DIR = path.join(INBOX, 'grpc_dumps');
fs.mkdirSync(DUMP_DIR, { recursive: true });

const TARGET_HOSTS = ['cloudcode-pa.googleapis.com', 'generativelanguage.googleapis.com', 'anthropic.com'];

let captureCount = 0;

function tryExtract(reqBody, reqUrl, reqMethod) {
    try {
        const str = Buffer.isBuffer(reqBody) ? reqBody.toString('utf8') : reqBody;
        if (str.length > 10) {
            captureCount++;
            const dumpFile = path.join(DUMP_DIR, `raw_dump_${Date.now()}_${captureCount}.bin`);
            fs.writeFileSync(dumpFile, reqBody);
            console.log(`[MITM CAPTURE] Dumped ${str.length} bytes from ${reqMethod} ${reqUrl}`);
            
            // Still write the md file to global inbox just in case it is readable text
            fs.writeFileSync(OUTPUT, `# Hijacked Context — ${new Date().toISOString()}\n\nURL: ${reqUrl}\n\n\`\`\`text\n${str.substring(0, 50000)}\n\`\`\`\n`, 'utf8');
        }
    } catch(e) { }
}

// Ensure PFX cert is ready
const pfxPath = path.join(process.cwd(), 'mitm.pfx');
const passphrase = '1234';
if (!fs.existsSync(pfxPath)) {
    console.log('[MITM] Generating self-signed PFX using PowerShell...');
    execSync(`powershell -NoProfile -Command "$cert = New-SelfSignedCertificate -DnsName 'cloudcode-pa.googleapis.com','generativelanguage.googleapis.com','anthropic.com','*.googleapis.com','gemini.googleapis.com' -CertStoreLocation 'cert:\\CurrentUser\\My' -FriendlyName 'NeuronFSMITM'; $pwd = ConvertTo-SecureString -String '${passphrase}' -Force -AsPlainText; Export-PfxCertificate -Cert $cert -FilePath '${pfxPath}' -Password $pwd"`, { stdio: 'inherit' });
    console.log('[MITM] PFX generated!');
} else {
    console.log('[MITM] Loading existing PFX cert...');
}

// 1. Create Local HTTPS Server (TLS Termination with Forced HTTP/1.1 ALPN)
// By forcing HTTP/1.1, Chromium won't use HTTP/2 gRPC ALPN, converting complex protobuf streams to simple POST.
const mitmServer = https.createServer({
    pfx: fs.readFileSync(pfxPath),
    passphrase,
    ALPNProtocols: ['http/1.1']
}, (req, res) => {
    let body = [];
    req.on('data', c => body.push(c));
    req.on('end', () => {
        const fullBody = Buffer.concat(body);
        const urlStr = `https://${req.headers.host}${req.url}`;
        
        if (TARGET_HOSTS.some(t => urlStr.includes(t))) {
            tryExtract(fullBody, urlStr, req.method);
        }

        // Forward Safely (Re-encryption)
        const reqOpts = {
            hostname: req.headers.host,
            port: 443,
            path: req.url,
            method: req.method,
            headers: req.headers,
            rejectUnauthorized: false // we don't care about upstream invalid certs here
        };
        
        const fwd = https.request(reqOpts, (upRes) => {
            res.writeHead(upRes.statusCode, upRes.headers);
            upRes.pipe(res);
        });
        fwd.on('error', e => {
            console.error(`[MITM Forward Error] ${req.url.substring(0, 50)}... -> ${e.message}`);
            res.end();
        });
        if (fullBody.length > 0) fwd.write(fullBody);
        fwd.end();
    });
});

mitmServer.listen(0, '127.0.0.1', () => {
    const mitmPort = mitmServer.address().port;
    console.log(`[MITM Server] TLS termination listening internally on port ${mitmPort}`);
    
    // 2. Main Proxy Tunnel (Port 8080)
    const proxy = http.createServer((req, res) => {
        const urlStr = req.url.startsWith('http') ? req.url : `http://${req.headers.host}${req.url}`;
        const u = new URL(urlStr);
        const reqOpts = { host: u.hostname, port: u.port || 80, path: u.pathname + u.search, method: req.method, headers: req.headers };
        const fwd = http.request(reqOpts, (upRes) => { res.writeHead(upRes.statusCode, upRes.headers); upRes.pipe(res); });
        req.pipe(fwd);
        fwd.on('error', () => res.end());
    });

    // Handle HTTPS CONNECT Tunneling
    proxy.on('connect', (req, clientSocket, head) => {
        const [host, port] = req.url.split(':');
        
        // Always ACK the connect
        clientSocket.write('HTTP/1.1 200 Connection Established\r\n\r\n');
        
        if (TARGET_HOSTS.some(t => host.includes(t))) {
            console.log(`[PROXY] Intercepting: ${host}:${port}`);
            // Forward to our local TLS termination server!
            const mitmSocket = net.connect(mitmPort, '127.0.0.1');
            mitmSocket.write(head);
            mitmSocket.pipe(clientSocket);
            clientSocket.pipe(mitmSocket);
            mitmSocket.on('error', () => clientSocket.end());
        } else {
            // Unmonitored traffic passthrough
            const srvSocket = net.connect(parseInt(port) || 443, host);
            srvSocket.write(head);
            srvSocket.pipe(clientSocket);
            clientSocket.pipe(srvSocket);
            srvSocket.on('error', () => clientSocket.end());
        }
        clientSocket.on('error', () => {});
    });

    proxy.listen(8080, '127.0.0.1', () => {
        console.log('\n======================================================');
        console.log('📡 NeuronFS Context Hijacker (TLS Decryption Engine)');
        console.log('======================================================');
        console.log('Listening for OS Proxy Traffic on 127.0.0.1:8080');
        console.log('Target Hosts: ', TARGET_HOSTS.join(' | '));
        console.log('Ensure process runs with $env:NODE_TLS_REJECT_UNAUTHORIZED="0"');
        console.log('======================================================\n');
    });
});

process.on('uncaughtException', e => {
    // Suppress console spam if sockets close abruptly
});
