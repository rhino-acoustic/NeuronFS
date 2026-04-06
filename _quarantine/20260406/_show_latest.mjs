import fs from 'fs';
import path from 'path';

const dir = 'C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\brain_v4\\_agents\\global_inbox\\cdp_captures';
const files = fs.readdirSync(dir).filter(f => f.startsWith('chat_net_') && !f.endsWith('.meta.json')).sort().reverse();

for (const f of files.slice(0, 3)) {
    const data = fs.readFileSync(path.join(dir, f), 'utf8');
    const meta = JSON.parse(fs.readFileSync(path.join(dir, f + '.meta.json'), 'utf8'));
    console.log(`\n${'='.repeat(60)}`);
    console.log(`File: ${f} | ${data.length}B | ${meta.label}`);
    console.log(`${'='.repeat(60)}`);
    
    try {
        const json = JSON.parse(data);
        console.log(JSON.stringify(json, null, 2).substring(0, 1000));
    } catch {
        console.log(data.substring(0, 500));
    }
}
