import { ConnectWidget } from '@mxenabled/web-widget-sdk'
import type { ActionArgs, LoaderArgs } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { useActionData, useLoaderData, useSubmit } from '@remix-run/react'
import { useEffect } from 'react'
import { route } from 'routes-gen'
import { getCurrentFlow, updateFlow } from '~/lib/flows.server'
import { grpcClient, isGrpcError, StatusError } from '~/lib/proto.server'

export async function loader({ request, params }: LoaderArgs) {
  const flow = await getCurrentFlow(request, params)
  let rpc = await grpcClient
    .getBankAccountWidget(
      {},
      {
        meta: {
          cookies: request.headers.get('cookie') || ''
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)
  if (isGrpcError(rpc)) {
    throw rpc
  }

  return json({ widgetUrl: rpc.response.url, flow })
}

export default function Page() {
  const submit = useSubmit()
  const { widgetUrl, flow } = useLoaderData<typeof loader>()
  const actionData = useActionData<typeof action>()

  useEffect(function () {
    const widget = new ConnectWidget({
      container: '#widget',
      url: widgetUrl,
      onMemberConnected: (event) => {
        let formData = new FormData()
        formData.append('memberGuid', event.member_guid)
        formData.append('userGuid', event.user_guid)

        submit(formData, { method: 'post', action: `/flows/${flow.id}/linked-account/type` })
        widget.unmount()
      }
    })
  }, [])

  return (
    <>
      <div id='widget' className='col-span-full sm:col-span-6 sm:col-start-2 lg:col-start-4 h-screen'></div>
    </>
  )
}

export async function action({ request, params }: ActionArgs) {
  const form = await request.formData()
  const memberGuid = form.get('memberGuid')?.toString() || ''
  const userGuid = form.get('userGuid')?.toString() || ''
  let rpc = await grpcClient
    .addBankAccount(
      {
        memberGuid,
        userGuid,
      },
      {
        meta: {
          cookies: request.headers.get('cookie') || ''
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)
  if (isGrpcError(rpc)) {
    throw rpc
  }

  const headers = await updateFlow(request, {
    memberGuid: memberGuid,
    userGuid: userGuid,
    fundingsourceId: rpc.response.fundingsourceId,
  })

  const flow = await getCurrentFlow(request, params)
  return redirect(
    route('/flows/:flowId/linked-account/review', {
      flowId: flow?.id as string
    }),
    { headers }
  )
}
