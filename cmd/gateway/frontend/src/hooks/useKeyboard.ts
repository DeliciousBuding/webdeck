import { useEffect } from 'react'

type Cmd =
  | { type: 'click'; x: number; y: number }
  | { type: 'swipe'; x1: number; y1: number; x2: number; y2: number }
  | { type: 'key'; key: string }
  | { type: 'dismiss' }
  | { type: 'ping' }

export function useKeyboard(send: (cmd: Cmd) => void) {
  useEffect(() => {
    const h = (e: KeyboardEvent) => {
      if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) return
      send({ type: 'key', key: e.key })
      e.preventDefault()
    }
    document.addEventListener('keydown', h)
    return () => document.removeEventListener('keydown', h)
  }, [send])
}
