import { Link } from 'react-router-dom'
import { formatPrice } from '../lib/money'
import type { ProductCardData } from '../lib/types'
import { Rating } from './Rating'
import { SafeImage } from './SafeImage'

export function ProductCard({ product }: { product: ProductCardData }) {
  return (
    <Link to={`/products/${product.slug}`} className="group block">
      <div className="relative aspect-[4/5] overflow-hidden rounded-lg bg-fog">
        <SafeImage
          src={product.image_url}
          alt={product.name}
          className="h-full w-full object-cover transition-transform duration-500 group-hover:scale-[1.03]"
        />
        {!product.in_stock && (
          <span className="absolute left-3 top-3 rounded bg-ink px-2 py-0.5 text-xs font-semibold text-paper">Sold out</span>
        )}
      </div>
      <div className="mt-3">
        {product.brand && <p className="text-sm text-ink/70">{product.brand}</p>}
        <h3 className="text-xl leading-tight group-hover:underline group-hover:underline-offset-4">{product.name}</h3>
        <div className="mt-1 flex items-center justify-between gap-3">
          <p className="font-display text-lg font-semibold">From {formatPrice(product.min_price)}</p>
          <Rating value={product.avg_rating} count={product.rating_count} />
        </div>
      </div>
    </Link>
  )
}

export function ProductCardSkeleton() {
  return (
    <div className="animate-pulse" aria-hidden="true">
      <div className="aspect-[4/5] rounded-lg bg-fog" />
      <div className="mt-3 h-4 w-2/3 rounded bg-fog" />
      <div className="mt-2 h-4 w-1/3 rounded bg-fog" />
    </div>
  )
}
