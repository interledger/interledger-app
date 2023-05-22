import type { LoaderArgs } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import {
  StatusError,
  grpcClient,
  httpMapping,
  isGrpcError
} from '~/lib/proto.server'
import { useLoaderData } from '@remix-run/react'
import { Layouts } from '~/components'

export const handle = {
  title: 'Twitter',
  layout: Layouts.FocusLayout
}

export async function loader({ request }: LoaderArgs) {
  // Check for state and code if not state create auth url
  let url = new URL(request.url)
  let state = url.searchParams.get('state')
  let code = url.searchParams.get('code')

  if (state && code) {
    let resp = await grpcClient.twitterCallback(
      {
        state: state.toString(),
        code: code.toString()
      },
      {
        meta: {
          cookies: String(request.headers.get('cookie')) || ''
        }
      }
    )

    if (isGrpcError(resp)) {
      throw json({}, httpMapping(resp.code))
    }

    return json({
      message: 'Success, close window.'
    })
  } else {
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
}

export default function Page() {
  let { message } = useLoaderData<typeof loader>()

  return <div>{message}</div>
}
