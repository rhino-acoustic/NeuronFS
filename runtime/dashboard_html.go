package main

// Dashboard HTML for NeuronFS v5.0 — 3D Brain Topology + Card UI

const dashboardHTML = `<!DOCTYPE html>
<html lang="ko">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>NeuronFS v5 — 인지 엔진</title>
<script src="https://cdnjs.cloudflare.com/ajax/libs/three.js/r128/three.min.js"></script>
<style>
  *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
  @import url('https://fonts.googleapis.com/css2?family=Inter:wght@400;600;700;900&display=swap');
  body {
    font-family: 'Inter', -apple-system, sans-serif;
    background: #09090b; color: #e0e0e0;
    min-height: 100vh; overflow: hidden;
  }

  /* ── Layout ── */
  .app { display: flex; height: 100vh; }
  #canvas3d { flex: 1; position: relative; }
  .sidebar {
    width: 420px; min-width: 380px; background: rgba(15,15,20,0.95);
    border-left: 1px solid #1a1a2e; overflow-y: auto;
    backdrop-filter: blur(20px); z-index: 10;
  }

  /* ── Header ── */
  .header {
    padding: 20px 24px; border-bottom: 1px solid #1a1a2e;
    display: flex; align-items: center; gap: 12px;
  }
  .header h1 { font-size: 16px; font-weight: 900; color: #fff; letter-spacing: -0.02em; }
  .badge {
    font-size: 10px; padding: 3px 10px; border-radius: 50px;
    font-weight: 700; letter-spacing: 0.05em;
  }
  .badge-ok { background: #064e3b; color: #34d399; }
  .badge-score {
    background: linear-gradient(135deg, #1e40af, #7c3aed);
    color: #fff; margin-left: auto; font-size: 12px; padding: 4px 14px;
  }

  /* ── Stats Bar ── */
  .stats {
    display: grid; grid-template-columns: repeat(3, 1fr);
    padding: 16px 24px; gap: 12px; border-bottom: 1px solid #1a1a2e;
  }
  .stat { text-align: center; }
  .stat-value {
    font-size: 28px; font-weight: 900;
    background: linear-gradient(135deg, #3b82f6, #8b5cf6);
    -webkit-background-clip: text; -webkit-text-fill-color: transparent;
  }
  .stat-label { font-size: 10px; color: #666; text-transform: uppercase; letter-spacing: 0.1em; }

  /* ── Detail Panel (appears on sphere click) ── */
  .detail-panel {
    padding: 20px 24px; border-bottom: 1px solid #1a1a2e;
    display: none; animation: slideIn 0.2s ease;
  }
  .detail-panel.active { display: block; }
  @keyframes slideIn { from { opacity: 0; transform: translateY(-8px); } to { opacity: 1; } }
  .detail-panel h2 {
    font-size: 14px; font-weight: 700; color: #fff;
    margin-bottom: 12px; display: flex; align-items: center; gap: 8px;
  }
  .detail-panel .close-btn {
    margin-left: auto; cursor: pointer; color: #666;
    font-size: 16px; background: none; border: none;
  }
  .detail-panel .close-btn:hover { color: #fff; }

  /* ── Connections (axons) ── */
  .connections { margin: 8px 0 12px; }
  .conn-line {
    display: flex; align-items: center; gap: 8px;
    font-size: 11px; color: #888; padding: 4px 0;
  }
  .conn-arrow { color: #3b82f6; font-weight: 700; }
  .conn-target { color: #e0e0e0; cursor: pointer; }
  .conn-target:hover { color: #3b82f6; text-decoration: underline; }

  /* ── Neuron list in detail ── */
  .neuron-list { max-height: 300px; overflow-y: auto; }
  .neuron-item {
    display: flex; align-items: center; gap: 8px;
    padding: 6px 0; border-bottom: 1px solid #111; font-size: 11px;
  }
  .neuron-item:last-child { border: none; }
  .n-name { flex: 1; color: #ccc; font-family: monospace; word-break: break-all; }
  .n-bar { width: 50px; height: 4px; background: #222; border-radius: 2px; overflow: hidden; }
  .n-fill { height: 100%; border-radius: 2px; }
  .n-counter { font-family: monospace; font-size: 10px; color: #666; width: 28px; text-align: right; }
  .n-signals { font-size: 10px; width: 32px; text-align: center; }
  .n-fire { background: none; border: 1px solid #333; border-radius: 4px; color: #888; font-size: 9px; padding: 2px 6px; cursor: pointer; transition: all 0.15s; }
  .n-fire:hover { border-color: #f59e0b; color: #f59e0b; }
  .n-strength { font-size: 8px; padding: 1px 5px; border-radius: 3px; font-weight: 700; margin-right: 4px; }
  .n-str-abs { background: #7f1d1d; color: #fca5a5; }
  .n-str-must { background: #1e3a5f; color: #93c5fd; }

  /* ── Search ── */
  .search-bar {
    padding: 12px 24px; border-bottom: 1px solid #1a1a2e;
  }
  .search-bar input {
    width: 100%; background: #111; border: 1px solid #222; border-radius: 8px;
    padding: 8px 12px; color: #fff; font-size: 12px; outline: none;
    transition: border-color 0.15s;
  }
  .search-bar input:focus { border-color: #3b82f6; }
  .search-bar input::placeholder { color: #555; }

  /* ── Add neuron form ── */
  .add-section {
    padding: 12px 24px; border-bottom: 1px solid #1a1a2e;
    display: none;
  }
  .add-section.visible { display: block; }
  .add-section .add-row { display: flex; gap: 6px; margin-top: 8px; }
  .add-section input, .add-section select {
    flex: 1; background: #111; border: 1px solid #222; border-radius: 6px;
    padding: 6px 10px; color: #fff; font-size: 11px; outline: none;
  }
  .add-section select { flex: none; width: 110px; }
  .add-section button {
    background: #059669; color: #fff; border: none; border-radius: 6px;
    padding: 6px 14px; font-size: 11px; cursor: pointer; font-weight: 600;
  }
  .add-section button:hover { background: #047857; }

  /* ── Sandbox ── */
  .sandbox-section {
    padding: 12px 24px; border-bottom: 1px solid #1a1a2e;
    display: none;
  }
  .sandbox-section.visible { display: block; }
  .sandbox-section textarea {
    width: 100%; background: #111; border: 1px solid #222; border-radius: 6px;
    padding: 8px 10px; color: #fff; font-size: 11px; font-family: monospace;
    outline: none; resize: vertical; min-height: 60px; max-height: 150px;
  }
  .sandbox-section textarea:focus { border-color: #f59e0b; }
  .sandbox-section .sandbox-row { display: flex; gap: 6px; margin-top: 8px; }
  .sandbox-section .sandbox-row button { flex: 1; }
  .sandbox-paths { margin-top: 8px; font-size: 10px; color: #34d399; font-family: monospace; }
  .sandbox-paths div { padding: 2px 0; }
  .add-label { font-size: 10px; color: #888; text-transform: uppercase; letter-spacing: 0.08em; }

  /* ── Region list ── */
  .region-list { padding: 8px 24px; }
  .region-chip {
    display: inline-flex; align-items: center; gap: 6px;
    padding: 6px 14px; margin: 4px; border-radius: 50px;
    font-size: 11px; font-weight: 600; cursor: pointer;
    border: 1px solid #222; background: #111; transition: all 0.15s;
  }
  .region-chip:hover { border-color: #3b82f6; background: #0f172a; }
  .region-chip.active { border-color: #3b82f6; background: #1e3a5f; color: #fff; }
  .region-chip .chip-count {
    background: #222; padding: 1px 6px; border-radius: 10px;
    font-size: 9px; color: #888;
  }

  /* ── Controls ── */
  .controls {
    padding: 16px 24px; border-top: 1px solid #1a1a2e;
    display: flex; gap: 8px; position: sticky; bottom: 0;
    background: rgba(15,15,20,0.95); backdrop-filter: blur(20px);
  }
  .btn {
    flex: 1; border: none; border-radius: 8px; padding: 10px;
    font-size: 11px; font-weight: 700; cursor: pointer; transition: all 0.15s;
  }
  .btn-primary { background: #1d4ed8; color: #fff; }
  .btn-primary:hover { background: #2563eb; }
  .btn-danger { background: #7f1d1d; color: #fca5a5; }
  .btn-danger:hover { background: #991b1b; }
  .btn-ghost { background: #1a1a2e; color: #888; }
  .btn-ghost:hover { background: #222; color: #fff; }
  .btn-add { background: #059669; color: #fff; }
  .btn-add:hover { background: #047857; }

  /* ── Toast notification ── */
  .toast {
    position: fixed; top: 16px; left: 16px; max-width: 300px;
    background: rgba(12,12,24,0.9); backdrop-filter: blur(16px);
    border: 1px solid rgba(255,255,255,0.08); border-left: 3px solid #34d399;
    border-radius: 8px; padding: 12px 20px; font-size: 9px; font-family: 'SUIT', sans-serif; color: #e0e0e0;
    z-index: 300; opacity: 0; pointer-events: auto; display: flex; justify-content: space-between; align-items: flex-start; gap: 12px;
    animation: slideInFromLeft 0.3s forwards; display: none;
  }
  .toast.show { display: flex; opacity: 1; }
  .toast.error { border-left-color: #ef4444; }
  .toast-close { background: none; border: none; color: #888; cursor: pointer; font-size: 12px; padding: 0; line-height: 1; }
  .toast-close:hover { color: #fff; }
  @keyframes slideInFromLeft { from { transform: translateX(-20px); opacity: 0; } to { transform: translateX(0); opacity: 1; } }

  /* ── System Health Panel ── */
  .system-health {
    padding: 12px 24px; border-bottom: 1px solid #1a1a2e;
  }
  .health-row {
    display: flex; align-items: center; gap: 8px;
    padding: 5px 0; font-size: 11px; border-bottom: 1px solid #111;
  }
  .health-row:last-child { border: none; }
  .health-dot {
    width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0;
  }
  .health-dot.on { background: #22c55e; box-shadow: 0 0 6px #22c55e; }
  .health-dot.off { background: #ef4444; box-shadow: 0 0 6px #ef4444; animation: pulse-red 1.5s infinite; }
  @keyframes pulse-red { 0%,100% { opacity: 1; } 50% { opacity: 0.4; } }
  .health-name { color: #e0e0e0; font-weight: 600; width: 100px; }
  .health-role { color: #666; font-size: 10px; flex: 1; }

  /* ── 3D overlay info ── */
  .hover-tooltip {
    position: absolute; pointer-events: none;
    background: rgba(255, 255, 255, 0.05); backdrop-filter: blur(20px); -webkit-backdrop-filter: blur(20px);
    border: 1px solid rgba(255, 255, 255, 0.15); border-radius: 12px;
    padding: 10px 16px; font-size: 12px; color: #fff;
    box-shadow: 0 4px 30px rgba(0, 0, 0, 0.5); display: none; z-index: 100; max-width: 320px; transition: opacity 0.2s ease;
  }
  .hover-tooltip .tt-region { color: #b6cfdd; font-weight: 700; display: block; margin-bottom: 4px; letter-spacing: 0.5px; }
  .hover-tooltip .tt-stats { color: #eadccf; font-size: 11px; }
</style>
</head>
<body>
<div class="app">
  <!-- 3D Canvas -->
  <div id="canvas3d">
    <div class="hover-tooltip" id="tooltip"></div>
  </div>

  <!-- Sidebar -->
  <div class="sidebar">
    <div class="header">
      <h1>🧠 NeuronFS v5</h1>
      <span class="badge badge-ok" id="status">NOMINAL</span>
      <span class="badge badge-score" id="score">0</span>
    </div>

    <div class="stats">
      <div class="stat"><div class="stat-value" id="s-neurons">0</div><div class="stat-label">뉴런</div></div>
      <div class="stat"><div class="stat-value" id="s-activation">0</div><div class="stat-label">활성도</div></div>
      <div class="stat"><div class="stat-value" id="s-regions">0</div><div class="stat-label">영역</div></div>
    </div>

    <div class="search-bar">
      <input type="text" id="searchInput" placeholder="뉴런 검색 (Ctrl+K)" oninput="filterNeurons()">
    </div>

    <div class="detail-panel" id="detail">
      <h2>
        <span id="detail-icon"></span>
        <span id="detail-name"></span>
        <button class="close-btn" onclick="closeDetail()">✕</button>
      </h2>
      <div class="connections" id="detail-axons"></div>
      <div class="neuron-list" id="detail-neurons"></div>
    </div>

    <div class="region-list" id="regionChips"></div>

    <div class="add-section" id="addSection">
      <div class="add-label">새 뉴런 생성</div>
      <div class="add-row">
        <select id="addRegion"></select>
        <input type="text" id="addPath" placeholder="경로 (예: methodology/tdd)">
        <button onclick="addNeuron()">+</button>
      </div>
    </div>

    <div class="sandbox-section" id="sandboxSection">
      <div class="add-label">🧪 Sandbox — 규칙 실험</div>
      <textarea id="sandboxText" placeholder="한 줄에 하나씩 규칙 입력&#10;예: 禁인라인스타일&#10;    항상_타입체크"></textarea>
      <div class="sandbox-row">
        <button class="btn btn-primary" onclick="applySandbox()" style="font-size:10px">✅ 적용</button>
        <button class="btn btn-ghost" onclick="clearSandbox()" style="font-size:10px">🗑 초기화</button>
      </div>
      <div class="sandbox-paths" id="sandboxPaths"></div>
    </div>

    <div class="system-health" id="healthPanel">
      <div class="add-label" style="margin-bottom:8px;display:flex;align-items:center;gap:6px">
        ⚙️ SYSTEM STATUS
        <span id="health-badge" style="font-size:9px;padding:2px 8px;border-radius:10px;background:#064e3b;color:#34d399;font-weight:700">ALL OK</span>
      </div>
      <div id="healthList"></div>
      <div style="margin-top:8px;font-size:9px;color:#444" id="healthMeta"></div>
    </div>

    <div class="controls">
      <button class="btn btn-primary" onclick="doInject()">⚡ INJECT</button>
      <button class="btn btn-add" onclick="toggleAdd()">+ 뉴런</button>
      <button class="btn btn-ghost" onclick="doDedup()">🔀 DEDUP</button>
      <button class="btn btn-ghost" onclick="toggleSandbox()" style="font-size:10px">🧪</button>
      <select id="bombRegion" style="background:#1a1a2e;color:#fca5a5;border:1px solid #7f1d1d;border-radius:6px;padding:6px 8px;font-size:10px;cursor:pointer;"><option value="">💀 영역 선택</option></select>
      <button class="btn btn-danger" onclick="doBomb()">💀 BOMB</button>
    </div>
  </div>
  <div class="toast" id="toast"><span id="toast-msg"></span><button class="toast-close" onclick="closeToast()">✕</button></div>
</div>

<script>
// ── State ──
let brainData = null;
let regionSpheres = {};
let selectedRegion = null;
let scene, camera, renderer, raycaster, mouse;
let particleSystem;
let axonLines = [];

// ── Colors per region ──
const regionColors = {
  brainstem: 0xff4444,
  limbic: 0xff8844,
  hippocampus: 0xffcc44,
  sensors: 0x44ff88,
  cortex: 0x4488ff,
  ego: 0xaa44ff,
  prefrontal: 0xff44aa
};
const regionEmoji = {
  brainstem: '🫀', limbic: '💓', hippocampus: '📝',
  sensors: '👁️', cortex: '🧠', ego: '🎭', prefrontal: '🎯'
};

// ── Three.js Setup ──
function initThree() {
  const container = document.getElementById('canvas3d');
  scene = new THREE.Scene();
  scene.fog = new THREE.FogExp2(0x09090b, 0.0015);

  camera = new THREE.PerspectiveCamera(60, container.clientWidth / container.clientHeight, 0.1, 2000);
  camera.position.set(0, 50, 250);
  camera.lookAt(0, 0, 0);

  renderer = new THREE.WebGLRenderer({ antialias: true, alpha: true });
  renderer.setSize(container.clientWidth, container.clientHeight);
  renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2));
  container.appendChild(renderer.domElement);

  // Lighting
  const ambient = new THREE.AmbientLight(0x222244, 0.5);
  scene.add(ambient);
  const point = new THREE.PointLight(0x4488ff, 1.5, 500);
  point.position.set(0, 100, 100);
  scene.add(point);
  const point2 = new THREE.PointLight(0xaa44ff, 0.8, 400);
  point2.position.set(-100, -50, -100);
  scene.add(point2);

  // Background particles
  const pGeo = new THREE.BufferGeometry();
  const pCount = 2000;
  const positions = new Float32Array(pCount * 3);
  for (let i = 0; i < pCount * 3; i++) positions[i] = (Math.random() - 0.5) * 800;
  pGeo.setAttribute('position', new THREE.BufferAttribute(positions, 3));
  const pMat = new THREE.PointsMaterial({ color: 0x334466, size: 1, transparent: true, opacity: 0.4 });
  particleSystem = new THREE.Points(pGeo, pMat);
  scene.add(particleSystem);

  raycaster = new THREE.Raycaster();
  mouse = new THREE.Vector2();

  // Events
  renderer.domElement.addEventListener('mousemove', onMouseMove);
  renderer.domElement.addEventListener('click', onSphereClick);
  window.addEventListener('resize', () => {
    camera.aspect = container.clientWidth / container.clientHeight;
    camera.updateProjectionMatrix();
    renderer.setSize(container.clientWidth, container.clientHeight);
  });
}

// ── Create Brain Spheres ──
function createBrain(data) {
  // Clear old
  Object.values(regionSpheres).forEach(s => scene.remove(s.mesh));
  axonLines.forEach(l => scene.remove(l));
  regionSpheres = {}; axonLines = [];

  const regions = data.regions || [];
  const angleStep = (Math.PI * 2) / Math.max(regions.length, 1);
  const radius = 100;

  regions.forEach((region, i) => {
    const angle = angleStep * i - Math.PI / 2;
    const x = Math.cos(angle) * radius;
    const z = Math.sin(angle) * radius;
    const y = (Math.random() - 0.5) * 40;

    const neuronCount = region.neurons ? region.neurons.length : 0;
    const totalAct = region.neurons ? region.neurons.reduce((s,n) => s + n.counter, 0) : 0;
    const sphereSize = 8 + Math.sqrt(neuronCount) * 3;

    const color = regionColors[region.name] || 0x888888;

    // Glowing sphere
    const geo = new THREE.SphereGeometry(sphereSize, 32, 32);
    const mat = new THREE.MeshPhongMaterial({
      color: color, emissive: color, emissiveIntensity: 0.3,
      transparent: true, opacity: 0.85, shininess: 80
    });
    const mesh = new THREE.Mesh(geo, mat);
    mesh.position.set(x, y, z);
    scene.add(mesh);

    // Glow ring & Pulse aura
    const ringGeo = new THREE.RingGeometry(sphereSize + 2, sphereSize + 6, 64);
    const ringMat = new THREE.MeshBasicMaterial({
      color: color, transparent: true, opacity: 0.25, side: THREE.DoubleSide, blending: THREE.AdditiveBlending
    });
    const ring = new THREE.Mesh(ringGeo, ringMat);
    ring.position.copy(mesh.position);
    ring.lookAt(camera.position);
    scene.add(ring);

    // Sub-neuron dots (orbit around sphere)
    const subDots = [];
    if (region.neurons) {
      const topN = [...region.neurons].sort((a,b) => b.counter - a.counter).slice(0, 12);
      topN.forEach((n, j) => {
        const subAngle = (Math.PI * 2 / topN.length) * j;
        const subR = sphereSize + 8 + Math.random() * 6;
        const dotGeo = new THREE.SphereGeometry(1.5 + Math.min(n.counter / 4, 4), 12, 12);
        const dotMat = new THREE.MeshBasicMaterial({
          color: n.dopamine > 0 ? 0x659b85 : n.hasBomb ? 0xc96d63 : 0xb6cfdd,
          transparent: true, opacity: 0.85, blending: THREE.AdditiveBlending
        });
        const dot = new THREE.Mesh(dotGeo, dotMat);
        dot.userData = { isNeuron: true, path: n.path, counter: n.counter, region: region.name };
        dot.position.set(
          x + Math.cos(subAngle) * subR,
          y + Math.sin(subAngle) * subR * 0.6,
          z + Math.sin(subAngle * 1.5) * subR * 0.4
        );
        scene.add(dot);
        subDots.push(dot);
      });
    }

    regionSpheres[region.name] = {
      mesh, ring, subDots, region,
      basePos: { x, y, z }, sphereSize,
      neuronCount, totalAct
    };
  });

  // Axon connections (lines between regions)
  regions.forEach(region => {
    if (!region.axons) return;
    const src = regionSpheres[region.name];
    if (!src) return;
    region.axons.forEach(target => {
      const dst = regionSpheres[target];
      if (!dst) return;
      const curve = new THREE.QuadraticBezierCurve3(
        src.mesh.position,
        new THREE.Vector3(
          (src.mesh.position.x + dst.mesh.position.x) / 2,
          (src.mesh.position.y + dst.mesh.position.y) / 2 + 30,
          (src.mesh.position.z + dst.mesh.position.z) / 2
        ),
        dst.mesh.position
      );
      const points = curve.getPoints(30);
      const lineGeo = new THREE.BufferGeometry().setFromPoints(points);
      const lineMat = new THREE.LineBasicMaterial({
        color: regionColors[region.name] || 0xeadccf,
        transparent: true, opacity: 0.25, blending: THREE.AdditiveBlending
      });
      const line = new THREE.Line(lineGeo, lineMat);
      scene.add(line);
      axonLines.push(line);
    });
  });
}

// ── Mouse interaction ──
function onMouseMove(e) {
  const rect = renderer.domElement.getBoundingClientRect();
  mouse.x = ((e.clientX - rect.left) / rect.width) * 2 - 1;
  mouse.y = -((e.clientY - rect.top) / rect.height) * 2 + 1;

  raycaster.setFromCamera(mouse, camera);
  
  let allMeshes = [];
  Object.values(regionSpheres).forEach(s => {
    allMeshes.push(s.mesh);
    allMeshes.push(...s.subDots);
  });
  
  const hits = raycaster.intersectObjects(allMeshes);

  const tooltip = document.getElementById('tooltip');
  if (hits.length > 0) {
    const hitObj = hits[0].object;
    tooltip.style.display = 'block';
    tooltip.style.left = (e.clientX - document.getElementById('canvas3d').getBoundingClientRect().left + 15) + 'px';
    tooltip.style.top = (e.clientY - document.getElementById('canvas3d').getBoundingClientRect().top - 10) + 'px';
    tooltip.style.opacity = '1';

    if (hitObj.userData && hitObj.userData.isNeuron) {
      let dispPath = hitObj.userData.path.replace(/\\/g, "/");
      tooltip.innerHTML = '<span class="tt-region">🧠 ' + hitObj.userData.region + '/' + dispPath.split('/').pop() + '</span>' +
        '<span class="tt-stats">활성도(Synapse): ' + hitObj.userData.counter + '</span>';
      renderer.domElement.style.cursor = 'crosshair';
      return;
    }

    const entry = Object.values(regionSpheres).find(s => s.mesh === hitObj);
    if (entry) {
      tooltip.innerHTML = '<span class="tt-region">' + (regionEmoji[entry.region.name]||'') + ' ' + entry.region.name + '</span>' +
        '<span class="tt-stats">뉴런 ' + entry.neuronCount + ' | 활성도 ' + entry.totalAct + '</span>';
      renderer.domElement.style.cursor = 'pointer';
    }
  } else {
    tooltip.style.opacity = '0';
    setTimeout(() => { if(tooltip.style.opacity==='0') tooltip.style.display = 'none'; }, 200);
    renderer.domElement.style.cursor = 'default';
  }
}

function onSphereClick(e) {
  raycaster.setFromCamera(mouse, camera);
  const meshes = Object.values(regionSpheres).map(s => s.mesh);
  const hits = raycaster.intersectObjects(meshes);
  if (hits.length > 0) {
    const entry = Object.values(regionSpheres).find(s => s.mesh === hits[0].object);
    if (entry) selectRegion(entry.region.name);
  }
}

// ── Select Region (show detail) ──
function selectRegion(name) {
  selectedRegion = name;
  const entry = regionSpheres[name];
  if (!entry) return;
  const region = entry.region;

  // Highlight sphere
  Object.values(regionSpheres).forEach(s => {
    s.mesh.material.emissiveIntensity = s.mesh === entry.mesh ? 0.8 : 0.15;
    s.mesh.material.opacity = s.mesh === entry.mesh ? 1 : 0.4;
    s.ring.material.opacity = s.mesh === entry.mesh ? 0.3 : 0.05;
  });
  // Highlight axon lines
  axonLines.forEach(l => { l.material.opacity = 0.05; });
  if (region.axons) {
    axonLines.forEach(l => { /* highlight relevant */ l.material.opacity = 0.3; });
  }

  // Detail panel
  const panel = document.getElementById('detail');
  panel.classList.add('active');
  document.getElementById('detail-icon').textContent = regionEmoji[name] || '📁';
  document.getElementById('detail-name').textContent = name.toUpperCase() + ' — ' + (region.ko || '');

  // Axons
  let axonHtml = '';
  if (region.axons && region.axons.length > 0) {
    axonHtml += '<div style="font-size:10px;color:#666;margin-bottom:4px;">축삭 연결:</div>';
    region.axons.forEach(a => {
      axonHtml += '<div class="conn-line"><span class="conn-arrow">→</span><span class="conn-target" onclick="selectRegion(\'' + a + '\')">' +
        (regionEmoji[a]||'') + ' ' + a + '</span></div>';
    });
  }
  document.getElementById('detail-axons').innerHTML = axonHtml;

  // Neurons
  let nHtml = '';
  if (region.neurons) {
    const sorted = [...region.neurons].sort((a,b) => (b.intensity||b.counter) - (a.intensity||a.counter));
    sorted.forEach(n => {
      const intensity = n.intensity || (n.counter + (n.dopamine||0));
      const pct = Math.min(100, intensity * 4);
      const pol = n.polarity !== undefined ? n.polarity : 0.5;
      let barColor;
      if (pol > 0.6) barColor = '#22c55e';
      else if (pol < 0.3) barColor = '#ef4444';
      else if (intensity >= 10) barColor = '#f59e0b';
      else barColor = '#3b82f6';

      let signals = '';
      if (n.dopamine > 0) signals += '🟢';
      if (n.hasBomb) signals += '💀';
      if (n.isDormant) signals += '💤';
      if (pol > 0.6) signals += '↑';
      else if (pol < 0.3 && intensity > 3) signals += '↓';

      let strengthHtml = '';
      if (intensity >= 10) strengthHtml = '<span class="n-strength n-str-abs">절대</span>';
      else if (intensity >= 5) strengthHtml = '<span class="n-strength n-str-must">반드시</span>';

      const path = n.path.replace(/\//g, ' > ');
      const regionName = selectedRegion;
      nHtml += '<div class="neuron-item">' +
        strengthHtml +
        '<div class="n-name">' + path + '</div>' +
        '<div class="n-bar"><div class="n-fill" style="width:' + pct + '%;background:' + barColor + '"></div></div>' +
        '<div class="n-counter">' + intensity + '</div>' +
        '<div class="n-signals">' + signals + '</div>' +
        '<button class="n-fire" onclick="event.stopPropagation();fireNeuron(\'' + regionName + '\',\'' + n.path.replace(/'/g,"\\'" ) + '\')">▲</button>' +
        '</div>';
    });
  }
  document.getElementById('detail-neurons').innerHTML = nHtml;

  // Region chips
  document.querySelectorAll('.region-chip').forEach(c => {
    c.classList.toggle('active', c.dataset.region === name);
  });
}

function closeDetail() {
  document.getElementById('detail').classList.remove('active');
  selectedRegion = null;
  Object.values(regionSpheres).forEach(s => {
    s.mesh.material.emissiveIntensity = 0.3;
    s.mesh.material.opacity = 0.85;
    s.ring.material.opacity = 0.25;
  });
  axonLines.forEach(l => { l.material.opacity = 0.15; });
  document.querySelectorAll('.region-chip').forEach(c => c.classList.remove('active'));
}

// ── Render loop ──
let frame = 0;
function animate() {
  requestAnimationFrame(animate);
  frame++;

  // Slow camera orbit (Dramatic close angle)
  const t = frame * 0.001;
  camera.position.x = Math.cos(t) * 200;
  camera.position.z = Math.sin(t) * 200;
  camera.position.y = 80 + Math.sin(t * 0.5) * 40;
  camera.lookAt(0, 0, 0);

  // Organic Breathing spheres & pulsing rings with complex synthetic waves
  Object.values(regionSpheres).forEach(s => {
    // Neural rhythm (slow + fast variation)
    const rhythm = Math.sin(frame * 0.03 + s.basePos.x) + Math.cos(frame * 0.015 + s.basePos.z) * 0.5;
    const breath = 1 + rhythm * 0.04;
    s.mesh.scale.set(breath, breath, breath);
    
    // Ring pulse (ethereal glow fluctuation)
    const ringScale = 1 + (Math.sin(frame * 0.05 + s.basePos.y) + Math.cos(frame * 0.025)) * 0.15;
    s.ring.scale.set(ringScale, ringScale, 1);
    s.ring.material.opacity = 0.25 + (Math.sin(frame * 0.08 + s.basePos.z) * 0.5 + 0.5) * 0.1;
    s.ring.lookAt(camera.position);

    // Orbit sub-dots (chaotic orbital pull)
    s.subDots.forEach((dot, i) => {
      const a = frame * 0.004 + i * 0.6 + Math.sin(frame * 0.002) * 0.5;
      const r = s.sphereSize + 12 + i * 2.5 + Math.cos(frame * 0.01 + i) * 2;
      dot.position.x = s.basePos.x + Math.cos(a) * r;
      dot.position.y = s.basePos.y + Math.sin(a * 0.8) * r * 0.4;
      dot.position.z = s.basePos.z + Math.sin(a * 1.2) * r * 0.4;
    });
  });

  // Axon breathing (fluctuating data stream)
  axonLines.forEach((l, i) => {
    l.material.opacity = 0.15 + (Math.sin(frame * 0.04 + i) + Math.cos(frame * 0.01)) * 0.1;
  });

  // Rotate particles
  particleSystem.rotation.y += 0.0001;

  renderer.render(scene, camera);
}

// ── UI Updates ──
function updateUI(data) {
  document.getElementById('s-neurons').textContent = data.totalNeurons;
  document.getElementById('s-activation').textContent = data.totalCounter;
  document.getElementById('s-regions').textContent = (data.regions || []).length;
  document.getElementById('score').textContent = '⚡ ' + data.totalCounter;

  if (data.bombSource) {
    document.getElementById('status').className = 'badge';
    document.getElementById('status').style.background = '#7f1d1d';
    document.getElementById('status').style.color = '#fca5a5';
    document.getElementById('status').textContent = '💀 BOMB: ' + data.bombSource;
  } else {
    document.getElementById('status').className = 'badge badge-ok';
    document.getElementById('status').textContent = 'NOMINAL';
  }

  // Region chips
  let chipsHtml = '';
  (data.regions || []).forEach(r => {
    const nc = r.neurons ? r.neurons.length : 0;
    const act = selectedRegion === r.name ? ' active' : '';
    chipsHtml += '<span class="region-chip' + act + '" data-region="' + r.name + '" onclick="selectRegion(\'' + r.name + '\')">' +
      (regionEmoji[r.name]||'') + ' ' + r.name +
      '<span class="chip-count">' + nc + '</span></span>';
  });
  document.getElementById('regionChips').innerHTML = chipsHtml;
  updateBombDropdown();
}

// ── API ──
let previousNeuronCount = 0;

async function loadBrain() {
  try {
    const res = await fetch('/api/brain');
    brainData = await res.json();
    if (!scene) { initThree(); animate(); }
    createBrain(brainData);
    updateUI(brainData);
    if (selectedRegion) selectRegion(selectedRegion);
    
    // Auto-Evolution Detection
    if (previousNeuronCount > 0 && brainData.totalNeurons > previousNeuronCount) {
        const diff = brainData.totalNeurons - previousNeuronCount;
        showToast('🌱 자가 진화 발생: +' + diff + ' 신규 규칙 흡수완료');
    }
    previousNeuronCount = brainData.totalNeurons;
  } catch(e) { /* silent — 禁console_log */ }
}

// ── Inject ──
async function doInject() {
  showToast('⚡ Injecting...');
  const res = await fetch('/api/inject', { method: 'POST' });
  const text = await res.text();
  showToast('✅ ' + text);
  loadBrain();
}
async function doDedup() {
  showToast('🔀 Dedup...');
  await fetch('/api/dedup', { method: 'POST' });
  showToast('✅ Dedup 완료');
  loadBrain();
}
function updateBombDropdown() {
  const sel = document.getElementById('bombRegion');
  if (!sel || !brainData) return;
  const current = sel.value;
  sel.innerHTML = '<option value="">💀 영역 선택</option>';
  (brainData.regions || []).forEach(r => {
    const opt = document.createElement('option');
    opt.value = r.name; opt.textContent = (regionEmoji[r.name]||'') + ' ' + r.name;
    sel.appendChild(opt);
  });
  if (current) sel.value = current;
}
async function doBomb() {
  const region = document.getElementById('bombRegion').value;
  if (!region) { showToast('⚠️ 영역을 먼저 선택하세요'); return; }
  await fetch('/api/signal', {
    method: 'POST', headers:{'Content-Type':'application/json'},
    body: JSON.stringify({path: region + '/halt', type: 'bomb'})
  });
  showToast('💀 BOMB: ' + region);
  document.getElementById('bombRegion').value = '';
  loadBrain();
}

// ── Fire neuron ──
async function fireNeuron(region, path) {
  await fetch('/api/fire', {
    method: 'POST', headers:{'Content-Type':'application/json'},
    body: JSON.stringify({path: region + '/' + path})
  });
  showToast('🔥 fired: ' + path);
  loadBrain();
}

// ── Add neuron ──
function toggleAdd() {
  const section = document.getElementById('addSection');
  section.classList.toggle('visible');
  if (section.classList.contains('visible')) {
    const sel = document.getElementById('addRegion');
    sel.innerHTML = '';
    if (brainData && brainData.regions) {
      brainData.regions.forEach(r => {
        const opt = document.createElement('option');
        opt.value = r.name; opt.textContent = (regionEmoji[r.name]||'') + ' ' + r.name;
        sel.appendChild(opt);
      });
    }
  }
}
async function addNeuron() {
  const region = document.getElementById('addRegion').value;
  const path = document.getElementById('addPath').value.trim();
  if (!region || !path) return;
  await fetch('/api/neuron', {
    method: 'POST', headers:{'Content-Type':'application/json'},
    body: JSON.stringify({region: region, path: path})
  });
  document.getElementById('addPath').value = '';
  document.getElementById('addSection').classList.remove('visible');
  showToast('🌱 뉴런 생성: ' + region + '/' + path);
  loadBrain();
}

// ── Search ──
function filterNeurons() {
  const q = document.getElementById('searchInput').value.toLowerCase();
  if (!q) { if (selectedRegion) selectRegion(selectedRegion); return; }
  if (!brainData) return;

  // Search all neurons across all regions
  let results = [];
  (brainData.regions || []).forEach(r => {
    (r.neurons || []).forEach(n => {
      if (n.path.toLowerCase().includes(q) || n.name.toLowerCase().includes(q)) {
        results.push({...n, regionName: r.name});
      }
    });
  });
  results.sort((a,b) => b.counter - a.counter);

  let nHtml = '<div style="font-size:10px;color:#666;margin-bottom:6px;">검색 결과: ' + results.length + '건</div>';
  results.forEach(n => {
    const pct = Math.min(100, n.counter * 5);
    let barColor = '#475569';
    if (n.counter >= 10) barColor = '#f59e0b';
    else if (n.counter >= 5) barColor = '#22c55e';
    else if (n.counter >= 2) barColor = '#3b82f6';
    let strengthHtml = '';
    if (n.counter >= 10) strengthHtml = '<span class="n-strength n-str-abs">절대</span>';
    else if (n.counter >= 5) strengthHtml = '<span class="n-strength n-str-must">반드시</span>';
    let signals = '';
    if (n.dopamine > 0) signals += '🟢';
    if (n.hasBomb) signals += '💀';
    nHtml += '<div class="neuron-item">' +
      strengthHtml +
      '<div class="n-name" style="font-size:10px"><span style="color:#3b82f6">' + n.regionName + '</span> > ' + n.path.replace(/\//g, ' > ') + '</div>' +
      '<div class="n-bar"><div class="n-fill" style="width:' + pct + '%;background:' + barColor + '"></div></div>' +
      '<div class="n-counter">' + n.counter + '</div>' +
      '<div class="n-signals">' + signals + '</div>' +
      '</div>';
  });

  const panel = document.getElementById('detail');
  panel.classList.add('active');
  document.getElementById('detail-icon').textContent = '🔍';
  document.getElementById('detail-name').textContent = '검색: "' + q + '"';
  document.getElementById('detail-axons').innerHTML = '';
  document.getElementById('detail-neurons').innerHTML = nHtml;
}

// ── Sandbox ──
function toggleSandbox() {
  const section = document.getElementById('sandboxSection');
  section.classList.toggle('visible');
  if (section.classList.contains('visible')) {
    // Load current sandbox rules
    fetch('/api/sandbox').then(r => r.json()).then(data => {
      document.getElementById('sandboxText').value = (data.rules || []).join('\n');
    }).catch(() => {});
  }
}
async function applySandbox() {
  const text = document.getElementById('sandboxText').value.trim();
  if (!text) { showToast('⚠️ 규칙을 입력하세요'); return; }
  const res = await fetch('/api/sandbox', {
    method: 'POST', headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({text: text})
  });
  const data = await res.json();
  const pathsDiv = document.getElementById('sandboxPaths');
  if (data.paths && data.paths.length > 0) {
    pathsDiv.innerHTML = data.paths.map(p => '<div>✓ ' + p + ' 생성됨</div>').join('');
  } else {
    pathsDiv.innerHTML = '<div>✓ ' + (data.created || 0) + '개 적용</div>';
  }
  showToast('🧪 Sandbox: ' + (data.created || 0) + '개 규칙 적용');
  loadBrain();
}
async function clearSandbox() {
  await fetch('/api/sandbox', {
    method: 'POST', headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({text: ''})
  });
  document.getElementById('sandboxText').value = '';
  document.getElementById('sandboxPaths').innerHTML = '';
  showToast('🗑 Sandbox 초기화');
  loadBrain();
}

let toastTimeout;
function showToast(msg) {
  const t = document.getElementById('toast');
  document.getElementById('toast-msg').textContent = msg;
  if (msg.includes('⚠️') || msg.includes('💀')) t.classList.add('error'); else t.classList.remove('error');
  t.classList.add('show');
  clearTimeout(toastTimeout);
  toastTimeout = setTimeout(() => t.classList.remove('show'), 5000);
}
function closeToast() {
  document.getElementById('toast').classList.remove('show');
}

// ── Health monitoring ──
async function loadHealth() {
  try {
    const res = await fetch('/api/health');
    const data = await res.json();
    const list = document.getElementById('healthList');
    const badge = document.getElementById('health-badge');
    const meta = document.getElementById('healthMeta');
    if (!list) return;

    let html = '';
    let allOk = true;
    (data.processes || []).forEach(p => {
      if (!p.running) allOk = false;
      html += '<div class="health-row">' +
        '<div class="health-dot ' + (p.running ? 'on' : 'off') + '"></div>' +
        '<span class="health-name">' + p.name + '</span>' +
        '<span class="health-role">' + p.role + '</span>' +
        '</div>';
    });
    list.innerHTML = html;

    if (allOk) {
      badge.textContent = 'ALL OK';
      badge.style.background = '#064e3b';
      badge.style.color = '#34d399';
    } else {
      badge.textContent = 'DEGRADED';
      badge.style.background = '#7f1d1d';
      badge.style.color = '#fca5a5';
    }

    meta.textContent = 'OS: ' + data.os + ' | .neuron files: ' + data.neuronFiles;
  } catch(e) { /* silent */ }
}

// ── Init ──
loadBrain();
loadHealth();
setInterval(loadBrain, 10000);
setInterval(loadHealth, 15000);

// ── Keyboard shortcuts ──
document.addEventListener('keydown', (e) => {
  if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
    e.preventDefault();
    const input = document.getElementById('searchInput');
    input.focus();
    input.select();
  }
  if (e.key === 'Escape') {
    const input = document.getElementById('searchInput');
    if (document.activeElement === input) {
      input.value = '';
      input.blur();
      closeDetail();
    }
  }
});
</script>
</body>
</html>` + "\n"

