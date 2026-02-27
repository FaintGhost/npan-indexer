import { describe, expect, it } from 'vitest'
import { categorizeFileName, matchesSearchFilter, normalizeSearchFilter } from './file-category'

describe('file-category', () => {
  it('normalizes unknown filter to all', () => {
    expect(normalizeSearchFilter(undefined)).toBe('all')
    expect(normalizeSearchFilter('unknown')).toBe('all')
    expect(normalizeSearchFilter('doc')).toBe('doc')
  })

  it('categorizes common extensions', () => {
    expect(categorizeFileName('report.pdf')).toBe('doc')
    expect(categorizeFileName('photo.jpg')).toBe('image')
    expect(categorizeFileName('demo.mp4')).toBe('video')
    expect(categorizeFileName('backup.tar.gz')).toBe('archive')
    expect(categorizeFileName('README')).toBe('other')
  })

  it('matches all filter for any file name', () => {
    expect(matchesSearchFilter('foo.bin', 'all')).toBe(true)
    expect(matchesSearchFilter('foo.pdf', 'all')).toBe(true)
  })
})
