import fs from 'fs';
import path from 'path';

const dir = 'C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\brain_v4\\_agents\\global_inbox\\cdp_captures';
const files = fs.readdirSync(dir)
    .filter(f => (f.startsWith('chat_net_') || f.startsWith('net_')) && !f.endsWith('.meta.json'))
    .sort().reverse();

console.log('=== Captured gRPC-Web Requests ===\n');
console.log(`Total: ${files.length} files\n`);

for (const f of files.slice(0, 15)) {
    try {
        const data = fs.readFileSync(path.join(dir, f), 'utf8');
        const metaPath = path.join(dir, f + '.meta.json');
        const meta = fs.existsSync(metaPath) ? JSON.parse(fs.readFileSync(metaPath, 'utf8')) : {};
        
        const url = meta.label ? meta.label.replace('CHAT_NET:', '').replace('cdp_net:', '') : '?';
        const rpc = url.split('/').pop();
        
        console.log(`--- ${f} (${data.length}B) | ${rpc} ---`);
        
        // JSON이면 파싱해서 핵심 필드만 출력
        try {
            const json = JSON.parse(data);
            if (json.cascadeId) console.log(`  cascadeId: ${json.cascadeId}`);
            if (json.items) {
                for (const item of json.items) {
                    if (item.text) console.log(`  📝 USER TEXT: "${item.text}"`);
                    if (item.userRequest) console.log(`  📝 USER REQUEST: "${JSON.stringify(item.userRequest).substring(0, 200)}"`);
                }
            }
            if (json.interaction) {
                const inter = json.interaction;
                if (inter.runCommand) console.log(`  🖥️ CMD: confirm=${inter.runCommand.confirm} "${(inter.runCommand.proposedCommandLine || '').substring(0, 100)}"`);
                if (inter.userMessage) console.log(`  📝 USER MSG: "${inter.userMessage}"`);
                if (inter.stepIndex) console.log(`  step: ${inter.stepIndex}`);
            }
            // 기타 키
            const keys = Object.keys(json).filter(k => k !== 'cascadeId' && k !== 'items' && k !== 'interaction');
            if (keys.length > 0) console.log(`  other keys: ${keys.join(', ')}`);
        } catch {
            console.log(`  (raw): ${data.substring(0, 200)}`);
        }
        console.log('');
    } catch {}
}
