import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useSearchParams } from 'react-router-dom'
import { CategoryTree, TagList, findCategory } from '../components/Filters'
import { CloseIcon } from '../components/icons'
import { Pagination } from '../components/Pagination'
import { ProductCard, ProductCardSkeleton } from '../components/ProductCard'
import { EmptyState, ErrorState } from '../components/States'
import { TestNotice } from '../components/TestNotice'
import { api } from '../lib/api'
import type { CategoryNode, ProductList, Tag } from '../lib/types'

const PAGE_SIZE = 12

const SORTS = [
  { value: 'newest', label: 'Newest' },
  { value: 'price_asc', label: 'Price: low to high' },
  { value: 'price_desc', label: 'Price: high to low' },
]

// The address bar is the single source of truth for the filters, so links can be shared and the back button works.
export default function HomePage() {
  const [params, setParams] = useSearchParams()
  const category = params.get('category') ?? ''
  const tag = params.get('tag') ?? ''
  const q = params.get('q') ?? ''
  const sort = params.get('sort') ?? 'newest'
  const page = Math.max(1, parseInt(params.get('page') ?? '1', 10) || 1)

  function update(changes: Record<string, string>, keepPage = false) {
    const next = new URLSearchParams(params)
    for (const [key, value] of Object.entries(changes)) {
      if (value) next.set(key, value)
      else next.delete(key)
    }
    if (!keepPage) next.delete('page')
    setParams(next)
  }

  const categories = useQuery({
    queryKey: ['categories'],
    queryFn: () => api.get<CategoryNode[]>('/categories'),
    staleTime: 5 * 60_000,
  })
  const tags = useQuery({
    queryKey: ['tags'],
    queryFn: () => api.get<Tag[]>('/tags'),
    staleTime: 5 * 60_000,
  })
  const products = useQuery({
    queryKey: ['products', { category, tag, q, sort, page }],
    queryFn: () =>
      api.get<ProductList>('/products', {
        category,
        tag,
        q,
        sort: sort === 'newest' ? undefined : sort,
        page: page > 1 ? page : undefined,
        page_size: PAGE_SIZE,
      }),
    placeholderData: keepPreviousData,
  })

  const categoryName = category ? findCategory(categories.data ?? [], category)?.name ?? category : ''
  const tagName = tag ? (tags.data ?? []).find((t) => t.slug === tag)?.name ?? tag : ''
  const heading = q ? `Results for “${q}”` : categoryName || 'All products'
  const hasFilters = !!(category || tag || q)
  const items = products.data?.items ?? []

  return (
    <>
      <TestNotice />

      <div className="grid grid-cols-[15rem_minmax(0,1fr)] gap-12">
        <aside className="sticky top-6 self-start">
          <h2 className="text-lg">Categories</h2>
          <div className="mt-3">
            <button
              type="button"
              onClick={() => update({ category: '' })}
              className={`w-full rounded-md px-2 py-1.5 text-left transition-colors hover:bg-mist/30 ${
                !category ? 'bg-ink font-semibold text-paper hover:bg-ink' : ''
              }`}
            >
              All products
            </button>
            {categories.isPending && <div className="mt-2 h-24 animate-pulse rounded-md bg-fog" />}
            {categories.data && (
              <CategoryTree
                nodes={categories.data}
                selected={category}
                onSelect={(slug) => update({ category: slug === category ? '' : slug })}
              />
            )}
          </div>

          {tags.data && tags.data.length > 0 && (
            <div className="mt-8">
              <h2 className="text-lg">Tags</h2>
              <div className="mt-3">
                <TagList tags={tags.data} selected={tag} onSelect={(slug) => update({ tag: slug === tag ? '' : slug })} />
              </div>
            </div>
          )}
        </aside>

        <section>
          <div className="flex flex-wrap items-end justify-between gap-4">
            <div>
              <h1 className="text-4xl">{heading}</h1>
              {hasFilters && (
                <div className="mt-3 flex flex-wrap items-center gap-2">
                  {q && <FilterChip label={`Search: ${q}`} onRemove={() => update({ q: '' })} />}
                  {category && <FilterChip label={`Category: ${categoryName}`} onRemove={() => update({ category: '' })} />}
                  {tag && <FilterChip label={`Tag: ${tagName}`} onRemove={() => update({ tag: '' })} />}
                  <button type="button" className="btn btn-quiet btn-sm" onClick={() => setParams(new URLSearchParams())}>
                    Clear all
                  </button>
                </div>
              )}
            </div>

            <label className="flex items-center gap-2 text-sm font-semibold">
              Sort by
              <select className="field !w-auto !py-1.5" value={sort} onChange={(e) => update({ sort: e.target.value === 'newest' ? '' : e.target.value })}>
                {SORTS.map((s) => (
                  <option key={s.value} value={s.value}>
                    {s.label}
                  </option>
                ))}
              </select>
            </label>
          </div>

          <div className="mt-8">
            {products.isPending ? (
              <div className="grid grid-cols-3 gap-x-6 gap-y-10">
                {Array.from({ length: 6 }).map((_, i) => (
                  <ProductCardSkeleton key={i} />
                ))}
              </div>
            ) : products.isError ? (
              <ErrorState error={products.error} onRetry={() => void products.refetch()} />
            ) : items.length === 0 ? (
              <EmptyState title={hasFilters ? 'No products match these filters' : 'No products yet'}>
                {hasFilters ? (
                  <button type="button" className="btn btn-outline mt-2" onClick={() => setParams(new URLSearchParams())}>
                    Clear all filters
                  </button>
                ) : (
                  <p>The shop is empty. Products appear here once they are active and have at least one variant.</p>
                )}
              </EmptyState>
            ) : (
              <div className={`grid grid-cols-3 gap-x-6 gap-y-10 transition-opacity ${products.isFetching ? 'opacity-60' : ''}`}>
                {items.map((p) => (
                  <ProductCard key={p.id} product={p} />
                ))}
              </div>
            )}
          </div>

          {products.data && (
            <Pagination
              page={page}
              hasMore={products.data.has_more}
              onPage={(next) => {
                update({ page: next > 1 ? String(next) : '' }, true)
                window.scrollTo({ top: 0, behavior: 'smooth' })
              }}
            />
          )}
        </section>
      </div>
    </>
  )
}

function FilterChip({ label, onRemove }: { label: string; onRemove: () => void }) {
  return (
    <span className="inline-flex items-center gap-1 rounded-full bg-ink py-1 pl-3 pr-1.5 text-sm font-medium text-paper">
      {label}
      <button type="button" onClick={onRemove} className="rounded-full p-0.5 hover:bg-paper/20" aria-label={`Remove filter ${label}`}>
        <CloseIcon className="h-4 w-4" />
      </button>
    </span>
  )
}
