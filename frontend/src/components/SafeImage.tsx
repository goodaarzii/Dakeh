import { useEffect, useState } from 'react'
import { PLACEHOLDER_IMAGE } from '../config'

// Shows the placeholder when there is no image, or when the image cannot be loaded.
export function SafeImage({ src, alt, className, eager }: { src?: string; alt: string; className?: string; eager?: boolean }) {
  const [failed, setFailed] = useState(false)

  useEffect(() => {
    setFailed(false)
  }, [src])

  return (
    <img
      src={failed || !src ? PLACEHOLDER_IMAGE : src}
      alt={alt}
      className={className}
      loading={eager ? 'eager' : 'lazy'}
      onError={() => setFailed(true)}
    />
  )
}
