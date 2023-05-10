import type { LoaderArgs } from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import {
  StatusError,
  grpcClient,
  httpMapping,
  isGrpcError
} from '~/lib/proto.server'

export async function loader({ request }: LoaderArgs) {
  // get username from request params
  const url = new URL(request.url);
  const username = url.searchParams.get("username");

  let identity = await grpcClient
    .addIdentity(
      {
        platform: 'twitter',
        handle: username as string,
        public: true
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

  return [identity.instructions, ' ', identity.code]
}
