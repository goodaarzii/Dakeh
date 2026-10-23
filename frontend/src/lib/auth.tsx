import { createContext, useCallback, useContext, useEffect, useMemo, useState, type ReactNode } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { ApiError, api, setSessionExpiredHandler, tokenStore } from './api'
import { useToast } from '../components/Toast'
import type { AuthResponse, User } from './types'

export interface RegisterInput {
  full_name: string
  email: string
  phone: string
  password: string
}

interface AuthContextValue {
  user: User | null
  ready: boolean // false until we know whether a saved session is still valid
  login: (email: string, password: string) => Promise<void>
  register: (input: RegisterInput) => Promise<void>
  logout: () => Promise<void>
  logoutAll: () => Promise<void>
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [ready, setReady] = useState(false)
  const queryClient = useQueryClient()
  const toast = useToast()

  // The API client calls this when a refresh token is rejected: the session is over.
  useEffect(() => {
    setSessionExpiredHandler(() => {
      setUser(null)
      queryClient.removeQueries({ queryKey: ['cart'] })
      queryClient.removeQueries({ queryKey: ['favorites'] })
      queryClient.removeQueries({ queryKey: ['my-review'] })
      toast('Your session ended. Please log in again.')
    })
    return () => setSessionExpiredHandler(null)
  }, [queryClient, toast])

  // On start: if tokens are saved, ask the API who we are.
  useEffect(() => {
    let cancelled = false
    async function boot() {
      if (!tokenStore.getAccess() && !tokenStore.getRefresh()) {
        setReady(true)
        return
      }
      try {
        const me = await api.get<User>('/users/me')
        if (!cancelled) setUser(me)
      } catch (err) {
        if (err instanceof ApiError && err.status !== 0) tokenStore.clear()
      } finally {
        if (!cancelled) setReady(true)
      }
    }
    void boot()
    return () => {
      cancelled = true
    }
  }, [])

  const startSession = useCallback((res: AuthResponse) => {
    tokenStore.set(res.access_token, res.refresh_token)
    setUser(res.user)
  }, [])

  const login = useCallback(
    async (email: string, password: string) => {
      const res = await api.post<AuthResponse>('/auth/login', { email, password }, { auth: false })
      startSession(res)
    },
    [startSession],
  )

  const register = useCallback(
    async (input: RegisterInput) => {
      const res = await api.post<AuthResponse>('/auth/register', input, { auth: false })
      startSession(res)
    },
    [startSession],
  )

  const endSession = useCallback(() => {
    tokenStore.clear()
    setUser(null)
    queryClient.removeQueries({ queryKey: ['cart'] })
    queryClient.removeQueries({ queryKey: ['favorites'] })
    queryClient.removeQueries({ queryKey: ['my-review'] })
  }, [queryClient])

  const logout = useCallback(async () => {
    const refresh = tokenStore.getRefresh()
    if (refresh) {
      try {
        await api.post('/auth/logout', { refresh_token: refresh }, { auth: false })
      } catch {
        // Logging out must always work on this device, even if the server cannot be reached.
      }
    }
    endSession()
  }, [endSession])

  const logoutAll = useCallback(async () => {
    await api.post('/auth/logout-all')
    endSession()
  }, [endSession])

  const value = useMemo(
    () => ({ user, ready, login, register, logout, logoutAll }),
    [user, ready, login, register, logout, logoutAll],
  )
  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used inside AuthProvider')
  return ctx
}
