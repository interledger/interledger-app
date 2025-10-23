import { ActionFunctionArgs, json } from '@remix-run/node'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'

export type Operation = 'freeze' | 'unfreeze' | 'block' | 'terminate'

export type OperationResponse = {
  success: boolean
  operation: Operation
  errors?: any
}

export async function action({ request }: ActionFunctionArgs) {
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
  if (operation === 'terminate') {
    return terminateCard(request, cardId)
  }

  return json({
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

  return json({
    success: true,
    operation: 'freeze'
  })
}

const unfreezeCard = async (request: Request, cardId: string) => {
  const unfreezeResponse = await grpc.unfreezeCard(request, {
    cardId
  })

  if (isConnectError(unfreezeResponse)) {
    return unfreezeResponse.error({ errors: { operation: 'unfreeze' } })
  }

  return json({
    success: true,
    operation: 'unfreeze'
  })
}

const blockCard = async (request: Request, cardId: string) => {
  const blockResponse = await grpc.blockCard(request, {
    cardId
  })

  if (isConnectError(blockResponse)) {
    return blockResponse.error({ errors: { operation: 'block' } })
  }

  return json({
    success: true,
    operation: 'block'
  })
}

const terminateCard = async (request: Request, cardId: string) => {
  const terminateResponse = await grpc.terminateCard(request, {
    cardId
  })

  if (isConnectError(terminateResponse)) {
    return terminateResponse.error({ errors: { operation: 'terminate' } })
  }

  return json({
    success: true,
    operation: 'terminate'
  })
}
