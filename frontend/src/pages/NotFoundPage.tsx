import { Link } from 'react-router-dom'
import { EmptyState } from '../components/States'

export default function NotFoundPage() {
  return (
    <EmptyState title="This page does not exist">
      <p>The address may be wrong, or the page was moved.</p>
      <Link to="/" className="btn btn-primary mt-4">
        Back to the shop
      </Link>
    </EmptyState>
  )
}
