// extension.js에서 protobuf descriptor base64를 추출하고 텍스트 부분만 출력
import fs from 'fs';

const content = fs.readFileSync(
    "C:\\Users\\BASEMENT_ADMIN\\AppData\\Local\\Programs\\Antigravity\\resources\\app\\extensions\\antigravity\\dist\\extension.js", 
    'utf8'
);

const regex = /fileDesc\)\("([A-Za-z0-9+\/=]+)"/g;
let match;
let i = 0;
const results = [];

while ((match = regex.exec(content)) !== null) {
    i++;
    const b64 = match[1];
    try {
        // pad if needed
        const padded = b64 + '='.repeat((4 - b64.length % 4) % 4);
        const buf = Buffer.from(padded, 'base64');
        
        // Extract readable strings from binary protobuf descriptor
        const strings = [];
        let current = '';
        for (const byte of buf) {
            if (byte >= 32 && byte <= 126) {
                current += String.fromCharCode(byte);
            } else {
                if (current.length > 3) strings.push(current);
                current = '';
            }
        }
        if (current.length > 3) strings.push(current);
        
        const label = strings.find(s => s.includes('.proto')) || strings[0] || 'unknown';
        
        fs.writeFileSync(`C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\runtime\\proto_desc_${i}.bin`, buf);
        
        const line = `=== Proto ${i} (${buf.length} bytes) — ${label} ===\n  Strings: ${strings.join(' | ')}\n`;
        results.push(line);
        console.log(line);
    } catch(e) {
        console.log(`Proto ${i}: decode error — ${e.message}`);
    }
}

fs.writeFileSync(
    'C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\runtime\\proto_analysis.txt', 
    results.join('\n'),
    'utf8'
);
console.log(`\nTotal: ${i} descriptors extracted`);
