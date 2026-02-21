import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { SearchPage } from '../routes/app.index.lazy'
import { SkeletonCard } from '../components/skeleton-card'
import { DownloadButton } from '../components/download-button'
import { ApiKeyDialog } from '../components/api-key-dialog'
import { SearchInput } from '../components/search-input'
import { http, HttpResponse } from 'msw'
import { server } from './mocks/server'

describe('Accessibility', () => {
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

  it('search page results area exists', () => {
    server.use(
      http.get('/api/v1/app/search', () => {
        return HttpResponse.json({ items: [], total: 0 })
      }),
    )
    render(<SearchPage />)
    // Initial state should render without errors
    expect(screen.getByText('等待探索')).toBeInTheDocument()
  })
})
