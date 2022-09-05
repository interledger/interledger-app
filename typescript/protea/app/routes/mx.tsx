import { ConnectWidget } from '@mxenabled/web-widget-sdk'
import { ActionArgs, json, LoaderArgs } from '@remix-run/node'
import { useActionData, useLoaderData, useSubmit } from '@remix-run/react'
import { useEffect } from 'react'
import { requireUserSession } from '~/lib/kratos.server'
import { grpcClient, isGrpcError, StatusError } from '~/lib/proto.server'

export async function loader({ request }: LoaderArgs) {
  await requireUserSession(request)
  let response = await grpcClient
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
  if (isGrpcError(response)) {
    throw response
  }

  return json({ widgetUrl: response.response.url })
}

export async function action({ request }: ActionArgs) {
  let data = await request.formData()
  let response = await grpcClient
    .addBankAccount(
      {
        memberGuid: data.get('memberGuid')?.toString() || '',
        userGuid: data.get('userGuid')?.toString() || '',
      },
      {
        meta: {
          cookies: request.headers.get('cookie') || ''
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)
  if (isGrpcError(response)) {
    throw response
  }

  return json({ fundingsourceId: response.response.fundingsourceId })
}

export default function Page() {
  const submit = useSubmit()
  const actionData = useActionData<typeof action>()
  const { widgetUrl } = useLoaderData<typeof loader>()
  useEffect(function () {
    const widget = new ConnectWidget({
      container: '#widget',
      url: widgetUrl,
      onMemberConnected: (event) => {
        console.log('memberGuid', event.member_guid)
        console.log('userGuid', event.user_guid)
        let formData = new FormData()
        formData.append('memberGuid', event.member_guid)
        formData.append('userGuid', event.user_guid)
        submit(formData, { method: 'post', action: '/mx' })
        widget.unmount()
      }
    })
  }, [])

  return (
    <>
      <div id='widget' className='mx-auto h-screen w-full'></div>
      <div>reference={actionData?.fundingsourceId}</div>
    </>
  )
}
