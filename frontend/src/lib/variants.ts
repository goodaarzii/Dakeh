import type { Variant } from './types'

// Variants have free-form attributes, for example { color: "red", size: "42" }.
// These helpers turn them into option groups ("color", "size") for the product page.

export function attributeKeys(variants: Variant[]): string[] {
  const keys: string[] = []
  for (const v of variants) {
    for (const key of Object.keys(v.attributes ?? {})) {
      if (!keys.includes(key)) keys.push(key)
    }
  }
  return keys
}

export function valuesFor(variants: Variant[], key: string): string[] {
  const values: string[] = []
  for (const v of variants) {
    const value = v.attributes?.[key]
    if (value !== undefined && !values.includes(value)) values.push(value)
  }
  return values
}

// The variant to show first: the first one that is in stock, otherwise the first one.
export function firstVariant(variants: Variant[]): Variant | undefined {
  return variants.find((v) => v.in_stock) ?? variants[0]
}

// The customer picks one value of one attribute. Keep the other choices if that combination exists,
// otherwise jump to the closest variant that has the new value.
export function pickByAttribute(variants: Variant[], current: Variant, key: string, value: string): Variant {
  const wanted: Record<string, string> = { ...current.attributes, [key]: value }
  const exact = variants.find((v) => Object.keys(wanted).every((k) => v.attributes?.[k] === wanted[k]))
  if (exact) return exact
  const sameValue = variants.filter((v) => v.attributes?.[key] === value)
  return sameValue.find((v) => v.in_stock) ?? sameValue[0] ?? current
}

// Is there an in-stock variant with this value, given the customer's other choices?
export function valueInStock(variants: Variant[], current: Variant, key: string, value: string): boolean {
  const others = Object.entries(current.attributes ?? {}).filter(([k]) => k !== key)
  return variants.some(
    (v) => v.in_stock && v.attributes?.[key] === value && others.every(([k, val]) => v.attributes?.[k] === val),
  )
}
