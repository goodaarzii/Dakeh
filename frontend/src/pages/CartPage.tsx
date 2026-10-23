import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useDialogs } from '../components/Dialogs'
import { TrashIcon } from '../components/icons'
import { QuantityStepper } from '../components/QuantityStepper'
import { SafeImage } from '../components/SafeImage'
import { EmptyState, ErrorState, Spinner } from '../components/States'
import { useToast } from '../components/Toast'
import { useCart, type CartLineView } from '../lib/cart'
import { friendlyError } from '../lib/messages'
import { formatPrice } from '../lib/money'

export default function CartPage() {
  const { cart, isGuest, loading, error, setQuantity, removeItem, clear, refetch } = useCart()
  const navigate = useNavigate()
  const toast = useToast()
  const dialogs = useDialogs()
  const [busyId, setBusyId] = useState<string | null>(null)

  async function run(id: string, action: () => Promise<void>) {
    setBusyId(id)
    try {
      await action()
    } catch (err) {
      toast(friendlyError(err), { type: 'error' })
    } finally {
      setBusyId(null)
    }
  }

  async function onClear() {
    const ok = await dialogs.confirm({
      title: 'Empty your cart?',
      message: 'Every item will be removed from the cart.',
      confirmLabel: 'Empty cart',
      danger: true,
    })
    if (ok) await run('all', clear)
  }

  function onCheckout() {
    if (isGuest) {
      navigate('/login?next=/cart')
      return
    }
    dialogs.underConstruction('Checkout and payment')
  }

  if (loading) return <Spinner label="Loading your cart" />
  if (error) return <ErrorState error={error} onRetry={refetch} />

  if (cart.lines.length === 0) {
    return (
      <EmptyState title="Your cart is empty">
        <p>Add something from the shop and it will show up here.</p>
        <Link to="/" className="btn btn-primary mt-4">
          Browse products
        </Link>
      </EmptyState>
    )
  }

  const hasProblems = cart.lines.some((l) => !l.available || !l.inStock)

  return (
    <div>
      <div className="flex items-end justify-between gap-4">
        <h1 className="text-4xl">Your cart</h1>
        <button type="button" className="btn btn-quiet btn-sm" onClick={() => void onClear()} disabled={busyId !== null}>
          Empty cart
        </button>
      </div>

      {isGuest && (
        <p className="mt-4 rounded-lg border-[1.5px] border-mist bg-paper px-4 py-3 text-sm">
          You are shopping as a guest. This cart is saved in this browser.{' '}
          <Link to="/login?next=/cart" className="font-semibold underline underline-offset-2">
            Log in
          </Link>{' '}
          to keep it on your account. Prices and stock are confirmed then.
        </p>
      )}

      <div className="mt-8 grid grid-cols-[minmax(0,1fr)_20rem] items-start gap-12">
        <ul className="divide-y-[1.5px] divide-mist border-y-[1.5px] border-mist">
          {cart.lines.map((line) => (
            <CartLine
              key={line.variantId}
              line={line}
              busy={busyId === line.variantId || busyId === 'all'}
              onQuantity={(q) => void run(line.variantId, () => setQuantity(line.variantId, q))}
              onRemove={() => void run(line.variantId, () => removeItem(line.variantId))}
            />
          ))}
        </ul>

        <aside className="sticky top-6 rounded-lg border-[1.5px] border-ink bg-paper p-6">
          <h2 className="text-xl">Summary</h2>
          <dl className="mt-4 space-y-2">
            <div className="flex justify-between">
              <dt className="text-ink/75">Items</dt>
              <dd>{cart.itemCount}</dd>
            </div>
            <div className="flex items-baseline justify-between">
              <dt className="text-ink/75">Subtotal</dt>
              <dd className="font-display text-2xl font-bold">{formatPrice(cart.subtotal)}</dd>
            </div>
          </dl>
          <p className="mt-2 text-sm text-ink/65">Shipping is added at checkout.</p>

          <button type="button" className="btn btn-primary mt-5 w-full !py-3" onClick={onCheckout} disabled={!cart.canCheckout}>
            {isGuest ? 'Log in to check out' : 'Check out'}
          </button>
          {hasProblems && (
            <p className="mt-3 text-sm text-brick">Fix the items marked in the list to continue.</p>
          )}
        </aside>
      </div>
    </div>
  )
}

function CartLine({
  line,
  busy,
  onQuantity,
  onRemove,
}: {
  line: CartLineView
  busy: boolean
  onQuantity: (quantity: number) => void
  onRemove: () => void
}) {
  const attributes = Object.values(line.attributes).join(' / ')

  if (!line.available) {
    return (
      <li className="flex items-center justify-between gap-4 py-5 opacity-80">
        <div>
          <p className="font-display text-lg">This item is no longer available</p>
          <p className="text-sm text-ink/70">The shop stopped selling it. Remove it to continue.</p>
        </div>
        <button type="button" className="btn btn-danger btn-sm" onClick={onRemove} disabled={busy}>
          Remove
        </button>
      </li>
    )
  }

  return (
    <li className="grid grid-cols-[6rem_minmax(0,1fr)_auto] gap-5 py-5">
      <Link to={`/products/${line.productSlug}`} className="block aspect-[4/5] overflow-hidden rounded-md bg-fog">
        <SafeImage src={line.imageUrl} alt={line.productName} className="h-full w-full object-cover" />
      </Link>

      <div className="min-w-0">
        <Link to={`/products/${line.productSlug}`} className="font-display text-xl leading-tight hover:underline">
          {line.productName}
        </Link>
        {(line.variantName || attributes) && <p className="mt-0.5 text-sm text-ink/70">{line.variantName || attributes}</p>}
        <p className="mt-2">
          {formatPrice(line.unitPrice)}
          {line.compareAtPrice > line.unitPrice && (
            <span className="ml-2 text-sm text-ink/55 line-through">{formatPrice(line.compareAtPrice)}</span>
          )}
        </p>
        {!line.inStock && (
          <p className="mt-2 text-sm font-semibold text-brick">Not enough in stock for {line.quantity}. Lower the quantity or remove it.</p>
        )}
        <div className="mt-3 flex items-center gap-3">
          <QuantityStepper value={line.quantity} onChange={onQuantity} disabled={busy} />
          <button type="button" className="btn btn-quiet btn-sm" onClick={onRemove} disabled={busy}>
            <TrashIcon className="h-4 w-4" />
            Remove
          </button>
        </div>
      </div>

      <p className="font-display text-xl font-semibold">{formatPrice(line.lineTotal)}</p>
    </li>
  )
}
