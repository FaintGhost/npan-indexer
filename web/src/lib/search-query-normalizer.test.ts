import { describe, expect, it } from 'vitest'
import { normalizeSearchQuery } from './search-query-normalizer'

describe('normalizeSearchQuery', () => {
  it('requires a dedicated query normalizer module for legacy preprocess alignment', () => {
    expect(typeof normalizeSearchQuery).toBe('function')
  })

  it.each([
    ['规格书.pdf', 'pdf 规格书'],
    ['firmware v3.2.1', 'firmware 3.2.1'],
    ['mx40 spec pdf', 'pdf mx40 spec'],
    ['mx6000 V1.5.0 pdf', 'pdf mx6000 1.5.0'],
  ])('aligns legacy preprocess semantics for %s', async (input, expected) => {
    expect(normalizeSearchQuery(input)).toBe(expected)
  })

  it.each([
    ['', ''],
    ['王晨 报告', '王晨 报告'],
    ['VIVO手机', 'VIVO手机'],
    ['4.9.4.0', '4.9.4.0'],
  ])('keeps unsupported query %s unchanged', async (input, expected) => {
    expect(normalizeSearchQuery(input)).toBe(expected)
  })
})
