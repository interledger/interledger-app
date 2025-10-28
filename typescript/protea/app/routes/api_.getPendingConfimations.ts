import type { LoaderFunctionArgs } from '@remix-run/node'
import { json } from '@remix-run/node'
import type { PendingConfirmation } from '~/lib/mocks/confirmations'
import { mockPendingConfirmations } from '~/lib/mocks/confirmations'

export type GetPendingConfirmationsResponse = {
  confirmations: PendingConfirmation[]
  errors?: any
}

export async function loader({ request }: LoaderFunctionArgs) {
  // TODO: Replace with actual gRPC call
  // const response = await grpc.getPendingConfirmations(request, {})
  // if (isConnectError(response)) {
  //   return response.error({ errors: { operation: 'get-pending-confirmations' } })
  // }
  console.log('[api_.getPendingConfimations] 🐳 loader called')

  return json({
    confirmations: mockPendingConfirmations
  })
}
