import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'

function Hello() {
  return <div>Hello, World!</div>
}

describe('testing infrastructure', () => {
  it('renders a React component', () => {
    render(<Hello />)
    expect(screen.getByText('Hello, World!')).toBeInTheDocument()
  })
})
