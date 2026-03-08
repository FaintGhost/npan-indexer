import { describe, expect, it } from 'vitest'
import { fromMeiliHit } from './meili-hit-adapter'

describe('fromMeiliHit', () => {
  it('maps Meilisearch hit fields into current IndexDocument shape', () => {
    const doc = fromMeiliHit({
      doc_id: 'file_9001',
      source_id: 9001,
      type: 'file',
      name: 'quarterly-report-2024.pdf',
      path_text: '/reports/quarterly-report-2024.pdf',
      parent_id: 7,
      modified_at: 1718444400,
      created_at: 1718440800,
      size: 1048576,
      sha1: 'abc123',
      in_trash: false,
      is_deleted: false,
      file_category: 'doc',
    })

    expect(doc).toMatchObject({
      doc_id: 'file_9001',
      source_id: 9001,
      type: 'file',
      name: 'quarterly-report-2024.pdf',
      path_text: '/reports/quarterly-report-2024.pdf',
      parent_id: 7,
      modified_at: 1718444400,
      created_at: 1718440800,
      size: 1048576,
      sha1: 'abc123',
      in_trash: false,
      is_deleted: false,
      file_category: 'doc',
    })
  })

  it('uses InstantSearch highlight payload for highlighted_name', () => {
    const doc = fromMeiliHit({
      doc_id: 'file_9002',
      source_id: 9002,
      type: 'file',
      name: 'project-design-spec.docx',
      path_text: '/docs/project-design-spec.docx',
      parent_id: 0,
      modified_at: 1721484000,
      created_at: 1721480400,
      size: 524288,
      sha1: 'def456',
      in_trash: false,
      is_deleted: false,
      _highlightResult: {
        name: {
          value: 'project-<mark>design</mark>-spec.docx',
        },
      },
    })

    expect(doc.highlighted_name).toBe('project-<mark>design</mark>-spec.docx')
  })

  it('sanitizes highlight payload and keeps only mark tags', () => {
    const doc = fromMeiliHit({
      doc_id: 'file_9003',
      source_id: 9003,
      type: 'file',
      name: 'architecture-diagram.png',
      path_text: '/design/architecture-diagram.png',
      parent_id: 0,
      modified_at: 1722502800,
      created_at: 1722499200,
      size: 2097152,
      sha1: 'ghi789',
      in_trash: false,
      is_deleted: false,
      _highlightResult: {
        name: {
          value: '<img src=x onerror=alert(1)><mark class="x">diagram</mark><script>alert(1)</script>.png',
        },
      },
    })

    expect(doc.highlighted_name).toBe('<mark>diagram</mark>.png')
  })

  it('falls back to the raw name when highlight payload is absent', () => {
    const doc = fromMeiliHit({
      doc_id: 'file_9004',
      source_id: 9004,
      type: 'file',
      name: 'architecture-diagram.png',
      path_text: '/design/architecture-diagram.png',
      parent_id: 0,
      modified_at: 1722502800,
      created_at: 1722499200,
      size: 2097152,
      sha1: 'ghi789',
      in_trash: false,
      is_deleted: false,
    })

    expect(doc.highlighted_name).toBeUndefined()
    expect(doc.name).toBe('architecture-diagram.png')
  })
})
