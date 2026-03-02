import type { LoaderFunctionArgs } from 'react-router';
import { redirect } from 'react-router';
import { href } from 'react-router'

export async function loader({ params }: LoaderFunctionArgs) {
  return redirect(
    href('/payments/:paymentId', { paymentId: params.transactionId as string })
  )
}
