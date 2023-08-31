import { useLoaderData } from '@remix-run/react'
import { MarketingPageWithSections } from '~/components/Content'

import type { SectionRecord } from '~/generated/dato-cms-graphql'
import type { loader } from './route'

export function MarketingPage() {
  const { homeRoute } = useLoaderData<typeof loader>()
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
