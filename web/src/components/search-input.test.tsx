import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { SearchInput } from './search-input'

describe('SearchInput', () => {
  it('shows keyboard shortcut badge when empty', () => {
    render(
      <SearchInput value="" onChange={() => {}} onSubmit={() => {}} onClear={() => {}} />,
    )
    // Should show ⌘K or Ctrl K
    const badge = screen.getByText(/⌘K|Ctrl/i)
    expect(badge).toBeInTheDocument()
  })

  it('shows clear button when has value', () => {
    render(
      <SearchInput value="test" onChange={() => {}} onSubmit={() => {}} onClear={() => {}} />,
    )
    expect(screen.getByLabelText('清空搜索')).toBeInTheDocument()
  })

  it('hides keyboard shortcut badge when has value', () => {
    render(
      <SearchInput value="test" onChange={() => {}} onSubmit={() => {}} onClear={() => {}} />,
    )
    expect(screen.queryByText(/⌘K|Ctrl/i)).not.toBeInTheDocument()
  })

  it('calls onChange on input', async () => {
    const onChange = vi.fn()
    render(
      <SearchInput value="" onChange={onChange} onSubmit={() => {}} onClear={() => {}} />,
    )
    const user = userEvent.setup()
    await user.type(screen.getByRole('searchbox'), 'hello')
    expect(onChange).toHaveBeenCalled()
  })

  it('calls onSubmit on Enter key', async () => {
    const onSubmit = vi.fn()
    render(
      <SearchInput value="test" onChange={() => {}} onSubmit={onSubmit} onClear={() => {}} />,
    )
    const user = userEvent.setup()
    await user.type(screen.getByRole('searchbox'), '{Enter}')
    expect(onSubmit).toHaveBeenCalledOnce()
  })

  it('calls onClear when clear button clicked', async () => {
    const onClear = vi.fn()
    render(
      <SearchInput value="test" onChange={() => {}} onSubmit={() => {}} onClear={onClear} />,
    )
    const user = userEvent.setup()
    await user.click(screen.getByLabelText('清空搜索'))
    expect(onClear).toHaveBeenCalledOnce()
  })

  it('accepts ref for focus management', () => {
    const ref = { current: null as HTMLInputElement | null }
    render(
      <SearchInput
        ref={ref}
        value=""
        onChange={() => {}}
        onSubmit={() => {}}
        onClear={() => {}}
      />,
    )
    expect(ref.current).toBeInstanceOf(HTMLInputElement)
  })
})
