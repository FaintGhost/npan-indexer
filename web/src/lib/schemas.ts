import { z } from 'zod'

export const FileCategorySchema = z.enum(['doc', 'image', 'video', 'archive', 'other'])
export type FileCategory = z.infer<typeof FileCategorySchema>

export const IndexDocumentSchema = z.object({
  doc_id: z.string(),
  source_id: z.number().int(),
  type: z.enum(['file', 'folder']),
  name: z.string(),
  path_text: z.string(),
  parent_id: z.number().int(),
  modified_at: z.number().int(),
  created_at: z.number().int(),
  size: z.number().int(),
  sha1: z.string(),
  in_trash: z.boolean(),
  is_deleted: z.boolean(),
  file_category: FileCategorySchema.optional(),
  highlighted_name: z.string().optional(),
})
export type IndexDocument = z.infer<typeof IndexDocumentSchema>

export const SearchResponseSchema = z.object({
  items: z.array(IndexDocumentSchema),
  total: z.number().int(),
})
export type SearchResponse = z.infer<typeof SearchResponseSchema>

export const DownloadURLResponseSchema = z.object({
  file_id: z.number().int(),
  download_url: z.string().url().min(1),
})
export type DownloadURLResponse = z.infer<typeof DownloadURLResponseSchema>

export const ErrorResponseSchema = z.object({
  code: z.string(),
  message: z.string(),
  request_id: z.string().optional(),
})
export type ErrorResponse = z.infer<typeof ErrorResponseSchema>
