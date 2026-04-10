package main

const opsDashboardHTML = `<!DOCTYPE html>
<html lang="ko">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>NeuronFS Ops Dashboard</title>
<link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap" rel="stylesheet">
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:'Inter',sans-serif;background:#0a0a0f;color:#e0e0e0;min-height:100vh}
.header{background:linear-gradient(135deg,#0f1923 0%,#1a1a2e 100%);border-bottom:1px solid #1e293b;padding:20px 32px;display:flex;align-items:center;justify-content:space-between}
.header h1{font-size:18px;font-weight:700;color:#fff;display:flex;align-items:center;gap:10px}
.header h1 span{color:#38bdf8}
.header .meta{font-size:12px;color:#64748b}
.grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(320px,1fr));gap:16px;padding:24px 32px}
.card{background:#111827;border:1px solid #1e293b;border-radius:12px;padding:20px;transition:border-color .2s}
.card:hover{border-color:#334155}
.card h2{font-size:13px;font-weight:600;color:#94a3b8;text-transform:uppercase;letter-spacing:.5px;margin-bottom:16px}
.svc{display:flex;align-items:center;gap:12px;padding:10px 0;border-bottom:1px solid #1e293b}
.svc:last-child{border:none}
.dot{width:10px;height:10px;border-radius:50%;flex-shrink:0}
.dot.up{background:#22c55e;box-shadow:0 0 8px #22c55e66}
.dot.down{background:#ef4444;box-shadow:0 0 8px #ef444466;animation:pulse 1.5s infinite}
@keyframes pulse{0%,100%{opacity:1}50%{opacity:.4}}
.svc-info{flex:1}
.svc-name{font-size:14px;font-weight:600;color:#f1f5f9}
.svc-detail{font-size:11px;color:#64748b;margin-top:2px}
.svc-sla{font-size:20px;font-weight:700;text-align:right}
.sla-good{color:#22c55e}
.sla-warn{color:#eab308}
.sla-bad{color:#ef4444}
.metric-row{display:flex;justify-content:space-between;padding:8px 0;border-bottom:1px solid #1e293b}
.metric-row:last-child{border:none}
.metric-label{font-size:12px;color:#94a3b8}
.metric-value{font-size:14px;font-weight:600;color:#f1f5f9}
.region-bar{display:flex;align-items:center;gap:8px;padding:6px 0}
.region-name{font-size:12px;color:#94a3b8;width:100px}
.bar-track{flex:1;height:6px;background:#1e293b;border-radius:3px;overflow:hidden}
.bar-fill{height:100%;border-radius:3px;transition:width .6s ease}
.bar-brainstem{background:linear-gradient(90deg,#ef4444,#f97316)}
.bar-limbic{background:linear-gradient(90deg,#f97316,#eab308)}
.bar-hippocampus{background:linear-gradient(90deg,#eab308,#22c55e)}
.bar-sensors{background:linear-gradient(90deg,#22c55e,#06b6d4)}
.bar-cortex{background:linear-gradient(90deg,#3b82f6,#8b5cf6)}
.bar-ego{background:linear-gradient(90deg,#8b5cf6,#ec4899)}
.bar-prefrontal{background:linear-gradient(90deg,#ec4899,#f43f5e)}
.region-count{font-size:11px;color:#64748b;width:50px;text-align:right}
.log-item{display:flex;justify-content:space-between;padding:6px 0;border-bottom:1px solid #0f172a;font-size:11px}
.log-name{color:#94a3b8}
.log-size{color:#64748b}
.log-time{color:#475569;font-size:10px}
.status-banner{text-align:center;padding:12px;font-size:13px;font-weight:600;border-radius:8px;margin-bottom:16px}
.banner-nominal{background:#052e1633;color:#22c55e;border:1px solid #22c55e33}
.banner-degraded{background:#422006;color:#eab308;border:1px solid #eab30833}
.banner-critical{background:#450a0a;color:#ef4444;border:1px solid #ef444433}
.refresh-dot{display:inline-block;width:6px;height:6px;background:#22c55e;border-radius:50%;margin-right:6px;animation:pulse 2s infinite}
.footer{text-align:center;padding:16px;color:#334155;font-size:11px}
</style>
</head>
<body>
<div class="header">
  <h1><span>⚡</span> NeuronFS Ops Dashboard</h1>
  <div class="meta"><span class="refresh-dot"></span>자동 갱신 10초 | <span id="lastUpdate">-</span></div>
</div>

<div id="banner" class="status-banner banner-nominal">시스템 상태 로딩 중...</div>

<div class="grid">
  <div class="card" id="servicesCard">
    <h2>서비스 상태</h2>
    <div id="services">로딩 중...</div>
  </div>

  <div class="card">
    <h2>Watchdog 메트릭</h2>
    <div id="watchdogMetrics">로딩 중...</div>
  </div>

  <div class="card">
    <h2>뇌 상태 (Brain)</h2>
    <div id="brainState">로딩 중...</div>
  </div>

  <div class="card">
    <h2>로그 파일</h2>
    <div id="logFiles">로딩 중...</div>
  </div>
</div>

<div class="footer">NeuronFS Watchdog v4 Enterprise · Folder-as-Neuron Engine</div>

<script>
function formatBytes(b){if(!b)return '0B';const k=1024;const s=['B','KB','MB','GB'];const i=Math.floor(Math.log(b)/Math.log(k));return (b/Math.pow(k,i)).toFixed(1)+s[i]}
function formatDuration(ms){const m=Math.floor(ms/60000);const h=Math.floor(m/60);if(h>0)return h+'h '+m%60+'m';return m+'m'}
function slaClass(v){if(v>=99.5)return 'sla-good';if(v>=95)return 'sla-warn';return 'sla-bad'}

async function refresh(){
  try{
    const r=await fetch('/api/ops');
    const d=await r.json();
    document.getElementById('lastUpdate').textContent=new Date().toLocaleTimeString('ko-KR');

    // Banner
    const banner=document.getElementById('banner');
    const wd=d.watchdog;
    if(!wd){
      banner.className='status-banner banner-degraded';
      banner.textContent='⚠️ Watchdog 메트릭 없음 — watchdog가 아직 시작되지 않았습니다';
    }else{
      const anyDown=wd.services?.some(s=>s.status==='DOWN');
      const anyZombie=wd.services?.some(s=>s.zombieCount>0);
      if(anyDown){banner.className='status-banner banner-critical';banner.textContent='🔴 CRITICAL — 서비스 다운 감지';}
      else if(anyZombie){banner.className='status-banner banner-degraded';banner.textContent='⚠️ DEGRADED — 좀비 프로세스 감지';}
      else{banner.className='status-banner banner-nominal';banner.textContent='🟢 NOMINAL — 모든 서비스 정상';}
    }

    // Services
    const svcEl=document.getElementById('services');
    if(wd?.services){
      svcEl.innerHTML=wd.services.map(s=>{
        const isUp=s.status==='UP';
        const sla=s.sla||100;
        return '<div class="svc">'+
          '<div class="dot '+(isUp?'up':'down')+'"></div>'+
          '<div class="svc-info"><div class="svc-name">'+s.name+'</div>'+
          '<div class="svc-detail">restarts: '+s.restarts+(s.zombieCount?' | zombie: '+s.zombieCount:'')+
          (s.healthChecks?.length?' | checks: '+s.healthChecks.join(', '):'')+
          '</div></div>'+
          '<div class="svc-sla '+slaClass(sla)+'">'+sla.toFixed(1)+'%</div></div>';
      }).join('');
    }

    // Watchdog metrics
    const wmEl=document.getElementById('watchdogMetrics');
    if(wd){
      wmEl.innerHTML=[
        ['가동 시간',formatDuration(wd.uptimeMs)],
        ['체크 횟수',wd.checkCount+'회'],
        ['평균 체크',wd.avgCheckMs+'ms'],
        ['부팅 시각',new Date(wd.bootTime).toLocaleString('ko-KR')],
        ['마지막 갱신',new Date(wd.ts).toLocaleTimeString('ko-KR')]
      ].map(([l,v])=>'<div class="metric-row"><span class="metric-label">'+l+'</span><span class="metric-value">'+v+'</span></div>').join('');
    }

    // Brain
    const brEl=document.getElementById('brainState');
    const br=d.brain;
    if(br){
      const maxAct=Math.max(...br.regions.map(r=>r.activation),1);
      let html='<div class="metric-row"><span class="metric-label">전체 뉴런</span><span class="metric-value">'+br.totalNeurons+'</span></div>';
      html+='<div class="metric-row"><span class="metric-label">전체 활성도</span><span class="metric-value">'+br.totalActivation.toLocaleString()+'</span></div>';
      html+='<div style="margin-top:12px">';
      br.regions.forEach(r=>{
        const pct=Math.max(2,r.activation/maxAct*100);
        html+='<div class="region-bar"><span class="region-name">'+r.name+'</span>'+
          '<div class="bar-track"><div class="bar-fill bar-'+r.name+'" style="width:'+pct+'%"></div></div>'+
          '<span class="region-count">'+r.neurons+'</span></div>';
      });
      html+='</div>';
      brEl.innerHTML=html;
    }

    // Logs
    const lgEl=document.getElementById('logFiles');
    if(d.logs){
      lgEl.innerHTML=d.logs.sort((a,b)=>new Date(b.modified)-new Date(a.modified)).map(l=>
        '<div class="log-item"><span class="log-name">'+l.name+'</span>'+
        '<span class="log-size">'+formatBytes(l.size)+'</span>'+
        '<span class="log-time">'+new Date(l.modified).toLocaleTimeString('ko-KR')+'</span></div>'
      ).join('');
    }
  }catch(e){
    document.getElementById('banner').className='status-banner banner-critical';
    document.getElementById('banner').textContent='🔴 API 연결 실패: '+e.message;
  }
}

refresh();
setInterval(refresh,10000);
</script>
</body>
</html>`
