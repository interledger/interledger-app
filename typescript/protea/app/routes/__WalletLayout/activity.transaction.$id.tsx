import React from 'react'
import { Link, LoaderFunction, useNavigate } from 'remix'
import { route } from 'routes-gen'
import { BackIcon, ReceivedIcon } from '~/components'
import { requireUserSession } from '~/lib/kratos'

export const loader: LoaderFunction = async ({ request, params }) => {
  // params.id should be used to fetch a specific transaction
  console.log(params)
  return requireUserSession(request)
}

export default function ActivityTransactionPage() {
  const navigate = useNavigate()
  return (
    <div className='w-full'>
      {/* Header */}
      <header className='sticky top-0 mx-auto flex h-16 w-full select-none items-center justify-start bg-white p-4 text-medium sm:max-w-lg lg:max-w-3xl xl:max-w-4xl'>
        <button
          onClick={() => {
            navigate(-1)
          }}
        >
          <div className='-ml-3 p-3 text-medium'>
            <BackIcon />
          </div>
        </button>
        <div className='flex items-center justify-start font-display text-2xl font-medium'>
          Transaction
        </div>
      </header>
      {/* Body */}
      <div className='mx-auto grid min-h-[calc(100vh-9rem)] w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-y-auto p-4 pb-24 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 xl:max-w-4xl'>
        {/* Activity item */}
        {/* TODO Form */}
        <div className='col-span-full flex h-12 items-center justify-between rounded-xl bg-container px-3 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
          <div className='flex items-center justify-between space-x-3'>
            <ReceivedIcon />
            <div className='flex flex-col'>
              <span className='font-display text-base font-medium'>
                Received
              </span>
              <span className='font-sans text-xs font-normal'>
                from Interledger
              </span>
            </div>
          </div>
          <span className='font-sans text-lg font-normal'>$ 1.00</span>
        </div>
      </div>
    </div>
  )
}
