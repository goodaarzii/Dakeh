import { useEffect, useMemo, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Link, useParams } from 'react-router-dom'
import { HeartIcon } from '../components/icons'
import { ReviewSection } from '../components/ReviewSection'
import { QuantityStepper } from '../components/QuantityStepper'
import { Rating } from '../components/Rating'
import { SafeImage } from '../components/SafeImage'
import { EmptyState, ErrorState, Spinner } from '../components/States'
import { useToast } from '../components/Toast'
import { ApiError, api } from '../lib/api'
import { useCart } from '../lib/cart'
import { useAuth } from '../lib/auth'
import { friendlyError } from '../lib/messages'
import { discountPercent, formatPrice } from '../lib/money'
import type { FavoriteList, FavoriteStatus, ProductDetail, Variant } from '../lib/types'
import { attributeKeys, firstVariant, pickByAttribute, valueInStock, valuesFor } from '../lib/variants'

export default function ProductPage() {
  const { slug = '' } = useParams()
  const query = useQuery({
    queryKey: ['product', slug],
    queryFn: () => api.get<ProductDetail>(`/products/${encodeURIComponent(slug)}`),
    enabled: slug !== '',
  })

  if (query.isPending) return <Spinner label="Loading product" />

  if (query.isError) {
    const notFound = query.error instanceof ApiError && query.error.status === 404
    if (notFound) {
      return (
        <EmptyState title="This product is not for sale">
          <p>It may have been removed, or it is not available yet.</p>
          <Link to="/" className="btn btn-outline mt-4">
            Back to the shop
          </Link>
        </EmptyState>
      )
    }
    return <ErrorState error={query.error} onRetry={() => void query.refetch()} />
  }

  // key: a different product starts with fresh choices (variant, quantity, picture)
  return <ProductView key={query.data.id} product={query.data} />
}

function ProductView({ product }: { product: ProductDetail }) {
  const { addItem } = useCart()
  const { user } = useAuth()
  const toast = useToast()
  const queryClient = useQueryClient()
  const favorites = useQuery({
    queryKey: ['favorites'],
    queryFn: () => api.get<FavoriteList>('/favorites'),
    enabled: !!user,
  })
  const isFavorite = (favorites.data?.items ?? []).some((item) => item.id === product.id)
  const favoriteMutation = useMutation({
    mutationFn: async (wasFavorite: boolean) => {
      if (wasFavorite) {
        await api.delete<void>(`/favorites/${product.id}`)
        return false
      }
      await api.put<FavoriteStatus>(`/favorites/${product.id}`)
      return true
    },
    onSuccess: async (saved) => {
      await queryClient.invalidateQueries({ queryKey: ['favorites'] })
      toast(saved ? 'Added to your favorites.' : 'Removed from your favorites.', { type: 'success', action: saved ? { label: 'View favorites', to: '/favorites' } : undefined })
    },
    onError: (err) => toast(friendlyError(err), { type: 'error' }),
  })

  const [variantId, setVariantId] = useState<string | null>(null)
  const [quantity, setQuantity] = useState(1)
  const [adding, setAdding] = useState(false)
  const [activeImage, setActiveImage] = useState(0)

  const images = product.images ?? []
  const keys = useMemo(() => attributeKeys(product.variants), [product.variants])
  const variant = product.variants.find((v) => v.id === variantId) ?? firstVariant(product.variants)
  const discount = variant ? discountPercent(variant.price, variant.compare_at_price) : 0

  // A variant can have its own picture: show it when that variant is chosen.
  useEffect(() => {
    if (!variant) return
    const index = images.findIndex((i) => i.variant_id === variant.id)
    if (index >= 0) setActiveImage(index)
  }, [variant?.id, images]) // eslint-disable-line react-hooks/exhaustive-deps

  async function onAdd() {
    if (!variant) return
    setAdding(true)
    try {
      const picture = images.find((i) => i.variant_id === variant.id) ?? images[0]
      await addItem(
        {
          variantId: variant.id,
          productName: product.name,
          productSlug: product.slug,
          variantName: variant.name,
          sku: variant.sku,
          attributes: variant.attributes ?? {},
          imageUrl: picture?.url ?? '',
          unitPrice: variant.price,
          compareAtPrice: variant.compare_at_price,
        },
        quantity,
      )
      toast(`${product.name} was added to your cart.`, { type: 'success', action: { label: 'View cart', to: '/cart' } })
    } catch (err) {
      toast(friendlyError(err), { type: 'error' })
    } finally {
      setAdding(false)
    }
  }

  return (
    <div>
      <nav className="mb-6 text-sm text-ink/70" aria-label="Breadcrumb">
        <Link to="/" className="hover:underline">
          Shop
        </Link>
        <span className="mx-2">/</span>
        <Link to={`/?category=${product.category.slug}`} className="hover:underline">
          {product.category.name}
        </Link>
        <span className="mx-2">/</span>
        <span className="text-ink">{product.name}</span>
      </nav>

      <div className="grid grid-cols-[minmax(0,1fr)_minmax(0,1fr)] gap-14">
        <div>
          <div className="aspect-[4/5] overflow-hidden rounded-lg bg-fog">
            <SafeImage
              src={images[activeImage]?.url}
              alt={images[activeImage]?.alt_text || product.name}
              className="h-full w-full object-cover"
              eager
            />
          </div>
          {images.length > 1 && (
            <div className="mt-3 flex gap-3">
              {images.map((img, index) => (
                <button
                  key={img.url}
                  type="button"
                  onClick={() => setActiveImage(index)}
                  aria-label={`Show picture ${index + 1}`}
                  aria-current={index === activeImage ? 'true' : undefined}
                  className={`h-20 w-16 overflow-hidden rounded-md border-2 bg-fog ${
                    index === activeImage ? 'border-ink' : 'border-transparent hover:border-mist'
                  }`}
                >
                  <SafeImage src={img.url} alt="" className="h-full w-full object-cover" />
                </button>
              ))}
            </div>
          )}
        </div>

        <div>
          {product.brand && <p className="text-ink/70">{product.brand}</p>}
          <h1 className="mt-1 text-4xl">{product.name}</h1>
          <div className="mt-2">
            <Rating value={product.avg_rating} count={product.rating_count} />
          </div>

          {variant ? (
            <>
              <div className="mt-6 flex flex-wrap items-baseline gap-3">
                <span className="font-display text-4xl font-bold">{formatPrice(variant.price)}</span>
                {discount > 0 && (
                  <>
                    <span className="text-lg text-ink/60 line-through">{formatPrice(variant.compare_at_price)}</span>
                    <span className="rounded bg-moss px-2 py-0.5 text-sm font-semibold text-paper">-{discount}%</span>
                  </>
                )}
              </div>
              <p className="mt-2 inline-flex items-center gap-2 text-sm">
                <span className={`h-2 w-2 rounded-full ${variant.in_stock ? 'bg-moss' : 'bg-brick'}`} />
                {variant.in_stock ? 'In stock' : 'Sold out'}
              </p>

              <div className="mt-6 space-y-5">
                <VariantPicker variants={product.variants} keys={keys} current={variant} onPick={(v) => setVariantId(v.id)} />
              </div>

              <div className="mt-8 flex items-center gap-3">
                <QuantityStepper value={quantity} onChange={setQuantity} disabled={!variant.in_stock} />
                <button type="button" className="btn btn-primary flex-1 !py-3" disabled={!variant.in_stock || adding} onClick={() => void onAdd()}>
                  {adding ? 'Adding…' : variant.in_stock ? 'Add to cart' : 'Sold out'}
                </button>
                <button
                  type="button"
                  className={`btn btn-outline !px-3 !py-3 ${isFavorite ? '!text-brick' : ''}`}
                  aria-label={isFavorite ? 'Remove from favorites' : 'Add to favorites'}
                  aria-pressed={isFavorite}
                  disabled={favoriteMutation.isPending || (!!user && favorites.isPending)}
                  onClick={() => {
                    if (!user) {
                      toast('Log in to save products to your favorites.', {
                        action: { label: 'Log in', to: `/login?next=${encodeURIComponent(`/products/${product.slug}`)}` },
                      })
                      return
                    }
                    favoriteMutation.mutate(isFavorite)
                  }}
                >
                  <HeartIcon filled={isFavorite} />
                </button>
              </div>
            </>
          ) : (
            <p className="mt-6 text-ink/75">This product has no variants to buy yet.</p>
          )}

          {product.tags.length > 0 && (
            <div className="mt-8 flex flex-wrap gap-2">
              {product.tags.map((t) => (
                <Link key={t.id} to={`/?tag=${t.slug}`} className="chip no-underline">
                  {t.name}
                </Link>
              ))}
            </div>
          )}

          {product.description && (
            <div className="mt-8 border-t-[1.5px] border-mist pt-6">
              <h2 className="text-xl">About this product</h2>
              <p className="mt-3 whitespace-pre-line text-ink/85">{product.description}</p>
            </div>
          )}

          <ReviewSection product={product} />
        </div>
      </div>
    </div>
  )
}

function VariantPicker({
  variants,
  keys,
  current,
  onPick,
}: {
  variants: Variant[]
  keys: string[]
  current: Variant
  onPick: (variant: Variant) => void
}) {
  if (variants.length <= 1) return null

  // Variants without attributes: one button per variant name.
  if (keys.length === 0) {
    return (
      <fieldset>
        <legend className="mb-2 text-sm font-semibold">Option</legend>
        <div className="flex flex-wrap gap-2">
          {variants.map((v) => (
            <button
              key={v.id}
              type="button"
              className="chip"
              aria-pressed={v.id === current.id}
              onClick={() => onPick(v)}
              style={!v.in_stock && v.id !== current.id ? { opacity: 0.5, textDecoration: 'line-through' } : undefined}
            >
              {v.name || v.sku}
            </button>
          ))}
        </div>
      </fieldset>
    )
  }

  return (
    <>
      {keys.map((key) => (
        <fieldset key={key}>
          <legend className="mb-2 text-sm font-semibold capitalize">
            {key}: <span className="font-normal">{current.attributes?.[key]}</span>
          </legend>
          <div className="flex flex-wrap gap-2">
            {valuesFor(variants, key).map((value) => {
              const selected = current.attributes?.[key] === value
              const available = valueInStock(variants, current, key, value)
              return (
                <button
                  key={value}
                  type="button"
                  className="chip"
                  aria-pressed={selected}
                  onClick={() => onPick(pickByAttribute(variants, current, key, value))}
                  style={!available && !selected ? { opacity: 0.5, textDecoration: 'line-through' } : undefined}
                >
                  {value}
                </button>
              )
            })}
          </div>
        </fieldset>
      ))}
    </>
  )
}
