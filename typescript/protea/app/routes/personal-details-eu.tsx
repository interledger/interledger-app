import type { LoaderFunctionArgs, MetaFunction } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { useLoaderData, useSearchParams } from '@remix-run/react'
import { route } from 'routes-gen'
import { Layouts } from '~/components'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { mergeMeta } from '~/lib/meta'

export async function loader({ request }: LoaderFunctionArgs) {
  if (process.env.FYNBOS_ENV === 'prod') {
    redirect('/')
  }

  let response = await grpc.getGatehubOnboardingWidget(request, {})
  if (isConnectError(response)) throw response.errorResponse

  return json({
    widget: response.widgetUrl
  })
}

export const handle = {
  layout: Layouts.Wallet,
  scaffold: {
    header: {
      title: 'Activate wallet',
      back: route('/')
    }
  }
}

export const meta: MetaFunction = mergeMeta(() => [
  {
    title: 'Activate wallet'
  }
])

export default function Page() {
  const [params] = useSearchParams()
  const { widget } = useLoaderData<typeof loader>()

  console.log('widgetURL', widget)

  return (
    <>
      <iframe
        src={widget}
        sandbox='allow-top-navigation allow-forms allow-same-origin allow-popups allow-scripts'
        scrolling='no'
        frameBorder='0'
        allow='camera;microphone'
        className='w-full'
      ></iframe>
    </>
  )
}
