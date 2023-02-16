import { ButtonRouter, Layouts, SuccessShapes } from '~/components'
import { route } from 'routes-gen'
import { useLoaderData, useParams } from '@remix-run/react'
import type { LoaderArgs , MetaFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import { flowType, requireFlow } from '~/lib/flows.server'

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

  return json({ flow })
}
export const handle = {
  layout: Layouts.FocusLayout
}

export const meta: MetaFunction = () => {
  return {
    title: 'Add linked account | Success'
  }
}

export default function Page() {
  const params = useParams()
  const { flow } = useLoaderData<typeof loader>()
  return (
    <div className='flex w-full flex-col rounded-2xl bg-page p-4 pb-8'>
      <SuccessShapes />

      <span className='mt-6 font-display text-2xl font-medium'>Success</span>
      <span className='mt-6 text-medium'>
        Your {params.type == 'card' ? 'debit card' : 'bank account'} has been
        added.
      </span>

      <div className='flex justify-end pt-12'>
        <ButtonRouter to={flow.returnTo ?? route('/settings/linked-accounts')}>
          Close
        </ButtonRouter>
      </div>
    </div>
  )
}
