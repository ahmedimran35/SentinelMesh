import React from 'react'

export default function WarRoom({ stats, investigations, activeInvestigation, onSelectInvestigation, onOpenSettings, onOpenAlerts, children }) {
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
    <div className="warroom">
      <header className="warroom-header">
        <div className="header-left">
          <svg className="logo-svg" width="22" height="22" viewBox="0 0 22 22" fill="none">
            <path d="M11 1L18 5.5V16.5L11 21L4 16.5V5.5L11 1Z" stroke="var(--cyan)" strokeWidth="1.5" fill="none" />
            <circle cx="11" cy="11" r="4" stroke="var(--cyan)" strokeWidth="1.2" fill="none" />
            <circle cx="11" cy="11" r="1.5" fill="var(--cyan)" />
          </svg>
          <span className="logo-text">SENTINELMESH</span>
          <span className="status-badge alive">
            <span className="status-pulse" />
            <span>SYSTEM ACTIVE</span>
          </span>
        </div>
        <div className="header-center">
          {stats.active_scans > 0 && (
            <span className="stat-tag scanning">
              <span className="stat-icon">⟳</span>
              {stats.active_scans} scanning
            </span>
          )}
          {stats.unacknowledged_alerts > 0 && (
            <span className="stat-tag alert clickable" onClick={onOpenAlerts}>
              <span className="stat-icon">⚠</span>
              {stats.unacknowledged_alerts} alerts
            </span>
          )}
          {stats.total_findings > 0 && (
            <span className="stat-tag">
              <span className="stat-icon">◈</span>
              {stats.total_findings} intel
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
                <option value="">History (click to view)</option>
                {investigations.slice(0, 20).map(inv => (
                  <option key={inv.id} value={inv.id}>
                    {inv.target} — {inv.status} — {inv.risk_rating || '?'}
                  </option>
                ))}
              </select>
            </div>
          )}
          <button className="icon-btn" onClick={onOpenSettings} title="Settings">
            <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
              <circle cx="8" cy="8" r="2.5" stroke="currentColor" strokeWidth="1.2" />
              <path d="M8 1V3M8 13V15M1 8H3M13 8H15M2.5 2.5L4 4M12 12L13.5 13.5M2.5 13.5L4 12M12 4L13.5 2.5"
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
            <span className="inv-target">{activeInvestigation.target}</span>
            <span className="inv-sep">|</span>
            <span className={`inv-status status-${activeInvestigation.status}`}>
              {activeInvestigation.status.toUpperCase()}
            </span>
            {activeInvestigation.risk_rating && (
              <>
                <span className="inv-sep">|</span>
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
        }

        .warroom-header {
          display: flex;
          align-items: center;
          justify-content: space-between;
          padding: 10px 20px;
          background: var(--bg-surface);
          border-bottom: 1px solid var(--border);
          z-index: 50;
          min-height: 48px;
          gap: 16px;
        }
        .header-left { display: flex; align-items: center; gap: 12px; }
        .logo-text {
          font-family: var(--font-mono);
          font-weight: 700;
          font-size: 14px;
          letter-spacing: 3px;
          color: var(--text-primary);
        }
        .status-badge {
          display: flex;
          align-items: center;
          gap: 6px;
          font-family: var(--font-mono);
          font-size: 9px;
          letter-spacing: 1.5px;
          padding: 3px 8px;
          border-radius: 4px;
        }
        .status-badge.alive {
          background: var(--green-dim);
          color: var(--green);
          border: 1px solid rgba(0, 230, 118, 0.2);
        }
        .status-pulse {
          width: 5px;
          height: 5px;
          border-radius: 50%;
          background: var(--green);
          animation: pulse-dot 2s ease-in-out infinite;
        }

        .header-center { display: flex; gap: 8px; align-items: center; }
        .stat-tag {
          display: flex;
          align-items: center;
          gap: 5px;
          font-family: var(--font-mono);
          font-size: 9px;
          letter-spacing: 0.5px;
          color: var(--text-secondary);
          background: var(--bg-elevated);
          padding: 4px 10px;
          border-radius: 4px;
          border: 1px solid var(--border);
        }
        .stat-tag.scanning {
          color: var(--cyan);
          border-color: var(--cyan-dim);
        }
        .stat-tag.alert {
          color: var(--red);
          border-color: var(--red-dim);
        }
        .stat-tag.clickable {
          cursor: pointer;
        }
        .stat-tag.clickable:hover {
          background: var(--bg-hover);
        }
        .stat-icon { font-size: 11px; }

        .header-right { display: flex; align-items: center; gap: 8px; }

        .history-dropdown select {
          background: var(--bg-elevated);
          border: 1px solid var(--border);
          border-radius: 4px;
          padding: 4px 8px;
          font-family: var(--font-mono);
          font-size: 9px;
          color: var(--text-secondary);
          cursor: pointer;
          max-width: 200px;
          appearance: none;
        }
        .history-dropdown select:focus {
          border-color: var(--cyan-dim);
          outline: none;
        }
        .history-dropdown select option {
          background: var(--bg-surface);
          color: var(--text-primary);
        }

        .icon-btn {
          background: transparent;
          border: 1px solid var(--border);
          border-radius: 6px;
          width: 32px;
          height: 32px;
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
          background: var(--bg-surface);
          border-top: 1px solid var(--border);
          z-index: 50;
          font-family: var(--font-mono);
          font-size: 10px;
          min-height: 32px;
          animation: slide-up 0.3s ease;
        }
        .inv-bar-left { display: flex; align-items: center; gap: 8px; }
        .inv-bar-right { display: flex; align-items: center; gap: 8px; }
        .inv-target { color: var(--cyan); font-weight: 600; letter-spacing: 0.5px; }
        .inv-sep { color: var(--text-dim); }
        .inv-status { letter-spacing: 1px; font-weight: 600; }
        .status-running { color: var(--cyan); }
        .status-complete { color: var(--green); }
        .status-error { color: var(--red); }
        .inv-risk { display: flex; align-items: center; gap: 5px; }
        .risk-dot {
          width: 6px; height: 6px;
          border-radius: 50%;
          display: inline-block;
        }
        .inv-time { color: var(--text-dim); }
      `}</style>
    </div>
  )
}