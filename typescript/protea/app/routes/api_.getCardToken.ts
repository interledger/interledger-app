import type { Route } from './+types/api_.getCardToken'
import { data } from 'react-router';
import type {
  CardTokenType,
  TokenLink
} from '~/generated/connect/backend/v1/backend_pb'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'

export type GetCardTokenResponse = {
  tokenType: CardTokenType
  token: string
  links: TokenLink[]
  shouldRevalidate?: boolean
  errors?: { operation: string }
}

export async function action({ request }: Route.ActionArgs) {
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
    return tokenResponse.error({ errors: { operation: 'card-token' } })
  }

  return data({
    tokenType,
    token: tokenResponse.token,
    links: tokenResponse.links,
    shouldRevalidate: false // Viewing sensitive data doesn't change card state
  })
}
