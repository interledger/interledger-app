import type { LoaderArgs } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import {
  StatusError,
  grpcClient,
  httpMapping,
  isGrpcError
} from '~/lib/proto.server'
import { useLoaderData, useNavigate } from '@remix-run/react'
import { Layouts } from '~/components'
import { useEffect } from 'react'

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
      id: resp.response.id
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
  let { id } = useLoaderData<typeof loader>()
  const nav = useNavigate()

  // We do this redirect clientside because the browser removes secure cookies when coming from another domain.
  useEffect(() => {
    nav(`/settings/linked-identities/${id}`)
  }, [id, nav])

  return null
}
