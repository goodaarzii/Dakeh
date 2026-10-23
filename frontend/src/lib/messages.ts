import { ApiError } from './api'

// The API sends a short machine-readable code with every error. Turn the common ones into friendly sentences.
const MESSAGES: Record<string, string> = {
  email_taken: 'That email address is already registered.',
  phone_taken: 'That phone number is already registered.',
  invalid_credentials: 'The email or password is wrong.',
  user_inactive: 'This account is disabled.',
  insufficient_stock: 'There is not enough stock for that quantity.',
  variant_unavailable: 'This item is not available any more.',
  cart_full: 'Your cart is full. Remove something first.',
  conflict: 'That conflicts with something you already saved. For example, you may already have a default address.',
  not_found: 'We could not find that.',
  unauthorized: 'Please log in again.',
}

export function friendlyError(err: unknown): string {
  if (err instanceof ApiError) return MESSAGES[err.code] ?? err.message
  if (err instanceof Error) return err.message
  return 'Something went wrong. Please try again.'
}
