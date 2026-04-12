'use client';

import React, { useMemo, useRef, useEffect } from 'react';
import { Canvas, useFrame } from '@react-three/fiber';
import { OrbitControls } from '@react-three/drei';
import * as THREE from 'three';

const NEURON_COUNT = 30000;

function InstancedBrain() {
  const meshRef = useRef<THREE.InstancedMesh>(null);
  
  // Calculate positions only once
  const { positions, colors, matrices } = useMemo(() => {
    const dummy = new THREE.Object3D();
    const tempPositions = [];
    const tempColors = new Float32Array(NEURON_COUNT * 3);
    const tempMatrices = new Float32Array(NEURON_COUNT * 16);
    const color = new THREE.Color();

    for (let i = 0; i < NEURON_COUNT; i++) {
      const t = Math.random() * Math.PI * 2;
      const p = Math.acos(2 * Math.random() - 1);
      
      let rx=55, ry=42, rz=48;
      const f=1+0.2*Math.pow(Math.max(0,Math.sin(t)*Math.cos(p)),2);
      const tm=1+0.12*Math.pow(Math.sin(t),2)*Math.pow(Math.sin(p),2);
      const bt=t>Math.PI*0.75 ? 1-0.3*(t-Math.PI*0.75)/(Math.PI*0.25) : 1;
      const gr=1-0.08*Math.pow(Math.max(0,Math.cos(p)),4)*Math.pow(Math.sin(t),2);
      const r=f*tm*bt*gr * (0.5 + Math.random()*0.5); 
      
      const x = rx*r*Math.sin(t)*Math.cos(p);
      const y = ry*r*Math.cos(t);
      const z = rz*r*Math.sin(t)*Math.sin(p);

      tempPositions.push(new THREE.Vector3(x, y, z));
      dummy.position.set(x, y, z);
      
      const dist = Math.sqrt(x*x + y*y + z*z);
      const s = Math.max(0.2, 1 - (dist / 60));
      dummy.scale.set(s, s, s);
      dummy.updateMatrix();
      
      dummy.matrix.toArray(tempMatrices, i * 16);

      if(y > 10) color.setHex(0x3b82f6); 
      else if(dist < 20) color.setHex(0xdc2626); 
      else if(y < -20) color.setHex(0x10b981);
      else color.setHex(0xa1a1aa);
      
      color.toArray(tempColors, i * 3);
    }
    
    return { positions: tempPositions, colors: tempColors, matrices: tempMatrices };
  }, []);

  useEffect(() => {
    if (meshRef.current) {
      meshRef.current.instanceMatrix.needsUpdate = true;
      if (meshRef.current.instanceColor) {
        meshRef.current.instanceColor.needsUpdate = true;
      }
    }
  }, []);

  const [dragSource, setDragSource] = React.useState<number | null>(null);

  const handlePointerDown = (e: any) => {
    e.stopPropagation();
    setDragSource(e.instanceId);
  };

  const handlePointerUp = (e: any) => {
    e.stopPropagation();
    if (dragSource !== null && dragSource !== e.instanceId && e.instanceId !== undefined) {
      console.log(`Merge Drag trigger: ${dragSource} -> ${e.instanceId}`);
      
      const ws = (window as any).NeuronWS;
      if (ws?.readyState === WebSocket.OPEN) {
        // Send Bidirectional action
        ws.send(JSON.stringify({
          action: "merge",
          source: `cortex/visual_node_${dragSource}`,
          target: `cortex/visual_node_${e.instanceId}`
        }));
      }
    }
    setDragSource(null);
  };

  return (
    <instancedMesh
      ref={meshRef}
      args={[undefined, undefined, NEURON_COUNT]}
      frustumCulled={false}
      onPointerDown={handlePointerDown}
      onPointerUp={handlePointerUp}
      onPointerMissed={() => setDragSource(null)}
    >
      <octahedronGeometry args={[0.8, 0]}>
        <instancedBufferAttribute attach="attributes-color" args={[colors, 3]} />
      </octahedronGeometry>
      <meshPhongMaterial
        vertexColors
        transparent
        opacity={0.7}
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
      <ambientLight color={0x10b981} intensity={0.2} />
      <pointLight ref={pLightRef} color={0x3b82f6} intensity={1} distance={300} position={[0, 50, 0]} />
      <fogExp2 attach="fog" args={[0x030305, 0.003]} />
    </>
  );
}

export default function BrainScene() {
  return (
    <div className="w-full h-full relative" style={{ background: 'radial-gradient(circle at center, #0a0a14 0%, #030305 100%)' }}>
      <Canvas
        camera={{ position: [100, 80, 150], fov: 45 }}
        gl={{ antialias: true, alpha: true }}
      >
        <SceneLighting />
        <InstancedBrain />
        <OrbitControls 
          enableDamping 
          dampingFactor={0.05} 
          autoRotate 
          autoRotateSpeed={0.5} 
        />
      </Canvas>
      <div className="absolute bottom-5 left-5 font-mono text-[10px] text-slate-400 pointer-events-none z-20">
        InstancedMesh | {NEURON_COUNT.toLocaleString()} Neurons | 1 Draw Call
      </div>
    </div>
  );
}
