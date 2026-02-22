import { z } from 'zod'
import {
  zIndexDocument,
  zQueryResult,
  zDownloadUrlResponse,
  zErrorResponse,
} from '@/api/generated/zod.gen'

export const IndexDocumentSchema = zIndexDocument
export type IndexDocument = z.infer<typeof IndexDocumentSchema>

export const SearchResponseSchema = zQueryResult
export type SearchResponse = z.infer<typeof SearchResponseSchema>

export const DownloadURLResponseSchema = zDownloadUrlResponse
export type DownloadURLResponse = z.infer<typeof DownloadURLResponseSchema>

export const ErrorResponseSchema = zErrorResponse
export type ErrorResponse = z.infer<typeof ErrorResponseSchema>
