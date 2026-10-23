import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { ProductCard } from '../components/ProductCard'
import { EmptyState, ErrorState, Spinner } from '../components/States'
import { useToast } from '../components/Toast'
import { api } from '../lib/api'
import { friendlyError } from '../lib/messages'
import type { FavoriteList } from '../lib/types'

export default function FavoritesPage() {
  const queryClient = useQueryClient()
  const toast = useToast()
  const favorites = useQuery({
    queryKey: ['favorites'],
    queryFn: () => api.get<FavoriteList>('/favorites'),
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
    <section>
      <div className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <p className="text-sm font-semibold uppercase tracking-[0.15em] text-ink/60">Your shortlist</p>
          <h1 className="mt-1 text-4xl">Favorites</h1>
          <p className="mt-2 text-ink/70">Products you want to come back to.</p>
        </div>
        <span className="rounded-full border border-mist px-3 py-1 text-sm font-semibold">
          {favorites.data.count} {favorites.data.count === 1 ? 'product' : 'products'}
        </span>
      </div>

      {items.length === 0 ? (
        <EmptyState title="No favorites yet">
          <p>Tap the heart on a product to keep it here for later.</p>
          <Link to="/" className="btn btn-primary mt-4">Explore the shop</Link>
        </EmptyState>
      ) : (
        <div className="mt-8 grid grid-cols-1 gap-x-6 gap-y-10 sm:grid-cols-2 lg:grid-cols-3">
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
