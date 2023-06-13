import * as widgetSdk from '@mxenabled/web-widget-sdk'
import type { ActionArgs, LoaderArgs } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { useLoaderData, useSubmit } from '@remix-run/react'
import { useEffect } from 'react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import { Card, Layouts } from '~/components'
import {
  StatusError,
  grpcClient,
  httpMapping,
  isGrpcError
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

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      back: route('/'),
      title: 'Bank details'
    }
  }
}

export default function Page() {
  const submit = useSubmit()
  const { url } = useLoaderData<typeof loader>()
  useEffect(() => {
    const widget = new widgetSdk.ConnectWidget({
      container: '#widget',
      url,
      onMemberConnected: (event) => {
        let formData = new FormData()
        formData.append('userGuid', event.user_guid)
        formData.append('memberGuid', event.member_guid)
        formData.append('sessionGuid', event.session_guid)
        submit(formData, {
          action: '/link-account/bank',
          method: 'post'
        })
      }
    })

    return () => {
      widget.unmount()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])
  return (
    <Card>
      <div id='widget' />
    </Card>
  )
}

export async function action({ request }: ActionArgs) {
  const form = await request.formData()
  let rpc = await grpcClient
    .createMXBankAccounts(
      {
        memberGuid: form.get('memberGuid') as string,
        sessionGuid: form.get('sessionGuid') as string,
        userGuid: form.get('userGuid') as string
      },
      {
        meta: {
          cookies: String(request.headers.get('cookie'))
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)

  if (isGrpcError(rpc)) {
    throw json(null, httpMapping(rpc.code))
  }

  return redirect(route('/accounts'))
}
