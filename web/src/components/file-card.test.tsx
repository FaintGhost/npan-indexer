import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { FileCard } from './file-card'
import type { IndexDocument } from '@/lib/schemas'

const baseDoc: IndexDocument = {
  doc_id: 'file_123',
  source_id: 456,
  type: 'file',
  name: 'report.pdf',
  path_text: '/docs/report.pdf',
  parent_id: 10,
  modified_at: 1700000000,
  created_at: 1700000000,
  size: 1048576,
  sha1: 'abc123',
  in_trash: false,
  is_deleted: false,
  highlighted_name: '',
}

describe('FileCard', () => {
  it('renders file name', () => {
    render(<FileCard doc={baseDoc} downloadStatus="idle" onDownload={() => {}} />)
    expect(screen.getByText('report.pdf')).toBeInTheDocument()
  })

  it('renders highlighted name with HTML', () => {
    const doc = { ...baseDoc, highlighted_name: '<mark>report</mark>.pdf' }
    render(<FileCard doc={doc} downloadStatus="idle" onDownload={() => {}} />)
    const mark = document.querySelector('mark')
    expect(mark).toBeInTheDocument()
    expect(mark?.textContent).toBe('report')
  })

  it('renders formatted size', () => {
    render(<FileCard doc={baseDoc} downloadStatus="idle" onDownload={() => {}} />)
    expect(screen.getByText(/1 MB/)).toBeInTheDocument()
  })

  it('renders formatted date', () => {
    render(<FileCard doc={baseDoc} downloadStatus="idle" onDownload={() => {}} />)
    // Check for date pattern YYYY-MM-DD HH:mm
    const dateEl = screen.getByText(/\d{4}-\d{2}-\d{2} \d{2}:\d{2}/)
    expect(dateEl).toBeInTheDocument()
  })

  it('renders source_id', () => {
    render(<FileCard doc={baseDoc} downloadStatus="idle" onDownload={() => {}} />)
    expect(screen.getByText(/456/)).toBeInTheDocument()
  })

  it('renders document icon for pdf', () => {
    const { container } = render(<FileCard doc={baseDoc} downloadStatus="idle" onDownload={() => {}} />)
    // pdf -> document category -> bg-rose-100
    const iconEl = container.querySelector('.bg-rose-100')
    expect(iconEl).toBeInTheDocument()
  })

  it('renders firmware icon for .bin file', () => {
    const doc = { ...baseDoc, name: 'firmware.bin' }
    const { container } = render(<FileCard doc={doc} downloadStatus="idle" onDownload={() => {}} />)
    const iconEl = container.querySelector('.bg-purple-100')
    expect(iconEl).toBeInTheDocument()
  })

  it('renders archive icon for .zip file', () => {
    const doc = { ...baseDoc, name: 'archive.zip' }
    const { container } = render(<FileCard doc={doc} downloadStatus="idle" onDownload={() => {}} />)
    const iconEl = container.querySelector('.bg-amber-100')
    expect(iconEl).toBeInTheDocument()
  })

  it('renders download button', () => {
    render(<FileCard doc={baseDoc} downloadStatus="idle" onDownload={() => {}} />)
    expect(screen.getByRole('button', { name: /下载/ })).toBeInTheDocument()
  })

  it('calls onDownload when download button clicked', async () => {
    const onDownload = vi.fn()
    render(<FileCard doc={baseDoc} downloadStatus="idle" onDownload={onDownload} />)
    const user = userEvent.setup()
    await user.click(screen.getByRole('button', { name: /下载/ }))
    expect(onDownload).toHaveBeenCalled()
  })
})
