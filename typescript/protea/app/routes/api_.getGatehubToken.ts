import { ActionFunctionArgs, json } from '@remix-run/node'
import {
  CardTokenType,
  TokenLink
} from '~/generated/connect/backend/v1/backend_pb'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'

export type GetGatehubTokenResponse = {
  tokenType: CardTokenType
  token: string
  links: TokenLink[]
  errors?: any
}

export async function action({ request }: ActionFunctionArgs) {
  let form = await request.formData()
  const cardId = form.get('cardId') as string
  const tokenType = Number(form.get('tokenType')) as CardTokenType
  const publicKey = form.get('publicKey') as string

  const tokenResponse = await grpc.getCardToken(request, {
    tokenType,
    cardId,
    publicKey
  })

  if (isConnectError(tokenResponse)) {
    return tokenResponse.error({ errors: { form: 'Failed to get card token' } })
  }

  return json({
    tokenType,
    token: tokenResponse.token,
    links: tokenResponse.links
  })
}
