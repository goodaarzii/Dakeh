import type { ReactNode } from 'react'
import { API_URL } from '../config'
import { ApiError } from '../lib/api'
import { friendlyError } from '../lib/messages'

export function Spinner({ label = 'Loading' }: { label?: string }) {
  return (
    <div role="status" className="flex items-center gap-3 py-10 text-ink/70">
      <span className="h-5 w-5 animate-spin rounded-full border-2 border-mist border-t-ink" />
      <span>{label}…</span>
    </div>
  )
}

export function ErrorState({ error, onRetry }: { error: unknown; onRetry?: () => void }) {
  const offline = error instanceof ApiError && error.status === 0
  return (
    <div role="alert" className="rounded-lg border-[1.5px] border-brick/60 bg-paper p-6">
      <h2 className="text-xl">{offline ? 'The shop cannot reach its server' : 'Something went wrong'}</h2>
      <p className="mt-2 text-ink/80">{friendlyError(error)}</p>
      {offline && (
        <p className="mt-2 text-sm text-ink/70">
          Check that the Go API is running and that it answers at <code className="rounded bg-fog px-1.5 py-0.5">{API_URL}</code>.
        </p>
      )}
      {onRetry && (
        <button type="button" onClick={onRetry} className="btn btn-outline btn-sm mt-4">
          Try again
        </button>
      )}
    </div>
  )
}

export function EmptyState({ title, children }: { title: string; children?: ReactNode }) {
  return (
    <div className="rounded-lg border-[1.5px] border-dashed border-mist px-6 py-14 text-center">
      <h2 className="text-2xl">{title}</h2>
      {children && <div className="mx-auto mt-3 max-w-md text-ink/75">{children}</div>}
    </div>
  )
}
