/**
 * ALPN Downgrade Reality Test
 * 
 * Groq API에 실제 AI 질문을 보내고, MITM 프록시가 
 * 응답(AI 생성 텍스트)을 평문으로 캡처하는지 검증한다.
 * 
 * Flow:
 *   Client --[HTTPS via MITM :RANDOM]--> Groq API
 *                    |
 *              MITM 캡처 파일 생성
 *                    |
 *              평문 AI 응답이 보이면 성공
 */

const https = require('https');
const fs = require('fs');
const path = require('path');

const GROQ_KEY = process.env.GROQ_API_KEY || '';
const DUMP_DIR = path.join(__dirname, 'alpn_captures');
if (!fs.existsSync(DUMP_DIR)) fs.mkdirSync(DUMP_DIR, { recursive: true });

// Clear previous captures
fs.readdirSync(DUMP_DIR).forEach(f => fs.unlinkSync(path.join(DUMP_DIR, f)));

const AI_QUESTION = 'What is 2+2? Answer in one word.';

// ─── MITM Server ───
function startMITM(pfxPath) {
    return new Promise((resolve, reject) => {
        const mitm = https.createServer({
            pfx: fs.readFileSync(pfxPath),
            passphrase: '1234',
            ALPNProtocols: ['http/1.1']  // <-- ALPN downgrade
        }, (req, res) => {
            console.log(`  [MITM] ${req.method} ${req.headers.host}${req.url}`);
            
            // Capture request body
            const reqChunks = [];
            req.on('data', c => reqChunks.push(c));
            req.on('end', () => {
                const reqBody = Buffer.concat(reqChunks);
                const reqText = reqBody.toString('utf-8');
                
                // Save captured request
                const reqFile = path.join(DUMP_DIR, 'captured_request.json');
                fs.writeFileSync(reqFile, reqText);
                console.log(`  [MITM] Request captured: ${reqText.length}B`);
                
                // Check if it's readable plaintext
                try {
                    const parsed = JSON.parse(reqText);
                    console.log(`  [MITM] Request IS JSON plaintext!`);
                    if (parsed.messages) {
                        const userMsg = parsed.messages.find(m => m.role === 'user');
                        if (userMsg) {
                            console.log(`  [MITM] >>> USER MESSAGE VISIBLE: "${userMsg.content}"`);
                        }
                    }
                } catch {
                    console.log(`  [MITM] Request is NOT JSON (binary?)`);
                }

                // Forward to real Groq API
                const fwdReq = https.request({
                    hostname: 'api.groq.com',
                    port: 443,
                    path: req.url,
                    method: req.method,
                    headers: {
                        ...req.headers,
                        host: 'api.groq.com'
                    }
                }, (upRes) => {
                    const resChunks = [];
                    upRes.on('data', c => resChunks.push(c));
                    upRes.on('end', () => {
                        const resBody = Buffer.concat(resChunks);
                        const resText = resBody.toString('utf-8');
                        
                        // Save captured response
                        const resFile = path.join(DUMP_DIR, 'captured_response.json');
                        fs.writeFileSync(resFile, resText);
                        console.log(`  [MITM] Response captured: ${resText.length}B`);
                        
                        // Check if AI response is readable
                        try {
                            const parsed = JSON.parse(resText);
                            console.log(`  [MITM] Response IS JSON plaintext!`);
                            const aiText = parsed.choices?.[0]?.message?.content;
                            if (aiText) {
                                console.log(`  [MITM] >>> AI RESPONSE VISIBLE: "${aiText}"`);
                            }
                        } catch {
                            console.log(`  [MITM] Response is NOT JSON`);
                        }
                        
                        res.writeHead(upRes.statusCode, upRes.headers);
                        res.end(resBody);
                    });
                });
                fwdReq.on('error', e => {
                    console.log(`  [MITM] Forward error: ${e.message}`);
                    res.writeHead(502);
                    res.end(e.message);
                });
                fwdReq.write(reqBody);
                fwdReq.end();
            });
        });

        mitm.listen(0, '127.0.0.1', () => {
            const port = mitm.address().port;
            resolve({ server: mitm, port });
        });
        mitm.on('error', reject);
    });
}

// ─── Client: send real AI request through MITM ───
function sendAIRequest(mitmPort) {
    return new Promise((resolve, reject) => {
        const body = JSON.stringify({
            model: 'llama-3.3-70b-versatile',
            messages: [{ role: 'user', content: AI_QUESTION }],
            temperature: 0,
            max_tokens: 10
        });

        const req = https.request({
            hostname: '127.0.0.1',
            port: mitmPort,
            path: '/openai/v1/chat/completions',
            method: 'POST',
            headers: {
                'Host': 'api.groq.com',
                'Authorization': `Bearer ${GROQ_KEY}`,
                'Content-Type': 'application/json',
                'Content-Length': Buffer.byteLength(body)
            },
            rejectUnauthorized: false  // Accept self-signed cert
        }, (res) => {
            const chunks = [];
            res.on('data', c => chunks.push(c));
            res.on('end', () => {
                const text = Buffer.concat(chunks).toString('utf-8');
                resolve({ status: res.statusCode, body: text });
            });
        });

        req.on('error', e => reject(e));
        req.write(body);
        req.end();
    });
}

// ─── Main ───
async function main() {
    console.log('');
    console.log('============================================');
    console.log('  ALPN Downgrade REALITY TEST');
    console.log('  Real AI request through MITM proxy');
    console.log('============================================');
    console.log('');
    console.log(`  Question: "${AI_QUESTION}"`);
    console.log('');

    const pfxPath = 'C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\mitm.pfx';
    if (!fs.existsSync(pfxPath)) {
        console.log('ERROR: mitm.pfx not found');
        return;
    }

    // 1. Start MITM
    console.log('[1] Starting MITM proxy (ALPNProtocols: http/1.1)...');
    const { server, port } = await startMITM(pfxPath);
    console.log(`    Listening on port ${port}`);
    console.log('');

    // 2. Send real AI request
    console.log('[2] Sending real AI request through MITM...');
    console.log('');
    
    try {
        const result = await sendAIRequest(port);
        console.log('');
        console.log('[3] Client received response:');
        console.log(`    Status: ${result.status}`);
        
        try {
            const parsed = JSON.parse(result.body);
            const aiText = parsed.choices?.[0]?.message?.content;
            console.log(`    AI Answer: "${aiText}"`);
        } catch {
            console.log(`    Raw: ${result.body.substring(0, 200)}`);
        }
    } catch (e) {
        console.log(`    ERROR: ${e.message}`);
    }

    // 3. Show captured files
    console.log('');
    console.log('[4] Captured files in', DUMP_DIR + ':');
    const files = fs.readdirSync(DUMP_DIR);
    for (const f of files) {
        const fp = path.join(DUMP_DIR, f);
        const content = fs.readFileSync(fp, 'utf-8');
        console.log(`    ${f} (${content.length}B)`);
        console.log(`    Preview: ${content.replace(/\n/g, ' ').substring(0, 150)}`);
        console.log('');
    }

    console.log('============================================');
    console.log('  VERDICT:');
    
    const reqFile = path.join(DUMP_DIR, 'captured_request.json');
    const resFile = path.join(DUMP_DIR, 'captured_response.json');
    
    if (fs.existsSync(reqFile) && fs.existsSync(resFile)) {
        let reqOk = false, resOk = false;
        try { JSON.parse(fs.readFileSync(reqFile, 'utf-8')); reqOk = true; } catch {}
        try { JSON.parse(fs.readFileSync(resFile, 'utf-8')); resOk = true; } catch {}
        
        if (reqOk && resOk) {
            console.log('  REQUEST:  PLAINTEXT JSON (user message visible)');
            console.log('  RESPONSE: PLAINTEXT JSON (AI answer visible)');
            console.log('  RESULT:   ALPN DOWNGRADE WORKS - FULL PLAINTEXT CAPTURE');
        } else {
            console.log(`  REQUEST:  ${reqOk ? 'JSON' : 'BINARY/UNKNOWN'}`);
            console.log(`  RESPONSE: ${resOk ? 'JSON' : 'BINARY/UNKNOWN'}`);
        }
    } else {
        console.log('  No captures found - MITM may have failed');
    }
    
    console.log('============================================');

    server.close();
}

main().catch(e => console.error('Fatal:', e));
