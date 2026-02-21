import { useState } from 'react'

interface ApiKeyDialogProps {
  open: boolean
  onSubmit: (key: string) => void
  error?: string | null
  loading?: boolean
}

export function ApiKeyDialog({ open, onSubmit, error, loading }: ApiKeyDialogProps) {
  const [key, setKey] = useState('')
  const [localError, setLocalError] = useState('')

  if (!open) return null

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!key.trim()) {
      setLocalError('请输入 API Key')
      return
    }
    setLocalError('')
    onSubmit(key)
  }

  const displayError = error || localError

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-sm">
      <div className="mx-4 w-full max-w-sm rounded-2xl bg-white p-6 shadow-2xl">
        <h2 className="text-lg font-semibold text-slate-900">管理认证</h2>
        <p className="mt-1 text-sm text-slate-500">请输入管理 API Key 以访问此页面</p>

        <form onSubmit={handleSubmit} className="mt-4">
          <input
            type="password"
            value={key}
            onChange={(e) => setKey(e.target.value)}
            placeholder="输入 API Key"
            className="w-full rounded-xl border border-slate-200 px-4 py-3 text-sm outline-none transition-shadow focus:border-blue-300 focus:ring-4 focus:ring-blue-100"
            autoFocus
          />

          {displayError && (
            <p className="mt-2 text-sm text-rose-500">{displayError}</p>
          )}

          <button
            type="submit"
            disabled={loading}
            className="mt-4 flex w-full items-center justify-center rounded-xl bg-slate-900 py-3 text-sm font-medium text-white transition-colors hover:bg-slate-800 disabled:opacity-60"
          >
            {loading ? '验证中...' : '确认'}
          </button>
        </form>
      </div>
    </div>
  )
}
