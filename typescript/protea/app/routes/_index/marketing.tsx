import { useLoaderData } from '@remix-run/react'
import { MarketingPageWithSections } from '~/components/Content'

import { route } from 'routes-gen'
import { Router } from '~/components'
import type { SectionRecord } from '~/generated/dato-cms-graphql'
import type { marketingLoader } from './route'

export function MarketingPage() {
  const { homeRoute } = useLoaderData<typeof marketingLoader>()
  return (
    <>
      <div className='mx-auto mb-5 mt-1 flex items-center gap-x-2 rounded-lg bg-mk-section p-1 px-2 lg:px-1'>
        <div className='rounded bg-primary px-3 py-1 text-xs font-medium text-on-color'>
          New
        </div>
        <div className='flex flex-col lg:contents'>
          <span className='text-xs text-medium'>
            Refer a friend using Fynbos and earn $20!
          </span>
          <Router
            className='mr-1 text-xs font-medium text-primary'
            to={route('/referral')}
          >
            Find out more
          </Router>
        </div>
      </div>
      <div>
        {homeRoute?.body.map((section) => (
          <MarketingPageWithSections
            key={section.id}
            section={section as SectionRecord}
          />
        ))}
      </div>
    </>
  )
}
