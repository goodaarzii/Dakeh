import { useState, type ChangeEvent, type FormEvent } from 'react'
import { Link, Navigate, useNavigate, useSearchParams } from 'react-router-dom'
import { TextField } from '../components/TextField'
import { useAuth } from '../lib/auth'
import { friendlyError } from '../lib/messages'
import { safeNext } from './LoginPage'

interface Form {
  full_name: string
  email: string
  phone: string
  password: string
  confirm: string
}

export default function RegisterPage() {
  const { user, register } = useAuth()
  const [params] = useSearchParams()
  const navigate = useNavigate()
  const next = safeNext(params.get('next'))

  const [form, setForm] = useState<Form>({ full_name: '', email: '', phone: '', password: '', confirm: '' })
  const [errors, setErrors] = useState<Partial<Record<keyof Form, string>>>({})
  const [formError, setFormError] = useState('')
  const [submitting, setSubmitting] = useState(false)

  if (user) return <Navigate to={next} replace />

  const set = (key: keyof Form) => (e: ChangeEvent<HTMLInputElement>) => setForm((f) => ({ ...f, [key]: e.target.value }))

  // The same limits as the API, so people get the message here and not as a server error.
  function validate(): Partial<Record<keyof Form, string>> {
    const found: Partial<Record<keyof Form, string>> = {}
    if (!form.full_name.trim()) found.full_name = 'Enter your name.'
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(form.email.trim())) found.email = 'Enter a valid email address.'
    const phone = form.phone.trim()
    if (phone.length < 7 || phone.length > 20) found.phone = 'Enter a phone number with 7 to 20 characters.'
    if (form.password.length < 8 || form.password.length > 72) found.password = 'Use 8 to 72 characters.'
    if (form.confirm !== form.password) found.confirm = 'The passwords do not match.'
    return found
  }

  async function onSubmit(e: FormEvent) {
    e.preventDefault()
    const found = validate()
    setErrors(found)
    setFormError('')
    if (Object.keys(found).length > 0) return

    setSubmitting(true)
    try {
      await register({
        full_name: form.full_name.trim(),
        email: form.email.trim(),
        phone: form.phone.trim(),
        password: form.password,
      })
      navigate(next, { replace: true })
    } catch (err) {
      setFormError(friendlyError(err))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="mx-auto max-w-md">
      <h1 className="text-4xl">Create an account</h1>
      <p className="mt-2 text-ink/75">One account for your cart, your addresses and, soon, your orders.</p>

      <form onSubmit={(e) => void onSubmit(e)} className="mt-8 space-y-5" noValidate>
        <TextField label="Full name" autoComplete="name" value={form.full_name} onChange={set('full_name')} error={errors.full_name} />
        <TextField label="Email" type="email" autoComplete="email" value={form.email} onChange={set('email')} error={errors.email} />
        <TextField label="Phone" type="tel" autoComplete="tel" value={form.phone} onChange={set('phone')} error={errors.phone} />
        <TextField
          label="Password"
          type="password"
          autoComplete="new-password"
          value={form.password}
          onChange={set('password')}
          error={errors.password}
          hint="At least 8 characters."
        />
        <TextField
          label="Repeat the password"
          type="password"
          autoComplete="new-password"
          value={form.confirm}
          onChange={set('confirm')}
          error={errors.confirm}
        />
        {formError && (
          <p role="alert" className="rounded-lg border-[1.5px] border-brick/60 bg-paper px-4 py-3 text-sm text-brick">
            {formError}
          </p>
        )}
        <button type="submit" className="btn btn-primary w-full !py-3" disabled={submitting}>
          {submitting ? 'Creating your account…' : 'Create account'}
        </button>
      </form>

      <p className="mt-6 text-ink/75">
        Already have an account?{' '}
        <Link to={`/login${params.get('next') ? `?next=${encodeURIComponent(next)}` : ''}`} className="font-semibold underline underline-offset-2">
          Log in
        </Link>
      </p>
    </div>
  )
}
