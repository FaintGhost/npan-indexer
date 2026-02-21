import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { InitialState, NoResultsState, ErrorState } from './empty-state'

describe('InitialState', () => {
  it('renders title', () => {
    render(<InitialState />)
    expect(screen.getByText('等待探索')).toBeInTheDocument()
  })

  it('renders description', () => {
    render(<InitialState />)
    expect(screen.getByText(/输入关键词/)).toBeInTheDocument()
  })
})

describe('NoResultsState', () => {
  it('renders title', () => {
    render(<NoResultsState />)
    expect(screen.getByText('未找到相关文件')).toBeInTheDocument()
  })
})

describe('ErrorState', () => {
  it('renders title', () => {
    render(<ErrorState />)
    expect(screen.getByText('加载出错了')).toBeInTheDocument()
  })
})
