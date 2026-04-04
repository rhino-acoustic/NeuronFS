/**
 * DOM 구조 프로브 — 실제 메시지 컨테이너 패턴 찾기
 */
import http from 'http';
import WebSocket from 'ws';

function getJson(url) {
    return new Promise((r,j) => http.get(url, res => {
        let d=''; res.on('data',c=>d+=c);
        res.on('end', () => { try{r(JSON.parse(d))}catch(e){j(e)} });
    }).on('error',j));
}

async function probe(wsUrl, targetName) {
    const ws = new WebSocket(wsUrl);
    await new Promise((r,j)=>{ws.on('open',r);ws.on('error',j)});
    let id=1; const p=new Map();
    ws.on('message', m=>{const d=JSON.parse(m);if(d.id&&p.has(d.id)){p.get(d.id)(d);p.delete(d.id)}});
    const call=(method,params)=>new Promise((resolve,reject)=>{
        const i=id++; const t=setTimeout(()=>{p.delete(i);reject('timeout')},8000);
        p.set(i,d=>{clearTimeout(t);resolve(d)});
        ws.send(JSON.stringify({id:i,method,params}));
    });
    await call('Runtime.enable',{});
    await new Promise(r=>setTimeout(r,300));

    const script = `(() => {
        function walk(root) {
            const f = [];
            const w = n => {
                if(!n) return;
                if(n.shadowRoot) w(n.shadowRoot);
                for(const c of (n.children||[])) if(c.nodeType===1){f.push(c);w(c)}
            };
            w(root); return f;
        }
        const all = walk(document.body);
        
        // 1. 텍스트 길이가 50자 이상인 보이는 요소들의 클래스/속성 수집
        const candidates = all.filter(el => {
            if(el.offsetParent === null) return false;
            const text = (el.innerText||'').trim();
            return text.length > 50 && text.length < 5000 && el.children.length < 20;
        }).slice(0,20).map(el => ({
            tag: el.tagName,
            cls: (el.className||'').toString().slice(0,120),
            attrs: Object.fromEntries([...el.attributes].map(a=>[a.name,a.value.slice(0,50)])),
            textLen: (el.innerText||'').length,
            text: (el.innerText||'').slice(0,100)
        }));
        
        // 2. data-* 속성 가진 요소들
        const dataEls = all.filter(el => {
            return [...el.attributes].some(a => a.name.startsWith('data-') && 
                (a.name.includes('turn') || a.name.includes('message') || a.name.includes('role')));
        }).map(el => ({
            tag: el.tagName,
            attrs: Object.fromEntries([...el.attributes].map(a=>[a.name,a.value.slice(0,50)])),
            text: (el.innerText||'').slice(0,80)
        }));
        
        return { candidates: candidates.slice(0,10), dataEls: dataEls.slice(0,10) };
    })()`;

    const res = await call('Runtime.evaluate', {expression:script, returnByValue:true});
    const val = res?.result?.result?.value;
    
    console.log(`\n=== ${targetName} ===`);
    console.log('--- DATA-* elements ---');
    for(const el of (val?.dataEls||[])) {
        console.log(' ', JSON.stringify(el));
    }
    console.log('--- Text candidates ---');
    for(const el of (val?.candidates||[])) {
        console.log(' ', JSON.stringify(el));
    }
    ws.close();
}

(async()=>{
    const list = await getJson('http://127.0.0.1:9000/json/list');
    const targets = list.filter(t => t.url?.includes('workbench') && t.type==='page');
    for(const t of targets) {
        try { await probe(t.webSocketDebuggerUrl, t.title); }
        catch(e) { console.log(`ERR ${t.title}: ${e}`); }
    }
    process.exit(0);
})();
