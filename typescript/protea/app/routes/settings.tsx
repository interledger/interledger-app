import { useState } from 'react'
import type { LoaderArgs } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Link, useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import { Icon, Router, Snackbar } from '~/components'
import { requireUserSession } from '~/lib/kratos.server'
import { getSession, commitSession } from '~/sessions'

export async function loader({ request }: LoaderArgs) {
  const userSettings = await getSession(request.headers.get('Cookie'))
  const url = new URL(request.url)
  const flowId = url.searchParams.get('flow')
  if (flowId)
    return redirect(`${route('/recovery/password')}?flow=${flowId}`, {
      headers: request.headers
    })
  const session = await requireUserSession(request)
  const snackbar = {
    // NOTE: userSettings.has must be called before userSettings.get
    show: userSettings.has('snackbar'),
    ...userSettings.get('snackbar')
  }

  return json(
    {
      session,
      snackbar: snackbar
    },
    {
      headers: { 'Set-Cookie': await commitSession(userSettings) }
    }
  )
}

export default function Page() {
  const { session, snackbar } = useLoaderData<typeof loader>()
  const [showSnackbar, setSnackbar] = useState<boolean>(snackbar.show)
  return (
    <div className='w-full'>
      <Snackbar
        message={snackbar.message}
        action={snackbar.action}
        show={showSnackbar}
        id='cookie-snackbar'
        onClose={() => setSnackbar(false)}
      />
      {/* Header */}
      <header className='sticky top-0 mx-auto flex h-16 w-full select-none items-center justify-start bg-app p-4 text-medium sm:min-w-full'>
        <Link className='sm:hidden' to={route('/')}>
          <div className='-ml-3 p-3 text-medium'>
            <Icon>arrow_back</Icon>
          </div>
        </Link>
        <div className='flex items-center justify-start font-display text-2xl font-medium'>
          Settings
        </div>
      </header>
      {/* Body */}
      <div className='mx-auto grid min-h-[calc(100vh-9rem)] w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-y-auto p-4 pb-24 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 xl:max-w-4xl'>
        {/* Activity item */}
        <span className='col-span-full font-display text-lg font-medium sm:col-span-6 sm:col-start-2 lg:col-start-4'>
          Personal info
        </span>
        <div className='col-span-full flex items-center justify-between rounded-xl bg-container p-3 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
          <div className='flex space-x-3'>
            <Icon>mail</Icon>
            <span>{session?.identity.traits.email}</span>
          </div>
        </div>
        <div className='col-span-full flex items-center justify-between rounded-xl bg-container p-3 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
          <div className='flex space-x-3'>
            <Icon>flag</Icon>
            {/* TODO: show actual country here rather */}
            <span>Country</span>
          </div>
        </div>
        <span className='col-span-full font-display text-lg font-medium sm:col-span-6 sm:col-start-2 lg:col-start-4'>
          Payments
        </span>
        <Router
          to='/settings/linked-accounts'
          className={`col-span-full flex items-center justify-between rounded-xl bg-container p-3 sm:col-span-6 sm:col-start-2 lg:col-start-4`}
        >
          <div className='flex space-x-3'>
            <Icon>credit_card</Icon>
            <span>Linked accounts</span>
          </div>
          <Icon>navigate_next</Icon>
        </Router>
        <Router
          to='/settings/statements'
          className={`col-span-full flex items-center justify-between rounded-xl bg-container p-3 sm:col-span-6 sm:col-start-2 lg:col-start-4`}
        >
          <div className='flex space-x-3'>
            <Icon>folder</Icon>
            <span className='font-sans text-base font-normal'>Statements</span>
          </div>
          <Icon>navigate_next</Icon>
        </Router>
        <span className='col-span-full font-display text-lg font-medium sm:col-span-6 sm:col-start-2 lg:col-start-4'>
          Security
        </span>
        <Router
          to='/login/challenge?challenge-flow=settings-password'
          className='col-span-full flex items-center justify-between rounded-xl bg-container p-3 sm:col-span-6 sm:col-start-2 lg:col-start-4'
        >
          <div className='flex space-x-3'>
            <Icon>password</Icon>
            <span>Password</span>
          </div>
          <Icon>navigate_next</Icon>
        </Router>
        <Router
          to='/logout'
          className='col-span-full mt-6 flex items-center justify-between rounded-xl bg-container p-3 sm:col-span-6 sm:col-start-2 lg:col-start-4'
        >
          <div className='flex space-x-3'>
            <Icon>logout</Icon>
            <span>Logout</span>
          </div>
          <Icon>navigate_next</Icon>
        </Router>
      </div>
    </div>
  )
}
