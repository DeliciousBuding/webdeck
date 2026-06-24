import { useEffect, useRef, useState } from 'react'

export function useFPS(): number {
  const [fps, setFps] = useState(0)
  const frames = useRef(0)
  const last = useRef(performance.now())

  useEffect(() => {
    const img = document.querySelector<HTMLImageElement>('#stream')
    if (!img) return
    const onLoad = () => {
      frames.current++
      const n = performance.now()
      if (n - last.current >= 1000) {
        setFps(Math.round(frames.current / ((n - last.current) / 1000)))
        frames.current = 0
        last.current = n
      }
    }
    img.addEventListener('load', onLoad)
    return () => img.removeEventListener('load', onLoad)
  }, [])

  return fps
}
