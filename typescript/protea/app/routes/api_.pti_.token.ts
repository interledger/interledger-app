import { json, type ActionFunctionArgs } from '@remix-run/node'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { getUserSession } from '~/lib/kratos.server'

export async function action({ request }: ActionFunctionArgs) {
  await getUserSession(request) // must be authenticated
  const payload = await request.json()
  const response = await grpc.createPtiToken(request, {
    url: payload.url,
    method: payload.method
  })

  if (isConnectError(response)) throw response.errorResponse

  return json({
    accessToken: response.accessToken,
    expiresAt: response.expiresAt,
    tokenType: response.tokenType
  })
}
