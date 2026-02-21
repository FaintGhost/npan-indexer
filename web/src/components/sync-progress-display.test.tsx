import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { SyncProgressDisplay } from './sync-progress-display'
import type { SyncProgress } from '@/lib/sync-schemas'

const baseProgress: SyncProgress = {
  status: 'running',
  startedAt: 1700000000,
  updatedAt: 1700000500,
  roots: [100, 200],
  completedRoots: [100],
  activeRoot: 200,
  aggregateStats: {
    foldersVisited: 50,
    filesIndexed: 300,
    pagesFetched: 60,
    failedRequests: 2,
    startedAt: 1700000000,
    endedAt: 0,
  },
  rootProgress: {},
  lastError: '',
}

describe('SyncProgressDisplay', () => {
  it('shows running status', () => {
    render(<SyncProgressDisplay progress={baseProgress} />)
    expect(screen.getByText('运行中')).toBeInTheDocument()
  })

  it('shows roots progress', () => {
    render(<SyncProgressDisplay progress={baseProgress} />)
    expect(screen.getByText(/1.*\/.*2/)).toBeInTheDocument()
  })

  it('shows aggregate stats', () => {
    render(<SyncProgressDisplay progress={baseProgress} />)
    expect(screen.getByText(/300/)).toBeInTheDocument() // filesIndexed
    expect(screen.getByText(/60/)).toBeInTheDocument() // pagesFetched
  })

  it('shows done status', () => {
    render(<SyncProgressDisplay progress={{ ...baseProgress, status: 'done' }} />)
    expect(screen.getByText('已完成')).toBeInTheDocument()
  })

  it('shows error status with lastError', () => {
    render(
      <SyncProgressDisplay
        progress={{ ...baseProgress, status: 'error', lastError: '网络超时' }}
      />,
    )
    expect(screen.getByText('出错')).toBeInTheDocument()
    expect(screen.getByText('网络超时')).toBeInTheDocument()
  })

  it('shows cancelled status', () => {
    render(<SyncProgressDisplay progress={{ ...baseProgress, status: 'cancelled' }} />)
    expect(screen.getByText('已取消')).toBeInTheDocument()
  })

  it('shows failed requests count when > 0', () => {
    render(<SyncProgressDisplay progress={baseProgress} />)
    expect(screen.getByText('失败请求')).toBeInTheDocument()
    const failedCard = screen.getByText('失败请求').closest('div')!
    expect(failedCard).toHaveTextContent('2')
  })
})
