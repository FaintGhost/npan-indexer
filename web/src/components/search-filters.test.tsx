import { describe, expect, it, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import { SearchFilters } from './search-filters'

const refineSpy = vi.fn()

vi.mock('react-instantsearch', () => ({
  useRefinementList: vi.fn(() => ({
    refine: refineSpy,
  })),
  useCurrentRefinements: vi.fn(() => ({
    items: [],
  })),
}))

describe('SearchFilters', () => {
  it('renders filter chips as a radio group for refinement-driven selection', () => {
    render(<SearchFilters />)

    expect(screen.getByRole('radiogroup', { name: '结果筛选' })).toBeInTheDocument()
    expect(screen.getByRole('radio', { name: '全部' })).toBeChecked()
    expect(screen.getByRole('radio', { name: '文档' })).not.toBeChecked()
  })
})
