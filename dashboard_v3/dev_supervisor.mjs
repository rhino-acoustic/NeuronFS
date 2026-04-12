import { spawn } from 'child_process';
import WebSocket from 'ws';

function connectWS() {
    let ws = new WebSocket('ws://localhost:7350/api/ws');
    
    ws.on('error', () => {
        // Silent reconnect
    });

    ws.on('close', () => {
        setTimeout(connectWS, 3000);
    });

    return ws;
}

let ws = connectWS();

console.log("🚀 NeuronFS V4 TurboPack Auto-Fix Supervisor Initiated.");
const next = spawn('npx', ['next', 'dev'], { stdio: ['ignore', 'pipe', 'pipe'], shell: true });

let errorBuffer = '';
let captureTimer = null;

function broadcastError() {
    if (ws.readyState === WebSocket.OPEN && errorBuffer.trim().length > 0) {
        console.log(`\n[Auto-Fix] Exception Intercepted. Broadcasting to NeuronFS Cortex...`);
        ws.send(JSON.stringify({
            action: "turbopack_error",
            message: errorBuffer.slice(0, 3000) // Truncate to avoid massive payloads
        }));
    }
    errorBuffer = '';
}

next.stdout.on('data', (data) => {
    const str = data.toString();
    process.stdout.write(str);
    
    if (str.includes('Failed to compile') || str.includes('Type error:') || str.includes('Unhandled Runtime Error')) {
        errorBuffer += str;
        
        if (captureTimer) clearTimeout(captureTimer);
        // Wait 2 seconds for the full stack trace to flush out
        captureTimer = setTimeout(broadcastError, 2000);
    } else if (errorBuffer.length > 0) {
        errorBuffer += str;
    }
});

next.stderr.on('data', (data) => {
    const str = data.toString();
    process.stderr.write(str);
    if (errorBuffer.length > 0) {
        errorBuffer += str;
    }
});

process.on('SIGINT', () => {
    next.kill('SIGINT');
    process.exit();
});
