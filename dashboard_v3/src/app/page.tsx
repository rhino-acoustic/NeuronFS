'use client';

import dynamic from 'next/dynamic';
import React from 'react';
import NerveCenterUI from '@/components/NerveCenterUI';

// Prevent SSR for Three.js WebGL canvas context
const BrainScene = dynamic(() => import('@/components/BrainScene'), { 
  ssr: false, 
  loading: () => <div className="w-full h-full flex items-center justify-center bg-[#030305] text-[var(--accent)] font-mono">Initializing Neural Render Engine...</div>
});

export default function Home() {
  return (
    <main className="flex w-full h-screen overflow-hidden">
      {/* 3D WebGL Visualization Area */}
      <section className="flex-1 relative border-r border-[var(--border)]">
        <BrainScene />
      </section>

      {/* Nerve Center HUD / Command Module */}
      <NerveCenterUI />
    </main>
  );
}
