import { describe, it, expect } from 'vitest'
import { formatBytes, formatTime } from './format'

describe('formatBytes', () => {
  it('returns "-" for 0', () => {
    expect(formatBytes(0)).toBe('-')
  })

  it('returns "-" for undefined', () => {
    expect(formatBytes(undefined as unknown as number)).toBe('-')
  })

  it('formats bytes', () => {
    expect(formatBytes(500)).toBe('500 B')
  })

  it('formats kilobytes', () => {
    expect(formatBytes(1024)).toBe('1 KB')
  })

  it('formats kilobytes with decimal', () => {
    expect(formatBytes(1536)).toBe('1.5 KB')
  })

  it('formats large KB without decimal', () => {
    expect(formatBytes(15360)).toBe('15 KB')
  })

  it('formats megabytes', () => {
    expect(formatBytes(1048576)).toBe('1 MB')
  })

  it('formats gigabytes', () => {
    expect(formatBytes(1073741824)).toBe('1 GB')
  })
})

describe('formatTime', () => {
  it('returns "-" for 0', () => {
    expect(formatTime(0)).toBe('-')
  })

  it('returns "-" for undefined', () => {
    expect(formatTime(undefined as unknown as number)).toBe('-')
  })

  it('formats unix seconds timestamp', () => {
    // 2023-11-14 in some timezone - just check format pattern
    const result = formatTime(1700000000)
    expect(result).toMatch(/^\d{4}-\d{2}-\d{2} \d{2}:\d{2}$/)
  })

  it('formats unix milliseconds timestamp', () => {
    const result = formatTime(1700000000000)
    expect(result).toMatch(/^\d{4}-\d{2}-\d{2} \d{2}:\d{2}$/)
  })

  it('produces same result for seconds and milliseconds of same instant', () => {
    expect(formatTime(1700000000)).toBe(formatTime(1700000000000))
  })
})
