import { describe, expect, it, beforeEach, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { SearchFilters } from './search-filters'

const { refineSpy, useCurrentRefinementsMock } = vi.hoisted(() => ({
  refineSpy: vi.fn(),
  useCurrentRefinementsMock: vi.fn(),
}))

vi.mock('react-instantsearch', () => ({
  useRefinementList: vi.fn(() => ({
    refine: refineSpy,
  })),
  useCurrentRefinements: useCurrentRefinementsMock,
}))

describe('SearchFilters', () => {
  beforeEach(() => {
    refineSpy.mockReset()
    useCurrentRefinementsMock.mockReturnValue({
      items: [],
    })
  })

  it('renders filter chips as a radio group for refinement-driven selection', () => {
    render(<SearchFilters />)

    expect(screen.getByRole('radiogroup', { name: '结果筛选' })).toBeInTheDocument()
    expect(screen.getByRole('radio', { name: '全部' })).toBeChecked()
    expect(screen.getByRole('radio', { name: '文档' })).not.toBeChecked()
  })

  it('keeps file_category refinement additive so public baseline filters can stay in the request layer', async () => {
    useCurrentRefinementsMock.mockReturnValue({
      items: [],
    })
    render(<SearchFilters />)

    const user = userEvent.setup()
    await user.click(screen.getByRole('radio', { name: '文档' }))

    expect(refineSpy).toHaveBeenCalledWith('doc')
    expect(refineSpy).toHaveBeenCalledTimes(1)
  })

  it('only swaps file_category refinement values instead of resetting unrelated request filters', async () => {
    useCurrentRefinementsMock.mockReturnValue({
      items: [
        {
          attribute: 'file_category',
          refinements: [{ value: 'image' }],
        },
      ],
    })
    render(<SearchFilters />)

    const user = userEvent.setup()
    await user.click(screen.getByRole('radio', { name: '文档' }))

    expect(refineSpy).toHaveBeenNthCalledWith(1, 'image')
    expect(refineSpy).toHaveBeenNthCalledWith(2, 'doc')
    expect(refineSpy).toHaveBeenCalledTimes(2)
  })
})
