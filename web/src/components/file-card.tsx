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
    <article className="group frost-panel rounded-2xl px-4 py-4 sm:px-5">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex min-w-0 flex-1 items-start gap-4">
          <div className={`flex h-12 w-12 shrink-0 items-center justify-center rounded-xl ring-1 ring-inset ring-white/60 ${icon.bg} ${icon.text}`}>
            <FileIcon category={icon.category} />
          </div>
          <div className="min-w-0 flex-1 pt-1">
            {doc.highlighted_name ? (
              <h3
                className="truncate text-[15px] font-semibold tracking-[-0.01em] text-slate-900"
                title={doc.name}
                dangerouslySetInnerHTML={{ __html: displayName }}
              />
            ) : (
              <h3
                className="truncate text-[15px] font-semibold tracking-[-0.01em] text-slate-900"
                title={doc.name}
              >
                {displayName}
              </h3>
            )}
            <p className="mt-1 truncate text-[13px] text-slate-600">
              {size}
              <span className="mx-1.5 text-slate-300">&middot;</span>
              {date}
              <span className="mx-1.5 text-slate-300">&middot;</span>
              <span className="font-mono tracking-tight">ID: {doc.source_id}</span>
            </p>
          </div>
        </div>
        <DownloadButton status={downloadStatus} onClick={onDownload} />
      </div>
    </article>
  )
})
