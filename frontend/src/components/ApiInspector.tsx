import { useEffect, useState } from 'react'
import { API_URL } from '../config'
import { inspector, useInspectorEntries, useInspectorOpen, type InspectorEntry } from '../lib/inspector'
import { CloseIcon, TerminalIcon } from './icons'
import { useToast } from './Toast'

// Turns a recorded request into a curl command that fits the Postman variables {{baseUrl}} and {{token}}.
function toCurl(entry: InspectorEntry): string {
  const parts = [`curl -X ${entry.method} "${entry.url.replace(API_URL, '{{baseUrl}}')}"`]
  if (entry.hadAuth) parts.push('-H "Authorization: Bearer {{token}}"')
  if (entry.requestBody !== undefined) {
    parts.push('-H "Content-Type: application/json"')
    parts.push(`-d '${JSON.stringify(entry.requestBody).replace(/'/g, "'\\''")}'`)
  }
  return parts.join(' ')
}

function statusClass(entry: InspectorEntry): string {
  if (entry.status === 0) return 'bg-brick text-paper'
  if (entry.ok) return 'bg-moss text-paper'
  return 'bg-brick text-paper'
}

function pretty(value: unknown): string {
  if (value === undefined) return '(empty)'
  return typeof value === 'string' ? value : JSON.stringify(value, null, 2)
}

export function InspectorButton() {
  const open = useInspectorOpen()
  const entries = useInspectorEntries()
  return (
    <button
      type="button"
      onClick={() => inspector.setOpen(!open)}
      className="fixed bottom-6 right-6 z-40 flex items-center gap-2 rounded-full bg-ink px-4 py-2.5 font-semibold text-paper shadow-lg transition-colors hover:bg-ink-deep"
      aria-label="Open the API inspector"
    >
      <TerminalIcon className="h-5 w-5" />
      API inspector
      <span className="min-w-6 rounded-full bg-sand px-1.5 text-center text-sm text-ink">{entries.length}</span>
    </button>
  )
}

export function ApiInspector() {
  const open = useInspectorOpen()
  const entries = useInspectorEntries()
  const toast = useToast()
  const [selectedId, setSelectedId] = useState<number | null>(null)

  useEffect(() => {
    if (!open) return
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') inspector.setOpen(false)
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [open])

  if (!open) return null

  const selected = entries.find((e) => e.id === selectedId) ?? entries[0]

  async function copyCurl(entry: InspectorEntry) {
    try {
      await navigator.clipboard.writeText(toCurl(entry))
      toast('curl command copied. It uses {{baseUrl}} and {{token}}.', { type: 'success' })
    } catch {
      toast('Could not copy to the clipboard.', { type: 'error' })
    }
  }

  return (
    <aside
      className="fixed right-0 top-0 z-50 flex h-full w-[36rem] max-w-full flex-col border-l-[1.5px] border-ink bg-paper shadow-2xl"
      aria-label="API inspector"
    >
      <header className="flex items-start justify-between gap-4 border-b-[1.5px] border-mist px-5 py-4">
        <div>
          <h2 className="text-xl">API inspector</h2>
          <p className="text-sm text-ink/70">Every request this page sends to the Go API, newest first.</p>
        </div>
        <div className="flex items-center gap-1">
          <button type="button" className="btn btn-quiet btn-sm" onClick={() => inspector.clear()}>
            Clear
          </button>
          <button type="button" className="btn btn-quiet btn-sm" onClick={() => inspector.setOpen(false)} aria-label="Close">
            <CloseIcon className="h-5 w-5" />
          </button>
        </div>
      </header>

      {entries.length === 0 ? (
        <p className="p-5 text-ink/75">Nothing yet. Click around the shop and the requests will show up here.</p>
      ) : (
        <>
          <ul className="max-h-[38%] shrink-0 overflow-y-auto border-b-[1.5px] border-mist">
            {entries.map((entry) => (
              <li key={entry.id}>
                <button
                  type="button"
                  onClick={() => setSelectedId(entry.id)}
                  className={`flex w-full items-center gap-3 px-5 py-2 text-left text-sm hover:bg-mist/25 ${
                    selected?.id === entry.id ? 'bg-mist/35' : ''
                  }`}
                >
                  <span className="w-14 shrink-0 font-bold">{entry.method}</span>
                  <span className="min-w-0 flex-1 truncate">{entry.path}</span>
                  <span className={`rounded px-1.5 py-0.5 text-xs font-bold ${statusClass(entry)}`}>
                    {entry.status === 0 ? 'ERR' : entry.status}
                  </span>
                  <span className="w-14 shrink-0 text-right text-ink/65">{entry.durationMs} ms</span>
                </button>
              </li>
            ))}
          </ul>

          {selected && (
            <div className="min-h-0 flex-1 overflow-y-auto px-5 py-4">
              <div className="flex flex-wrap items-center justify-between gap-2">
                <p className="text-sm text-ink/70">{new Date(selected.time).toLocaleTimeString('en-GB')}</p>
                <button type="button" className="btn btn-outline btn-sm" onClick={() => void copyCurl(selected)}>
                  Copy as curl
                </button>
              </div>

              <h3 className="mt-3 text-base">Request</h3>
              <p className="mt-1 break-all text-sm">
                <strong>{selected.method}</strong> {selected.url}
              </p>
              <p className="mt-1 text-sm text-ink/70">
                {selected.hadAuth ? 'Sent with the access token.' : 'Sent without a token.'}
              </p>
              {selected.requestBody !== undefined && (
                <pre className="mt-2 overflow-x-auto rounded-md bg-fog p-3 font-mono text-[0.8rem] leading-relaxed">
                  {pretty(selected.requestBody)}
                </pre>
              )}

              <h3 className="mt-5 text-base">
                Response{' '}
                <span className={`ml-1 rounded px-1.5 py-0.5 align-middle text-xs font-bold ${statusClass(selected)}`}>
                  {selected.status === 0 ? 'no answer' : selected.status}
                </span>
              </h3>
              {selected.error && <p className="mt-1 text-sm text-brick">{selected.error}</p>}
              <pre className="mt-2 overflow-x-auto rounded-md bg-fog p-3 font-mono text-[0.8rem] leading-relaxed">
                {pretty(selected.responseBody)}
              </pre>
              <p className="mt-3 text-xs text-ink/60">Passwords are hidden and tokens are shortened in this panel.</p>
            </div>
          )}
        </>
      )}
    </aside>
  )
}
