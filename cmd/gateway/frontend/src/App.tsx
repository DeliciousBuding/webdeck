import { useState, useRef, type MouseEvent as ReactMouseEvent } from 'react'
import { useWebSocket, useFPS, useKeyboard } from './hooks'

const W = 1280, H = 720

// ── Coordinate math ──
function toGame(e: ReactMouseEvent, el: HTMLElement) {
  const r = el.getBoundingClientRect()
  const scale = Math.min(r.width / W, r.height / H)
  const dw = W * scale, dh = H * scale
  const ox = (r.width - dw) / 2, oy = (r.height - dh) / 2
  return {
    x: Math.round(Math.max(0, Math.min(W, (e.clientX - r.left - ox) / scale))),
    y: Math.round(Math.max(0, Math.min(H, (e.clientY - r.top - oy) / scale))),
  }
}

// ── StatusBar ──
function StatusBar({ connected, fps }: { connected: boolean; fps: number }) {
  return (
    <header className="bar">
      <span className="brand">webdeck</span>
      <span className={`dot ${connected ? 'live' : ''}`} />
      <span className="stat"><strong>{fps || '--'}</strong> fps</span>
      <span className="sep" />
      <span className="stat">1280×720</span>
      <span className="muted">click to control</span>
    </header>
  )
}

// ── StreamView ──
function StreamView() {
  const [swiping, setSwiping] = useState(false)
  const [pos, setPos] = useState('')
  const [swStart, setSwStart] = useState<{ x: number; y: number } | null>(null)
  const [ripples, setRipples] = useState<{ id: number; x: number; y: number }[]>([])
  const wrapRef = useRef<HTMLDivElement>(null)
  const { connected, send } = useWebSocket()
  const fps = useFPS()
  useKeyboard(send)
  const n = useRef(0)

  const handleClick = (e: ReactMouseEvent) => {
    if (swiping) return
    const g = toGame(e, wrapRef.current!)
    if (g.x < 0 || g.y < 0 || g.x > W || g.y > H) return
    setPos(`${g.x}, ${g.y}`)
    const id = ++n.current
    const rect = wrapRef.current!.getBoundingClientRect()
    setRipples(p => [...p.slice(-5), { id, x: e.clientX - rect.left, y: e.clientY - rect.top }])
    setTimeout(() => setRipples(p => p.filter(r => r.id !== id)), 400)
    send({ type: 'click', x: g.x, y: g.y })
  }

  const handleDown = (e: ReactMouseEvent) => { if (swiping) setSwStart(toGame(e, wrapRef.current!)) }
  const handleUp = (e: ReactMouseEvent) => {
    if (!swiping || !swStart) return
    const g = toGame(e, wrapRef.current!)
    if (Math.abs(g.x - swStart.x) < 6 && Math.abs(g.y - swStart.y) < 6) {
      send({ type: 'click', x: g.x, y: g.y })
    } else {
      send({ type: 'swipe', x1: swStart.x, y1: swStart.y, x2: g.x, y2: g.y })
    }
    setSwStart(null)
  }

  return (
    <>
      <StatusBar connected={connected} fps={fps} />
      <main ref={wrapRef} className="viewport" onClick={handleClick} onMouseDown={handleDown} onMouseUp={handleUp}>
        <img id="stream" src="/stream" alt="game stream" />
        {ripples.map(r => <div key={r.id} className="ripple" style={{ left: r.x, top: r.y }} />)}
        <div className="controls">
          <span className="coords">{pos || '—'}</span>
          <button className={`btn ${swiping ? 'active' : ''}`} onClick={e => { e.stopPropagation(); setSwiping(s => !s) }}>
            {swiping ? 'Swipe ON' : 'Swipe'}
          </button>
          <button className="btn danger" onClick={e => { e.stopPropagation(); send({ type: 'dismiss' }) }}>Dismiss</button>
          <button className="btn" onClick={e => { e.stopPropagation(); send({ type: 'key', key: 'Escape' }) }}>Esc</button>
        </div>
      </main>
    </>
  )
}

export default StreamView
