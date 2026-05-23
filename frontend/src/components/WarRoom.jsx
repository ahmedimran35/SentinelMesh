import React, { useState, useEffect } from 'react'

export default function WarRoom({ stats, investigations, activeInvestigation, onSelectInvestigation, onOpenSettings, onOpenAlerts, children }) {
  const [clock, setClock] = useState('')
  const [booted, setBooted] = useState(false)

  useEffect(() => {
    const tick = () => {
      const now = new Date()
      setClock(now.toLocaleTimeString('en-US', { hour12: false, hour: '2-digit', minute: '2-digit', second: '2-digit' }))
    }
    tick()
    const id = setInterval(tick, 1000)
    setTimeout(() => setBooted(true), 100)
    return () => clearInterval(id)
  }, [])

  const renderRiskBadge = (rating) => {
    const colors = { critical: 'var(--red)', high: 'var(--amber)', medium: 'var(--cyan)', low: 'var(--green)', info: 'var(--text-dim)' }
    const color = colors[rating] || 'var(--text-dim)'
    return <span className="risk-dot" style={{ background: color, boxShadow: `0 0 6px ${color}` }} />
  }

  const formatTime = (iso) => {
    if (!iso) return '—'
    const d = new Date(iso)
    return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })
  }

  return (
    <div className={`warroom ${booted ? 'booted' : ''}`}>
      {/* Grid background */}
      <div className="grid-bg" />

      <header className="warroom-header">
        <div className="header-left">
          <div className="logo-mark">
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none">
              <path d="M12 2L20 7V17L12 22L4 17V7L12 2Z" stroke="var(--cyan)" strokeWidth="1.5" fill="none" opacity="0.6" />
              <path d="M12 6L16 8.5V13.5L12 16L8 13.5V8.5L12 6Z" stroke="var(--cyan)" strokeWidth="1.2" fill="none" />
              <circle cx="12" cy="11" r="2" fill="var(--cyan)" opacity="0.8" />
              <line x1="12" y1="2" x2="12" y2="6" stroke="var(--cyan)" strokeWidth="0.8" opacity="0.4" />
              <line x1="12" y1="16" x2="12" y2="22" stroke="var(--cyan)" strokeWidth="0.8" opacity="0.4" />
            </svg>
          </div>
          <div className="logo-block">
            <span className="logo-text">SENTINEL<span className="logo-accent">MESH</span></span>
            <span className="logo-sub">AUTONOMOUS SECURITY OPERATIONS</span>
          </div>
          <span className="status-badge alive">
            <span className="status-pulse" />
            <span>ONLINE</span>
          </span>
        </div>

        <div className="header-center">
          <div className="hud-clock">{clock}</div>
          {stats.active_scans > 0 && (
            <span className="stat-tag scanning">
              <span className="stat-icon">⟳</span>
              {stats.active_scans} ACTIVE
            </span>
          )}
          {stats.unacknowledged_alerts > 0 && (
            <span className="stat-tag alert clickable" onClick={onOpenAlerts}>
              <span className="stat-icon">▲</span>
              {stats.unacknowledged_alerts} ALERTS
            </span>
          )}
          {stats.total_findings > 0 && (
            <span className="stat-tag">
              <span className="stat-icon">◆</span>
              {stats.total_findings} INTEL
            </span>
          )}
        </div>

        <div className="header-right">
          {investigations.length > 0 && (
            <div className="history-dropdown">
              <select
                onChange={(e) => {
                  if (e.target.value) onSelectInvestigation({ id: e.target.value })
                }}
                value={activeInvestigation?.id || ''}
              >
                <option value="">HISTORY ▾</option>
                {investigations.slice(0, 20).map(inv => (
                  <option key={inv.id} value={inv.id}>
                    {inv.target} — {inv.status} — {inv.risk_rating || '?'}
                  </option>
                ))}
              </select>
            </div>
          )}
          <button className="icon-btn" onClick={onOpenAlerts} title="Alerts">
            <svg width="15" height="15" viewBox="0 0 15 15" fill="none">
              <path d="M7.5 1.5L13.5 12.5H1.5L7.5 1.5Z" stroke="currentColor" strokeWidth="1.2" fill="none" />
              <line x1="7.5" y1="5.5" x2="7.5" y2="8.5" stroke="currentColor" strokeWidth="1.2" />
              <circle cx="7.5" cy="10.5" r="0.6" fill="currentColor" />
            </svg>
          </button>
          <button className="icon-btn" onClick={onOpenSettings} title="Settings">
            <svg width="15" height="15" viewBox="0 0 15 15" fill="none">
              <circle cx="7.5" cy="7.5" r="2.2" stroke="currentColor" strokeWidth="1.2" />
              <path d="M7.5 1.5V3.5M7.5 11.5V13.5M1.5 7.5H3.5M11.5 7.5H13.5M3 3L4.5 4.5M10.5 10.5L12 12M3 12L4.5 10.5M10.5 4.5L12 3"
                stroke="currentColor" strokeWidth="1.2" strokeLinecap="round" />
            </svg>
          </button>
        </div>
      </header>

      <main className="warroom-main">
        {children}
      </main>

      {activeInvestigation && (
        <aside className="investigation-bar">
          <div className="inv-bar-left">
            <span className="inv-label">TARGET</span>
            <span className="inv-target">{activeInvestigation.target}</span>
            <span className="inv-sep">│</span>
            <span className={`inv-status status-${activeInvestigation.status}`}>
              {activeInvestigation.status === 'running' && <span className="inv-running-dot" />}
              {activeInvestigation.status.toUpperCase()}
            </span>
            {activeInvestigation.risk_rating && (
              <>
                <span className="inv-sep">│</span>
                <span className="inv-risk">{renderRiskBadge(activeInvestigation.risk_rating)}{activeInvestigation.risk_rating.toUpperCase()}</span>
              </>
            )}
          </div>
          <div className="inv-bar-right">
            <span className="inv-time">{formatTime(activeInvestigation.created_at)}</span>
          </div>
        </aside>
      )}

      <style>{`
        .warroom {
          display: flex;
          flex-direction: column;
          height: 100vh;
          width: 100vw;
          overflow: hidden;
          position: relative;
          opacity: 0;
          transition: opacity 0.6s ease;
        }
        .warroom.booted {
          opacity: 1;
        }

        .grid-bg {
          position: fixed;
          top: 0; left: 0; right: 0; bottom: 0;
          background-image:
            linear-gradient(var(--border) 1px, transparent 1px),
            linear-gradient(90deg, var(--border) 1px, transparent 1px);
          background-size: 60px 60px;
          opacity: 0.06;
          pointer-events: none;
          z-index: 0;
        }

        .warroom-header {
          display: flex;
          align-items: center;
          justify-content: space-between;
          padding: 0 20px;
          background: linear-gradient(180deg, rgba(12,12,20,0.98), rgba(12,12,20,0.92));
          border-bottom: 1px solid var(--border);
          z-index: 50;
          min-height: 52px;
          gap: 16px;
          backdrop-filter: blur(12px);
          animation: slide-up 0.4s ease;
        }
        .header-left { display: flex; align-items: center; gap: 12px; }
        .logo-mark {
          display: flex;
          align-items: center;
          animation: flicker 8s infinite;
        }
        .logo-block { display: flex; flex-direction: column; gap: 1px; }
        .logo-text {
          font-family: var(--font-display);
          font-weight: 700;
          font-size: 14px;
          letter-spacing: 4px;
          color: var(--text-primary);
        }
        .logo-accent { color: var(--cyan); }
        .logo-sub {
          font-family: var(--font-mono);
          font-size: 7px;
          letter-spacing: 2.5px;
          color: var(--text-dim);
          text-transform: uppercase;
        }
        .status-badge {
          display: flex;
          align-items: center;
          gap: 6px;
          font-family: var(--font-mono);
          font-size: 8px;
          letter-spacing: 2px;
          padding: 3px 10px;
          border-radius: 3px;
        }
        .status-badge.alive {
          background: var(--green-dim);
          color: var(--green);
          border: 1px solid rgba(0, 230, 118, 0.15);
        }
        .status-pulse {
          width: 5px;
          height: 5px;
          border-radius: 50%;
          background: var(--green);
          animation: pulse-dot 2s ease-in-out infinite;
        }

        .header-center { display: flex; gap: 8px; align-items: center; }
        .hud-clock {
          font-family: var(--font-display);
          font-size: 11px;
          letter-spacing: 3px;
          color: var(--text-dim);
          min-width: 80px;
          text-align: center;
        }
        .stat-tag {
          display: flex;
          align-items: center;
          gap: 5px;
          font-family: var(--font-mono);
          font-size: 9px;
          letter-spacing: 1px;
          color: var(--text-secondary);
          background: var(--bg-elevated);
          padding: 4px 10px;
          border-radius: 3px;
          border: 1px solid var(--border);
        }
        .stat-tag.scanning {
          color: var(--cyan);
          border-color: rgba(0, 229, 255, 0.15);
        }
        .stat-tag.alert {
          color: var(--red);
          border-color: rgba(255, 23, 68, 0.15);
        }
        .stat-tag.clickable { cursor: pointer; }
        .stat-tag.clickable:hover { background: var(--bg-hover); }
        .stat-icon { font-size: 10px; }

        .header-right { display: flex; align-items: center; gap: 8px; }

        .history-dropdown select {
          background: var(--bg-elevated);
          border: 1px solid var(--border);
          border-radius: 3px;
          padding: 5px 10px;
          font-family: var(--font-mono);
          font-size: 9px;
          letter-spacing: 0.5px;
          color: var(--text-secondary);
          cursor: pointer;
          max-width: 200px;
          appearance: none;
        }
        .history-dropdown select:focus { border-color: var(--cyan-dim); outline: none; }
        .history-dropdown select option { background: var(--bg-surface); color: var(--text-primary); }

        .icon-btn {
          background: transparent;
          border: 1px solid var(--border);
          border-radius: 4px;
          width: 30px;
          height: 30px;
          display: flex;
          align-items: center;
          justify-content: center;
          cursor: pointer;
          color: var(--text-secondary);
          transition: all 0.2s;
        }
        .icon-btn:hover {
          background: var(--bg-elevated);
          color: var(--cyan);
          border-color: var(--cyan-dim);
        }

        .warroom-main {
          flex: 1;
          overflow: hidden;
          display: flex;
          flex-direction: column;
          position: relative;
          z-index: 1;
        }

        .investigation-bar {
          display: flex;
          align-items: center;
          justify-content: space-between;
          padding: 6px 20px;
          background: rgba(12,12,20,0.95);
          border-top: 1px solid var(--border);
          z-index: 50;
          font-family: var(--font-mono);
          font-size: 10px;
          min-height: 30px;
          animation: slide-up 0.3s ease;
          backdrop-filter: blur(12px);
        }
        .inv-bar-left { display: flex; align-items: center; gap: 8px; }
        .inv-bar-right { display: flex; align-items: center; gap: 8px; }
        .inv-label {
          font-size: 8px;
          letter-spacing: 1.5px;
          color: var(--text-dim);
          font-family: var(--font-display);
        }
        .inv-target { color: var(--cyan); font-weight: 600; letter-spacing: 0.5px; }
        .inv-sep { color: var(--text-dim); opacity: 0.3; }
        .inv-status { letter-spacing: 1.5px; font-weight: 600; font-size: 9px; display: flex; align-items: center; gap: 5px; }
        .inv-running-dot {
          width: 5px; height: 5px; border-radius: 50%;
          background: var(--cyan);
          animation: pulse-dot 1.5s ease-in-out infinite;
        }
        .status-running { color: var(--cyan); }
        .status-complete { color: var(--green); }
        .status-error { color: var(--red); }
        .inv-risk { display: flex; align-items: center; gap: 5px; }
        .risk-dot { width: 6px; height: 6px; border-radius: 50%; display: inline-block; }
        .inv-time { color: var(--text-dim); letter-spacing: 1px; font-size: 9px; }
      `}</style>
    </div>
  )
}
