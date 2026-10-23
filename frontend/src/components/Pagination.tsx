import { ChevronIcon } from './icons'

// The API says only whether another page exists (has_more), not how many pages there are.
export function Pagination({ page, hasMore, onPage }: { page: number; hasMore: boolean; onPage: (page: number) => void }) {
  if (page === 1 && !hasMore) return null
  return (
    <nav className="mt-12 flex items-center justify-center gap-4" aria-label="Pages">
      <button type="button" className="btn btn-outline btn-sm" disabled={page <= 1} onClick={() => onPage(page - 1)}>
        <ChevronIcon direction="left" className="h-4 w-4" />
        Previous
      </button>
      <span className="font-display text-lg font-semibold">Page {page}</span>
      <button type="button" className="btn btn-outline btn-sm" disabled={!hasMore} onClick={() => onPage(page + 1)}>
        Next
        <ChevronIcon direction="right" className="h-4 w-4" />
      </button>
    </nav>
  )
}
