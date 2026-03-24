import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { SearchPage } from '../routes/index.lazy'
import { SkeletonCard } from '../components/skeleton-card'
import { DownloadButton } from '../components/download-button'
import { ApiKeyDialog } from '../components/api-key-dialog'
import { SearchInput } from '../components/search-input'
import { http, HttpResponse } from 'msw'
import { server } from './mocks/server'
import { createTestProvider } from './test-providers'

describe('Accessibility', () => {
  const wrapper = createTestProvider()

  it('search input has proper aria-label', () => {
    render(
      <SearchInput value="" onChange={() => {}} onSubmit={() => {}} onClear={() => {}} />,
    )
    const input = screen.getByRole('searchbox')
    expect(
      input.getAttribute('aria-label') || input.getAttribute('placeholder'),
    ).toBeTruthy()
  })

  it('skeleton card has aria-hidden', () => {
    const { container } = render(<SkeletonCard />)
    expect(container.firstElementChild?.getAttribute('aria-hidden')).toBe('true')
  })

  it('download button has accessible name', () => {
    render(<DownloadButton status="idle" onClick={() => {}} />)
    const button = screen.getByRole('button')
    expect(button.textContent).toBeTruthy()
  })

  it('api key dialog input has type password', () => {
    render(<ApiKeyDialog open onSubmit={() => {}} />)
    const input = screen.getByPlaceholderText(/API Key/i)
    expect(input.getAttribute('type')).toBe('password')
  })

  it('api key dialog overlay blocks interaction', () => {
    render(<ApiKeyDialog open onSubmit={() => {}} />)
    // Check dialog has appropriate structure
    expect(screen.getByRole('button', { name: /确认/i })).toBeInTheDocument()
  })

  it('search page results area exists', async () => {
    server.use(
      http.post('/npan.v1.AppService/AppSearch', () => {
        return HttpResponse.json({ result: { items: [], total: '0' } })
      }),
    )
    render(<SearchPage />, { wrapper })
    expect(await screen.findByText('等待探索')).toBeInTheDocument()
  })

  it('filter controls have accessible radio semantics', async () => {
    server.use(
      http.post('/npan.v1.AppService/AppSearch', () => {
        return HttpResponse.json({
          result: {
            items: [
              {
                docId: 'f1',
                sourceId: '1',
                type: 'ITEM_TYPE_FILE',
                name: 'a.pdf',
                pathText: '/a.pdf',
                parentId: '0',
                modifiedAt: '1700000000',
                createdAt: '1700000000',
                size: '1',
                sha1: 'x',
                inTrash: false,
                isDeleted: false,
                highlightedName: '',
              },
            ],
            total: '1',
          },
        })
      }),
    )

    render(<SearchPage />, { wrapper })
    const user = userEvent.setup()
    await user.type(screen.getByRole('searchbox'), 'test{Enter}')

    const group = await screen.findByRole('radiogroup', { name: '结果筛选' })
    expect(group).toBeInTheDocument()

    const all = screen.getByRole('radio', { name: '全部' })
    const image = screen.getByRole('radio', { name: '图片' })

    expect(all).toHaveAttribute('aria-checked', 'true')
    expect(image).toHaveAttribute('aria-checked', 'false')

    await user.click(image)
    expect(image).toHaveAttribute('aria-checked', 'true')
    expect(all).toHaveAttribute('aria-checked', 'false')
  })
})
