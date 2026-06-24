import { useEffect, useRef, useState, useCallback } from 'react'

type Cmd =
  | { type: 'click'; x: number; y: number }
  | { type: 'swipe'; x1: number; y1: number; x2: number; y2: number }
  | { type: 'key'; key: string }
  | { type: 'dismiss' }
  | { type: 'ping' }

export function useWebSocket() {
  const wsRef = useRef<WebSocket | null>(null)
  const reconnectRef = useRef<ReturnType<typeof setTimeout> | undefined>(undefined)
  const [connected, setConnected] = useState(false)

  const connect = useCallback(() => {
    try { wsRef.current?.close() } catch { /* */ }
    const ws = new WebSocket(`ws://${location.host}/ws`)
    ws.onopen = (_ev: Event) => { wsRef.current = ws; setConnected(true) }
    ws.onclose = (_ev: Event) => { setConnected(false); wsRef.current = null; reconnectRef.current = setTimeout(connect, 2000) }
    ws.onerror = (_ev: Event) => ws.close()
  }, [])

  useEffect(() => {
    connect()
    return () => { clearTimeout(reconnectRef.current); wsRef.current?.close() }
  }, [connect])

  const send = useCallback((cmd: Cmd) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) wsRef.current.send(JSON.stringify(cmd))
  }, [])

  useEffect(() => {
    const t = setInterval(() => send({ type: 'ping' }), 15000)
    return () => clearInterval(t)
  }, [send])

  return { connected, send }
}
