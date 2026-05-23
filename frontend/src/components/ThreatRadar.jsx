import React, { useEffect, useRef, useState } from 'react'

export default function ThreatRadar({ blips = [] }) {
  const [sweep, setSweep] = useState(0)
  const animRef = useRef()

  useEffect(() => {
    let a = 0
    let lastTime = 0
    const anim = (time) => {
      if (time - lastTime >= 33) { // ~30fps instead of 60fps
        a = (a + 1.2) % 360
        setSweep(a)
        lastTime = time
      }
      animRef.current = requestAnimationFrame(anim)
    }
    animRef.current = requestAnimationFrame(anim)
    return () => cancelAnimationFrame(animRef.current)
  }, [])

  const sevColor = (s) => {
    switch (s) {
      case 'critical': return 'var(--red)'
      case 'high': return 'var(--amber)'
      case 'medium': return 'var(--cyan)'
      case 'low': return 'var(--green)'
      default: return 'var(--text-dim)'
    }
  }

  const size = 260
  const cx = size / 2
  const cy = size / 2
  const mr = size / 2 - 18

  return (
    <div className="radar-box">
      <svg width="100%" height="100%" viewBox={`0 0 ${size} ${size}`}>
        <defs>
          <radialGradient id="sweep-grad" cx="50%" cy="50%" r="50%">
            <stop offset="0%" stopColor="var(--cyan)" stopOpacity="0.12" />
            <stop offset="100%" stopColor="var(--cyan)" stopOpacity="0" />
          </radialGradient>
        </defs>

        {[0.2, 0.4, 0.6, 0.8, 1].map((r, i) => (
          <circle key={i} cx={cx} cy={cy} r={mr * r} fill="none"
            stroke="var(--border)" strokeWidth="0.8" opacity={0.6 - i * 0.1} />
        ))}

        {[0, 45, 90, 135].map((a, i) => (
          <line key={i}
            x1={cx + mr * 0.2 * Math.cos((a * Math.PI) / 180)}
            y1={cy + mr * 0.2 * Math.sin((a * Math.PI) / 180)}
            x2={cx + mr * Math.cos((a * Math.PI) / 180)}
            y2={cy + mr * Math.sin((a * Math.PI) / 180)}
            stroke="var(--border)" strokeWidth="0.4" opacity="0.3" />
        ))}

        <line x1={cx} y1={cy - mr} x2={cx} y2={cy + mr}
          stroke="var(--border)" strokeWidth="0.4" opacity="0.2" />
        <line x1={cx - mr} y1={cy} x2={cx + mr} y2={cy}
          stroke="var(--border)" strokeWidth="0.4" opacity="0.2" />

        <line
          x1={cx} y1={cy}
          x2={cx + mr * Math.cos((sweep * Math.PI) / 180)}
          y2={cy + mr * Math.sin((sweep * Math.PI) / 180)}
          stroke="var(--cyan)" strokeWidth="1.2" opacity="0.7"
        />

        <path
          d={`M ${cx} ${cy} L ${cx + mr * Math.cos(((sweep - 25) * Math.PI) / 180)} ${cy + mr * Math.sin(((sweep - 25) * Math.PI) / 180)} A ${mr} ${mr} 0 0 1 ${cx + mr * Math.cos((sweep * Math.PI) / 180)} ${cy + mr * Math.sin((sweep * Math.PI) / 180)} Z`}
          fill="url(#sweep-grad)"
        />

        {blips.slice(-25).map((b, i) => {
          const a = (b.angle * Math.PI) / 180
          const d = b.distance * mr
          const x = cx + d * Math.cos(a)
          const y = cy + d * Math.sin(a)
          const fresh = i >= blips.length - 3

          return (
            <g key={b.id}>
              <circle cx={x} cy={y} r={fresh ? 4 : 2.5}
                fill={sevColor(b.severity)}
                opacity={fresh ? 1 : 0.5}>
                {fresh && <animate attributeName="r" values="2.5;5;2.5" dur="1.2s" repeatCount="indefinite" />}
              </circle>
              {fresh && (
                <circle cx={x} cy={y} r="8" fill="none"
                  stroke={sevColor(b.severity)} strokeWidth="0.8" opacity="0.3">
                  <animate attributeName="r" values="4;12;4" dur="1.8s" repeatCount="indefinite" />
                  <animate attributeName="opacity" values="0.3;0;0.3" dur="1.8s" repeatCount="indefinite" />
                </circle>
              )}
            </g>
          )
        })}

        <circle cx={cx} cy={cy} r="2.5" fill="var(--cyan)" opacity="0.9" />

        <text x={cx} y={12} textAnchor="middle" fill="var(--text-dim)" fontSize="7" fontFamily="var(--font-mono)">CRITICAL</text>
        <text x={cx} y={size - 4} textAnchor="middle" fill="var(--text-dim)" fontSize="7" fontFamily="var(--font-mono)">INFO</text>
      </svg>

      <div className="radar-footer">
        <span>{blips.length} signals</span>
      </div>

      <style>{`
        .radar-box {
          flex: 1;
          display: flex;
          flex-direction: column;
          align-items: center;
          justify-content: center;
          padding: 8px;
          position: relative;
        }
        .radar-footer {
          position: absolute;
          bottom: 6px;
          right: 10px;
          font-family: var(--font-mono);
          font-size: 9px;
          color: var(--text-dim);
        }
      `}</style>
    </div>
  )
}