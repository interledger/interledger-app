import type { LoaderArgs } from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { toRemixMeta } from 'react-datocms'
import type { ApplicationProps } from '~/components'
import { Fab, Layouts, MarketingPageWithSections } from '~/components'
import type { SectionRecord } from '~/generated/dato-cms-graphql'
import { getLegalRoute } from '~/lib/marketing.server'

export async function loader({ request }: LoaderArgs) {
  const { legalRoute, footer } = await getLegalRoute()
  return json({ legalRoute, footer })
}

export const handle: ApplicationProps = {
  layout: Layouts.Marketing,
  scaffold: {
    header: {},
    fab: Fab.Pay,
    footer: (match) => match.data.footer
  }
}

export function meta({ data, params }: any) {
  return {
    ...toRemixMeta(data.legalRoute.seoMeta),
    'twitter:url': 'https://fynbos.app/legal',
    'og:url': 'https://fynbos.app/legal'
  }
}

export default function Page() {
  const { legalRoute } = useLoaderData<typeof loader>()
  return (
    <>
      {legalRoute?.body.map((section) => (
        <MarketingPageWithSections
          key={section.id}
          section={section as SectionRecord}
        />
      ))}
    </>
  )
}
