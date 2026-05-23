import React, { useState, useRef, useEffect } from 'react'

export default function CommandInput({ onInvestigate }) {
  const [input, setInput] = useState('')
  const [history, setHistory] = useState([])
  const [historyIdx, setHistoryIdx] = useState(-1)
  const inputRef = useRef(null)

  useEffect(() => {
    inputRef.current?.focus()
  }, [])

  const handleSubmit = (e) => {
    e.preventDefault()
    const target = input.trim()
    if (!target) return
    setHistory(prev => [target, ...prev].slice(0, 50))
    setHistoryIdx(-1)
    onInvestigate(target)
    setInput('')
  }

  const handleKeyDown = (e) => {
    if (e.key === 'ArrowUp') {
      e.preventDefault()
      const ni = Math.min(historyIdx + 1, history.length - 1)
      setHistoryIdx(ni)
      if (history[ni]) setInput(history[ni])
    } else if (e.key === 'ArrowDown') {
      e.preventDefault()
      const ni = Math.max(historyIdx - 1, -1)
      setHistoryIdx(ni)
      setInput(ni >= 0 ? history[ni] : '')
    }
  }

  const detectType = (v) => {
    if (/^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\/\d{1,2}$/.test(v)) return 'CIDR'
    if (/^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$/.test(v)) return 'IP'
    return 'DOMAIN'
  }

  return (
    <form className="terminal-input" onSubmit={handleSubmit}>
      <span className="terminal-prompt">{'>'}_</span>
      <input
        ref={inputRef}
        type="text"
        className="terminal-field"
        value={input}
        onChange={(e) => setInput(e.target.value)}
        onKeyDown={handleKeyDown}
        placeholder="Enter target: domain, IP, or CIDR range..."
        spellCheck="false"
        autoComplete="off"
      />
      {input && (
        <span className="terminal-type">{detectType(input)}</span>
      )}
      <button type="submit" className="terminal-submit" disabled={!input.trim()}>
        SCAN
      </button>
      <style>{`
        .terminal-input {
          display: flex;
          align-items: center;
          gap: 10px;
          background: var(--bg-surface);
          border: 1px solid var(--border);
          border-radius: 4px;
          padding: 10px 14px;
          transition: border-color 0.3s, box-shadow 0.3s;
          position: relative;
        }
        .terminal-input:focus-within {
          border-color: var(--cyan);
          box-shadow: 0 0 16px rgba(0, 229, 255, 0.08), inset 0 0 30px rgba(0, 229, 255, 0.02);
        }
        .terminal-prompt {
          font-family: var(--font-display);
          font-size: 13px;
          color: var(--cyan);
          font-weight: 700;
          letter-spacing: 1px;
          opacity: 0.8;
        }
        .terminal-field {
          flex: 1;
          background: transparent;
          border: none;
          outline: none;
          font-family: var(--font-mono);
          font-size: 13px;
          color: var(--text-primary);
          caret-color: var(--cyan);
          letter-spacing: 0.5px;
        }
        .terminal-field::placeholder {
          color: var(--text-dim);
          font-size: 11px;
          letter-spacing: 0.3px;
        }
        .terminal-type {
          font-family: var(--font-display);
          font-size: 8px;
          color: var(--cyan);
          background: var(--cyan-dim);
          padding: 2px 8px;
          border-radius: 2px;
          letter-spacing: 2px;
          font-weight: 600;
        }
        .terminal-submit {
          font-family: var(--font-display);
          font-size: 9px;
          letter-spacing: 2px;
          color: var(--cyan);
          background: var(--cyan-dim);
          border: 1px solid rgba(0, 229, 255, 0.2);
          border-radius: 3px;
          padding: 4px 12px;
          cursor: pointer;
          transition: all 0.2s;
        }
        .terminal-submit:hover:not(:disabled) {
          background: rgba(0, 229, 255, 0.2);
          border-color: var(--cyan);
        }
        .terminal-submit:disabled {
          opacity: 0.3;
          cursor: not-allowed;
        }
      `}</style>
    </form>
  )
}
