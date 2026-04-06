import http from 'http';
import WebSocket from 'ws';

function getJson(url) {
    return new Promise((r,j)=>http.get(url,res=>{let d='';res.on('data',c=>d+=c);res.on('end',()=>r(JSON.parse(d)))}).on('error',j));
}

const list = await getJson('http://127.0.0.1:9000/json/list');
const target = list.find(t=>t.url?.includes('workbench') && t.type==='page' && t.title?.includes('BASEMENT_ADMIN'));
if(!target){ console.log('Not found. Targets:', list.map(t=>t.title)); process.exit(1); }

const ws = new WebSocket(target.webSocketDebuggerUrl);
await new Promise((r,j)=>{ws.on('open',r);ws.on('error',j)});
let id=1; const p=new Map();
ws.on('message',m=>{const d=JSON.parse(m);if(d.id&&p.has(d.id)){p.get(d.id)(d);p.delete(d.id)}});
const call=(method,params)=>new Promise((resolve,reject)=>{
    const i=id++; const t=setTimeout(()=>{p.delete(i);reject('timeout')},5000);
    p.set(i,d=>{clearTimeout(t);resolve(d)});
    ws.send(JSON.stringify({id:i,method,params}));
});
await call('Runtime.enable',{});
await new Promise(r=>setTimeout(r,300));

// Command Palette로 Restart Extension Host 실행
const result = await call('Runtime.evaluate', {
    expression: `(() => {
        // VS Code workbench service에 직접 접근
        try {
            const s = window._WORKBENCH_SERVICE || 
                      window.require && window.require('vs/workbench/services/extensions/common/extensionHostManager');
            return 'no_direct_api';
        } catch(e) { return String(e); }
    })()`,
    returnByValue: true
});
console.log('Result:', JSON.stringify(result?.result?.result?.value));

// 키보드 단축키로 Command Palette 열기 + restartExtensionHost
await call('Input.dispatchKeyEvent', {type:'keyDown', key:'P', code:'KeyP', modifiers:8, windowsVirtualKeyCode:80}); // Ctrl+Shift+P
await new Promise(r=>setTimeout(r,800));
await call('Input.insertText', {text:'Restart Extension Host'});
await new Promise(r=>setTimeout(r,600));
await call('Input.dispatchKeyEvent', {type:'keyDown', key:'Return', code:'Enter', windowsVirtualKeyCode:13});
await call('Input.dispatchKeyEvent', {type:'keyUp', key:'Return', code:'Enter', windowsVirtualKeyCode:13});
console.log('Restart Extension Host command sent');
ws.close();
