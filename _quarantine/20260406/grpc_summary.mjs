import fs from 'fs';
import path from 'path';

const dir = 'C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\brain_v4\\_agents\\global_inbox\\cdp_captures';
const paths = {};

for (const f of fs.readdirSync(dir).filter(f => f.endsWith('.meta.json'))) {
    try {
        const m = JSON.parse(fs.readFileSync(path.join(dir, f)));
        if (m.label && m.label.startsWith('h2_')) {
            const colonIdx = m.label.indexOf(':/');
            const type = m.label.substring(0, colonIdx);
            const rpcPath = m.label.substring(colonIdx + 1).split(':')[0];
            const key = type + '|' + rpcPath;
            if (!paths[key]) paths[key] = { c: 0, s: 0 };
            paths[key].c++;
            paths[key].s += m.size || 0;
        }
    } catch {}
}

const sorted = Object.entries(paths).sort((a, b) => b[1].s - a[1].s);
console.log('=== gRPC Captures Summary ===');
for (const [k, v] of sorted) {
    console.log(`${v.c}x (${(v.s / 1024).toFixed(1)}KB)  ${k}`);
}
console.log(`\nTotal: ${sorted.reduce((a, b) => a + b[1].c, 0)} captures`);
