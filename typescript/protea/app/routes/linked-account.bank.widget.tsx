import * as widgetSdk from '@mxenabled/web-widget-sdk'
import type { LoaderArgs } from '@remix-run/node';
import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { useEffect } from 'react'
import { Layouts } from '~/components'
import { getUserSession } from '~/lib/kratos.server'
import {
  grpcClient,
  httpMapping,
  isGrpcError,
  StatusError
} from '~/lib/proto.server'

export async function loader({ request }: LoaderArgs) {
  let rpc = await grpcClient
    .getMXWidget(
      {},
      {
        meta: { cookies: String(request.headers.get('cookie')) }
      }
    )
    .then((v) => v)
    .catch(StatusError)
  if (isGrpcError(rpc)) {
    throw json({}, httpMapping(rpc.code))
  }

  return json({ url: rpc.response.url })
}

export const handle = { layout: Layouts.FocusLayout }

export default function Page() {
  const { url } = useLoaderData<typeof loader>()
  useEffect(() => {
    const widget = new widgetSdk.ConnectWidget({
      container: '#widget',
      url,
      onMessage: (event) => {
        console.log(event.data)
      }
    })

    return () => {
      widget.unmount()
    }
  }, [])
  return (
    <>
      <div id='widget' />
    </>
  )
}
