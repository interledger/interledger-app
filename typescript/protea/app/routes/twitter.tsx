import type { LoaderArgs } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import {
  StatusError,
  grpcClient,
  httpMapping,
  isGrpcError
} from '~/lib/proto.server'

export async function loader({ request }: LoaderArgs) {
  let resp = await grpcClient
    .createTwitterAuthURL(
      {},
      {
        meta: {
          cookies: String(request.headers.get('cookie')) || ''
        }
      }
    )
    .then((resp) => resp.response)
    .catch(StatusError)

  if (isGrpcError(resp)) {
    throw json({}, httpMapping(resp.code))
  }

  return redirect(resp.url)
}

export default function Page() {
  return []
}
