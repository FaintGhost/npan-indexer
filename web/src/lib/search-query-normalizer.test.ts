import { describe, expect, it } from 'vitest'
import { existsSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath, pathToFileURL } from 'node:url'

const currentDir = dirname(fileURLToPath(import.meta.url))
const normalizerPath = resolve(currentDir, 'search-query-normalizer.ts')
const hasNormalizerModule = existsSync(normalizerPath)

async function loadNormalizer() {
  const moduleUrl = pathToFileURL(normalizerPath).href
  return import(/* @vite-ignore */ moduleUrl)
}

describe('normalizeSearchQuery', () => {
  it('requires a dedicated query normalizer module for legacy preprocess alignment', () => {
    expect(
      hasNormalizerModule,
      'expected search-query-normalizer module to exist so public search can align outbound query with legacy preprocessQuery semantics for extension, version, and multi-word queries',
    ).toBe(true)
  })

  if (!hasNormalizerModule) {
    return
  }

  it.each([
    ['规格书.pdf', 'pdf 规格书'],
    ['firmware v3.2.1', 'firmware 3.2.1'],
    ['mx40 spec pdf', 'pdf mx40 spec'],
    ['mx6000 V1.5.0 pdf', 'pdf mx6000 1.5.0'],
  ])('aligns legacy preprocess semantics for %s', async (input, expected) => {
    const { normalizeSearchQuery } = await loadNormalizer()
    expect(normalizeSearchQuery(input)).toBe(expected)
  })

  it.each([
    ['', ''],
    ['王晨 报告', '王晨 报告'],
    ['VIVO手机', 'VIVO手机'],
    ['4.9.4.0', '4.9.4.0'],
  ])('keeps unsupported query %s unchanged', async (input, expected) => {
    const { normalizeSearchQuery } = await loadNormalizer()
    expect(normalizeSearchQuery(input)).toBe(expected)
  })
})
