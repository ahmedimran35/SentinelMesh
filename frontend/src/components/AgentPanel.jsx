import React from 'react'

export default function AgentPanel({ agent }) {
  const { name, icon, color, state, progress, message } = agent

  const stateColors = {
    idle: 'var(--text-dim)',
    scanning: 'var(--cyan)',
    analyzing: 'var(--amber)',
    complete: 'var(--green)',
    error: 'var(--red)',
  }
  const stateColor = stateColors[state] || 'var(--text-dim)'
  const isActive = state === 'scanning' || state === 'analyzing'
  const pct = Math.round((progress || 0) * 100)

  return (
    <div className={`agent ${isActive ? 'active' : ''} ${state}`}>
      <div className="agent-head">
        <span className="agent-icon" style={{ '--accent': color }}>{icon}</span>
        <span className="agent-name" style={{ color }}>{name}</span>
      </div>
      <div className="agent-body">
        <span className="agent-state" style={{ color: stateColor }}>
          {state === 'idle' ? '—' : message || state}
        </span>
      </div>
      <div className="agent-track">
        <div className="agent-fill" style={{
          width: `${pct}%`,
          background: `linear-gradient(90deg, ${color}66, ${color})`,
          boxShadow: isActive ? `0 0 6px ${color}44` : 'none',
        }} />
      </div>

      <style>{`
        .agent {
          background: var(--bg-surface);
          border: 1px solid var(--border);
          border-radius: 6px;
          padding: 10px;
          display: flex;
          flex-direction: column;
          gap: 6px;
          transition: all 0.3s;
        }
        .agent.active {
          border-color: var(--border-hover);
          animation: agent-breathe 2s ease-in-out infinite;
          --agent-color: var(--cyan-glow);
        }
        .agent.complete {
          border-color: rgba(0, 230, 118, 0.15);
        }
        .agent.error {
          border-color: rgba(255, 23, 68, 0.25);
          animation: critical-pulse 1s ease-in-out infinite;
        }
        .agent-head {
          display: flex;
          align-items: center;
          gap: 6px;
        }
        .agent-icon {
          font-size: 16px;
          line-height: 1;
        }
        .agent-name {
          font-family: var(--font-mono);
          font-size: 10px;
          font-weight: 600;
          letter-spacing: 0.8px;
          text-transform: uppercase;
        }
        .agent-body {
          min-height: 14px;
        }
        .agent-state {
          font-family: var(--font-mono);
          font-size: 9px;
          line-height: 1.4;
          display: -webkit-box;
          -webkit-line-clamp: 2;
          -webkit-box-orient: vertical;
          overflow: hidden;
        }
        .agent-track {
          height: 2px;
          background: var(--bg-elevated);
          border-radius: 2px;
          overflow: hidden;
        }
        .agent-fill {
          height: 100%;
          border-radius: 2px;
          transition: width 0.5s ease;
          animation: progress-fill 0.4s ease;
        }
      `}</style>
    </div>
  )
}