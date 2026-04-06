import fs from 'fs';
import path from 'path';

const dir = 'C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\brain_v4\\_agents\\global_inbox\\cdp_captures';
const files = fs.readdirSync(dir).filter(f => f.startsWith('chat_') && f.endsWith('.meta.json'));
const paths = {};

for (const f of files) {
    try {
        const m = JSON.parse(fs.readFileSync(path.join(dir, f)));
        if (m.label) {
            const key = m.label.substring(0, 100);
            if (!paths[key]) paths[key] = { c: 0, s: 0 };
            paths[key].c++;
            paths[key].s += m.size || 0;
        }
    } catch {}
}

const sorted = Object.entries(paths).sort((a, b) => b[1].s - a[1].s);
console.log('=== PID 50980 (chat) gRPC Summary ===');
console.log('Total chat_* captures:', files.length);
console.log('');
for (const [k, v] of sorted) {
    console.log(`${v.c}x (${(v.s / 1024).toFixed(1)}KB)  ${k}`);
}
