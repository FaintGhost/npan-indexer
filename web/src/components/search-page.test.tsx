import { describe, it, expect } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'
import { server } from '../tests/mocks/server'
import { SearchPage } from '../routes/app.index.lazy'

// Helper to create search response
function makeSearchResponse(count: number, total: number) {
  return {
    items: Array.from({ length: count }, (_, i) => ({
      doc_id: `f${i + 1}`,
      source_id: i + 1,
      type: 'file',
      name: `file${i + 1}.pdf`,
      path_text: `/file${i + 1}.pdf`,
      parent_id: 0,
      modified_at: 1700000000,
      created_at: 1700000000,
      size: 1024 * (i + 1),
      sha1: `hash${i}`,
      in_trash: false,
      is_deleted: false,
      highlighted_name: '',
    })),
    total,
  }
}

describe('SearchPage', () => {
  it('shows initial state on load', () => {
    render(<SearchPage />)
    expect(screen.getByText('等待探索')).toBeInTheDocument()
  })

  it('shows results after search', async () => {
    server.use(
      http.get('/api/v1/app/search', () => {
        return HttpResponse.json(makeSearchResponse(3, 3))
      }),
    )

    render(<SearchPage />)
    const user = userEvent.setup()
    const input = screen.getByRole('searchbox')
    await user.type(input, 'test{Enter}')

    await waitFor(() => {
      expect(screen.getByText('file1.pdf')).toBeInTheDocument()
      expect(screen.getByText('file2.pdf')).toBeInTheDocument()
      expect(screen.getByText('file3.pdf')).toBeInTheDocument()
    })
  })

  it('shows no results state', async () => {
    server.use(
      http.get('/api/v1/app/search', () => {
        return HttpResponse.json({ items: [], total: 0 })
      }),
    )

    render(<SearchPage />)
    const user = userEvent.setup()
    await user.type(screen.getByRole('searchbox'), 'nonexistent{Enter}')

    await waitFor(() => {
      expect(screen.getByText('未找到相关文件')).toBeInTheDocument()
    })
  })

  it('shows error state on API failure', async () => {
    server.use(
      http.get('/api/v1/app/search', () => {
        return HttpResponse.json(
          { code: 'INTERNAL_ERROR', message: 'Server error' },
          { status: 500 },
        )
      }),
    )

    render(<SearchPage />)
    const user = userEvent.setup()
    await user.type(screen.getByRole('searchbox'), 'test{Enter}')

    await waitFor(() => {
      expect(screen.getByText('加载出错了')).toBeInTheDocument()
    })
  })

  it('returns to initial state on clear', async () => {
    server.use(
      http.get('/api/v1/app/search', () => {
        return HttpResponse.json(makeSearchResponse(1, 1))
      }),
    )

    render(<SearchPage />)
    const user = userEvent.setup()
    await user.type(screen.getByRole('searchbox'), 'test{Enter}')

    await waitFor(() => {
      expect(screen.getByText('file1.pdf')).toBeInTheDocument()
    })

    // Click clear button
    await user.click(screen.getByLabelText('清空搜索'))

    await waitFor(() => {
      expect(screen.getByText('等待探索')).toBeInTheDocument()
    })
  })

  it('shows result count', async () => {
    server.use(
      http.get('/api/v1/app/search', () => {
        return HttpResponse.json(makeSearchResponse(3, 50))
      }),
    )

    render(<SearchPage />)
    const user = userEvent.setup()
    await user.type(screen.getByRole('searchbox'), 'test{Enter}')

    await waitFor(() => {
      expect(screen.getByText(/50/)).toBeInTheDocument()
    })
  })
})
