import type { LoaderFunctionArgs } from '@remix-run/node'
import { redirect } from '@remix-run/node'
import { route } from 'routes-gen'

export async function loader({ params }: LoaderFunctionArgs) {
  return redirect(
    route('/payments/:paymentId', { paymentId: params.transactionId as string })
  )
}
