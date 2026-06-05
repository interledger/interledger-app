import { href, redirect } from 'react-router'
import type { Route } from './+types/transactions.$transactionId'

export async function loader({ params }: Route.LoaderArgs) {
  return redirect(
    href('/payments/:paymentId', { paymentId: params.transactionId as string })
  )
}
