const formatter = new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' })

// The API sends cents: 1999 -> "$19.99".
export function formatPrice(cents: number): string {
  return formatter.format(cents / 100)
}

// 0 when there is no old price or it is not higher than the current one.
export function discountPercent(price: number, compareAt: number): number {
  if (compareAt <= 0 || compareAt <= price) return 0
  return Math.round((1 - price / compareAt) * 100)
}
