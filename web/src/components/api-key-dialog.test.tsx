import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { ApiKeyDialog } from './api-key-dialog'

describe('ApiKeyDialog', () => {
  it('renders password input and confirm button', () => {
    render(<ApiKeyDialog open onSubmit={() => {}} />)
    expect(screen.getByPlaceholderText(/API Key/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /确认/i })).toBeInTheDocument()
  })

  it('shows validation error on empty submit', async () => {
    const onSubmit = vi.fn()
    render(<ApiKeyDialog open onSubmit={onSubmit} />)
    const user = userEvent.setup()
    await user.click(screen.getByRole('button', { name: /确认/i }))
    expect(screen.getByText('请输入 API Key')).toBeInTheDocument()
    expect(onSubmit).not.toHaveBeenCalled()
  })

  it('calls onSubmit with key value', async () => {
    const onSubmit = vi.fn()
    render(<ApiKeyDialog open onSubmit={onSubmit} />)
    const user = userEvent.setup()
    await user.type(screen.getByPlaceholderText(/API Key/i), 'my-secret-key')
    await user.click(screen.getByRole('button', { name: /确认/i }))
    expect(onSubmit).toHaveBeenCalledWith('my-secret-key')
  })

  it('displays error message', () => {
    render(<ApiKeyDialog open onSubmit={() => {}} error="API Key 无效" />)
    expect(screen.getByText('API Key 无效')).toBeInTheDocument()
  })

  it('disables button when loading', () => {
    render(<ApiKeyDialog open onSubmit={() => {}} loading />)
    expect(screen.getByRole('button', { name: /确认|验证中/i })).toBeDisabled()
  })

  it('does not render when not open', () => {
    const { container } = render(<ApiKeyDialog open={false} onSubmit={() => {}} />)
    expect(container.innerHTML).toBe('')
  })
})
