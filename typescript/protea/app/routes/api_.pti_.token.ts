import { data } from 'react-router'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import type { Route } from './+types/api_.pti_.token'

export async function action({ request }: Route.ActionArgs) {
  const payload = await request.json()
  const response = await grpc.createPtiToken(request, {
    url: payload['x-pti-token-payload'].url,
    method: payload['x-pti-token-payload'].method
  })

  if (isConnectError(response)) throw response.errorResponse

  return data({
    accessToken: response.accessToken,
    expiresAt: response.expiresAt,
    tokenType: response.tokenType
  })
}
