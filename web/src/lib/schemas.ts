import { z } from 'zod'

export const IndexDocumentSchema = z.object({
  doc_id: z.string(),
  source_id: z.number(),
  type: z.enum(['file', 'folder']),
  name: z.string(),
  path_text: z.string(),
  parent_id: z.number(),
  modified_at: z.number(),
  created_at: z.number(),
  size: z.number(),
  sha1: z.string(),
  in_trash: z.boolean(),
  is_deleted: z.boolean(),
  highlighted_name: z.string().optional().default(''),
})

export type IndexDocument = z.infer<typeof IndexDocumentSchema>

export const SearchResponseSchema = z.object({
  items: z.array(IndexDocumentSchema),
  total: z.number(),
})

export type SearchResponse = z.infer<typeof SearchResponseSchema>

export const DownloadURLResponseSchema = z.object({
  file_id: z.number(),
  download_url: z.string().min(1),
})

export type DownloadURLResponse = z.infer<typeof DownloadURLResponseSchema>

export const ErrorResponseSchema = z.object({
  code: z.string(),
  message: z.string(),
  request_id: z.string().optional(),
})

export type ErrorResponse = z.infer<typeof ErrorResponseSchema>
