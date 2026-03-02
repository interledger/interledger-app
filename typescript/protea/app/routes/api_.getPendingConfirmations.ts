import type { ActionFunctionArgs } from 'react-router';
import { data } from 'react-router';
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
    return data<GetPendingConfirmationsResponse>(
      {
        confirmations: [],
        errors: { operation: 'get-pending-confirmations' }
      },
      { status: 400 }
    )
  }

  return data<GetPendingConfirmationsResponse>({
    confirmations: response.confirmations
  })
}
