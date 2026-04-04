import WebSocket from 'ws';
import http from 'http';
import { execSync } from 'child_process';

// Signal remaining NodeService PIDs
const otherPids = [38320, 43048, 34812, 41784];
for (const pid of otherPids) {
    try { process._debugProcess(pid); console.log('signaled', pid); }
    catch(e) { console.log('fail', pid, e.message); }
}

await new Promise(r => setTimeout(r, 2000));

// Check what ports opened for those PIDs
const out = execSync('netstat -ano', { encoding: 'utf8' });
const listening = out.split('\n').filter(l =>
    l.includes('LISTENING') && otherPids.some(p => l.trim().endsWith(String(p)))
);
console.log('New NodeService ports:\n', listening.join('\n'));

// Now probe ExtHost-A for require/mainModule
const wsUrl = await new Promise((resolve, reject) => {
    http.get({ host: '127.0.0.1', port: 6188, path: '/json' }, res => {
        let d = ''; res.on('data', c => d += c);
        res.on('end', () => { try { resolve(JSON.parse(d)[0]?.webSocketDebuggerUrl); } catch(e) { reject(e); } });
    }).on('error', reject);
});

if (wsUrl) {
    await new Promise(resolve => {
        const ws = new WebSocket(wsUrl);
        ws.on('open', () => {
            ws.send(JSON.stringify({ id: 1, method: 'Runtime.evaluate', params: {
                expression: `JSON.stringify({
                    hasPMR: typeof process !== 'undefined' && !!(process.mainModule && process.mainModule.require),
                    hasGlobalRequire: typeof require !== 'undefined',
                    processKeys: typeof process !== 'undefined' ? Object.keys(process).filter(k => k.match(/require|module|load/i)) : [],
                    nodeEnv: typeof process !== 'undefined' ? process.env.NODE_ENV : null,
                    execPath: typeof process !== 'undefined' ? process.execPath.substring(0,40) : null
                })`,
                returnByValue: true
            }}));
        });
        ws.on('message', raw => {
            const msg = JSON.parse(raw.toString());
            if (msg.id === 1) {
                console.log('ExtHost-A inspect result:', msg.result?.result?.value);
                ws.close(); resolve();
            }
        });
        ws.on('error', e => { console.log('ws error:', e.message); resolve(); });
        setTimeout(resolve, 5000);
    });
}
