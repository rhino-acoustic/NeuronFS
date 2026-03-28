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
    position: fixed; bottom: 80px; left: 50%; transform: translateX(-50%);
    background: rgba(0,0,0,0.9); border: 1px solid #333;
    border-radius: 8px; padding: 8px 20px; font-size: 12px; color: #34d399;
    z-index: 200; opacity: 0; transition: opacity 0.3s;
    pointer-events: none;
  }
  .toast.show { opacity: 1; }

  /* ── 3D overlay info ── */
  .hover-tooltip {
    position: absolute; pointer-events: none;
    background: rgba(0,0,0,0.85); backdrop-filter: blur(12px);
    border: 1px solid #333; border-radius: 8px;
    padding: 8px 14px; font-size: 11px; color: #fff;
    display: none; z-index: 100; max-width: 280px;
  }
  .hover-tooltip .tt-region { color: #3b82f6; font-weight: 700; }
  .hover-tooltip .tt-stats { color: #888; font-size: 10px; }
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
      <input type="text" id="searchInput" placeholder="뉴런 검색 (경로, 이름...)" oninput="filterNeurons()">
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

    <div class="controls">
      <button class="btn btn-primary" onclick="doInject()">⚡ INJECT</button>
      <button class="btn btn-add" onclick="toggleAdd()">+ 뉴런</button>
      <button class="btn btn-ghost" onclick="doDedup()">🔀 DEDUP</button>
      <button class="btn btn-danger" onclick="doBomb()">💀 BOMB</button>
    </div>
  </div>
  <div class="toast" id="toast"></div>
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

    // Glow ring
    const ringGeo = new THREE.RingGeometry(sphereSize + 2, sphereSize + 4, 32);
    const ringMat = new THREE.MeshBasicMaterial({
      color: color, transparent: true, opacity: 0.15, side: THREE.DoubleSide
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
        const dotGeo = new THREE.SphereGeometry(1 + Math.min(n.counter / 5, 3), 8, 8);
        const dotMat = new THREE.MeshBasicMaterial({
          color: n.dopamine > 0 ? 0x22ff66 : n.hasBomb ? 0xff2222 : color,
          transparent: true, opacity: 0.7
        });
        const dot = new THREE.Mesh(dotGeo, dotMat);
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
        color: regionColors[region.name] || 0x444444,
        transparent: true, opacity: 0.15
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
  const meshes = Object.values(regionSpheres).map(s => s.mesh);
  const hits = raycaster.intersectObjects(meshes);

  const tooltip = document.getElementById('tooltip');
  if (hits.length > 0) {
    const entry = Object.values(regionSpheres).find(s => s.mesh === hits[0].object);
    if (entry) {
      tooltip.style.display = 'block';
      tooltip.style.left = (e.clientX - document.getElementById('canvas3d').getBoundingClientRect().left + 15) + 'px';
      tooltip.style.top = (e.clientY - document.getElementById('canvas3d').getBoundingClientRect().top - 10) + 'px';
      tooltip.innerHTML = '<span class="tt-region">' + (regionEmoji[entry.region.name]||'') + ' ' + entry.region.name + '</span><br>' +
        '<span class="tt-stats">뉴런 ' + entry.neuronCount + ' | 활성도 ' + entry.totalAct + '</span>';
      renderer.domElement.style.cursor = 'pointer';
    }
  } else {
    tooltip.style.display = 'none';
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
    const sorted = [...region.neurons].sort((a,b) => b.counter - a.counter);
    sorted.forEach(n => {
      const pct = Math.min(100, n.counter * 5);
      let barColor = '#475569';
      if (n.counter >= 10) barColor = '#f59e0b';
      else if (n.counter >= 5) barColor = '#22c55e';
      else if (n.counter >= 2) barColor = '#3b82f6';

      let signals = '';
      if (n.dopamine > 0) signals += '🟢';
      if (n.hasBomb) signals += '💀';
      if (n.isDormant) signals += '💤';

      let strengthHtml = '';
      if (n.counter >= 10) strengthHtml = '<span class="n-strength n-str-abs">절대</span>';
      else if (n.counter >= 5) strengthHtml = '<span class="n-strength n-str-must">반드시</span>';

      const path = n.path.replace(/\//g, ' > ');
      const regionName = selectedRegion;
      nHtml += '<div class="neuron-item">' +
        strengthHtml +
        '<div class="n-name">' + path + '</div>' +
        '<div class="n-bar"><div class="n-fill" style="width:' + pct + '%;background:' + barColor + '"></div></div>' +
        '<div class="n-counter">' + n.counter + '</div>' +
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
    s.ring.material.opacity = 0.15;
  });
  axonLines.forEach(l => { l.material.opacity = 0.15; });
  document.querySelectorAll('.region-chip').forEach(c => c.classList.remove('active'));
}

// ── Render loop ──
let frame = 0;
function animate() {
  requestAnimationFrame(animate);
  frame++;

  // Slow camera orbit
  const t = frame * 0.001;
  camera.position.x = Math.cos(t) * 250;
  camera.position.z = Math.sin(t) * 250;
  camera.position.y = 50 + Math.sin(t * 0.5) * 20;
  camera.lookAt(0, 0, 0);

  // Breathing spheres
  Object.values(regionSpheres).forEach(s => {
    const breath = 1 + Math.sin(frame * 0.02 + s.basePos.x) * 0.03;
    s.mesh.scale.set(breath, breath, breath);
    s.ring.lookAt(camera.position);
    // Orbit sub-dots
    s.subDots.forEach((dot, i) => {
      const a = frame * 0.005 + i * 0.5;
      const r = s.sphereSize + 10 + i * 2;
      dot.position.x = s.basePos.x + Math.cos(a) * r;
      dot.position.y = s.basePos.y + Math.sin(a * 0.7) * r * 0.5;
      dot.position.z = s.basePos.z + Math.sin(a * 1.3) * r * 0.3;
    });
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
}

// ── API ──
async function loadBrain() {
  try {
    const res = await fetch('/api/brain');
    brainData = await res.json();
    if (!scene) { initThree(); animate(); }
    createBrain(brainData);
    updateUI(brainData);
    if (selectedRegion) selectRegion(selectedRegion);
  } catch(e) { console.error(e); }
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
async function doBomb() {
  const region = prompt('Bomb 영역:');
  if (!region) return;
  await fetch('/api/signal', {
    method: 'POST', headers:{'Content-Type':'application/json'},
    body: JSON.stringify({path: region + '/halt', type: 'bomb'})
  });
  showToast('💀 BOMB: ' + region);
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

// ── Toast ──
function showToast(msg) {
  const t = document.getElementById('toast');
  t.textContent = msg;
  t.classList.add('show');
  setTimeout(() => t.classList.remove('show'), 2000);
}

// ── Init ──
loadBrain();
setInterval(loadBrain, 10000);
</script>
</body>
</html>` + "\n"

