import { json, type ActionFunctionArgs } from '@remix-run/node'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'

export async function action({ request }: ActionFunctionArgs) {
  try {
    const payload = await request.json()
    console.log(payload)
    const response = await grpc.createPtiToken(request, {
      url: payload['x-pti-token-payload'].url,
      method: payload['x-pti-token-payload'].method
    })

    if (isConnectError(response)) throw response.errorResponse

    return json({
      accessToken: response.accessToken,
      expiresAt: response.expiresAt,
      tokenType: response.tokenType
    })
  } catch (err) {
    console.log(err)
  }
}
