import React, { useRef, useState, useLayoutEffect } from 'react'

export default function EventStream({ events = [] }) {
  const containerRef = useRef(null)
  const [autoScroll, setAutoScroll] = useState(true)
  const [expanded, setExpanded] = useState(null)

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
          <div className="stream-empty-icon">◇</div>
          <span className="stream-empty-text">AWAITING INTELLIGENCE FEED</span>
          <span className="stream-hint">Submit a target to commence reconnaissance</span>
        </div>
      ) : (
        <div className="stream-list">
          {events.map((ev) => (
            <div key={ev.id} className={`stream-row ${sevClass(ev.severity)}`}>
              <span className="stream-time">{ev.time}</span>
              <span className="stream-badge">{sevBadge(ev.severity)}</span>
              <span className="stream-agent">{ev.agent}</span>
              <span className="stream-title" title={ev.title}>{ev.title}</span>
            </div>
          ))}
        </div>
      )}

      <button className="stream-toggle" onClick={toggleScroll}>
        {autoScroll ? '‖' : '▸'}
      </button>

      <style>{`
        .stream-wrap {
          flex: 1;
          overflow-y: auto;
          padding: 4px;
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
          gap: 8px;
          color: var(--text-dim);
        }
        .stream-empty-icon {
          font-size: 28px;
          color: var(--cyan);
          opacity: 0.15;
          animation: pulse-dot 3s ease-in-out infinite;
        }
        .stream-empty-text {
          font-family: var(--font-display);
          font-size: 9px;
          letter-spacing: 3px;
          opacity: 0.4;
        }
        .stream-hint {
          font-size: 9px;
          opacity: 0.3;
        }
        .stream-list {
          display: flex;
          flex-direction: column;
          gap: 1px;
        }
        .stream-row {
          display: flex;
          align-items: center;
          gap: 8px;
          padding: 3px 8px;
          border-radius: 2px;
          animation: event-slide-in 0.25s ease;
          line-height: 1.6;
          transition: background 0.15s;
        }
        .stream-row:hover { background: var(--bg-elevated); }
        .stream-time {
          font-size: 8px;
          color: var(--text-dim);
          min-width: 55px;
          letter-spacing: 0.5px;
          flex-shrink: 0;
        }
        .stream-badge {
          font-family: var(--font-display);
          font-size: 7px;
          font-weight: 600;
          letter-spacing: 1.5px;
          padding: 1px 5px;
          border-radius: 2px;
          flex-shrink: 0;
          min-width: 30px;
          text-align: center;
        }
        .sev-critical .stream-badge { background: var(--red-dim); color: var(--red); border: 1px solid rgba(255, 23, 68, 0.15); }
        .sev-high .stream-badge { background: var(--amber-dim); color: var(--amber); border: 1px solid rgba(255, 171, 0, 0.15); }
        .sev-medium .stream-badge { background: var(--cyan-dim); color: var(--cyan); border: 1px solid rgba(0, 229, 255, 0.15); }
        .sev-low .stream-badge { background: var(--green-dim); color: var(--green); border: 1px solid rgba(0, 230, 118, 0.15); }
        .sev-info .stream-badge { background: var(--bg-elevated); color: var(--text-dim); border: 1px solid var(--border); }
        .stream-agent {
          color: var(--text-dim);
          font-size: 8px;
          text-transform: uppercase;
          letter-spacing: 1px;
          flex-shrink: 0;
          min-width: 52px;
          font-family: var(--font-display);
        }
        .stream-title {
          color: var(--text-secondary);
          overflow: hidden;
          text-overflow: ellipsis;
          white-space: nowrap;
          flex: 1;
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
          padding: 2px 8px;
          cursor: pointer;
          font-size: 10px;
          color: var(--text-dim);
          opacity: 0.5;
          transition: opacity 0.2s;
        }
        .stream-toggle:hover { opacity: 1; }
      `}</style>
    </div>
  )
}
