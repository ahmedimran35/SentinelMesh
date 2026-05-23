import React, { useState, useEffect } from 'react'
import { getInvestigation, getFindings, getRules, deleteInvestigation } from '../services/api'
import { riskColors } from '../utils/severity'

export default function InvestigationDetail({ investigationId, onClose, onDeleted }) {
  const [inv, setInv] = useState(null)
  const [findings, setFindings] = useState([])
  const [rules, setRules] = useState([])
  const [loading, setLoading] = useState(true)
  const [expandedFinding, setExpandedFinding] = useState(null)
  const [showDelete, setShowDelete] = useState(false)
  const [activeTab, setActiveTab] = useState('findings')

  useEffect(() => {
    if (!investigationId) return
    setLoading(true)
    Promise.all([
      getInvestigation(investigationId),
      getFindings(investigationId),
      getRules(investigationId),
    ])
      .then(([invData, findingsData, rulesData]) => {
        setInv(invData)
        setFindings(Array.isArray(findingsData) ? findingsData : [])
        setRules(Array.isArray(rulesData) ? rulesData : [])
      })
      .catch(err => console.error('Failed to load investigation:', err))
      .finally(() => setLoading(false))
  }, [investigationId])

  const handleDelete = async () => {
    try {
      await deleteInvestigation(investigationId)
      onDeleted?.(investigationId)
      onClose?.()
    } catch (e) {
      console.error('Delete failed:', e)
    }
  }

  const rc = riskColors[inv?.risk_rating] || 'var(--text-secondary)'

  const exportReport = (format) => {
    const url = `/api/investigations/${investigationId}/export?format=${format}`
    const a = document.createElement('a')
    a.href = url
    a.download = `sentinelmesh-${investigationId}.${format === 'stix' ? 'json' : format}`
    a.click()
  }

  return (
    <div className="detail-overlay" onClick={onClose}>
      <div className="detail-modal" onClick={(e) => e.stopPropagation()}>
        <div className="detail-head">
          <span className="detail-title">
            <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
              <circle cx="7" cy="7" r="5.5" stroke="currentColor" strokeWidth="1.2" fill="none" />
              <circle cx="7" cy="7" r="2" fill="currentColor" />
            </svg>
            INVESTIGATION DETAIL
          </span>
          <button className="detail-close" onClick={onClose}>✕</button>
        </div>

        <div className="detail-body">
          {loading ? (
            <div className="detail-loading">Loading investigation data...</div>
          ) : !inv ? (
            <div className="detail-loading">Investigation not found.</div>
          ) : (
            <>
              <div className="detail-meta">
                <div className="meta-cell">
                  <span className="meta-lbl">ID</span>
                  <span className="meta-val mono">{inv.id}</span>
                </div>
                <div className="meta-cell">
                  <span className="meta-lbl">Target</span>
                  <span className="meta-val mono glow-cyan">{inv.target}</span>
                </div>
                <div className="meta-cell">
                  <span className="meta-lbl">Status</span>
                  <span className={`meta-val status-${inv.status}`}>{inv.status?.toUpperCase()}</span>
                </div>
                <div className="meta-cell">
                  <span className="meta-lbl">Risk</span>
                  <span className="meta-risk" style={{ color: rc, borderColor: rc }}>
                    {(inv.risk_rating || 'unknown').toUpperCase()}
                  </span>
                </div>
                <div className="meta-cell">
                  <span className="meta-lbl">Findings</span>
                  <span className="meta-val">{findings.length}</span>
                </div>
                <div className="meta-cell">
                  <span className="meta-lbl">Created</span>
                  <span className="meta-val">{inv.created_at ? new Date(inv.created_at).toLocaleString() : '—'}</span>
                </div>
              </div>

              {inv.executive_summary && (
                <div className="detail-section">
                  <h4 className="section-head">EXECUTIVE SUMMARY</h4>
                  <p className="section-text">{inv.executive_summary}</p>
                </div>
              )}

              <div className="detail-tabs">
                <button className={`tab ${activeTab === 'findings' ? 'active' : ''}`} onClick={() => setActiveTab('findings')}>
                  Findings ({findings.length})
                </button>
                <button className={`tab ${activeTab === 'rules' ? 'active' : ''}`} onClick={() => setActiveTab('rules')}>
                  Rules ({rules.length})
                </button>
                {inv.full_report && (
                  <button className={`tab ${activeTab === 'report' ? 'active' : ''}`} onClick={() => setActiveTab('report')}>
                    Full Report
                  </button>
                )}
              </div>

              {activeTab === 'findings' && findings.length > 0 && (
                <div className="findings-list">
                  {findings.map(f => (
                    <div key={f.id} className={`finding-item sev-${f.severity} ${expandedFinding === f.id ? 'expanded' : ''}`}>
                      <div className="finding-header" onClick={() => setExpandedFinding(expandedFinding === f.id ? null : f.id)}>
                        <span className="finding-sev">{f.severity?.toUpperCase()}</span>
                        <span className="finding-agent">{f.agent}</span>
                        <span className="finding-type">{f.type}</span>
                        <span className="finding-title">{f.title}</span>
                        <span className="finding-expand">{expandedFinding === f.id ? '▼' : '▶'}</span>
                      </div>
                      {expandedFinding === f.id && (
                        <div className="finding-details">
                          <pre>{f.details}</pre>
                        </div>
                      )}
                    </div>
                  ))}
                </div>
              )}

              {activeTab === 'rules' && (
                <div className="rules-list">
                  {rules.length === 0 ? (
                    <div className="empty-msg">No rules generated for this investigation.</div>
                  ) : rules.map(r => (
                    <div key={r.id} className="rule-item">
                      <div className="rule-header">
                        <span className={`rule-type ${r.type}`}>{r.type?.toUpperCase()}</span>
                        <button className="btn-copy" onClick={() => navigator.clipboard?.writeText(r.content)}>Copy</button>
                      </div>
                      <pre className="rule-content">{r.content}</pre>
                    </div>
                  ))}
                </div>
              )}

              {activeTab === 'report' && inv.full_report && (
                <pre className="report-raw">{inv.full_report}</pre>
              )}

              <div className="detail-actions">
                <button className="action-btn" onClick={() => exportReport('json')}>Export JSON</button>
                <button className="action-btn" onClick={() => exportReport('csv')}>Export CSV</button>
                <button className="action-btn" onClick={() => exportReport('stix')}>Export STIX</button>
                <div className="action-spacer" />
                {!showDelete ? (
                  <button className="action-btn danger" onClick={() => setShowDelete(true)}>Delete</button>
                ) : (
                  <div className="delete-confirm">
                    <span>Confirm delete?</span>
                    <button className="action-btn danger" onClick={handleDelete}>Yes, Delete</button>
                    <button className="action-btn" onClick={() => setShowDelete(false)}>Cancel</button>
                  </div>
                )}
              </div>
            </>
          )}
        </div>
      </div>

      <style>{`
        .detail-overlay {
          position: fixed; top: 0; left: 0; right: 0; bottom: 0;
          background: rgba(0,0,0,0.85); backdrop-filter: blur(6px);
          display: flex; align-items: center; justify-content: center;
          z-index: 100; animation: fade-in 0.2s ease;
        }
        .detail-modal {
          background: var(--bg-surface); border: 1px solid var(--border);
          border-radius: 10px; width: 780px; max-width: 95vw; max-height: 90vh;
          overflow: hidden; display: flex; flex-direction: column;
          box-shadow: 0 24px 80px rgba(0,0,0,0.6);
        }
        .detail-head {
          display: flex; align-items: center; justify-content: space-between;
          padding: 14px 20px; border-bottom: 1px solid var(--border);
          background: var(--bg-elevated);
        }
        .detail-title {
          display: flex; align-items: center; gap: 10px;
          font-family: var(--font-mono); font-size: 12px; font-weight: 600;
          letter-spacing: 2px; color: var(--text-primary);
        }
        .detail-close {
          background: transparent; border: 1px solid var(--border);
          color: var(--text-secondary); width: 28px; height: 28px;
          border-radius: 5px; cursor: pointer; font-size: 13px;
          display: flex; align-items: center; justify-content: center;
          transition: all 0.2s;
        }
        .detail-close:hover { background: var(--bg-hover); color: var(--cyan); border-color: var(--cyan-dim); }
        .detail-body { padding: 20px; overflow-y: auto; display: flex; flex-direction: column; gap: 16px; }
        .detail-loading { color: var(--text-dim); font-family: var(--font-mono); font-size: 12px; text-align: center; padding: 40px 0; }
        .detail-meta { display: grid; grid-template-columns: repeat(3, 1fr); gap: 12px; }
        .meta-cell { display: flex; flex-direction: column; gap: 3px; }
        .meta-lbl { font-family: var(--font-mono); font-size: 8px; letter-spacing: 1.5px; color: var(--text-dim); text-transform: uppercase; }
        .meta-val { font-size: 12px; color: var(--text-primary); }
        .meta-val.mono { font-family: var(--font-mono); font-size: 10px; }
        .meta-risk { font-family: var(--font-mono); font-weight: 700; font-size: 13px; padding: 2px 8px; border: 1px solid; border-radius: 3px; display: inline-block; width: fit-content; }
        .status-running { color: var(--cyan); }
        .status-complete { color: var(--green); }
        .status-error { color: var(--red); }
        .detail-section { display: flex; flex-direction: column; gap: 8px; }
        .section-head { font-family: var(--font-mono); font-size: 10px; letter-spacing: 1.5px; color: var(--text-secondary); font-weight: 600; }
        .section-text { font-size: 13px; line-height: 1.7; color: var(--text-primary); }
        .detail-tabs { display: flex; gap: 2px; border-bottom: 1px solid var(--border); }
        .tab {
          background: transparent; border: none; padding: 8px 16px;
          font-family: var(--font-mono); font-size: 10px; letter-spacing: 0.5px;
          color: var(--text-dim); cursor: pointer; border-bottom: 2px solid transparent;
          transition: all 0.2s;
        }
        .tab:hover { color: var(--text-secondary); }
        .tab.active { color: var(--cyan); border-bottom-color: var(--cyan); }
        .findings-list { display: flex; flex-direction: column; gap: 2px; }
        .finding-item {
          background: var(--bg-primary); border: 1px solid var(--border);
          border-radius: 4px; overflow: hidden;
        }
        .finding-header {
          display: flex; align-items: center; gap: 8px;
          padding: 8px 12px; cursor: pointer; font-family: var(--font-mono); font-size: 10px;
        }
        .finding-header:hover { background: var(--bg-elevated); }
        .finding-sev {
          font-size: 8px; font-weight: 700; letter-spacing: 0.5px;
          padding: 2px 5px; border-radius: 2px; min-width: 36px; text-align: center;
        }
        .sev-critical .finding-sev { background: var(--red-dim); color: var(--red); }
        .sev-high .finding-sev { background: var(--amber-dim); color: var(--amber); }
        .sev-medium .finding-sev { background: var(--cyan-dim); color: var(--cyan); }
        .sev-low .finding-sev { background: var(--green-dim); color: var(--green); }
        .sev-info .finding-sev { background: var(--bg-elevated); color: var(--text-dim); }
        .finding-agent { color: var(--text-secondary); min-width: 56px; text-transform: uppercase; }
        .finding-type { color: var(--text-dim); min-width: 48px; }
        .finding-title { flex: 1; color: var(--text-primary); }
        .sev-critical .finding-title { color: var(--red); }
        .sev-high .finding-title { color: var(--amber); }
        .finding-expand { color: var(--text-dim); font-size: 8px; }
        .finding-details {
          padding: 10px 12px; border-top: 1px solid var(--border);
          background: var(--bg-elevated);
        }
        .finding-details pre {
          font-family: var(--font-code); font-size: 11px; line-height: 1.5;
          color: var(--text-secondary); white-space: pre-wrap; margin: 0;
        }
        .rules-list { display: flex; flex-direction: column; gap: 8px; }
        .empty-msg { color: var(--text-dim); font-family: var(--font-mono); font-size: 11px; padding: 20px; text-align: center; }
        .rule-item { background: var(--bg-primary); border: 1px solid var(--border); border-radius: 6px; overflow: hidden; }
        .rule-header { display: flex; align-items: center; justify-content: space-between; padding: 8px 12px; border-bottom: 1px solid var(--border); }
        .rule-type { font-family: var(--font-mono); font-size: 9px; font-weight: 700; letter-spacing: 1px; padding: 2px 8px; border-radius: 3px; }
        .rule-type.sigma { background: var(--cyan-dim); color: var(--cyan); }
        .rule-type.yara { background: var(--amber-dim); color: var(--amber); }
        .btn-copy {
          background: var(--bg-elevated); border: 1px solid var(--border);
          color: var(--text-secondary); padding: 3px 10px; border-radius: 3px;
          cursor: pointer; font-family: var(--font-mono); font-size: 9px;
          transition: all 0.2s;
        }
        .btn-copy:hover { background: var(--bg-hover); color: var(--cyan); border-color: var(--cyan-dim); }
        .rule-content {
          padding: 12px; font-family: var(--font-code); font-size: 11px;
          line-height: 1.5; color: var(--text-secondary); white-space: pre-wrap;
          margin: 0; max-height: 200px; overflow-y: auto;
        }
        .report-raw {
          font-family: var(--font-code); font-size: 11px; line-height: 1.6;
          color: var(--text-secondary); background: var(--bg-primary);
          padding: 14px; border-radius: 6px; border: 1px solid var(--border);
          overflow-x: auto; white-space: pre-wrap; max-height: 400px; overflow-y: auto;
        }
        .detail-actions {
          display: flex; gap: 8px; padding-top: 8px;
          border-top: 1px solid var(--border);
        }
        .action-btn {
          display: flex; align-items: center; gap: 6px;
          background: var(--bg-elevated); border: 1px solid var(--border);
          color: var(--text-primary); padding: 8px 14px; border-radius: 5px;
          cursor: pointer; font-family: var(--font-mono); font-size: 10px;
          letter-spacing: 0.3px; transition: all 0.2s;
        }
        .action-btn:hover { background: var(--bg-hover); border-color: var(--cyan-dim); color: var(--cyan); }
        .action-btn.danger { color: var(--red); border-color: var(--red-dim); }
        .action-btn.danger:hover { background: var(--red-dim); }
        .action-spacer { flex: 1; }
        .delete-confirm { display: flex; align-items: center; gap: 8px; }
        .delete-confirm span { font-family: var(--font-mono); font-size: 10px; color: var(--red); }
      `}</style>
    </div>
  )
}
