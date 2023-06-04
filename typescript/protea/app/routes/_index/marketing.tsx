import { useLoaderData } from '@remix-run/react'
import type { ApplicationProps } from '~/components'
import { Layouts, MarketingPageWithSections } from '~/components'

import type { SectionRecord } from '~/generated/dato-cms-graphql'
import type { loader } from './route'

export const handle: ApplicationProps = {
  layout: Layouts.Marketing,
  scaffold: {
    header: {},
    footer: (match) => match.data.footer
  }
}

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
