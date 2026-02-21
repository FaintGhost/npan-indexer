import { describe, it, expect } from 'vitest'
import { render } from '@testing-library/react'
import { SkeletonCard } from './skeleton-card'

describe('SkeletonCard', () => {
  it('renders with pulse animation class', () => {
    const { container } = render(<SkeletonCard />)
    const el = container.firstElementChild as HTMLElement
    expect(el.className).toContain('animate-pulse')
  })

  it('sets aria-hidden', () => {
    const { container } = render(<SkeletonCard />)
    expect(container.firstElementChild?.getAttribute('aria-hidden')).toBe('true')
  })

  it('renders multiple skeleton cards', () => {
    const { container } = render(
      <>
        {Array.from({ length: 5 }, (_, i) => (
          <SkeletonCard key={i} />
        ))}
      </>,
    )
    const cards = container.querySelectorAll('[aria-hidden="true"]')
    expect(cards).toHaveLength(5)
  })
})
