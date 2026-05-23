import React, { useState, useCallback, useEffect, useRef } from 'react'
import { useSSE } from './hooks/useSSE'
import { startInvestigation, listInvestigations, getStats, loadApiKey } from './services/api'
import WarRoom from './components/WarRoom'
import ThreatRadar from './components/ThreatRadar'
import AgentPanel from './components/AgentPanel'
import EventStream from './components/EventStream'
import CommandInput from './components/CommandInput'
import ReportViewer from './components/ReportViewer'
import InvestigationDetail from './components/InvestigationDetail'
import AlertPanel from './components/AlertPanel'
import Settings from './components/Settings'

const AGENTS = [
  { id: 'commander', name: 'Commander', icon: '🧠', color: '#d500f9' },
  { id: 'recon', name: 'Recon', icon: '🔍', color: '#00e5ff' },
  { id: 'vuln', name: 'Vuln', icon: '🛡️', color: '#ffab00' },
  { id: 'malware', name: 'Malware', icon: '🦠', color: '#ff1744' },
  { id: 'threat_intel', name: 'Threat Intel', icon: '📡', color: '#00e676' },
  { id: 'news_intel', name: 'News Intel', icon: '📰', color: '#d500f9' },
]

export default function App() {
  const [events, setEvents] = useState([])
  const [agentStates, setAgentStates] = useState({})
  const [investigations, setInvestigations] = useState([])
  const [activeInvestigation, setActiveInvestigation] = useState(null)
  const [report, setReport] = useState(null)
  const [stats, setStats] = useState({})
  const [radarBlips, setRadarBlips] = useState([])
  const [commanderMessages, setCommanderMessages] = useState([])
  const [showSettings, setShowSettings] = useState(false)
  const [showDetail, setShowDetail] = useState(null)
  const [showAlerts, setShowAlerts] = useState(false)
  const eventCounter = useRef(0)

  const handleSSE = useCallback((type, data) => {
    const timestamp = new Date().toLocaleTimeString('en-US', { hour12: false })
    eventCounter.current++

    switch (type) {
      case 'agent-status':
        setAgentStates(prev => ({
          ...prev,
          [data.agent]: {
            state: data.state,
            progress: data.progress,
            message: data.message,
          }
        }))
        break

      case 'finding':
        setEvents(prev => [{
          id: eventCounter.current,
          time: timestamp,
          agent: data.agent,
          type: data.type,
          severity: data.severity,
          title: data.title,
          details: data.details,
        }, ...prev].slice(0, 200))

        setRadarBlips(prev => [...prev, {
          id: eventCounter.current,
          severity: data.severity,
          agent: data.agent,
          title: data.title,
          angle: Math.random() * 360,
          distance: data.severity === 'critical' ? 0.15 : data.severity === 'high' ? 0.35 : data.severity === 'medium' ? 0.55 : 0.75,
        }].slice(-100))
        break

      case 'commander-message':
        setCommanderMessages(prev => [{
          id: eventCounter.current,
          time: timestamp,
          message: data.message,
          type: data.type,
        }, ...prev].slice(0, 50))
        break

      case 'investigation-status':
        if (data.status === 'complete' || data.status === 'error') {
          loadInvestigations()
          loadStats()
        }
        break

      case 'report-ready':
        setReport(data)
        break

      case 'alert':
        setEvents(prev => [{
          id: eventCounter.current,
          time: timestamp,
          agent: 'system',
          type: 'alert',
          severity: data.severity,
          title: `[ALERT] ${data.message}`,
          details: `Target: ${data.target}`,
        }, ...prev].slice(0, 200))
        break
    }
  }, [])

  useSSE(handleSSE)

  const loadInvestigations = async () => {
    try {
      const data = await listInvestigations()
      setInvestigations(Array.isArray(data) ? data : [])
    } catch (e) {
      console.error('Failed to load investigations:', e)
    }
  }

  const loadStats = async () => {
    try {
      const data = await getStats()
      setStats(data || {})
    } catch (e) {
      console.error('Failed to load stats:', e)
    }
  }

  useEffect(() => {
    loadApiKey()
    loadInvestigations()
    loadStats()
  }, [])

  const handleInvestigate = async (target) => {
    try {
      setAgentStates({})
      setEvents([])
      setRadarBlips([])
      setCommanderMessages([])
      setReport(null)
      setActiveInvestigation({ target, status: 'running' })

      const result = await startInvestigation(target)
      setActiveInvestigation(result)
      loadInvestigations()
    } catch (e) {
      console.error('Investigation failed:', e)
    }
  }

  const activeAgentStates = AGENTS.map(agent => ({
    ...agent,
    ...(agentStates[agent.id] || { state: 'idle', progress: 0, message: 'Standing by' }),
  }))

  return (
    <WarRoom
      stats={stats}
      investigations={investigations}
      activeInvestigation={activeInvestigation}
      onSelectInvestigation={(inv) => {
        setActiveInvestigation(inv)
        setReport(null)
        if (inv.id) setShowDetail(inv.id)
      }}
      onOpenSettings={() => setShowSettings(true)}
      onOpenAlerts={() => setShowAlerts(true)}
    >
      <div className="warroom-grid">
        <div className="warroom-top">
          <div className="panel radar-panel">
            <div className="panel-header">
              <span className="panel-badge">SENSOR</span>
              <span className="panel-title">THREAT RADAR</span>
              {activeInvestigation?.status === 'running' && (
                <span className="live-badge">LIVE</span>
              )}
            </div>
            <ThreatRadar blips={radarBlips} />
          </div>

          <div className="panel stream-panel">
            <div className="panel-header">
              <span className="panel-badge">STREAM</span>
              <span className="panel-title">EVENT LOG</span>
              <span className="event-count">{events.length}</span>
            </div>
            <EventStream events={events} />
          </div>
        </div>

        <div className="warroom-agents">
          {activeAgentStates.map(agent => (
            <AgentPanel key={agent.id} agent={agent} />
          ))}
        </div>

        <div className="warroom-bottom">
          <div className="commander-output">
            {commanderMessages.length > 0 ? (
              commanderMessages.slice(0, 3).map(msg => (
                <div key={msg.id} className={`cmd-msg cmd-msg-${msg.type}`}>
                  <svg className="cmd-icon" width="14" height="14" viewBox="0 0 14 14" fill="none">
                    <path d="M7 1L13 4.5V9.5L7 13L1 9.5V4.5L7 1Z" stroke="currentColor" strokeWidth="1.2" fill="none" />
                    <circle cx="7" cy="7" r="1" fill="currentColor" />
                  </svg>
                  <span className="cmd-label">CMD</span>
                  <span className="cmd-text">"{msg.message}"</span>
                </div>
              ))
            ) : (
              <div className="cmd-msg cmd-msg-idle">
                <svg className="cmd-icon" width="14" height="14" viewBox="0 0 14 14" fill="none">
                  <path d="M7 1L13 4.5V9.5L7 13L1 9.5V4.5L7 1Z" stroke="currentColor" strokeWidth="1.2" fill="none" />
                  <circle cx="7" cy="7" r="1" fill="currentColor" />
                </svg>
                <span className="cmd-label">CMD</span>
                <span className="cmd-text">"Awaiting orders. Enter target to initiate reconnaissance."</span>
              </div>
            )}
          </div>
          <CommandInput onInvestigate={handleInvestigate} />
        </div>
      </div>

      {report && (
        <ReportViewer report={report} onClose={() => setReport(null)} />
      )}

      {showDetail && (
        <InvestigationDetail
          investigationId={showDetail}
          onClose={() => setShowDetail(null)}
          onDeleted={(id) => {
            loadInvestigations()
            loadStats()
            if (activeInvestigation?.id === id) setActiveInvestigation(null)
          }}
        />
      )}

      {showSettings && (
        <Settings onClose={() => setShowSettings(false)} />
      )}

      {showAlerts && (
        <AlertPanel onClose={() => {
          setShowAlerts(false)
          loadStats()
        }} />
      )}

      <style>{`
        .warroom-grid {
          display: flex;
          flex-direction: column;
          height: 100%;
          gap: 10px;
          padding: 12px;
          position: relative;
          z-index: 1;
        }
        .warroom-top {
          display: grid;
          grid-template-columns: 320px 1fr;
          gap: 10px;
          flex: 1;
          min-height: 0;
        }
        .warroom-agents {
          display: grid;
          grid-template-columns: repeat(6, 1fr);
          gap: 8px;
        }
        .warroom-bottom {
          display: flex;
          flex-direction: column;
          gap: 6px;
        }

        .panel {
          background: var(--bg-surface);
          border: 1px solid var(--border);
          border-radius: 8px;
          overflow: hidden;
          display: flex;
          flex-direction: column;
          transition: border-color 0.3s;
        }
        .panel:hover {
          border-color: var(--border-hover);
        }
        .panel-header {
          display: flex;
          align-items: center;
          gap: 8px;
          padding: 8px 12px;
          font-family: var(--font-mono);
          font-size: 9px;
          font-weight: 600;
          letter-spacing: 1.5px;
          color: var(--text-secondary);
          border-bottom: 1px solid var(--border);
          background: var(--bg-elevated);
        }
        .panel-badge {
          background: var(--bg-primary);
          color: var(--text-dim);
          padding: 2px 6px;
          border-radius: 3px;
          font-size: 8px;
          letter-spacing: 1px;
        }
        .panel-title {
          flex: 1;
          color: var(--text-primary);
        }
        .live-badge {
          background: var(--red-dim);
          color: var(--red);
          padding: 2px 8px;
          border-radius: 3px;
          font-size: 8px;
          animation: agent-breathe 1.5s ease-in-out infinite;
          --agent-color: rgba(255, 23, 68, 0.3);
        }
        .event-count {
          background: var(--bg-primary);
          color: var(--text-dim);
          padding: 2px 6px;
          border-radius: 3px;
          font-size: 9px;
        }

        .radar-panel { min-height: 260px; }
        .stream-panel { min-height: 260px; }

        .commander-output {
          background: var(--bg-surface);
          border: 1px solid var(--border);
          border-radius: 8px;
          padding: 10px 14px;
          min-height: 48px;
          max-height: 72px;
          overflow-y: auto;
        }
        .cmd-msg {
          display: flex;
          align-items: center;
          gap: 8px;
          font-family: var(--font-mono);
          font-size: 11px;
          line-height: 1.5;
          animation: event-slide-in 0.3s ease;
        }
        .cmd-icon { color: var(--purple); flex-shrink: 0; }
        .cmd-label {
          color: var(--purple);
          font-weight: 700;
          letter-spacing: 1px;
          font-size: 10px;
          flex-shrink: 0;
        }
        .cmd-text {
          color: var(--text-primary);
          font-style: italic;
        }
        .cmd-msg-idle .cmd-text {
          color: var(--text-dim);
        }
        .cmd-msg-idle .cmd-icon,
        .cmd-msg-idle .cmd-label {
          opacity: 0.5;
        }

        @media (max-width: 1200px) {
          .warroom-agents { grid-template-columns: repeat(3, 1fr); }
        }
        @media (max-width: 800px) {
          .warroom-top { grid-template-columns: 1fr; }
          .warroom-agents { grid-template-columns: repeat(2, 1fr); }
        }
      `}</style>
    </WarRoom>
  )
}