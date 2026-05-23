import React, { useState, useEffect, useCallback } from 'react'
import { getAlerts, acknowledgeAlert } from '../services/api'
import { severityColors } from '../utils/severity'

export default function AlertPanel({ onClose }) {
  const [alerts, setAlerts] = useState([])
  const [loading, setLoading] = useState(true)
  const [filter, setFilter] = useState('unacknowledged')

  const loadAlerts = useCallback(async () => {
    setLoading(true)
    try {
      const data = await getAlerts(filter === 'unacknowledged')
      setAlerts(Array.isArray(data) ? data : [])
    } catch (e) {
      console.error('Failed to load alerts:', e)
    }
    setLoading(false)
  }, [filter])

  useEffect(() => { loadAlerts() }, [loadAlerts])

  const handleAcknowledge = async (id) => {
    try {
      await acknowledgeAlert(id)
      setAlerts(prev => prev.map(a => a.id === id ? { ...a, acknowledged: true } : a))
    } catch (e) {
      console.error('Ack failed:', e)
    }
  }

  return (
    <div className="alert-overlay" onClick={onClose}>
      <div className="alert-modal" onClick={(e) => e.stopPropagation()}>
        <div className="alert-head">
          <span className="alert-title">
            <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
              <path d="M7 1L13 12H1L7 1Z" stroke="currentColor" strokeWidth="1.2" fill="none" />
              <line x1="7" y1="5" x2="7" y2="8" stroke="currentColor" strokeWidth="1.2" />
              <circle cx="7" cy="10" r="0.8" fill="currentColor" />
            </svg>
            ALERTS
          </span>
          <button className="alert-close" onClick={onClose}>✕</button>
        </div>

        <div className="alert-toolbar">
          <button className={`filter-btn ${filter === 'unacknowledged' ? 'active' : ''}`} onClick={() => setFilter('unacknowledged')}>Unacknowledged</button>
          <button className={`filter-btn ${filter === 'all' ? 'active' : ''}`} onClick={() => setFilter('all')}>All</button>
          <div className="alert-spacer" />
          <span className="alert-count">{alerts.length} alerts</span>
        </div>

        <div className="alert-list">
          {loading ? (
            <div className="alert-empty">Loading...</div>
          ) : alerts.length === 0 ? (
            <div className="alert-empty">No {filter === 'unacknowledged' ? 'unacknowledged ' : ''}alerts.</div>
          ) : alerts.map(a => (
            <div key={a.id} className={`alert-item sev-${a.severity} ${a.acknowledged ? 'acked' : ''}`}>
              <div className="alert-sev" style={{ color: severityColors[a.severity] }}>
                {a.severity?.toUpperCase()}
              </div>
              <div className="alert-body">
                <div className="alert-msg">{a.message}</div>
                <div className="alert-meta">
                  <span>{a.alert_type}</span>
                  <span>{a.target}</span>
                  <span>{a.created_at ? new Date(a.created_at).toLocaleString() : ''}</span>
                </div>
              </div>
              {!a.acknowledged && (
                <button className="btn-ack" onClick={() => handleAcknowledge(a.id)}>ACK</button>
              )}
            </div>
          ))}
        </div>
      </div>

      <style>{`
        .alert-overlay {
          position: fixed; top: 0; left: 0; right: 0; bottom: 0;
          background: rgba(0,0,0,0.85); backdrop-filter: blur(6px);
          display: flex; align-items: center; justify-content: center;
          z-index: 100; animation: fade-in 0.2s ease;
        }
        .alert-modal {
          background: var(--bg-surface); border: 1px solid var(--border);
          border-radius: 10px; width: 600px; max-width: 92vw; max-height: 80vh;
          overflow: hidden; display: flex; flex-direction: column;
          box-shadow: 0 24px 80px rgba(0,0,0,0.6);
        }
        .alert-head {
          display: flex; align-items: center; justify-content: space-between;
          padding: 14px 20px; border-bottom: 1px solid var(--border);
          background: var(--bg-elevated);
        }
        .alert-title {
          display: flex; align-items: center; gap: 10px;
          font-family: var(--font-mono); font-size: 12px; font-weight: 600;
          letter-spacing: 2px; color: var(--text-primary);
        }
        .alert-close {
          background: transparent; border: 1px solid var(--border);
          color: var(--text-secondary); width: 28px; height: 28px;
          border-radius: 5px; cursor: pointer; font-size: 13px;
          display: flex; align-items: center; justify-content: center;
          transition: all 0.2s;
        }
        .alert-close:hover { background: var(--bg-hover); color: var(--cyan); border-color: var(--cyan-dim); }
        .alert-toolbar {
          display: flex; align-items: center; gap: 4px;
          padding: 10px 20px; border-bottom: 1px solid var(--border);
        }
        .filter-btn {
          background: var(--bg-elevated); border: 1px solid var(--border);
          color: var(--text-dim); padding: 4px 12px; border-radius: 4px;
          cursor: pointer; font-family: var(--font-mono); font-size: 9px;
          letter-spacing: 0.5px; transition: all 0.2s;
        }
        .filter-btn:hover { color: var(--text-secondary); }
        .filter-btn.active { color: var(--cyan); border-color: var(--cyan-dim); background: var(--cyan-dim); }
        .alert-spacer { flex: 1; }
        .alert-count { font-family: var(--font-mono); font-size: 9px; color: var(--text-dim); }
        .alert-list { overflow-y: auto; flex: 1; }
        .alert-empty { color: var(--text-dim); font-family: var(--font-mono); font-size: 11px; text-align: center; padding: 40px; }
        .alert-item {
          display: flex; align-items: center; gap: 12px;
          padding: 10px 20px; border-bottom: 1px solid var(--border);
          transition: background 0.15s;
        }
        .alert-item:hover { background: var(--bg-elevated); }
        .alert-item.acked { opacity: 0.5; }
        .alert-sev {
          font-family: var(--font-mono); font-size: 8px; font-weight: 700;
          letter-spacing: 1px; min-width: 50px;
        }
        .alert-body { flex: 1; display: flex; flex-direction: column; gap: 2px; }
        .alert-msg { font-size: 12px; color: var(--text-primary); }
        .alert-meta {
          display: flex; gap: 12px;
          font-family: var(--font-mono); font-size: 9px; color: var(--text-dim);
        }
        .btn-ack {
          background: var(--bg-elevated); border: 1px solid var(--border);
          color: var(--green); padding: 4px 10px; border-radius: 3px;
          cursor: pointer; font-family: var(--font-mono); font-size: 9px;
          font-weight: 700; letter-spacing: 1px; transition: all 0.2s;
        }
        .btn-ack:hover { background: var(--green-dim); border-color: var(--green); }
      `}</style>
    </div>
  )
}
