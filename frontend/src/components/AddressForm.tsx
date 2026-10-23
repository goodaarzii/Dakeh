import { useState, type FormEvent } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { api } from '../lib/api'
import { friendlyError } from '../lib/messages'
import type { Address, AddressInput } from '../lib/types'
import { MapPicker, type Point } from './MapPicker'
import { TextField } from './TextField'

interface TextValues {
  label: string
  recipient_name: string
  phone: string
  country: string
  province: string
  city: string
  street: string
  postal_code: string
}

type Errors = Partial<Record<keyof TextValues | 'location', string>>

function validate(values: TextValues, point: Point | null): Errors {
  const e: Errors = {}
  if (values.label.length > 50) e.label = 'Use at most 50 characters.'
  if (!values.recipient_name.trim()) e.recipient_name = 'Enter the name of the person who receives the order.'
  const phone = values.phone.trim()
  if (phone.length < 7 || phone.length > 20) e.phone = 'Enter a phone number with 7 to 20 characters.'
  if (!values.country.trim()) e.country = 'Enter the country.'
  if (!values.province.trim()) e.province = 'Enter the province or state.'
  if (!values.city.trim()) e.city = 'Enter the city.'
  if (!values.street.trim()) e.street = 'Enter the street and the house number.'
  if (values.postal_code.length > 20) e.postal_code = 'Use at most 20 characters.'
  if (!point) e.location = 'Click the map to drop a pin on the address.'
  return e
}

export function AddressForm({
  initial,
  hasDefault,
  onDone,
  onCancel,
}: {
  initial?: Address
  hasDefault: boolean
  onDone: () => void
  onCancel: () => void
}) {
  const queryClient = useQueryClient()
  const [values, setValues] = useState<TextValues>({
    label: initial?.label ?? '',
    recipient_name: initial?.recipient_name ?? '',
    phone: initial?.phone ?? '',
    country: initial?.country ?? '',
    province: initial?.province ?? '',
    city: initial?.city ?? '',
    street: initial?.street ?? '',
    postal_code: initial?.postal_code ?? '',
  })
  const [point, setPoint] = useState<Point | null>(initial ? { lat: initial.lat, lng: initial.lng } : null)
  const [isDefault, setIsDefault] = useState(initial?.is_default ?? false)
  const [errors, setErrors] = useState<Errors>({})
  const [formError, setFormError] = useState('')
  const [submitting, setSubmitting] = useState(false)

  const set = (key: keyof TextValues) => (e: { target: { value: string } }) => setValues((v) => ({ ...v, [key]: e.target.value }))

  async function onSubmit(e: FormEvent) {
    e.preventDefault()
    const found = validate(values, point)
    setErrors(found)
    setFormError('')
    if (Object.keys(found).length > 0 || !point) return

    const body: AddressInput = {
      label: values.label.trim(),
      recipient_name: values.recipient_name.trim(),
      phone: values.phone.trim(),
      country: values.country.trim(),
      province: values.province.trim(),
      city: values.city.trim(),
      street: values.street.trim(),
      postal_code: values.postal_code.trim(),
      lat: point.lat,
      lng: point.lng,
      is_default: isDefault,
    }

    setSubmitting(true)
    try {
      if (initial) await api.put<Address>(`/users/address/${initial.id}`, body)
      else await api.post<Address>('/users/address', body)
      await queryClient.invalidateQueries({ queryKey: ['addresses'] })
      onDone()
    } catch (err) {
      setFormError(friendlyError(err))
    } finally {
      setSubmitting(false)
    }
  }

  // The API sets the default only when an address is created, and allows one default per customer.
  const canChooseDefault = !initial && !hasDefault

  return (
    <form onSubmit={(e) => void onSubmit(e)} className="rounded-lg border-[1.5px] border-ink bg-paper p-6" noValidate>
      <h3 className="text-2xl">{initial ? 'Edit address' : 'New address'}</h3>

      <div className="mt-5">
        <MapPicker value={point} onChange={setPoint} error={errors.location} />
      </div>

      <div className="mt-6 grid grid-cols-2 gap-x-5 gap-y-4">
        <TextField label="Label" placeholder="Home, work…" value={values.label} onChange={set('label')} error={errors.label} />
        <TextField label="Recipient" autoComplete="name" value={values.recipient_name} onChange={set('recipient_name')} error={errors.recipient_name} />
        <TextField label="Phone" type="tel" autoComplete="tel" value={values.phone} onChange={set('phone')} error={errors.phone} />
        <TextField label="Postal code" autoComplete="postal-code" value={values.postal_code} onChange={set('postal_code')} error={errors.postal_code} />
        <TextField label="Country" autoComplete="country-name" value={values.country} onChange={set('country')} error={errors.country} />
        <TextField label="Province or state" value={values.province} onChange={set('province')} error={errors.province} />
        <TextField label="City" value={values.city} onChange={set('city')} error={errors.city} />
        <TextField label="Street and number" autoComplete="street-address" value={values.street} onChange={set('street')} error={errors.street} />
      </div>

      {canChooseDefault && (
        <label className="mt-5 flex items-center gap-2 text-sm font-semibold">
          <input type="checkbox" checked={isDefault} onChange={(e) => setIsDefault(e.target.checked)} className="h-4 w-4 accent-[#213448]" />
          Use as my default address
        </label>
      )}

      {formError && (
        <p role="alert" className="mt-5 rounded-lg border-[1.5px] border-brick/60 bg-sand px-4 py-3 text-sm text-brick">
          {formError}
        </p>
      )}

      <div className="mt-6 flex justify-end gap-3">
        <button type="button" className="btn btn-quiet" onClick={onCancel} disabled={submitting}>
          Cancel
        </button>
        <button type="submit" className="btn btn-primary" disabled={submitting}>
          {submitting ? 'Saving…' : initial ? 'Save changes' : 'Save address'}
        </button>
      </div>
    </form>
  )
}
