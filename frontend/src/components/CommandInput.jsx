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
      <span className="terminal-prompt">❯</span>
      <input
        ref={inputRef}
        type="text"
        className="terminal-field"
        value={input}
        onChange={(e) => setInput(e.target.value)}
        onKeyDown={handleKeyDown}
        placeholder="investigate target.com, 192.168.1.1, or 10.0.0.0/24"
        spellCheck="false"
        autoComplete="off"
      />
      {input && (
        <span className="terminal-type">{detectType(input)}</span>
      )}
      <style>{`
        .terminal-input {
          display: flex;
          align-items: center;
          gap: 8px;
          background: var(--bg-surface);
          border: 1px solid var(--border);
          border-radius: 6px;
          padding: 10px 14px;
          transition: border-color 0.3s, box-shadow 0.3s;
        }
        .terminal-input:focus-within {
          border-color: var(--cyan-dim);
          box-shadow: 0 0 12px var(--cyan-dim);
        }
        .terminal-prompt {
          font-family: var(--font-mono);
          font-size: 14px;
          color: var(--cyan);
          font-weight: 700;
        }
        .terminal-field {
          flex: 1;
          background: transparent;
          border: none;
          outline: none;
          font-family: var(--font-code);
          font-size: 13px;
          color: var(--text-primary);
          caret-color: var(--cyan);
        }
        .terminal-field::placeholder {
          color: var(--text-dim);
          font-size: 12px;
        }
        .terminal-type {
          font-family: var(--font-mono);
          font-size: 8px;
          color: var(--text-dim);
          background: var(--bg-elevated);
          padding: 2px 6px;
          border-radius: 2px;
          letter-spacing: 1px;
          font-weight: 600;
        }
      `}</style>
    </form>
  )
}