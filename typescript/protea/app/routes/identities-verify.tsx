import type { LoaderArgs } from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import {
  StatusError,
  grpcClient,
  httpMapping,
  isGrpcError
} from '~/lib/proto.server'

export async function loader({ request, params }: LoaderArgs) {
  // get proof id and proof url from request params
  const { id, proof } = params

  let identity = await grpcClient
    .startIdentityVerification(
      {
        id: id as string,
        proof: proof as string
      },
      {
        meta: {
          cookies: String(request.headers.get('cookie')) || ''
        }
      }
    )
    .then((resp) => resp.response)
    .catch(StatusError)

  if (isGrpcError(identity)) {
    throw json({}, httpMapping(identity.code))
  }

  return json({ identity })
}

export default function Page() {
  const { identity } = useLoaderData<typeof loader>()

  return [identity.state]
}
