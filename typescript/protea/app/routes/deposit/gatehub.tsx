import type { LoaderFunctionArgs } from '@remix-run/node'
import { useLoaderData, useNavigate } from '@remix-run/react'
import { useEffect } from 'react'
import { jsonWithCSRF } from '~/lib/csrf.server'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { useScaffoldStore } from '~/lib/useScaffoldStore'

export async function gatehubDepositLoader({ request }: LoaderFunctionArgs) {
  const widgetResponse = await grpc.getGatehubDepositWidget(request, {})
  if (isConnectError(widgetResponse)) throw widgetResponse.error

  return jsonWithCSRF(request, {
    provider: 'gatehub',
    gatehubWidgetUrl: widgetResponse.widgetUrl
  })
}

export function GatehubDepositPage() {
  const { gatehubWidgetUrl } = useLoaderData<typeof gatehubDepositLoader>()
  const [pushSnackbar] = useScaffoldStore((state) => [state.pushSnackbar])
  const navigate = useNavigate()
  useEffect(() => {
    console.log('registering message event handler')
    const url = new URL(gatehubWidgetUrl)

    const handler = (event: MessageEvent) => {
      if (
        event.origin === url.origin &&
        event.data.type === 'StripeDepositCompleted'
      ) {
        pushSnackbar({
          id: 'deposit-success',
          message: 'Deposit submitted successfully, it may take a few minutes to reflect in your account',
          icon: 'close',
          canShow: true
        })
        navigate('/')
      }
    }

    window.addEventListener('message', handler)

    return () => {
      window.removeEventListener('message', handler)
    }
  }, [gatehubWidgetUrl, navigate, pushSnackbar])


  return (
    <iframe
      title='Deposit'
      src={gatehubWidgetUrl}
      sandbox='allow-top-navigation allow-forms allow-same-origin allow-popups allow-scripts'
      scrolling='no'
      frameBorder='0'
      className='h-[750px]'
    />
  )
}
