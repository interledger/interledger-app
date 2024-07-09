import type { LoaderFunctionArgs } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { jsonWithCSRF } from '~/lib/csrf.server'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'

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
