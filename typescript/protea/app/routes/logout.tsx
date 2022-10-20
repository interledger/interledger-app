import type { LoaderArgs } from '@remix-run/node'
import { json } from '@remix-run/node'
import { Form, useLoaderData } from '@remix-run/react'
import { AnchorButtonRouter, Button, Logo, Router, Shape } from '~/components'
import { route } from 'routes-gen'
import { KRATOS_URL, handleFlowError } from '~/lib/kratos.server'

export async function loader({ request }: LoaderArgs) {
  const cookie = String(request.headers.get('cookie'))
  let flow
  const flowRes = await fetch(`${KRATOS_URL}/self-service/logout/browser`, {
    headers: {
      cookie: cookie,
      Accept: 'application/json'
    }
  })
  flow = await flowRes.json()
  if (flowRes.status >= 400) handleFlowError(flow, 'logout')
  return json({ logoutUrl: flow.logout_url })
}

export default function Page() {
  const { logoutUrl } = useLoaderData<typeof loader>()

  return (
    <div className='flex w-full flex-col rounded-2xl bg-page p-4 pb-8'>
      <h1 className='mb-6 font-display text-2xl font-medium'>Log out</h1>
      <span>Are you sure you want to log out?</span>

      <div className='mt-6'>
        <AnchorButtonRouter to={logoutUrl}>Log out</AnchorButtonRouter>
      </div>
    </div>
  )
}
