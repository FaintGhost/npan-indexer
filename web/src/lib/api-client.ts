import type { ZodTypeAny } from 'zod'
import { ErrorResponseSchema } from './schemas'

export class ApiError extends Error {
  readonly status: number
  readonly code: string

  constructor(status: number, code: string, message: string) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.code = code
  }
}

function buildURL(path: string, params?: Record<string, unknown>): string {
  const url = new URL(path, window.location.origin)
  if (params) {
    for (const [key, value] of Object.entries(params)) {
      if (value === undefined || value === null || value === '') continue
      url.searchParams.set(key, String(value))
    }
  }
  return url.toString()
}

export async function apiGet<S extends ZodTypeAny>(
  path: string,
  params: Record<string, unknown>,
  schema: S,
  options?: { signal?: AbortSignal; headers?: Record<string, string> },
): Promise<S['_output']> {
  const url = buildURL(path, params)
  const res = await fetch(url, {
    signal: options?.signal,
    headers: options?.headers,
  })

  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    const parsed = ErrorResponseSchema.safeParse(body)
    if (parsed.success) {
      throw new ApiError(res.status, parsed.data.code, parsed.data.message)
    }
    throw new ApiError(res.status, 'UNKNOWN', `HTTP ${res.status}`)
  }

  const data = await res.json()
  return schema.parse(data)
}

export async function apiPost<T = { message: string }>(
  path: string,
  body: unknown,
  options?: { signal?: AbortSignal; headers?: Record<string, string> },
): Promise<T> {
  const url = new URL(path, window.location.origin).toString()
  const res = await fetch(url, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
    body: JSON.stringify(body),
    signal: options?.signal,
  })

  if (!res.ok) {
    const errBody = await res.json().catch(() => ({}))
    const parsed = ErrorResponseSchema.safeParse(errBody)
    if (parsed.success) {
      throw new ApiError(res.status, parsed.data.code, parsed.data.message)
    }
    throw new ApiError(res.status, 'UNKNOWN', `HTTP ${res.status}`)
  }

  return res.json() as Promise<T>
}

export async function apiDelete<T = { message: string }>(
  path: string,
  options?: { signal?: AbortSignal; headers?: Record<string, string> },
): Promise<T> {
  const url = new URL(path, window.location.origin).toString()
  const res = await fetch(url, {
    method: 'DELETE',
    headers: options?.headers,
    signal: options?.signal,
  })

  if (!res.ok) {
    const errBody = await res.json().catch(() => ({}))
    const parsed = ErrorResponseSchema.safeParse(errBody)
    if (parsed.success) {
      throw new ApiError(res.status, parsed.data.code, parsed.data.message)
    }
    throw new ApiError(res.status, 'UNKNOWN', `HTTP ${res.status}`)
  }

  return res.json() as Promise<T>
}
