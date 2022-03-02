import { json, redirect, useLoaderData } from 'remix'
import type { LoaderFunction } from 'remix'
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
    <main className='mx-auto grid min-h-screen w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-y-auto p-4 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 lg:content-center xl:max-w-4xl'>
      <div className='col-span-full sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <Router to={route('/')}>
          <Logo className='h-8' />
        </Router>
      </div>
      <div className='col-span-full pt-4 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <h1 className='font-display text-4xl font-medium leading-normal text-strong'>
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
