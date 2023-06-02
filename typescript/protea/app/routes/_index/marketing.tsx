import { useLoaderData } from '@remix-run/react'
import type { ApplicationProps } from '~/components'
import { Layouts, renderMarketingPageWithSections } from '~/components'

import type { SectionRecord } from '~/generated/dato-cms-graphql'
import type { loader } from './route'

export const handle: ApplicationProps = {
  title: 'Fynbos',
  layout: Layouts.Marketing,
  scaffold: {
    header: {},
    footer: (match) => match.data.footer
  }
}

export function MarketingPage() {
  const { homepage } = useLoaderData<typeof loader>()
  return (
    <>
      {homepage?.body.map((section) =>
        renderMarketingPageWithSections(section as SectionRecord)
      )}
    </>
  )
}
