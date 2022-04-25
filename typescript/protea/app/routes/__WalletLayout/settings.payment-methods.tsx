import React from 'react'
import type { LoaderFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import { Link, useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import { Icon, Router } from '~/components'
import type {
  GetFundingSourcesQuery,
  GetFundingSourcesQueryVariables
} from '~/generated/types'
import { GetFundingSourcesDocument } from '~/generated/types'
import { apolloClient } from '~/lib/apollo.server'

type PaymentMethod = {
  id: string
  name: string
  description: string
  icon: string
}

export const loader: LoaderFunction = async ({ request }) => {
  const cookie = String(request.headers.get('cookie'))

  const res = await apolloClient.query<
    GetFundingSourcesQuery,
    GetFundingSourcesQueryVariables
  >({
    query: GetFundingSourcesDocument,
    context: {
      headers: {
        cookie: cookie
      }
    }
  })
  const paymentMethods = res.data.fundingSources.map((fs) => ({
    id: fs?.id,
    name: fs?.name,
    description: fs?.mask,
    icon: 'credit_card' // TODO: get actual icon from fundingsource subtype
  }))

  return json({
    paymentMethods
  })
}

export default function SettingsPaymentMethodsPage() {
  const { paymentMethods } = useLoaderData()

  return (
    <div className='w-full'>
      {/* Header */}
      <header className='sticky top-0 mx-auto flex h-16 w-full select-none items-center justify-start bg-white p-4 text-medium sm:max-w-lg lg:max-w-3xl xl:max-w-4xl'>
        <Link to={route('/settings')}>
          <div className='-ml-3 p-3 text-medium'>
            <Icon>arrow_back</Icon>
          </div>
        </Link>
        <div className='flex items-center justify-start font-display text-2xl font-medium'>
          Payment Methods
        </div>
      </header>
      {/* Body */}
      <div className='mx-auto grid min-h-[calc(100vh-9rem)] w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-y-auto p-4 pb-24 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 xl:max-w-4xl'>
        {paymentMethods.length == 0 && (
          <div className='col-span-full flex items-center justify-between space-x-3 rounded-xl bg-container p-3 text-medium sm:col-span-6 sm:col-start-2 lg:col-start-4'>
            <Icon>tips_and_updates</Icon>
            <span className='font-sans text-sm font-normal'>
              You need to add a payment method before you can deposit money.
            </span>
          </div>
        )}
        {paymentMethods.length > 0 &&
          paymentMethods.map((method: PaymentMethod) => (
            <div
              key={method.id}
              className='col-span-full flex items-center justify-between rounded-xl bg-container px-4 py-2 sm:col-span-6 sm:col-start-2 lg:col-start-4'
            >
              <div className='flex items-center space-x-3 text-medium'>
                {method.icon && <Icon>{method.icon}</Icon>}
                <div className='flex flex-col'>
                  <span className='font-sans text-base font-normal'>
                    {method.name}
                  </span>
                  <span className='font-sans text-xs font-normal text-weak'>
                    {method.description}
                  </span>
                </div>
              </div>
            </div>
          ))}
        <Router
          to={route('/flows/:flowId/payment-method/type', {
            flowId: 'init'
          })}
          className='col-span-full mt-2 flex items-center justify-between rounded-xl bg-container p-3 text-medium sm:col-span-6 sm:col-start-2 lg:col-start-4'
        >
          <span className='font-sans text-base font-normal'>
            New payment method
          </span>
          <Icon>navigate_next</Icon>
        </Router>
      </div>
    </div>
  )
}
