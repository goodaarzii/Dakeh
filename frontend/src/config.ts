// Where the Go API lives. Set VITE_API_URL in frontend/.env to change it.
export const API_URL: string = (
  (import.meta.env.VITE_API_URL as string | undefined) ?? 'http://localhost:8080/api/v1'
).replace(/\/$/, '')

export const SHOP_NAME = 'Dakeh'

// Static images live in frontend/public/images. Put your own files there with these exact names.
export const PLACEHOLDER_IMAGE = '/images/placeholder-product.png'
export const LOGO_IMAGE = '/images/logo.svg'

// Keep in step with domain.MaxQuantityPerLine and domain.MaxCartLines in the backend.
export const MAX_QTY_PER_LINE = 20
export const MAX_CART_LINES = 50
