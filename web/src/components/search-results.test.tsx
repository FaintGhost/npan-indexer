import { describe, expect, it, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { SearchResults } from './search-results'

const showMoreSpy = vi.fn()

vi.mock('react-instantsearch', () => ({
  useInfiniteHits: vi.fn(() => ({
    items: [
      {
        doc_id: 'file_1',
        source_id: 1,
        type: 'file',
        name: 'report.pdf',
        path_text: '/docs/report.pdf',
        parent_id: 0,
        modified_at: 1700000000,
        created_at: 1700000000,
        size: 1024,
        sha1: 'abc',
        in_trash: false,
        is_deleted: false,
        _highlightResult: {
          name: {
            value: '<mark>report</mark>.pdf',
          },
        },
      },
      {
        doc_id: 'file_2',
        source_id: 2,
        type: 'file',
        name: 'manual.pdf',
        path_text: '/docs/manual.pdf',
        parent_id: 0,
        modified_at: 1700000001,
        created_at: 1700000001,
        size: 2048,
        sha1: 'def',
        in_trash: false,
        is_deleted: false,
      },
    ],
    isLastPage: false,
    showMore: showMoreSpy,
    status: 'idle',
  })),
  useStats: vi.fn(() => ({
    nbHits: 2,
  })),
  useInstantSearch: vi.fn(() => ({
    status: 'idle',
    error: undefined,
  })),
  useSearchBox: vi.fn(() => ({
    query: 'report',
  })),
}))

describe('SearchResults', () => {
  it('renders InstantSearch hits and result count text', () => {
    render(
      <SearchResults
        download={{
          getStatus: () => 'idle',
          download: () => {},
        }}
      />,
    )

    expect(screen.getByTitle('report.pdf')).toBeInTheDocument()
    expect(screen.getByText('manual.pdf')).toBeInTheDocument()
    expect(screen.getByText('已加载 2 / 2 个文件')).toBeInTheDocument()
    expect(document.querySelector('mark')?.textContent).toBe('report')
  })

  it('loads more hits from InfiniteHits when clicking the pagination control', async () => {
    render(
      <SearchResults
        download={{
          getStatus: () => 'idle',
          download: () => {},
        }}
      />,
    )

    const user = userEvent.setup()
    await user.click(screen.getByRole('button', { name: '加载更多结果' }))

    expect(showMoreSpy).toHaveBeenCalledTimes(1)
  })
})
