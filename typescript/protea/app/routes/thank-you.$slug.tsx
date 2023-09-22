import type { LoaderArgs, V2_MetaFunction } from '@remix-run/node'

import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import type { ApplicationProps } from '~/components'
import { Layouts } from '~/components'
import { MarketingPageWithSections } from '~/components/Content'
import { getCurrentThankYouPage } from '~/data/content.server'
import type { SectionRecord } from '~/generated/dato-cms-graphql'
import { datoMeta, mergeMeta } from '~/lib/meta'

export async function loader({ request, params }: LoaderArgs) {
  const { thankYou, footer } = await getCurrentThankYouPage({
    filter: { slug: { eq: params.slug } }
  })
  return json({ thankYou, footer })
}

export const handle: ApplicationProps = {
  layout: Layouts.Marketing,
  scaffold: {
    header: {},
    footer: (match) => match.data.footer
  }
}

export const meta: V2_MetaFunction<typeof loader> = mergeMeta(
  ({ data }) => datoMeta(data?.thankYou?.seoMeta),
  ({ location }) => [
    {
      name: 'og:url',
      content: `https://fynbos.app${location.pathname}`
    },
    {
      name: 'twitter:url',
      content: `https://fynbos.app${location.pathname}`
    }
  ]
)

export default function Page() {
  const { thankYou } = useLoaderData<typeof loader>()
  return (
    <>
      {thankYou?.body.map((section) => (
        <MarketingPageWithSections
          key={section.id}
          section={section as SectionRecord}
        />
      ))}
    </>
  )
}
