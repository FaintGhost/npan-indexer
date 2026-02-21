import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { DownloadButton } from './download-button'

describe('DownloadButton', () => {
  it('renders idle state with download text', () => {
    render(<DownloadButton status="idle" onClick={() => {}} />)
    expect(screen.getByText('下载')).toBeInTheDocument()
  })

  it('renders loading state with spinner text', () => {
    render(<DownloadButton status="loading" onClick={() => {}} />)
    expect(screen.getByText('获取中')).toBeInTheDocument()
  })

  it('is disabled when loading', () => {
    render(<DownloadButton status="loading" onClick={() => {}} />)
    expect(screen.getByRole('button')).toBeDisabled()
  })

  it('renders success state', () => {
    render(<DownloadButton status="success" onClick={() => {}} />)
    expect(screen.getByText('成功')).toBeInTheDocument()
  })

  it('is disabled when success', () => {
    render(<DownloadButton status="success" onClick={() => {}} />)
    expect(screen.getByRole('button')).toBeDisabled()
  })

  it('renders error state with retry text', () => {
    render(<DownloadButton status="error" onClick={() => {}} />)
    expect(screen.getByText('重试')).toBeInTheDocument()
  })

  it('is not disabled in error state', () => {
    render(<DownloadButton status="error" onClick={() => {}} />)
    expect(screen.getByRole('button')).not.toBeDisabled()
  })

  it('calls onClick when clicked', async () => {
    const onClick = vi.fn()
    render(<DownloadButton status="idle" onClick={onClick} />)
    const user = userEvent.setup()
    await user.click(screen.getByRole('button'))
    expect(onClick).toHaveBeenCalledOnce()
  })

  it('does not call onClick when disabled', async () => {
    const onClick = vi.fn()
    render(<DownloadButton status="loading" onClick={onClick} />)
    const user = userEvent.setup()
    await user.click(screen.getByRole('button'))
    expect(onClick).not.toHaveBeenCalled()
  })
})
