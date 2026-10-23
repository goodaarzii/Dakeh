import { createContext, useCallback, useContext, useEffect, useMemo, useRef, useState, type ReactNode } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from './api'
import { useAuth } from './auth'
import { useToast } from '../components/Toast'
import { MAX_CART_LINES, MAX_QTY_PER_LINE } from '../config'
import type { ApiCart } from './types'

// One cart for the whole app, with two sources:
//  - a GUEST cart in this browser (localStorage), until the customer logs in
//  - the SERVER cart (Redis, through the API) for a logged-in customer
// Pages only use useCart() and do not care which one it is.

export interface CartLineView {
  variantId: string
  quantity: number
  productName: string
  productSlug: string
  variantName: string
  sku: string
  attributes: Record<string, string>
  imageUrl: string
  unitPrice: number
  compareAtPrice: number
  lineTotal: number
  available: boolean // false: the shop stopped selling it
  inStock: boolean // false: not enough stock for this quantity
}

export interface CartView {
  lines: CartLineView[]
  itemCount: number
  subtotal: number
  canCheckout: boolean
}

// What the product page knows about a variant. A guest cart stores it, because a guest has no server cart to ask.
export interface NewCartItem {
  variantId: string
  productName: string
  productSlug: string
  variantName: string
  sku: string
  attributes: Record<string, string>
  imageUrl: string
  unitPrice: number
  compareAtPrice: number
}

interface GuestItem extends NewCartItem {
  quantity: number
}

const GUEST_KEY = 'dakeh.guestCart'
const EMPTY_CART: CartView = { lines: [], itemCount: 0, subtotal: 0, canCheckout: false }

function readGuest(): GuestItem[] {
  try {
    const raw = localStorage.getItem(GUEST_KEY)
    if (!raw) return []
    const parsed: unknown = JSON.parse(raw)
    return Array.isArray(parsed) ? (parsed as GuestItem[]) : []
  } catch {
    return []
  }
}

function fromGuest(items: GuestItem[]): CartView {
  const lines: CartLineView[] = items.map((i) => ({
    variantId: i.variantId,
    quantity: i.quantity,
    productName: i.productName,
    productSlug: i.productSlug,
    variantName: i.variantName,
    sku: i.sku,
    attributes: i.attributes ?? {},
    imageUrl: i.imageUrl,
    unitPrice: i.unitPrice,
    compareAtPrice: i.compareAtPrice,
    lineTotal: i.unitPrice * i.quantity,
    available: true,
    inStock: true, // the stock is checked when the customer logs in
  }))
  return {
    lines,
    itemCount: lines.reduce((sum, l) => sum + l.quantity, 0),
    subtotal: lines.reduce((sum, l) => sum + l.lineTotal, 0),
    canCheckout: lines.length > 0,
  }
}

function fromServer(cart: ApiCart | undefined): CartView {
  if (!cart) return EMPTY_CART
  return {
    lines: cart.items.map((i) => ({
      variantId: i.variant_id,
      quantity: i.quantity,
      productName: i.product_name ?? '',
      productSlug: i.product_slug ?? '',
      variantName: i.variant_name ?? '',
      sku: i.sku ?? '',
      attributes: i.attributes ?? {},
      imageUrl: i.image_url ?? '',
      unitPrice: i.unit_price,
      compareAtPrice: i.compare_at_price,
      lineTotal: i.line_total,
      available: i.available,
      inStock: i.in_stock,
    })),
    itemCount: cart.item_count,
    subtotal: cart.subtotal,
    canCheckout: cart.can_checkout,
  }
}

interface CartContextValue {
  cart: CartView
  isGuest: boolean
  loading: boolean
  error: Error | null
  addItem: (item: NewCartItem, quantity: number) => Promise<void>
  setQuantity: (variantId: string, quantity: number) => Promise<void>
  removeItem: (variantId: string) => Promise<void>
  clear: () => Promise<void>
  refetch: () => void
}

const CartContext = createContext<CartContextValue | null>(null)

export function CartProvider({ children }: { children: ReactNode }) {
  const { user } = useAuth()
  const queryClient = useQueryClient()
  const toast = useToast()
  const [guestItems, setGuestItems] = useState<GuestItem[]>(readGuest)

  useEffect(() => {
    localStorage.setItem(GUEST_KEY, JSON.stringify(guestItems))
  }, [guestItems])

  const serverCart = useQuery({
    queryKey: ['cart'],
    queryFn: () => api.get<ApiCart>('/cart'),
    enabled: !!user,
  })

  // When a customer logs in with a guest cart, add its items to the account cart (once).
  const merging = useRef(false)
  const userId = user?.id
  useEffect(() => {
    if (!userId || guestItems.length === 0 || merging.current) return
    merging.current = true
    const items = guestItems.map((i) => ({ variant_id: i.variantId, quantity: i.quantity }))
    api
      .post<ApiCart>('/cart/merge', { items })
      .then((cart) => {
        queryClient.setQueryData(['cart'], cart)
        setGuestItems([])
        toast('Your guest cart was added to your account.', { type: 'success' })
      })
      .catch(() => {
        toast('Your guest cart could not be added to your account. It stays on this device.', { type: 'error' })
      })
      .finally(() => {
        merging.current = false
      })
  }, [userId]) // eslint-disable-line react-hooks/exhaustive-deps

  const cart = useMemo(
    () => (user ? fromServer(serverCart.data) : fromGuest(guestItems)),
    [user, serverCart.data, guestItems],
  )

  const putServerQuantity = useCallback(
    async (variantId: string, quantity: number) => {
      const updated = await api.put<ApiCart>(`/cart/items/${variantId}`, { quantity })
      queryClient.setQueryData(['cart'], updated)
    },
    [queryClient],
  )

  const addItem = useCallback(
    async (item: NewCartItem, quantity: number) => {
      if (user) {
        const current = queryClient
          .getQueryData<ApiCart>(['cart'])
          ?.items.find((i) => i.variant_id === item.variantId)?.quantity ?? 0
        await putServerQuantity(item.variantId, Math.min(MAX_QTY_PER_LINE, current + quantity))
        return
      }
      setGuestItems((prev) => {
        const existing = prev.find((i) => i.variantId === item.variantId)
        if (existing) {
          return prev.map((i) =>
            i.variantId === item.variantId
              ? { ...i, ...item, quantity: Math.min(MAX_QTY_PER_LINE, i.quantity + quantity) }
              : i,
          )
        }
        if (prev.length >= MAX_CART_LINES) return prev
        return [...prev, { ...item, quantity: Math.min(MAX_QTY_PER_LINE, quantity) }]
      })
    },
    [user, queryClient, putServerQuantity],
  )

  const setQuantity = useCallback(
    async (variantId: string, quantity: number) => {
      const next = Math.max(1, Math.min(MAX_QTY_PER_LINE, quantity))
      if (user) {
        await putServerQuantity(variantId, next)
        return
      }
      setGuestItems((prev) => prev.map((i) => (i.variantId === variantId ? { ...i, quantity: next } : i)))
    },
    [user, putServerQuantity],
  )

  const removeItem = useCallback(
    async (variantId: string) => {
      if (user) {
        const updated = await api.delete<ApiCart>(`/cart/items/${variantId}`)
        queryClient.setQueryData(['cart'], updated)
        return
      }
      setGuestItems((prev) => prev.filter((i) => i.variantId !== variantId))
    },
    [user, queryClient],
  )

  const clear = useCallback(async () => {
    if (user) {
      await api.delete<void>('/cart')
      queryClient.setQueryData<ApiCart>(['cart'], { items: [], subtotal: 0, item_count: 0, can_checkout: false })
      return
    }
    setGuestItems([])
  }, [user, queryClient])

  const value = useMemo<CartContextValue>(
    () => ({
      cart,
      isGuest: !user,
      loading: !!user && serverCart.isLoading,
      error: user ? (serverCart.error as Error | null) : null,
      addItem,
      setQuantity,
      removeItem,
      clear,
      refetch: () => void serverCart.refetch(),
    }),
    [cart, user, serverCart.isLoading, serverCart.error, serverCart, addItem, setQuantity, removeItem, clear],
  )

  return <CartContext.Provider value={value}>{children}</CartContext.Provider>
}

export function useCart(): CartContextValue {
  const ctx = useContext(CartContext)
  if (!ctx) throw new Error('useCart must be used inside CartProvider')
  return ctx
}
