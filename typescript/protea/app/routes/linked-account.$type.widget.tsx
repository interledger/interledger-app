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
import { Layouts, Shape } from '~/components'
import { route } from 'routes-gen'
import { flowType, requireFlow, updateFlow } from '~/lib/flows.server'
import { getLinkedAccounts } from '~/lib/wallet.server'

export async function loader({ request, params }: LoaderArgs) {
  await requireUserSession(request)

  const linkedAccounts = await getLinkedAccounts(request)
  if (params.type == 'card') {
    await requireFlow(request, flowType.LinkCardAccount)
    await updateFlow(request, flowType.LinkCardAccount, {
      linkedAccountLength: linkedAccounts.linkedAccounts.length
    })
  } else if (params.type == 'bank') {
    await requireFlow(request, flowType.LinkBankAccount)
    await updateFlow(request, flowType.LinkBankAccount, {
      linkedAccountLength: linkedAccounts.linkedAccounts.length
    })
  } else {
    throw json(
      { title: `Linking type ${params.type} not allowed.` },
      { status: 400 }
    )
  }

  let cardRpc = await grpcClient
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
  if (isGrpcError(cardRpc)) {
    throw json({}, httpMapping(cardRpc.code))
  }

  return json({
    widgetScriptUrl:
      process.env.MACHNET_WIDGET_URL ||
      'https://widget.v4sandbox.machpay.com/widget/widget.js',
    widgetUserId: cardRpc.response.userId,
    widgetToken: cardRpc.response.value
  })
}

export const handle = {
  layout: Layouts.FocusLayout
}

export default function Page() {
  const params = useParams()
  const navigate = useNavigate()

  const { widgetScriptUrl, widgetUserId, widgetToken } =
    useLoaderData<typeof loader>()
  const scriptStatus = useScript(widgetScriptUrl)

  const listener = useCallback(
    (event: any) => {
      if (
        (event.data.type == 'CARD' && event.data.status == 'CARD_ADDED') ||
        (event.data.type == 'BANK' && event.data.status == 'BANK_ADDED')
      ) {
        navigate(
          route('/linked-account/:type/almost-there', {
            type: params.type as string
          })
        )
      }
    },
    [navigate, params.type]
  )

  useEffect(() => {
    if (scriptStatus === 'ready') {
      const widget = new (window as any).MachnetWidget({
        elementId: 'widget',
        userId: widgetUserId,
        width: '100%',
        height: params.type == 'card' ? '258px' : '591px',
        type: params.type,
        locale: 'en',
        stylesheet: 'https://cdn.fynbos.app/machnet/widget_css.css',
        token: widgetToken
      })
      widget.init()

      window.addEventListener('message', listener)
    }
  }, [widgetToken, widgetUserId, listener, scriptStatus, params.type])

  return (
    <div className='flex w-full flex-col rounded-2xl bg-page p-4 pb-8'>
      {params.type == 'card' && (
        <>
          <div className='flex justify-between'>
            <h1 className='font-display text-2xl font-medium'>Debit card</h1>
            <div className='hidden sm:flex'>
              <Shape
                width={'w-8'}
                radius={'rounded-tl-full'}
                color={'bg-lime-500'}
              />
              <Shape
                width={'w-8'}
                radius={'rounded-tl-full'}
                color={'bg-slate-600'}
              />
            </div>
          </div>
          <p className='mt-6 text-medium'>
            Please provide your debit card details.
          </p>
        </>
      )}
      {params.type == 'bank' && (
        <>
          <div className='flex justify-between'>
            <h1 className='font-display text-2xl font-medium'>Bank details</h1>
            <div className='hidden sm:flex'>
              <Shape
                width={'w-8'}
                radius={'rounded-tr-full'}
                color={'bg-slate-500'}
              />
              <Shape
                width={'w-8'}
                radius={'rounded-br-full'}
                color={'bg-yellow-300'}
              />
            </div>
          </div>
          <p className='mt-6 text-medium'>
            Please provide your bank account details.
          </p>
        </>
      )}
      <div id='widget' className='-mx-4 mt-6 w-[100vw-2rem]' />
    </div>
  )
}
