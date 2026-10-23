import { StarIcon } from './icons'

export function Rating({ value, count }: { value: number; count: number }) {
  if (count <= 0) return null
  return (
    <span className="inline-flex items-center gap-1 text-sm text-ink/80">
      <StarIcon className="h-4 w-4 text-steel" />
      <span className="font-semibold">{value.toFixed(1)}</span>
      <span>({count})</span>
    </span>
  )
}
