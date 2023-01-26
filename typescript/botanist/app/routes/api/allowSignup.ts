import type { ActionArgs } from '@remix-run/node'
import { json } from '@remix-run/node'
import {
  grpcClient,
  httpMapping,
  isGrpcError,
  StatusError
} from '~/lib/proto.server'

export async function action({ request }: ActionArgs) {
  const form = await request.formData()
  const id = form.get('id') as string

  let response = await grpcClient
    .allowWaitlistSignup(
      {
        id
      },
      {
        meta: {
          cookies: String(request.headers.get('cookie')) || ''
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)
  if (isGrpcError(response)) {
    throw json({}, httpMapping(response.code))
  }

  return json({ success: true })
}
