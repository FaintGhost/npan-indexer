import { memo } from 'react'
import type { IndexDocument } from '@/lib/schemas'
import { formatBytes, formatTime } from '@/lib/format'
import { getFileIcon } from '@/lib/file-icon'
import { FileIcon } from './file-icon'

interface FileCardProps {
  doc: IndexDocument
  onDownload: (sourceId: number) => void
}

export const FileCard = memo(function FileCard({ doc, onDownload }: FileCardProps) {
  const displayName = doc.highlighted_name || doc.name
  const size = formatBytes(doc.size)
  const date = formatTime(doc.modified_at)
  const icon = getFileIcon(doc.name)

  return (
    <article className="group rounded-2xl border border-slate-200 bg-white px-4 py-4 shadow-sm transition hover:-translate-y-0.5 hover:shadow-md sm:px-5">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex items-start gap-4 min-w-0 flex-1">
          <div
            className={`flex shrink-0 items-center justify-center w-12 h-12 rounded-xl ${icon.bg} ${icon.text}`}
          >
            <FileIcon category={icon.category} />
          </div>
          <div className="min-w-0 flex-1 pt-1">
            {doc.highlighted_name ? (
              <h3
                className="truncate text-[15px] font-semibold text-slate-900"
                title={doc.name}
                dangerouslySetInnerHTML={{ __html: displayName }}
              />
            ) : (
              <h3
                className="truncate text-[15px] font-semibold text-slate-900"
                title={doc.name}
              >
                {displayName}
              </h3>
            )}
            <p className="mt-1 text-[13px] text-slate-500 truncate">
              {size}
              <span className="mx-1.5 text-slate-300">&middot;</span>
              {date}
              <span className="mx-1.5 text-slate-300">&middot;</span>
              ID: {doc.source_id}
            </p>
          </div>
        </div>
        <button
          type="button"
          onClick={() => onDownload(doc.source_id)}
          className="relative flex h-10 w-full sm:w-auto min-w-[96px] shrink-0 items-center justify-center rounded-xl bg-slate-900 px-4 text-sm font-medium text-white transition-all hover:bg-slate-800"
        >
          <span className="flex items-center gap-1.5">
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="14"
              height="14"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2.5"
              strokeLinecap="round"
              strokeLinejoin="round"
            >
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
              <polyline points="7 10 12 15 17 10" />
              <line x1="12" x2="12" y1="15" y2="3" />
            </svg>
            下载
          </span>
        </button>
      </div>
    </article>
  )
})
