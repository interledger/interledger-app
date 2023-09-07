import type { LoaderArgs } from '@remix-run/node'

import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { toRemixMeta } from 'react-datocms'
import type { ApplicationProps } from '~/components'
import { Layouts } from '~/components'
import { MarketingPageWithSections } from '~/components/Content'
import type { SectionRecord } from '~/generated/dato-cms-graphql'
import { getCurrentThankYouPage } from '~/lib/marketing.server'

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

export function meta({ data, params }: any) {
  return {
    ...toRemixMeta(data.thankYou.seoMeta),
    'twitter:url': 'https://fynbos.app/thank-you/' + params.slug,
    'og:url': 'https://fynbos.app/thank-you/' + params.slug
  }
}

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
