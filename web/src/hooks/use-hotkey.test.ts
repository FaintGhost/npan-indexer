import { describe, it, expect, vi, afterEach } from 'vitest'
import { renderHook } from '@testing-library/react'
import { useHotkey } from './use-hotkey'

describe('useHotkey', () => {
  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('triggers callback on Cmd+K (Mac)', () => {
    Object.defineProperty(navigator, 'platform', {
      value: 'MacIntel',
      configurable: true,
    })

    const callback = vi.fn()
    renderHook(() => useHotkey('k', callback))

    const event = new KeyboardEvent('keydown', {
      key: 'k',
      metaKey: true,
      bubbles: true,
    })
    const spy = vi.spyOn(event, 'preventDefault')
    document.dispatchEvent(event)

    expect(callback).toHaveBeenCalledOnce()
    expect(spy).toHaveBeenCalled()
  })

  it('triggers callback on Ctrl+K (non-Mac)', () => {
    Object.defineProperty(navigator, 'platform', {
      value: 'Win32',
      configurable: true,
    })

    const callback = vi.fn()
    renderHook(() => useHotkey('k', callback))

    const event = new KeyboardEvent('keydown', {
      key: 'k',
      ctrlKey: true,
      bubbles: true,
    })
    document.dispatchEvent(event)

    expect(callback).toHaveBeenCalledOnce()
  })

  it('does not trigger without modifier key', () => {
    const callback = vi.fn()
    renderHook(() => useHotkey('k', callback))

    document.dispatchEvent(
      new KeyboardEvent('keydown', { key: 'k', bubbles: true }),
    )

    expect(callback).not.toHaveBeenCalled()
  })

  it('removes listener on unmount', () => {
    const callback = vi.fn()
    const { unmount } = renderHook(() => useHotkey('k', callback))

    unmount()

    document.dispatchEvent(
      new KeyboardEvent('keydown', {
        key: 'k',
        metaKey: true,
        bubbles: true,
      }),
    )

    expect(callback).not.toHaveBeenCalled()
  })
})
