import React, { useRef, useState, useLayoutEffect } from 'react'

export default function EventStream({ events = [] }) {
  const containerRef = useRef(null)
  const [autoScroll, setAutoScroll] = useState(true)

  const toggleScroll = () => setAutoScroll(!autoScroll)

  useLayoutEffect(() => {
    if (autoScroll && containerRef.current) {
      containerRef.current.scrollTop = 0
    }
  })

  const sevClass = (s) => {
    switch (s) {
      case 'critical': return 'sev-critical'
      case 'high': return 'sev-high'
      case 'medium': return 'sev-medium'
      case 'low': return 'sev-low'
      default: return 'sev-info'
    }
  }

  const sevBadge = (s) => {
    switch (s) {
      case 'critical': return 'CRIT'
      case 'high': return 'HIGH'
      case 'medium': return 'MED'
      case 'low': return 'LOW'
      default: return 'INFO'
    }
  }

  return (
    <div className="stream-wrap" ref={containerRef}>
      {events.length === 0 ? (
        <div className="stream-empty">
          <div className="stream-empty-icon">◈</div>
          <span>Awaiting intelligence feed...</span>
          <span className="stream-hint">Submit a target to commence scan</span>
        </div>
      ) : (
        <div className="stream-list">
          {events.map((ev) => (
            <div key={ev.id} className={`stream-row ${sevClass(ev.severity)}`}>
              <span className="stream-badge">{sevBadge(ev.severity)}</span>
              <span className="stream-agent">{ev.agent}</span>
              <span className="stream-title">{ev.title}</span>
            </div>
          ))}
        </div>
      )}

      <button className="stream-toggle" onClick={toggleScroll}>
        {autoScroll ? '⏸' : '▶'}
      </button>

      <style>{`
        .stream-wrap {
          flex: 1;
          overflow-y: auto;
          padding: 6px;
          font-family: var(--font-mono);
          font-size: 10px;
          position: relative;
        }
        .stream-empty {
          display: flex;
          flex-direction: column;
          align-items: center;
          justify-content: center;
          height: 100%;
          gap: 6px;
          color: var(--text-dim);
        }
        .stream-empty-icon {
          font-size: 24px;
          color: var(--text-dim);
          opacity: 0.3;
        }
        .stream-hint {
          font-size: 9px;
          opacity: 0.5;
        }
        .stream-list {
          display: flex;
          flex-direction: column;
          gap: 1px;
        }
        .stream-row {
          display: flex;
          align-items: center;
          gap: 6px;
          padding: 3px 6px;
          border-radius: 3px;
          animation: event-slide-in 0.25s ease;
          line-height: 1.5;
        }
        .stream-row:hover {
          background: var(--bg-elevated);
        }
        .stream-badge {
          font-size: 8px;
          font-weight: 700;
          letter-spacing: 1px;
          padding: 1px 4px;
          border-radius: 2px;
          flex-shrink: 0;
          min-width: 28px;
          text-align: center;
        }
        .sev-critical .stream-badge { background: var(--red-dim); color: var(--red); border: 1px solid rgba(255, 23, 68, 0.2); }
        .sev-high .stream-badge { background: var(--amber-dim); color: var(--amber); border: 1px solid rgba(255, 171, 0, 0.2); }
        .sev-medium .stream-badge { background: var(--cyan-dim); color: var(--cyan); border: 1px solid rgba(0, 229, 255, 0.2); }
        .sev-low .stream-badge { background: var(--green-dim); color: var(--green); border: 1px solid rgba(0, 230, 118, 0.2); }
        .sev-info .stream-badge { background: var(--bg-elevated); color: var(--text-dim); border: 1px solid var(--border); }
        .stream-agent {
          color: var(--text-secondary);
          font-size: 9px;
          text-transform: uppercase;
          letter-spacing: 0.5px;
          flex-shrink: 0;
          min-width: 52px;
        }
        .stream-title {
          color: var(--text-primary);
          overflow: hidden;
          text-overflow: ellipsis;
          white-space: nowrap;
        }
        .sev-critical .stream-title { color: var(--red); }
        .sev-high .stream-title { color: var(--amber); }
        .sev-medium .stream-title { color: var(--cyan); }
        .stream-toggle {
          position: absolute;
          bottom: 6px;
          right: 6px;
          background: var(--bg-elevated);
          border: 1px solid var(--border);
          border-radius: 3px;
          padding: 2px 6px;
          cursor: pointer;
          font-size: 10px;
          color: var(--text-dim);
          opacity: 0.6;
          transition: opacity 0.2s;
        }
        .stream-toggle:hover {
          opacity: 1;
        }
      `}</style>
    </div>
  )
}