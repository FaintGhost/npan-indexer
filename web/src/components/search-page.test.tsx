import { describe, it, expect } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'
import { server } from '../tests/mocks/server'
import { createTestProvider } from '../tests/test-providers'
import { SearchPage } from '../routes/index.lazy'

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

function toConnectSearchResponse(items: ReturnType<typeof makeSearchResponse>['items'], total: number) {
  return {
    result: {
      items: items.map((item) => ({
        docId: item.doc_id,
        sourceId: String(item.source_id),
        type: item.type === 'folder' ? 'ITEM_TYPE_FOLDER' : 'ITEM_TYPE_FILE',
        name: item.name,
        pathText: item.path_text,
        parentId: String(item.parent_id),
        modifiedAt: String(item.modified_at),
        createdAt: String(item.created_at),
        size: String(item.size),
        sha1: item.sha1,
        inTrash: item.in_trash,
        isDeleted: item.is_deleted,
        highlightedName: item.highlighted_name,
      })),
      total: String(total),
    },
  }
}

describe('SearchPage', () => {
  const wrapper = createTestProvider()

  it('shows initial state on load', () => {
    render(<SearchPage />, { wrapper })
    expect(screen.getByText('等待探索')).toBeInTheDocument()
  })

  it('shows results after search', async () => {
    server.use(
      http.post('/npan.v1.AppService/AppSearch', () => {
        const response = makeSearchResponse(3, 3)
        return HttpResponse.json(toConnectSearchResponse(response.items, response.total))
      }),
    )

    render(<SearchPage />, { wrapper })
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
      http.post('/npan.v1.AppService/AppSearch', () => {
        return HttpResponse.json(toConnectSearchResponse([], 0))
      }),
    )

    render(<SearchPage />, { wrapper })
    const user = userEvent.setup()
    await user.type(screen.getByRole('searchbox'), 'nonexistent{Enter}')

    await waitFor(() => {
      // Empty state card has the description text
      expect(screen.getByText(/没有找到匹配的内容/)).toBeInTheDocument()
    })
  })

  it('shows error state on API failure', async () => {
    server.use(
      http.post('/npan.v1.AppService/AppSearch', () => {
        return HttpResponse.json(
          { code: 'internal', message: 'Server error' },
          { status: 500 },
        )
      }),
    )

    render(<SearchPage />, { wrapper })
    const user = userEvent.setup()
    await user.type(screen.getByRole('searchbox'), 'test{Enter}')

    await waitFor(() => {
      expect(screen.getByText('加载出错了')).toBeInTheDocument()
    })
  })

  it('returns to initial state on clear', async () => {
    server.use(
      http.post('/npan.v1.AppService/AppSearch', () => {
        const response = makeSearchResponse(1, 1)
        return HttpResponse.json(toConnectSearchResponse(response.items, response.total))
      }),
    )

    render(<SearchPage />, { wrapper })
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
      http.post('/npan.v1.AppService/AppSearch', () => {
        const response = makeSearchResponse(3, 50)
        return HttpResponse.json(toConnectSearchResponse(response.items, response.total))
      }),
    )

    render(<SearchPage />, { wrapper })
    const user = userEvent.setup()
    await user.type(screen.getByRole('searchbox'), 'test{Enter}')

    await waitFor(() => {
      // Counter shows "3 / 50"
      expect(screen.getByText('3 / 50')).toBeInTheDocument()
    })
  })
})
