import { createContext, useCallback, useContext, useEffect, useId, useMemo, useRef, useState, type ReactNode } from 'react'
import { CloseIcon } from './icons'

// A modal built on the native <dialog> element: it traps focus and closes with Esc by itself.
export function Modal({
  open,
  onClose,
  title,
  children,
}: {
  open: boolean
  onClose: () => void
  title: string
  children: ReactNode
}) {
  const ref = useRef<HTMLDialogElement>(null)
  const titleId = useId()

  useEffect(() => {
    const dialog = ref.current
    if (!dialog) return
    if (open && !dialog.open) dialog.showModal()
    if (!open && dialog.open) dialog.close()
  }, [open])

  return (
    <dialog
      ref={ref}
      aria-labelledby={titleId}
      onClose={onClose}
      onClick={(e) => {
        // A click on the dark backdrop lands on the <dialog> itself, not on its content.
        if (e.target === e.currentTarget) onClose()
      }}
    >
      <div className="p-6">
        <div className="flex items-start justify-between gap-4">
          <h2 id={titleId} className="text-2xl">
            {title}
          </h2>
          <button type="button" onClick={onClose} className="btn btn-quiet btn-sm -mr-2 -mt-1" aria-label="Close">
            <CloseIcon className="h-5 w-5" />
          </button>
        </div>
        {children}
      </div>
    </dialog>
  )
}

interface ConfirmOptions {
  title: string
  message: string
  confirmLabel: string
  danger?: boolean
}

interface DialogApi {
  underConstruction: (feature: string) => void
  confirm: (options: ConfirmOptions) => Promise<boolean>
}

const DialogContext = createContext<DialogApi | null>(null)

export function DialogProvider({ children }: { children: ReactNode }) {
  const [feature, setFeature] = useState<string | null>(null)
  const [confirmState, setConfirmState] = useState<{ options: ConfirmOptions; resolve: (value: boolean) => void } | null>(null)

  const underConstruction = useCallback((name: string) => setFeature(name), [])
  const confirm = useCallback(
    (options: ConfirmOptions) => new Promise<boolean>((resolve) => setConfirmState({ options, resolve })),
    [],
  )

  const answer = (value: boolean) => {
    confirmState?.resolve(value)
    setConfirmState(null)
  }

  const api = useMemo(() => ({ underConstruction, confirm }), [underConstruction, confirm])

  return (
    <DialogContext.Provider value={api}>
      {children}

      <Modal open={feature !== null} onClose={() => setFeature(null)} title="Under construction">
        <p className="mt-3 text-ink/80">
          We are still building <strong>{feature}</strong>. It is not part of this test version yet.
        </p>
        <div className="mt-6 flex justify-end">
          <button type="button" className="btn btn-primary" onClick={() => setFeature(null)}>
            Got it
          </button>
        </div>
      </Modal>

      <Modal open={confirmState !== null} onClose={() => answer(false)} title={confirmState?.options.title ?? ''}>
        <p className="mt-3 text-ink/80">{confirmState?.options.message}</p>
        <div className="mt-6 flex justify-end gap-3">
          <button type="button" className="btn btn-quiet" onClick={() => answer(false)}>
            Cancel
          </button>
          <button
            type="button"
            className={`btn ${confirmState?.options.danger ? 'btn-danger' : 'btn-primary'}`}
            onClick={() => answer(true)}
          >
            {confirmState?.options.confirmLabel}
          </button>
        </div>
      </Modal>
    </DialogContext.Provider>
  )
}

export function useDialogs(): DialogApi {
  const ctx = useContext(DialogContext)
  if (!ctx) throw new Error('useDialogs must be used inside DialogProvider')
  return ctx
}
