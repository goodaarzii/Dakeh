import { useSyncExternalStore } from 'react'

// A tiny store that remembers every request the app makes, for the "API inspector" panel.

export interface InspectorEntry {
  id: number
  time: number
  method: string
  url: string
  path: string
  requestBody?: unknown
  hadAuth: boolean
  status: number // 0 = the server did not answer
  ok: boolean
  durationMs: number
  responseBody?: unknown
  error?: string
}

type Listener = () => void

let entries: InspectorEntry[] = []
let nextId = 1
let isOpen = false
const listeners = new Set<Listener>()

function emit() {
  listeners.forEach((l) => l())
}

function subscribe(listener: Listener) {
  listeners.add(listener)
  return () => {
    listeners.delete(listener)
  }
}

export const inspector = {
  add(entry: Omit<InspectorEntry, 'id'>) {
    entries = [{ ...entry, id: nextId++ }, ...entries].slice(0, 100)
    emit()
  },
  clear() {
    entries = []
    emit()
  },
  setOpen(open: boolean) {
    isOpen = open
    emit()
  },
}

export function useInspectorEntries(): InspectorEntry[] {
  return useSyncExternalStore(subscribe, () => entries)
}

export function useInspectorOpen(): boolean {
  return useSyncExternalStore(subscribe, () => isOpen)
}

// Passwords are hidden and tokens are shortened before anything is stored for display.
const TOKEN_KEYS = new Set(['access_token', 'refresh_token'])

export function redact(value: unknown): unknown {
  if (Array.isArray(value)) return value.map(redact)
  if (value && typeof value === 'object') {
    const out: Record<string, unknown> = {}
    for (const [key, v] of Object.entries(value as Record<string, unknown>)) {
      if (key === 'password') out[key] = '••••••••'
      else if (TOKEN_KEYS.has(key) && typeof v === 'string') out[key] = v.slice(0, 10) + '…'
      else out[key] = redact(v)
    }
    return out
  }
  return value
}
