import fs from 'fs';
import path from 'path';

const dir = 'C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\brain_v4\\_agents\\global_inbox\\cdp_captures';
const files = fs.readdirSync(dir).filter(f => f.endsWith('.meta.json'));
const allPaths = new Set();

for (const f of files) {
    try {
        const m = JSON.parse(fs.readFileSync(path.join(dir, f)));
        if (m.label) {
            // h2_req, h2_res, h2_hdrs 에서 경로 추출
            const match = m.label.match(/h2_(?:req|res|hdrs):(.+)/);
            if (match) allPaths.add(match[1]);
        }
    } catch {}
}

console.log('=== ALL unique gRPC paths captured ===');
console.log('Total unique:', allPaths.size);
console.log('');
for (const p of [...allPaths].sort()) {
    console.log('  ' + p);
}

// Also check: is there ANY file containing "SendUser" or "cascade" in content?
console.log('\n=== Searching raw content for chat keywords ===');
const keywords = ['SendUser', 'sendUser', 'StartCascade', 'chatMessage', 'userMessage', 'executePrompt'];
let found = 0;

const allFiles = fs.readdirSync(dir).filter(f => !f.endsWith('.meta.json') && (f.startsWith('raw_') || f.startsWith('chat_') || f.startsWith('mem_')));
for (const f of allFiles.slice(-200)) { // last 200 files
    try {
        const buf = fs.readFileSync(path.join(dir, f));
        const str = buf.toString('utf8');
        for (const kw of keywords) {
            if (str.includes(kw)) {
                found++;
                console.log(`  FOUND "${kw}" in ${f} (${buf.length}B)`);
                console.log(`    context: ...${str.substring(Math.max(0, str.indexOf(kw) - 30), str.indexOf(kw) + kw.length + 50)}...`);
                break;
            }
        }
    } catch {}
}
if (found === 0) console.log('  No matches found in last 200 files');
