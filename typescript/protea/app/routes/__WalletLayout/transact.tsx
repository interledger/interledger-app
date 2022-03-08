import React from 'react'
import { LoaderFunction, NavLink } from 'remix'
import { route } from 'routes-gen'
import { SettingsIcon } from '~/components'
import { requireUserSession } from '~/lib/kratos.server'

export const loader: LoaderFunction = async ({ request }) => {
  return requireUserSession(request)
}

export default function TransactPage() {
  return (
    <>
      {/* Header */}
      <header className='sticky top-0 flex h-16 min-w-full select-none items-center justify-between bg-white p-4 text-medium'>
        <div className='flex items-center justify-start font-display text-2xl font-medium'>
          Home
        </div>
        <NavLink className='sm:hidden' to={route('/home')}>
          <div className='-mr-3 p-3 text-medium'>
            <SettingsIcon />
          </div>
        </NavLink>
      </header>
      {/* Body */}
      <div className='mx-auto grid min-h-[calc(100vh-9rem)] w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-y-auto p-4 pb-24 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 xl:max-w-4xl'>
        Body
      </div>
    </>
  )
}
