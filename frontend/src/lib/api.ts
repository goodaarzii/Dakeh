import { API_URL } from '../config'
import { inspector, redact } from './inspector'

const ACCESS_KEY = 'dakeh.access'
const REFRESH_KEY = 'dakeh.refresh'

// Tokens live in localStorage: simple and standard for a single-page app.
// (The trade-off: a script injected into the page could read them.)
export const tokenStore = {
  getAccess: () => localStorage.getItem(ACCESS_KEY),
  getRefresh: () => localStorage.getItem(REFRESH_KEY),
  set(access: string, refresh: string) {
    localStorage.setItem(ACCESS_KEY, access)
    localStorage.setItem(REFRESH_KEY, refresh)
  },
  clear() {
    localStorage.removeItem(ACCESS_KEY)
    localStorage.removeItem(REFRESH_KEY)
  },
}

export class ApiError extends Error {
  status: number
  code: string

  constructor(status: number, code: string, message: string) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.code = code
  }
}

type Query = Record<string, string | number | undefined | null>

interface Options {
  body?: unknown
  query?: Query
  // false = never send the access token and never try to refresh (login, register, refresh, logout)
  auth?: boolean
}

let onSessionExpired: (() => void) | null = null

export function setSessionExpiredHandler(handler: (() => void) | null) {
  onSessionExpired = handler
}

function buildUrl(path: string, query?: Query): string {
  const url = new URL(API_URL + path)
  if (query) {
    for (const [key, value] of Object.entries(query)) {
      if (value !== undefined && value !== null && value !== '') url.searchParams.set(key, String(value))
    }
  }
  return url.toString()
}

// One real HTTP call. Every call is recorded for the API inspector.
async function rawRequest<T>(method: string, path: string, opts: Options = {}): Promise<T> {
  const url = buildUrl(path, opts.query)
  const headers: Record<string, string> = { Accept: 'application/json' }

  let body: string | undefined
  if (opts.body !== undefined) {
    headers['Content-Type'] = 'application/json'
    body = JSON.stringify(opts.body)
  }

  const access = opts.auth === false ? null : tokenStore.getAccess()
  if (access) headers.Authorization = `Bearer ${access}`

  const started = performance.now()
  let res: Response
  try {
    res = await fetch(url, { method, headers, body })
  } catch {
    inspector.add({
      time: Date.now(),
      method,
      url,
      path,
      requestBody: redact(opts.body),
      hadAuth: !!access,
      status: 0,
      ok: false,
      durationMs: Math.round(performance.now() - started),
      error: 'The server did not answer (network error or CORS).',
    })
    throw new ApiError(0, 'network_error', 'Cannot reach the server. Is the API running?')
  }

  const text = await res.text()
  let data: unknown = undefined
  if (text) {
    try {
      data = JSON.parse(text)
    } catch {
      data = text
    }
  }

  inspector.add({
    time: Date.now(),
    method,
    url,
    path,
    requestBody: redact(opts.body),
    hadAuth: !!access,
    status: res.status,
    ok: res.ok,
    durationMs: Math.round(performance.now() - started),
    responseBody: redact(data),
  })

  if (!res.ok) {
    const apiError = (data as { error?: { code?: string; message?: string } } | undefined)?.error
    throw new ApiError(res.status, apiError?.code ?? 'http_error', apiError?.message ?? `Request failed (${res.status})`)
  }
  return data as T
}

// Several requests can get a 401 at the same moment. They all wait for ONE refresh call,
// because a refresh token can only be used once.
let refreshing: Promise<boolean> | null = null

function refreshTokens(): Promise<boolean> {
  if (!refreshing) {
    refreshing = (async () => {
      const refresh = tokenStore.getRefresh()
      if (!refresh) return false
      try {
        const res = await rawRequest<{ access_token: string; refresh_token: string }>('POST', '/auth/refresh', {
          body: { refresh_token: refresh },
          auth: false,
        })
        tokenStore.set(res.access_token, res.refresh_token)
        return true
      } catch (err) {
        // Only a real answer from the server ends the session. A network error does not.
        if (err instanceof ApiError && err.status !== 0) {
          tokenStore.clear()
          onSessionExpired?.()
        }
        return false
      }
    })().finally(() => {
      refreshing = null
    })
  }
  return refreshing
}

// The call the app uses: on a 401 it refreshes the tokens once and repeats the request.
async function request<T>(method: string, path: string, opts: Options = {}): Promise<T> {
  try {
    return await rawRequest<T>(method, path, opts)
  } catch (err) {
    const canRetry =
      err instanceof ApiError && err.status === 401 && opts.auth !== false && !!tokenStore.getRefresh()
    if (!canRetry) throw err
    const refreshed = await refreshTokens()
    if (!refreshed) throw err
    return rawRequest<T>(method, path, opts)
  }
}

export const api = {
  get: <T>(path: string, query?: Query) => request<T>('GET', path, { query }),
  post: <T>(path: string, body?: unknown, opts?: { auth?: boolean }) =>
    request<T>('POST', path, { body, auth: opts?.auth }),
  put: <T>(path: string, body?: unknown) => request<T>('PUT', path, { body }),
  patch: <T>(path: string, body?: unknown) => request<T>('PATCH', path, { body }),
  delete: <T>(path: string) => request<T>('DELETE', path),
}

// A message that is safe to show to a person.
export function errorMessage(err: unknown): string {
  if (err instanceof ApiError) return err.message
  if (err instanceof Error) return err.message
  return 'Something went wrong.'
}
