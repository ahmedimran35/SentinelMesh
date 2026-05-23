import React, { useState, useEffect } from 'react'
import { getInvestigation, getFindings } from '../services/api'
import { riskColors } from '../utils/severity'

export default function ReportViewer({ report, onClose }) {
  const [inv, setInv] = useState(null)
  const [findings, setFindings] = useState([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (!report?.id) return
    setLoading(true)
    Promise.all([
      getInvestigation(report.id),
      getFindings(report.id),
    ])
      .then(([invData, findingsData]) => {
        setInv(invData)
        setFindings(Array.isArray(findingsData) ? findingsData : [])
      })
      .catch(err => console.error('Failed to load report data:', err))
      .finally(() => setLoading(false))
  }, [report?.id])

  const rc = riskColors[inv?.risk_rating || report?.risk_rating] || 'var(--text-secondary)'

  return (
    <div className="report-overlay" onClick={onClose}>
      <div className="report-modal" onClick={(e) => e.stopPropagation()}>
        <div className="report-head">
          <div className="report-title">
            <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
              <rect x="1" y="2" width="14" height="12" rx="2" stroke="currentColor" strokeWidth="1.2" fill="none" />
              <line x1="4" y1="6" x2="12" y2="6" stroke="currentColor" strokeWidth="1.2" />
              <line x1="4" y1="9" x2="10" y2="9" stroke="currentColor" strokeWidth="1.2" />
            </svg>
            INTELLIGENCE REPORT
          </div>
          <button className="report-close" onClick={onClose}>✕</button>
        </div>

        <div className="report-body">
          {loading ? (
            <div className="report-loading">Loading report data...</div>
          ) : (
            <>
              <div className="report-meta">
                <div className="meta-cell">
                  <span className="meta-lbl">Investigation</span>
                  <span className="meta-val mono">{inv?.id || report?.id}</span>
                </div>
                <div className="meta-cell">
                  <span className="meta-lbl">Target</span>
                  <span className="meta-val mono glow-cyan">{inv?.target || report?.target || report?.id}</span>
                </div>
                <div className="meta-cell">
                  <span className="meta-lbl">Risk</span>
                  <span className="meta-risk" style={{ color: rc, borderColor: rc }}>
                    {(inv?.risk_rating || report?.risk_rating || 'unknown').toUpperCase()}
                  </span>
                </div>
                <div className="meta-cell">
                  <span className="meta-lbl">Findings</span>
                  <span className="meta-val">{findings.length}</span>
                </div>
              </div>

              <div className="report-section">
                <h4 className="section-head">EXECUTIVE SUMMARY</h4>
                <p className="section-text">
                  {inv?.executive_summary || report?.summary || 'No executive summary available.'}
                </p>
              </div>

              {findings.length > 0 && (
                <div className="report-section">
                  <h4 className="section-head">FINDINGS ({findings.length})</h4>
                  <div className="findings-table">
                    {findings.map(f => (
                      <div key={f.id} className={`finding-row sev-${f.severity}`}>
                        <span className="finding-sev">{f.severity.toUpperCase()}</span>
                        <span className="finding-agent">{f.agent}</span>
                        <span className="finding-title">{f.title}</span>
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {inv?.full_report && (
                <div className="report-section">
                  <h4 className="section-head">FULL ANALYSIS</h4>
                  <pre className="report-raw">{inv.full_report}</pre>
                </div>
              )}

              <div className="report-actions">
                <button className="action-btn" onClick={() => {
                  const blob = new Blob([JSON.stringify({ investigation: inv, findings }, null, 2)], { type: 'application/json' })
                  const url = URL.createObjectURL(blob)
                  const a = document.createElement('a')
                  a.href = url
                  a.download = `sentinelmesh-${report.id}.json`
                  a.click()
                  URL.revokeObjectURL(url)
                }}>
                  <svg width="12" height="12" viewBox="0 0 12 12" fill="none">
                    <path d="M6 1V8M3 5L6 8L9 5" stroke="currentColor" strokeWidth="1.2" strokeLinecap="round" strokeLinejoin="round" />
                    <path d="M1 9V10.5C1 10.8 1.2 11 1.5 11H10.5C10.8 11 11 10.8 11 10.5V9" stroke="currentColor" strokeWidth="1.2" />
                  </svg>
                  Export JSON
                </button>
              </div>
            </>
          )}
        </div>
      </div>

      <style>{`
        .report-overlay {
          position: fixed;
          top: 0; left: 0; right: 0; bottom: 0;
          background: rgba(0, 0, 0, 0.8);
          backdrop-filter: blur(6px);
          display: flex;
          align-items: center;
          justify-content: center;
          z-index: 100;
          animation: fade-in 0.2s ease;
        }
        .report-modal {
          background: var(--bg-surface);
          border: 1px solid var(--border);
          border-radius: 10px;
          width: 680px;
          max-width: 92vw;
          max-height: 85vh;
          overflow: hidden;
          display: flex;
          flex-direction: column;
          box-shadow: 0 24px 80px rgba(0, 0, 0, 0.6);
        }
        .report-head {
          display: flex;
          align-items: center;
          justify-content: space-between;
          padding: 14px 20px;
          border-bottom: 1px solid var(--border);
          background: var(--bg-elevated);
        }
        .report-title {
          display: flex;
          align-items: center;
          gap: 10px;
          font-family: var(--font-mono);
          font-size: 12px;
          font-weight: 600;
          letter-spacing: 2px;
          color: var(--text-primary);
        }
        .report-close {
          background: transparent;
          border: 1px solid var(--border);
          color: var(--text-secondary);
          width: 28px;
          height: 28px;
          border-radius: 5px;
          cursor: pointer;
          font-size: 13px;
          display: flex;
          align-items: center;
          justify-content: center;
          transition: all 0.2s;
        }
        .report-close:hover {
          background: var(--bg-hover);
          color: var(--cyan);
          border-color: var(--cyan-dim);
        }
        .report-body {
          padding: 20px;
          overflow-y: auto;
          display: flex;
          flex-direction: column;
          gap: 18px;
        }
        .report-loading {
          color: var(--text-dim);
          font-family: var(--font-mono);
          font-size: 12px;
          text-align: center;
          padding: 40px 0;
        }
        .report-meta {
          display: grid;
          grid-template-columns: repeat(4, 1fr);
          gap: 12px;
        }
        .meta-cell {
          display: flex;
          flex-direction: column;
          gap: 3px;
        }
        .meta-lbl {
          font-family: var(--font-mono);
          font-size: 8px;
          letter-spacing: 1.5px;
          color: var(--text-dim);
          text-transform: uppercase;
        }
        .meta-val {
          font-size: 12px;
          color: var(--text-primary);
        }
        .meta-val.mono {
          font-family: var(--font-mono);
          font-size: 10px;
        }
        .meta-risk {
          font-family: var(--font-mono);
          font-weight: 700;
          font-size: 13px;
          padding: 2px 8px;
          border: 1px solid;
          border-radius: 3px;
          display: inline-block;
          width: fit-content;
        }
        .report-section { display: flex; flex-direction: column; gap: 8px; }
        .section-head {
          font-family: var(--font-mono);
          font-size: 10px;
          letter-spacing: 1.5px;
          color: var(--text-secondary);
          font-weight: 600;
        }
        .section-text {
          font-size: 13px;
          line-height: 1.7;
          color: var(--text-primary);
        }
        .findings-table {
          display: flex;
          flex-direction: column;
          gap: 2px;
          background: var(--bg-primary);
          border-radius: 6px;
          overflow: hidden;
          border: 1px solid var(--border);
        }
        .finding-row {
          display: flex;
          align-items: center;
          gap: 8px;
          padding: 6px 10px;
          font-family: var(--font-mono);
          font-size: 10px;
          border-bottom: 1px solid var(--border);
        }
        .finding-row:last-child { border-bottom: none; }
        .finding-row:hover { background: var(--bg-elevated); }
        .finding-sev {
          font-size: 8px;
          font-weight: 700;
          letter-spacing: 0.5px;
          padding: 1px 4px;
          border-radius: 2px;
          min-width: 30px;
          text-align: center;
        }
        .sev-critical .finding-sev { background: var(--red-dim); color: var(--red); }
        .sev-high .finding-sev { background: var(--amber-dim); color: var(--amber); }
        .sev-medium .finding-sev { background: var(--cyan-dim); color: var(--cyan); }
        .sev-low .finding-sev { background: var(--green-dim); color: var(--green); }
        .finding-agent {
          color: var(--text-secondary);
          min-width: 56px;
          text-transform: uppercase;
        }
        .finding-title { color: var(--text-primary); }
        .sev-critical .finding-title { color: var(--red); }
        .sev-high .finding-title { color: var(--amber); }
        .sev-medium .finding-title { color: var(--cyan); }
        .report-raw {
          font-family: var(--font-code);
          font-size: 11px;
          line-height: 1.6;
          color: var(--text-secondary);
          background: var(--bg-primary);
          padding: 14px;
          border-radius: 6px;
          border: 1px solid var(--border);
          overflow-x: auto;
          white-space: pre-wrap;
          max-height: 300px;
          overflow-y: auto;
        }
        .report-actions {
          display: flex;
          gap: 8px;
        }
        .action-btn {
          display: flex;
          align-items: center;
          gap: 6px;
          background: var(--bg-elevated);
          border: 1px solid var(--border);
          color: var(--text-primary);
          padding: 8px 14px;
          border-radius: 5px;
          cursor: pointer;
          font-family: var(--font-mono);
          font-size: 10px;
          letter-spacing: 0.3px;
          transition: all 0.2s;
        }
        .action-btn:hover {
          background: var(--bg-hover);
          border-color: var(--cyan-dim);
          color: var(--cyan);
        }
      `}</style>
    </div>
  )
}