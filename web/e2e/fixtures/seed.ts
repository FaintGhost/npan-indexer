const MEILI_HOST = process.env.MEILI_HOST ?? 'http://localhost:7700'
const MEILI_API_KEY = process.env.MEILI_API_KEY ?? 'ci-test-meili-key-5678'
const MEILI_INDEX = process.env.MEILI_INDEX ?? 'npan_items'

interface TestDocument {
  source_id: number
  name: string
  size: number
  modified_at: number
  type: string
  file_category: 'doc' | 'image' | 'video' | 'archive' | 'other'
  is_deleted: boolean
  in_trash: boolean
}

const namedDocs: TestDocument[] = [
  {
    source_id: 9001,
    name: 'quarterly-report-2024.pdf',
    size: 1048576,
    modified_at: 1718444400, // 2024-06-15T10:30:00Z
    type: 'file',
    file_category: 'doc',
    is_deleted: false,
    in_trash: false,
  },
  {
    source_id: 9002,
    name: 'project-design-spec.docx',
    size: 524288,
    modified_at: 1721484000, // 2024-07-20T14:00:00Z
    type: 'file',
    file_category: 'doc',
    is_deleted: false,
    in_trash: false,
  },
  {
    source_id: 9003,
    name: 'architecture-diagram.png',
    size: 2097152,
    modified_at: 1722502800, // 2024-08-01T09:00:00Z
    type: 'file',
    file_category: 'image',
    is_deleted: false,
    in_trash: false,
  },
  {
    source_id: 9004,
    name: 'launch-demo.mp4',
    size: 3145728,
    modified_at: 1725170400, // 2024-09-01T10:00:00Z
    type: 'file',
    file_category: 'video',
    is_deleted: false,
    in_trash: false,
  },
  {
    source_id: 9005,
    name: 'release-bundle.tar.gz',
    size: 4194304,
    modified_at: 1727852400, // 2024-10-02T11:00:00Z
    type: 'file',
    file_category: 'archive',
    is_deleted: false,
    in_trash: false,
  },
]

const bulkDocs: TestDocument[] = Array.from({ length: 35 }, (_, i) => ({
  source_id: 1000 + i,
  name: `test-file-${String(i).padStart(3, '0')}.txt`,
  size: 1024 * (i + 1),
  modified_at: 1704067200, // 2024-01-01T00:00:00Z
  type: 'file',
  file_category: 'doc',
  is_deleted: false,
  in_trash: false,
}))

export const TEST_DOCUMENTS: TestDocument[] = [...namedDocs, ...bulkDocs]

const FILTERABLE_ATTRS = ['source_id', 'type', 'file_category', 'is_deleted', 'in_trash']

function headers(): Record<string, string> {
  return {
    'Content-Type': 'application/json',
    Authorization: `Bearer ${MEILI_API_KEY}`,
  }
}

export async function waitForMeiliTask(taskUid: number): Promise<void> {
  const url = `${MEILI_HOST}/tasks/${taskUid}`
  while (true) {
    const res = await fetch(url, { headers: headers() })
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

export async function seedMeilisearch(): Promise<void> {
  // Create or update index with filterable attributes
  const settingsRes = await fetch(
    `${MEILI_HOST}/indexes/${MEILI_INDEX}/settings`,
    {
      method: 'PATCH',
      headers: headers(),
      body: JSON.stringify({ filterableAttributes: FILTERABLE_ATTRS }),
    },
  )
  if (!settingsRes.ok) {
    // Index may not exist yet; create it first
    const createRes = await fetch(`${MEILI_HOST}/indexes`, {
      method: 'POST',
      headers: headers(),
      body: JSON.stringify({
        uid: MEILI_INDEX,
        primaryKey: 'source_id',
      }),
    })
    if (createRes.ok) {
      const createTask = (await createRes.json()) as { taskUid: number }
      await waitForMeiliTask(createTask.taskUid)
    }

    // Retry settings update
    const retryRes = await fetch(
      `${MEILI_HOST}/indexes/${MEILI_INDEX}/settings`,
      {
        method: 'PATCH',
        headers: headers(),
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

  // Add documents
  const addRes = await fetch(
    `${MEILI_HOST}/indexes/${MEILI_INDEX}/documents`,
    {
      method: 'POST',
      headers: headers(),
      body: JSON.stringify(TEST_DOCUMENTS),
    },
  )
  if (!addRes.ok) {
    throw new Error(`Failed to add documents: ${addRes.status}`)
  }
  const addTask = (await addRes.json()) as { taskUid: number }
  await waitForMeiliTask(addTask.taskUid)
}

export async function clearMeilisearch(): Promise<void> {
  const res = await fetch(
    `${MEILI_HOST}/indexes/${MEILI_INDEX}/documents`,
    {
      method: 'DELETE',
      headers: headers(),
    },
  )
  if (!res.ok) {
    throw new Error(`Failed to delete documents: ${res.status}`)
  }
  const task = (await res.json()) as { taskUid: number }
  await waitForMeiliTask(task.taskUid)
}
