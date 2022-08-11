import type { LoaderArgs } from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { Logo, Router } from '~/components'
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
  return json(flow)
}

export default function Page() {
  const loaderData = useLoaderData<typeof loader>()

  return (
    <main className='mx-auto grid min-h-screen w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-y-auto p-4 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 lg:content-center xl:max-w-4xl'>
      <div className='col-span-full sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <Router to={route('/')}>
          <Logo className='h-8' />
        </Router>
      </div>
      <div className='col-span-full pt-4 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <h1 className='font-display text-4xl font-medium leading-normal'>
          Logout of your account
        </h1>
      </div>
      <div className='col-span-full pb-8 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <p className='text-medium'>Are you sure you want to logout?</p>
      </div>
      <div className='col-span-full flex justify-end pt-4 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        {/* TODO allow Router to handle external href */}
        <a
          href={loaderData.logout_url}
          className='flex h-10 cursor-pointer items-center rounded-full bg-container-primary px-6 font-display text-sm font-medium text-medium hover:bg-container-primary-hover focus:outline-none focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus active:bg-container-primary-active'
        >
          Logout
        </a>
      </div>
    </main>
  )
}
