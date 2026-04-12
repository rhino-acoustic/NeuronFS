'use client';

import React, { useEffect, useState, useRef } from 'react';
import { Activity, Server } from 'lucide-react';

interface MetricState {
  neurons: number | string;
  activation: number | string;
}

interface LogEntry {
  ts: string;
  level: string;
  msg: string;
}

export default function NerveCenterUI() {
  const [metrics, setMetrics] = useState<MetricState>({ neurons: '---', activation: '---' });
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const logsEndRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    // Auto-scroll TTY
    if (logsEndRef.current) {
      logsEndRef.current.scrollIntoView({ behavior: 'smooth' });
    }
  }, [logs]);

  useEffect(() => {
    let ws: WebSocket;
    let reconnectTimer: NodeJS.Timeout;

    const connect = () => {
      ws = new WebSocket('ws://localhost:7350/api/ws');

      ws.onmessage = (e) => {
        try {
          const data = JSON.parse(e.data);
          const now = new Date();
          const timeStr = `${now.getHours().toString().padStart(2, '0')}:${now.getMinutes().toString().padStart(2, '0')}:${now.getSeconds().toString().padStart(2, '0')}`;
          
          if (data.type === 'stat') {
            setMetrics({
              neurons: data.totalNeurons.toLocaleString(),
              activation: data.totalActivation.toLocaleString()
            });
          } else if (data.level) {
            setLogs(prev => [...prev, { ts: timeStr, level: data.level, msg: data.msg }].slice(-100));
          } else if (data.type === 'fire') {
            setLogs(prev => [...prev, { ts: timeStr, level: 'INFO', msg: `FIRE EVENT: ${data.path} (${data.old}→${data.new})` }].slice(-100));
            window.dispatchEvent(new CustomEvent('neuron-fire', { detail: data }));
          }
        } catch (err) {
          console.error('WS Error:', err);
        }
      };

      ws.onerror = () => {
        setLogs(prev => [...prev, { ts: '---', level: 'ERROR', msg: 'WebSocket Connection Error. Retrying...' }].slice(-100));
      };

      ws.onclose = () => {
        reconnectTimer = setTimeout(connect, 3000);
      };

      // expose for BrainScene
      (window as any).NeuronWS = ws;
    };

    connect();

    return () => {
      clearTimeout(reconnectTimer);
      if (ws) ws.close();
    };
  }, []);

  return (
    <div className="w-[450px] bg-[var(--surface)] backdrop-blur-[28px] border-l border-[var(--border)] flex flex-col z-10 h-full">
      <div className="p-5 border-b border-[var(--border)] flex justify-between items-center">
        <div className="brand">
          <h1 className="text-lg font-extrabold tracking-widest text-white drop-shadow-[0_0_10px_rgba(255,255,255,0.2)]">
            <span className="text-[var(--accent)]">Neuron</span>FS V3
          </h1>
          <span className="text-[10px] text-[var(--text-muted)] font-mono">React Three Fiber Port</span>
        </div>
      </div>
      
      <div className="p-5 border-b border-[var(--border)] flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="w-2.5 h-2.5 rounded-full bg-[var(--accent)] shadow-[0_0_15px_var(--accent)] animate-[pulse_2s_infinite]"></div>
          <div className="text-xs font-semibold text-white">System Listening...</div>
        </div>
      </div>

      <div className="flex-1 p-5 font-mono text-[11px] overflow-y-auto text-zinc-400 leading-relaxed tty-feed">
        {logs.map((log, i) => (
          <div key={i} className="mb-1.5 flex gap-2">
            <span className="text-zinc-600 shrink-0">[{log.ts}]</span>
            <span className={log.level === 'ERROR' ? 'text-red-500' : log.level === 'WARN' ? 'text-amber-400' : 'text-blue-400'}>
              {log.msg}
            </span>
          </div>
        ))}
        <div ref={logsEndRef} />
      </div>

      <div className="p-5 border-t border-[var(--border)] grid grid-cols-2 gap-[15px] bg-black/30">
        <div className="flex flex-col gap-1">
          <span className="text-[9px] uppercase tracking-widest text-[var(--text-muted)]">Active Neurons</span>
          <span className="text-base font-extrabold text-white font-mono">{metrics.neurons}</span>
        </div>
        <div className="flex flex-col gap-1">
          <span className="text-[9px] uppercase tracking-widest text-[var(--text-muted)]">Total Activation</span>
          <span className="text-base font-extrabold text-white font-mono">{metrics.activation}</span>
        </div>
        <div className="flex flex-col gap-1">
          <span className="text-[9px] uppercase tracking-widest text-[var(--text-muted)]">Node Daemon</span>
          <span className="text-xs font-bold text-[var(--accent)] flex items-center gap-1"><Server size={12}/> ONLINE</span>
        </div>
        <div className="flex flex-col gap-1">
          <span className="text-[9px] uppercase tracking-widest text-[var(--text-muted)]">Go Supervisor</span>
          <span className="text-xs font-bold text-[var(--accent)] flex items-center gap-1"><Activity size={12}/> ONLINE</span>
        </div>
      </div>
    </div>
  );
}
