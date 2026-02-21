import { describe, it, expect, vi, afterEach } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { useViewMode } from './use-view-mode'

describe('useViewMode', () => {
  afterEach(() => {
    // Clean up any startViewTransition we may have added
    if ('startViewTransition' in document) {
      delete (document as any).startViewTransition
    }
  })

  it('starts in hero mode', () => {
    const { result } = renderHook(() => useViewMode())
    expect(result.current.isDocked).toBe(false)
  })

  it('switches to docked mode', () => {
    const { result } = renderHook(() => useViewMode())
    act(() => {
      result.current.setDocked(true)
    })
    expect(result.current.isDocked).toBe(true)
  })

  it('switches back to hero mode', () => {
    const { result } = renderHook(() => useViewMode())
    act(() => {
      result.current.setDocked(true)
    })
    act(() => {
      result.current.setDocked(false)
    })
    expect(result.current.isDocked).toBe(false)
  })

  it('calls startViewTransition when supported', () => {
    const mockTransition = vi.fn((cb: () => void) => cb())
    ;(document as any).startViewTransition = mockTransition

    const { result } = renderHook(() => useViewMode())
    act(() => {
      result.current.setDocked(true)
    })

    expect(mockTransition).toHaveBeenCalledOnce()
    expect(result.current.isDocked).toBe(true)
  })

  it('works without startViewTransition (graceful fallback)', () => {
    // jsdom doesn't have startViewTransition by default
    const { result } = renderHook(() => useViewMode())
    act(() => {
      result.current.setDocked(true)
    })
    expect(result.current.isDocked).toBe(true)
  })
})
