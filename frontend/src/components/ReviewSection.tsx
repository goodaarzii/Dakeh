import { useEffect, useState } from 'react'
import { useInfiniteQuery, useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { ApiError, api } from '../lib/api'
import { useAuth } from '../lib/auth'
import { friendlyError } from '../lib/messages'
import { useToast } from './Toast'
import type { ProductDetail, Review, ReviewList } from '../lib/types'

const PAGE_SIZE = 10

export function ReviewSection({ product }: { product: ProductDetail }) {
  const { user } = useAuth()
  const toast = useToast()
  const queryClient = useQueryClient()
  const [rating, setRating] = useState(5)
  const [title, setTitle] = useState('')
  const [body, setBody] = useState('')

  const reviews = useInfiniteQuery({
    queryKey: ['reviews', product.id],
    queryFn: ({ pageParam }) => api.get<ReviewList>(`/products/${product.id}/reviews`, {
      page: pageParam,
      page_size: PAGE_SIZE,
    }),
    initialPageParam: 1,
    getNextPageParam: (lastPage) => lastPage.has_more ? lastPage.page + 1 : undefined,
  })

  const myReview = useQuery({
    queryKey: ['my-review', product.id],
    enabled: !!user,
    queryFn: async () => {
      try {
        return await api.get<Review>('/reviews/mine', { product_id: product.id })
      } catch (err) {
        if (err instanceof ApiError && err.status === 404) return null
        throw err
      }
    },
  })

  useEffect(() => {
    if (!myReview.data) {
      setRating(5)
      setTitle('')
      setBody('')
      return
    }
    setRating(myReview.data.rating)
    setTitle(myReview.data.title)
    setBody(myReview.data.body)
  }, [myReview.data?.id, myReview.data?.updated_at])

  const refreshReviewData = async () => {
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: ['reviews', product.id] }),
      queryClient.invalidateQueries({ queryKey: ['my-review', product.id] }),
      queryClient.invalidateQueries({ queryKey: ['product', product.slug] }),
    ])
  }

  const saveReview = useMutation({
    mutationFn: () => {
      const input = { rating, title: title.trim(), body: body.trim() }
      return myReview.data?.id
        ? api.patch<Review>(`/reviews/${myReview.data.id}`, input)
        : api.post<Review>(`/products/${product.id}/reviews`, input)
    },
    onSuccess: async () => {
      toast('Your review has been saved.', { type: 'success' })
      await refreshReviewData()
    },
    onError: (err) => toast(friendlyError(err), { type: 'error' }),
  })

  const deleteReview = useMutation({
    mutationFn: () => api.delete<void>(`/reviews/${myReview.data?.id}`),
    onSuccess: async () => {
      toast('Your review has been deleted.', { type: 'success' })
      setRating(5)
      setTitle('')
      setBody('')
      await refreshReviewData()
    },
    onError: (err) => toast(friendlyError(err), { type: 'error' }),
  })

  const allReviews = reviews.data?.pages.flatMap((page) => page.items) ?? []
  const reviewCount = reviews.data?.pages[0]?.count ?? product.rating_count
  const busy = saveReview.isPending || deleteReview.isPending

  return (
    <section className="mt-8 border-t-[1.5px] border-mist pt-6" aria-labelledby="reviews-heading">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <h2 id="reviews-heading" className="text-xl">Reviews</h2>
          <p className="mt-1 text-sm text-ink/70">
            {reviewCount > 0
              ? `${product.avg_rating.toFixed(1)} out of 5 · ${reviewCount} ${reviewCount === 1 ? 'review' : 'reviews'}`
              : 'No reviews yet. Be the first to share your experience.'}
          </p>
        </div>
      </div>

      <div className="mt-5 rounded-lg border border-mist p-4">
        <h3 className="font-semibold">{myReview.data ? 'Edit your review' : 'Write a review'}</h3>
        {!user ? (
          <p className="mt-2 text-sm text-ink/75">
            <Link className="underline underline-offset-2" to={`/login?next=${encodeURIComponent(`/products/${product.slug}`)}`}>Log in</Link> to rate this product and share your experience.
          </p>
        ) : (
          <form className="mt-3 space-y-3" onSubmit={(event) => { event.preventDefault(); saveReview.mutate() }}>
            <fieldset>
              <legend className="mb-2 text-sm font-medium">Your rating</legend>
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
                    className={`text-2xl leading-none ${value <= rating ? 'text-steel' : 'text-mist'}`}
                  >★</button>
                ))}
              </div>
            </fieldset>
            <label className="block text-sm">
              <span className="mb-1 block font-medium">Title (optional)</span>
              <input
                className="w-full rounded-md border border-mist bg-paper px-3 py-2"
                value={title}
                onChange={(event) => setTitle(event.target.value)}
                maxLength={120}
                disabled={busy}
                placeholder="Sum up your experience"
              />
            </label>
            <label className="block text-sm">
              <span className="mb-1 block font-medium">Your review</span>
              <textarea
                className="w-full rounded-md border border-mist bg-paper px-3 py-2"
                value={body}
                onChange={(event) => setBody(event.target.value)}
                minLength={3}
                maxLength={5000}
                rows={4}
                required
                disabled={busy}
                placeholder="What should other shoppers know?"
              />
              <span className="mt-1 block text-xs text-ink/60">3–5,000 characters</span>
            </label>
            <div className="flex flex-wrap gap-2">
              <button type="submit" className="btn btn-primary" disabled={busy || body.trim().length < 3 || (!!user && myReview.isPending)}>
                {saveReview.isPending ? 'Saving…' : myReview.data ? 'Update review' : 'Publish review'}
              </button>
              {myReview.data && (
                <button type="button" className="btn btn-outline" disabled={busy} onClick={() => deleteReview.mutate()}>
                  {deleteReview.isPending ? 'Deleting…' : 'Delete review'}
                </button>
              )}
            </div>
            {myReview.isError && <p className="text-sm text-brick">Could not load your existing review. {friendlyError(myReview.error)}</p>}
          </form>
        )}
      </div>

      <div className="mt-6 space-y-4">
        {reviews.isPending && <p className="text-sm text-ink/70">Loading reviews…</p>}
        {reviews.isError && <p role="alert" className="text-sm text-brick">Could not load reviews. {friendlyError(reviews.error)}</p>}
        {!reviews.isPending && !reviews.isError && allReviews.length === 0 && (
          <p className="text-sm text-ink/70">There are no written reviews for this product yet.</p>
        )}
        {allReviews.map((review) => (
          <article key={review.id} className="border-b border-mist pb-4">
            <div className="flex flex-wrap items-center justify-between gap-2">
              <div className="font-semibold">{review.reviewer_name || 'Shopper'}</div>
              <time className="text-xs text-ink/60" dateTime={review.created_at}>
                {new Date(review.created_at).toLocaleDateString()}
              </time>
            </div>
            <div className="mt-1 text-sm" aria-label={`${review.rating} out of 5 stars`}>
              <span className="text-steel">{'★'.repeat(review.rating)}</span><span className="text-mist">{'★'.repeat(5 - review.rating)}</span>
              <span className="ml-2 text-ink/70">{review.rating}/5</span>
            </div>
            {review.title && <h4 className="mt-2 font-semibold">{review.title}</h4>}
            <p className="mt-1 whitespace-pre-line text-sm text-ink/85">{review.body}</p>
            {review.updated_at !== review.created_at && <p className="mt-2 text-xs text-ink/55">Edited</p>}
          </article>
        ))}
        {reviews.hasNextPage && (
          <button type="button" className="btn btn-outline btn-sm" disabled={reviews.isFetchingNextPage} onClick={() => void reviews.fetchNextPage()}>
            {reviews.isFetchingNextPage ? 'Loading…' : 'Load older reviews'}
          </button>
        )}
      </div>
    </section>
  )
}
