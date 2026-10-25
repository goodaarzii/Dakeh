import { useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { EmptyState, ErrorState, Spinner } from '../components/States'
import { SafeImage } from '../components/SafeImage'
import { useToast } from '../components/Toast'
import { api } from '../lib/api'
import { friendlyError } from '../lib/messages'
import { formatPrice } from '../lib/money'
import { useCart } from '../lib/cart'
import type { Address, ApiCart, CreateOrderResponse } from '../lib/types'

// Keep this aligned with DefaultShippingFeeCents in internal/usecase/order/service.go.
const SHIPPING_FEE_CENTS = 500
const EMPTY_API_CART: ApiCart = { items: [], subtotal: 0, item_count: 0, can_checkout: false }

export default function CheckoutPage() {
  const { cart, loading: cartLoading, error: cartError, refetch } = useCart()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const toast = useToast()
  const [selectedAddressId, setSelectedAddressId] = useState('')
  const [idempotencyKey] = useState(() => crypto.randomUUID())

  const addresses = useQuery({
    queryKey: ['addresses'],
    queryFn: () => api.get<Address[]>('/users/address'),
  })

  useEffect(() => {
    const list = addresses.data
    if (!list?.length) return
    if (!selectedAddressId || !list.some((address) => address.id === selectedAddressId)) {
      setSelectedAddressId((list.find((address) => address.is_default) ?? list[0]).id)
    }
  }, [addresses.data, selectedAddressId])

  const submitOrder = useMutation({
    mutationFn: () => api.post<CreateOrderResponse>('/orders', { address_id: selectedAddressId }, { headers: { 'Idempotency-Key': idempotencyKey } }),
    onSuccess: async (result) => {
      if (result.cart_cleared) {
        queryClient.setQueryData<ApiCart>(['cart'], EMPTY_API_CART)
      } else {
        await queryClient.invalidateQueries({ queryKey: ['cart'] })
      }
      await queryClient.invalidateQueries({ queryKey: ['orders'] })
      toast('Your order has been submitted. No payment has been taken.', { type: 'success' })
      if (!result.cart_cleared) {
        toast('Your order is safe, but the cart could not be cleared automatically. Check it before placing another order.', { type: 'error' })
      }
      navigate('/account?tab=orders')
    },
    onError: (err) => toast(friendlyError(err), { type: 'error' }),
  })

  if (cartLoading || addresses.isPending) return <Spinner label="Preparing checkout" />
  if (cartError) return <ErrorState error={cartError} onRetry={refetch} />
  if (addresses.isError) return <ErrorState error={addresses.error} onRetry={() => void addresses.refetch()} />

  if (cart.lines.length === 0) {
    return (
      <EmptyState title="Your cart is empty">
        <p>Add products before starting checkout.</p>
        <Link to="/" className="btn btn-primary mt-4">Browse products</Link>
      </EmptyState>
    )
  }

  const addressList = addresses.data ?? []
  const hasCartProblems = cart.lines.some((line) => !line.available || !line.inStock)
  const canSubmit = selectedAddressId !== '' && addressList.some((address) => address.id === selectedAddressId)
    && cart.canCheckout && !hasCartProblems && !submitOrder.isPending
  const total = cart.subtotal + SHIPPING_FEE_CENTS

  return (
    <div className="pb-8">
      <div className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <p className="text-sm font-semibold uppercase tracking-[0.15em] text-ink/60">Almost there</p>
          <h1 className="mt-1 text-4xl">Checkout</h1>
          <p className="mt-2 text-ink/70">Choose a delivery address and review your order before submitting it.</p>
        </div>
        <Link to="/cart" className="btn btn-outline btn-sm">Back to cart</Link>
      </div>

      <div className="mt-8 grid grid-cols-1 items-start gap-8 lg:grid-cols-[minmax(0,1fr)_22rem]">
        <div className="space-y-8">
          <section className="rounded-lg border-[1.5px] border-mist bg-paper p-5 sm:p-6" aria-labelledby="delivery-heading">
            <div className="flex flex-wrap items-start justify-between gap-3">
              <div>
                <p className="text-sm font-semibold uppercase tracking-[0.12em] text-ink/60">Step 1</p>
                <h2 id="delivery-heading" className="mt-1 text-2xl">Delivery address</h2>
              </div>
              <Link to="/account?tab=addresses" className="btn btn-outline btn-sm">Manage addresses</Link>
            </div>

            {addressList.length === 0 ? (
              <div className="mt-5 rounded-lg border border-dashed border-mist p-5">
                <p className="font-semibold">You haven't saved a delivery address yet.</p>
                <p className="mt-1 text-sm text-ink/70">Add an address to your account, then return here to finish checkout.</p>
                <Link to="/account?tab=addresses" className="btn btn-primary mt-4">Add an address</Link>
              </div>
            ) : (
              <fieldset className="mt-5 grid gap-3">
                <legend className="sr-only">Choose a delivery address</legend>
                {addressList.map((address) => (
                  <label
                    key={address.id}
                    className={`flex cursor-pointer gap-3 rounded-lg border-[1.5px] p-4 transition-colors ${selectedAddressId === address.id ? 'border-ink bg-fog' : 'border-mist hover:border-ink/50'}`}
                  >
                    <input
                      type="radio"
                      name="delivery-address"
                      value={address.id}
                      checked={selectedAddressId === address.id}
                      onChange={() => setSelectedAddressId(address.id)}
                      className="mt-1 accent-ink"
                    />
                    <span className="min-w-0 flex-1">
                      <span className="flex flex-wrap items-center gap-2 font-semibold">
                        {address.label || 'Delivery address'}
                        {address.is_default && <span className="rounded bg-ink px-2 py-0.5 text-xs text-paper">Default</span>}
                      </span>
                      <span className="mt-1 block">{address.recipient_name} · {address.phone}</span>
                      <span className="mt-1 block text-sm text-ink/75">
                        {address.street}, {address.city}, {address.province} {address.postal_code}, {address.country}
                      </span>
                    </span>
                  </label>
                ))}
              </fieldset>
            )}
          </section>

          <section className="rounded-lg border-[1.5px] border-mist bg-paper p-5 sm:p-6" aria-labelledby="items-heading">
            <p className="text-sm font-semibold uppercase tracking-[0.12em] text-ink/60">Step 2</p>
            <h2 id="items-heading" className="mt-1 text-2xl">Review your items</h2>
            {hasCartProblems && (
              <p className="mt-4 rounded-md border border-brick/40 bg-brick/5 p-3 text-sm text-brick">
                At least one item is unavailable or its stock changed. Return to your cart to fix these items before submitting.
              </p>
            )}
            <ul className="mt-4 divide-y divide-mist border-y border-mist">
              {cart.lines.map((line) => {
                const attributes = Object.values(line.attributes).join(' / ')
                return (
                  <li key={line.variantId} className="flex gap-4 py-4">
                    <div className="h-20 w-16 shrink-0 overflow-hidden rounded-md bg-fog">
                      <SafeImage src={line.imageUrl} alt={line.productName || 'Product'} className="h-full w-full object-cover" />
                    </div>
                    <div className="min-w-0 flex-1">
                      {line.productSlug ? (
                        <Link to={`/products/${line.productSlug}`} className="font-semibold hover:underline">{line.productName}</Link>
                      ) : (
                        <p className="font-semibold">{line.productName || 'Product unavailable'}</p>
                      )}
                      {(line.variantName || attributes) && <p className="mt-0.5 text-sm text-ink/65">{line.variantName || attributes}</p>}
                      <p className="mt-1 text-sm text-ink/70">Qty {line.quantity} · {formatPrice(line.unitPrice)} each</p>
                      {(!line.available || !line.inStock) && <p className="mt-1 text-sm font-semibold text-brick">Needs attention</p>}
                    </div>
                    <p className="shrink-0 font-semibold">{formatPrice(line.lineTotal)}</p>
                  </li>
                )
              })}
            </ul>
          </section>
        </div>

        <aside className="sticky top-6 rounded-lg border-[1.5px] border-ink bg-paper p-6">
          <h2 className="text-xl">Order summary</h2>
          <dl className="mt-5 space-y-3 text-sm">
            <div className="flex justify-between gap-4"><dt className="text-ink/70">Items ({cart.itemCount})</dt><dd className="font-semibold">{formatPrice(cart.subtotal)}</dd></div>
            <div className="flex justify-between gap-4"><dt className="text-ink/70">Shipping (flat rate)</dt><dd className="font-semibold">{formatPrice(SHIPPING_FEE_CENTS)}</dd></div>
            <div className="flex justify-between gap-4 border-t border-mist pt-4 text-base"><dt className="font-semibold">Total</dt><dd className="font-display text-2xl font-bold">{formatPrice(total)}</dd></div>
          </dl>
          <div className="mt-5 rounded-md bg-fog p-3 text-sm text-ink/75">
            <p className="font-semibold text-ink">Payment is not enabled yet.</p>
            <p className="mt-1">Submitting this order records it in your account. It does not charge you.</p>
          </div>
          <button
            type="button"
            className="btn btn-primary mt-5 w-full !py-3"
            disabled={!canSubmit}
            onClick={() => submitOrder.mutate()}
          >
            {submitOrder.isPending ? 'Submitting order…' : 'Submit order'}
          </button>
          {addressList.length > 0 && !selectedAddressId && <p className="mt-3 text-sm text-brick">Choose a delivery address to continue.</p>}
          {!cart.canCheckout && !hasCartProblems && <p className="mt-3 text-sm text-brick">Your cart cannot be checked out yet.</p>}
          <p className="mt-4 text-xs leading-relaxed text-ink/55">Stock and prices are checked again by the server when the order is submitted.</p>
        </aside>
      </div>
    </div>
  )
}
