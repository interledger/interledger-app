import { useLoaderData } from '@remix-run/react'
import { MarketingPageWithSections } from '~/components/Content'

import type { SectionRecord } from '~/generated/dato-cms-graphql'
import type { marketingLoader } from './route'

export function MarketingPage() {
  const { homeRoute } = useLoaderData<typeof marketingLoader>()
  return (
    <>
      {homeRoute?.body.map((section) => (
        <MarketingPageWithSections
          key={section.id}
          section={section as SectionRecord}
        />
      ))}
    </>
  )
}
