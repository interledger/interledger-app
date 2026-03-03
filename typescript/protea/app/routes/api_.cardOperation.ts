import type { Route } from './+types/api_.cardOperation'
import { data } from 'react-router';
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'

export type Operation = 'freeze' | 'unfreeze' | 'block'

export type OperationResponse = {
  success: boolean
  operation: Operation
  shouldRevalidate?: boolean
  errors?: { operation: string }
}

export async function action({ request }: Route.ActionArgs) {
  let form = await request.formData()
  const cardId = form.get('cardId') as string
  const operation = form.get('operation') as Operation

  if (operation === 'freeze') {
    return freezeCard(request, cardId)
  }
  if (operation === 'unfreeze') {
    return unfreezeCard(request, cardId)
  }
  if (operation === 'block') {
    return blockCard(request, cardId)
  }

  return data({
    success: false,
    errors: { operation: 'invalidOperation' }
  })
}

const freezeCard = async (request: Request, cardId: string) => {
  const freezeResponse = await grpc.freezeCard(request, {
    cardId
  })

  if (isConnectError(freezeResponse)) {
    return freezeResponse.error({ errors: { operation: 'freeze' } })
  }

  return data({
    success: true,
    operation: 'freeze',
    shouldRevalidate: true
  })
}

const unfreezeCard = async (request: Request, cardId: string) => {
  const unfreezeResponse = await grpc.unfreezeCard(request, {
    cardId
  })

  if (isConnectError(unfreezeResponse)) {
    return unfreezeResponse.error({ errors: { operation: 'unfreeze' } })
  }

  return data({
    success: true,
    operation: 'unfreeze',
    shouldRevalidate: true
  })
}

const blockCard = async (request: Request, cardId: string) => {
  const blockResponse = await grpc.blockCard(request, {
    cardId
  })

  if (isConnectError(blockResponse)) {
    return blockResponse.error({ errors: { operation: 'block' } })
  }

  return data({
    success: true,
    operation: 'block',
    shouldRevalidate: true
  })
}
