// These types mirror the JSON that the Go API sends. Prices are USD cents.

export interface User {
  id: string
  email: string
  phone: string
  full_name: string
  role: string
  created_at: string
}

export interface AuthResponse {
  user: User
  access_token: string
  refresh_token: string
  token_type: string
  expires_in: number
}

export interface CategoryNode {
  id: string
  name: string
  slug: string
  image_url: string
  children: CategoryNode[]
}

export interface Tag {
  id: string
  name: string
  slug: string
}

export interface ProductCardData {
  id: string
  name: string
  slug: string
  brand: string
  min_price: number
  in_stock: boolean
  image_url: string
  avg_rating: number
  rating_count: number
}

export interface ProductList {
  items: ProductCardData[]
  page: number
  page_size: number
  has_more: boolean
}

export interface FavoriteProduct extends ProductCardData {
  category_id: string
  favorited_at: string
}

export interface FavoriteList {
  items: FavoriteProduct[]
  count: number
}

export interface FavoriteStatus {
  product_id: string
  favorited_at: string
  is_favorite: boolean
}

export interface ProductImage {
  url: string
  alt_text: string
  is_primary: boolean
  variant_id: string | null
}

export interface Variant {
  id: string
  sku: string
  name: string
  price: number
  compare_at_price: number
  attributes: Record<string, string>
  in_stock: boolean
}

export interface ProductDetail {
  id: string
  name: string
  slug: string
  description: string
  brand: string
  category: { id: string; name: string; slug: string }
  tags: Tag[]
  images: ProductImage[]
  variants: Variant[]
  avg_rating: number
  rating_count: number
}

export interface ApiCartLine {
  variant_id: string
  quantity: number
  available: boolean
  in_stock: boolean
  product_id?: string
  product_name?: string
  product_slug?: string
  variant_name?: string
  sku?: string
  attributes?: Record<string, string>
  image_url?: string
  unit_price: number
  compare_at_price: number
  line_total: number
}

export interface ApiCart {
  items: ApiCartLine[]
  subtotal: number
  item_count: number
  can_checkout: boolean
}

export interface Address {
  id: string
  label: string
  recipient_name: string
  phone: string
  country: string
  province: string
  city: string
  street: string
  postal_code: string
  lat: number
  lng: number
  is_default: boolean
  created_at: string
  updated_at: string
}

export interface AddressInput {
  label: string
  recipient_name: string
  phone: string
  country: string
  province: string
  city: string
  street: string
  postal_code: string
  lat: number
  lng: number
  is_default: boolean
}

export interface StatusResponse {
  code: number
  message: string
  is_success: boolean
}


export interface Review {
  id: string
  product_id: string
  reviewer_name: string
  rating: number
  title: string
  body: string
  created_at: string
  updated_at: string
}

export interface ReviewList {
  items: Review[]
  count: number
  page: number
  page_size: number
  has_more: boolean
}

export interface OrderAddressSnapshot {
  label: string
  recipient_name: string
  phone: string
  country: string
  province: string
  city: string
  street: string
  postal_code: string
}

export interface OrderItem {
  id: string
  product_id: string
  variant_id: string
  product_name: string
  product_slug: string
  variant_name: string
  sku: string
  attributes: Record<string, string>
  image_url: string
  unit_price: number
  quantity: number
  line_total: number
}

export interface Order {
  id: string
  status: 'submitted' | 'cancelled' | string
  currency: string
  subtotal: number
  shipping_fee: number
  total: number
  shipping_address: OrderAddressSnapshot
  items: OrderItem[]
  created_at: string
  updated_at: string
  cancelled_at?: string | null
}

export interface OrderList {
  items: Order[]
  count: number
  page: number
  page_size: number
  has_more: boolean
}

export interface CreateOrderResponse {
  order: Order
  cart_cleared: boolean
}
