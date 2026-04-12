'use client';

import React, { useMemo, useRef, useEffect, useState } from 'react';
import { Canvas, useFrame } from '@react-three/fiber';
import { OrbitControls } from '@react-three/drei';
import * as THREE from 'three';
import { forceSimulation, forceLink, forceManyBody, forceCenter } from 'd3-force-3d';

// Data fetching and Graph initialization
function useBrainGraph() {
  const [graphData, setGraphData] = useState<{ nodes: any[]; links: any[] } | null>(null);

  useEffect(() => {
    fetch('http://localhost:7350/api/brain')
      .then(r => r.json())
      .then(data => {
        const nodes: any[] = [];
        const links: any[] = [];
        const parentMap = new Map<string, string>();

        data.regions?.forEach((region: any, rIdx: number) => {
          region.neurons?.forEach((neuron: any) => {
            const pathParts = neuron.path.replace(/\\/g, '/').split('/');
            const id = neuron.path.replace(/\\/g, '/');
            nodes.push({
              id,
              region: region.name,
              counter: neuron.counter || 0,
              yPriority: (Object.values(data.regions).length - rIdx) * 50 // pseudo-gravity
            });
            
            // Reconstruct tree links
            if (pathParts.length > 1) {
              const parentPath = pathParts.slice(0, -1).join('/');
              if (!parentMap.has(id)) {
                links.push({ source: parentPath, target: id });
              }
            }
            // Ensure parent folders exist as empty nodes if they aren't explicitly listed
            pathParts.reduce((acc: string, curr: string) => {
              const full = acc ? `${acc}/${curr}` : curr;
              if (full !== id && !nodes.find(n => n.id === full)) {
                nodes.push({ id: full, region: region.name, counter: 1, isFolder: true });
              }
              return full;
            }, '');
          });
        });

        // Filter valid links where both source & target exist
        const validLinks = links.filter(l => nodes.find(n => n.id === l.source) && nodes.find(n => n.id === l.target));

        setGraphData({ nodes, links: validLinks });
      })
      .catch(err => console.error("Could not fetch brain:", err));
  }, []);

  return graphData;
}

function InstancedGraph({ nodes, links }: { nodes: any[]; links: any[] }) {
  const meshRef = useRef<THREE.InstancedMesh>(null);
  const colorArray = useMemo(() => new Float32Array(nodes.length * 3), [nodes]);

  // Initiate Simulation
  const simulation = useMemo(() => {
    nodes.forEach(n => {
      n.x = (Math.random() - 0.5) * 100;
      n.y = (Math.random() - 0.5) * 100;
      n.z = (Math.random() - 0.5) * 100;
    });

    const sim = forceSimulation(nodes, 3)
      .force('link', forceLink(links).id((d: any) => d.id).distance(20))
      .force('charge', forceManyBody().strength(-80))
      .force('center', forceCenter(0, 0, 0));

    // Warmup
    sim.tick(100);
    return sim;
  }, [nodes, links]);

  // Initialize Colors
  useMemo(() => {
    const dummyColor = new THREE.Color();
    nodes.forEach((n, i) => {
      if (n.region === "cortex") dummyColor.setHex(0x3b82f6);
      else if (n.region === "limbic") dummyColor.setHex(0x10b981);
      else if (n.region === "prefrontal") dummyColor.setHex(0x8b5cf6);
      else if (n.region === "brainstem") dummyColor.setHex(0xdc2626);
      else dummyColor.setHex(0xa1a1aa);

      if (n.isFolder) dummyColor.setHex(0x27272a); // dark gray
      
      dummyColor.toArray(colorArray, i * 3);
    });
  }, [nodes, colorArray]);

  const [dragSource, setDragSource] = useState<number | null>(null);
  const dummy = useMemo(() => new THREE.Object3D(), []);

  useFrame(() => {
    if (!meshRef.current) return;
    
    // Slight simulation tick for living effect
    simulation.tick(1);

    nodes.forEach((node, i) => {
      dummy.position.set(node.x, node.y, node.z);
      
      // Highlight dragged source
      const s = dragSource === i ? 2.5 : Math.max(0.5, Math.log10(node.counter || 1) + 0.5);
      dummy.scale.set(s, s, s);
      dummy.updateMatrix();
      meshRef.current!.setMatrixAt(i, dummy.matrix);
    });
    
    meshRef.current.instanceMatrix.needsUpdate = true;
  });

  const handlePointerDown = (e: any) => {
    e.stopPropagation();
    setDragSource(e.instanceId);
  };

  const handlePointerUp = (e: any) => {
    e.stopPropagation();
    if (dragSource !== null && dragSource !== e.instanceId && e.instanceId !== undefined) {
      const sourceId = nodes[dragSource].id;
      const targetId = nodes[e.instanceId].id;
      console.log(`Merge/Link Drag trigger: ${sourceId} -> ${targetId}`);
      
      const ws = (window as any).NeuronWS;
      if (ws?.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({
          action: "merge", // Default action for Phase 21 drag
          source: sourceId,
          target: targetId
        }));
      }
    }
    setDragSource(null);
  };

  return (
    <instancedMesh
      ref={meshRef}
      args={[undefined, undefined, nodes.length]}
      frustumCulled={false}
      onPointerDown={handlePointerDown}
      onPointerUp={handlePointerUp}
      onPointerMissed={() => setDragSource(null)}
    >
      <octahedronGeometry args={[0.8, 0]}>
        <instancedBufferAttribute attach="attributes-color" args={[colorArray, 3]} />
      </octahedronGeometry>
      <meshPhongMaterial
        vertexColors
        transparent
        opacity={0.8}
        shininess={100}
        blending={THREE.AdditiveBlending}
      />
    </instancedMesh>
  );
}

function SceneLighting() {
  const pLightRef = useRef<THREE.PointLight>(null);

  useEffect(() => {
    const handleFire = () => {
      if (pLightRef.current) {
        pLightRef.current.intensity = 5;
        pLightRef.current.color.setHex(0x10b981);
      }
    };
    window.addEventListener('neuron-fire', handleFire);
    return () => window.removeEventListener('neuron-fire', handleFire);
  }, []);

  useFrame((state, delta) => {
    if (pLightRef.current && pLightRef.current.intensity > 1) {
      pLightRef.current.intensity = Math.max(1, pLightRef.current.intensity - delta * 4);
      if (pLightRef.current.intensity <= 1) {
        pLightRef.current.color.setHex(0x3b82f6); 
      }
    }
  });

  return (
    <>
      <ambientLight color={0x10b981} intensity={0.3} />
      <pointLight ref={pLightRef} color={0x3b82f6} intensity={1} distance={500} position={[0, 80, 0]} />
      <fogExp2 attach="fog" args={[0x030305, 0.001]} />
    </>
  );
}

export default function BrainScene() {
  const data = useBrainGraph();

  return (
    <div className="w-full h-full relative" style={{ background: 'radial-gradient(circle at center, #0a0a14 0%, #030305 100%)' }}>
      {!data ? (
        <div className="absolute inset-0 flex items-center justify-center text-[var(--accent)] font-mono animate-pulse z-10 text-xs">
          FETCHING COGNITIVE TOPOLOGY...
        </div>
      ) : null}
      
      {data && (
        <Canvas
          camera={{ position: [150, 100, 200], fov: 50 }}
          gl={{ antialias: true, alpha: true }}
        >
          <SceneLighting />
          <InstancedGraph nodes={data.nodes} links={data.links} />
          <OrbitControls 
            enableDamping 
            dampingFactor={0.05} 
            // autoRotate 
            // autoRotateSpeed={0.5} 
          />
        </Canvas>
      )}
      
      {data && (
        <div className="absolute bottom-5 left-5 font-mono text-[10px] text-slate-400 pointer-events-none z-20">
          Force-Directed Graph | {data.nodes.length.toLocaleString()} Nodes | {data.links.length.toLocaleString()} Edges
        </div>
      )}
    </div>
  );
}
