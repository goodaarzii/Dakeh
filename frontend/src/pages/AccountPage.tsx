import { useState } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { AddressForm } from '../components/AddressForm'
import { useDialogs } from '../components/Dialogs'
import { PinIcon, PlusIcon } from '../components/icons'
import { EmptyState, ErrorState, Spinner } from '../components/States'
import { useToast } from '../components/Toast'
import { ApiError, api } from '../lib/api'
import { useAuth } from '../lib/auth'
import { friendlyError } from '../lib/messages'
import type { Address, StatusResponse } from '../lib/types'

export default function AccountPage() {
  const [params] = useSearchParams()
  const tab = params.get('tab') === 'addresses' ? 'addresses' : 'profile'

  const tabClass = (active: boolean) =>
    `rounded-full px-4 py-1.5 font-semibold transition-colors ${active ? 'bg-ink text-paper' : 'hover:bg-mist/35'}`

  return (
    <div>
      <h1 className="text-4xl">Your account</h1>
      <nav className="mt-6 flex gap-2" aria-label="Account sections">
        <Link to="/account" className={tabClass(tab === 'profile')} aria-current={tab === 'profile' ? 'page' : undefined}>
          Profile
        </Link>
        <Link to="/account?tab=addresses" className={tabClass(tab === 'addresses')} aria-current={tab === 'addresses' ? 'page' : undefined}>
          Addresses
        </Link>
      </nav>

      <div className="mt-8">{tab === 'profile' ? <ProfileTab /> : <AddressesTab />}</div>
    </div>
  )
}

function ProfileTab() {
  const { user, logout, logoutAll } = useAuth()
  const navigate = useNavigate()
  const toast = useToast()
  const dialogs = useDialogs()
  const [busy, setBusy] = useState(false)

  if (!user) return null

  async function onLogout() {
    setBusy(true)
    await logout()
    navigate('/', { replace: true })
  }

  async function onLogoutAll() {
    const ok = await dialogs.confirm({
      title: 'Log out everywhere?',
      message: 'Every device where you are logged in will have to log in again.',
      confirmLabel: 'Log out everywhere',
      danger: true,
    })
    if (!ok) return
    setBusy(true)
    try {
      await logoutAll()
      navigate('/', { replace: true })
    } catch (err) {
      toast(friendlyError(err), { type: 'error' })
      setBusy(false)
    }
  }

  const rows: [string, string][] = [
    ['Name', user.full_name],
    ['Email', user.email],
    ['Phone', user.phone],
    ['Member since', new Date(user.created_at).toLocaleDateString('en-GB', { dateStyle: 'long' })],
  ]

  return (
    <div className="max-w-xl">
      <dl className="divide-y-[1.5px] divide-mist border-y-[1.5px] border-mist">
        {rows.map(([label, value]) => (
          <div key={label} className="grid grid-cols-[10rem_1fr] gap-4 py-4">
            <dt className="text-ink/70">{label}</dt>
            <dd className="font-semibold">{value}</dd>
          </div>
        ))}
      </dl>
      <div className="mt-8 flex gap-3">
        <button type="button" className="btn btn-primary" onClick={() => void onLogout()} disabled={busy}>
          Log out
        </button>
        <button type="button" className="btn btn-outline" onClick={() => void onLogoutAll()} disabled={busy}>
          Log out everywhere
        </button>
      </div>
    </div>
  )
}

function AddressesTab() {
  const queryClient = useQueryClient()
  const toast = useToast()
  const dialogs = useDialogs()
  const [editing, setEditing] = useState<Address | 'new' | null>(null)
  const [busyId, setBusyId] = useState<string | null>(null)

  const query = useQuery({
    queryKey: ['addresses'],
    queryFn: () => api.get<Address[]>('/users/address'),
  })

  const addresses = query.data ?? []
  const hasDefault = addresses.some((a) => a.is_default)

  async function onDelete(address: Address) {
    const ok = await dialogs.confirm({
      title: 'Delete this address?',
      message: `${address.label || address.street} will be removed from your account.`,
      confirmLabel: 'Delete',
      danger: true,
    })
    if (!ok) return
    setBusyId(address.id)
    try {
      await api.delete<StatusResponse>(`/users/address/${address.id}`)
      toast('Address deleted.', { type: 'success' })
    } catch (err) {
      // 404 means it is already gone, which is what the customer wanted.
      if (!(err instanceof ApiError && err.status === 404)) toast(friendlyError(err), { type: 'error' })
    } finally {
      setBusyId(null)
      await queryClient.invalidateQueries({ queryKey: ['addresses'] })
    }
  }

  if (query.isPending) return <Spinner label="Loading your addresses" />
  if (query.isError) return <ErrorState error={query.error} onRetry={() => void query.refetch()} />

  if (editing) {
    return (
      <div className="max-w-3xl">
        <AddressForm
          key={editing === 'new' ? 'new' : editing.id}
          initial={editing === 'new' ? undefined : editing}
          hasDefault={hasDefault}
          onCancel={() => setEditing(null)}
          onDone={() => {
            setEditing(null)
            toast('Address saved.', { type: 'success' })
          }}
        />
      </div>
    )
  }

  return (
    <div>
      <div className="flex items-center justify-between gap-4">
        <p className="text-ink/75">Where we deliver your orders. You pick the exact spot on a map.</p>
        <button type="button" className="btn btn-primary" onClick={() => setEditing('new')}>
          <PlusIcon className="h-4 w-4" />
          Add an address
        </button>
      </div>

      {addresses.length === 0 ? (
        <div className="mt-8">
          <EmptyState title="No addresses yet">
            <p>Add one now so checkout will be quick later.</p>
          </EmptyState>
        </div>
      ) : (
        <ul className="mt-8 grid grid-cols-2 gap-6">
          {addresses.map((a) => (
            <li key={a.id} className="flex flex-col rounded-lg border-[1.5px] border-mist bg-paper p-5">
              <div className="flex items-center gap-2">
                <PinIcon className="h-5 w-5 text-steel" />
                <h3 className="text-xl">{a.label || 'Address'}</h3>
                {a.is_default && <span className="rounded bg-ink px-2 py-0.5 text-xs font-semibold text-paper">Default</span>}
              </div>
              <address className="mt-3 flex-1 not-italic text-ink/85">
                <p className="font-semibold text-ink">{a.recipient_name}</p>
                <p>{a.street}</p>
                <p>
                  {a.city}, {a.province}
                  {a.postal_code ? ` ${a.postal_code}` : ''}
                </p>
                <p>{a.country}</p>
                <p className="mt-1">{a.phone}</p>
              </address>
              <div className="mt-4 flex gap-2">
                <button type="button" className="btn btn-outline btn-sm" onClick={() => setEditing(a)} disabled={busyId === a.id}>
                  Edit
                </button>
                <button type="button" className="btn btn-danger btn-sm" onClick={() => void onDelete(a)} disabled={busyId === a.id}>
                  Delete
                </button>
              </div>
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
