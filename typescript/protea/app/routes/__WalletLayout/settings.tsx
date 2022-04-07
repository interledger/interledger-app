import React, { useState } from 'react'
import { Link, LoaderFunction, redirect, useLoaderData, json } from 'remix'
import { route } from 'routes-gen'
import { BackIcon, CardIcon, NextIcon, Router, Snackbar } from '~/components'
import { requireUserSession } from '~/lib/kratos.server'
import { getSession, commitSession } from '~/sessions'

export const loader: LoaderFunction = async ({ request }) => {
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

export default function SettingsPage() {
  const { session, snackbar } = useLoaderData()
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
      <header className='sticky top-0 mx-auto flex h-16 w-full select-none items-center justify-start bg-white p-4 text-medium sm:min-w-full'>
        <Link className='sm:hidden' to={route('/home')}>
          <div className='-ml-3 p-3 text-medium'>
            <BackIcon />
          </div>
        </Link>
        <div className='flex items-center justify-start font-display text-2xl font-medium'>
          Settings
        </div>
      </header>
      {/* Body */}
      <div className='mx-auto grid min-h-[calc(100vh-9rem)] w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-y-auto p-4 pb-24 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 xl:max-w-4xl'>
        {/* Activity item */}
        <span className='col-span-full ml-4 font-display text-lg font-medium sm:col-span-6 sm:col-start-2 lg:col-start-4'>
          Personal info
        </span>
        <div className='col-span-full flex items-center justify-between rounded-xl bg-container px-4 py-2 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
          <div className='flex flex-col'>
            <span className='font-display text-xs font-medium'>Email</span>
            <span className='font-sans text-base font-normal'>
              {session?.identity.traits.email}
            </span>
          </div>
        </div>
        <div className='col-span-full flex items-center justify-between rounded-xl bg-container px-4 py-2 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
          <div className='flex flex-col'>
            <span className='font-display text-xs font-medium'>Country</span>
            <span className='font-sans text-base font-normal'>TODO</span>
          </div>
        </div>
        <Router
          to='/settings/payment-methods'
          className={`col-span-full flex items-center justify-between rounded-xl bg-container p-3 sm:col-span-6 sm:col-start-2 lg:col-start-4`}
        >
          <div className='flex space-x-3'>
            <CardIcon />
            <span className='font-sans text-base font-normal'>
              Payment methods
            </span>
          </div>
          <NextIcon />
        </Router>
        <span className='col-span-full ml-4 font-display text-lg font-medium sm:col-span-6 sm:col-start-2 lg:col-start-4'>
          Security
        </span>
        <Router
          to='/login/challenge?challenge-flow=settings-password'
          className='col-span-full flex items-center justify-between rounded-xl bg-container px-4 py-2 sm:col-span-6 sm:col-start-2 lg:col-start-4'
        >
          <div className='flex flex-col'>
            <span className='font-display text-xs font-medium'>Password</span>
            <span className='font-sans text-base font-normal'>
              &#9679;&#9679;&#9679;&#9679;&#9679;&#9679;&#9679;&#9679;
            </span>
          </div>
          <NextIcon />
        </Router>
      </div>
    </div>
  )
}
