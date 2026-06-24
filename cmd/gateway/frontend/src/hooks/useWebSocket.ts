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
  const [frameUrl, setFrameUrl] = useState<string>('')
  const [fps, setFps] = useState(0)
  const frames = useRef(0)
  const lastFps = useRef(performance.now())
  const frameObjUrl = useRef<string>('')

  const connect = useCallback(() => {
    try { wsRef.current?.close() } catch { /* */ }
    // Release previous blob URL
    if (frameObjUrl.current) URL.revokeObjectURL(frameObjUrl.current)

    const ws = new WebSocket(`ws://${location.host}/ws`)
    ws.binaryType = 'blob'

    ws.onopen = (_ev: Event) => { wsRef.current = ws; setConnected(true) }

    ws.onmessage = (ev: MessageEvent) => {
      if (ev.data instanceof Blob) {
        // New frame from hub
        if (frameObjUrl.current) URL.revokeObjectURL(frameObjUrl.current)
        frameObjUrl.current = URL.createObjectURL(ev.data)
        setFrameUrl(frameObjUrl.current)

        // FPS counting
        frames.current++
        const n = performance.now()
        if (n - lastFps.current >= 1000) {
          setFps(Math.round(frames.current / ((n - lastFps.current) / 1000)))
          frames.current = 0
          lastFps.current = n
        }
      }
      // Text messages (commands) are handled by the send side only
    }

    ws.onclose = (_ev: Event) => {
      setConnected(false)
      wsRef.current = null
      reconnectRef.current = setTimeout(connect, 2000)
    }
    ws.onerror = (_ev: Event) => ws.close()
  }, [])

  useEffect(() => {
    connect()
    return () => {
      clearTimeout(reconnectRef.current)
      wsRef.current?.close()
      if (frameObjUrl.current) URL.revokeObjectURL(frameObjUrl.current)
    }
  }, [connect])

  const send = useCallback((cmd: Cmd) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) wsRef.current.send(JSON.stringify(cmd))
  }, [])

  useEffect(() => {
    const t = setInterval(() => send({ type: 'ping' }), 15000)
    return () => clearInterval(t)
  }, [send])

  return { connected, send, frameUrl, fps }
}
