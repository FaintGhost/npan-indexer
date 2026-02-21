import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { SyncProgressDisplay } from './sync-progress-display'
import type { SyncProgress } from '@/lib/sync-schemas'

const baseProgress: SyncProgress = {
  status: 'running',
  startedAt: 1700000000,
  updatedAt: 1700000500,
  roots: [100, 200],
  rootNames: {},
  completedRoots: [100],
  activeRoot: 200,
  mode: 'full',
  aggregateStats: {
    foldersVisited: 50,
    filesIndexed: 300,
    pagesFetched: 60,
    failedRequests: 2,
    startedAt: 1700000000,
    endedAt: 0,
    filesDiscovered: 300,
    skippedFiles: 0,
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
    expect(screen.getByText('已索引文件').closest('div')).toHaveTextContent('300') // filesIndexed
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

  it('shows filesDiscovered stat', () => {
    const progress = {
      ...baseProgress,
      aggregateStats: { ...baseProgress.aggregateStats, filesDiscovered: 50 },
    }
    render(<SyncProgressDisplay progress={progress} />)
    expect(screen.getByText('已发现')).toBeInTheDocument()
  })

  it('shows verification success', () => {
    const progress: SyncProgress = {
      ...baseProgress,
      status: 'done',
      verification: {
        meiliDocCount: 120,
        crawledDocCount: 120,
        discoveredDocCount: 120,
        skippedCount: 0,
        verified: true,
        warnings: [],
      },
    }
    render(<SyncProgressDisplay progress={progress} />)
    expect(screen.getByText('验证通过')).toBeInTheDocument()
  })

  it('shows verification warnings', () => {
    const progress: SyncProgress = {
      ...baseProgress,
      status: 'done',
      verification: {
        meiliDocCount: 110,
        crawledDocCount: 120,
        discoveredDocCount: 120,
        skippedCount: 0,
        verified: false,
        warnings: ['MeiliSearch 文档数(110) < 爬取写入数(120)'],
      },
    }
    render(<SyncProgressDisplay progress={progress} />)
    expect(screen.getByText('MeiliSearch 文档数(110) < 爬取写入数(120)')).toBeInTheDocument()
  })

  it('hides verification when null', () => {
    render(<SyncProgressDisplay progress={baseProgress} />)
    expect(screen.queryByText('验证通过')).not.toBeInTheDocument()
  })

  it('renders incremental mode stats', () => {
    const incrementalProgress: SyncProgress = {
      ...baseProgress,
      mode: 'incremental',
      incrementalStats: {
        changesFetched: 42,
        upserted: 30,
        deleted: 5,
        skippedUpserts: 3,
        skippedDeletes: 1,
        cursorBefore: 100,
        cursorAfter: 200,
      },
    }
    render(<SyncProgressDisplay progress={incrementalProgress} />)
    expect(screen.getByText('增量同步')).toBeInTheDocument()
    expect(screen.getByText('变更').closest('div')).toHaveTextContent('42')
    expect(screen.getByText('写入').closest('div')).toHaveTextContent('30')
    expect(screen.getByText('删除').closest('div')).toHaveTextContent('5')
    expect(screen.getByText('跳过写入').closest('div')).toHaveTextContent('3')
    expect(screen.getByText('跳过删除').closest('div')).toHaveTextContent('1')
  })
})
