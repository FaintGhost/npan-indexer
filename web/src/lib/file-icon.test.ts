import { describe, it, expect } from 'vitest'
import { getFileIcon } from './file-icon'

describe('getFileIcon', () => {
  it('returns archive category for .zip', () => {
    const icon = getFileIcon('archive.zip')
    expect(icon.category).toBe('archive')
    expect(icon.bg).toBe('bg-amber-100')
    expect(icon.text).toBe('text-amber-600')
  })

  it('returns archive category for .rar', () => {
    expect(getFileIcon('file.rar').category).toBe('archive')
  })

  it('returns archive category for .7z', () => {
    expect(getFileIcon('file.7z').category).toBe('archive')
  })

  it('returns archive category for .tar', () => {
    expect(getFileIcon('file.tar').category).toBe('archive')
  })

  it('returns archive category for .gz', () => {
    expect(getFileIcon('file.gz').category).toBe('archive')
  })

  it('returns installer category for .apk', () => {
    const icon = getFileIcon('app.apk')
    expect(icon.category).toBe('installer')
    expect(icon.bg).toBe('bg-emerald-100')
  })

  it('returns installer category for .exe', () => {
    expect(getFileIcon('setup.exe').category).toBe('installer')
  })

  it('returns installer category for .dmg', () => {
    expect(getFileIcon('app.dmg').category).toBe('installer')
  })

  it('returns installer category for name containing 安装包', () => {
    expect(getFileIcon('微信安装包.bin').category).toBe('installer')
  })

  it('returns firmware category for .bin', () => {
    const icon = getFileIcon('firmware.bin')
    expect(icon.category).toBe('firmware')
    expect(icon.bg).toBe('bg-purple-100')
  })

  it('returns firmware category for .iso', () => {
    expect(getFileIcon('ubuntu.iso').category).toBe('firmware')
  })

  it('returns firmware category for .img', () => {
    expect(getFileIcon('disk.img').category).toBe('firmware')
  })

  it('returns firmware category for name containing 固件', () => {
    expect(getFileIcon('路由器固件v2.dat').category).toBe('firmware')
  })

  it('returns document category for .pdf', () => {
    const icon = getFileIcon('report.pdf')
    expect(icon.category).toBe('document')
    expect(icon.bg).toBe('bg-rose-100')
  })

  it('returns document category for .doc', () => {
    expect(getFileIcon('file.doc').category).toBe('document')
  })

  it('returns document category for .docx', () => {
    expect(getFileIcon('file.docx').category).toBe('document')
  })

  it('returns document category for .txt', () => {
    expect(getFileIcon('readme.txt').category).toBe('document')
  })

  it('returns document category for .md', () => {
    expect(getFileIcon('README.md').category).toBe('document')
  })

  it('returns default category for unknown extension', () => {
    const icon = getFileIcon('data.xyz')
    expect(icon.category).toBe('default')
    expect(icon.bg).toBe('bg-blue-50')
  })

  it('returns default category for no extension', () => {
    expect(getFileIcon('Makefile').category).toBe('default')
  })

  it('installer takes priority over firmware for 安装包.bin', () => {
    // "安装包" in name should match installer before firmware checks .bin
    expect(getFileIcon('微信安装包.bin').category).toBe('installer')
  })
})
