const SEARCH_BACKEND = (process.env.NPA_SEARCH_BACKEND ?? 'meilisearch').trim().toLowerCase()

const MEILI_HOST = process.env.MEILI_HOST ?? 'http://localhost:7700'
const MEILI_API_KEY = process.env.MEILI_API_KEY ?? 'ci-test-meili-key-5678'
const MEILI_INDEX = process.env.MEILI_INDEX ?? 'npan_items'

const TYPESENSE_HOST = process.env.TYPESENSE_HOST ?? 'http://localhost:8108'
const TYPESENSE_API_KEY = process.env.TYPESENSE_API_KEY ?? 'ci-test-typesense-key-5678'
const TYPESENSE_COLLECTION = process.env.TYPESENSE_COLLECTION ?? 'npan_items'

interface TestDocument {
  doc_id: string
  source_id: number
  name: string
  name_base: string
  name_ext: string
  path_text: string
  parent_id: number
  size: number
  modified_at: number
  created_at: number
  type: string
  file_category: 'doc' | 'image' | 'video' | 'archive' | 'other'
  is_deleted: boolean
  in_trash: boolean
  sha1: string
}

function makeDoc(
  sourceId: number,
  name: string,
  fileCategory: TestDocument['file_category'],
  modifiedAt: number,
  size: number,
): TestDocument {
  const dot = name.lastIndexOf('.')
  const nameExt = dot > 0 ? name.slice(dot + 1).toLowerCase() : ''
  const nameBase = dot > 0 ? name.slice(0, dot) : name
  return {
    doc_id: `file_${sourceId}`,
    source_id: sourceId,
    name,
    name_base: nameBase,
    name_ext: nameExt,
    path_text: `/${name}`,
    parent_id: 0,
    size,
    modified_at: modifiedAt,
    created_at: modifiedAt,
    type: 'file',
    file_category: fileCategory,
    is_deleted: false,
    in_trash: false,
    sha1: '',
  }
}

const namedDocs: TestDocument[] = [
  makeDoc(9001, 'quarterly-report-2024.pdf', 'doc', 1718444400, 1048576),
  makeDoc(9002, 'project-design-spec.docx', 'doc', 1721484000, 524288),
  makeDoc(9003, 'architecture-diagram.png', 'image', 1722502800, 2097152),
  makeDoc(9004, 'launch-demo.mp4', 'video', 1725170400, 3145728),
  makeDoc(9005, 'release-bundle.tar.gz', 'archive', 1727852400, 4194304),
]

const bulkDocs: TestDocument[] = Array.from({ length: 35 }, (_, i) =>
  makeDoc(1000 + i, `test-file-${String(i).padStart(3, '0')}.txt`, 'doc', 1704067200, 1024 * (i + 1)),
)

export const TEST_DOCUMENTS: TestDocument[] = [...namedDocs, ...bulkDocs]

const FILTERABLE_ATTRS = ['source_id', 'type', 'file_category', 'is_deleted', 'in_trash']

function meiliHeaders(): Record<string, string> {
  return {
    'Content-Type': 'application/json',
    Authorization: `Bearer ${MEILI_API_KEY}`,
  }
}

function typesenseHeaders(contentType = 'application/json'): Record<string, string> {
  return {
    'Content-Type': contentType,
    'X-TYPESENSE-API-KEY': TYPESENSE_API_KEY,
  }
}

export async function waitForMeiliTask(taskUid: number): Promise<void> {
  const url = `${MEILI_HOST}/tasks/${taskUid}`
  while (true) {
    const res = await fetch(url, { headers: meiliHeaders() })
    if (!res.ok) {
      throw new Error(`Failed to fetch task ${taskUid}: ${res.status}`)
    }
    const task = (await res.json()) as { status: string; error?: unknown }
    if (task.status === 'succeeded') return
    if (task.status === 'failed') {
      throw new Error(`Meilisearch task ${taskUid} failed: ${JSON.stringify(task.error)}`)
    }
    await new Promise((r) => setTimeout(r, 200))
  }
}

async function seedMeilisearch(): Promise<void> {
  const settingsRes = await fetch(
    `${MEILI_HOST}/indexes/${MEILI_INDEX}/settings`,
    {
      method: 'PATCH',
      headers: meiliHeaders(),
      body: JSON.stringify({ filterableAttributes: FILTERABLE_ATTRS }),
    },
  )
  if (!settingsRes.ok) {
    const createRes = await fetch(`${MEILI_HOST}/indexes`, {
      method: 'POST',
      headers: meiliHeaders(),
      body: JSON.stringify({
        uid: MEILI_INDEX,
        primaryKey: 'doc_id',
      }),
    })
    if (createRes.ok) {
      const createTask = (await createRes.json()) as { taskUid: number }
      await waitForMeiliTask(createTask.taskUid)
    }

    const retryRes = await fetch(
      `${MEILI_HOST}/indexes/${MEILI_INDEX}/settings`,
      {
        method: 'PATCH',
        headers: meiliHeaders(),
        body: JSON.stringify({ filterableAttributes: FILTERABLE_ATTRS }),
      },
    )
    if (retryRes.ok) {
      const retryTask = (await retryRes.json()) as { taskUid: number }
      await waitForMeiliTask(retryTask.taskUid)
    }
  } else {
    const settingsTask = (await settingsRes.json()) as { taskUid: number }
    await waitForMeiliTask(settingsTask.taskUid)
  }

  const addRes = await fetch(
    `${MEILI_HOST}/indexes/${MEILI_INDEX}/documents?primaryKey=doc_id`,
    {
      method: 'POST',
      headers: meiliHeaders(),
      body: JSON.stringify(TEST_DOCUMENTS),
    },
  )
  if (!addRes.ok) {
    throw new Error(`Failed to add documents: ${addRes.status}`)
  }
  const addTask = (await addRes.json()) as { taskUid: number }
  await waitForMeiliTask(addTask.taskUid)
}

async function clearMeilisearch(): Promise<void> {
  const res = await fetch(
    `${MEILI_HOST}/indexes/${MEILI_INDEX}/documents`,
    {
      method: 'DELETE',
      headers: meiliHeaders(),
    },
  )
  if (!res.ok) {
    throw new Error(`Failed to delete documents: ${res.status}`)
  }
  const task = (await res.json()) as { taskUid: number }
  await waitForMeiliTask(task.taskUid)
}

async function createTypesenseCollection(): Promise<void> {
  const createRes = await fetch(`${TYPESENSE_HOST}/collections`, {
    method: 'POST',
    headers: typesenseHeaders(),
    body: JSON.stringify({
      name: TYPESENSE_COLLECTION,
      default_sorting_field: 'modified_at',
      token_separators: ['-', '_'],
      fields: [
        { name: 'doc_id', type: 'string' },
        { name: 'source_id', type: 'int64', sort: true },
        { name: 'type', type: 'string', facet: true },
        { name: 'name', type: 'string' },
        { name: 'name_base', type: 'string' },
        { name: 'name_ext', type: 'string', optional: true },
        { name: 'file_category', type: 'string', facet: true, optional: true },
        { name: 'path_text', type: 'string' },
        { name: 'parent_id', type: 'int64', facet: true, sort: true },
        { name: 'modified_at', type: 'int64', facet: true, sort: true },
        { name: 'created_at', type: 'int64', sort: true },
        { name: 'size', type: 'int64', sort: true },
        { name: 'sha1', type: 'string', optional: true },
        { name: 'in_trash', type: 'bool', facet: true },
        { name: 'is_deleted', type: 'bool', facet: true },
      ],
    }),
  })
  if (createRes.ok || createRes.status === 409) {
    return
  }
  throw new Error(`Failed to create Typesense collection: ${createRes.status} ${await createRes.text()}`)
}

async function seedTypesense(): Promise<void> {
  await clearTypesense()
  await createTypesenseCollection()

  const body = TEST_DOCUMENTS.map((doc) => JSON.stringify(doc)).join('\n')
  const res = await fetch(
    `${TYPESENSE_HOST}/collections/${TYPESENSE_COLLECTION}/documents/import?action=upsert`,
    {
      method: 'POST',
      headers: typesenseHeaders('text/plain'),
      body,
    },
  )
  if (!res.ok) {
    throw new Error(`Failed to import Typesense documents: ${res.status} ${await res.text()}`)
  }

  const results = (await res.text()).trim().split('\n').filter(Boolean)
  for (const line of results) {
    const parsed = JSON.parse(line) as { success?: boolean; error?: string }
    if (parsed.success === false) {
      throw new Error(`Typesense import failed: ${parsed.error ?? line}`)
    }
  }
}

async function clearTypesense(): Promise<void> {
  const res = await fetch(`${TYPESENSE_HOST}/collections/${TYPESENSE_COLLECTION}`, {
    method: 'DELETE',
    headers: typesenseHeaders(),
  })
  if (res.status === 404) {
    return
  }
  if (!res.ok) {
    throw new Error(`Failed to delete Typesense collection: ${res.status} ${await res.text()}`)
  }
}

export async function seedSearchBackend(): Promise<void> {
  if (SEARCH_BACKEND === 'typesense') {
    await seedTypesense()
    return
  }
  await seedMeilisearch()
}

export async function clearSearchBackend(): Promise<void> {
  if (SEARCH_BACKEND === 'typesense') {
    await clearTypesense()
    return
  }
  await clearMeilisearch()
}
