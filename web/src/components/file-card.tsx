import { memo } from 'react'
import type { IndexDocument } from '@/lib/schemas'
import { formatBytes, formatTime } from '@/lib/format'
import { getFileIcon } from '@/lib/file-icon'
import { FileIcon } from './file-icon'
import { DownloadButton } from './download-button'

type DownloadStatus = 'idle' | 'loading' | 'success' | 'error'

interface FileCardProps {
  doc: IndexDocument
  downloadStatus: DownloadStatus
  onDownload: () => void
}

export const FileCard = memo(function FileCard({ doc, downloadStatus, onDownload }: FileCardProps) {
  const displayName = doc.highlighted_name || doc.name
  const size = formatBytes(doc.size)
  const date = formatTime(doc.modified_at)
  const icon = getFileIcon(doc.name)

  return (
    <article className="group rounded-2xl border border-slate-200 bg-white px-4 py-4 shadow-sm transition hover:-translate-y-0.5 hover:shadow-md sm:px-5">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex min-w-0 flex-1 items-start gap-4">
          <div
            className={`flex h-12 w-12 shrink-0 items-center justify-center rounded-xl ${icon.bg} ${icon.text}`}
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
            <p className="mt-1 truncate text-[13px] text-slate-500">
              {size}
              <span className="mx-1.5 text-slate-300">&middot;</span>
              {date}
              <span className="mx-1.5 text-slate-300">&middot;</span>
              ID: {doc.source_id}
            </p>
          </div>
        </div>
        <DownloadButton status={downloadStatus} onClick={onDownload} />
      </div>
    </article>
  )
})
