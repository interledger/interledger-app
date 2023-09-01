import type { LoaderArgs } from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { toRemixMeta } from 'react-datocms'
import type { ApplicationProps } from '~/components'
import { Layouts } from '~/components'
import { MarketingPageWithSections } from '~/components/Content'
import type { SectionRecord } from '~/generated/dato-cms-graphql'
import { getSlackRoute } from '~/lib/marketing.server'

export async function loader({ request }: LoaderArgs) {
  const { slackRoute, footer } = await getSlackRoute()
  return json({ slackRoute, footer })
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
    ...toRemixMeta(data.slackRoute.seoMeta),
    'twitter:url': 'https://fynbos.app/slack',
    'og:url': 'https://fynbos.app/slack'
  }
}

export default function Page() {
  const { slackRoute } = useLoaderData<typeof loader>()
  return (
    <>
      {slackRoute?.body.map((section) => (
        <MarketingPageWithSections
          key={section.id}
          section={section as SectionRecord}
        />
      ))}
    </>
  )
}
