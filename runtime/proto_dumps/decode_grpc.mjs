import fs from 'fs';
import path from 'path';

const dir = 'C:\\Users\\BASEMENT_ADMIN\\NeuronFS\\brain_v4\\_agents\\global_inbox\\cdp_captures';
const target = process.argv[2] || 'HandleStreamingCommand';
const maxFiles = parseInt(process.argv[3]) || 3;

function rv(b, o) {
    let r = 0, s = 0, i = o;
    while (i < b.length) { const c = b[i]; r |= (c & 0x7f) << s; i++; if (!(c & 0x80)) break; s += 7; if (s > 49) break; }
    return { v: r, n: i - o };
}

function decode(buf, maxDepth = 20) {
    const fields = [];
    let p = 0;
    while (p < buf.length && fields.length < maxDepth) {
        const t = rv(buf, p);
        if (t.n === 0) break;
        p += t.n;
        const fn = t.v >>> 3, wt = t.v & 7;
        if (fn === 0 || fn > 999) break;
        if (wt === 0) {
            const v = rv(buf, p); p += v.n;
            fields.push({ f: fn, type: 'int', v: v.v });
        } else if (wt === 2) {
            const l = rv(buf, p); p += l.n;
            const d = buf.slice(p, p + l.v); p += l.v;
            let isStr = true;
            for (let i = 0; i < Math.min(d.length, 50); i++) {
                if (d[i] < 9 || (d[i] > 13 && d[i] < 32 && d[i] !== 27)) { isStr = false; break; }
            }
            if (isStr && d.length > 0 && d.length < 500) {
                fields.push({ f: fn, type: 'str', v: d.toString('utf8') });
            } else {
                fields.push({ f: fn, type: 'bytes', len: d.length, data: d });
            }
        } else if (wt === 1) { p += 8; fields.push({ f: fn, type: 'f64' }); }
        else if (wt === 5) { p += 4; fields.push({ f: fn, type: 'f32' }); }
        else break;
    }
    return fields;
}

// Find matching files
const files = [];
for (const f of fs.readdirSync(dir).filter(f => f.endsWith('.meta.json'))) {
    try {
        const m = JSON.parse(fs.readFileSync(path.join(dir, f)));
        if (m.label && m.label.includes(target)) {
            files.push({ bin: f.replace('.meta.json', ''), label: m.label, size: m.size });
        }
    } catch {}
}
files.sort((a, b) => b.size - a.size);

console.log(`\n=== ${target} — ${files.length} captures ===\n`);

for (const f of files.slice(0, maxFiles)) {
    const buf = fs.readFileSync(path.join(dir, f.bin));
    console.log(`--- ${f.bin} (${f.size}B) ---`);
    console.log(`    Label: ${f.label}`);
    
    // Strip gRPC frame if present
    let payload = buf;
    if (buf.length >= 5 && (buf[0] === 0 || buf[0] === 1)) {
        const frameLen = buf.readUInt32BE(1);
        if (frameLen + 5 <= buf.length && frameLen > 0) {
            payload = buf.slice(5, 5 + frameLen);
            console.log(`    gRPC frame: flag=${buf[0]}, payload=${frameLen}B`);
        }
    }
    
    // Level 1 decode
    const level1 = decode(payload, 30);
    for (const field of level1) {
        if (field.type === 'str') {
            const truncated = field.v.length > 200 ? field.v.substring(0, 200) + '...' : field.v;
            console.log(`  F${field.f}: "${truncated}"`);
        } else if (field.type === 'int') {
            console.log(`  F${field.f}: ${field.v}`);
        } else if (field.type === 'bytes') {
            // Try to decode nested level
            const nested = decode(field.data, 15);
            if (nested.length > 1) {
                console.log(`  F${field.f}: {${field.len}B, ${nested.length} subfields}`);
                for (const sub of nested.slice(0, 10)) {
                    if (sub.type === 'str') {
                        const t = sub.v.length > 150 ? sub.v.substring(0, 150) + '...' : sub.v;
                        console.log(`    F${sub.f}: "${t}"`);
                    } else if (sub.type === 'int') {
                        console.log(`    F${sub.f}: ${sub.v}`);
                    } else if (sub.type === 'bytes') {
                        console.log(`    F${sub.f}: [${sub.len}B]`);
                    }
                }
            } else {
                console.log(`  F${field.f}: [${field.len}B]`);
            }
        }
    }
    console.log('');
}
