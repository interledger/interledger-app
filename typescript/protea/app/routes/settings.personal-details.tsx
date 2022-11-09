import { useState } from 'react'
import type { LoaderArgs } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import { Icon, Layouts, Router, Snackbar, WalletGrid } from '~/components'
import { requireUserSession } from '~/lib/kratos.server'
import { getSnackbar } from '~/lib/snackbar.server'

export async function loader({ request }: LoaderArgs) {
  const url = new URL(request.url)

  const flowId = url.searchParams.get('flow')
  if (flowId) return redirect(`${route('/recovery/password')}?flow=${flowId}`)

  const session = await requireUserSession(request)

  const snackbar = await getSnackbar(request)

  return json({
    traits: session.identity.traits,
    snackbar
  })
}

export const handle = {
  layout: Layouts.WalletLayout
}

export default function Page() {
  const { traits, snackbar } = useLoaderData<typeof loader>()
  const [showSnackbar, setSnackbar] = useState<boolean>(snackbar.show ?? false)
  return (
    <WalletGrid>
      <div className='col-span-full flex flex-col rounded-2xl bg-page p-4 pb-8 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <h1 className='font-display text-2xl font-medium'>Personal details</h1>
        <Router
          to={route('/settings/personal-details')}
          className='mt-6 flex items-center justify-between rounded-xl bg-container p-3 text-medium hover:bg-container-hover'
        >
          <div className='flex space-x-3'>
            <Icon>face</Icon>
            <span>
              {traits.firstName} {traits.lastName}
            </span>
          </div>
          <Icon>navigate_next</Icon>
        </Router>
        <Router
          to={route('/settings/linked-accounts')}
          className='mt-2 flex items-center justify-between rounded-xl bg-container p-3 text-medium hover:bg-container-hover'
        >
          <div className='flex space-x-3'>
            <Icon>mail</Icon>
            <span>{traits.email}</span>
          </div>
          <Icon>navigate_next</Icon>
        </Router>
        <Router
          to={route('/settings/linked-accounts')}
          className='mt-2 flex items-center justify-between rounded-xl bg-container p-3 text-medium hover:bg-container-hover'
        >
          <div className='flex space-x-3'>
            <Icon>call</Icon>
            <span>{traits.phone}</span>
          </div>
          <Icon>navigate_next</Icon>
        </Router>
        <h2 className='mt-6 text-sm font-medium'>Country of residence</h2>
        <Router
          to='/login/challenge?challenge-flow=settings-password'
          className='mt-2 flex items-center justify-between rounded-xl bg-container p-3 text-medium hover:bg-container-hover'
        >
          <div className='flex space-x-3'>
            <Icon>password</Icon>
            <span>Password</span>
          </div>
          <Icon>navigate_next</Icon>
        </Router>
        <Router
          to={route('/logout')}
          className='mt-2 flex items-center space-x-3 rounded-xl p-3 text-primary'
        >
          <Icon>logout</Icon>
          <span>Log out</span>
        </Router>
      </div>
      <Snackbar
        message={snackbar.message}
        action={snackbar.action}
        icon={snackbar.icon}
        show={showSnackbar}
        id='cookie-snackbar'
        dismissAfter={3000}
        offset
        onClose={() => setSnackbar(false)}
      />
    </WalletGrid>
  )
}
