import fs from 'fs';
import path from 'path';

const dir = 'C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\brain_v4\\_agents\\global_inbox\\cdp_captures';
const files = fs.readdirSync(dir).filter(f => f.endsWith('.meta.json'));
const methods = {};
let cascadeCount = 0;
let chatKeywords = 0;

for (const f of files) {
    try {
        const m = JSON.parse(fs.readFileSync(path.join(dir, f)));
        if (m.label && m.label.startsWith('sock_')) {
            const binFile = f.replace('.meta.json', '');
            const buf = fs.readFileSync(path.join(dir, binFile));
            const str = buf.toString('utf8');
            
            // 채팅/cascade 관련 키워드 검색
            const keywords = ['cascade', 'Cascade', 'SendUser', 'sendUser', 'chat', 'Chat', 'userMessage', 'cortex', 'Cortex'];
            let found = false;
            for (const kw of keywords) {
                if (str.includes(kw)) {
                    found = true;
                    chatKeywords++;
                    break;
                }
            }
            
            // LSP method 추출
            const methodMatch = str.match(/"method"\s*:\s*"([^"]+)"/g);
            if (methodMatch) {
                for (const mm of methodMatch) {
                    const m2 = mm.match(/"method"\s*:\s*"([^"]+)"/);
                    if (m2) {
                        if (!methods[m2[1]]) methods[m2[1]] = { count: 0, hasCascade: false };
                        methods[m2[1]].count++;
                        if (found) methods[m2[1]].hasCascade = true;
                    }
                }
            }
            
            if (found) cascadeCount++;
        }
    } catch {}
}

console.log('=== LSP Methods with Chat/Cascade Keywords ===');
console.log('Total sock captures with cascade/chat:', cascadeCount);
console.log('Total keyword hits:', chatKeywords);
console.log('');

const sorted = Object.entries(methods).sort((a, b) => b[1].count - a[1].count);
for (const [m, v] of sorted) {
    const marker = v.hasCascade ? ' ***' : '';
    console.log(`${v.count}x  ${m}${marker}`);
}
