import { useState } from 'react'
import type { LoaderArgs } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import { Icon, Router, Snackbar } from '~/components'
import { requireUserSession } from '~/lib/kratos.server'
import { getSession, commitSession } from '~/sessions'

const shapes = [
  [
    'bg-slate-600 rounded-tl-full',
    'bg-transparent',
    'bg-yellow-400 rounded-tr-full',
    'bg-rose-300 rounded-tl-full',
    'bg-lime-400 rounded-full',
    'bg-transparent',
    'bg-rose-500 rounded-full',
    'bg-lime-300 rounded-tr-full',
    'bg-transparent',
    'bg-transparent'
  ],
  [
    'bg-transparent',
    'bg-rose-400 rounded-full',
    'bg-lime-500 rounded-bl-full',
    'bg-transparent',
    'bg-slate-300 rounded-tl-full',
    'bg-yellow-200 rounded-tl-full',
    'bg-slate-500 rounded-br-full',
    'bg-transparent',
    'bg-rose-100 rounded-full',
    'bg-rose-300 rounded-bl-full'
  ]
]

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
      <div className='grid w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-y-auto py-4 px-4 sm:grid-cols-8 sm:px-0 lg:grid-cols-12'>
        <div className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2 lg:col-start-4'>
          {shapes.map((shapeRow) => (
            <div className='flex' key={shapeRow.toString()}>
              {shapeRow.map((shape, index) => (
                <div
                  key={shape + index}
                  className={`aspect-square w-full ${shape}`}
                />
              ))}
            </div>
          ))}
        </div>
        <span className='col-span-full mt-4 font-display text-2xl font-medium sm:col-span-6 sm:col-start-2 lg:col-start-4'>
          Settings
        </span>
        <span className='col-span-full mt-3 font-display text-lg font-medium sm:col-span-6 sm:col-start-2 lg:col-start-4'>
          Profile
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
