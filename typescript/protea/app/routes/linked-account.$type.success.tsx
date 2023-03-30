import { ButtonRouter, Card, Layouts, SuccessShapes } from '~/components'
import { route } from 'routes-gen'
import { useLoaderData, useParams } from '@remix-run/react'
import type { LoaderArgs, MetaFunction } from '@remix-run/node'
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
  title: 'Success',
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
    <>
      <Card>
        <SuccessShapes />

        <span className='mt-6 text-medium'>
          Your {params.type == 'card' ? 'debit card' : 'bank account'} has been
          added.
        </span>
      </Card>
      <ButtonRouter to={flow.returnTo ?? route('/settings/linked-accounts')}>
        Close
      </ButtonRouter>
    </>
  )
}
