import React, { useState, useEffect } from 'react'
import { loadApiKey } from '../services/api'

export default function Settings({ onClose }) {
  const [settings, setSettings] = useState({})
  const [apiKey, setApiKey] = useState('')
  const [endpoint, setEndpoint] = useState('https://integrate.api.nvidia.com/v1/chat/completions')
  const [model, setModel] = useState('')
  const [provider, setProvider] = useState('nim')
  const [models, setModels] = useState([])
  const [loadingModels, setLoadingModels] = useState(false)
  const [testResult, setTestResult] = useState(null)
  const [testing, setTesting] = useState(false)
  const [saving, setSaving] = useState(false)
  const [saved, setSaved] = useState(false)

  const makeHeaders = () => {
    const h = { 'Content-Type': 'application/json' }
    const key = apiKey || loadApiKey()
    if (key) h['X-API-Key'] = key
    return h
  }

  useEffect(() => { loadSettings() }, [])

  const loadSettings = async () => {
    try {
      const res = await fetch('/api/settings', { headers: makeHeaders() })
      const data = await res.json()
      setSettings(data)
      if (data.nim_api_key) setApiKey(data.nim_api_key)
      if (data.nim_endpoint) setEndpoint(data.nim_endpoint)
      if (data.nim_model) setModel(data.nim_model)
      if (data.llm_provider) setProvider(data.llm_provider)
    } catch (e) { console.error('Failed to load settings:', e) }
  }

  const fetchModels = async () => {
    if (!apiKey) { setTestResult({ status: 'error', message: 'Enter API key first' }); return }
    setLoadingModels(true)
    setTestResult(null)
    try {
      const res = await fetch('/api/nim/models', {
        method: 'POST',
        headers: makeHeaders(),
        body: JSON.stringify({ api_key: apiKey, endpoint }),
      })
      const data = await res.json()
      if (data.error) setTestResult({ status: 'error', message: data.error })
      else { setModels(data.models || []); setTestResult({ status: 'ok', message: `Found ${data.count} models` }) }
    } catch (e) { setTestResult({ status: 'error', message: `Failed: ${e.message}` }) }
    finally { setLoadingModels(false) }
  }

  const testConnection = async () => {
    setTesting(true)
    setTestResult(null)
    try {
      const res = await fetch('/api/nim/test', {
        method: 'POST',
        headers: makeHeaders(),
        body: JSON.stringify({ api_key: apiKey, endpoint, model: model || 'meta/llama-3.1-8b-instruct' }),
      })
      setTestResult(await res.json())
    } catch (e) { setTestResult({ status: 'error', message: `Failed: ${e.message}` }) }
    finally { setTesting(false) }
  }

  const saveSettings = async () => {
    setSaving(true)
    setSaved(false)
    try {
      const toSave = {}
      if (apiKey) toSave.nim_api_key = apiKey
      if (endpoint) toSave.nim_endpoint = endpoint
      if (model) toSave.nim_model = model
      toSave.llm_provider = provider
      const res = await fetch('/api/settings', {
        method: 'POST',
        headers: makeHeaders(),
        body: JSON.stringify(toSave),
      })
      const data = await res.json()
      if (data.status === 'saved') { setSaved(true); setTimeout(() => setSaved(false), 3000) }
    } catch (e) { setTestResult({ status: 'error', message: `Save failed: ${e.message}` }) }
    finally { setSaving(false) }
  }

  return (
    <div className="set-overlay" onClick={onClose}>
      <div className="set-modal" onClick={(e) => e.stopPropagation()}>
        <div className="set-head">
          <span className="set-title">
            <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
              <circle cx="7" cy="7" r="2.5" stroke="currentColor" strokeWidth="1.2" />
              <path d="M7 .5V2.5M7 11.5V13.5M.5 7H2.5M11.5 7H13.5M2 2L3.5 3.5M10.5 10.5L12 12M2 12L3.5 10.5M10.5 3.5L12 2"
                stroke="currentColor" strokeWidth="1.2" strokeLinecap="round" />
            </svg>
            SETTINGS
          </span>
          <button className="set-close" onClick={onClose}>&#10005;</button>
        </div>

        <div className="set-body">
          <div className="set-group">
            <h4 className="set-label">AI PROVIDER</h4>
            <div className="provider-grid">
              <button className={`provider-opt ${provider === 'nim' ? 'active' : ''}`} onClick={() => setProvider('nim')}>
                <span className="provider-icon">&#9729;</span>
                <span className="provider-name">NVIDIA NIM</span>
                <span className="provider-desc">Cloud &mdash; requires API key</span>
              </button>
              <button className={`provider-opt ${provider === 'ollama' ? 'active' : ''}`} onClick={() => setProvider('ollama')}>
                <span className="provider-icon">&#9670;</span>
                <span className="provider-name">Ollama</span>
                <span className="provider-desc">Local &mdash; free, no key needed</span>
              </button>
            </div>
          </div>

          {provider === 'ollama' && (
            <div className="set-group">
              <h4 className="set-label">OLLAMA (LOCAL)</h4>
              <div className="info-box">
                <p>Ollama runs fully offline &mdash; no API key or internet required.</p>
                <p>Install: <code>curl -fsSL https://ollama.ai/install.sh | sh</code></p>
                <p>Pull: <code>ollama pull llama3.1</code></p>
              </div>
              <button className="btn-save" onClick={saveSettings} disabled={saving}>
                {saving ? 'Saving...' : saved ? 'Saved' : 'Save Settings'}
              </button>
            </div>
          )}

          {provider === 'nim' && (
            <>
              <div className="set-group">
                <h4 className="set-label">NVIDIA NIM</h4>
                <div className="field">
                  <label className="field-label">API Key</label>
                  <input type="password" className="field-input"
                    value={apiKey} onChange={(e) => setApiKey(e.target.value)}
                    placeholder="nvapi-..." spellCheck="false" />
                  {settings.nim_api_key_masked && (
                    <span className="field-hint">Saved: {settings.nim_api_key_masked}</span>
                  )}
                </div>
                <div className="field">
                  <label className="field-label">Endpoint</label>
                  <input type="text" className="field-input"
                    value={endpoint} onChange={(e) => setEndpoint(e.target.value)}
                    placeholder="https://integrate.api.nvidia.com/v1/chat/completions" spellCheck="false" />
                </div>
                <div className="field">
                  <label className="field-label">Model</label>
                  <div className="field-row">
                    <select className="field-input field-select"
                      value={model} onChange={(e) => setModel(e.target.value)}>
                      <option value="">&mdash; Select model &mdash;</option>
                      {models.map(m => <option key={m.id} value={m.id}>{m.id}</option>)}
                      {model && !models.find(m => m.id === model) && <option value={model}>{model}</option>}
                    </select>
                    <button className="btn-secondary" onClick={fetchModels} disabled={loadingModels || !apiKey}>
                      {loadingModels ? '\u231B' : '\u21BB'}
                    </button>
                  </div>
                </div>
                <div className="set-actions">
                  <button className="btn-test" onClick={testConnection} disabled={testing || !apiKey}>
                    {testing ? 'Testing...' : 'Test Connection'}
                  </button>
                  <button className="btn-save" onClick={saveSettings} disabled={saving}>
                    {saving ? 'Saving...' : saved ? 'Saved' : 'Save'}
                  </button>
                </div>
                {testResult && (
                  <div className={`test-box ${testResult.status}`}>
                    <span>{testResult.status === 'ok' ? '\u2713' : '\u2717'} {testResult.message}</span>
                    {testResult.latency_ms && <span className="test-detail">{testResult.latency_ms}ms</span>}
                    {testResult.response && <span className="test-detail">"{testResult.response}"</span>}
                  </div>
                )}
              </div>
            </>
          )}
        </div>

        <style>{`
          .set-overlay {
            position: fixed; top: 0; left: 0; right: 0; bottom: 0;
            background: rgba(0,0,0,0.85); backdrop-filter: blur(6px);
            display: flex; align-items: center; justify-content: center;
            z-index: 100; animation: fade-in 0.2s ease;
          }
          .set-modal {
            background: var(--bg-surface); border: 1px solid var(--border);
            border-radius: 10px; width: 520px; max-width: 92vw; max-height: 85vh;
            overflow: hidden; display: flex; flex-direction: column;
            box-shadow: 0 24px 80px rgba(0,0,0,0.6);
          }
          .set-head {
            display: flex; align-items: center; justify-content: space-between;
            padding: 14px 20px; border-bottom: 1px solid var(--border);
            background: var(--bg-elevated);
          }
          .set-title {
            display: flex; align-items: center; gap: 10px;
            font-family: var(--font-mono); font-size: 12px; font-weight: 600;
            letter-spacing: 2px; color: var(--text-primary);
          }
          .set-close {
            background: transparent; border: 1px solid var(--border);
            color: var(--text-secondary); width: 28px; height: 28px;
            border-radius: 5px; cursor: pointer; font-size: 13px;
            display: flex; align-items: center; justify-content: center;
            transition: all 0.2s;
          }
          .set-close:hover { background: var(--bg-hover); color: var(--cyan); border-color: var(--cyan-dim); }
          .set-body { padding: 20px; overflow-y: auto; display: flex; flex-direction: column; gap: 20px; }
          .set-group { display: flex; flex-direction: column; gap: 12px; }
          .set-label {
            font-family: var(--font-mono); font-size: 9px; letter-spacing: 1.5px;
            color: var(--cyan); font-weight: 600;
            padding-bottom: 6px; border-bottom: 1px solid var(--border);
          }
          .provider-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 10px; }
          .provider-opt {
            background: var(--bg-elevated); border: 2px solid var(--border);
            border-radius: 6px; padding: 14px; cursor: pointer;
            text-align: left; transition: all 0.2s;
            display: flex; flex-direction: column; gap: 4px;
          }
          .provider-opt:hover { border-color: var(--border-hover); background: var(--bg-hover); }
          .provider-opt.active { border-color: var(--cyan-dim); background: var(--cyan-dim); }
          .provider-icon { color: var(--text-secondary); font-size: 16px; }
          .provider-name { font-family: var(--font-mono); font-size: 12px; font-weight: 600; color: var(--text-primary); }
          .provider-desc { font-size: 10px; color: var(--text-dim); }
          .info-box {
            background: var(--bg-primary); border-radius: 6px;
            padding: 14px; display: flex; flex-direction: column; gap: 6px;
            font-size: 12px; color: var(--text-secondary); line-height: 1.6;
          }
          .info-box code {
            background: var(--bg-elevated); padding: 2px 5px; border-radius: 3px;
            font-family: var(--font-code); font-size: 11px; color: var(--cyan);
          }
          .field { display: flex; flex-direction: column; gap: 4px; }
          .field-label {
            font-family: var(--font-mono); font-size: 10px; letter-spacing: 1px;
            color: var(--text-secondary); text-transform: uppercase;
          }
          .field-input {
            background: var(--bg-primary); border: 1px solid var(--border);
            border-radius: 5px; padding: 8px 12px;
            font-family: var(--font-code); font-size: 12px;
            color: var(--text-primary); outline: none;
            transition: border-color 0.2s;
          }
          .field-input:focus { border-color: var(--cyan-dim); box-shadow: 0 0 8px var(--cyan-dim); }
          .field-input::placeholder { color: var(--text-dim); }
          .field-select { cursor: pointer; }
          .field-row { display: flex; gap: 6px; }
          .field-row .field-input { flex: 1; }
          .field-hint { font-family: var(--font-mono); font-size: 9px; color: var(--text-dim); }
          .set-actions { display: flex; gap: 8px; }
          .btn-test, .btn-save, .btn-secondary {
            font-family: var(--font-mono); font-size: 10px;
            padding: 8px 14px; border-radius: 5px; cursor: pointer;
            border: 1px solid var(--border); transition: all 0.2s;
            background: var(--bg-elevated); color: var(--text-primary);
          }
          .btn-test { background: var(--cyan-dim); border-color: var(--cyan); color: var(--cyan); }
          .btn-test:hover:not(:disabled) { background: rgba(0,229,255,0.25); }
          .btn-save { background: var(--green-dim); border-color: var(--green); color: var(--green); }
          .btn-save:hover:not(:disabled) { background: rgba(0,230,118,0.25); }
          .btn-secondary { padding: 8px 10px; font-size: 14px; }
          .btn-secondary:hover:not(:disabled) { background: var(--bg-hover); border-color: var(--cyan-dim); }
          .btn-test:disabled, .btn-save:disabled, .btn-secondary:disabled { opacity: 0.5; cursor: not-allowed; }
          .test-box {
            background: var(--bg-primary); border-radius: 6px; padding: 10px 14px;
            display: flex; flex-direction: column; gap: 4px;
            font-family: var(--font-mono); font-size: 10px;
            animation: fade-in 0.3s ease;
          }
          .test-box.ok { border: 1px solid rgba(0,230,118,0.3); color: var(--green); }
          .test-box.error { border: 1px solid rgba(255,23,68,0.3); color: var(--red); }
          .test-detail { color: var(--text-secondary); font-size: 9px; }
        `}</style>
      </div>
    </div>
  )
}
