import { json } from '@remix-run/node'
import type { ApplicationProps } from '~/components'
import { Layouts } from '~/components'

import type { LoaderFunctionArgs } from '@remix-run/node'
import type { UIMatch } from '@remix-run/react'
import { useLoaderData } from '@remix-run/react'
import { MarketingPageWithSections } from '~/components/Content'
import { getCurrentMarketingPage } from '~/data/content.server'
import type { SectionRecord } from '~/generated/dato-cms-graphql'

export async function loader({ params }: LoaderFunctionArgs) {
  const { marketingPage, footer } = await getCurrentMarketingPage({
    filter: {
      slug: { eq: params.slug }
    }
  })

  if (marketingPage === null)
    throw json(null, { status: 404, statusText: 'Not Found' })

  return json({
    footer,
    marketingPage
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Marketing,
  scaffold: {
    header: {},
    footer: (match: UIMatch<typeof loader>) => match.data.footer as any
  }
}

export default function Page() {
  const { marketingPage } = useLoaderData<typeof loader>()
  return (
    <>
      {marketingPage?.body?.map((section) => (
        <MarketingPageWithSections
          key={section.id}
          section={section as SectionRecord}
        />
      ))}
    </>
  )
}
