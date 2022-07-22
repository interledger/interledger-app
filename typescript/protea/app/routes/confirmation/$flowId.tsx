import type { LoaderArgs } from '@remix-run/node'
import { Outlet } from '@remix-run/react'
import { route } from 'routes-gen'
import { Logo, Router } from '~/components'
import { requireUserSession } from '~/lib/kratos.server'

export async function loader({ request, params }: LoaderArgs) {
  return requireUserSession(request)
}

export default function Page() {
  return (
    <div className='mx-auto grid min-h-screen w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-y-auto p-4 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 lg:content-center xl:max-w-4xl'>
      <div className='col-span-full sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <Router to={route('/')}>
          <Logo className='h-8' />
        </Router>
      </div>
      <Outlet />
    </div>
  )
}
