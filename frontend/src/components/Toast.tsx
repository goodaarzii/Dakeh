import { createContext, useCallback, useContext, useRef, useState, type ReactNode } from 'react'
import { Link } from 'react-router-dom'

export interface ToastOptions {
  type?: 'info' | 'success' | 'error'
  action?: { label: string; to: string }
}

interface ToastItem {
  id: number
  message: string
  type: 'info' | 'success' | 'error'
  action?: { label: string; to: string }
}

type ToastFn = (message: string, options?: ToastOptions) => void

const ToastContext = createContext<ToastFn>(() => {})

export function ToastProvider({ children }: { children: ReactNode }) {
  const [items, setItems] = useState<ToastItem[]>([])
  const nextId = useRef(1)

  const remove = useCallback((id: number) => {
    setItems((prev) => prev.filter((t) => t.id !== id))
  }, [])

  // useCallback keeps this function the same between renders, so it is safe to use in effect dependencies.
  const toast = useCallback<ToastFn>(
    (message, options = {}) => {
      const id = nextId.current++
      setItems((prev) => [...prev, { id, message, type: options.type ?? 'info', action: options.action }])
      window.setTimeout(() => remove(id), options.action ? 6000 : 4500)
    },
    [remove],
  )

  const bar = { info: 'bg-mist', success: 'bg-moss', error: 'bg-brick' } as const

  return (
    <ToastContext.Provider value={toast}>
      {children}
      <div className="pointer-events-none fixed bottom-6 left-1/2 z-[60] flex -translate-x-1/2 flex-col items-center gap-2" aria-live="polite">
        {items.map((t) => (
          <div
            key={t.id}
            role={t.type === 'error' ? 'alert' : 'status'}
            className="pointer-events-auto flex max-w-[28rem] items-stretch overflow-hidden rounded-lg bg-ink text-paper shadow-lg"
          >
            <span className={`w-1.5 shrink-0 ${bar[t.type]}`} />
            <div className="flex items-center gap-4 px-4 py-3 text-[0.95rem]">
              <span>{t.message}</span>
              {t.action && (
                <Link to={t.action.to} onClick={() => remove(t.id)} className="whitespace-nowrap font-semibold underline underline-offset-2">
                  {t.action.label}
                </Link>
              )}
            </div>
          </div>
        ))}
      </div>
    </ToastContext.Provider>
  )
}

export function useToast(): ToastFn {
  return useContext(ToastContext)
}
