import React from 'react'

export default function AgentPanel({ agent, index = 0 }) {
  const { name, icon, color, state, progress, message } = agent

  const stateColors = {
    idle: 'var(--text-dim)',
    scanning: 'var(--cyan)',
    analyzing: 'var(--amber)',
    complete: 'var(--green)',
    error: 'var(--red)',
  }
  const stateLabels = {
    idle: 'STANDBY',
    scanning: 'SCANNING',
    analyzing: 'ANALYZING',
    complete: 'COMPLETE',
    error: 'ERROR',
  }
  const stateColor = stateColors[state] || 'var(--text-dim)'
  const isActive = state === 'scanning' || state === 'analyzing'
  const pct = Math.round((progress || 0) * 100)

  return (
    <div
      className={`agent ${isActive ? 'active' : ''} ${state}`}
      style={{ animationDelay: `${index * 80}ms` }}
    >
      <div className="agent-head">
        <span className="agent-icon" style={{ '--accent': color }}>{icon}</span>
        <div className="agent-info">
          <span className="agent-name" style={{ color }}>{name}</span>
          <span className="agent-state-tag" style={{ color: stateColor }}>
            {stateLabels[state] || state.toUpperCase()}
          </span>
        </div>
        {isActive && <span className="agent-spinner" style={{ borderColor: `${color}33 ${color}33 ${color} ${color}` }} />}
      </div>
      <div className="agent-body">
        <span className="agent-state" style={{ color: state === 'idle' ? 'var(--text-dim)' : 'var(--text-secondary)' }}>
          {state === 'idle' ? 'Awaiting tasking...' : message || state}
        </span>
      </div>
      <div className="agent-track">
        <div className="agent-fill" style={{
          width: `${pct}%`,
          background: `linear-gradient(90deg, ${color}44, ${color})`,
          boxShadow: isActive ? `0 0 8px ${color}33` : 'none',
        }} />
      </div>
      <div className="agent-pct">{pct}%</div>

      <style>{`
        .agent {
          background: var(--bg-surface);
          border: 1px solid var(--border);
          border-radius: 4px;
          padding: 10px 12px;
          display: flex;
          flex-direction: column;
          gap: 6px;
          transition: all 0.3s;
          animation: stagger-in 0.4s ease both;
          position: relative;
          overflow: hidden;
        }
        .agent::before {
          content: '';
          position: absolute;
          top: 0; left: 0; right: 0;
          height: 1px;
          background: linear-gradient(90deg, transparent, var(--cyan-dim), transparent);
          opacity: 0;
          transition: opacity 0.3s;
        }
        .agent.active::before { opacity: 1; }
        .agent.active {
          border-color: rgba(0, 229, 255, 0.12);
          animation: stagger-in 0.4s ease both, agent-breathe 2.5s ease-in-out infinite;
          --agent-color: var(--cyan-glow);
        }
        .agent.complete { border-color: rgba(0, 230, 118, 0.12); }
        .agent.error {
          border-color: rgba(255, 23, 68, 0.2);
          animation: stagger-in 0.4s ease both, critical-pulse 1.5s ease-in-out infinite;
        }
        .agent-head {
          display: flex;
          align-items: center;
          gap: 8px;
        }
        .agent-icon { font-size: 18px; line-height: 1; }
        .agent-info { display: flex; flex-direction: column; gap: 1px; flex: 1; }
        .agent-name {
          font-family: var(--font-display);
          font-size: 9px;
          font-weight: 600;
          letter-spacing: 1.5px;
          text-transform: uppercase;
        }
        .agent-state-tag {
          font-family: var(--font-mono);
          font-size: 7px;
          letter-spacing: 1.5px;
        }
        .agent-spinner {
          width: 12px; height: 12px;
          border: 2px solid;
          border-radius: 50%;
          animation: spin 1s linear infinite;
          flex-shrink: 0;
        }
        @keyframes spin { to { transform: rotate(360deg); } }
        .agent-body { min-height: 14px; }
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
          animation: progress-fill 0.5s ease;
        }
        .agent-pct {
          position: absolute;
          bottom: 6px;
          right: 8px;
          font-family: var(--font-display);
          font-size: 9px;
          color: var(--text-dim);
          letter-spacing: 1px;
        }
      `}</style>
    </div>
  )
}
