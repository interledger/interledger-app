import type { LoaderArgs } from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { toRemixMeta } from 'react-datocms'
import type { ApplicationProps } from '~/components'
import { Layouts } from '~/components'
import { MarketingPageWithSections } from '~/components/Content'
import { getCollectablesRoute } from '~/data/content.server'
import type { SectionRecord } from '~/generated/dato-cms-graphql'

export async function loader({ request }: LoaderArgs) {
  const { collectablesRoute, footer } = await getCollectablesRoute()
  return json({ collectablesRoute, footer })
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
    ...toRemixMeta(data.collectablesRoute.seoMeta),
    'twitter:url': 'https://fynbos.app/collectables',
    'og:url': 'https://fynbos.app/collectables'
  }
}

export default function Page() {
  const { collectablesRoute } = useLoaderData<typeof loader>()
  return (
    <>
      {collectablesRoute?.body.map((section) => (
        <MarketingPageWithSections
          key={section.id}
          section={section as SectionRecord}
        />
      ))}
    </>
  )
}
