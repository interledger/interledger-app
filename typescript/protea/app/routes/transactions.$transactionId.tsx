import type { Route } from './+types/transactions.$transactionId'
import { redirect } from 'react-router';
import { href } from 'react-router'

export async function loader({ params }: Route.LoaderArgs) {
  return redirect(
    href('/payments/:paymentId', { paymentId: params.transactionId as string })
  )
}
