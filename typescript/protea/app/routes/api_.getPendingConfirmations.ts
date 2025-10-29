import type { ActionFunctionArgs } from '@remix-run/node'
import { json } from '@remix-run/node'
import type { PendingThreeDSConfirmation } from '~/generated/connect/backend/v1/backend_pb'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'

export type GetPendingConfirmationsResponse = {
  confirmations: PendingThreeDSConfirmation[]
  errors?: {
    operation: string
  }
}

export async function action({ request }: ActionFunctionArgs) {
  const response = await grpc.getPendingThreeDSConfirmations(request, {})
  if (isConnectError(response)) {
    return json<GetPendingConfirmationsResponse>(
      {
        confirmations: [],
        errors: { operation: 'get-pending-confirmations' }
      },
      { status: 400 }
    )
  }

  return json<GetPendingConfirmationsResponse>({
    confirmations: response.confirmations
  })
}
