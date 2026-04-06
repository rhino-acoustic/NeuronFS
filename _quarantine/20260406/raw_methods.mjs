import fs from 'fs';
import path from 'path';

const dir = 'C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\brain_v4\\_agents\\global_inbox\\cdp_captures';
const files = fs.readdirSync(dir).filter(f => f.startsWith('raw_') && !f.endsWith('.meta.json')).sort();
const methods = {};
let totalSize = 0;

for (const f of files) {
    const buf = fs.readFileSync(path.join(dir, f));
    totalSize += buf.length;
    const str = buf.toString('utf8');
    
    // Extract all JSON-RPC methods
    const ms = [...str.matchAll(/"method"\s*:\s*"([^"]+)"/g)];
    for (const m of ms) {
        if (!methods[m[1]]) methods[m[1]] = { count: 0, size: 0 };
        methods[m[1]].count++;
        methods[m[1]].size += buf.length;
    }
}

const sorted = Object.entries(methods).sort((a, b) => b[1].size - a[1].size);
console.log('=== PID 50980 -> port 14912 LSP Methods ===');
console.log('Files:', files.length, '| Total:', (totalSize / 1024).toFixed(1) + 'KB');
console.log('');
for (const [m, v] of sorted) {
    console.log(`${v.count}x (${(v.size / 1024).toFixed(1)}KB)  ${m}`);
}
