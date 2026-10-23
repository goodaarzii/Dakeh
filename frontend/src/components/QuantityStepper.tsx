import { MAX_QTY_PER_LINE } from '../config'
import { MinusIcon, PlusIcon } from './icons'

export function QuantityStepper({
  value,
  onChange,
  disabled,
  min = 1,
  max = MAX_QTY_PER_LINE,
}: {
  value: number
  onChange: (value: number) => void
  disabled?: boolean
  min?: number
  max?: number
}) {
  return (
    <div className="inline-flex items-center rounded-lg border-[1.5px] border-mist bg-paper">
      <button
        type="button"
        className="btn btn-quiet btn-sm !rounded-r-none"
        onClick={() => onChange(value - 1)}
        disabled={disabled || value <= min}
        aria-label="Decrease quantity"
      >
        <MinusIcon className="h-4 w-4" />
      </button>
      <span className="min-w-8 text-center font-semibold tabular-nums" aria-live="polite">
        {value}
      </span>
      <button
        type="button"
        className="btn btn-quiet btn-sm !rounded-l-none"
        onClick={() => onChange(value + 1)}
        disabled={disabled || value >= max}
        aria-label="Increase quantity"
      >
        <PlusIcon className="h-4 w-4" />
      </button>
    </div>
  )
}
