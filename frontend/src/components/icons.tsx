import type { ReactNode } from 'react'

interface IconProps {
  className?: string
}

function base(className: string | undefined, children: ReactNode) {
  return (
    <svg
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth={1.8}
      strokeLinecap="round"
      strokeLinejoin="round"
      className={className ?? 'h-5 w-5'}
      aria-hidden="true"
    >
      {children}
    </svg>
  )
}

export const SearchIcon = ({ className }: IconProps) =>
  base(className, <><circle cx="11" cy="11" r="7" /><path d="m20 20-3.5-3.5" /></>)

export const CartIcon = ({ className }: IconProps) =>
  base(className, <><path d="M3 4h2.2l2.1 10.2a1.5 1.5 0 0 0 1.5 1.2h8.1a1.5 1.5 0 0 0 1.5-1.1L20 8H6.2" /><circle cx="9.5" cy="19.5" r="1.2" /><circle cx="17" cy="19.5" r="1.2" /></>)

export const HeartIcon = ({ className, filled = false }: IconProps & { filled?: boolean }) => (
  <svg
    viewBox="0 0 24 24"
    fill={filled ? 'currentColor' : 'none'}
    stroke="currentColor"
    strokeWidth={1.8}
    strokeLinecap="round"
    strokeLinejoin="round"
    className={className ?? 'h-5 w-5'}
    aria-hidden="true"
  >
    <path d="M12 20s-7-4.4-7-10a4 4 0 0 1 7-2.6A4 4 0 0 1 19 10c0 5.6-7 10-7 10Z" />
  </svg>
)

export const UserIcon = ({ className }: IconProps) =>
  base(className, <><circle cx="12" cy="8" r="3.6" /><path d="M4.5 20a7.5 7.5 0 0 1 15 0" /></>)

export const CloseIcon = ({ className }: IconProps) => base(className, <path d="M6 6l12 12M18 6 6 18" />)

export const PlusIcon = ({ className }: IconProps) => base(className, <path d="M12 5v14M5 12h14" />)

export const MinusIcon = ({ className }: IconProps) => base(className, <path d="M5 12h14" />)

export const TrashIcon = ({ className }: IconProps) =>
  base(className, <><path d="M4 7h16M9 7V4.5h6V7M6.5 7l.8 12.5h9.4L17.5 7" /><path d="M10 11v5M14 11v5" /></>)

export const PinIcon = ({ className }: IconProps) =>
  base(className, <><path d="M12 21s-6.5-5.6-6.5-11a6.5 6.5 0 0 1 13 0c0 5.4-6.5 11-6.5 11Z" /><circle cx="12" cy="10" r="2.3" /></>)

export const ChevronIcon = ({ className, direction = 'right' }: IconProps & { direction?: 'left' | 'right' | 'down' }) => {
  const d = direction === 'left' ? 'm14.5 6-6 6 6 6' : direction === 'down' ? 'm6 9.5 6 6 6-6' : 'm9.5 6 6 6-6 6'
  return base(className, <path d={d} />)
}

export const TerminalIcon = ({ className }: IconProps) =>
  base(className, <><rect x="3" y="4.5" width="18" height="15" rx="2.2" /><path d="m7.5 10 3 2.5-3 2.5M13 15h4" /></>)

export const CheckIcon = ({ className }: IconProps) => base(className, <path d="m5 12.5 4.5 4.5L19 7.5" />)

export const StarIcon = ({ className }: IconProps) => (
  <svg viewBox="0 0 24 24" fill="currentColor" className={className ?? 'h-4 w-4'} aria-hidden="true">
    <path d="m12 3.2 2.6 5.5 6 .8-4.4 4.2 1.1 6-5.3-2.9-5.3 2.9 1.1-6L3.4 9.5l6-.8Z" />
  </svg>
)
