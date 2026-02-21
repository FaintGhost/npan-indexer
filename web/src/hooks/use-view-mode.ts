import { useState, useCallback } from 'react'

export function useViewMode() {
  const [isDocked, setIsDocked] = useState(false)

  const setDocked = useCallback((docked: boolean) => {
    const update = () => setIsDocked(docked)

    if (typeof document !== 'undefined' && 'startViewTransition' in document) {
      ;(document as any).startViewTransition(update)
    } else {
      update()
    }
  }, [])

  return { isDocked, setDocked }
}
