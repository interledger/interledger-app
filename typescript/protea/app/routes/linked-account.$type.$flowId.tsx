import type { LoaderArgs } from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData, useNavigate, useParams } from '@remix-run/react'
import { useCallback, useEffect } from 'react'
import { useScript } from '~/lib/useScript'
import { requireUserSession } from '~/lib/kratos.server'
import {
  grpcClient,
  httpMapping,
  isGrpcError,
  StatusError
} from '~/lib/proto.server'
import { Shape } from '~/components'
import { route } from 'routes-gen'

export async function loader({ request }: LoaderArgs) {
  await requireUserSession(request)
  let rpc = await grpcClient
    .getMachnetWidgetToken(
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
    throw json({}, httpMapping(rpc.code))
  }

  let widgetScriptUrl = 'https://widget.v4sandbox.machpay.com/widget/widget.js'
  return json({
    widgetScriptUrl,
    widgetUserId: rpc.response.userId,
    widgetToken: rpc.response.value
  })
}

export default function Page() {
  const data = useLoaderData<typeof loader>()
  const scriptStatus = useScript(data.widgetScriptUrl)
  const navigate = useNavigate()
  const params = useParams()

  const listener = useCallback(
    (event: any) => {
      if (event.data.type == 'CARD' && event.data.status == 'CARD_ADDED') {
        navigate(
          route('/linked-account/:type/:flowId/success', {
            type: params.type as string,
            flowId: params.flowId as string
          })
        )
      }
    },
    [navigate, params.flowId, params.type]
  )

  useEffect(() => {
    if (scriptStatus === 'ready') {
      const widget = new (window as any).MachnetWidget({
        elementId: 'widget',
        userId: data.widgetUserId,
        width: '100%',
        height: '200px',
        type: 'card',
        locale: 'en',
        stylesheet: '',
        token: data.widgetToken
      })
      widget.init()

      window.addEventListener('message', listener)
    }
  }, [data.widgetToken, data.widgetUserId, listener, scriptStatus])

  return (
    <div className='flex w-full flex-col rounded-2xl bg-page p-4 pb-8'>
      <div className='flex justify-between'>
        <h1 className='font-display text-2xl font-medium'>Debit card</h1>
        <div className='hidden sm:flex'>
          <Shape
            width={'w-8'}
            radius={'rounded-br-full'}
            color={'bg-lime-300'}
          />
          <Shape
            width={'w-8'}
            radius={'rounded-br-full'}
            color={'bg-slate-500'}
          />
        </div>
      </div>
      <p className='mt-6 text-medium'>
        Please provide your debit card details.
      </p>
      <div id='widget' className='mt-6 w-full' />
    </div>
  )
}
