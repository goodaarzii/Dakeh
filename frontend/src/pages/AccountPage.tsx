import { useEffect, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { AddressForm } from '../components/AddressForm'
import { useDialogs } from '../components/Dialogs'
import { PinIcon, PlusIcon, StarIcon } from '../components/icons'
import { ProductCard } from '../components/ProductCard'
import { SafeImage } from '../components/SafeImage'
import { EmptyState, ErrorState, Spinner } from '../components/States'
import { useToast } from '../components/Toast'
import { ApiError, api } from '../lib/api'
import { useAuth } from '../lib/auth'
import { friendlyError } from '../lib/messages'
import { formatPrice } from '../lib/money'
import type { Address, FavoriteList, Order, OrderList, Review, ReviewList, StatusResponse } from '../lib/types'

const PAGE_SIZE = 10
type MyReview = Review & { product_name?: string; product_slug?: string; product_image_url?: string }
type MyReviewList = Omit<ReviewList, 'items'> & { items: MyReview[] }

type AccountTab = 'profile' | 'addresses' | 'favorites' | 'reviews' | 'orders'

export default function AccountPage() {
  const [params] = useSearchParams()
  const requestedTab = params.get('tab')
  const tab: AccountTab = requestedTab === 'addresses' || requestedTab === 'favorites' || requestedTab === 'reviews' || requestedTab === 'orders'
    ? requestedTab
    : 'profile'

  const tabClass = (active: boolean) =>
    `rounded-full px-4 py-1.5 font-semibold transition-colors ${active ? 'bg-ink text-paper' : 'hover:bg-mist/35'}`

  return (
    <div>
      <h1 className="text-4xl">Your account</h1>
      <nav className="mt-6 flex flex-wrap gap-2" aria-label="Account sections">
        <Link to="/account" className={tabClass(tab === 'profile')} aria-current={tab === 'profile' ? 'page' : undefined}>
          Details
        </Link>
        <Link to="/account?tab=addresses" className={tabClass(tab === 'addresses')} aria-current={tab === 'addresses' ? 'page' : undefined}>
          Addresses
        </Link>
        <Link to="/account?tab=favorites" className={tabClass(tab === 'favorites')} aria-current={tab === 'favorites' ? 'page' : undefined}>
          Favorites
        </Link>
        <Link to="/account?tab=reviews" className={tabClass(tab === 'reviews')} aria-current={tab === 'reviews' ? 'page' : undefined}>
          My reviews
        </Link>
        <Link to="/account?tab=orders" className={tabClass(tab === 'orders')} aria-current={tab === 'orders' ? 'page' : undefined}>
          My orders
        </Link>
      </nav>

      <div className="mt-8">
        {tab === 'profile' && <ProfileTab />}
        {tab === 'addresses' && <AddressesTab />}
        {tab === 'favorites' && <FavoritesTab />}
        {tab === 'reviews' && <MyReviewsTab />}
        {tab === 'orders' && <OrdersTab />}
      </div>
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
            <dd className="font-semibold">{value || '—'}</dd>
          </div>
        ))}
      </dl>
      <div className="mt-8 flex flex-wrap gap-3">
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
      // A 404 means the address is already gone, which is what the customer wanted.
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
            void queryClient.invalidateQueries({ queryKey: ['addresses'] })
          }}
        />
      </div>
    )
  }

  return (
    <div>
      <div className="flex flex-wrap items-center justify-between gap-4">
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
        <ul className="mt-8 grid grid-cols-1 gap-6 md:grid-cols-2">
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

function FavoritesTab() {
  const { user } = useAuth()
  const queryClient = useQueryClient()
  const toast = useToast()
  const favorites = useQuery({
    queryKey: ['favorites'],
    queryFn: () => api.get<FavoriteList>('/favorites'),
    enabled: !!user,
  })

  const removeFavorite = useMutation({
    mutationFn: (productID: string) => api.delete<void>(`/favorites/${productID}`),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['favorites'] })
      toast('Removed from your favorites.', { type: 'success' })
    },
    onError: (err) => toast(friendlyError(err), { type: 'error' }),
  })

  if (favorites.isPending) return <Spinner label="Loading favorites" />
  if (favorites.isError) return <ErrorState error={favorites.error} onRetry={() => void favorites.refetch()} />

  const items = favorites.data.items ?? []
  return (
    <section aria-labelledby="favorites-heading">
      <div className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <p className="text-sm font-semibold uppercase tracking-[0.15em] text-ink/60">Your shortlist</p>
          <h2 id="favorites-heading" className="mt-1 text-3xl">Favorites</h2>
          <p className="mt-2 text-ink/70">Products you want to come back to.</p>
        </div>
        <span className="rounded-full border border-mist px-3 py-1 text-sm font-semibold">
          {favorites.data.count} {favorites.data.count === 1 ? 'product' : 'products'}
        </span>
      </div>

      {items.length === 0 ? (
        <div className="mt-6">
          <EmptyState title="No favorites yet">
            <p>Tap the heart on a product to keep it here for later.</p>
            <Link to="/" className="btn btn-primary mt-4">Explore the shop</Link>
          </EmptyState>
        </div>
      ) : (
        <div className="mt-8 grid grid-cols-1 gap-x-6 gap-y-8 sm:grid-cols-2 lg:grid-cols-3">
          {items.map((product) => (
            <div key={product.id} className="flex flex-col items-start">
              <div className="w-full">
                <ProductCard product={product} />
              </div>
              <button
                type="button"
                className="btn btn-outline btn-sm mt-3"
                disabled={removeFavorite.isPending && removeFavorite.variables === product.id}
                onClick={() => removeFavorite.mutate(product.id)}
              >
                {removeFavorite.isPending && removeFavorite.variables === product.id ? 'Removing…' : 'Remove favorite'}
              </button>
            </div>
          ))}
        </div>
      )}
    </section>
  )
}

function MyReviewsTab() {
  const { user } = useAuth()
  const queryClient = useQueryClient()
  const toast = useToast()
  const dialogs = useDialogs()
  const [page, setPage] = useState(1)
  const [editingId, setEditingId] = useState<string | null>(null)
  const [rating, setRating] = useState(5)
  const [title, setTitle] = useState('')
  const [body, setBody] = useState('')

  const reviews = useQuery({
    queryKey: ['my-reviews', user?.id, page],
    queryFn: () => api.get<MyReviewList>('/reviews', { page, page_size: PAGE_SIZE }),
    enabled: !!user,
  })

  const updateReview = useMutation({
    mutationFn: ({ reviewID, input }: { reviewID: string; input: { rating: number; title: string; body: string } }) =>
      api.patch<Review>(`/reviews/${reviewID}`, input),
    onSuccess: async () => {
      setEditingId(null)
      toast('Your review has been updated.', { type: 'success' })
      await queryClient.invalidateQueries({ queryKey: ['my-reviews', user?.id] })
      await queryClient.invalidateQueries({ queryKey: ['my-review'] })
      await queryClient.invalidateQueries({ queryKey: ['reviews'] })
      await queryClient.invalidateQueries({ queryKey: ['product'] })
    },
    onError: (err) => toast(friendlyError(err), { type: 'error' }),
  })

  const deleteReview = useMutation({
    mutationFn: (reviewID: string) => api.delete<void>(`/reviews/${reviewID}`),
    onSuccess: async () => {
      setEditingId(null)
      toast('Your review has been deleted.', { type: 'success' })
      await queryClient.invalidateQueries({ queryKey: ['my-reviews', user?.id] })
      await queryClient.invalidateQueries({ queryKey: ['my-review'] })
      await queryClient.invalidateQueries({ queryKey: ['reviews'] })
      await queryClient.invalidateQueries({ queryKey: ['product'] })
    },
    onError: (err) => toast(friendlyError(err), { type: 'error' }),
  })

  useEffect(() => {
    if (reviews.data && page > 1 && reviews.data.items.length === 0 && !reviews.isFetching) {
      setPage((current) => Math.max(1, current - 1))
    }
  }, [reviews.data, reviews.isFetching, page])

  if (reviews.isPending) return <Spinner label="Loading your reviews" />
  if (reviews.isError) return <ErrorState error={reviews.error} onRetry={() => void reviews.refetch()} />

  const items = reviews.data.items ?? []
  const busy = updateReview.isPending || deleteReview.isPending

  function startEditing(review: MyReview) {
    setEditingId(review.id)
    setRating(review.rating)
    setTitle(review.title ?? '')
    setBody(review.body)
  }

  async function onDelete(review: MyReview) {
    const ok = await dialogs.confirm({
      title: 'Delete your review?',
      message: 'This review will be permanently removed from the product page.',
      confirmLabel: 'Delete review',
      danger: true,
    })
    if (ok) deleteReview.mutate(review.id)
  }

  return (
    <section aria-labelledby="my-reviews-heading">
      <div className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <p className="text-sm font-semibold uppercase tracking-[0.15em] text-ink/60">Your feedback</p>
          <h2 id="my-reviews-heading" className="mt-1 text-3xl">My reviews</h2>
          <p className="mt-2 text-ink/70">Reviews you have written for products in the shop.</p>
        </div>
        <span className="rounded-full border border-mist px-3 py-1 text-sm font-semibold">
          {reviews.data.count} {reviews.data.count === 1 ? 'review' : 'reviews'}
        </span>
      </div>

      {items.length === 0 ? (
        <div className="mt-6">
          <EmptyState title="You haven't written any reviews yet">
            <p>Visit a product page to share your experience and help other shoppers.</p>
            <Link to="/" className="btn btn-primary mt-4">Explore the shop</Link>
          </EmptyState>
        </div>
      ) : (
        <div className="mt-6 space-y-4">
          {items.map((review) => {
            const isEditing = editingId === review.id
            const productName = review.product_name || `Product ${review.product_id.slice(0, 8)}`
            return (
              <article key={review.id} className="rounded-lg border-[1.5px] border-mist bg-paper p-5">
                <div className="flex flex-wrap items-start justify-between gap-3">
                  <div className="min-w-0">
                    <p className="text-xs font-semibold uppercase tracking-[0.12em] text-ink/60">Product review</p>
                    {review.product_slug ? (
                      <Link to={`/products/${review.product_slug}`} className="mt-1 block text-xl font-bold hover:underline">
                        {productName}
                      </Link>
                    ) : (
                      <h3 className="mt-1 text-xl">{productName}</h3>
                    )}
                    {!review.product_name && (
                      <p className="mt-1 break-all text-xs text-ink/55" title={review.product_id}>Product ID: {review.product_id}</p>
                    )}
                  </div>
                  <time className="text-sm text-ink/60" dateTime={review.created_at}>
                    {new Date(review.created_at).toLocaleDateString('en-GB', { dateStyle: 'medium' })}
                  </time>
                </div>

                {isEditing ? (
                  <form
                    className="mt-5 space-y-4 border-t border-mist pt-5"
                    onSubmit={(event) => {
                      event.preventDefault()
                      updateReview.mutate({ reviewID: review.id, input: { rating, title: title.trim(), body: body.trim() } })
                    }}
                  >
                    <fieldset>
                      <legend className="mb-2 text-sm font-semibold">Your rating</legend>
                      <div className="flex gap-1" role="radiogroup" aria-label="Your rating">
                        {[1, 2, 3, 4, 5].map((value) => (
                          <button
                            key={value}
                            type="button"
                            role="radio"
                            aria-checked={rating === value}
                            aria-label={`${value} ${value === 1 ? 'star' : 'stars'}`}
                            disabled={busy}
                            onClick={() => setRating(value)}
                            className={`rounded p-1 ${value <= rating ? 'text-steel' : 'text-mist'}`}
                          >
                            <StarIcon className="h-6 w-6" />
                          </button>
                        ))}
                      </div>
                    </fieldset>
                    <label className="block text-sm">
                      <span className="mb-1 block font-semibold">Title (optional)</span>
                      <input className="field" value={title} onChange={(event) => setTitle(event.target.value)} maxLength={120} disabled={busy} />
                    </label>
                    <label className="block text-sm">
                      <span className="mb-1 block font-semibold">Review</span>
                      <textarea
                        className="field"
                        value={body}
                        onChange={(event) => setBody(event.target.value)}
                        rows={4}
                        minLength={3}
                        maxLength={5000}
                        required
                        disabled={busy}
                      />
                      <span className="mt-1 block text-xs text-ink/60">3–5,000 characters</span>
                    </label>
                    <div className="flex flex-wrap gap-2">
                      <button type="submit" className="btn btn-primary btn-sm" disabled={busy || body.trim().length < 3}>
                        {updateReview.isPending ? 'Saving…' : 'Save changes'}
                      </button>
                      <button type="button" className="btn btn-outline btn-sm" disabled={busy} onClick={() => setEditingId(null)}>
                        Cancel
                      </button>
                    </div>
                  </form>
                ) : (
                  <>
                    <div className="mt-3 flex items-center gap-2" aria-label={`${review.rating} out of 5 stars`}>
                      <span className="inline-flex text-steel">{Array.from({ length: review.rating }, (_, index) => <StarIcon key={index} className="h-4 w-4" />)}</span>
                      <span className="text-sm font-semibold">{review.rating}/5</span>
                    </div>
                    {review.title && <h4 className="mt-3 text-lg">{review.title}</h4>}
                    <p className="mt-2 whitespace-pre-line text-ink/85">{review.body}</p>
                    {review.updated_at && review.updated_at !== review.created_at && (
                      <p className="mt-2 text-xs text-ink/55">Edited</p>
                    )}
                    <div className="mt-4 flex flex-wrap gap-2 border-t border-mist pt-4">
                      <button type="button" className="btn btn-outline btn-sm" disabled={busy} onClick={() => startEditing(review)}>
                        Edit review
                      </button>
                      <button type="button" className="btn btn-danger btn-sm" disabled={busy} onClick={() => void onDelete(review)}>
                        {deleteReview.isPending && deleteReview.variables === review.id ? 'Deleting…' : 'Delete review'}
                      </button>
                    </div>
                  </>
                )}
              </article>
            )
          })}
        </div>
      )}

      {(page > 1 || reviews.data.has_more) && (
        <div className="mt-6 flex items-center justify-between gap-3">
          <button type="button" className="btn btn-outline btn-sm" disabled={page <= 1} onClick={() => setPage((current) => Math.max(1, current - 1))}>
            Newer reviews
          </button>
          <span className="text-sm text-ink/65">Page {page}</span>
          <button type="button" className="btn btn-outline btn-sm" disabled={!reviews.data.has_more} onClick={() => setPage((current) => current + 1)}>
            Older reviews
          </button>
        </div>
      )}
    </section>
  )
}


function OrdersTab() {
  const { user } = useAuth()
  const queryClient = useQueryClient()
  const toast = useToast()
  const dialogs = useDialogs()
  const [page, setPage] = useState(1)

  const orders = useQuery({
    queryKey: ['orders', user?.id, page],
    queryFn: () => api.get<OrderList>('/orders', { page, page_size: PAGE_SIZE }),
    enabled: !!user,
  })

  const cancelOrder = useMutation({
    mutationFn: (orderID: string) => api.post<Order>(`/orders/${orderID}/cancel`),
    onSuccess: async () => {
      toast('Your order was cancelled and its reserved stock was released.', { type: 'success' })
      await queryClient.invalidateQueries({ queryKey: ['orders', user?.id] })
    },
    onError: (err) => toast(friendlyError(err), { type: 'error' }),
  })

  useEffect(() => {
    if (orders.data && page > 1 && orders.data.items.length === 0 && !orders.isFetching) {
      setPage((current) => Math.max(1, current - 1))
    }
  }, [orders.data, orders.isFetching, page])

  if (orders.isPending) return <Spinner label="Loading your orders" />
  if (orders.isError) return <ErrorState error={orders.error} onRetry={() => void orders.refetch()} />

  const items = orders.data.items ?? []
  const isBusy = cancelOrder.isPending

  async function onCancel(order: Order) {
    const ok = await dialogs.confirm({
      title: 'Cancel this order?',
      message: 'This submitted order will be cancelled and the reserved stock will be released. No payment has been taken.',
      confirmLabel: 'Cancel order',
      danger: true,
    })
    if (ok) cancelOrder.mutate(order.id)
  }

  return (
    <section aria-labelledby="my-orders-heading">
      <div className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <p className="text-sm font-semibold uppercase tracking-[0.15em] text-ink/60">Your purchases</p>
          <h2 id="my-orders-heading" className="mt-1 text-3xl">My orders</h2>
          <p className="mt-2 text-ink/70">Orders you have submitted from this account.</p>
        </div>
        <span className="rounded-full border border-mist px-3 py-1 text-sm font-semibold">
          {orders.data.count} {orders.data.count === 1 ? 'order' : 'orders'}
        </span>
      </div>
      <p className="mt-4 rounded-md bg-fog px-4 py-3 text-sm text-ink/75">
        Online payment is not enabled yet. These orders are submitted but have not been charged.
      </p>

      {items.length === 0 ? (
        <div className="mt-6">
          <EmptyState title="You haven't submitted an order yet">
            <p>Your submitted orders will appear here.</p>
            <Link to="/" className="btn btn-primary mt-4">Explore the shop</Link>
          </EmptyState>
        </div>
      ) : (
        <div className="mt-6 space-y-5">
          {items.map((order) => (
            <article key={order.id} className="rounded-lg border-[1.5px] border-mist bg-paper p-5 sm:p-6">
              <div className="flex flex-wrap items-start justify-between gap-4 border-b border-mist pb-4">
                <div>
                  <p className="text-xs font-semibold uppercase tracking-[0.12em] text-ink/60">Order #{order.id.slice(0, 8).toUpperCase()}</p>
                  <time className="mt-1 block text-sm text-ink/65" dateTime={order.created_at}>
                    {new Date(order.created_at).toLocaleDateString('en-GB', { dateStyle: 'long' })}
                  </time>
                </div>
                <span className={`rounded-full px-3 py-1 text-sm font-semibold ${order.status === 'cancelled' ? 'bg-mist text-ink/70' : 'bg-sand text-ink'}`}>
                  {order.status === 'cancelled' ? 'Cancelled' : 'Submitted'}
                </span>
              </div>

              <ul className="divide-y divide-mist">
                {order.items.map((item) => {
                  const attributes = Object.values(item.attributes ?? {}).join(' / ')
                  return (
                    <li key={item.id} className="flex gap-3 py-4">
                      <div className="h-16 w-12 shrink-0 overflow-hidden rounded-md bg-fog">
                        <SafeImage src={item.image_url} alt={item.product_name} className="h-full w-full object-cover" />
                      </div>
                      <div className="min-w-0 flex-1">
                        {item.product_slug ? (
                          <Link to={`/products/${item.product_slug}`} className="font-semibold hover:underline">{item.product_name}</Link>
                        ) : (
                          <p className="font-semibold">{item.product_name}</p>
                        )}
                        {(item.variant_name || attributes) && <p className="text-sm text-ink/65">{item.variant_name || attributes}</p>}
                        <p className="mt-1 text-sm text-ink/65">Qty {item.quantity} · {formatPrice(item.unit_price)} each</p>
                      </div>
                      <p className="shrink-0 font-semibold">{formatPrice(item.line_total)}</p>
                    </li>
                  )
                })}
              </ul>

              <div className="grid gap-5 border-t border-mist pt-4 sm:grid-cols-[minmax(0,1fr)_13rem]">
                <div className="text-sm text-ink/75">
                  <p className="font-semibold text-ink">Deliver to {order.shipping_address.recipient_name}</p>
                  <p>{order.shipping_address.street}</p>
                  <p>{order.shipping_address.city}, {order.shipping_address.province} {order.shipping_address.postal_code}</p>
                  <p>{order.shipping_address.country} · {order.shipping_address.phone}</p>
                </div>
                <dl className="space-y-1 text-sm">
                  <div className="flex justify-between gap-3"><dt className="text-ink/65">Subtotal</dt><dd>{formatPrice(order.subtotal)}</dd></div>
                  <div className="flex justify-between gap-3"><dt className="text-ink/65">Shipping</dt><dd>{formatPrice(order.shipping_fee)}</dd></div>
                  <div className="flex justify-between gap-3 border-t border-mist pt-2 text-base font-bold"><dt>Total</dt><dd>{formatPrice(order.total)}</dd></div>
                </dl>
              </div>
              {order.status === 'submitted' && (
                <div className="mt-4 border-t border-mist pt-4">
                  <button type="button" className="btn btn-danger btn-sm" disabled={isBusy} onClick={() => void onCancel(order)}>
                    {cancelOrder.isPending && cancelOrder.variables === order.id ? 'Cancelling…' : 'Cancel order'}
                  </button>
                </div>
              )}
            </article>
          ))}
        </div>
      )}

      {(page > 1 || orders.data.has_more) && (
        <div className="mt-6 flex items-center justify-between gap-3">
          <button type="button" className="btn btn-outline btn-sm" disabled={page <= 1} onClick={() => setPage((current) => Math.max(1, current - 1))}>
            Newer orders
          </button>
          <span className="text-sm text-ink/65">Page {page}</span>
          <button type="button" className="btn btn-outline btn-sm" disabled={!orders.data.has_more} onClick={() => setPage((current) => current + 1)}>
            Older orders
          </button>
        </div>
      )}
    </section>
  )
}
