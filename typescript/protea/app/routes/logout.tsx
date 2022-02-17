import { LoaderFunction, redirect, useLoaderData } from 'remix'
import { json } from 'remix'
import { Logo, Router } from '~/components'
import { AxiosError } from 'axios'
import React from 'react'
import { route } from 'routes-gen'
import { kratos } from '~/lib/kratos'

export const loader: LoaderFunction = async ({ request }) => {
  return kratos
    .createSelfServiceLogoutFlowUrlForBrowsers(
      request.headers.get('cookie') ?? undefined
    )
    .then(({ data }) => json(data))
    .catch((err: AxiosError) => {
      switch (err.response?.status) {
        case 401:
          // do nothing, the user is not logged in
          return redirect(route('/login'))
      }

      // Something else happened!
      return Promise.reject(err)
    })
}

export default function LoginPage() {
  const loaderData = useLoaderData()

  return (
    <main className='mx-auto flex h-screen max-w-sm flex-col items-start justify-center px-4'>
      <Router to={route('/')} aria-label='Fynbos logo'>
        <Logo className='h-8' />
      </Router>
      <h1 className='mt-6 mb-1 font-display text-4xl font-medium leading-normal text-strong'>
        Logout of your account
      </h1>
      <p className='mb-10 text-medium'>Are you sure you want to logout?</p>
      {/* className={`focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus focus-visible:ring-offset-1 ${className}`} */}
      <div className='flex min-w-full flex-col items-end'>
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
