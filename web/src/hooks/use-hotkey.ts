import { useEffect } from 'react'

export function isMac(): boolean {
  return typeof navigator !== 'undefined' && /Mac|iPod|iPhone|iPad/.test(navigator.platform)
}

export function useHotkey(key: string, callback: () => void) {
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      const modifier = isMac() ? e.metaKey : e.ctrlKey
      if (modifier && e.key === key) {
        e.preventDefault()
        callback()
      }
    }

    document.addEventListener('keydown', handler)
    return () => document.removeEventListener('keydown', handler)
  }, [key, callback])
}
