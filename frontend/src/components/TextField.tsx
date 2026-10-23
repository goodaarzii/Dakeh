import { useId, type InputHTMLAttributes } from 'react'

interface TextFieldProps extends InputHTMLAttributes<HTMLInputElement> {
  label: string
  error?: string
  hint?: string
}

export function TextField({ label, error, hint, className, ...rest }: TextFieldProps) {
  const id = useId()
  const messageId = `${id}-message`
  return (
    <div>
      <label htmlFor={id} className="mb-1.5 block text-sm font-semibold">
        {label}
      </label>
      <input
        id={id}
        className={`field ${className ?? ''}`}
        aria-invalid={error ? true : undefined}
        aria-describedby={error || hint ? messageId : undefined}
        {...rest}
      />
      {error ? (
        <p id={messageId} className="mt-1.5 text-sm text-brick">
          {error}
        </p>
      ) : hint ? (
        <p id={messageId} className="mt-1.5 text-sm text-ink/65">
          {hint}
        </p>
      ) : null}
    </div>
  )
}
