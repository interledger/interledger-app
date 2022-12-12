import { HomeShapes, Layouts } from '~/components'
import { useFetcher, useParams } from '@remix-run/react'
import type { LoaderArgs } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { getLinkedAccounts } from '~/lib/wallet.server'
import { flowType, requireFlow } from '~/lib/flows.server'
import { useEffect } from 'react'
import { route } from 'routes-gen'

export async function loader({ request, params }: LoaderArgs) {
  let flow
  if (params.type == 'card') {
    flow = await requireFlow(request, flowType.LinkCardAccount)
  } else if (params.type == 'bank') {
    flow = await requireFlow(request, flowType.LinkBankAccount)
  } else {
    throw json(
      { title: `Linking type ${params.type} not allowed.` },
      { status: 400 }
    )
  }
  const linkedAccounts = await getLinkedAccounts(request)

  if (linkedAccounts.linkedAccounts.length > flow.data.linkedAccountLength)
    return redirect(
      route('/linked-account/:type/success', { type: params.type })
    )

  return json({})
}

export const handle = {
  layout: Layouts.FocusLayout
}

export default function Page() {
  const params = useParams()
  const fetcher = useFetcher()

  // Revalidate every 1s
  useEffect(() => {
    const interval: NodeJS.Timeout = setInterval(() => {
      if (document.visibilityState === 'visible') {
        fetcher.load(
          route('/linked-account/:type/almost-there', {
            type: params.type as string
          })
        )
      }
    }, 1000)

    return () => clearInterval(interval)
  }, [fetcher, params.type])

  return (
    <div className='flex w-full flex-col rounded-2xl bg-page p-4 pb-8'>
      <HomeShapes animate />

      <span className='mt-6 font-display text-2xl font-medium'>
        Almost there
      </span>
      <span className='mt-6 text-medium'>
        Please wait a moment while we verify a few final details. You will be
        redirected shortly.
      </span>
    </div>
  )
}
