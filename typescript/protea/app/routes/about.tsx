import type { LoaderArgs } from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { toRemixMeta } from 'react-datocms'
import type { ApplicationProps } from '~/components'
import { Layouts } from '~/components'
import { MarketingPageWithSections } from '~/components/Content'
import { getAboutRoute } from '~/data/content.server'
import type { SectionRecord } from '~/generated/dato-cms-graphql'

export async function loader({ request }: LoaderArgs) {
  const { aboutRoute, footer } = await getAboutRoute()
  return json({ aboutRoute, footer })
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
    ...toRemixMeta(data.aboutRoute.seoMeta),
    'twitter:url': 'https://fynbos.app/about',
    'og:url': 'https://fynbos.app/about'
  }
}

export default function Page() {
  const { aboutRoute } = useLoaderData<typeof loader>()
  return (
    <>
      {aboutRoute?.body.map((section) => (
        <MarketingPageWithSections
          key={section.id}
          section={section as SectionRecord}
        />
      ))}
    </>
  )
}
