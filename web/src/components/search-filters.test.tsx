import { describe, expect, it, beforeEach, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { SearchFilters } from './search-filters'

const { refineSpy, useCurrentRefinementsMock, useRefinementListMock } = vi.hoisted(() => ({
  refineSpy: vi.fn(),
  useCurrentRefinementsMock: vi.fn(),
  useRefinementListMock: vi.fn(),
}))

vi.mock('react-instantsearch', () => ({
  useRefinementList: useRefinementListMock,
  useCurrentRefinements: useCurrentRefinementsMock,
}))

describe('SearchFilters', () => {
  beforeEach(() => {
    refineSpy.mockReset()
    useRefinementListMock.mockReturnValue({
      refine: refineSpy,
      items: [
        { label: 'doc', value: 'file_category:doc-token', isRefined: false },
        { label: 'image', value: 'file_category:image-token', isRefined: false },
      ],
    })
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

  it('uses the real refinement token from refinementList items when selecting a category', async () => {
    useCurrentRefinementsMock.mockReturnValue({
      items: [],
    })
    render(<SearchFilters />)

    const user = userEvent.setup()
    await user.click(screen.getByRole('radio', { name: '文档' }))

    expect(refineSpy).toHaveBeenCalledWith('file_category:doc-token')
    expect(refineSpy).toHaveBeenCalledTimes(1)
  })

  it('swaps categories by clearing the previous token before applying the next token', async () => {
    useRefinementListMock.mockReturnValue({
      refine: refineSpy,
      items: [
        { label: 'doc', value: 'file_category:doc-token', isRefined: false },
        { label: 'image', value: 'file_category:image-token', isRefined: true },
      ],
    })
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

    expect(refineSpy).toHaveBeenNthCalledWith(1, 'file_category:image-token')
    expect(refineSpy).toHaveBeenNthCalledWith(2, 'file_category:doc-token')
    expect(refineSpy).toHaveBeenCalledTimes(2)
  })
})
