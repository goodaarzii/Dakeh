import { useState, type FormEvent } from 'react'
import { Link, Navigate, useNavigate, useSearchParams } from 'react-router-dom'
import { TextField } from '../components/TextField'
import { useAuth } from '../lib/auth'
import { friendlyError } from '../lib/messages'

// Only allow redirects to pages of this app (never to another site).
export function safeNext(raw: string | null): string {
  return raw && raw.startsWith('/') && !raw.startsWith('//') ? raw : '/'
}

export default function LoginPage() {
  const { user, login } = useAuth()
  const [params] = useSearchParams()
  const navigate = useNavigate()
  const next = safeNext(params.get('next'))

  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [errors, setErrors] = useState<{ email?: string; password?: string }>({})
  const [formError, setFormError] = useState('')
  const [submitting, setSubmitting] = useState(false)

  if (user) return <Navigate to={next} replace />

  async function onSubmit(e: FormEvent) {
    e.preventDefault()
    const found: { email?: string; password?: string } = {}
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.trim())) found.email = 'Enter a valid email address.'
    if (password.length < 8) found.password = 'The password has at least 8 characters.'
    setErrors(found)
    setFormError('')
    if (Object.keys(found).length > 0) return

    setSubmitting(true)
    try {
      await login(email.trim(), password)
      navigate(next, { replace: true })
    } catch (err) {
      setFormError(friendlyError(err))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="mx-auto max-w-md">
      <h1 className="text-4xl">Log in</h1>
      <p className="mt-2 text-ink/75">Welcome back. Your cart and addresses are waiting.</p>

      <form onSubmit={(e) => void onSubmit(e)} className="mt-8 space-y-5" noValidate>
        <TextField label="Email" type="email" autoComplete="email" value={email} onChange={(e) => setEmail(e.target.value)} error={errors.email} />
        <TextField
          label="Password"
          type="password"
          autoComplete="current-password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          error={errors.password}
        />
        {formError && (
          <p role="alert" className="rounded-lg border-[1.5px] border-brick/60 bg-paper px-4 py-3 text-sm text-brick">
            {formError}
          </p>
        )}
        <button type="submit" className="btn btn-primary w-full !py-3" disabled={submitting}>
          {submitting ? 'Logging in…' : 'Log in'}
        </button>
      </form>

      <p className="mt-6 text-ink/75">
        New here?{' '}
        <Link to={`/register${params.get('next') ? `?next=${encodeURIComponent(next)}` : ''}`} className="font-semibold underline underline-offset-2">
          Create an account
        </Link>
      </p>
    </div>
  )
}
